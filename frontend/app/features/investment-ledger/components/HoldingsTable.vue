<script setup lang="ts">
import { computed } from "vue";

import { formatMoney, formatQuantity, shortenId } from "../lib/ledgerFormat";
import type {
  ApiHolding,
  ApiInstrument,
} from "../services/investmentLedgerApi";

interface Props {
  holdings: ApiHolding[];
  instruments?: ApiInstrument[];
  baseCurrency?: string;
  loading?: boolean;
  error?: string | null;
}

const props = withDefaults(defineProps<Props>(), {
  instruments: () => [],
  baseCurrency: "",
  loading: false,
  error: null,
});

const emit = defineEmits<{
  select: [holding: ApiHolding];
  refresh: [];
}>();

const instrumentMap = computed(() => {
  const map = new Map<string, ApiInstrument>();
  for (const inst of props.instruments) {
    if (inst.id) map.set(inst.id, inst);
  }
  return map;
});

function instrumentLabel(h: ApiHolding): string {
  if (!h.instrument_id) return "—";
  const inst = instrumentMap.value.get(h.instrument_id);
  if (!inst) return shortenId(h.instrument_id);
  const ticker = inst.primary_ticker ?? "";
  const name = inst.name ?? "";
  return ticker && name ? `${ticker} — ${name}` : ticker || name || shortenId(h.instrument_id);
}

function instrumentExchange(h: ApiHolding): string {
  if (!h.instrument_id) return "";
  return instrumentMap.value.get(h.instrument_id)?.primary_exchange ?? "";
}
</script>

<template>
  <div class="holdings-table">
    <header v-if="loading || error" class="holdings-table__state">
      <span v-if="loading">Loading holdings…</span>
      <span v-else-if="error" class="holdings-table__error" role="alert">
        {{ error }}
      </span>
    </header>

    <table v-if="holdings.length > 0" class="holdings-table__table">
      <thead>
        <tr>
          <th>Instrument</th>
          <th>Exchange</th>
          <th class="holdings-table__col-num">Quantity</th>
          <th class="holdings-table__col-num">Avg cost</th>
          <th class="holdings-table__col-num">Cost basis</th>
          <th>Last txn</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="h in holdings"
          :key="h.instrument_id ?? Math.random()"
          @click="emit('select', h)"
        >
          <td class="holdings-table__cell-primary">{{ instrumentLabel(h) }}</td>
          <td>{{ instrumentExchange(h) || "—" }}</td>
          <td class="holdings-table__col-num">{{ formatQuantity(h.quantity) }}</td>
          <td class="holdings-table__col-num">
            {{ formatMoney(h.average_cost, baseCurrency) }}
          </td>
          <td class="holdings-table__col-num">
            {{ formatMoney(h.cost_basis, baseCurrency) }}
          </td>
          <td class="holdings-table__col-mono">
            {{ h.last_business_date ?? "—" }}
          </td>
        </tr>
      </tbody>
    </table>

    <p v-else-if="!loading && !error" class="holdings-table__empty">
      No holdings recorded for this portfolio yet.
    </p>
  </div>
</template>

<style scoped>
.holdings-table {
  display: grid;
  gap: var(--space-2);
}

.holdings-table__state {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  padding: 0 var(--space-3);
}

.holdings-table__error {
  color: var(--state-danger, #dc2626);
}

.holdings-table__table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
}

.holdings-table__table th,
.holdings-table__table td {
  padding: var(--space-2) var(--space-3);
  text-align: left;
  border-bottom: 1px solid var(--border-subtle);
}

.holdings-table__table th {
  background: var(--bg-card-muted);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
  font-weight: var(--font-weight-semibold);
}

.holdings-table__table tbody tr {
  cursor: pointer;
}

.holdings-table__table tbody tr:hover {
  background: var(--bg-card-muted);
}

.holdings-table__cell-primary {
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.holdings-table__col-num {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.holdings-table__col-mono {
  font-family: var(--font-family-mono);
  color: var(--text-secondary);
}

.holdings-table__empty {
  margin: 0;
  padding: var(--space-5) var(--space-3);
  text-align: center;
  color: var(--text-tertiary);
  font-size: var(--font-size-sm);
}
</style>
