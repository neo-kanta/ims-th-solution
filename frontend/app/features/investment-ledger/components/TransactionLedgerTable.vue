<script setup lang="ts">
import { computed } from "vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

import AppBadge from "~/shared/ui/AppBadge.vue";

import {
  formatMoney,
  formatQuantity,
  shortenId,
  statusBadgeVariant,
} from "../lib/ledgerFormat";
import type {
  ApiInstrument,
  ApiTransaction,
} from "../services/investmentLedgerApi";

interface Props {
  transactions: ApiTransaction[];
  instruments?: ApiInstrument[];
  loading?: boolean;
  error?: string | null;
}

const props = withDefaults(defineProps<Props>(), {
  instruments: () => [],
  loading: false,
  error: null,
});

const instrumentMap = computed(() => {
  const m = new Map<string, ApiInstrument>();
  for (const inst of props.instruments) {
    if (inst.id) m.set(inst.id, inst);
  }
  return m;
});

function instrumentLabel(t: ApiTransaction): string {
  if (!t.instrument_id) return "—";
  const inst = instrumentMap.value.get(t.instrument_id);
  if (!inst) return shortenId(t.instrument_id);
  return inst.primary_ticker ?? inst.name ?? shortenId(t.instrument_id);
}

function sideTone(side: string | null | undefined): "success" | "error" | "neutral" {
  if (!side) return "neutral";
  const upper = side.toUpperCase();
  if (upper === "BUY") return "success";
  if (upper === "SELL") return "error";
  return "neutral";
}
</script>

<template>
  <div class="txn-ledger">
    <AppLoadingState v-if="loading && transactions.length === 0" message="Loading transaction ledger…" />

    <header v-else-if="loading || error" class="txn-ledger__state">
      <span v-if="loading">Loading transaction ledger…</span>
      <span v-else-if="error" class="txn-ledger__error" role="alert">
        {{ error }}
      </span>
    </header>

    <table v-if="transactions.length > 0" class="txn-ledger__table">
      <thead>
        <tr>
          <th>Business date</th>
          <th>Type</th>
          <th>Side</th>
          <th>Instrument</th>
          <th class="txn-ledger__col-num">Quantity</th>
          <th class="txn-ledger__col-num">Price</th>
          <th class="txn-ledger__col-num">Gross</th>
          <th class="txn-ledger__col-num">Net</th>
          <th>Status</th>
          <th>Ref</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="t in transactions" :key="t.id ?? `${t.business_date}-${t.instrument_id}`">
          <td class="txn-ledger__col-mono">{{ t.business_date ?? "—" }}</td>
          <td>
            <span class="txn-ledger__type">{{ t.transaction_type ?? "—" }}</span>
          </td>
          <td>
            <AppBadge
              v-if="t.side"
              :variant="sideTone(t.side) === 'success' ? 'success' : sideTone(t.side) === 'error' ? 'error' : 'neutral'"
              size="sm"
            >
              {{ t.side }}
            </AppBadge>
            <span v-else>—</span>
          </td>
          <td class="txn-ledger__cell-primary">{{ instrumentLabel(t) }}</td>
          <td class="txn-ledger__col-num">{{ formatQuantity(t.quantity) }}</td>
          <td class="txn-ledger__col-num">
            {{ formatMoney(t.price, t.currency ?? "") }}
          </td>
          <td class="txn-ledger__col-num">
            {{ formatMoney(t.gross_amount, t.currency ?? "") }}
          </td>
          <td class="txn-ledger__col-num">
            {{ formatMoney(t.net_amount, t.currency ?? "") }}
          </td>
          <td>
            <AppBadge :variant="statusBadgeVariant(t.status)" size="sm">
              {{ t.status ?? "—" }}
            </AppBadge>
          </td>
          <td class="txn-ledger__col-mono" :title="t.external_ref ?? t.id ?? ''">
            {{ t.external_ref ? t.external_ref : shortenId(t.id) }}
          </td>
        </tr>
      </tbody>
    </table>

    <p v-else-if="!loading && !error" class="txn-ledger__empty">
      No transactions on the ledger yet — the first post will appear here.
    </p>
  </div>
</template>

<style scoped>
.txn-ledger {
  display: grid;
  gap: var(--space-2);
}

.txn-ledger__state {
  padding: 0 var(--space-3);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.txn-ledger__error {
  color: var(--state-danger, #dc2626);
}

.txn-ledger__table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
}

.txn-ledger__table th,
.txn-ledger__table td {
  padding: var(--space-2) var(--space-3);
  text-align: left;
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.txn-ledger__table th {
  background: var(--bg-card-muted);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
  font-weight: var(--font-weight-semibold);
}

.txn-ledger__col-num {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.txn-ledger__col-mono {
  font-family: var(--font-family-mono);
  color: var(--text-secondary);
}

.txn-ledger__type {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.txn-ledger__cell-primary {
  font-weight: var(--font-weight-semibold);
}

.txn-ledger__empty {
  margin: 0;
  padding: var(--space-5) var(--space-3);
  text-align: center;
  color: var(--text-tertiary);
  font-size: var(--font-size-sm);
}
</style>
