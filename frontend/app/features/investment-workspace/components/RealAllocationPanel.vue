<script setup lang="ts">
/**
 * Real allocation panel — renders the four dimensions returned by
 *   GET /investment/funds/{id}/allocation
 *
 * Switching dimensions is a pure client-side concern; the API returns all
 * four breakdowns in a single call. Each bar is a percent-of-NAV value so
 * the totals line up at 100% (minus rounding) regardless of which slice
 * is on screen.
 */
import { computed, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import type {
  AllocationBucket,
  FundAllocation,
} from "~/features/my-funds/services/myFundsApi";

const props = defineProps<{
  payload: FundAllocation | null;
  loading?: boolean;
}>();

const { t } = useI18n();

type Dimension = "asset_class" | "sector" | "country" | "currency";

const activeDimension = ref<Dimension>("asset_class");

const dimensions: { key: Dimension; label: string }[] = [
  { key: "asset_class", label: t("holdings.allocation.dimensions.assetClass", "Asset class") },
  { key: "sector",      label: t("holdings.allocation.dimensions.sector", "Sector") },
  { key: "country",     label: t("holdings.allocation.dimensions.country", "Country") },
  { key: "currency",    label: t("holdings.allocation.dimensions.currency", "Currency") },
];

const buckets = computed<AllocationBucket[]>(() => {
  if (!props.payload) return [];
  switch (activeDimension.value) {
    case "asset_class": return props.payload.by_asset_class ?? [];
    case "sector":      return props.payload.by_sector ?? [];
    case "country":     return props.payload.by_country ?? [];
    case "currency":    return props.payload.by_currency ?? [];
  }
});

function fmtPct(s: string | undefined): string {
  if (!s) return "0.00%";
  const n = Number(s);
  if (!Number.isFinite(n)) return s + "%";
  return `${n.toFixed(2)}%`;
}

function barWidth(s: string | undefined): string {
  if (!s) return "0%";
  const n = Number(s);
  if (!Number.isFinite(n) || n <= 0) return "0%";
  if (n >= 100) return "100%";
  return `${n.toFixed(2)}%`;
}

const isEmpty = computed(() => !props.payload || buckets.value.length === 0);
</script>

<template>
  <div class="ralloc">
    <div class="ralloc__tabs" role="tablist" :aria-label="t('holdings.allocation.dimensionAria', 'Allocation dimension')">
      <button
        v-for="d in dimensions"
        :key="d.key"
        type="button"
        role="tab"
        :aria-selected="activeDimension === d.key"
        class="ralloc__tab"
        :class="{ 'is-active': activeDimension === d.key }"
        @click="activeDimension = d.key"
      >
        {{ d.label }}
      </button>
    </div>

    <p v-if="isEmpty && !loading" class="ralloc__empty">
      {{
        t(
          "holdings.allocation.emptyForDimension",
          "No allocation data for this dimension.",
        )
      }}
    </p>

    <p v-if="loading" class="ralloc__empty ralloc__empty--loading">
      {{ t("holdings.allocation.loading", "Loading allocation…") }}
    </p>

    <ul v-if="!isEmpty" class="ralloc__list">
      <li v-for="b in buckets" :key="b.key" class="ralloc__row">
        <div class="ralloc__row-label">{{ b.label }}</div>
        <div class="ralloc__row-bar-wrap">
          <div class="ralloc__row-bar" :style="{ width: barWidth(b.pct_of_nav) }" />
        </div>
        <div class="ralloc__row-pct">{{ fmtPct(b.pct_of_nav) }}</div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.ralloc {
  display: grid;
  gap: 10px;
}

.ralloc__tabs {
  display: inline-flex;
  align-items: center;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  padding: 2px;
  width: max-content;
}

.ralloc__tab {
  padding: 4px 10px;
  font-size: 11px;
  font-weight: 500;
  color: var(--text-secondary);
  background: transparent;
  border: 0;
  border-radius: 4px;
  cursor: pointer;
}
.ralloc__tab:hover {
  color: var(--text-primary);
}
.ralloc__tab.is-active {
  background: var(--bg-card);
  color: var(--text-primary);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
}

.ralloc__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 8px;
}

.ralloc__row {
  display: grid;
  grid-template-columns: minmax(110px, 1fr) minmax(80px, 2fr) auto;
  align-items: center;
  gap: 10px;
  font-size: 12px;
}

.ralloc__row-label {
  color: var(--text-primary);
  font-weight: 500;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ralloc__row-bar-wrap {
  height: 6px;
  background: var(--surface-1);
  border-radius: 3px;
  overflow: hidden;
}

.ralloc__row-bar {
  height: 100%;
  background: var(--color-primary-500, #1f6feb);
  border-radius: 3px;
  transition: width 0.25s ease;
}

.ralloc__row-pct {
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
  font-weight: 500;
  min-width: 56px;
  text-align: right;
}

.ralloc__empty {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary);
  padding: 8px 0;
}
.ralloc__empty--loading {
  color: var(--text-secondary);
}
</style>
