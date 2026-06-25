<script setup lang="ts">
import DecisionStatusBadge from "./DecisionStatusBadge.vue";
import type { ApiDecision } from "../services/decisionApi";

const props = defineProps<{
  items: ApiDecision[];
  loading: boolean;
  total: number;
  page: number;
  limit: number;
  selectedNos: string[];
}>();

const emit = defineEmits<{
  (e: "update:selectedNos", value: string[]): void;
  (e: "update:page", value: number): void;
  (e: "open", decision: ApiDecision): void;
}>();

const totalPages = computed(() =>
  props.limit > 0 ? Math.ceil(props.total / props.limit) : 1,
);

function isSelected(no: string) {
  return props.selectedNos.includes(no);
}

function toggleRow(no: string) {
  if (isSelected(no)) {
    emit(
      "update:selectedNos",
      props.selectedNos.filter((n) => n !== no),
    );
  } else {
    emit("update:selectedNos", [...props.selectedNos, no]);
  }
}

const allChecked = computed(
  () =>
    props.items.length > 0 &&
    props.items.every((d) => isSelected(d.decision_number ?? "")),
);

function toggleAll() {
  if (allChecked.value) {
    emit("update:selectedNos", []);
  } else {
    emit(
      "update:selectedNos",
      props.items.map((d) => d.decision_number ?? "").filter(Boolean),
    );
  }
}

function formatDate(s?: string) {
  if (!s) return "—";
  return s.slice(0, 10);
}

function formatDateTime(s?: string) {
  if (!s) return "—";
  return s.replace("T", " ").slice(0, 16);
}
</script>

<template>
  <div class="table-wrapper">
    <table class="decision-table">
      <thead>
        <tr>
          <th class="col-check">
            <input
              type="checkbox"
              :checked="allChecked"
              :indeterminate="selectedNos.length > 0 && !allChecked"
              @change="toggleAll"
            />
          </th>
          <th>Decision No</th>
          <th>Instrument</th>
          <th>Side</th>
          <th>Qty / Amount</th>
          <th>Status</th>
          <th>Business Date</th>
          <th>Submitted At</th>
          <th class="col-action">Action</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td colspan="9" class="empty">Loading…</td>
        </tr>
        <tr v-else-if="items.length === 0">
          <td colspan="9" class="empty">No decisions found.</td>
        </tr>
        <tr
          v-for="d in items"
          :key="d.id"
          :class="{ selected: isSelected(d.decision_number ?? '') }"
        >
          <td class="col-check">
            <input
              type="checkbox"
              :checked="isSelected(d.decision_number ?? '')"
              @change="toggleRow(d.decision_number ?? '')"
            />
          </td>
          <td class="cell-mono">{{ d.decision_number ?? "—" }}</td>
          <td>{{ d.instrument_code ?? "—" }}</td>
          <td>
            <span class="side-badge" :data-side="d.side">{{ d.side ?? "—" }}</span>
          </td>
          <td class="cell-mono">
            <template v-if="d.quantity">{{ d.quantity }}</template>
            <template v-else-if="d.amount">{{ d.amount }}</template>
            <template v-else>—</template>
            <span v-if="d.currency" class="cell-ccy"> {{ d.currency }}</span>
          </td>
          <td><DecisionStatusBadge :status="d.status" /></td>
          <td>{{ formatDate(d.business_date) }}</td>
          <td>{{ formatDateTime(d.submitted_at) }}</td>
          <td class="col-action">
            <button class="btn-open" @click="emit('open', d)">Open</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="totalPages > 1" class="pagination">
      <button :disabled="page <= 1" @click="emit('update:page', page - 1)">
        ‹ Prev
      </button>
      <span>Page {{ page }} / {{ totalPages }}</span>
      <button
        :disabled="page >= totalPages"
        @click="emit('update:page', page + 1)"
      >
        Next ›
      </button>
    </div>
  </div>
</template>

<style scoped>
.table-wrapper {
  overflow-x: auto;
}

.decision-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.decision-table th,
.decision-table td {
  padding: 8px 10px;
  text-align: left;
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.decision-table th {
  background: var(--bg-card-muted);
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary);
}

.decision-table tr.selected td {
  background: rgba(9, 105, 218, 0.06);
}

.decision-table tr:hover td {
  background: var(--bg-card-hover);
}

.col-check {
  width: 36px;
  text-align: center;
}

.col-action {
  width: 72px;
  text-align: center;
}

.cell-mono {
  font-family: monospace;
  font-size: 12px;
}

.cell-ccy {
  color: var(--text-secondary);
  font-size: 11px;
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

.pagination {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  font-size: 13px;
  color: var(--text-secondary);
}

.pagination button {
  padding: 4px 10px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  cursor: pointer;
  font-size: 13px;
}

.pagination button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>
