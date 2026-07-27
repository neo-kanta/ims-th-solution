<script setup lang="ts">
/**
 * Submitter-facing panel of a LIVE portfolio's cash-movement approval
 * requests (Stage 2 of the LIVE cash-transaction approval gate — see
 * docs/investment/live-cash-transaction-approval-design.md §5). Shows
 * PENDING/APPROVED/REJECTED/CANCELLED requests and lets the submitter
 * cancel a request that is still PENDING. Only meaningful for LIVE
 * portfolios — SIMULATION/MODEL never create cash requests, so callers
 * should only render this for ctx.portfolioType === "LIVE".
 */
import { onMounted, ref, watch } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import AppConfirmDialog from "~/shared/ui/AppConfirmDialog.vue";
import { useI18n } from "~/composables/useI18n";
import { useAuthStore } from "~/stores/useAuthStore";
import { formatMoney } from "~/features/investment-ledger/lib/ledgerFormat";
import { canCancelCashRequest } from "../lib/cashRequestAuthz";
import { portfolioApi, type ApiCashRequestV2 } from "../services/portfolioApi";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();
const authStore = useAuthStore();

function canCancel(item: ApiCashRequestV2): boolean {
  return canCancelCashRequest(item, authStore.user?.id);
}

const items = ref<ApiCashRequestV2[]>([]);
const loading = ref(false);
const loadError = ref<string | null>(null);

const cancelTarget = ref<ApiCashRequestV2 | null>(null);
const cancelling = ref(false);
const cancelError = ref<string | null>(null);

async function load() {
  loading.value = true;
  loadError.value = null;
  try {
    const res = await portfolioApi.listCashRequests(props.portfolioCode, {
      limit: 50,
    });
    items.value = res.items ?? [];
  } catch (err) {
    loadError.value =
      (err as { data?: { error?: string } })?.data?.error ||
      (err as Error)?.message ||
      t("portfolio.cashRequests.loadError");
  } finally {
    loading.value = false;
  }
}

onMounted(() => void load());
watch(
  () => props.portfolioCode,
  () => void load(),
);

defineExpose({ reload: load });

function statusVariant(status?: string): string {
  switch (status) {
    case "PENDING":
      return "pending";
    case "APPROVED":
      return "approved";
    case "REJECTED":
      return "rejected";
    case "CANCELLED":
      return "neutral";
    default:
      return "neutral";
  }
}

function statusLabel(status?: string): string {
  switch (status) {
    case "PENDING":
      return t("portfolio.cashRequests.status.pending");
    case "APPROVED":
      return t("portfolio.cashRequests.status.approved");
    case "REJECTED":
      return t("portfolio.cashRequests.status.rejected");
    case "CANCELLED":
      return t("portfolio.cashRequests.status.cancelled");
    default:
      return status || t("common.notAvailable");
  }
}

function typeLabel(type?: string): string {
  switch (type) {
    case "CASH_IN":
      return t("portfolio.ledgerNew.cashTypeCashIn");
    case "CASH_OUT":
      return t("portfolio.ledgerNew.cashTypeCashOut");
    case "FEE":
      return t("portfolio.ledgerNew.cashTypeFee");
    case "DIVIDEND":
      return t("portfolio.ledgerNew.cashTypeDividend");
    default:
      return type || t("common.notAvailable");
  }
}

function formatDateTime(value?: string): string {
  if (!value) return t("common.notAvailable");
  return value.replace("T", " ").slice(0, 16);
}

function openCancel(item: ApiCashRequestV2) {
  cancelTarget.value = item;
  cancelError.value = null;
}

function closeCancel() {
  if (cancelling.value) return;
  cancelTarget.value = null;
  cancelError.value = null;
}

async function confirmCancel() {
  const target = cancelTarget.value;
  if (!target?.id) return;
  cancelling.value = true;
  cancelError.value = null;
  try {
    await portfolioApi.cancelCashRequest(props.portfolioCode, target.id);
    cancelTarget.value = null;
    await load();
  } catch (err) {
    cancelError.value =
      (err as { data?: { error?: string } })?.data?.error ||
      (err as Error)?.message ||
      t("portfolio.cashRequests.cancelError");
  } finally {
    cancelling.value = false;
  }
}
</script>

<template>
  <AppCard
    :title="t('portfolio.cashRequests.title')"
    :subtitle="t('portfolio.cashRequests.subtitle')"
  >
    <div v-if="loading" class="cash-requests-panel__notice" role="status">
      {{ t("portfolio.cashRequests.loading") }}
    </div>
    <div v-else-if="loadError" class="cash-requests-panel__error" role="alert">
      {{ loadError }}
    </div>
    <div v-else-if="items.length === 0" class="cash-requests-panel__notice">
      {{ t("portfolio.cashRequests.empty") }}
    </div>
    <div v-else class="cash-requests-panel__table">
      <table>
        <thead>
          <tr>
            <th>{{ t("portfolio.cashRequests.columns.type") }}</th>
            <th>{{ t("portfolio.cashRequests.columns.amount") }}</th>
            <th>{{ t("portfolio.cashRequests.columns.status") }}</th>
            <th>{{ t("portfolio.cashRequests.columns.submitted") }}</th>
            <th class="col-action">{{ t("portfolio.cashRequests.columns.action") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td>{{ typeLabel(item.transaction_type) }}</td>
            <td class="cell-mono">{{ formatMoney(item.amount, item.currency) }}</td>
            <td>
              <AppStatusBadge
                :status="statusVariant(item.status)"
                :label="statusLabel(item.status)"
              />
            </td>
            <td>{{ formatDateTime(item.submitted_at) }}</td>
            <td class="col-action">
              <button
                v-if="canCancel(item)"
                type="button"
                class="btn btn-danger btn-sm"
                @click="openCancel(item)"
              >
                {{ t("portfolio.cashRequests.cancel") }}
              </button>
              <span v-else>—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <AppConfirmDialog
      :open="cancelTarget !== null"
      :title="t('portfolio.cashRequests.cancelConfirmTitle')"
      :description="t('portfolio.cashRequests.cancelConfirmDescription')"
      :confirm-label="t('portfolio.cashRequests.cancelConfirmAction')"
      :cancel-label="t('portfolio.cashRequests.cancelConfirmBack')"
      tone="danger"
      :loading="cancelling"
      @cancel="closeCancel"
      @confirm="confirmCancel"
    />
    <p v-if="cancelError" class="cash-requests-panel__error" role="alert">
      {{ cancelError }}
    </p>
  </AppCard>
</template>

<style scoped>
.cash-requests-panel__notice,
.cash-requests-panel__error {
  padding: var(--space-4, 16px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.cash-requests-panel__error {
  color: var(--alert-danger-text, #cf222e);
}

.cash-requests-panel__table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

th,
td {
  padding: 8px 10px;
  text-align: left;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  white-space: nowrap;
}

th {
  background: var(--bg-card-muted, #f6f8fa);
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary, #57606a);
}

.cell-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}

.col-action {
  width: 96px;
}
</style>
