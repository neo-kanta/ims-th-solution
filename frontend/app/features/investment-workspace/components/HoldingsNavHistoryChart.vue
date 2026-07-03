<script setup lang="ts">
import { computed, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import { formatSignedPercent } from "../lib/holdingsFormat";
import type { NavHistoryPayload, NavHistoryRange, NavHistoryPoint, HoldingsSummary, FundFooter } from "../types";
import NavHistoryGraph from "./NavHistoryGraph.vue";
import FullscreenChartOverlay from "./FullscreenChartOverlay.vue";

const props = defineProps<{
  payload: NavHistoryPayload | null;
  range: NavHistoryRange;
  summary?: HoldingsSummary | null;
  footer?: FundFooter | null;
  loading?: boolean;
  error?: string | null;
}>();

const emit = defineEmits<{
  changeRange: [r: NavHistoryRange];
}>();

const { t } = useI18n();

const ranges: NavHistoryRange[] = ["1M", "3M", "6M", "1Y"];

const hoveredPoint = ref<NavHistoryPoint | null>(null);
const isFullscreenOpen = ref(false);

function handleHover(p: NavHistoryPoint) {
  hoveredPoint.value = p;
}

function handleLeave() {
  hoveredPoint.value = null;
}

function openFullscreen() {
  isFullscreenOpen.value = true;
}

function closeFullscreen() {
  isFullscreenOpen.value = false;
}

const subtitle = computed(() => {
  const days = props.payload?.series.length ?? 0;
  const quarter = Math.ceil((new Date().getUTCMonth() + 1) / 3);
  return t(
    "holdings.navHistory.subtitle",
    { days, quarter },
    `${days} business days (Q${quarter})`,
  );
});

function fmt(v: number | undefined): string {
  if (v === undefined || !Number.isFinite(v)) return "—";
  return v.toFixed(4);
}

function formatDateLabel(dateStr: string): string {
  try {
    const parts = dateStr.split("-");
    if (parts.length === 3) {
      const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
      const mIdx = parseInt(parts[1] || "", 10) - 1;
      const monthStr = months[mIdx] || "";
      const month = t(`holdings.months.${monthStr}` as any, monthStr);
      const day = parseInt(parts[2] || "", 10).toString();
      return t("holdings.dateFormats.dayMonth" as any, { month, day }, `${month} ${day}`);
    }
  } catch (e) {}
  return dateStr;
}
</script>

<template>
  <div class="ht-nav">
    <header class="ht-nav__head">
      <div>
        <h3 class="ht-nav__title">{{ t("holdings.navHistory.title", "Unit NAV history") }}</h3>
        <p v-if="hoveredPoint" class="ht-nav__subtitle">
          <span class="ht-nav__active-date">{{ formatDateLabel(hoveredPoint.business_date) }}</span>
          <span class="ht-nav__active-val">{{ t("holdings.navHistory.navLabel", "NAV") }}: {{ hoveredPoint.nav_per_unit.toFixed(4) }}</span>
        </p>
        <p v-else class="ht-nav__subtitle">{{ subtitle }}</p>
      </div>
      <div class="ht-nav__head-right">
        <button
          type="button"
          class="ht-nav__expand-button"
          :aria-label="t('holdings.navHistory.expandAria', 'Expand chart to fullscreen')"
          @click="openFullscreen"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" class="ht-nav__expand-icon">
            <path d="M15 3h6v6M9 21H3v-6M21 3l-7 7M3 21l7-7"/>
          </svg>
        </button>
        <div class="ht-nav__delta" :class="payload && payload.delta_pct >= 0 ? 'tone-positive' : 'tone-negative'">
          <template v-if="payload">{{ formatSignedPercent(payload.delta_pct) }}</template>
          <template v-else>—</template>
        </div>
      </div>
    </header>

    <div class="ht-nav__ranges" role="tablist">
      <button
        v-for="r in ranges"
        :key="r"
        type="button"
        role="tab"
        :aria-selected="range === r"
        class="ht-nav__range"
        :class="{ 'is-active': range === r }"
        @click="emit('changeRange', r)"
      >
        {{ t(`holdings.navHistory.range.${r}` as any, r) }}
      </button>
    </div>

    <!-- Mini/Compact Chart Graph -->
    <NavHistoryGraph
      :payload="payload"
      :view-w="500"
      :view-h="160"
      :pad-top="15"
      :pad-bottom="25"
      :pad-left="15"
      :pad-right="50"
      @hover="handleHover"
      @leave="handleLeave"
    />

    <dl class="ht-nav__stats">
      <div class="ht-nav__stat">
        <dt>{{ t("holdings.navHistory.high", "High") }}</dt>
        <dd>{{ fmt(payload?.high) }}</dd>
      </div>
      <div class="ht-nav__stat">
        <dt>{{ t("holdings.navHistory.low", "Low") }}</dt>
        <dd>{{ fmt(payload?.low) }}</dd>
      </div>
      <div class="ht-nav__stat">
        <dt>{{ t("holdings.navHistory.latest", "Latest") }}</dt>
        <dd>{{ fmt(payload?.latest) }}</dd>
      </div>
    </dl>

    <!-- Teleported Fullscreen Workspace -->
    <FullscreenChartOverlay
      :open="isFullscreenOpen"
      :payload="payload"
      :range="range"
      :summary="summary ?? null"
      :footer="footer ?? null"
      :loading="loading"
      :error="error"
      @close="closeFullscreen"
      @changeRange="emit('changeRange', $event)"
    />
  </div>
</template>

<style scoped>
.ht-nav {
  display: grid;
  gap: var(--space-2);
}

.ht-nav__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
}

.ht-nav__head-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ht-nav__expand-button {
  background: none;
  border: none;
  padding: 4px;
  color: var(--text-tertiary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.15s ease;
}

.ht-nav__expand-button:hover {
  background: var(--bg-row-hover);
  color: var(--text-primary);
}

.ht-nav__expand-icon {
  width: 14px;
  height: 14px;
}

.ht-nav__title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.ht-nav__subtitle {
  margin: 0;
  font-size: 11px;
  color: var(--text-tertiary);
}

.ht-nav__delta {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
  background: var(--status-approved-bg);
  font-variant-numeric: tabular-nums;
}

.tone-positive {
  background: var(--status-approved-bg);
  color: var(--status-approved-text);
}
.tone-negative {
  background: var(--status-rejected-bg);
  color: var(--status-rejected-text);
}

.ht-nav__ranges {
  display: flex;
  gap: 4px;
}

.ht-nav__range {
  height: 22px;
  padding: 0 8px;
  background: var(--bg-table-header);
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-secondary);
  cursor: pointer;
}

.ht-nav__range.is-active {
  background: var(--status-executed-bg);
  border-color: var(--border-focus);
  color: var(--status-executed-text);
}

.ht-nav__active-date {
  font-weight: 600;
  margin-right: 8px;
  color: var(--text-tertiary);
}

.ht-nav__active-val {
  color: var(--text-primary);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.ht-nav__stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-2);
  margin: 0;
}

.ht-nav__stat {
  display: grid;
  gap: 2px;
  padding: 8px 10px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
}

.ht-nav__stat dt {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}

.ht-nav__stat dd {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
}
</style>
