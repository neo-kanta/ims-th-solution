<script setup lang="ts">
import { computed, watch } from "vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";
import AppErrorState from "~/shared/ui/AppErrorState.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import AppDescriptionList from "~/shared/ui/AppDescriptionList.vue";
import { useDecisionDetail } from "../composables/useDecisionApproval";
import type { ApiDecision } from "../services/decisionApprovalApi";
import ApprovalSessionPanel from "~/features/approval/components/ApprovalSessionPanel.vue";

interface Props {
  open: boolean;
  decisionId: string | null;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  close: [];
}>();

const { decision, loading, error, fetch, clear } = useDecisionDetail();

watch(
  () => props.decisionId,
  (id) => {
    if (id) fetch(id);
    else clear();
  },
  { immediate: true },
);

const headerFields = computed(() => {
  if (!decision.value) return [];
  const d = decision.value;
  return [
    { label: "Decision No", value: d.decision_number ?? "—" },
    { label: "Business Date", value: d.business_date ?? "—" },
    { label: "Process Type", value: (d.process_type ?? "—").replace(/_/g, " ") },
    { label: "Decision Type", value: (d.decision_type ?? "—").replace(/_/g, " ") },
    { label: "Product Type", value: d.product_type ?? "—" },
    { label: "Status", value: (d.status ?? "—").replace(/_/g, " ") },
    { label: "Side", value: d.side ?? "BASKET" },
    { label: "Instrument", value: d.instrument_code ?? "—" },
    { label: "Quantity", value: d.quantity ?? "—" },
    { label: "Amount", value: d.amount ?? "—" },
    { label: "Currency", value: d.currency ?? "—" },
    { label: "Strategy", value: d.strategy_code ?? "—" },
    { label: "Research Report", value: d.research_report_no ?? "—" },
    { label: "Rationale", value: d.rationale ?? "—" },
  ].filter((f) => f.value !== "—" || ["Decision No", "Business Date", "Status"].includes(f.label));
});

const approvalFields = computed(() => {
  if (!decision.value) return [];
  const d = decision.value;
  const fields = [];
  if (d.approval_stage) {
    fields.push({
      label: "Approval Stage",
      value: d.approval_total_stages
        ? `${d.approval_stage} / ${d.approval_total_stages}`
        : String(d.approval_stage),
    });
  }
  if (d.current_approvers?.length) {
    fields.push({ label: "Pending Approvers", value: d.current_approvers.join(", ") });
  }
  if (d.previous_approvers?.length) {
    fields.push({ label: "Previous Approvers", value: d.previous_approvers.join(", ") });
  }
  return fields;
});

function statusType(status: string | undefined): string {
  switch (status) {
    case "PENDING_APPROVAL": return "pending";
    case "APPROVED": return "approved";
    case "REJECTED": return "rejected";
    case "CANCELLED": return "inactive";
    case "READY_FOR_EXECUTION": return "success";
    default: return "neutral";
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="open" class="drawer-overlay" @click.self="emit('close')">
        <aside class="drawer" role="dialog" aria-modal="true" aria-label="Decision Detail">
          <header class="drawer__header">
            <div class="drawer__title-row">
              <h2 class="drawer__title">Decision Detail</h2>
              <AppStatusBadge
                v-if="decision?.status"
                :status="statusType(decision.status)"
                :label="(decision.status ?? '').replace(/_/g, ' ')"
                size="sm"
              />
            </div>
            <AppButton variant="ghost" size="sm" icon @click="emit('close')">✕</AppButton>
          </header>

          <div class="drawer__body">
            <AppLoadingState v-if="loading" />
            <AppErrorState v-else-if="error" :message="error" />
            <template v-else-if="decision">
              <section class="drawer__section">
                <h3 class="drawer__section-title">Header</h3>
                <AppDescriptionList :items="headerFields" />
              </section>

              <section class="drawer__section">
                <h3 class="drawer__section-title">Approval Session</h3>
                <ApprovalSessionPanel
                  v-if="decision?.id"
                  :target="{
                    moduleCode: 'INVESTMENT',
                    processType: decision.process_type || 'INVESTMENT_DECISION',
                    recordType: 'INVESTMENT_DECISION',
                    recordId: decision.id,
                    title: decision.decision_number
                  }"
                  @action-completed="fetch(decision.id)"
                />
              </section>

              <section v-if="decision.lines && decision.lines.length > 0" class="drawer__section">
                <h3 class="drawer__section-title">Lines ({{ decision.lines.length }})</h3>
                <div class="drawer__lines">
                  <table class="drawer__lines-table">
                    <thead>
                      <tr>
                        <th>#</th>
                        <th>Instrument</th>
                        <th>Product</th>
                        <th>Side</th>
                        <th class="right">Qty</th>
                        <th class="right">Amount</th>
                        <th class="right">Weight %</th>
                        <th>CCY</th>
                        <th>Notes</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="line in decision.lines" :key="line.id ?? line.line_number">
                        <td>{{ line.line_number }}</td>
                        <td>{{ line.instrument_code ?? "—" }}</td>
                        <td>{{ line.product_type ?? "—" }}</td>
                        <td>{{ line.side ?? "—" }}</td>
                        <td class="right">{{ line.quantity ?? "—" }}</td>
                        <td class="right">{{ line.amount ?? "—" }}</td>
                        <td class="right">
                          {{ line.target_weight ? `${line.target_weight}%` : "—" }}
                        </td>
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
  background: rgba(0, 0, 0, 0.4);
  z-index: 900;
  display: flex;
  justify-content: flex-end;
}

.drawer {
  width: min(720px, 90vw);
  height: 100%;
  background: var(--surface-primary, #fff);
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-xl);
  overflow: hidden;
}

.drawer__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-5) var(--space-6);
  border-bottom: 1px solid var(--border-primary);
  flex-shrink: 0;
}

.drawer__title-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.drawer__title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  margin: 0;
}

.drawer__body {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-5) var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.drawer__section-title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin: 0 0 var(--space-3) 0;
}

.drawer__lines {
  overflow-x: auto;
}

.drawer__lines-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
}

.drawer__lines-table th,
.drawer__lines-table td {
  padding: var(--space-2) var(--space-3);
  text-align: left;
  border-bottom: 1px solid var(--border-secondary);
  white-space: nowrap;
}

.drawer__lines-table th {
  font-weight: var(--font-weight-semibold);
  color: var(--text-secondary);
  background: var(--surface-secondary);
}

.drawer__lines-table .right {
  text-align: right;
}

.drawer__lines-table .notes {
  white-space: normal;
  max-width: 200px;
  word-break: break-word;
}

.drawer-enter-active,
.drawer-leave-active {
  transition: opacity 0.2s ease;
}

.drawer-enter-active .drawer,
.drawer-leave-active .drawer {
  transition: transform 0.25s ease;
}

.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}

.drawer-enter-from .drawer,
.drawer-leave-to .drawer {
  transform: translateX(100%);
}
</style>
