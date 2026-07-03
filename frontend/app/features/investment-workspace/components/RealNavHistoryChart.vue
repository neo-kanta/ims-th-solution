<script setup lang="ts">
/**
 * Real NAV history chart — backed by
 *   GET /investment/funds/{id}/nav-history?range=1M|3M|6M|1Y|5Y|YTD
 *
 * Renders a minimal inline SVG line chart with high/low/latest tiles below.
 * Uses NAV-per-unit for unitised funds and AUM for non-unitised funds; the
 * label adapts.
 */
import { computed, onMounted, onBeforeUnmount, ref, watch } from "vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

import { useI18n } from "~/composables/useI18n";
import type {
  FundNavHistory,
  FundNavHistoryRange,
} from "~/features/my-funds/services/myFundsApi";

const props = defineProps<{
  payload: FundNavHistory | null;
  range: FundNavHistoryRange;
  loading?: boolean;
}>();

const emit = defineEmits<{
  changeRange: [r: FundNavHistoryRange];
}>();

const { t } = useI18n();

const ranges: FundNavHistoryRange[] = ["1M", "3M", "6M", "1Y"];
const fullscreenRanges: FundNavHistoryRange[] = ["1M", "3M", "6M", "1Y", "5Y", "YTD"];

// Fullscreen overlay — teleported into <body> so it escapes the right-rail
// card's overflow context. Closed on ESC or backdrop click.
const isFullscreen = ref(false);

function openFullscreen() {
  isFullscreen.value = true;
}

function closeFullscreen() {
  isFullscreen.value = false;
}

function onEsc(e: KeyboardEvent) {
  if (e.key === "Escape" && isFullscreen.value) {
    closeFullscreen();
  }
}

onMounted(() => {
  if (typeof window !== "undefined") {
    window.addEventListener("keydown", onEsc);
  }
});
onBeforeUnmount(() => {
  if (typeof window !== "undefined") {
    window.removeEventListener("keydown", onEsc);
  }
});

// Lock body scroll while the overlay is open so the page underneath
// stays still as the user pans/zooms (matches existing overlay UX).
watch(isFullscreen, (open) => {
  if (typeof document === "undefined") return;
  document.body.style.overflow = open ? "hidden" : "";
});

interface Point {
  x: number;
  y: number;
  raw: number;
  date: string;
}

const W = 320;
const H = 96;
const PAD_X = 4;
const PAD_Y = 6;

const numericSeries = computed(() => {
  const p = props.payload;
  if (!p || !p.series || p.series.length === 0) return [] as { date: string; value: number }[];
  return p.series
    .map((s) => {
      const v = s.nav_per_unit && s.nav_per_unit !== ""
        ? Number(s.nav_per_unit)
        : Number(s.aum ?? "0");
      return { date: s.business_date ?? "", value: Number.isFinite(v) ? v : 0 };
    })
    .filter((s) => Number.isFinite(s.value));
});

const linePath = computed(() => {
  const pts = mappedPoints.value;
  if (pts.length === 0) return "";
  return pts
    .map((p, i) => `${i === 0 ? "M" : "L"}${p.x.toFixed(2)},${p.y.toFixed(2)}`)
    .join(" ");
});

const areaPath = computed(() => {
  const pts = mappedPoints.value;
  const last = pts[pts.length - 1];
  const first = pts[0];
  if (!first || !last) return "";
  const baseY = H - PAD_Y;
  const segments = pts.map((p, i) => `${i === 0 ? "M" : "L"}${p.x.toFixed(2)},${p.y.toFixed(2)}`);
  segments.push(`L${last.x.toFixed(2)},${baseY}`);
  segments.push(`L${first.x.toFixed(2)},${baseY}`);
  segments.push("Z");
  return segments.join(" ");
});

const mappedPoints = computed<Point[]>(() => {
  const series = numericSeries.value;
  if (series.length === 0) return [];
  const min = Math.min(...series.map((s) => s.value));
  const max = Math.max(...series.map((s) => s.value));
  const span = max - min || max || 1;
  const usableW = W - PAD_X * 2;
  const usableH = H - PAD_Y * 2;
  return series.map((s, i) => {
    const x = PAD_X + (series.length === 1 ? usableW / 2 : (usableW * i) / (series.length - 1));
    const norm = (s.value - min) / span;
    const y = PAD_Y + (1 - norm) * usableH;
    return { x, y, raw: s.value, date: s.date };
  });
});

const trendTone = computed(() => {
  const d = props.payload?.delta_pct;
  if (!d) return "neutral";
  const n = Number(d);
  if (!Number.isFinite(n) || n === 0) return "neutral";
  return n > 0 ? "positive" : "negative";
});

const deltaLabel = computed(() => {
  const d = props.payload?.delta_pct;
  if (!d) return "0.00%";
  const n = Number(d);
  if (!Number.isFinite(n)) return "0.00%";
  const sign = n > 0 ? "+" : "";
  return `${sign}${n.toFixed(2)}%`;
});

function fmtVal(v: string | undefined, isUnit: boolean): string {
  if (!v) return "—";
  const n = Number(v);
  if (!Number.isFinite(n)) return v;
  return n.toLocaleString("en-US", {
    minimumFractionDigits: isUnit ? 4 : 0,
    maximumFractionDigits: isUnit ? 4 : 0,
  });
}

const seriesLabel = computed(() =>
  props.payload?.has_units
    ? t("holdings.navHistory.unitNavLabel", "NAV / unit")
    : t("holdings.navHistory.aumLabel", "AUM"),
);

const valueIsUnit = computed(() => !!props.payload?.has_units);

const isEmpty = computed(
  () => !props.payload || props.payload.is_empty || mappedPoints.value.length === 0,
);

function onRange(r: FundNavHistoryRange) {
  emit("changeRange", r);
}
</script>

<template>
  <div class="rnav">
    <header class="rnav__header">
      <div class="rnav__title-block">
        <span class="rnav__series">{{ seriesLabel }}</span>
        <span class="rnav__delta" :class="`tone-${trendTone}`">{{ deltaLabel }}</span>
      </div>
      <div class="rnav__header-actions">
        <div class="rnav__ranges" role="tablist">
          <button
            v-for="r in ranges"
            :key="r"
            type="button"
            role="tab"
            :aria-selected="range === r"
            class="rnav__range"
            :class="{ 'is-active': range === r }"
            @click="onRange(r)"
          >
            {{ r }}
          </button>
        </div>
        <button
          type="button"
          class="rnav__expand"
          :aria-label="t('holdings.navHistory.expandAria', 'Expand chart to fullscreen')"
          :title="t('holdings.navHistory.expandTitle', 'Expand')"
          @click="openFullscreen"
        >
          <svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
            <path
              d="M2 2h5v1.5H4.06l3.97 3.97-1.06 1.06L3 4.56V7.5H1.5V2zM9 2h5v5.5h-1.5V4.56l-3.97 3.97-1.06-1.06L11.44 3.5H8.5V2zM3.5 8.5l3.97 3.97V9.53H9v5.47H3.5v-1.5h2.94L2.47 9.56 3.5 8.5zm5.97 0L13.44 12.5H10.5V14h5.5V8.5h-1.5v2.94l-3.97-3.97-1.06 1.03z"
              fill="currentColor"
            />
          </svg>
        </button>
      </div>
    </header>

    <AppLoadingState v-if="loading" :message="t('holdings.navHistory.loading', 'Loading NAV history…')" />

    <p v-else-if="isEmpty" class="rnav__empty">
      {{
        t(
          "holdings.navHistory.empty",
          "No NAV history available for this range.",
        )
      }}
    </p>

    <svg
      v-else
      class="rnav__chart"
      :viewBox="`0 0 ${W} ${H}`"
      preserveAspectRatio="none"
      role="img"
      :aria-label="t('holdings.navHistory.chartAria', 'NAV history line chart')"
    >
      <path :d="areaPath" class="rnav__area" />
      <path :d="linePath" class="rnav__line" />
    </svg>

    <div v-if="!isEmpty" class="rnav__tiles">
      <div class="rnav__tile">
        <span class="rnav__tile-label">{{ t("holdings.navHistory.high", "High") }}</span>
        <span class="rnav__tile-value">{{ fmtVal(payload?.high, valueIsUnit) }}</span>
      </div>
      <div class="rnav__tile">
        <span class="rnav__tile-label">{{ t("holdings.navHistory.low", "Low") }}</span>
        <span class="rnav__tile-value">{{ fmtVal(payload?.low, valueIsUnit) }}</span>
      </div>
      <div class="rnav__tile">
        <span class="rnav__tile-label">{{ t("holdings.navHistory.latest", "Latest") }}</span>
        <span class="rnav__tile-value">{{ fmtVal(payload?.latest, valueIsUnit) }}</span>
      </div>
    </div>

    <!-- Teleported fullscreen overlay (mounted on <body> so it escapes the right-rail card's overflow context) -->
    <Teleport to="body">
      <div
        v-if="isFullscreen"
        class="rnav-fs"
        role="dialog"
        aria-modal="true"
        :aria-label="t('holdings.navHistory.fullscreenAria', 'NAV history — fullscreen')"
        @click.self="closeFullscreen"
      >
        <div class="rnav-fs__panel">
          <header class="rnav-fs__header">
            <div class="rnav-fs__title-block">
              <h2 class="rnav-fs__title">{{ t("holdings.navHistory.title", "Unit NAV history") }}</h2>
              <span class="rnav-fs__series">{{ seriesLabel }}</span>
              <span class="rnav-fs__delta" :class="`tone-${trendTone}`">{{ deltaLabel }}</span>
            </div>
            <div class="rnav-fs__header-actions">
              <div class="rnav-fs__ranges" role="tablist">
                <button
                  v-for="r in fullscreenRanges"
                  :key="r"
                  type="button"
                  role="tab"
                  :aria-selected="range === r"
                  class="rnav-fs__range"
                  :class="{ 'is-active': range === r }"
                  @click="onRange(r)"
                >
                  {{ r }}
                </button>
              </div>
              <button
                type="button"
                class="rnav-fs__close"
                :aria-label="t('holdings.navHistory.closeAria', 'Close fullscreen chart')"
                @click="closeFullscreen"
              >
                <svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
                  <path
                    d="M3.72 3.72a.75.75 0 011.06 0L8 6.94l3.22-3.22a.75.75 0 111.06 1.06L9.06 8l3.22 3.22a.75.75 0 11-1.06 1.06L8 9.06l-3.22 3.22a.75.75 0 01-1.06-1.06L6.94 8 3.72 4.78a.75.75 0 010-1.06z"
                    fill="currentColor"
                  />
                </svg>
              </button>
            </div>
          </header>

          <div class="rnav-fs__body">
            <p v-if="loading" class="rnav__empty rnav__empty--loading">
              {{ t("holdings.navHistory.loading", "Loading NAV history…") }}
            </p>

            <p v-else-if="isEmpty" class="rnav__empty">
              {{
                t(
                  "holdings.navHistory.empty",
                  "No NAV history available for this range.",
                )
              }}
            </p>

            <svg
              v-else
              class="rnav-fs__chart"
              :viewBox="`0 0 ${W} ${H}`"
              preserveAspectRatio="none"
              role="img"
              :aria-label="t('holdings.navHistory.chartAria', 'NAV history line chart')"
            >
              <path :d="areaPath" class="rnav__area" />
              <path :d="linePath" class="rnav__line" />
            </svg>
          </div>

          <footer v-if="!isEmpty" class="rnav-fs__footer">
            <div class="rnav-fs__tile">
              <span class="rnav-fs__tile-label">{{ t("holdings.navHistory.high", "High") }}</span>
              <span class="rnav-fs__tile-value">{{ fmtVal(payload?.high, valueIsUnit) }}</span>
            </div>
            <div class="rnav-fs__tile">
              <span class="rnav-fs__tile-label">{{ t("holdings.navHistory.low", "Low") }}</span>
              <span class="rnav-fs__tile-value">{{ fmtVal(payload?.low, valueIsUnit) }}</span>
            </div>
            <div class="rnav-fs__tile">
              <span class="rnav-fs__tile-label">{{ t("holdings.navHistory.latest", "Latest") }}</span>
              <span class="rnav-fs__tile-value">{{ fmtVal(payload?.latest, valueIsUnit) }}</span>
            </div>
            <div class="rnav-fs__tile">
              <span class="rnav-fs__tile-label">{{ t("holdings.navHistory.points", "Points") }}</span>
              <span class="rnav-fs__tile-value">{{ payload?.series?.length ?? 0 }}</span>
            </div>
            <div class="rnav-fs__tile">
              <span class="rnav-fs__tile-label">{{ t("holdings.navHistory.window", "Window") }}</span>
              <span class="rnav-fs__tile-value">{{ payload?.from }} → {{ payload?.to }}</span>
            </div>
          </footer>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.rnav {
  display: grid;
  gap: 8px;
}

.rnav__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.rnav__title-block {
  display: inline-flex;
  align-items: baseline;
  gap: 8px;
}

.rnav__series {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
  font-weight: 600;
}

.rnav__delta {
  font-size: 12px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.rnav__delta.tone-positive { color: var(--state-success, #1a7f37); }
.rnav__delta.tone-negative { color: var(--state-danger, #cf222e); }
.rnav__delta.tone-neutral  { color: var(--text-secondary); }

.rnav__ranges {
  display: inline-flex;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  padding: 2px;
}

.rnav__range {
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 500;
  color: var(--text-secondary);
  background: transparent;
  border: 0;
  border-radius: 4px;
  cursor: pointer;
}
.rnav__range:hover { color: var(--text-primary); }
.rnav__range.is-active {
  background: var(--bg-card);
  color: var(--text-primary);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
}

.rnav__chart {
  width: 100%;
  height: 96px;
  display: block;
}

.rnav__area {
  fill: var(--color-primary-500, #1f6feb);
  fill-opacity: 0.08;
  stroke: none;
}

.rnav__line {
  fill: none;
  stroke: var(--color-primary-500, #1f6feb);
  stroke-width: 1.6;
  stroke-linejoin: round;
  stroke-linecap: round;
}

.rnav__tiles {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  border-top: 1px solid var(--border-subtle);
  padding-top: 8px;
}

.rnav__tile {
  display: grid;
  gap: 2px;
}

.rnav__tile-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
}

.rnav__tile-value {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
}

.rnav__empty {
  margin: 0;
  font-size: 12px;
  color: var(--text-tertiary);
  padding: 8px 0;
}
.rnav__empty--loading { color: var(--text-secondary); }

.rnav__header-actions {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.rnav__expand {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  padding: 0;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease;
}
.rnav__expand:hover {
  background: var(--action-secondary-hover, var(--surface-2, var(--surface-1)));
  color: var(--text-primary);
  border-color: var(--border-strong, var(--border-default));
}

/* ─── Fullscreen overlay ────────────────────────────────────────────── */
.rnav-fs {
  position: fixed;
  inset: 0;
  background: rgba(15, 17, 21, 0.55);
  z-index: 9999;
  display: grid;
  place-items: center;
  padding: 24px;
  backdrop-filter: blur(2px);
}

.rnav-fs__panel {
  width: min(1100px, 100%);
  max-height: calc(100vh - 48px);
  display: grid;
  grid-template-rows: auto 1fr auto;
  gap: 12px;
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: 10px;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.25);
  padding: 18px 22px;
}

.rnav-fs__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.rnav-fs__title-block {
  display: inline-flex;
  align-items: baseline;
  gap: 12px;
  min-width: 0;
  flex-wrap: wrap;
}

.rnav-fs__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.rnav-fs__series {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
  font-weight: 600;
}

.rnav-fs__delta {
  font-size: 13px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.rnav-fs__delta.tone-positive { color: var(--state-success, #1a7f37); }
.rnav-fs__delta.tone-negative { color: var(--state-danger, #cf222e); }
.rnav-fs__delta.tone-neutral  { color: var(--text-secondary); }

.rnav-fs__header-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.rnav-fs__ranges {
  display: inline-flex;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  padding: 2px;
}

.rnav-fs__range {
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  background: transparent;
  border: 0;
  border-radius: 4px;
  cursor: pointer;
}
.rnav-fs__range:hover { color: var(--text-primary); }
.rnav-fs__range.is-active {
  background: var(--bg-card);
  color: var(--text-primary);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
}

.rnav-fs__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
}
.rnav-fs__close:hover {
  background: var(--action-secondary-hover, var(--surface-2, var(--surface-1)));
  color: var(--text-primary);
}

.rnav-fs__body {
  min-height: 360px;
  display: grid;
  align-items: stretch;
}

.rnav-fs__chart {
  width: 100%;
  height: clamp(340px, 60vh, 560px);
  display: block;
}

.rnav-fs__footer {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
  border-top: 1px solid var(--border-subtle);
  padding-top: 12px;
}

.rnav-fs__tile {
  display: grid;
  gap: 2px;
}

.rnav-fs__tile-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
}

.rnav-fs__tile-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

@media (max-width: 720px) {
  .rnav-fs__footer {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
