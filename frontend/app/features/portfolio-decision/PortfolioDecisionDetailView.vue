<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "#imports";

import AppButton from "~/shared/ui/AppButton.vue";
import AppTextarea from "~/shared/ui/AppTextarea.vue";
import IMSPermissionGuard from "~/shared/ui/IMSPermissionGuard.vue";
import DecisionStatusBadge from "~/features/investment-decision/components/DecisionStatusBadge.vue";
import { usePortfolioContext } from "~/features/portfolio-workspace/composables/usePortfolioContext";
import PortfolioWorkspaceHeader from "~/features/portfolio-workspace/components/PortfolioWorkspaceHeader.vue";

import DecisionWorkflowPanel from "./components/DecisionWorkflowPanel.vue";
import { portfolioDecisionApi, type ApiDecisionV2 } from "./services/portfolioDecisionApi";
import { describeDecisionError } from "./lib/decisionErrors";
import { formatMoney, formatQuantity } from "./lib/decisionFormat";
import { decisionsListPath } from "./lib/decisionRoutes";
import { useI18n } from "~/composables/useI18n";

const props = defineProps<{ portfolioCode: string; decisionId: string }>();
const router = useRouter();
const { t } = useI18n();

const ctx = usePortfolioContext(() => props.portfolioCode);

const decision = ref<ApiDecisionV2 | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

const busy = ref(false);
const actionError = ref<string | null>(null);
const showCancelForm = ref(false);
const cancelReason = ref("");

async function load() {
  loading.value = true;
  error.value = null;
  try {
    decision.value = await portfolioDecisionApi.get(props.portfolioCode, props.decisionId);
  } catch (err) {
    decision.value = null;
    error.value = describeDecisionError(err, t("portfolio.decisionDetail.errors.load"));
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void ctx.reload();
  void load();
});
watch(() => [props.portfolioCode, props.decisionId], load);

const canSubmit = computed(() => decision.value?.status === "DRAFT");
const canCancel = computed(() =>
  ["DRAFT", "PENDING_APPROVAL", "PENDING_COMPLIANCE_RELEASE"].includes(decision.value?.status ?? ""),
);

async function onSubmit() {
  if (busy.value || !decision.value?.id) return;
  busy.value = true;
  actionError.value = null;
  try {
    decision.value = await portfolioDecisionApi.submit(props.portfolioCode, decision.value.id);
  } catch (err) {
    actionError.value = describeDecisionError(err, t("portfolio.decisionDetail.errors.submit"));
  } finally {
    busy.value = false;
  }
}

function openCancelForm() {
  showCancelForm.value = true;
  cancelReason.value = "";
}

async function confirmCancel() {
  if (busy.value || !decision.value?.id || !cancelReason.value.trim()) return;
  busy.value = true;
  actionError.value = null;
  try {
    decision.value = await portfolioDecisionApi.cancel(
      props.portfolioCode,
      decision.value.id,
      cancelReason.value.trim(),
    );
    showCancelForm.value = false;
  } catch (err) {
    actionError.value = describeDecisionError(err, t("portfolio.decisionDetail.errors.cancel"));
  } finally {
    busy.value = false;
  }
}

function backToList() {
  void router.push(decisionsListPath(props.portfolioCode));
}

function sideLabel(side: string | undefined): string {
  if (side === "BUY") return t("portfolio.decisionNew.buy");
  if (side === "SELL") return t("portfolio.decisionNew.sell");
  return side || t("common.notAvailable");
}
</script>

<template>
  <section class="decision-detail">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <div v-if="loading" class="decision-detail__notice">{{ t("portfolio.decisionDetail.loading") }}</div>
    <div v-else-if="error" class="decision-detail__error" role="alert">
      <span>{{ error }}</span>
      <button type="button" class="decision-detail__retry" @click="load">{{ t("portfolio.decisionDetail.retry") }}</button>
    </div>

    <template v-else-if="decision">
      <header class="decision-detail__header">
        <div>
          <button type="button" class="decision-detail__back" @click="backToList">
            ← {{ t("portfolio.decisionDetail.allDecisions") }}
          </button>
          <h1 class="decision-detail__title">
            {{ sideLabel(decision.side) }} {{ decision.instrument_code }}
            <DecisionStatusBadge :status="decision.status" />
          </h1>
          <p class="decision-detail__subtitle">{{ decision.decision_number }}</p>
        </div>
        <div class="decision-detail__actions">
          <IMSPermissionGuard v-if="canSubmit" permission="INVESTMENT_DECISION_SUBMIT" mode="disable">
            <AppButton variant="primary" size="sm" :loading="busy" :disabled="busy" @click="onSubmit">
              {{ t("portfolio.decisionDetail.submit") }}
            </AppButton>
          </IMSPermissionGuard>
          <IMSPermissionGuard v-if="canCancel" permission="INVESTMENT_DECISION_CANCEL" mode="disable">
            <AppButton variant="secondary" size="sm" :disabled="busy" @click="openCancelForm">
              {{ t("portfolio.decisionDetail.cancel") }}
            </AppButton>
          </IMSPermissionGuard>
        </div>
      </header>

      <p v-if="actionError" class="decision-detail__error" role="alert">{{ actionError }}</p>

      <div v-if="showCancelForm" class="decision-detail__cancel-form">
        <label class="decision-detail__label" for="cancel-reason">{{ t("portfolio.decisionDetail.cancellationReason") }}</label>
        <AppTextarea id="cancel-reason" v-model="cancelReason" :rows="2" :placeholder="t('portfolio.decisionDetail.cancellationPlaceholder')" />
        <div class="decision-detail__cancel-actions">
          <AppButton variant="ghost" size="sm" @click="showCancelForm = false">{{ t("portfolio.decisionDetail.back") }}</AppButton>
          <AppButton
            variant="danger"
            size="sm"
            :loading="busy"
            :disabled="busy || !cancelReason.trim()"
            @click="confirmCancel"
          >
            {{ t("portfolio.decisionDetail.confirmCancel") }}
          </AppButton>
        </div>
      </div>

      <div class="decision-detail__grid">
        <dl class="decision-detail__facts">
          <div><dt>{{ t("portfolio.decisionDetail.fields.side") }}</dt><dd>{{ sideLabel(decision.side) }}</dd></div>
          <div><dt>{{ t("portfolio.decisionDetail.fields.instrument") }}</dt><dd>{{ decision.instrument_code }}</dd></div>
          <div><dt>{{ t("portfolio.decisionDetail.fields.quantity") }}</dt><dd>{{ decision.quantity ? formatQuantity(decision.quantity) : t("common.notAvailable") }}</dd></div>
          <div><dt>{{ t("portfolio.decisionDetail.fields.amount") }}</dt><dd>{{ decision.amount ? formatMoney(decision.amount, decision.currency) : t("common.notAvailable") }}</dd></div>
          <div><dt>{{ t("portfolio.decisionDetail.fields.limitPrice") }}</dt><dd>{{ decision.limit_price ? formatMoney(decision.limit_price, decision.currency) : t("common.notAvailable") }}</dd></div>
          <div><dt>{{ t("portfolio.decisionDetail.fields.currency") }}</dt><dd>{{ decision.currency }}</dd></div>
          <div><dt>{{ t("portfolio.decisionDetail.fields.exchange") }}</dt><dd>{{ decision.exchange || t("common.notAvailable") }}</dd></div>
          <div><dt>{{ t("portfolio.decisionDetail.fields.businessDate") }}</dt><dd>{{ decision.business_date }}</dd></div>
          <div><dt>{{ t("portfolio.decisionDetail.fields.rationale") }}</dt><dd>{{ decision.rationale || t("common.notAvailable") }}</dd></div>
        </dl>

        <div class="decision-detail__workflow">
          <div class="decision-detail__workflow-title">{{ t("portfolio.decisionDetail.lifecycle") }}</div>
          <DecisionWorkflowPanel :status="decision.status" />
          <p class="decision-detail__note">
            {{ t("portfolio.decisionDetail.lifecycleNote") }}
          </p>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.decision-detail {
  display: grid;
  gap: 16px;
}

.decision-detail__notice {
  padding: 16px;
  font-size: 13px;
  color: var(--text-secondary);
}

.decision-detail__error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  background: var(--alert-danger-bg, #ffebe9);
  border: 1px solid var(--alert-danger-border, #cf222e);
  border-radius: var(--radius-sm, 4px);
  color: var(--alert-danger-text, #cf222e);
  font-size: 13px;
}

.decision-detail__retry {
  border: 1px solid currentColor;
  background: transparent;
  color: inherit;
  border-radius: var(--radius-sm, 4px);
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.decision-detail__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.decision-detail__back {
  border: none;
  background: transparent;
  color: var(--action-primary, #0969da);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  padding: 0;
  margin-bottom: 6px;
}

.decision-detail__title {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 10px;
}

.decision-detail__subtitle {
  margin: 4px 0 0;
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: var(--text-tertiary);
}

.decision-detail__actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.decision-detail__cancel-form {
  display: grid;
  gap: 8px;
  padding: 12px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card-muted, #f6f8fa);
}

.decision-detail__label {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--text-secondary);
}

.decision-detail__cancel-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.decision-detail__grid {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 16px;
  align-items: start;
}

.decision-detail__facts {
  margin: 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
  padding: 14px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 8px);
  background: var(--bg-card, #fff);
}

.decision-detail__facts dt {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--text-tertiary);
}

.decision-detail__facts dd {
  margin: 2px 0 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.decision-detail__workflow {
  padding: 14px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 8px);
  background: var(--bg-card, #fff);
  display: grid;
  gap: 12px;
}

.decision-detail__workflow-title {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
}

.decision-detail__note {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary);
  line-height: 1.5;
}

@media (max-width: 900px) {
  .decision-detail__grid {
    grid-template-columns: 1fr;
  }
}
</style>
