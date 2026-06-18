<script setup lang="ts">
import { computed } from "vue";
import AppDataTable from "~/shared/ui/AppDataTable.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import type { ApiDecision } from "../services/decisionApprovalApi";

interface Props {
  items: ApiDecision[];
  loading?: boolean;
  error?: string | null;
  selectedIds?: string[];
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  error: null,
  selectedIds: () => [],
});

const emit = defineEmits<{
  "select-change": [ids: string[]];
  "row-click": [item: ApiDecision];
}>();

const columns = [
  { key: "decision_number", label: "Decision No", width: "140px" },
  { key: "business_date", label: "Date", width: "110px" },
  { key: "process_type", label: "Process", width: "130px" },
  { key: "product_type", label: "Product", width: "110px" },
  { key: "instrument_code", label: "Instrument", width: "140px" },
  { key: "side", label: "Side", width: "90px" },
  { key: "amount_qty", label: "Amt / Qty", width: "110px", align: "right" as const },
  { key: "stage", label: "Stage", width: "90px", align: "center" as const },
  { key: "current_approvers", label: "Pending Approvers", width: "200px" },
  { key: "status", label: "Status", width: "130px" },
];

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

function formatAmtQty(item: ApiDecision): string {
  if (item.amount) return item.amount;
  if (item.quantity) return item.quantity;
  if (item.lines && item.lines.length > 0) return `${item.lines.length} lines`;
  return "—";
}

function stageLabel(item: ApiDecision): string {
  if (!item.approval_stage) return "—";
  return item.approval_total_stages
    ? `${item.approval_stage} / ${item.approval_total_stages}`
    : String(item.approval_stage);
}
</script>

<template>
  <AppDataTable
    :columns="columns"
    :items="items"
    :loading="loading"
    :error="error"
    :selected-ids="selectedIds"
    :row-clickable="true"
    empty-text="No decisions pending approval."
    density="compact"
    @select-change="emit('select-change', $event)"
    @row-click="emit('row-click', $event)"
  >
    <template #cell-decision_number="{ item }">
      <span class="decision-grid__no">{{ item.decision_number ?? "—" }}</span>
    </template>

    <template #cell-process_type="{ item }">
      {{ (item.process_type ?? "").replace(/_/g, " ") }}
    </template>

    <template #cell-product_type="{ item }">
      {{ item.product_type ?? "—" }}
    </template>

    <template #cell-instrument_code="{ item }">
      {{ item.instrument_code ?? "—" }}
    </template>

    <template #cell-side="{ item }">
      <span :class="['decision-grid__side', item.side ? `decision-grid__side--${item.side.toLowerCase()}` : '']">
        {{ item.side ?? "BASKET" }}
      </span>
    </template>

    <template #cell-amount_qty="{ item }">
      {{ formatAmtQty(item) }}
    </template>

    <template #cell-stage="{ item }">
      {{ stageLabel(item) }}
    </template>

    <template #cell-current_approvers="{ item }">
      <span v-if="item.current_approvers && item.current_approvers.length">
        {{ item.current_approvers.join(", ") }}
      </span>
      <span v-else class="decision-grid__empty">—</span>
    </template>

    <template #cell-status="{ item }">
      <AppStatusBadge
        :status="statusType(item.status)"
        :label="(item.status ?? '').replace(/_/g, ' ')"
        size="sm"
      />
    </template>
  </AppDataTable>
</template>

<style scoped>
.decision-grid__no {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.decision-grid__side {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.decision-grid__side--buy { color: var(--color-success-600, #16a34a); }
.decision-grid__side--sell { color: var(--color-danger-600, #dc2626); }

.decision-grid__empty {
  color: var(--text-tertiary);
}
</style>
