<script setup lang="ts">
import { computed, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppDescriptionList from "~/shared/ui/AppDescriptionList.vue";
import AppErrorState from "~/shared/ui/AppErrorState.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import { useDecisionDetail } from "../composables/useDecisionApproval";
import { usePortfolioCodeLookup } from "../composables/usePortfolioCodeLookup";

const props = defineProps<{
  open: boolean;
  decisionId: string | null;
}>();

const emit = defineEmits<{
  close: [];
  "view-full-decision": [portfolioCode: string, decisionId: string];
}>();

const { t } = useI18n();
const { decision, loading, error, fetch, clear } = useDecisionDetail();
const codeLookup = usePortfolioCodeLookup();
const resolvedPortfolioCode = ref<string | null>(null);

watch(
  () => props.decisionId,
  async (id) => {
    resolvedPortfolioCode.value = null;
    if (id) {
      await fetch(id);
      resolvedPortfolioCode.value = await codeLookup.resolve(decision.value?.portfolio_id);
    } else {
      clear();
    }
  },
  { immediate: true },
);

function onViewFullDecision() {
  if (!resolvedPortfolioCode.value || !props.decisionId) return;
  emit("view-full-decision", resolvedPortfolioCode.value, props.decisionId);
}

const headerFields = computed(() => {
  if (!decision.value) return [];
  const d = decision.value;
  return [
    {
      label: t("approval.decisionWorkbench.drawer.decisionNumber", "Decision number"),
      value: d.decision_number ?? "—",
      required: true,
    },
    {
      label: t("approval.decisionWorkbench.drawer.businessDate", "Business date"),
      value: d.business_date ?? "—",
      required: true,
    },
    {
      label: t("approval.decisionWorkbench.drawer.processType", "Process type"),
      value: humanize(d.process_type),
    },
    {
      label: t("approval.decisionWorkbench.drawer.decisionType", "Decision type"),
      value: humanize(d.decision_type),
    },
    {
      label: t("approval.decisionWorkbench.drawer.productType", "Product type"),
      value: humanize(d.product_type),
    },
    {
      label: t("approval.decisionWorkbench.drawer.status", "Status"),
      value: statusLabel(d.approval_status ?? d.status),
      required: true,
    },
    {
      label: t("approval.decisionWorkbench.drawer.side", "Side"),
      value: d.side ?? t("approval.decisionWorkbench.table.basket", "Basket"),
    },
    {
      label: t("approval.decisionWorkbench.drawer.instrument", "Instrument"),
      value: d.instrument_code ?? "—",
    },
    {
      label: t("approval.decisionWorkbench.drawer.quantity", "Quantity"),
      value: d.quantity ?? "—",
    },
    {
      label: t("approval.decisionWorkbench.drawer.amount", "Amount"),
      value: d.amount ?? "—",
    },
    {
      label: t("approval.decisionWorkbench.drawer.currency", "Currency"),
      value: d.currency ?? "—",
    },
    {
      label: t("approval.decisionWorkbench.drawer.strategy", "Strategy"),
      value: d.strategy_code ?? "—",
    },
    {
      label: t("approval.decisionWorkbench.drawer.researchReport", "Research report"),
      value: d.research_report_no ?? "—",
    },
    {
      label: t("approval.decisionWorkbench.drawer.rationale", "Rationale"),
      value: d.rationale ?? "—",
    },
  ]
    .filter((field) => field.required || field.value !== "—")
    .map(({ label, value }) => ({ label, value }));
});

const approvalFields = computed(() => {
  if (!decision.value) return [];
  const d = decision.value;
  const fields: { label: string; value: string }[] = [];

  if (d.approval_stage) {
    fields.push({
      label: t("approval.decisionWorkbench.drawer.approvalStage", "Approval stage"),
      value: d.approval_total_stages
        ? `${d.approval_stage} / ${d.approval_total_stages}`
        : String(d.approval_stage),
    });
  }
  if (d.current_approvers?.length) {
    fields.push({
      label: t("approval.decisionWorkbench.drawer.pendingApprovers", "Pending approvers"),
      value: d.current_approvers.join(", "),
    });
  }
  if (d.previous_approvers?.length) {
    fields.push({
      label: t("approval.decisionWorkbench.drawer.previousApprovers", "Previous approvers"),
      value: d.previous_approvers.join(", "),
    });
  }
  return fields;
});

function humanize(value: string | undefined): string {
  return value ? value.replace(/_/g, " ") : "—";
}

function statusType(status: string | undefined): string {
  switch (status) {
    case "PENDING":
    case "PENDING_APPROVAL":
      return "pending";
    case "APPROVED":
      return "approved";
    case "REJECTED":
      return "rejected";
    case "CANCELLED":
      return "inactive";
    case "READY_FOR_EXECUTION":
      return "success";
    default:
      return "neutral";
  }
}

function statusLabel(status: string | undefined): string {
  switch (status) {
    case "PENDING":
    case "PENDING_APPROVAL":
      return t("approval.decisionWorkbench.status.pendingApproval", "Pending approval");
    case "APPROVED":
      return t("approval.decisionWorkbench.status.approved", "Approved");
    case "REJECTED":
      return t("approval.decisionWorkbench.status.rejected", "Rejected");
    case "CANCELLED":
      return t("approval.decisionWorkbench.status.cancelled", "Cancelled");
    case "READY_FOR_EXECUTION":
      return t("approval.decisionWorkbench.status.readyForExecution", "Ready for execution");
    default:
      return humanize(status);
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="open" class="drawer-overlay" @click.self="emit('close')">
        <aside
          class="drawer"
          role="dialog"
          aria-modal="true"
          :aria-label="t('approval.decisionWorkbench.drawer.ariaLabel', 'Decision approval details')"
          @keydown.esc="emit('close')"
        >
          <header class="drawer__header">
            <div class="drawer__title-row">
              <div>
                <p class="drawer__eyebrow">
                  {{ t("approval.decisionWorkbench.drawer.eyebrow", "Approval inspection") }}
                </p>
                <h2 class="drawer__title">
                  {{ decision?.decision_number || t("approval.decisionWorkbench.drawer.title", "Decision details") }}
                </h2>
              </div>
              <AppStatusBadge
                v-if="decision?.status || decision?.approval_status"
                :status="statusType(decision.approval_status ?? decision.status)"
                :label="statusLabel(decision.approval_status ?? decision.status)"
                size="sm"
              />
            </div>
            <AppButton
              variant="ghost"
              size="sm"
              icon
              :aria-label="t('approval.decisionWorkbench.drawer.close', 'Close details')"
              @click="emit('close')"
            >
              ×
            </AppButton>
          </header>

          <div class="drawer__body">
            <AppLoadingState
              v-if="loading"
              :message="t('approval.decisionWorkbench.drawer.loading', 'Loading decision details…')"
            />
            <AppErrorState
              v-else-if="error"
              :title="t('approval.decisionWorkbench.drawer.errorTitle', 'Could not load decision')"
              :message="error"
            />
            <template v-else-if="decision">
              <section class="drawer__section">
                <div class="drawer__section-row">
                  <h3 class="drawer__section-title">
                    {{ t("approval.decisionWorkbench.drawer.orderHeader", "Decision header") }}
                  </h3>
                  <AppButton
                    variant="secondary"
                    size="xs"
                    :disabled="!resolvedPortfolioCode"
                    @click="onViewFullDecision"
                  >
                    {{ t("approval.decisionWorkbench.drawer.viewFullDecision", "View full decision") }}
                  </AppButton>
                </div>
                <AppDescriptionList :items="headerFields" />
              </section>

              <section class="drawer__section drawer__section--approval">
                <h3 class="drawer__section-title">
                  {{ t("approval.decisionWorkbench.drawer.approvalRoute", "Approval route") }}
                </h3>
                <AppDescriptionList v-if="approvalFields.length" :items="approvalFields" />
                <p v-else class="drawer__empty">
                  {{
                    t(
                      "approval.decisionWorkbench.drawer.noApprovalRoute",
                      "No approval-stage details are available for this decision.",
                    )
                  }}
                </p>
              </section>

              <section v-if="decision.lines?.length" class="drawer__section">
                <h3 class="drawer__section-title">
                  {{
                    t(
                      "approval.decisionWorkbench.drawer.lines",
                      { count: decision.lines.length },
                      "Order lines ({count})",
                    )
                  }}
                </h3>
                <div class="drawer__lines">
                  <table class="drawer__lines-table">
                    <thead>
                      <tr>
                        <th>#</th>
                        <th>{{ t("approval.decisionWorkbench.drawer.instrument", "Instrument") }}</th>
                        <th>{{ t("approval.decisionWorkbench.drawer.productType", "Product") }}</th>
                        <th>{{ t("approval.decisionWorkbench.drawer.side", "Side") }}</th>
                        <th class="right">{{ t("approval.decisionWorkbench.drawer.quantity", "Quantity") }}</th>
                        <th class="right">{{ t("approval.decisionWorkbench.drawer.amount", "Amount") }}</th>
                        <th class="right">{{ t("approval.decisionWorkbench.drawer.weight", "Weight %") }}</th>
                        <th>{{ t("approval.decisionWorkbench.drawer.currency", "Currency") }}</th>
                        <th>{{ t("approval.decisionWorkbench.drawer.notes", "Notes") }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="line in decision.lines" :key="line.id ?? line.line_number">
                        <td>{{ line.line_number }}</td>
                        <td>{{ line.instrument_code ?? "—" }}</td>
                        <td>{{ humanize(line.product_type) }}</td>
                        <td>{{ humanize(line.side) }}</td>
                        <td class="right">{{ line.quantity ?? "—" }}</td>
                        <td class="right">{{ line.amount ?? "—" }}</td>
                        <td class="right">{{ line.target_weight ? `${line.target_weight}%` : "—" }}</td>
                        <td>{{ line.currency ?? "—" }}</td>
                        <td class="notes">{{ line.notes ?? "" }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </section>
            </template>
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.drawer-overlay {
  position: fixed;
  inset: 0;
  z-index: 900;
  display: flex;
  justify-content: flex-end;
  background: rgba(15, 23, 42, 0.42);
}

.drawer {
  display: flex;
  flex-direction: column;
  width: min(760px, 92vw);
  height: 100%;
  overflow: hidden;
  border-left: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card, #fff);
  box-shadow: -18px 0 48px rgba(15, 23, 42, 0.16);
}

.drawer__header {
  display: flex;
  flex-shrink: 0;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3, 12px);
  padding: var(--space-4, 16px) var(--space-5, 20px);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.drawer__title-row {
  display: flex;
  align-items: center;
  gap: var(--space-3, 12px);
  min-width: 0;
}

.drawer__eyebrow {
  margin: 0 0 2px;
  color: var(--text-tertiary, #6e7781);
  font-size: 10px;
  font-weight: var(--font-weight-semibold, 600);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.drawer__title {
  overflow: hidden;
  margin: 0;
  color: var(--text-primary, #1f2328);
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-lg, 18px);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.drawer__body {
  display: grid;
  flex: 1;
  gap: var(--space-6, 24px);
  overflow-y: auto;
  padding: var(--space-5, 20px);
  align-content: start;
}

.drawer__section--approval {
  padding: var(--space-4, 16px);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 8px);
  background: var(--bg-card-muted, #f6f8fa);
}

.drawer__section-title {
  margin: 0 0 var(--space-3, 12px);
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.drawer__section-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3, 12px);
}

.drawer__section-row .drawer__section-title {
  margin-bottom: 0;
}

.drawer__empty {
  margin: 0;
  color: var(--text-tertiary, #6e7781);
  font-size: var(--font-size-sm, 14px);
}

.drawer__lines {
  overflow-x: auto;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
}

.drawer__lines-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-xs, 12px);
}

.drawer__lines-table th,
.drawer__lines-table td {
  padding: var(--space-2, 8px) var(--space-3, 12px);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  text-align: left;
  white-space: nowrap;
}

.drawer__lines-table th {
  background: var(--bg-card-muted, #f6f8fa);
  color: var(--text-secondary, #57606a);
  font-weight: var(--font-weight-semibold, 600);
}

.drawer__lines-table tbody tr:last-child td {
  border-bottom: 0;
}

.drawer__lines-table .right {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.drawer__lines-table .notes {
  max-width: 220px;
  white-space: normal;
  word-break: break-word;
}

.drawer-enter-active,
.drawer-leave-active {
  transition: opacity 0.18s ease;
}

.drawer-enter-active .drawer,
.drawer-leave-active .drawer {
  transition: transform 0.22s ease;
}

.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}

.drawer-enter-from .drawer,
.drawer-leave-to .drawer {
  transform: translateX(100%);
}

@media (prefers-reduced-motion: reduce) {
  .drawer-enter-active,
  .drawer-leave-active,
  .drawer-enter-active .drawer,
  .drawer-leave-active .drawer {
    transition: none;
  }
}

@media (max-width: 640px) {
  .drawer {
    width: 100vw;
  }

  .drawer__header,
  .drawer__body {
    padding-right: var(--space-4, 16px);
    padding-left: var(--space-4, 16px);
  }

  .drawer__title-row {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
