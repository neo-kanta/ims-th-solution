<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import DecisionCockpit from "~/features/investment-decision/components/DecisionCockpit.vue";
import DecisionDetailPanel from "~/features/investment-decision/components/DecisionDetailPanel.vue";
import DecisionApprovalPanel from "~/features/investment-decision/components/DecisionApprovalPanel.vue";
import { useDecisionDetail } from "~/features/investment-decision/composables/useDecisionDetail";
import { useDecisionMutations } from "~/features/investment-decision/composables/useDecisionMutations";
import { investmentLedgerApi } from "~/features/investment-ledger/services/investmentLedgerApi";
import { myFundsApi } from "~/features/my-funds/services/myFundsApi";
import { useI18n } from "~/composables/useI18n";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const decisionId = computed(() => {
  const value = route.params.decisionId;
  return typeof value === "string" ? value : "";
});

const { decision, loading, error, fetch } = useDecisionDetail();
const mutations = useDecisionMutations();

const cancelDialog = ref(false);
const cancelReason = ref("");
const portfolioLabel = ref("");
const fundLabel = ref("");

function joinBusinessLabel(code?: string, name?: string): string {
  return [code, name].filter((value): value is string => Boolean(value?.trim())).join(" — ");
}

async function loadBusinessLabels() {
  portfolioLabel.value = "";
  fundLabel.value = "";

  const portfolioId = decision.value?.portfolio_id;
  const fundId = decision.value?.fund_id;
  await Promise.all([
    portfolioId
      ? investmentLedgerApi
          .getPortfolio(portfolioId)
          .then((portfolio) => {
            portfolioLabel.value = joinBusinessLabel(portfolio.code, portfolio.name);
          })
          .catch(() => undefined)
      : Promise.resolve(),
    fundId
      ? myFundsApi
          .listMyFunds()
          .then((funds) => {
            const fund = funds.find((item) => item.id === fundId);
            fundLabel.value = joinBusinessLabel(fund?.code, fund?.name);
          })
          .catch(() => undefined)
      : Promise.resolve(),
  ]);
}

async function handleSubmit() {
  if (!decision.value?.id) return;
  const result = await mutations.submit(decision.value.id);
  if (result) decision.value = result;
}

async function confirmCancel() {
  if (!cancelReason.value.trim() || !decision.value?.id) return;
  const result = await mutations.cancel(decision.value.id, cancelReason.value);
  if (result) {
    decision.value = result;
    cancelDialog.value = false;
    cancelReason.value = "";
  }
}

const canSubmit = computed(
  () => decision.value?.status === "DRAFT" && !mutations.loading.value,
);
const canCancel = computed(
  () =>
    (decision.value?.status === "DRAFT" ||
      decision.value?.status === "SUBMITTED") &&
    !mutations.loading.value,
);

onMounted(() => {
  void fetch(decisionId.value);
});

watch(decision, () => {
  void loadBusinessLabels();
});
</script>

<template>
  <div>
    <div v-if="loading" class="loading-state">{{ t("operator.operation.loadingDecision") }}</div>
    <div v-else-if="error" class="error-banner">{{ error }}</div>
    <div v-else-if="!decision" class="empty-state">{{ t("operator.operation.decisionNotFound") }}</div>

    <DecisionCockpit v-if="decision">
      <template #title>
        {{ decision.decision_number ?? t("operator.operation.decisionFallback") }}
      </template>
      <template #toolbar>
        <button
          v-if="canSubmit"
          class="toolbar-btn toolbar-btn--primary"
          :disabled="mutations.loading.value"
          @click="handleSubmit"
        >
          {{ t("operator.actions.submit") }}
        </button>
        <button
          v-if="canCancel"
          class="toolbar-btn toolbar-btn--danger"
          :disabled="mutations.loading.value"
          @click="cancelDialog = true"
        >
          {{ t("operator.operation.cancelTitle") }}
        </button>
        <button class="toolbar-btn" @click="router.back()">← {{ t("operator.actions.back") }}</button>
      </template>

      <template #left>
        <div class="panel-header">{{ t("operator.operation.portfolio") }}</div>
        <div class="panel-info">
          <div class="info-item">
            <span class="info-label">{{ t("operator.operation.portfolio") }}</span>
            <span class="info-value">{{ portfolioLabel || t("operator.operation.unavailableBusinessLabel") }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">{{ t("operator.operation.fund") }}</span>
            <span class="info-value">{{ fundLabel || t("operator.operation.unavailableBusinessLabel") }}</span>
          </div>
        </div>
      </template>

      <template #middle>
        <DecisionDetailPanel :decision="decision" />
      </template>

      <template #right>
        <DecisionApprovalPanel :decision="decision" />
      </template>
    </DecisionCockpit>

    <div v-if="mutations.error.value" class="error-banner mt">
      {{ mutations.error.value }}
    </div>

    <div v-if="cancelDialog" class="dialog-overlay" @click.self="cancelDialog = false">
      <div class="dialog">
        <div class="dialog__title">{{ t("operator.operation.cancelTitle") }}</div>
        <div class="dialog__body">
          <label class="dialog__label">{{ t("operator.operation.reason") }} <span class="required">*</span></label>
          <textarea v-model="cancelReason" rows="3" class="dialog__textarea" />
        </div>
        <div class="dialog__actions">
          <button class="dialog__btn" @click="cancelDialog = false">{{ t("operator.actions.back") }}</button>
          <button
            class="dialog__btn dialog__btn--danger"
            :disabled="mutations.loading.value || !cancelReason.trim()"
            @click="confirmCancel"
          >
            {{ t("operator.actions.confirmCancel") }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.loading-state,
.empty-state {
  padding: 32px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 14px;
}

.error-banner {
  padding: 10px 14px;
  margin-bottom: 12px;
  background: rgba(207, 34, 46, 0.1);
  border: 1px solid var(--state-danger, #cf222e);
  border-radius: var(--radius-sm, 4px);
  color: var(--state-danger, #cf222e);
  font-size: 13px;
}

.mt {
  margin-top: 12px;
}

.toolbar-btn {
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: background-color 0.1s;
}

.toolbar-btn:hover {
  background: var(--bg-card-hover);
}

.toolbar-btn--primary {
  background: var(--state-success, #1a7f37);
  color: #ffffff;
  border-color: transparent;
}

.toolbar-btn--primary:hover {
  opacity: 0.9;
}

.toolbar-btn--danger {
  background: var(--state-danger, #cf222e);
  color: #ffffff;
  border-color: transparent;
}

.toolbar-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.panel-header {
  background: var(--bg-card-hover);
  border-bottom: 1px solid var(--border-subtle);
  padding: 8px 12px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
}

.panel-info {
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-secondary);
}

.info-value {
  font-size: 12px;
  color: var(--text-primary);
  word-break: break-all;
}

.mono {
  font-family: monospace;
}

.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.dialog {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 20px 24px;
  min-width: 360px;
  max-width: 480px;
  box-shadow: var(--shadow-md);
}

.dialog__title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 14px;
}

.dialog__body {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.dialog__label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}

.dialog__textarea {
  width: 100%;
  padding: 7px 10px;
  font-size: 13px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  color: var(--text-primary);
  box-sizing: border-box;
  resize: none;
}

.dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.dialog__btn {
  padding: 6px 16px;
  font-size: 13px;
  font-weight: 600;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  color: var(--text-primary);
  cursor: pointer;
}

.dialog__btn--danger {
  background: var(--state-danger, #cf222e);
  color: #fff;
  border-color: transparent;
}

.dialog__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.required {
  color: var(--state-danger, #cf222e);
}
</style>
