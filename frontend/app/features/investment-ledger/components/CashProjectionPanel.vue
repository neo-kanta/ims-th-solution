<script setup lang="ts">
import { computed } from "vue";

import { formatMoney, formatMoneySigned } from "../lib/ledgerFormat";
import type { ApiCashProjection } from "../services/investmentLedgerApi";

interface Props {
  projection: ApiCashProjection | null | undefined;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const currency = computed(() => props.projection?.currency ?? "");
const cashImpact = computed(() => props.projection?.cash_impact ?? null);
const currentBalance = computed(() => props.projection?.current_balance ?? null);
const projectedBalance = computed(
  () => props.projection?.projected_balance ?? null,
);

const impactTone = computed<"positive" | "negative" | "neutral">(() => {
  if (!cashImpact.value) return "neutral";
  const n = Number(cashImpact.value);
  if (!Number.isFinite(n) || n === 0) return "neutral";
  return n > 0 ? "positive" : "negative";
});

const projectionTone = computed<"positive" | "negative" | "neutral">(() => {
  if (!projectedBalance.value) return "neutral";
  const n = Number(projectedBalance.value);
  if (!Number.isFinite(n)) return "neutral";
  return n < 0 ? "negative" : "neutral";
});
</script>

<template>
  <div class="cash-projection">
    <div class="cash-projection__header">
      <h3 class="cash-projection__title">Cash impact</h3>
      <span v-if="currency" class="cash-projection__chip">{{ currency }}</span>
    </div>

    <div v-if="loading" class="cash-projection__placeholder">Calculating…</div>

    <div v-else-if="!projection" class="cash-projection__placeholder">
      Run a simulation to preview the cash impact.
    </div>

    <div v-else class="cash-projection__grid">
      <div class="cash-projection__row">
        <span class="cash-projection__label">Current balance</span>
        <span class="cash-projection__value">
          {{ formatMoney(currentBalance, currency) }}
        </span>
      </div>
      <div class="cash-projection__row" :class="`cash-projection__row--${impactTone}`">
        <span class="cash-projection__label">Cash impact</span>
        <span class="cash-projection__value">
          {{ formatMoneySigned(cashImpact, currency) }}
        </span>
      </div>
      <div
        class="cash-projection__row cash-projection__row--total"
        :class="`cash-projection__row--${projectionTone}`"
      >
        <span class="cash-projection__label">Projected balance</span>
        <span class="cash-projection__value">
          {{ formatMoney(projectedBalance, currency) }}
        </span>
      </div>
    </div>

    <p v-if="projection && projectionTone === 'negative'" class="cash-projection__warn">
      Projected balance is negative — review available cash before posting.
    </p>
  </div>
</template>

<style scoped>
.cash-projection {
  display: grid;
  gap: var(--space-3);
}

.cash-projection__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.cash-projection__title {
  margin: 0;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}

.cash-projection__chip {
  font-size: 10px;
  font-weight: var(--font-weight-bold);
  letter-spacing: 0.04em;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--bg-card-muted);
  color: var(--text-secondary);
  border: 1px solid var(--border-subtle);
}

.cash-projection__placeholder {
  padding: var(--space-3);
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
  border: 1px dashed var(--border-subtle);
  border-radius: var(--radius-md);
  text-align: center;
}

.cash-projection__grid {
  display: grid;
  gap: var(--space-2);
}

.cash-projection__row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  font-variant-numeric: tabular-nums;
}

.cash-projection__row--total {
  border-color: var(--border-default);
  background: var(--bg-card-muted);
  font-weight: var(--font-weight-semibold);
}

.cash-projection__row--positive {
  color: var(--state-success, #059669);
}

.cash-projection__row--negative {
  color: var(--state-danger, #dc2626);
}

.cash-projection__label {
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary);
}

.cash-projection__value {
  font-size: var(--font-size-md);
  font-variant-numeric: tabular-nums;
}

.cash-projection__warn {
  margin: 0;
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-xs);
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  border: 1px solid var(--alert-warning-border);
  border-radius: var(--radius-sm);
}
</style>
