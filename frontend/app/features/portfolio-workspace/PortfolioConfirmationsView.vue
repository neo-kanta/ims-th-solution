<script setup lang="ts">
/**
 * Trade-confirmation list for a portfolio (Portfolio V2), reached from the
 * portfolio workspace tabs. Confirmations are recorded against an execution
 * elsewhere (broker import / execution-side action); this screen's own
 * write action is resolving a PENDING_REVIEW/MISMATCHED row to a target
 * status via the real `POST .../confirmations/{confirmationId}/resolve`
 * endpoint — never a frontend-only status change.
 */
import { computed, onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import IMSPermissionGuard from "~/shared/ui/IMSPermissionGuard.vue";
import { useI18n } from "~/composables/useI18n";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import { portfolioApi, type ApiTradeConfirmationV2 } from "./services/portfolioApi";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.confirmationsPage.title"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true },
);

const ctx = usePortfolioContext(() => props.portfolioCode);

const confirmations = ref<ApiTradeConfirmationV2[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const statusFilter = ref("");

const resolvingId = ref<string | null>(null);
const resolveError = ref<string | null>(null);
const mismatchReasonDraftId = ref<string | null>(null);
const mismatchReason = ref("");

const statusOptions = computed(() => [
  { value: "", label: t("portfolio.confirmationsPage.filters.allStatuses") },
  { value: "PENDING_REVIEW", label: t("portfolio.confirmationsPage.status.pendingReview") },
  { value: "MATCHED", label: t("portfolio.confirmationsPage.status.matched") },
  { value: "MISMATCHED", label: t("portfolio.confirmationsPage.status.mismatched") },
  { value: "REVIEWED", label: t("portfolio.confirmationsPage.status.reviewed") },
]);

function statusLabel(status: string | undefined): string {
  switch (status) {
    case "PENDING_REVIEW":
      return t("portfolio.confirmationsPage.status.pendingReview");
    case "MATCHED":
      return t("portfolio.confirmationsPage.status.matched");
    case "MISMATCHED":
      return t("portfolio.confirmationsPage.status.mismatched");
    case "REVIEWED":
      return t("portfolio.confirmationsPage.status.reviewed");
    default:
      return status ?? t("common.notAvailable");
  }
}

function statusTone(status: string | undefined): string {
  switch (status) {
    case "MATCHED":
    case "REVIEWED":
      return "success";
    case "MISMATCHED":
      return "rejected";
    case "PENDING_REVIEW":
      return "pending";
    default:
      return "neutral";
  }
}

const hasRows = computed(() => confirmations.value.length > 0);

async function loadConfirmations() {
  if (!props.portfolioCode) return;
  loading.value = true;
  error.value = null;
  try {
    const list = await portfolioApi.listConfirmations(props.portfolioCode, {
      limit: 200,
      status: statusFilter.value || undefined,
    });
    confirmations.value = list.items ?? [];
  } catch (err) {
    confirmations.value = [];
    error.value = err instanceof Error ? err.message : t("portfolio.confirmationsPage.errorTitle");
  } finally {
    loading.value = false;
  }
}

function canResolve(confirmation: ApiTradeConfirmationV2): boolean {
  return confirmation.status === "PENDING_REVIEW" || confirmation.status === "MISMATCHED";
}

function openMismatchReason(confirmation: ApiTradeConfirmationV2) {
  mismatchReasonDraftId.value = confirmation.id ?? null;
  mismatchReason.value = "";
}

function cancelMismatchReason() {
  mismatchReasonDraftId.value = null;
  mismatchReason.value = "";
}

async function resolve(
  confirmation: ApiTradeConfirmationV2,
  targetStatus: "MATCHED" | "MISMATCHED" | "REVIEWED",
  discrepancyReason = "",
) {
  if (!confirmation.id || resolvingId.value) return;
  resolvingId.value = confirmation.id;
  resolveError.value = null;
  try {
    await portfolioApi.resolveConfirmation(props.portfolioCode, confirmation.id, {
      target_status: targetStatus,
      discrepancy_reason: discrepancyReason || undefined,
    });
    mismatchReasonDraftId.value = null;
    mismatchReason.value = "";
    await loadConfirmations();
  } catch (err) {
    resolveError.value = err instanceof Error ? err.message : t("portfolio.confirmationsPage.resolveError");
  } finally {
    resolvingId.value = null;
  }
}

onMounted(() => {
  void ctx.reload();
  void loadConfirmations();
});
watch(
  () => props.portfolioCode,
  () => {
    void ctx.reload();
    void loadConfirmations();
  },
);
</script>

<template>
  <section class="portfolio-confirmations">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <AppCard :title="t('portfolio.confirmationsPage.title')" :subtitle="t('portfolio.confirmationsPage.subtitle')">
      <div class="portfolio-confirmations__filters">
        <label class="portfolio-confirmations__filter-label" for="confirmations-status-filter">
          {{ t("portfolio.confirmationsPage.filters.status") }}
        </label>
        <AppSelect
          id="confirmations-status-filter"
          v-model="statusFilter"
          :options="statusOptions"
          :placeholder="''"
          @update:model-value="loadConfirmations"
        />
      </div>

      <p v-if="resolveError" class="portfolio-confirmations__error" role="alert">{{ resolveError }}</p>

      <div v-if="loading && confirmations.length === 0" class="portfolio-confirmations__notice" role="status">
        {{ t("portfolio.confirmationsPage.loading") }}
      </div>
      <div v-else-if="error" class="portfolio-confirmations__error" role="alert">
        <div>{{ t("portfolio.confirmationsPage.errorTitle") }}: {{ error }}</div>
        <button type="button" class="portfolio-confirmations__retry" @click="loadConfirmations">
          {{ t("portfolio.confirmationsPage.retry") }}
        </button>
      </div>
      <div v-else-if="!hasRows" class="portfolio-confirmations__notice">
        {{ t("portfolio.confirmationsPage.empty") }}
      </div>
      <div v-else class="portfolio-confirmations__table-wrap">
        <table class="portfolio-confirmations__table">
          <thead>
            <tr>
              <th>{{ t("portfolio.confirmationsPage.columns.status") }}</th>
              <th class="right">{{ t("portfolio.confirmationsPage.columns.confirmedQty") }}</th>
              <th class="right">{{ t("portfolio.confirmationsPage.columns.confirmedPrice") }}</th>
              <th class="right">{{ t("portfolio.confirmationsPage.columns.confirmedAmount") }}</th>
              <th>{{ t("portfolio.confirmationsPage.columns.brokerReference") }}</th>
              <th>{{ t("portfolio.confirmationsPage.columns.businessDate") }}</th>
              <th class="right">{{ t("portfolio.confirmationsPage.columns.actions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="confirmation in confirmations" :key="confirmation.id">
              <td>
                <AppStatusBadge :status="statusTone(confirmation.status)" :label="statusLabel(confirmation.status)" size="sm" />
              </td>
              <td class="right">{{ confirmation.confirmed_quantity ?? "—" }}</td>
              <td class="right">{{ confirmation.confirmed_price ?? "—" }}</td>
              <td class="right">{{ confirmation.confirmed_amount ?? "—" }}</td>
              <td>{{ confirmation.broker_reference ?? t("common.notAvailable") }}</td>
              <td>{{ confirmation.business_date ?? "—" }}</td>
              <td class="right">
                <IMSPermissionGuard permission="INVESTMENT_CONFIRMATION_MANAGE" mode="disable">
                  <div v-if="canResolve(confirmation)" class="portfolio-confirmations__actions">
                    <AppButton
                      variant="secondary"
                      size="xs"
                      :disabled="resolvingId === confirmation.id"
                      :loading="resolvingId === confirmation.id"
                      @click="resolve(confirmation, 'MATCHED')"
                    >
                      {{ t("portfolio.confirmationsPage.actions.markMatched") }}
                    </AppButton>
                    <AppButton
                      variant="danger"
                      size="xs"
                      :disabled="resolvingId === confirmation.id"
                      @click="openMismatchReason(confirmation)"
                    >
                      {{ t("portfolio.confirmationsPage.actions.markMismatched") }}
                    </AppButton>
                  </div>
                  <span v-else>—</span>
                </IMSPermissionGuard>
              </td>
            </tr>
            <tr v-if="mismatchReasonDraftId" class="portfolio-confirmations__reason-row">
              <td colspan="7">
                <label class="portfolio-confirmations__reason-label" :for="`mismatch-reason-${mismatchReasonDraftId}`">
                  {{ t("portfolio.confirmationsPage.discrepancyLabel") }}
                </label>
                <div class="portfolio-confirmations__reason-controls">
                  <input
                    :id="`mismatch-reason-${mismatchReasonDraftId}`"
                    v-model="mismatchReason"
                    type="text"
                    class="portfolio-confirmations__reason-input"
                    :placeholder="t('portfolio.confirmationsPage.discrepancyPlaceholder')"
                  />
                  <AppButton variant="ghost" size="xs" @click="cancelMismatchReason">
                    {{ t("portfolio.confirmationsPage.actions.cancel") }}
                  </AppButton>
                  <AppButton
                    variant="danger"
                    size="xs"
                    :disabled="!mismatchReason.trim() || resolvingId === mismatchReasonDraftId"
                    :loading="resolvingId === mismatchReasonDraftId"
                    @click="
                      resolve(
                        confirmations.find((c) => c.id === mismatchReasonDraftId)!,
                        'MISMATCHED',
                        mismatchReason.trim(),
                      )
                    "
                  >
                    {{ t("portfolio.confirmationsPage.actions.confirmMismatch") }}
                  </AppButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </AppCard>
  </section>
</template>

<style scoped>
.portfolio-confirmations {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-confirmations__filters {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.portfolio-confirmations__filter-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary, #57606a);
}

.portfolio-confirmations__notice,
.portfolio-confirmations__error {
  padding: var(--space-4, 16px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-confirmations__error {
  color: var(--alert-danger-text, #cf222e);
  display: grid;
  gap: 6px;
}

.portfolio-confirmations__retry {
  justify-self: start;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card, #ffffff);
  border-radius: var(--radius-md, 5px);
  padding: 4px 10px;
  cursor: pointer;
}

.portfolio-confirmations__table-wrap {
  overflow-x: auto;
}

.portfolio-confirmations__table {
  width: 100%;
  min-width: 720px;
  border-collapse: collapse;
  font-size: 13px;
}

.portfolio-confirmations__table th {
  text-align: left;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  white-space: nowrap;
}

.portfolio-confirmations__table td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  white-space: nowrap;
}

.portfolio-confirmations__table th.right,
.portfolio-confirmations__table td.right {
  text-align: right;
}

.portfolio-confirmations__actions {
  display: inline-flex;
  gap: 6px;
  justify-content: flex-end;
}

.portfolio-confirmations__reason-row td {
  background: var(--bg-card-muted, #f6f8fa);
  white-space: normal;
}

.portfolio-confirmations__reason-label {
  display: block;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary, #57606a);
  margin-bottom: 6px;
}

.portfolio-confirmations__reason-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.portfolio-confirmations__reason-input {
  flex: 1;
  min-width: 0;
  height: 30px;
  padding: 0 10px;
  font-size: 13px;
  border: 1px solid var(--border-default, #d0d7de);
  border-radius: 6px;
  background: var(--bg-input, #fff);
  color: var(--text-primary, #1f2328);
}
</style>
