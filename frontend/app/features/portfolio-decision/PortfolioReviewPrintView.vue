<script setup lang="ts">
/**
 * OP-03 "Review & Print Summary" — portfolio-first. Lets the operator pick
 * any portfolio they can view, filter its processed decisions by business
 * date / lifecycle status / decision number, open the authoritative V2
 * decision detail, and print a summary sheet.
 *
 * Print sections backed by real, typed endpoints: decision header,
 * transaction detail, lines, rationale, compliance result
 * (`GET /compliance/checks/{groupID}`), approval history
 * (`GET /approvals/requests/{requestId}/timeline`), and — now that a typed
 * contract exists — execution/fill and trade-confirmation status
 * (`GET /portfolios/{portfolioCode}/executions` matched by
 * `decision_id`, then `GET /portfolios/{portfolioCode}/confirmations`
 * matched by `execution_id`; see usePrintSummarySections.ts).
 */
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "#imports";

import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppBadge from "~/shared/ui/AppBadge.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppDateField from "~/shared/ui/AppDateField.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import { useI18n } from "~/composables/useI18n";

import { usePortfolioPicker } from "./composables/usePortfolioPicker";
import { usePortfolioDecisionsList } from "./composables/usePortfolioDecisionsList";
import { usePrintSummarySections } from "./composables/usePrintSummarySections";
import { formatMoney, formatQuantity, decisionStatusKey } from "./lib/decisionFormat";
import { decisionDetailPath } from "./lib/decisionRoutes";
import { portfolioDecisionApi, type ApiDecisionV2 } from "./services/portfolioDecisionApi";
import { myFundsApi } from "~/features/my-funds/services/myFundsApi";
import type { ApiExecutionV2 } from "~/features/portfolio-workspace/services/portfolioApi";

const router = useRouter();
const { t } = useI18n();

const { filtered: filteredPortfolios, search: portfolioSearch, loading: portfoliosLoading, load: loadPortfolios } =
  usePortfolioPicker();

const selectedPortfolioCode = ref("");
const list = usePortfolioDecisionsList();
// OP-03 exists to browse a portfolio's decision history, so default to the
// API's maximum page size (200) rather than the 20-row default meant for a
// quick recent-activity list; pagination below still covers portfolios with
// more than 200 decisions.
list.limit.value = 200;

const businessDateFilter = ref("");
const statusFilter = ref("ALL");
const decisionNumberFilter = ref("");

const STATUS_OPTIONS = [
  "ALL",
  "DRAFT",
  "SUBMITTED",
  "PENDING_APPROVAL",
  "APPROVED",
  "READY_FOR_EXECUTION",
  "REJECTED",
  "CANCELLED",
] as const;

const portfolioOptions = computed(() =>
  filteredPortfolios.value.map((p) => ({ value: p.code, label: `${p.code} — ${p.name}` })),
);

const statusOptions = computed(() =>
  STATUS_OPTIONS.map((s) => ({
    value: s,
    label: s === "ALL" ? t("operator.review.allDecisions") : t(decisionStatusKey(s) ?? "common.notAvailable"),
  })),
);

const filteredDecisions = computed<ApiDecisionV2[]>(() => {
  const numberQuery = decisionNumberFilter.value.trim().toLowerCase();
  return list.items.value.filter((d) => {
    if (statusFilter.value !== "ALL" && d.status !== statusFilter.value) return false;
    if (businessDateFilter.value && d.business_date?.slice(0, 10) !== businessDateFilter.value) return false;
    if (numberQuery && !d.decision_number?.toLowerCase().includes(numberQuery)) return false;
    return true;
  });
});

const selectedDecision = ref<ApiDecisionV2 | null>(null);
const selectedDecisionLoading = ref(false);
const selectedDecisionError = ref<string | null>(null);
const selectedFundLabel = ref("");
const summary = usePrintSummarySections();

async function onPortfolioSelected() {
  selectedDecision.value = null;
  businessDateFilter.value = "";
  statusFilter.value = "ALL";
  decisionNumberFilter.value = "";
  if (!selectedPortfolioCode.value) return;
  list.page.value = 1;
  await list.load(selectedPortfolioCode.value);
}

async function openDecision(decision: ApiDecisionV2) {
  if (!decision.id) return;
  selectedDecisionLoading.value = true;
  selectedDecisionError.value = null;
  selectedFundLabel.value = "";
  try {
    const detail = await portfolioDecisionApi.get(selectedPortfolioCode.value, decision.id);
    selectedDecision.value = detail;
    await summary.load(detail, selectedPortfolioCode.value);
    if (detail.fund_id) {
      const funds = await myFundsApi.listMyFunds(200);
      const fund = funds.find((f) => f.id === detail.fund_id);
      selectedFundLabel.value = fund ? `${fund.code} — ${fund.name}` : "";
    }
  } catch (err) {
    selectedDecisionError.value = err instanceof Error ? err.message : t("operator.review.loadError");
  } finally {
    selectedDecisionLoading.value = false;
  }
}

function openAuthoritativeDetail() {
  if (!selectedDecision.value?.id) return;
  void router.push(decisionDetailPath(selectedPortfolioCode.value, selectedDecision.value.id));
}

function triggerPrint() {
  window.print();
}

function onDecisionRowKeydown(event: KeyboardEvent, item: ApiDecisionV2) {
  if (event.key !== "Enter" && event.key !== " ") return;
  event.preventDefault();
  void openDecision(item);
}

function sideLabel(side: string | undefined): string {
  if (side === "BUY") return t("portfolio.decisionNew.buy");
  if (side === "SELL") return t("portfolio.decisionNew.sell");
  return side || t("common.notAvailable");
}

function statusLabel(status: string | undefined): string {
  const key = decisionStatusKey(status);
  return key ? t(key) : status || t("common.notAvailable");
}

function verdictLabel(verdict: string | undefined): string {
  switch (verdict) {
    case "PASS":
      return t("operator.review.compliance.verdictPass");
    case "WARN":
      return t("operator.review.compliance.verdictWarn");
    case "BLOCK":
      return t("operator.review.compliance.verdictBlock");
    default:
      return verdict || t("common.notAvailable");
  }
}

function eventLabel(event: { event_type?: string }): string {
  return event.event_type ? event.event_type.replace(/_/g, " ") : t("common.notAvailable");
}

function executionStatusLabel(status: string | undefined): string {
  switch (status) {
    case "PENDING":
      return t("operator.review.execution.status.pending");
    case "EXECUTED":
      return t("operator.review.execution.status.filled");
    case "PARTIALLY_EXECUTED":
      return t("operator.review.execution.status.partiallyFilled");
    case "CANCELLED":
      return t("operator.review.execution.status.cancelled");
    default:
      return status || t("common.notAvailable");
  }
}

function confirmationStatusLabel(status: string | undefined): string {
  switch (status) {
    case "PENDING_REVIEW":
      return t("operator.review.execution.confirmation.status.pendingReview");
    case "MATCHED":
      return t("operator.review.execution.confirmation.status.matched");
    case "MISMATCHED":
      return t("operator.review.execution.confirmation.status.mismatched");
    case "REVIEWED":
      return t("operator.review.execution.confirmation.status.reviewed");
    default:
      return status || t("common.notAvailable");
  }
}

function confirmationFor(execution: ApiExecutionV2) {
  if (!execution.id) return null;
  return summary.confirmationsByExecutionId.value[execution.id] ?? null;
}

onMounted(() => {
  void loadPortfolios();
});

watch(selectedPortfolioCode, onPortfolioSelected);
</script>

<template>
  <div>
    <AppPageHeader :title="t('operator.review.title')">
      <template #actions>
        <AppButton variant="secondary" @click="router.push('/investment/operator')">
          ← {{ t("operator.actions.backToOperator") }}
        </AppButton>
      </template>
    </AppPageHeader>

    <section class="review-page">
      <div class="split-layout no-print">
        <!-- Left Column: Filters + List -->
        <div class="split-layout__left">
          <AppCard>
            <div class="controls-row">
              <div class="control-group search-group">
                <label class="control-label" for="op03-portfolio">{{ t("operator.review.filters.portfolioCode") }}</label>
                <AppInput
                  id="op03-portfolio-search"
                  v-model="portfolioSearch"
                  type="search"
                  :placeholder="t('operator.review.filters.portfolioSearchPlaceholder')"
                />
                <AppSelect
                  id="op03-portfolio"
                  v-model="selectedPortfolioCode"
                  :options="portfolioOptions"
                  :placeholder="t('operator.review.filters.portfolioPlaceholder')"
                  :disabled="portfoliosLoading"
                />
              </div>
            </div>

            <template v-if="selectedPortfolioCode">
              <div class="controls-row mt">
                <div class="control-group">
                  <label class="control-label">{{ t("operator.review.statusFilter") }}</label>
                  <AppSelect v-model="statusFilter" :options="statusOptions" :placeholder="''" />
                </div>
                <div class="control-group">
                  <label class="control-label">{{ t("operator.review.filters.businessDate") }}</label>
                  <AppDateField v-model="businessDateFilter" />
                </div>
                <div class="control-group search-group">
                  <label class="control-label">{{ t("operator.review.search") }}</label>
                  <AppInput
                    v-model="decisionNumberFilter"
                    type="search"
                    :placeholder="t('operator.review.searchPlaceholder')"
                  />
                </div>
              </div>
            </template>
          </AppCard>

          <p v-if="!selectedPortfolioCode" class="empty-hint">
            {{ t("operator.review.filters.selectPortfolioHint") }}
          </p>

          <AppCard v-else class="mt">
            <div v-if="list.error.value" class="list-error" role="alert">
              <span>{{ list.error.value }}</span>
              <button type="button" class="list-error__retry" @click="onPortfolioSelected">
                {{ t("portfolio.decisionDetail.retry") }}
              </button>
            </div>
            <div class="table-container">
              <table class="operator-table">
                <thead>
                  <tr>
                    <th>{{ t("operator.review.columns.decisionNumber") }}</th>
                    <th>{{ t("operator.review.columns.side") }}</th>
                    <th>{{ t("operator.review.columns.instrument") }}</th>
                    <th class="right">{{ t("operator.review.columns.quantity") }}</th>
                    <th>{{ t("operator.review.columns.status") }}</th>
                    <th>{{ t("operator.review.columns.date") }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="list.loading.value">
                    <td colspan="6" class="text-center p-4">{{ t("operator.review.loading") }}</td>
                  </tr>
                  <tr v-else-if="!list.error.value && filteredDecisions.length === 0">
                    <td colspan="6" class="text-center p-4 text-secondary">{{ t("operator.review.empty") }}</td>
                  </tr>
                  <tr
                    v-for="item in filteredDecisions"
                    :key="item.id"
                    class="clickable-row"
                    :class="{ 'is-active': selectedDecision?.id === item.id }"
                    tabindex="0"
                    role="button"
                    @click="openDecision(item)"
                    @keydown="onDecisionRowKeydown($event, item)"
                  >
                    <td class="bold">{{ item.decision_number || '—' }}</td>
                    <td>
                      <AppBadge :variant="item.side === 'BUY' ? 'success' : 'error'">{{ sideLabel(item.side) }}</AppBadge>
                    </td>
                    <td class="bold">{{ item.instrument_code || '—' }}</td>
                    <td class="right">{{ item.quantity || '—' }}</td>
                    <td>
                      <AppBadge :variant="item.status === 'APPROVED' || item.status === 'READY_FOR_EXECUTION' ? 'success' : 'neutral'">
                        {{ statusLabel(item.status) }}
                      </AppBadge>
                    </td>
                    <td>{{ item.business_date ? item.business_date.slice(0, 10) : '—' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-if="list.total.value > list.limit.value" class="pagination-row">
              <button
                type="button"
                class="pagination-button"
                :disabled="list.page.value <= 1 || list.loading.value"
                @click="list.setPage(list.page.value - 1, selectedPortfolioCode)"
              >
                ‹ {{ t("portfolio.decisionList.previous") }}
              </button>
              <span class="paging-note">
                {{
                  t("operator.review.filters.pageInfo", {
                    page: list.page.value,
                    total: Math.ceil(list.total.value / list.limit.value),
                  })
                }}
              </span>
              <button
                type="button"
                class="pagination-button"
                :disabled="list.page.value * list.limit.value >= list.total.value || list.loading.value"
                @click="list.setPage(list.page.value + 1, selectedPortfolioCode)"
              >
                {{ t("portfolio.decisionList.next") }} ›
              </button>
            </div>
            <p v-if="list.total.value > list.limit.value" class="paging-note">
              {{ t("operator.review.filters.pageCapNote") }}
            </p>
          </AppCard>
        </div>

        <!-- Right Column: Print Sheet Preview -->
        <div class="split-layout__right">
          <AppCard :title="t('operator.review.previewTitle')" :subtitle="t('operator.review.previewSubtitle')">
            <div v-if="selectedDecisionLoading" class="p-6 text-center">
              <AppLoadingState :message="t('operator.review.previewLoading')" />
            </div>
            <div v-else-if="selectedDecisionError" class="p-6 text-center text-secondary">
              {{ selectedDecisionError }}
            </div>
            <div v-else-if="!selectedDecision" class="p-6 text-center text-secondary">
              {{ t("operator.review.previewEmpty") }}
            </div>
            <div v-else class="review-preview-card">
              <div class="preview-actions mb-4">
                <AppButton variant="secondary" size="sm" @click="openAuthoritativeDetail">
                  {{ t("operator.review.openAuthoritative") }}
                </AppButton>
                <AppButton variant="primary" size="sm" @click="triggerPrint">
                  {{ t("operator.actions.print") }}
                </AppButton>
              </div>

              <div class="preview-paper">
                <div class="paper-title">{{ t("operator.review.sheetTitle") }}</div>
                <div class="paper-metadata">
                  <div><strong>{{ t("operator.review.number") }}:</strong> {{ selectedDecision.decision_number }}</div>
                  <div><strong>{{ t("operator.review.date") }}:</strong> {{ selectedDecision.business_date }}</div>
                  <div><strong>{{ t("operator.review.status") }}:</strong> {{ statusLabel(selectedDecision.status) }}</div>
                </div>
                <div class="paper-body">
                  <div class="paper-field"><strong>{{ t("operator.review.instrument") }}:</strong> {{ selectedDecision.instrument_code }}</div>
                  <div class="paper-field"><strong>{{ t("operator.review.side") }}:</strong> {{ sideLabel(selectedDecision.side) }}</div>
                  <div class="paper-field"><strong>{{ t("operator.review.quantityPrice") }}:</strong> {{ selectedDecision.quantity }} @ {{ selectedDecision.limit_price }} {{ selectedDecision.currency }}</div>
                </div>
              </div>
            </div>
          </AppCard>
        </div>
      </div>

      <!-- Printable Document Element -->
      <div v-if="selectedDecision" class="print-area">
        <div class="print-header">
          <div class="print-logo">TH-IMS</div>
          <div class="print-title-block">
            <h1 class="print-title">{{ t("operator.review.reportTitle") }}</h1>
            <div class="print-subtitle">{{ t("operator.review.referenceNumber") }}: {{ selectedDecision.decision_number || t("common.notAvailable") }}</div>
          </div>
        </div>

        <div class="print-section">
          <h2 class="print-section-title">{{ t("operator.review.sections.metadata") }}</h2>
          <table class="print-table">
            <tbody>
              <tr>
                <th>{{ t("operator.review.fields.portfolio") }}</th>
                <td>{{ selectedPortfolioCode }}</td>
                <th>{{ t("operator.review.fields.fund") }}</th>
                <td>{{ selectedFundLabel || t("operator.operation.unavailableBusinessLabel") }}</td>
              </tr>
              <tr>
                <th>{{ t("operator.review.fields.businessDate") }}</th>
                <td>{{ selectedDecision.business_date ? selectedDecision.business_date.slice(0, 10) : '—' }}</td>
                <th>{{ t("operator.review.fields.lifecycleStatus") }}</th>
                <td>{{ statusLabel(selectedDecision.status) }}</td>
              </tr>
              <tr>
                <th>{{ t("operator.review.fields.submittedAt") }}</th>
                <td>{{ selectedDecision.submitted_at || '—' }}</td>
                <th>{{ t("operator.review.fields.createdAt") }}</th>
                <td>{{ selectedDecision.created_at || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="print-section">
          <h2 class="print-section-title">{{ t("operator.review.sections.transaction") }}</h2>
          <table class="print-table">
            <tbody>
              <tr>
                <th>{{ t("operator.review.fields.instrumentCode") }}</th>
                <td>{{ selectedDecision.instrument_code }}</td>
                <th>{{ t("operator.review.fields.exchange") }}</th>
                <td>{{ selectedDecision.exchange || t("common.notAvailable") }}</td>
              </tr>
              <tr>
                <th>{{ t("operator.review.side") }}</th>
                <td><strong>{{ sideLabel(selectedDecision.side) }}</strong></td>
                <th>{{ t("operator.review.fields.currency") }}</th>
                <td>{{ selectedDecision.currency }}</td>
              </tr>
              <tr>
                <th>{{ t("operator.review.fields.quantity") }}</th>
                <td>{{ selectedDecision.quantity ? formatQuantity(selectedDecision.quantity) : '—' }}</td>
                <th>{{ t("operator.review.fields.limitPrice") }}</th>
                <td>{{ selectedDecision.limit_price ? formatMoney(selectedDecision.limit_price, selectedDecision.currency) : '—' }}</td>
              </tr>
              <tr>
                <th>{{ t("operator.review.fields.grossAmount") }}</th>
                <td colspan="3">{{ selectedDecision.amount ? formatMoney(selectedDecision.amount, selectedDecision.currency) : '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="print-section">
          <h2 class="print-section-title">{{ t("operator.review.sections.rationale") }}</h2>
          <div class="print-rationale">
            {{ selectedDecision.rationale || t("operator.review.noRationale") }}
          </div>
        </div>

        <div class="print-section">
          <h2 class="print-section-title">{{ t("operator.review.sections.compliance") }}</h2>
          <template v-if="!selectedDecision.compliance_check_group_id">
            <p class="print-not-recorded">{{ t("operator.review.compliance.noRecord") }}</p>
          </template>
          <template v-else-if="summary.compliance.loading.value">
            <p class="print-not-recorded">{{ t("operator.review.compliance.loading") }}</p>
          </template>
          <template v-else-if="summary.compliance.error.value">
            <p class="print-not-recorded">{{ summary.compliance.error.value }}</p>
          </template>
          <template v-else-if="summary.compliance.result.value">
            <table class="print-table">
              <thead>
                <tr>
                  <th>{{ t("operator.review.compliance.ruleType") }}</th>
                  <th>{{ t("operator.review.compliance.verdict") }}</th>
                  <th>{{ t("operator.review.compliance.message") }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="record in summary.compliance.result.value.records" :key="record.id">
                  <td>{{ record.ruleTypeID }}</td>
                  <td>{{ verdictLabel(record.finalVerdict) }}</td>
                  <td>{{ record.message }}</td>
                </tr>
              </tbody>
            </table>
            <p v-if="summary.compliance.result.value.breaches.length" class="print-breach-note">
              {{ t("operator.review.compliance.breachCount", { count: summary.compliance.result.value.breaches.length }) }}
            </p>
          </template>
        </div>

        <div class="print-section">
          <h2 class="print-section-title">{{ t("operator.review.sections.approval") }}</h2>
          <template v-if="!selectedDecision.approval_request_id">
            <p class="print-not-recorded">{{ t("operator.review.approvalHistory.noRecord") }}</p>
          </template>
          <template v-else-if="summary.approvalHistoryLoading.value">
            <p class="print-not-recorded">{{ t("operator.review.compliance.loading") }}</p>
          </template>
          <template v-else-if="summary.approvalHistoryError.value">
            <p class="print-not-recorded">{{ summary.approvalHistoryError.value }}</p>
          </template>
          <template v-else-if="summary.approvalHistory.value.length">
            <table class="print-table">
              <thead>
                <tr>
                  <th>{{ t("operator.review.approvalHistory.actor") }}</th>
                  <th>{{ t("operator.review.approvalHistory.action") }}</th>
                  <th>{{ t("operator.review.approvalHistory.comment") }}</th>
                  <th>{{ t("operator.review.approvalHistory.date") }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="event in summary.approvalHistory.value" :key="event.id">
                  <td>{{ event.actor_name || event.actor_user_id || t("common.notAvailable") }}</td>
                  <td>{{ eventLabel(event) }}</td>
                  <td>{{ event.comment || '—' }}</td>
                  <td>{{ event.created_at || '—' }}</td>
                </tr>
              </tbody>
            </table>
          </template>
          <template v-else>
            <p class="print-not-recorded">{{ t("operator.review.approvalHistory.noRecord") }}</p>
          </template>
        </div>

        <div class="print-section">
          <h2 class="print-section-title">{{ t("operator.review.sections.executionSummary") }}</h2>
          <template v-if="summary.executionsLoading.value">
            <p class="print-not-recorded">{{ t("operator.review.execution.loading") }}</p>
          </template>
          <template v-else-if="summary.executionsError.value">
            <p class="print-not-recorded">{{ summary.executionsError.value }}</p>
          </template>
          <template v-else-if="!summary.executions.value.length">
            <p class="print-not-recorded">{{ t("operator.review.execution.noRecord") }}</p>
          </template>
          <template v-else>
            <table class="print-table">
              <thead>
                <tr>
                  <th>{{ t("operator.review.execution.columns.status") }}</th>
                  <th>{{ t("operator.review.execution.columns.executedQty") }}</th>
                  <th>{{ t("operator.review.execution.columns.executedAmount") }}</th>
                  <th>{{ t("operator.review.execution.columns.executionPrice") }}</th>
                  <th>{{ t("operator.review.execution.columns.executedAt") }}</th>
                  <th>{{ t("operator.review.execution.columns.confirmation") }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="execution in summary.executions.value" :key="execution.id">
                  <td>{{ executionStatusLabel(execution.status) }}</td>
                  <td>{{ execution.executed_quantity ?? "—" }}</td>
                  <td>{{ execution.executed_amount ? formatMoney(execution.executed_amount, execution.currency) : "—" }}</td>
                  <td>{{ execution.execution_price ?? "—" }}</td>
                  <td>{{ execution.executed_at ?? "—" }}</td>
                  <td>
                    <span v-if="confirmationFor(execution)">
                      {{ confirmationStatusLabel(confirmationFor(execution)?.status) }}
                    </span>
                    <span v-else class="print-not-recorded">
                      {{ t("operator.review.execution.confirmation.noRecord") }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </template>
        </div>

        <div class="print-signatures">
          <div class="print-signature-box">
            <div class="print-signature-line"></div>
            <div class="print-signature-label">{{ t("operator.review.signatures.preparedBy") }}</div>
          </div>
          <div class="print-signature-box">
            <div class="print-signature-line"></div>
            <div class="print-signature-label">{{ t("operator.review.signatures.approvedBy") }}</div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.review-page {
  display: grid;
  gap: var(--space-4);
  max-width: 100%;
}

.controls-row {
  display: flex;
  align-items: flex-end;
  gap: var(--space-3);
  margin-bottom: var(--space-2);
  flex-wrap: wrap;
}

.control-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.control-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
}

.search-group {
  flex-grow: 1;
  max-width: 320px;
}

.empty-hint {
  padding: 16px;
  color: var(--text-secondary);
  font-size: 13px;
}

.list-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  margin-bottom: 12px;
  background: var(--alert-danger-bg, #ffebe9);
  border: 1px solid var(--alert-danger-border, #cf222e);
  border-radius: var(--radius-sm, 4px);
  color: var(--alert-danger-text, #cf222e);
  font-size: 13px;
}

.list-error__retry {
  border: 1px solid currentColor;
  background: transparent;
  color: inherit;
  border-radius: var(--radius-sm, 4px);
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.paging-note {
  margin: 8px 0 0;
  font-size: 11px;
  color: var(--text-tertiary);
}

.pagination-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 12px;
  font-size: 13px;
  color: var(--text-secondary);
}

.pagination-button {
  padding: 4px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  color: var(--text-primary);
  cursor: pointer;
  font-size: 13px;
}

.pagination-button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Split Layout */
.split-layout {
  display: grid;
  grid-template-columns: minmax(0, 3fr) minmax(320px, 2fr);
  gap: var(--space-4);
  align-items: start;
}

.split-layout__left {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  min-width: 0;
}

.split-layout__right {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  min-width: 0;
}

/* Table styling */
.table-container {
  overflow-x: auto;
  border-radius: var(--radius-md);
}

.operator-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.operator-table th,
.operator-table td {
  padding: 8px 12px;
  text-align: left;
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.operator-table th {
  background: var(--bg-card-muted);
  font-size: 11px;
  text-transform: uppercase;
  color: var(--text-tertiary);
  font-weight: 600;
}

.operator-table td.right,
.operator-table th.right {
  text-align: right;
}

.operator-table td.bold {
  font-weight: 600;
}

.clickable-row {
  cursor: pointer;
  transition: background-color 0.1s ease;
}

.clickable-row:hover {
  background: var(--bg-card-hover);
}

.clickable-row:focus-visible {
  outline: 2px solid var(--border-focus, #0969da);
  outline-offset: -2px;
}

.clickable-row.is-active {
  background: var(--bg-card-muted);
  border-left: 3px solid var(--state-info);
}

.preview-paper {
  border: 1px dashed var(--border-default);
  border-radius: 4px;
  padding: 20px;
  background: #ffffff;
  color: #1f2328;
  box-shadow: inset 0 0 8px rgba(0, 0, 0, 0.02);
  font-family: inherit;
}

.paper-title {
  text-align: center;
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 15px;
  letter-spacing: 0.05em;
  border-bottom: 2px solid #1f2328;
  padding-bottom: 6px;
}

.paper-metadata {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  font-size: 11px;
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: 8px;
  margin-bottom: 12px;
}

.paper-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 12px;
}

.preview-actions {
  display: flex;
  gap: 8px;
}

.print-area {
  display: none;
}

@media (max-width: 900px) {
  .split-layout {
    grid-template-columns: 1fr;
  }
}

.print-not-recorded {
  margin: 0;
  font-size: 11pt;
  color: var(--text-tertiary, #6e7781);
  font-style: italic;
}

.print-breach-note {
  margin: 6px 0 0;
  font-size: 10pt;
  font-weight: 700;
  color: var(--state-danger, #cf222e);
}

.bold { font-weight: 600; }
.right { text-align: right; }
.mb-4 { margin-bottom: 16px; }
.mt { margin-top: 12px; }

@media print {
  .no-print,
  header,
  footer,
  aside,
  nav,
  .split-layout__left,
  .split-layout__right,
  button,
  .btn,
  .preview-actions {
    display: none !important;
  }

  .print-area {
    display: block !important;
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    margin: 0;
    padding: 0;
    background: #ffffff !important;
    color: #000000 !important;
    font-family: "Times New Roman", Times, serif;
    font-size: 12pt;
    line-height: 1.5;
  }

  .print-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 3px double #000000;
    padding-bottom: 10px;
    margin-bottom: 20px;
  }

  .print-logo {
    font-size: 24pt;
    font-weight: bold;
    color: #000000;
  }

  .print-title-block {
    text-align: right;
  }

  .print-title {
    font-size: 16pt;
    font-weight: bold;
    margin: 0;
  }

  .print-subtitle {
    font-size: 10pt;
    color: #333333;
    margin-top: 4px;
  }

  .print-section {
    margin-bottom: 20px;
    page-break-inside: avoid;
  }

  .print-section-title {
    font-size: 12pt;
    font-weight: bold;
    text-transform: uppercase;
    border-bottom: 1px solid #000000;
    padding-bottom: 4px;
    margin-bottom: 8px;
    margin-top: 0;
  }

  .print-table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 10px;
  }

  .print-table th,
  .print-table td {
    padding: 6px 8px;
    font-size: 10pt;
    text-align: left;
    vertical-align: top;
    border: 1px solid #000000;
  }

  .print-table th {
    background: #f2f2f2 !important;
    font-weight: bold;
    width: 25%;
  }

  .print-rationale {
    padding: 10px;
    border: 1px solid #000000;
    font-size: 10pt;
    white-space: pre-wrap;
    min-height: 80px;
    background: #ffffff;
  }

  .print-signatures {
    margin-top: 50px;
    display: flex;
    justify-content: space-between;
    page-break-inside: avoid;
  }

  .print-signature-box {
    width: 45%;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .print-signature-line {
    width: 100%;
    border-bottom: 1px solid #000000;
    margin-bottom: 6px;
  }

  .print-signature-label {
    font-size: 10pt;
    text-align: center;
  }
}
</style>
