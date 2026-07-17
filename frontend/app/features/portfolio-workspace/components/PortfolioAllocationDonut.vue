<script setup lang="ts">
import { computed, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import { formatMoneyCompact } from "~/features/my-funds/lib/format";
import {
  buildAllocationDonutArcs,
  buildAllocationDonutSlices,
  type AllocationDonutItem,
} from "../lib/allocationDonut";

const props = withDefaults(
  defineProps<{
    items: AllocationDonutItem[];
    totalValue?: number;
    currency?: string;
    loading?: boolean;
    compact?: boolean;
  }>(),
  {
    totalValue: undefined,
    currency: "",
    loading: false,
    compact: false,
  },
);

const { t } = useI18n();
const activeKey = ref<string | null>(null);

const RADIUS = 48;
const slices = computed(() =>
  buildAllocationDonutSlices(
    props.items,
    t("portfolio.allocation.other"),
  ),
);
const arcs = computed(() => buildAllocationDonutArcs(slices.value, RADIUS));
const chartTotal = computed(() =>
  slices.value.reduce((sum, slice) => sum + slice.value, 0),
);
const displayTotal = computed(() =>
  Number.isFinite(props.totalValue) && (props.totalValue ?? 0) > 0
    ? (props.totalValue ?? 0)
    : chartTotal.value,
);
const activeSlice = computed(
  () => slices.value.find((slice) => slice.key === activeKey.value) ?? null,
);

watch(slices, (nextSlices) => {
  if (
    activeKey.value &&
    !nextSlices.some((slice) => slice.key === activeKey.value)
  ) {
    activeKey.value = null;
  }
});

function formatPercentage(value: number): string {
  return `${value.toFixed(value >= 10 ? 1 : 2)}%`;
}

function formatValue(value: number): string {
  return formatMoneyCompact(value, props.currency);
}

function setActive(key: string | null) {
  activeKey.value = key;
}

function arcLabel(key: string): string {
  const slice = slices.value.find((item) => item.key === key);
  if (!slice) return "";

  return t(
    "portfolio.allocation.sliceAria",
    {
      label: slice.label,
      percentage: formatPercentage(slice.percentage),
      value: formatValue(slice.value),
    },
    "{label}: {percentage}, {value}",
  );
}
</script>

<template>
  <div
    class="allocation-donut"
    :class="{ 'allocation-donut--compact': compact }"
  >
    <div
      v-if="loading && slices.length === 0"
      class="allocation-donut__state"
      role="status"
    >
      <span class="allocation-donut__spinner" aria-hidden="true" />
      {{ t("portfolio.allocation.loading") }}
    </div>

    <div v-else-if="slices.length === 0" class="allocation-donut__state">
      {{
        t(
          "portfolio.allocation.empty",
          "No valued positions are available yet.",
        )
      }}
    </div>

    <template v-else>
      <div class="allocation-donut__chart-wrap">
        <svg
          class="allocation-donut__chart"
          viewBox="0 0 120 120"
          role="img"
          :aria-label="
            t(
              'portfolio.allocation.chartAria',
              'Portfolio allocation donut chart',
            )
          "
        >
          <circle
            class="allocation-donut__track"
            cx="60"
            cy="60"
            :r="RADIUS"
            fill="none"
            stroke-width="15"
          />
          <g transform="rotate(-90 60 60)">
            <circle
              v-for="arc in arcs"
              :key="arc.key"
              class="allocation-donut__arc"
              :class="{
                'is-active': activeKey === arc.key,
                'is-muted': activeKey !== null && activeKey !== arc.key,
              }"
              cx="60"
              cy="60"
              :r="RADIUS"
              fill="none"
              :stroke="slices.find((slice) => slice.key === arc.key)?.color"
              :stroke-width="activeKey === arc.key ? 18 : 15"
              :stroke-dasharray="arc.dashArray"
              :stroke-dashoffset="arc.dashOffset"
              role="button"
              tabindex="0"
              :aria-label="arcLabel(arc.key)"
              @mouseenter="setActive(arc.key)"
              @mouseleave="setActive(null)"
              @focus="setActive(arc.key)"
              @blur="setActive(null)"
            />
          </g>
        </svg>

        <div class="allocation-donut__centre" aria-live="polite">
          <span class="allocation-donut__centre-label">
            {{ activeSlice?.label ?? t("portfolio.allocation.total") }}
          </span>
          <strong class="allocation-donut__centre-value">
            {{ formatValue(activeSlice?.value ?? displayTotal) }}
          </strong>
          <span class="allocation-donut__centre-meta">
            {{
              activeSlice ? formatPercentage(activeSlice.percentage) : currency
            }}
          </span>
        </div>
      </div>

      <ul class="allocation-donut__legend">
        <li v-for="slice in slices" :key="slice.key">
          <button
            type="button"
            class="allocation-donut__legend-button"
            :class="{
              'is-active': activeKey === slice.key,
              'is-muted': activeKey !== null && activeKey !== slice.key,
            }"
            :aria-label="arcLabel(slice.key)"
            @mouseenter="setActive(slice.key)"
            @mouseleave="setActive(null)"
            @focus="setActive(slice.key)"
            @blur="setActive(null)"
          >
            <span
              class="allocation-donut__swatch"
              :style="{ backgroundColor: slice.color }"
              aria-hidden="true"
            />
            <span class="allocation-donut__legend-copy">
              <span class="allocation-donut__legend-label">{{
                slice.label
              }}</span>
              <span
                v-if="slice.description"
                class="allocation-donut__legend-description"
              >
                {{ slice.description }}
              </span>
            </span>
            <span class="allocation-donut__legend-value">
              {{ formatPercentage(slice.percentage) }}
            </span>
          </button>
        </li>
      </ul>
    </template>
  </div>
</template>

<style scoped>
.allocation-donut {
  display: grid;
  gap: 16px;
  justify-items: center;
  min-width: 0;
}

.allocation-donut__state {
  min-height: 176px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px 8px;
  color: var(--text-secondary, #57606a);
  font-size: 12px;
  text-align: center;
}

.allocation-donut__spinner {
  width: 14px;
  height: 14px;
  border: 2px solid var(--border-subtle, #d0d7de);
  border-top-color: var(--accent-primary, #2563eb);
  border-radius: 50%;
  animation: allocation-donut-spin 0.7s linear infinite;
}

.allocation-donut__chart-wrap {
  position: relative;
  width: min(100%, 180px);
  aspect-ratio: 1;
}

.allocation-donut__chart {
  display: block;
  width: 100%;
  height: 100%;
  overflow: visible;
}

.allocation-donut__track {
  stroke: var(--bg-card-muted, #eef2f6);
}

.allocation-donut__arc {
  cursor: pointer;
  opacity: 1;
  transition:
    opacity 160ms ease,
    stroke-width 160ms ease;
}

.allocation-donut__arc.is-muted {
  opacity: 0.34;
}

.allocation-donut__arc:focus-visible {
  outline: none;
  filter: drop-shadow(0 0 2px var(--text-primary, #1f2328));
}

.allocation-donut__centre {
  position: absolute;
  inset: 24%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 0;
  text-align: center;
  pointer-events: none;
}

.allocation-donut__centre-label {
  width: 100%;
  overflow: hidden;
  color: var(--text-secondary, #57606a);
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.25;
  text-overflow: ellipsis;
  text-transform: uppercase;
  white-space: nowrap;
}

.allocation-donut__centre-value {
  max-width: 100%;
  overflow: hidden;
  color: var(--text-primary, #1f2328);
  font-size: 16px;
  font-variant-numeric: tabular-nums;
  line-height: 1.3;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.allocation-donut__centre-meta {
  color: var(--text-tertiary, #6e7781);
  font-size: 10px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.allocation-donut__legend {
  width: 100%;
  display: grid;
  gap: 2px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.allocation-donut__legend-button {
  width: 100%;
  display: grid;
  grid-template-columns: 9px minmax(0, 1fr) auto;
  align-items: center;
  gap: 8px;
  padding: 6px 7px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
  transition:
    background-color 160ms ease,
    opacity 160ms ease;
}

.allocation-donut__legend-button:hover,
.allocation-donut__legend-button.is-active {
  background: var(--bg-card-muted, #f6f8fa);
}

.allocation-donut__legend-button.is-muted {
  opacity: 0.48;
}

.allocation-donut__legend-button:focus-visible {
  outline: 2px solid var(--accent-primary, #2563eb);
  outline-offset: 1px;
}

.allocation-donut__swatch {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.allocation-donut__legend-copy {
  min-width: 0;
  display: grid;
}

.allocation-donut__legend-label,
.allocation-donut__legend-description {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.allocation-donut__legend-label {
  color: var(--text-primary, #1f2328);
  font-size: 12px;
  font-weight: 650;
}

.allocation-donut__legend-description {
  color: var(--text-tertiary, #6e7781);
  font-size: 10px;
}

.allocation-donut__legend-value {
  color: var(--text-secondary, #57606a);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
}

.allocation-donut--compact {
  grid-template-columns: minmax(104px, 124px) minmax(0, 1fr);
  align-items: center;
  gap: 10px;
}

.allocation-donut--compact .allocation-donut__chart-wrap {
  width: 116px;
}

.allocation-donut--compact .allocation-donut__legend-label {
  font-size: 10px;
}

.allocation-donut--compact .allocation-donut__legend-button {
  padding: 5px 4px;
}

.allocation-donut--compact .allocation-donut__legend-description {
  display: none;
}

@keyframes allocation-donut-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .allocation-donut__spinner {
    animation: none;
  }

  .allocation-donut__arc,
  .allocation-donut__legend-button {
    transition: none;
  }
}

@media (max-width: 360px) {
  .allocation-donut--compact {
    grid-template-columns: 1fr;
  }
}
</style>
