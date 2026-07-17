<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useDecisionList } from "~/features/investment-decision/composables/useDecisionList";
import { decisionApi } from "~/features/investment-decision/services/decisionApi";
import { investmentLedgerApi } from "~/features/investment-ledger/services/investmentLedgerApi";
import { myFundsApi } from "~/features/my-funds/services/myFundsApi";
import { useI18n } from "~/composables/useI18n";
import { decisionStatusKey } from "~/features/portfolio-decision/lib/decisionFormat";

import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppBadge from "~/shared/ui/AppBadge.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import type { components } from "~/api/ims-api";

type ApiDecision = components["schemas"]["DecisionResponse"];

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const fundId = computed(() => {
  const value = route.params.fundId;
  return typeof value === "string" ? value : "";
});

// Review List & Print State
const reviewList = useDecisionList();
const reviewSearch = ref("");
const reviewStatusFilter = ref<string>("ALL");
const selectedReviewDecision = ref<ApiDecision | null>(null);
const activeReviewLoading = ref(false);
const reviewError = ref<string | null>(null);
const selectedFundLabel = ref("");
const selectedPortfolioLabel = ref("");

function joinBusinessLabel(code?: string, name?: string): string {
  return [code, name].filter((value): value is string => Boolean(value?.trim())).join(" — ");
}

function statusLabel(status: string | undefined): string {
  const key = decisionStatusKey(status);
  return key ? t(key) : status || t("common.notAvailable");
}

function sideLabel(side: string | undefined): string {
  if (side === "BUY") return t("portfolio.decisionNew.buy");
  if (side === "SELL") return t("portfolio.decisionNew.sell");
  return side || t("common.notAvailable");
}

async function loadSelectedBusinessLabels(decision: ApiDecision) {
  selectedFundLabel.value = "";
  selectedPortfolioLabel.value = "";
  await Promise.all([
    decision.portfolio_id
      ? investmentLedgerApi
          .getPortfolio(decision.portfolio_id)
          .then((portfolio) => {
            selectedPortfolioLabel.value = joinBusinessLabel(portfolio.code, portfolio.name);
          })
          .catch(() => undefined)
      : Promise.resolve(),
    decision.fund_id
      ? myFundsApi
          .listMyFunds()
          .then((funds) => {
            const fund = funds.find((item) => item.id === decision.fund_id);
            selectedFundLabel.value = joinBusinessLabel(fund?.code, fund?.name);
          })
          .catch(() => undefined)
      : Promise.resolve(),
  ]);
}

async function loadReviewDecisions() {
  if (!fundId.value) return;
  reviewList.page.value = 1;
  await reviewList.fetchList({
    fund_id: fundId.value,
    status: reviewStatusFilter.value === "ALL" ? undefined : reviewStatusFilter.value,
    search: reviewSearch.value || undefined,
  });
  // Clear selection
  selectedReviewDecision.value = null;
  selectedFundLabel.value = "";
  selectedPortfolioLabel.value = "";
}

async function selectReviewDecision(decision: ApiDecision) {
  if (!decision.id) return;
  activeReviewLoading.value = true;
  reviewError.value = null;
  try {
    const detail = await decisionApi.getDecisionDetail(decision.id);
    selectedReviewDecision.value = detail;
    await loadSelectedBusinessLabels(detail);
  } catch (err) {
    reviewError.value = err instanceof Error ? err.message : t("operator.review.loadError");
  } finally {
    activeReviewLoading.value = false;
  }
}

function triggerPrint() {
  window.print();
}

watch(fundId, () => {
  void loadReviewDecisions();
}, { immediate: true });
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
        <!-- Left Column: List -->
        <div class="split-layout__left">
          <div class="controls-row">
            <div class="control-group">
              <label class="control-label">{{ t("operator.review.statusFilter") }}</label>
              <select v-model="reviewStatusFilter" class="control-select" @change="loadReviewDecisions">
                <option value="ALL">{{ t("operator.review.allDecisions") }}</option>
                <option value="APPROVED">{{ t("operator.review.approved") }}</option>
                <option value="READY_FOR_EXECUTION">{{ t("operator.review.readyForExecution") }}</option>
                <option value="CANCELLED">{{ t("operator.review.cancelled") }}</option>
              </select>
            </div>
            <div class="control-group search-group">
              <label class="control-label">{{ t("operator.review.search") }}</label>
              <input
                v-model="reviewSearch"
                type="text"
                :placeholder="t('operator.review.searchPlaceholder')"
                class="control-input"
                @keydown.enter="loadReviewDecisions"
              />
            </div>
            <AppButton variant="secondary" size="sm" @click="loadReviewDecisions">{{ t("operator.actions.search") }}</AppButton>
          </div>

          <AppCard class="mt">
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
                  <tr v-if="reviewList.loading.value">
                    <td colspan="6" class="text-center p-4">{{ t("operator.review.loading") }}</td>
                  </tr>
                  <tr v-else-if="reviewList.items.value.length === 0">
                    <td colspan="6" class="text-center p-4 text-secondary">{{ t("operator.review.empty") }}</td>
                  </tr>
                  <tr
                    v-for="item in reviewList.items.value"
                    :key="item.id"
                    class="clickable-row"
                    :class="{ 'is-active': selectedReviewDecision?.id === item.id }"
                    @click="selectReviewDecision(item)"
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
          </AppCard>
        </div>

        <!-- Right Column: Print Sheet Preview -->
        <div class="split-layout__right">
          <AppCard :title="t('operator.review.previewTitle')" :subtitle="t('operator.review.previewSubtitle')">
            <div v-if="activeReviewLoading" class="p-6 text-center">
              <AppLoadingState :message="t('operator.review.previewLoading')" />
            </div>
            <div v-else-if="!selectedReviewDecision" class="p-6 text-center text-secondary">
              {{ t("operator.review.previewEmpty") }}
            </div>
            <div v-else class="review-preview-card">
              <div class="preview-actions mb-4">
                <AppButton variant="primary" size="sm" @click="triggerPrint">
                  {{ t("operator.actions.print") }}
                </AppButton>
              </div>

              <!-- Dotted border preview card -->
              <div class="preview-paper">
                <div class="paper-title">{{ t("operator.review.sheetTitle") }}</div>
                <div class="paper-metadata">
                  <div><strong>{{ t("operator.review.number") }}:</strong> {{ selectedReviewDecision.decision_number }}</div>
                  <div><strong>{{ t("operator.review.date") }}:</strong> {{ selectedReviewDecision.business_date }}</div>
                  <div><strong>{{ t("operator.review.status") }}:</strong> {{ statusLabel(selectedReviewDecision.status) }}</div>
                </div>
                <div class="paper-body">
                  <div class="paper-field"><strong>{{ t("operator.review.instrument") }}:</strong> {{ selectedReviewDecision.instrument_code }}</div>
                  <div class="paper-field"><strong>{{ t("operator.review.side") }}:</strong> {{ sideLabel(selectedReviewDecision.side) }}</div>
                  <div class="paper-field"><strong>{{ t("operator.review.quantityPrice") }}:</strong> {{ selectedReviewDecision.quantity }} @ {{ selectedReviewDecision.limit_price }} {{ selectedReviewDecision.currency }}</div>
                  <div class="paper-field"><strong>{{ t("operator.review.rationale") }}:</strong></div>
                  <div class="paper-text">{{ selectedReviewDecision.rationale || t("operator.review.noRationale") }}</div>
                </div>
              </div>
            </div>
          </AppCard>
        </div>
      </div>

      <!-- Printable Document Element (Always rendered, only visible during Print window) -->
      <div v-if="selectedReviewDecision" class="print-area">
        <div class="print-header">
          <div class="print-logo">TH-IMS</div>
          <div class="print-title-block">
            <h1 class="print-title">{{ t("operator.review.reportTitle") }}</h1>
            <div class="print-subtitle">{{ t("operator.review.referenceNumber") }}: {{ selectedReviewDecision.decision_number || t("common.notAvailable") }}</div>
          </div>
        </div>

        <div class="print-section">
          <h2 class="print-section-title">{{ t("operator.review.sections.metadata") }}</h2>
          <table class="print-table">
            <tbody>
              <tr>
                <th>{{ t("operator.review.fields.fund") }}</th>
                <td>{{ selectedFundLabel || t("operator.operation.unavailableBusinessLabel") }}</td>
                <th>{{ t("operator.review.fields.portfolio") }}</th>
                <td>{{ selectedPortfolioLabel || t("operator.operation.unavailableBusinessLabel") }}</td>
              </tr>
              <tr>
                <th>{{ t("operator.review.fields.businessDate") }}</th>
                <td>{{ selectedReviewDecision.business_date ? selectedReviewDecision.business_date.slice(0, 10) : '—' }}</td>
                <th>{{ t("operator.review.fields.lifecycleStatus") }}</th>
                <td>{{ statusLabel(selectedReviewDecision.status) }}</td>
              </tr>
              <tr>
                <th>{{ t("operator.review.fields.submittedAt") }}</th>
                <td>{{ selectedReviewDecision.submitted_at || '—' }}</td>
                <th>{{ t("operator.review.fields.createdAt") }}</th>
                <td>{{ selectedReviewDecision.created_at || '—' }}</td>
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
                <td>{{ selectedReviewDecision.instrument_code }}</td>
                <th>{{ t("operator.review.fields.exchange") }}</th>
                <td>{{ selectedReviewDecision.exchange || t("common.notAvailable") }}</td>
              </tr>
              <tr>
                <th>{{ t("operator.review.side") }}</th>
                <td><strong>{{ sideLabel(selectedReviewDecision.side) }}</strong></td>
                <th>{{ t("operator.review.fields.currency") }}</th>
                <td>{{ selectedReviewDecision.currency }}</td>
              </tr>
              <tr>
                <th>{{ t("operator.review.fields.quantity") }}</th>
                <td>{{ selectedReviewDecision.quantity || '—' }}</td>
                <th>{{ t("operator.review.fields.limitPrice") }}</th>
                <td>{{ selectedReviewDecision.limit_price || '—' }}</td>
              </tr>
              <tr>
                <th>{{ t("operator.review.fields.grossAmount") }}</th>
                <td colspan="3">{{ selectedReviewDecision.amount || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-if="selectedReviewDecision.lines && selectedReviewDecision.lines.length > 0" class="print-section">
          <h2 class="print-section-title">{{ t("operator.review.sections.lines") }}</h2>
          <table class="print-table print-table--striped">
            <thead>
              <tr>
                <th>{{ t("operator.review.fields.lineNumber") }}</th>
                <th>{{ t("operator.review.instrument") }}</th>
                <th>{{ t("operator.review.side") }}</th>
                <th class="right">{{ t("operator.review.fields.quantity") }}</th>
                <th class="right">{{ t("operator.review.fields.price") }}</th>
                <th>{{ t("operator.review.fields.currency") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="line in selectedReviewDecision.lines" :key="line.id || line.line_number">
                <td>{{ line.line_number }}</td>
                <td>{{ line.instrument_code || '—' }}</td>
                <td>{{ sideLabel(line.side) }}</td>
                <td class="right">{{ line.quantity || '—' }}</td>
                <td class="right">{{ line.amount || '—' }}</td>
                <td>{{ line.currency || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="print-section">
          <h2 class="print-section-title">{{ t("operator.review.sections.rationale") }}</h2>
          <div class="print-rationale">
            {{ selectedReviewDecision.rationale || t("operator.review.noRationale") }}
          </div>
        </div>

        <div class="print-section">
          <h2 class="print-section-title">{{ t("operator.review.sections.approval") }}</h2>
          <table class="print-table">
            <tbody>
              <tr>
                <th>{{ t("operator.review.fields.currentStage") }}</th>
                <td>{{ t("operator.review.fields.stageValue", { current: selectedReviewDecision.approval_stage || t("common.notAvailable"), total: selectedReviewDecision.approval_total_stages || t("common.notAvailable") }) }}</td>
              </tr>
              <tr v-if="selectedReviewDecision.previous_approvers && selectedReviewDecision.previous_approvers.length > 0">
                <th>{{ t("operator.review.fields.previousApprovers") }}</th>
                <td>{{ selectedReviewDecision.previous_approvers.join(', ') }}</td>
              </tr>
              <tr v-if="selectedReviewDecision.current_approvers && selectedReviewDecision.current_approvers.length > 0">
                <th>{{ t("operator.review.fields.pendingApprovers") }}</th>
                <td>{{ selectedReviewDecision.current_approvers.join(', ') }}</td>
              </tr>
            </tbody>
          </table>
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

.control-select,
.control-input {
  height: 32px;
  padding: 0 10px;
  font-size: 13px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  background: var(--bg-input);
  color: var(--text-primary);
}

.control-select {
  min-width: 150px;
}

.search-group {
  flex-grow: 1;
  max-width: 320px;
}

.control-input {
  width: 100%;
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

.clickable-row.is-active {
  background: var(--bg-card-muted);
  border-left: 3px solid var(--state-info);
}

/* Preview Paper container */
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

.paper-text {
  background: #f6f8fa;
  border-radius: 4px;
  padding: 8px;
  white-space: pre-line;
  font-style: italic;
}

/* Print container */
.print-area {
  display: none;
}

/* Helpers */
.bold { font-weight: 600; }
.right { text-align: right; }
.mb-4 { margin-bottom: 16px; }
.mt { margin-top: 12px; }

/* =========================================================
   PRINT MEDIA STYLES
   ========================================================= */

@media print {
  /* Hide all dashboard chrome, selectors, tabs, headers, side columns */
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

  /* Make print area visible and expand to page */
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

  .print-table td.right,
  .print-table th.right {
    text-align: right;
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
