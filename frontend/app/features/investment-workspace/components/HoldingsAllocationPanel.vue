<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import { clampPct, formatPercent } from "../lib/holdingsFormat";
import type { AllocationBar, AllocationDimension, AllocationPayload } from "../types";

const props = defineProps<{
  payload: AllocationPayload | null;
  dimension: AllocationDimension;
}>();

const emit = defineEmits<{
  changeDimension: [dimension: AllocationDimension];
}>();

const { t } = useI18n();

const dimensions: AllocationDimension[] = ["country", "category", "industry", "currency"];

const bars = computed<AllocationBar[]>(() => {
  if (!props.payload) return [];
  return props.payload.dimensions[props.dimension] ?? [];
});

// Largest bar drives the visual scale so smaller weights remain visible.
const scaleMax = computed(() => {
  const values = bars.value
    .filter((b) => b.key !== "TOTAL")
    .map((b) => b.value_numeric);
  if (values.length === 0) return 1;
  const max = Math.max(...values);
  return max > 0 ? max : 1;
});

function barWidth(b: AllocationBar): number {
  return clampPct((b.value_numeric / scaleMax.value) * 100);
}
</script>

<template>
  <div class="ht-alloc">
    <div class="ht-alloc__tabs" role="tablist">
      <button
        v-for="d in dimensions"
        :key="d"
        type="button"
        role="tab"
        :aria-selected="dimension === d"
        class="ht-alloc__tab"
        :class="{ 'is-active': dimension === d }"
        @click="emit('changeDimension', d)"
      >
        {{ t(`holdings.allocation.dimensions.${d}` as any, d) }}
      </button>
    </div>

    <div class="ht-alloc__head">
      <span class="ht-alloc__head-col">{{ t(`holdings.allocation.dimensions.${dimension}` as any, dimension) }}</span>
    </div>

    <ul class="ht-alloc__list">
      <li
        v-for="b in bars"
        :key="b.key"
        class="ht-alloc__row"
        :class="{ 'is-total': b.key === 'TOTAL' }"
      >
        <span class="ht-alloc__label">{{ b.label }}</span>
        <div class="ht-alloc__bar-wrap" aria-hidden="true">
          <div class="ht-alloc__bar" :style="{ width: `${barWidth(b)}%` }" />
        </div>
        <span class="ht-alloc__value">{{ formatPercent(b.pct_of_nav) }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.ht-alloc {
  display: grid;
  gap: var(--space-2);
}

.ht-alloc__tabs {
  display: flex;
  gap: 4px;
  padding: 3px;
  background: var(--bg-table-header);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  align-self: flex-end;
}

.ht-alloc__tab {
  padding: 4px 10px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
}

.ht-alloc__tab:hover {
  background: var(--bg-row-hover);
}

.ht-alloc__tab.is-active {
  background: var(--bg-card);
  border-color: var(--border-default);
  color: var(--text-link);
}

.ht-alloc__head {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-tertiary);
  padding: 0 4px;
}

.ht-alloc__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 6px;
}

.ht-alloc__row {
  display: grid;
  grid-template-columns: 110px 1fr 56px;
  align-items: center;
  gap: 10px;
  font-size: 12px;
}

.ht-alloc__row.is-total {
  border-top: 1px solid var(--border-subtle);
  padding-top: 6px;
  margin-top: 4px;
  font-weight: 600;
  color: var(--text-primary);
}

.ht-alloc__label {
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ht-alloc__bar-wrap {
  height: 6px;
  background: var(--border-subtle);
  border-radius: 3px;
  overflow: hidden;
}

.ht-alloc__bar {
  height: 100%;
  background: var(--color-primary-500, #2563eb);
  border-radius: 3px;
}

.ht-alloc__row.is-total .ht-alloc__bar {
  background: var(--color-primary-700, #1e40af);
}

.ht-alloc__value {
  text-align: right;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
}
</style>
