<script setup lang="ts">
import { computed, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import {
  currencySymbol,
  formatMoneyCompact,
  formatPercent,
} from "~/features/my-funds/lib/format";
import type { FundNavHistory } from "~/features/my-funds/services/myFundsApi";
import type { ApiValuationV2 } from "../services/portfolioApi";
import {
  buildValuationHistorySeries,
  buildValuationSparkline,
  valuationHistoryChange,
} from "../lib/valuationHistory";

type HistoryRange = "1M" | "3M" | "6M" | "1Y";

const props = withDefaults(
  defineProps<{
    snapshots?: ApiValuationV2[];
    fundHistory?: FundNavHistory | null;
    currency?: string;
    loading?: boolean;
    range?: HistoryRange;
  }>(),
  {
    snapshots: () => [],
    fundHistory: null,
    currency: "",
    loading: false,
    range: "3M",
  },
);

const emit = defineEmits<{
  rangeChange: [range: HistoryRange];
}>();

const { t, locale } = useI18n();
const rangeDays = { "1M": 30, "3M": 90, "6M": 180, "1Y": 365 } as const;
const ranges = ["1M", "3M", "6M", "1Y"] as const;
const selectedIndex = ref<number | null>(null);

const isUnitNav = computed(() => Boolean(props.fundHistory?.has_units));
const fundSeries = computed(() =>
  buildValuationHistorySeries(
    (props.fundHistory?.series ?? []).map((point) => ({
      business_date: point.business_date,
      aum: isUnitNav.value ? point.nav_per_unit : point.aum,
    })),
    0,
  ),
);
const points = computed(() =>
  fundSeries.value.length > 0
    ? fundSeries.value
    : buildValuationHistorySeries(props.snapshots, rangeDays[props.range]),
);
const chart = computed(() => buildValuationSparkline(points.value, 300, 108));

function numberOrNull(value: string | undefined): number | null {
  if (!value) return null;
  const numeric = Number(value);
  return Number.isFinite(numeric) ? numeric : null;
}

const values = computed(() => points.value.map((point) => point.value));
const high = computed(
  () =>
    numberOrNull(props.fundHistory?.high) ??
    (values.value.length ? Math.max(...values.value) : null),
);
const low = computed(
  () =>
    numberOrNull(props.fundHistory?.low) ??
    (values.value.length ? Math.min(...values.value) : null),
);
const latest = computed(
  () =>
    numberOrNull(props.fundHistory?.latest) ??
    points.value.at(-1)?.value ??
    null,
);
const change = computed(
  () =>
    numberOrNull(props.fundHistory?.delta_pct) ??
    valuationHistoryChange(points.value),
);
const trend = computed(() => {
  if ((change.value ?? 0) > 0) return "up";
  if ((change.value ?? 0) < 0) return "down";
  return "flat";
});
const selectedCoordinate = computed(() =>
  selectedIndex.value === null
    ? null
    : (chart.value.coordinates[selectedIndex.value] ?? null),
);
const tooltipStyle = computed(() => {
  const point = selectedCoordinate.value;
  if (!point) return undefined;
  const horizontal = point.x < 55 ? "0%" : point.x > 245 ? "-100%" : "-50%";
  return {
    left: `${(point.x / 300) * 100}%`,
    top: `${(point.y / 108) * 100}%`,
    transform: `translate(${horizontal}, calc(-100% - 10px))`,
  };
});

watch(points, () => {
  selectedIndex.value = null;
});

function rangeLabel(range: HistoryRange): string {
  return t(`portfolio.terminal.history.ranges.${range}`);
}

function formatValue(value: number | null): string {
  if (value === null || !Number.isFinite(value)) return "—";
  if (!isUnitNav.value) return formatMoneyCompact(value, props.currency);
  return `${currencySymbol(props.currency)}${value.toLocaleString("en-US", {
    minimumFractionDigits: 4,
    maximumFractionDigits: 6,
  })}`;
}

function formatDate(value: string): string {
  const parsed = Date.parse(value);
  if (!Number.isFinite(parsed)) return value;
  const localeCode =
    locale.value === "th" ? "th-TH" : locale.value === "zh" ? "zh-CN" : "en-GB";
  return new Intl.DateTimeFormat(localeCode, {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(parsed);
}

function selectRange(range: HistoryRange) {
  if (range !== props.range) emit("rangeChange", range);
}

function selectFromPointer(event: PointerEvent) {
  const target = event.currentTarget as SVGSVGElement | null;
  if (!target || chart.value.coordinates.length === 0) return;
  const bounds = target.getBoundingClientRect();
  const chartX = ((event.clientX - bounds.left) / bounds.width) * 300;
  let nearestIndex = 0;
  let nearestDistance = Number.POSITIVE_INFINITY;
  chart.value.coordinates.forEach((point, index) => {
    const distance = Math.abs(point.x - chartX);
    if (distance < nearestDistance) {
      nearestDistance = distance;
      nearestIndex = index;
    }
  });
  selectedIndex.value = nearestIndex;
}

function moveSelection(delta: number) {
  const lastIndex = chart.value.coordinates.length - 1;
  if (lastIndex < 0) return;
  const current = selectedIndex.value ?? lastIndex;
  selectedIndex.value = Math.min(lastIndex, Math.max(0, current + delta));
}

function ensureSelection() {
  if (selectedIndex.value === null && chart.value.coordinates.length > 0) {
    selectedIndex.value = chart.value.coordinates.length - 1;
  }
}
</script>

<template>
  <div class="valuation-history">
    <div
      v-if="loading && points.length === 0"
      class="valuation-history__state"
      role="status"
    >
      <span class="valuation-history__spinner" aria-hidden="true" />
      {{ t("portfolio.terminal.history.loading") }}
    </div>

    <div v-else-if="points.length < 2" class="valuation-history__state">
      <svg aria-hidden="true" viewBox="0 0 24 24" width="22" height="22">
        <path
          d="M4 18V6m0 12h16M7 15l4-4 3 2 5-6"
          fill="none"
          stroke="currentColor"
          stroke-width="1.6"
        />
      </svg>
      <span>{{ t("portfolio.terminal.history.empty") }}</span>
    </div>

    <template v-else>
      <div class="valuation-history__headline">
        <strong>{{ formatValue(latest) }}</strong>
        <span :data-trend="trend">{{ formatPercent(change, 2, true) }}</span>
        <small>{{ range }}</small>
        <span
          v-if="loading"
          class="valuation-history__spinner"
          aria-hidden="true"
        />
      </div>

      <div
        class="valuation-history__chart-wrap"
        @pointerleave="selectedIndex = null"
      >
        <svg
          class="valuation-history__chart"
          viewBox="0 0 300 108"
          role="img"
          tabindex="0"
          :aria-label="
            t('portfolio.terminal.history.chartAria', {
              range: rangeLabel(range),
            })
          "
          preserveAspectRatio="none"
          @pointermove="selectFromPointer"
          @focus="ensureSelection"
          @blur="selectedIndex = null"
          @keydown.left.prevent="moveSelection(-1)"
          @keydown.right.prevent="moveSelection(1)"
          @keydown.home.prevent="selectedIndex = 0"
          @keydown.end.prevent="selectedIndex = chart.coordinates.length - 1"
        >
          <defs>
            <linearGradient
              id="portfolio-history-fill"
              x1="0"
              y1="0"
              x2="0"
              y2="1"
            >
              <stop
                offset="0%"
                stop-color="var(--terminal-green, #1f883d)"
                stop-opacity="0.22"
              />
              <stop
                offset="100%"
                stop-color="var(--terminal-green, #1f883d)"
                stop-opacity="0"
              />
            </linearGradient>
          </defs>
          <path
            class="valuation-history__grid"
            d="M0 27H300 M0 54H300 M0 81H300"
          />
          <path :d="chart.areaPath" fill="url(#portfolio-history-fill)" />
          <path class="valuation-history__line" :d="chart.linePath" />
          <template v-if="selectedCoordinate">
            <line
              class="valuation-history__crosshair"
              :x1="selectedCoordinate.x"
              :x2="selectedCoordinate.x"
              y1="4"
              y2="104"
            />
            <circle
              class="valuation-history__selected-point"
              :cx="selectedCoordinate.x"
              :cy="selectedCoordinate.y"
              r="3.5"
            />
          </template>
        </svg>

        <div
          v-if="selectedCoordinate"
          class="valuation-history__tooltip"
          :style="tooltipStyle"
          aria-live="polite"
        >
          <strong>{{ formatValue(selectedCoordinate.value) }}</strong>
          <span>{{ formatDate(selectedCoordinate.date) }}</span>
        </div>
      </div>

      <div
        class="valuation-history__ranges"
        :aria-label="t('portfolio.terminal.history.rangeLabel')"
      >
        <button
          v-for="rangeOption in ranges"
          :key="rangeOption"
          type="button"
          :class="{ 'is-active': range === rangeOption }"
          :aria-pressed="range === rangeOption"
          @click="selectRange(rangeOption)"
        >
          {{ rangeLabel(rangeOption) }}
        </button>
      </div>

      <dl class="valuation-history__stats">
        <div>
          <dt>{{ t("portfolio.terminal.history.high") }}</dt>
          <dd>{{ formatValue(high) }}</dd>
        </div>
        <div>
          <dt>{{ t("portfolio.terminal.history.low") }}</dt>
          <dd>{{ formatValue(low) }}</dd>
        </div>
        <div>
          <dt>{{ t("portfolio.terminal.history.latest") }}</dt>
          <dd>{{ formatValue(latest) }}</dd>
        </div>
      </dl>
    </template>
  </div>
</template>

<style scoped>
.valuation-history {
  min-width: 0;
  padding: 13px;
}
.valuation-history__state {
  min-height: 190px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 18px;
  color: var(--terminal-muted, #8292ab);
  font-size: 11px;
  line-height: 1.5;
  text-align: center;
}
.valuation-history__spinner {
  width: 14px;
  height: 14px;
  display: inline-block;
  flex: 0 0 auto;
  border: 2px solid var(--terminal-border, #29313d);
  border-top-color: var(--terminal-blue, #4b6ff0);
  border-radius: 50%;
  animation: valuation-history-spin 0.7s linear infinite;
}
.valuation-history__headline {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 7px;
}
.valuation-history__headline strong {
  color: var(--terminal-text, #eff5ff);
  font-size: 18px;
  font-variant-numeric: tabular-nums;
}
.valuation-history__headline span:not(.valuation-history__spinner) {
  color: var(--terminal-muted, #8292ab);
  font-family: var(--terminal-mono, ui-monospace, monospace);
  font-size: 10px;
  font-weight: 700;
}
.valuation-history__headline span[data-trend="up"] {
  color: var(--terminal-green, #55db78);
}
.valuation-history__headline span[data-trend="down"] {
  color: var(--terminal-red, #ff6b78);
}
.valuation-history__headline small {
  color: var(--terminal-muted, #8292ab);
  font-size: 9px;
}
.valuation-history__chart-wrap {
  position: relative;
}
.valuation-history__chart {
  display: block;
  width: 100%;
  height: 108px;
  overflow: visible;
  cursor: crosshair;
}
.valuation-history__chart:focus-visible {
  border-radius: 4px;
  outline: 2px solid var(--terminal-blue, #4b6ff0);
  outline-offset: 2px;
}
.valuation-history__grid {
  fill: none;
  stroke: var(--terminal-border, #29313d);
  stroke-opacity: 0.42;
  stroke-width: 0.7;
}
.valuation-history__line {
  fill: none;
  stroke: var(--terminal-green, #55db78);
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 1.5;
}
.valuation-history__crosshair {
  stroke: var(--terminal-muted, #8292ab);
  stroke-dasharray: 2 2;
  stroke-width: 0.8;
}
.valuation-history__selected-point {
  fill: var(--terminal-panel, #ffffff);
  stroke: var(--terminal-green, #55db78);
  stroke-width: 2;
}
.valuation-history__tooltip {
  position: absolute;
  z-index: 2;
  min-width: 94px;
  display: grid;
  gap: 2px;
  padding: 6px 8px;
  border: 1px solid var(--terminal-border, #29313d);
  border-radius: 5px;
  background: var(--terminal-panel, #ffffff);
  box-shadow: 0 4px 14px
    color-mix(in srgb, var(--terminal-text, #1f2328) 15%, transparent);
  pointer-events: none;
  white-space: nowrap;
}
.valuation-history__tooltip strong {
  color: var(--terminal-text, #1f2328);
  font-family: var(--terminal-mono, ui-monospace, monospace);
  font-size: 10px;
}
.valuation-history__tooltip span {
  color: var(--terminal-muted, #57606a);
  font-size: 8px;
}
.valuation-history__ranges {
  display: flex;
  gap: 4px;
  margin-top: 7px;
}
.valuation-history__ranges button {
  min-width: 27px;
  height: 22px;
  padding: 0 7px;
  border: 1px solid var(--terminal-border, #29313d);
  border-radius: 5px;
  background: transparent;
  color: var(--terminal-muted, #8292ab);
  font-family: var(--terminal-mono, ui-monospace, monospace);
  font-size: 9px;
  font-weight: 700;
  cursor: pointer;
}
.valuation-history__ranges button:hover,
.valuation-history__ranges button:focus-visible {
  color: var(--terminal-text, #eff5ff);
  border-color: var(--terminal-blue, #4b6ff0);
  outline: none;
}
.valuation-history__ranges button.is-active {
  border-color: var(--terminal-blue, #4b6ff0);
  background: var(--terminal-blue, #4b6ff0);
  color: var(--text-on-primary, #ffffff);
}
.valuation-history__stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 6px;
  margin: 10px 0 0;
}
.valuation-history__stats > div {
  min-width: 0;
  padding: 8px;
  border: 1px solid var(--terminal-border, #29313d);
  border-radius: 5px;
  background: var(--terminal-raised, #171c24);
}
.valuation-history__stats dt {
  color: var(--terminal-muted, #8292ab);
  font-size: 8px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.valuation-history__stats dd {
  margin: 4px 0 0;
  overflow: hidden;
  color: var(--terminal-text, #eff5ff);
  font-family: var(--terminal-mono, ui-monospace, monospace);
  font-size: 10px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}
@keyframes valuation-history-spin {
  to {
    transform: rotate(360deg);
  }
}
@media (prefers-reduced-motion: reduce) {
  .valuation-history__spinner {
    animation: none;
  }
}
</style>
