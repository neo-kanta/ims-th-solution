<script setup lang="ts">
import type { MarketPricePoint } from "../market-data.types";
import { useI18n } from "~/composables/useI18n";

const props = defineProps<{
  history: MarketPricePoint[];
  loading?: boolean;
  isRunning?: boolean;
}>();

const emit = defineEmits<{
  (e: "import-history-30"): void;
  (e: "import-history-250"): void;
}>();
const { t } = useI18n();

const CHART_WIDTH = 720;
const CHART_HEIGHT = 200;
const PADDING_X = 24;
const PADDING_Y = 18;

const stats = computed(() => {
  const closes = props.history.map((p) => p.close).filter((n) => Number.isFinite(n));
  if (closes.length === 0) return null;
  const min = Math.min(...closes);
  const max = Math.max(...closes);
  const range = max - min || 1;
  return { min, max, range };
});

const polyline = computed(() => {
  const s = stats.value;
  if (!s || props.history.length === 0) return "";
  const innerW = CHART_WIDTH - PADDING_X * 2;
  const innerH = CHART_HEIGHT - PADDING_Y * 2;
  const stepX = props.history.length > 1 ? innerW / (props.history.length - 1) : 0;
  return props.history
    .map((point, i) => {
      const x = (PADDING_X + i * stepX).toFixed(2);
      const y = (PADDING_Y + (1 - (point.close - s.min) / s.range) * innerH).toFixed(2);
      return `${x},${y}`;
    })
    .join(" ");
});

const isPositive = computed(() => {
  if (props.history.length < 2) return true;
  const first = props.history[0]?.close ?? 0;
  const last = props.history[props.history.length - 1]?.close ?? 0;
  return last >= first;
});

const summary = computed(() => {
  if (props.history.length === 0) return null;
  const first = props.history[0];
  const last = props.history[props.history.length - 1];
  if (!first || !last) return null;
  const range = last.close - first.close;
  const pct = first.close !== 0 ? (range / first.close) * 100 : 0;
  return { first, last, range, pct, count: props.history.length };
});

function formatNumber(n: number, opts?: Intl.NumberFormatOptions): string {
  return n.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2, ...opts });
}
function formatDate(s: string): string {
  if (!s) return "";
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  return d.toISOString().slice(0, 10);
}
</script>

<template>
  <article class="card md-sec-card md-sec-history">
    <header class="card-header md-sec-history__head">
      <div>
        <span class="card-title">{{ t("marketData.headings.priceHistory") }}</span>
        <div class="card-subtitle">
          <template v-if="summary">
            {{ summary.count }} bars · {{ formatDate(summary.first.date) }} → {{ formatDate(summary.last.date) }}
          </template>
          <template v-else>{{ t("marketData.messages.noHistoryLoaded") }}</template>
        </div>
      </div>
      <div
        v-if="summary"
        class="md-sec-history__summary md-num"
        :class="isPositive ? 'md-text-success' : 'md-text-danger'"
      >
        {{ summary.range >= 0 ? "+" : "" }}{{ formatNumber(summary.range) }}
        ({{ summary.pct >= 0 ? "+" : "" }}{{ formatNumber(summary.pct) }}%)
      </div>
    </header>

    <div class="card-body">
      <div v-if="history.length > 0" class="md-sec-history__chart-wrap">
        <svg
          class="md-sec-history__chart"
          :class="isPositive ? 'md-sec-history__chart--up' : 'md-sec-history__chart--down'"
          :viewBox="`0 0 ${CHART_WIDTH} ${CHART_HEIGHT}`"
          role="img"
          :aria-label="t('marketData.headings.priceHistory')"
        >
          <polyline
            :points="polyline"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linejoin="round"
          />
        </svg>
        <div v-if="stats" class="md-sec-history__axis md-num">
          <span>{{ formatNumber(stats.min) }}</span>
          <span>{{ formatNumber(stats.max) }}</span>
        </div>
      </div>

      <div v-else-if="loading" class="md-sec-history__empty">
        {{ t("marketData.messages.loadingPriceHistory") }}
      </div>

      <div v-else class="md-sec-history__empty">
        <p class="md-sec-history__empty-title">{{ t("marketData.messages.emptyPriceHistory") }}</p>
        <p class="md-sec-history__empty-text">
          {{ t("marketData.messages.importRecentBars") }}
        </p>
        <div class="md-sec-history__empty-actions">
          <button
            class="btn btn-secondary btn-sm"
            type="button"
            :disabled="isRunning"
            @click="emit('import-history-30')"
          >{{ t("marketData.actions.importHistory30") }}</button>
          <button
            class="btn btn-primary btn-sm"
            type="button"
            :disabled="isRunning"
            @click="emit('import-history-250')"
          >{{ t("marketData.actions.importHistory250") }}</button>
        </div>
      </div>
    </div>
  </article>
</template>

<style scoped>
.md-sec-history__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}
.md-sec-history__summary {
  font-size: 0.95rem;
  font-weight: 600;
}
.md-sec-history__chart-wrap {
  position: relative;
}
.md-sec-history__chart {
  width: 100%;
  height: auto;
  display: block;
  color: var(--md-text-muted);
}
.md-sec-history__chart--up   { color: var(--md-success); }
.md-sec-history__chart--down { color: var(--md-danger); }
.md-sec-history__axis {
  display: flex;
  justify-content: space-between;
  font-size: 0.72rem;
  color: var(--md-text-muted);
  padding: 0 0.4rem;
}
.md-sec-history__empty {
  display: grid;
  gap: 0.4rem;
  padding: 1.5rem 0.5rem;
  text-align: center;
}
.md-sec-history__empty-title {
  margin: 0;
  font-weight: 600;
}
.md-sec-history__empty-text {
  margin: 0;
  font-size: 0.85rem;
  color: var(--md-text-muted);
}
.md-sec-history__empty-actions {
  display: flex;
  gap: 0.4rem;
  justify-content: center;
  flex-wrap: wrap;
  margin-top: 0.6rem;
}
.md-num { font-variant-numeric: tabular-nums; }
</style>
