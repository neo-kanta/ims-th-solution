<script setup lang="ts">
/**
 * Allocation donut chart — renders the four allocation dimensions returned by
 *   GET /investment/funds/{id}/allocation
 * as a dependency-free SVG doughnut (consistent with the project's custom-SVG
 * NAV chart; no charting library is pulled in).
 *
 * The donut centre surfaces the live intraday figures the user asked for:
 *   - Estimated AUM (the total the allocation represents)
 *   - Estimated Unit NAV
 *   - Units Outstanding
 *
 * Switching dimensions is a pure client-side concern; the API returns all four
 * breakdowns in a single call. Slices are sorted largest-first and the long
 * tail is folded into an "Other" slice so the ring never degrades into
 * unreadable slivers.
 */
import { computed, ref } from "vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

import { useI18n } from "~/composables/useI18n";
import type {
  AllocationBucket,
  FundAllocation,
} from "~/features/my-funds/services/myFundsApi";
import type { IntradayValuation } from "../services/intradayValuationApi";

const props = defineProps<{
  payload: FundAllocation | null;
  valuation: IntradayValuation | null;
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

// Categorical palette — distinct hues that read on the dark GitHub-style
// surface. The final colour is reserved for the aggregated "Other" slice.
const PALETTE = [
  "#2563eb", // blue
  "#12b76a", // green
  "#f59e0b", // amber
  "#a855f7", // purple
  "#06b6d4", // cyan
  "#ec4899", // pink
  "#84cc16", // lime
  "#64748b", // slate (Other)
];
const OTHER_COLOR = "#94a3b8";

const MAX_SLICES = 7; // top N, remainder folds into "Other"

const rawBuckets = computed<AllocationBucket[]>(() => {
  if (!props.payload) return [];
  switch (activeDimension.value) {
    case "asset_class": return props.payload.by_asset_class ?? [];
    case "sector":      return props.payload.by_sector ?? [];
    case "country":     return props.payload.by_country ?? [];
    case "currency":    return props.payload.by_currency ?? [];
  }
});

function num(s: string | undefined | null): number {
  if (!s) return 0;
  const n = Number(s);
  return Number.isFinite(n) ? n : 0;
}

interface Slice {
  key: string;
  label: string;
  pct: number; // backend pct_of_nav, for the legend
  value: number; // market value, for ring proportion
  color: string;
}

// Sort largest-first, cap to MAX_SLICES, fold the tail into "Other".
const slices = computed<Slice[]>(() => {
  const sorted = [...rawBuckets.value]
    .map((b) => ({
      key: b.key ?? "",
      label: b.label ?? b.key ?? "—",
      pct: num(b.pct_of_nav),
      value: num(b.market_value),
    }))
    .sort((a, b) => b.value - a.value);

  if (sorted.length <= MAX_SLICES) {
    return sorted.map((s, i) => ({ ...s, color: PALETTE[i] ?? OTHER_COLOR }));
  }

  const head = sorted.slice(0, MAX_SLICES);
  const tail = sorted.slice(MAX_SLICES);
  const other = tail.reduce(
    (acc, s) => ({ pct: acc.pct + s.pct, value: acc.value + s.value }),
    { pct: 0, value: 0 },
  );
  return [
    ...head.map((s, i) => ({ ...s, color: PALETTE[i] ?? OTHER_COLOR })),
    {
      key: "__other__",
      label: t("holdings.allocation.other", "Other"),
      pct: Number(other.pct.toFixed(2)),
      value: other.value,
      color: OTHER_COLOR,
    },
  ];
});

const totalValue = computed(() =>
  slices.value.reduce((sum, s) => sum + s.value, 0),
);

// SVG geometry. A 120×120 viewBox keeps the maths simple; the ring is drawn
// with stroke-dasharray on stacked circles rotated to start at 12 o'clock.
const RADIUS = 52;
const CIRCUMFERENCE = 2 * Math.PI * RADIUS;

interface Arc {
  key: string;
  color: string;
  dash: string;
  offset: number;
}

const arcs = computed<Arc[]>(() => {
  const total = totalValue.value;
  if (total <= 0) return [];
  let accumulated = 0;
  return slices.value
    .filter((s) => s.value > 0)
    .map((s) => {
      const fraction = s.value / total;
      const arc: Arc = {
        key: s.key,
        color: s.color,
        dash: `${fraction * CIRCUMFERENCE} ${CIRCUMFERENCE}`,
        offset: -accumulated * CIRCUMFERENCE,
      };
      accumulated += fraction;
      return arc;
    });
});

const isEmpty = computed(
  () => !props.payload || slices.value.length === 0 || totalValue.value <= 0,
);

// ── Centre metrics (live intraday figures) ────────────────────────────────
const ccy = computed(() => props.valuation?.valuation_ccy || props.payload?.valuation_ccy || "THB");

function fmtCompact(s: string | undefined | null): string {
  const n = num(s);
  if (n === 0) return "—";
  return new Intl.NumberFormat("en-US", {
    notation: "compact",
    maximumFractionDigits: 2,
  }).format(n);
}

function fmtPlain(s: string | undefined | null, digits = 2): string {
  if (!s) return "—";
  const n = Number(s);
  if (!Number.isFinite(n) || n === 0) return "—";
  return n.toLocaleString("en-US", {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  });
}

const centreAum = computed(() => fmtCompact(props.valuation?.estimated_aum));
const centreNav = computed(() => fmtPlain(props.valuation?.estimated_nav_per_unit, 4));
const centreUnits = computed(() => fmtCompact(props.valuation?.units_outstanding));

function fmtPct(n: number): string {
  return `${n.toFixed(2)}%`;
}
</script>

<template>
  <div class="adonut">
    <div
      class="adonut__tabs"
      role="tablist"
      :aria-label="t('holdings.allocation.dimensionAria', 'Allocation dimension')"
    >
      <button
        v-for="d in dimensions"
        :key="d.key"
        type="button"
        role="tab"
        :aria-selected="activeDimension === d.key"
        class="adonut__tab"
        :class="{ 'is-active': activeDimension === d.key }"
        @click="activeDimension = d.key"
      >
        {{ d.label }}
      </button>
    </div>

    <AppLoadingState v-if="loading" :message="t('holdings.allocation.loading', 'Loading allocation…')" />

    <p v-else-if="isEmpty" class="adonut__empty">
      {{
        t(
          "holdings.allocation.emptyForDimension",
          "No allocation data for this dimension.",
        )
      }}
    </p>

    <template v-else>
      <div class="adonut__content">
        <div class="adonut__chart-wrap">
          <svg
            class="adonut__svg"
            viewBox="0 0 120 120"
            role="img"
            :aria-label="t('holdings.allocation.donutAria', 'Asset allocation donut chart')"
          >
            <g transform="rotate(-90 60 60)">
              <circle
                v-for="arc in arcs"
                :key="arc.key"
                class="adonut__arc"
                cx="60"
                cy="60"
                :r="RADIUS"
                fill="none"
                :stroke="arc.color"
                stroke-width="14"
                :stroke-dasharray="arc.dash"
                :stroke-dashoffset="arc.offset"
              />
            </g>
          </svg>

          <div class="adonut__centre">
            <span class="adonut__centre-label">{{ t("holdings.allocation.centreAum", "Est. AUM") }}</span>
            <span class="adonut__centre-value">{{ centreAum }}</span>
            <span class="adonut__centre-ccy">{{ ccy }}</span>
          </div>
        </div>

        <ul class="adonut__legend">
          <li v-for="s in slices" :key="s.key" class="adonut__legend-row">
            <span class="adonut__swatch" :style="{ background: s.color }" aria-hidden="true" />
            <span class="adonut__legend-label">{{ s.label }}</span>
            <span class="adonut__legend-pct">{{ fmtPct(s.pct) }}</span>
          </li>
        </ul>
      </div>
    </template>
  </div>
</template>

<style scoped>
.adonut {
  display: grid;
  gap: 16px;
}

.adonut__tabs {
  display: inline-flex;
  align-items: center;
  background: var(--bg-input);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  padding: 2px;
  width: max-content;
}

.adonut__tab {
  padding: 4px 10px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  background: transparent;
  border: 0;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.adonut__tab:hover {
  color: var(--text-primary);
}
.adonut__tab.is-active {
  background: var(--status-executed-bg);
  color: var(--status-executed-text);
}

.adonut__content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.adonut__chart-wrap {
  position: relative;
  width: 130px;
  height: 130px;
  flex-shrink: 0;
}

.adonut__svg {
  width: 100%;
  height: 100%;
  display: block;
}

.adonut__arc {
  transition: stroke-dasharray 0.3s ease, stroke-dashoffset 0.3s ease;
}

.adonut__centre {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1px;
  text-align: center;
  pointer-events: none;
  padding: 0 10%;
}

.adonut__centre-label {
  font-size: 8px;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}

.adonut__centre-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
  line-height: 1.1;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.adonut__centre-ccy {
  font-size: 9px;
  font-weight: 600;
  color: var(--text-tertiary);
}

.adonut__legend {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-grow: 1;
  min-width: 0;
}

.adonut__legend-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.adonut__swatch {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
  flex-shrink: 0;
}

.adonut__legend-label {
  color: var(--text-secondary);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.adonut__legend-pct {
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
  font-weight: 600;
  margin-left: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.adonut__empty {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary);
  padding: 8px 0;
}
</style>
