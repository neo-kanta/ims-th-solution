<script setup lang="ts">
import DecisionStatusBadge from "~/features/investment-decision/components/DecisionStatusBadge.vue";
import type { ApiDecisionV2 } from "../services/portfolioDecisionApi";
import { formatMoney, formatQuantity } from "../lib/decisionFormat";

defineProps<{
  items: ApiDecisionV2[];
  loading: boolean;
}>();

const emit = defineEmits<{
  open: [decision: ApiDecisionV2];
}>();

function formatDateTime(value?: string): string {
  if (!value) return "—";
  return value.replace("T", " ").slice(0, 16);
}
</script>

<template>
  <div class="decision-list-table">
    <table>
      <thead>
        <tr>
          <th>Decision No</th>
          <th>Instrument</th>
          <th>Side</th>
          <th>Quantity</th>
          <th>Limit price</th>
          <th>Status</th>
          <th>Business date</th>
          <th>Created / submitted</th>
          <th class="col-action">Action</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="9" class="empty">Loading decisions…</td>
        </tr>
        <tr v-else-if="items.length === 0">
          <td colspan="9" class="empty">No decisions yet for this portfolio.</td>
        </tr>
        <tr v-for="d in items" :key="d.id" class="decision-row" @click="emit('open', d)">
          <td class="cell-mono">{{ d.decision_number ?? "—" }}</td>
          <td>{{ d.instrument_code ?? "—" }}</td>
          <td>
            <span class="side-badge" :data-side="d.side">{{ d.side ?? "—" }}</span>
          </td>
          <td class="cell-mono">
            <template v-if="d.quantity">{{ formatQuantity(d.quantity) }}</template>
            <template v-else-if="d.amount">{{ formatMoney(d.amount, d.currency) }}</template>
            <template v-else>—</template>
          </td>
          <td class="cell-mono">{{ d.limit_price ? formatMoney(d.limit_price, d.currency) : "—" }}</td>
          <td><DecisionStatusBadge :status="d.status" /></td>
          <td>{{ d.business_date ?? "—" }}</td>
          <td>{{ formatDateTime(d.submitted_at ?? d.created_at) }}</td>
          <td class="col-action">
            <button type="button" class="btn-open" @click.stop="emit('open', d)">Open</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.decision-list-table {
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
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

th {
  background: var(--bg-card-muted);
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary);
}

.decision-row {
  cursor: pointer;
}

.decision-row:hover td {
  background: var(--bg-card-hover);
}

.cell-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}

.side-badge {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
}

.side-badge[data-side="BUY"] {
  background: rgba(26, 127, 55, 0.15);
  color: var(--state-success, #1a7f37);
}

.side-badge[data-side="SELL"] {
  background: rgba(207, 34, 46, 0.15);
  color: var(--state-danger, #cf222e);
}

.empty {
  text-align: center;
  padding: 32px;
  color: var(--text-secondary);
}

.col-action {
  width: 72px;
}

.btn-open {
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  color: var(--text-primary);
  cursor: pointer;
}

.btn-open:hover {
  background: var(--bg-card-hover);
}
</style>
