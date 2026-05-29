<script setup lang="ts">
import { formatMoney } from "../lib/ledgerFormat";
import type { ApiCashBalance } from "../services/investmentLedgerApi";

interface Props {
  balances: ApiCashBalance[];
  loading?: boolean;
  error?: string | null;
}

withDefaults(defineProps<Props>(), {
  loading: false,
  error: null,
});
</script>

<template>
  <div class="cash-balances">
    <header class="cash-balances__header">
      <span class="cash-balances__title">Cash by currency</span>
      <span v-if="loading" class="cash-balances__hint">Loading…</span>
    </header>

    <p v-if="error" class="cash-balances__error" role="alert">{{ error }}</p>

    <ul v-if="balances.length > 0" class="cash-balances__list">
      <li
        v-for="balance in balances"
        :key="balance.currency ?? `bal-${Math.random()}`"
        class="cash-balances__item"
      >
        <div class="cash-balances__currency">{{ balance.currency ?? "—" }}</div>
        <div class="cash-balances__amount">
          {{ formatMoney(balance.balance) }}
        </div>
        <div class="cash-balances__meta">
          as of {{ balance.last_business_date ?? "—" }}
        </div>
      </li>
    </ul>

    <p v-else-if="!loading && !error" class="cash-balances__empty">
      No cash balances on record.
    </p>
  </div>
</template>

<style scoped>
.cash-balances {
  display: grid;
  gap: var(--space-2);
}

.cash-balances__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.cash-balances__title {
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-weight: var(--font-weight-bold);
  color: var(--text-tertiary);
}

.cash-balances__hint {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.cash-balances__error {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--state-danger, #dc2626);
}

.cash-balances__list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: var(--space-2);
}

.cash-balances__item {
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: var(--space-2);
  align-items: baseline;
  font-variant-numeric: tabular-nums;
}

.cash-balances__currency {
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
}

.cash-balances__amount {
  text-align: right;
  font-weight: var(--font-weight-semibold);
}

.cash-balances__meta {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.cash-balances__empty {
  margin: 0;
  padding: var(--space-3);
  text-align: center;
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
  border: 1px dashed var(--border-subtle);
  border-radius: var(--radius-md);
}
</style>
