<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppDataTable from "~/shared/ui/AppDataTable.vue";
import type { TableColumn } from "~/shared/ui/AppDataTable.vue";
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

const { t } = useI18n();

const columns = computed<TableColumn[]>(() => [
  {
    key: "decision_number",
    label: t("approval.decisionWorkbench.table.decisionNumber", "Decision number"),
    width: "156px",
  },
  {
    key: "business_date",
    label: t("approval.decisionWorkbench.table.businessDate", "Business date"),
    width: "118px",
  },
  {
    key: "process_type",
    label: t("approval.decisionWorkbench.table.process", "Process"),
    width: "142px",
  },
  {
    key: "product_type",
    label: t("approval.decisionWorkbench.table.product", "Product"),
    width: "110px",
  },
  {
    key: "instrument_code",
    label: t("approval.decisionWorkbench.table.instrument", "Instrument"),
    width: "132px",
  },
  {
    key: "side",
    label: t("approval.decisionWorkbench.table.side", "Side"),
    width: "80px",
  },
  {
    key: "amount_qty",
    label: t("approval.decisionWorkbench.table.amountQuantity", "Amount / quantity"),
    width: "132px",
    align: "right",
  },
  {
    key: "stage",
    label: t("approval.decisionWorkbench.table.stage", "Stage"),
    width: "88px",
    align: "center",
  },
  {
    key: "current_approvers",
    label: t("approval.decisionWorkbench.table.pendingApprovers", "Pending approvers"),
    width: "200px",
  },
  {
    key: "status",
    label: t("approval.decisionWorkbench.table.status", "Status"),
    width: "132px",
  },
  {
    key: "actions",
    label: t("approval.decisionWorkbench.table.action", "Action"),
    width: "92px",
    align: "right",
  },
]);

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
      return (status ?? "—").replace(/_/g, " ");
  }
}

function formatAmtQty(item: ApiDecision): string {
  if (item.amount) return item.currency ? `${item.amount} ${item.currency}` : item.amount;
  if (item.quantity) return item.quantity;
  if (item.lines && item.lines.length > 0) {
    return t(
      "approval.decisionWorkbench.table.lineCount",
      { count: item.lines.length },
      "{count} line(s)",
    );
  }
  return "—";
}

function stageLabel(item: ApiDecision): string {
  if (!item.approval_stage) return "—";
  return item.approval_total_stages
    ? `${item.approval_stage} / ${item.approval_total_stages}`
    : String(item.approval_stage);
}

function humanize(value: string | undefined): string {
  return value ? value.replace(/_/g, " ") : "—";
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
    selectable
    :empty-text="t('approval.decisionWorkbench.table.empty', 'No pending decisions match this query.')"
    density="compact"
    @select-change="emit('select-change', $event)"
    @row-click="emit('row-click', $event)"
  >
    <template #[`cell(decision_number)`]="{ item }">
      <span class="decision-grid__no">{{ item.decision_number ?? "—" }}</span>
    </template>

    <template #[`cell(process_type)`]="{ item }">
      {{ humanize(item.process_type) }}
    </template>

    <template #[`cell(product_type)`]="{ item }">
      {{ humanize(item.product_type) }}
    </template>

    <template #[`cell(instrument_code)`]="{ item }">
      {{ item.instrument_code ?? "—" }}
    </template>

    <template #[`cell(side)`]="{ item }">
      <span
        :class="[
          'decision-grid__side',
          item.side ? `decision-grid__side--${item.side.toLowerCase()}` : '',
        ]"
      >
        {{ item.side ?? t("approval.decisionWorkbench.table.basket", "Basket") }}
      </span>
    </template>

    <template #[`cell(amount_qty)`]="{ item }">
      <span class="decision-grid__number">{{ formatAmtQty(item) }}</span>
    </template>

    <template #[`cell(stage)`]="{ item }">
      <span class="decision-grid__stage">{{ stageLabel(item) }}</span>
    </template>

    <template #[`cell(current_approvers)`]="{ item }">
      <span v-if="item.current_approvers?.length">
        {{ item.current_approvers.join(", ") }}
      </span>
      <span v-else class="decision-grid__empty">—</span>
    </template>

    <template #[`cell(status)`]="{ item }">
      <AppStatusBadge
        :status="statusType(item.approval_status ?? item.status)"
        :label="statusLabel(item.approval_status ?? item.status)"
        size="sm"
      />
    </template>

    <template #[`cell(actions)`]="{ item }">
      <AppButton variant="ghost" size="xs" @click.stop="emit('row-click', item)">
        {{ t("approval.decisionWorkbench.table.inspect", "Inspect") }}
      </AppButton>
    </template>
  </AppDataTable>
</template>

<style scoped>
.decision-grid__no,
.decision-grid__number,
.decision-grid__stage {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 12px);
  font-variant-numeric: tabular-nums;
}

.decision-grid__no {
  color: var(--text-primary, #1f2328);
  font-weight: var(--font-weight-semibold, 600);
}

.decision-grid__side {
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.decision-grid__side--buy {
  color: var(--state-success, #1a7f37);
}

.decision-grid__side--sell {
  color: var(--state-danger, #cf222e);
}

.decision-grid__empty {
  color: var(--text-tertiary, #6e7781);
}
</style>
