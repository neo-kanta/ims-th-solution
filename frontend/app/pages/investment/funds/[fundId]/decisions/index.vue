<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "~/composables/useI18n";
import { FundDetailLayout } from "~/features/my-funds";
import { useDecisionList } from "~/features/investment-decision/composables/useDecisionList";
import { decisionApi } from "~/features/investment-decision/services/decisionApi";

import AppCard from "~/shared/ui/AppCard.vue";
import AppBadge from "~/shared/ui/AppBadge.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

import type { components } from "~/api/ims-api";

type ApiDecision = components["schemas"]["DecisionResponse"];

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const { t } = useI18n();
const route = useRoute();
const fundId = computed(() => String(route.params.fundId ?? ""));

// Review List & Print State
const reviewList = useDecisionList();
const reviewSearch = ref("");
const reviewStatusFilter = ref<string>("ALL");
const selectedReviewDecision = ref<ApiDecision | null>(null);
const activeReviewLoading = ref(false);
const reviewError = ref<string | null>(null);

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
}

async function selectReviewDecision(decision: ApiDecision) {
  if (!decision.id) return;
  activeReviewLoading.value = true;
  reviewError.value = null;
  try {
    selectedReviewDecision.value = await decisionApi.getDecisionDetail(decision.id);
  } catch (err: any) {
    reviewError.value = err.message || "Failed to load decision details for printing";
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
  <FundDetailLayout :fund-id="fundId" active-tab="decisions">
    <template #default>
      <section class="review-page">
        <div class="split-layout no-print">
          <!-- Left Column: List -->
          <div class="split-layout__left">
            <div class="controls-row">
              <div class="control-group">
                <label class="control-label">Status Filter</label>
                <select v-model="reviewStatusFilter" class="control-select" @change="loadReviewDecisions">
                  <option value="ALL">All Decisions</option>
                  <option value="APPROVED">Approved</option>
                  <option value="READY_FOR_EXECUTION">Ready for Execution</option>
                  <option value="CANCELLED">Cancelled</option>
                </select>
              </div>
              <div class="control-group search-group">
                <label class="control-label">Search</label>
                <input
                  v-model="reviewSearch"
                  type="text"
                  placeholder="Search decision no..."
                  class="control-input"
                  @keydown.enter="loadReviewDecisions"
                />
              </div>
              <AppButton variant="secondary" size="sm" @click="loadReviewDecisions">Search</AppButton>
            </div>

            <AppCard class="mt">
              <div class="table-container">
                <table class="operator-table">
                  <thead>
                    <tr>
                      <th>Decision No</th>
                      <th>Side</th>
                      <th>Instrument</th>
                      <th class="right">Qty</th>
                      <th>Status</th>
                      <th>Date</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-if="reviewList.loading.value">
                      <td colspan="6" class="text-center p-4">Loading decisions…</td>
                    </tr>
                    <tr v-else-if="reviewList.items.value.length === 0">
                      <td colspan="6" class="text-center p-4 text-secondary">No processed decisions found.</td>
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
                        <AppBadge :variant="item.side === 'BUY' ? 'success' : 'error'">{{ item.side }}</AppBadge>
                      </td>
                      <td class="bold">{{ item.instrument_code || '—' }}</td>
                      <td class="right">{{ item.quantity || '—' }}</td>
                      <td>
                        <AppBadge :variant="item.status === 'APPROVED' || item.status === 'READY_FOR_EXECUTION' ? 'success' : 'neutral'">
                          {{ item.status }}
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
            <AppCard title="Summary Sheet Preview" subtitle="Review decision details and execute printing">
              <div v-if="activeReviewLoading" class="p-6 text-center">
                <AppLoadingState message="Loading summary sheet data..." />
              </div>
              <div v-else-if="!selectedReviewDecision" class="p-6 text-center text-secondary">
                Select a decision from the left list to generate and print its formal summary report.
              </div>
              <div v-else class="review-preview-card">
                <div class="preview-actions mb-4">
                  <AppButton variant="primary" size="sm" @click="triggerPrint">
                    🖨 Print Summary Sheet
                  </AppButton>
                </div>

                <!-- Dotted border preview card -->
                <div class="preview-paper">
                  <div class="paper-title">DECISION SHEET SUMMARY</div>
                  <div class="paper-metadata">
                    <div><strong>No:</strong> {{ selectedReviewDecision.decision_number }}</div>
                    <div><strong>Date:</strong> {{ selectedReviewDecision.business_date }}</div>
                    <div><strong>Status:</strong> {{ selectedReviewDecision.status }}</div>
                  </div>
                  <div class="paper-body">
                    <div class="paper-field"><strong>Instrument:</strong> {{ selectedReviewDecision.instrument_code }}</div>
                    <div class="paper-field"><strong>Side:</strong> {{ selectedReviewDecision.side }}</div>
                    <div class="paper-field"><strong>Qty / Price:</strong> {{ selectedReviewDecision.quantity }} @ {{ selectedReviewDecision.limit_price }} {{ selectedReviewDecision.currency }}</div>
                    <div class="paper-field"><strong>Rationale:</strong></div>
                    <div class="paper-text">{{ selectedReviewDecision.rationale || '—' }}</div>
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
              <h1 class="print-title">INVESTMENT DECISION SUMMARY REPORT</h1>
              <div class="print-subtitle">Reference No: {{ selectedReviewDecision.decision_number || selectedReviewDecision.id }}</div>
            </div>
          </div>

          <div class="print-section">
            <h2 class="print-section-title">1. GENERAL METADATA</h2>
            <table class="print-table">
              <tbody>
                <tr>
                  <th>Fund ID</th>
                  <td>{{ selectedReviewDecision.fund_id }}</td>
                  <th>Portfolio ID</th>
                  <td>{{ selectedReviewDecision.portfolio_id }}</td>
                </tr>
                <tr>
                  <th>Business Date</th>
                  <td>{{ selectedReviewDecision.business_date ? selectedReviewDecision.business_date.slice(0, 10) : '—' }}</td>
                  <th>Lifecycle Status</th>
                  <td>{{ selectedReviewDecision.status }}</td>
                </tr>
                <tr>
                  <th>Submitted At</th>
                  <td>{{ selectedReviewDecision.submitted_at || '—' }}</td>
                  <th>Created At</th>
                  <td>{{ selectedReviewDecision.created_at || '—' }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="print-section">
            <h2 class="print-section-title">2. TRANSACTION DETAILS</h2>
            <table class="print-table">
              <tbody>
                <tr>
                  <th>Instrument Code</th>
                  <td>{{ selectedReviewDecision.instrument_code }}</td>
                  <th>Exchange</th>
                  <td>{{ selectedReviewDecision.exchange || 'SET' }}</td>
                </tr>
                <tr>
                  <th>Side</th>
                  <td><strong>{{ selectedReviewDecision.side }}</strong></td>
                  <th>Currency</th>
                  <td>{{ selectedReviewDecision.currency }}</td>
                </tr>
                <tr>
                  <th>Quantity</th>
                  <td>{{ selectedReviewDecision.quantity || '—' }}</td>
                  <th>Limit Price</th>
                  <td>{{ selectedReviewDecision.limit_price || '—' }}</td>
                </tr>
                <tr>
                  <th>Gross Amount</th>
                  <td colspan="3">{{ selectedReviewDecision.amount || '—' }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="selectedReviewDecision.lines && selectedReviewDecision.lines.length > 0" class="print-section">
            <h2 class="print-section-title">3. DECISION LINES</h2>
            <table class="print-table print-table--striped">
              <thead>
                <tr>
                  <th>Line No</th>
                  <th>Instrument</th>
                  <th>Side</th>
                  <th class="right">Qty</th>
                  <th class="right">Price</th>
                  <th>CCY</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="line in selectedReviewDecision.lines" :key="line.id || line.line_number">
                  <td>{{ line.line_number }}</td>
                  <td>{{ line.instrument_code || '—' }}</td>
                  <td>{{ line.side || '—' }}</td>
                  <td class="right">{{ line.quantity || '—' }}</td>
                  <td class="right">{{ line.amount || '—' }}</td>
                  <td>{{ line.currency || '—' }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="print-section">
            <h2 class="print-section-title">4. INVESTMENT RATIONALE</h2>
            <div class="print-rationale">
              {{ selectedReviewDecision.rationale || 'No rationale provided.' }}
            </div>
          </div>

          <div class="print-section">
            <h2 class="print-section-title">5. APPROVAL TIMELINE</h2>
            <table class="print-table">
              <tbody>
                <tr>
                  <th>Current Stage</th>
                  <td>Stage {{ selectedReviewDecision.approval_stage || '—' }} / {{ selectedReviewDecision.approval_total_stages || '—' }}</td>
                </tr>
                <tr v-if="selectedReviewDecision.previous_approvers && selectedReviewDecision.previous_approvers.length > 0">
                  <th>Previous Approvers</th>
                  <td>{{ selectedReviewDecision.previous_approvers.join(', ') }}</td>
                </tr>
                <tr v-if="selectedReviewDecision.current_approvers && selectedReviewDecision.current_approvers.length > 0">
                  <th>Pending Approvers</th>
                  <td>{{ selectedReviewDecision.current_approvers.join(', ') }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="print-signatures">
            <div class="print-signature-box">
              <div class="print-signature-line"></div>
              <div class="print-signature-label">Prepared By (Fund Manager)</div>
            </div>
            <div class="print-signature-box">
              <div class="print-signature-line"></div>
              <div class="print-signature-label">Approved By (Compliance Officer)</div>
            </div>
          </div>
        </div>
      </section>
    </template>
  </FundDetailLayout>
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
