<script setup lang="ts">
import { computed } from "vue";

import {
  compareDecimal,
  formatMoney,
  formatQuantity,
  shortenId,
} from "../lib/ledgerFormat";
import type {
  ApiInstrument,
  ApiPositionProjection,
} from "../services/investmentLedgerApi";

interface Props {
  projection: ApiPositionProjection | null | undefined;
  instrument?: ApiInstrument | null;
  currency?: string;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  instrument: null,
  currency: "",
  loading: false,
});

const qtyDelta = computed(() =>
  compareDecimal(
    props.projection?.projected_quantity,
    props.projection?.current_quantity,
  ),
);

const qtyTone = computed<"positive" | "negative" | "neutral">(() => {
  if (qtyDelta.value > 0) return "positive";
  if (qtyDelta.value < 0) return "negative";
  return "neutral";
});

const instrumentLabel = computed(() => {
  const inst = props.instrument;
  if (!inst) {
    return props.projection?.instrument_id
      ? shortenId(props.projection.instrument_id)
      : "—";
  }
  const ticker = inst.primary_ticker ? inst.primary_ticker : "";
  const name = inst.name ?? "";
  if (ticker && name) return `${ticker} — ${name}`;
  return ticker || name || shortenId(inst.id);
});
</script>

<template>
  <div class="position-projection">
    <div class="position-projection__header">
      <h3 class="position-projection__title">Position impact</h3>
      <span class="position-projection__chip">{{ instrumentLabel }}</span>
    </div>

    <div v-if="loading" class="position-projection__placeholder">
      Calculating…
    </div>

    <div v-else-if="!projection" class="position-projection__placeholder">
      Run a simulation to preview the position impact.
    </div>

    <div v-else class="position-projection__grid">
      <div class="position-projection__col">
        <span class="position-projection__col-label">Current</span>
        <dl class="position-projection__defs">
          <div>
            <dt>Quantity</dt>
            <dd>{{ formatQuantity(projection.current_quantity) }}</dd>
          </div>
          <div>
            <dt>Avg cost</dt>
            <dd>{{ formatMoney(projection.current_average_cost, currency) }}</dd>
          </div>
          <div>
            <dt>Cost basis</dt>
            <dd>{{ formatMoney(projection.current_cost_basis, currency) }}</dd>
          </div>
        </dl>
      </div>

      <div class="position-projection__arrow" aria-hidden="true">→</div>

      <div
        class="position-projection__col position-projection__col--projected"
        :class="`position-projection__col--${qtyTone}`"
      >
        <span class="position-projection__col-label">Projected</span>
        <dl class="position-projection__defs">
          <div>
            <dt>Quantity</dt>
            <dd>{{ formatQuantity(projection.projected_quantity) }}</dd>
          </div>
          <div>
            <dt>Avg cost</dt>
            <dd>
              {{ formatMoney(projection.projected_average_cost, currency) }}
            </dd>
          </div>
          <div>
            <dt>Cost basis</dt>
            <dd>{{ formatMoney(projection.projected_cost_basis, currency) }}</dd>
          </div>
        </dl>
      </div>
    </div>
  </div>
</template>

<style scoped>
.position-projection {
  display: grid;
  gap: var(--space-3);
}

.position-projection__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-2);
}

.position-projection__title {
  margin: 0;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
}

.position-projection__chip {
  font-size: 10px;
  font-weight: var(--font-weight-bold);
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--bg-card-muted);
  border: 1px solid var(--border-subtle);
  color: var(--text-secondary);
  max-width: 50%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.position-projection__placeholder {
  padding: var(--space-3);
  text-align: center;
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
  border: 1px dashed var(--border-subtle);
  border-radius: var(--radius-md);
}

.position-projection__grid {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: var(--space-3);
  align-items: stretch;
}

.position-projection__col {
  padding: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  display: grid;
  gap: var(--space-2);
}

.position-projection__col--projected {
  background: var(--bg-card-muted);
}

.position-projection__col--positive {
  border-color: var(--state-success, #059669);
}

.position-projection__col--negative {
  border-color: var(--state-danger, #dc2626);
}

.position-projection__col-label {
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
  font-weight: var(--font-weight-bold);
}

.position-projection__arrow {
  display: grid;
  place-items: center;
  font-size: var(--font-size-lg);
  color: var(--text-tertiary);
}

.position-projection__defs {
  margin: 0;
  display: grid;
  gap: 2px;
  font-size: var(--font-size-sm);
  font-variant-numeric: tabular-nums;
}

.position-projection__defs > div {
  display: flex;
  justify-content: space-between;
  gap: var(--space-2);
}

.position-projection__defs dt {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.position-projection__defs dd {
  margin: 0;
  font-variant-numeric: tabular-nums;
}
</style>
