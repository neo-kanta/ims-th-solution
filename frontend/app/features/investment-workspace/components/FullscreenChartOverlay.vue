<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import { formatSignedPercent } from "../lib/holdingsFormat";
import type { NavHistoryPayload, NavHistoryRange, HoldingsSummary, FundFooter, NavHistoryPoint } from "../types";
import NavHistoryGraph from "./NavHistoryGraph.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

interface Props {
  open: boolean;
  payload: NavHistoryPayload | null;
  range: NavHistoryRange;
  summary: HoldingsSummary | null;
  footer: FundFooter | null;
  loading?: boolean;
  error?: string | null;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  error: null,
});

const emit = defineEmits<{
  close: [];
  changeRange: [r: NavHistoryRange];
}>();

const { t } = useI18n();

const ranges: NavHistoryRange[] = ["1D", "5D", "1M", "3M", "6M", "YTD", "1Y", "5Y", "All"];

const containerRef = ref<HTMLElement | null>(null);
let previouslyFocused: HTMLElement | null = null;

const hoveredPoint = ref<NavHistoryPoint | null>(null);

function handleHover(point: NavHistoryPoint) {
  hoveredPoint.value = point;
}

function handleLeave() {
  hoveredPoint.value = null;
}

// Display NAV
const displayNav = computed(() => {
  if (hoveredPoint.value) {
    return hoveredPoint.value.nav_per_unit.toFixed(5);
  }
  return props.payload?.latest ? props.payload.latest.toFixed(5) : "—";
});

// Display Date
const displayDate = computed(() => {
  const rawDate = hoveredPoint.value
    ? hoveredPoint.value.business_date
    : (props.payload?.series && props.payload.series.length > 0
        ? props.payload.series[props.payload.series.length - 1].business_date
        : props.summary?.business_date || "—");
  return formatDate(rawDate);
});

function formatDate(dateStr: string): string {
  if (!dateStr || dateStr === "—") return "—";
  try {
    const parts = dateStr.split("-");
    if (parts.length === 3) {
      const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
      const mIdx = parseInt(parts[1], 10) - 1;
      const month = t(`holdings.months.${months[mIdx]}` as any, months[mIdx]);
      const day = parseInt(parts[2], 10).toString();
      const year = parts[0];
      return t("holdings.dateFormats.fullDate" as any, { month, day, year }, `${month} ${day}, ${year}`);
    }
  } catch (e) {}
  return dateStr;
}

const isPositive = computed(() => (props.payload?.delta_pct ?? 0) >= 0);

const isStale = computed(() => props.summary?.kpis.stale_warning.is_stale ?? false);

function focusableElements(): HTMLElement[] {
  if (!containerRef.value) return [];
  return Array.from(
    containerRef.value.querySelectorAll<HTMLElement>(
      'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
    ),
  );
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    event.preventDefault();
    emit("close");
    return;
  }

  if (event.key !== "Tab") return;

  const items = focusableElements();
  if (items.length === 0) {
    event.preventDefault();
    containerRef.value?.focus();
    return;
  }

  const first = items[0];
  const last = items[items.length - 1];
  const active = document.activeElement as HTMLElement | null;

  if (event.shiftKey && (active === first || !containerRef.value?.contains(active))) {
    event.preventDefault();
    last.focus();
  } else if (!event.shiftKey && active === last) {
    event.preventDefault();
    first.focus();
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (!import.meta.client) return;
    if (isOpen) {
      document.body.style.overflow = "hidden";
      previouslyFocused = document.activeElement as HTMLElement | null;
      void nextTick(() => {
        containerRef.value?.focus();
        const items = focusableElements();
        items[0]?.focus();
      });
    } else {
      document.body.style.overflow = "";
      if (previouslyFocused && document.body.contains(previouslyFocused)) {
        previouslyFocused.focus();
      }
    }
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  if (import.meta.client) {
    document.body.style.overflow = "";
  }
});
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fc-overlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="fc-overlay-title"
      tabindex="-1"
      ref="containerRef"
      @keydown="handleKeydown"
    >
      <!-- Top header workspace -->
      <header class="fc-overlay__header">
        <div class="fc-overlay__header-left">
          <div class="fc-overlay__brand">
            <h2 id="fc-overlay-title" class="fc-overlay__title">
              {{ summary?.fund.short_name.toUpperCase() ?? 'FUND' }}
            </h2>
            <span class="fc-overlay__meta-badge">
              {{ summary?.fund.code ?? '—' }} ({{ summary?.fund.contract_code ?? '—' }})
            </span>
          </div>

          <div class="fc-overlay__separator" aria-hidden="true" />

          <!-- Dynamic NAV values -->
          <div class="fc-overlay__stats-nav">
            <div class="fc-overlay__nav-val">{{ displayNav }}</div>
            <div 
              class="fc-overlay__nav-delta" 
              :class="isPositive ? 'is-positive' : 'is-negative'"
            >
              {{ payload ? formatSignedPercent(payload.delta_pct) : '—' }}
            </div>
          </div>

          <div class="fc-overlay__separator" aria-hidden="true" />

          <!-- Date as of -->
          <div class="fc-overlay__as-of">
            <span class="fc-overlay__as-of-label">{{ t("holdings.chartOverlay.asOf", "As of:") }}</span>
            <span class="fc-overlay__as-of-value">{{ displayDate }} (GMT+7)</span>
            <span 
              class="fc-overlay__status-dot" 
              :class="isStale ? 'is-stale' : 'is-fresh'" 
              :title="isStale ? t('holdings.chartOverlay.staleData', 'Stale Data') : t('holdings.toolbar.freshOk', 'Feeds OK')"
            />
          </div>
        </div>

        <div class="fc-overlay__header-right">
          <!-- Close button -->
          <button
            type="button"
            class="fc-overlay__close-button"
            :aria-label="t('holdings.chartOverlay.closeAria', 'Close fullscreen chart')"
            @click="emit('close')"
          >
            <AppIcon name="close" class="fc-overlay__close-icon" />
            <span class="fc-overlay__close-text">{{ t("holdings.chartOverlay.close", "Close") }}</span>
          </button>
        </div>
      </header>

      <!-- Range selectors and Toolbar -->
      <div class="fc-overlay__toolbar">
        <div class="fc-overlay__ranges" role="tablist" :aria-label="t('holdings.chartOverlay.rangesAria', 'Chart time ranges')">
          <button
            v-for="r in ranges"
            :key="r"
            type="button"
            role="tab"
            :aria-selected="range === r"
            class="fc-overlay__range-tab"
            :class="{ 'is-active': range === r }"
            @click="emit('changeRange', r)"
          >
            {{ t('holdings.navHistory.range.' + r as any, r) }}
          </button>
        </div>
        <div class="fc-overlay__toolbar-meta">
          <span v-if="loading" class="fc-overlay__loading-spinner">
            {{ t("holdings.chartOverlay.updating", "Updating chart...") }}
          </span>
        </div>
      </div>

      <!-- Main chart workspace -->
      <main class="fc-overlay__chart-workspace">
        <div v-if="error" class="fc-overlay__error-state">
          <AppIcon name="warning" class="fc-overlay__state-icon" />
          <p class="fc-overlay__state-title">{{ t("holdings.chartOverlay.errorTitle", "Failed to load chart data") }}</p>
          <p class="fc-overlay__state-desc">{{ error }}</p>
        </div>
        <div v-else-if="!payload || payload.series.length === 0" class="fc-overlay__empty-state">
          <AppIcon name="info" class="fc-overlay__state-icon" />
          <p class="fc-overlay__state-title">{{ t("holdings.chartOverlay.emptyTitle", "No chart data available") }}</p>
          <p class="fc-overlay__state-desc">{{ t("holdings.chartOverlay.emptyDesc", "There are no NAV records in the selected range.") }}</p>
        </div>
        <div v-else class="fc-overlay__graph-container">
          <!-- Graph is scaled to fit nicely -->
          <NavHistoryGraph
            :payload="payload"
            :view-w="1200"
            :view-h="480"
            :pad-top="20"
            :pad-bottom="35"
            :pad-left="30"
            :pad-right="60"
            @hover="handleHover"
            @leave="handleLeave"
          />
        </div>
      </main>

      <!-- Quick statistics widget -->
      <div class="fc-overlay__stats-strip">
        <div class="fc-overlay__stat-item">
          <span class="fc-overlay__stat-label">{{ t("holdings.navHistory.high", "High").toUpperCase() }}</span>
          <span class="fc-overlay__stat-value">{{ payload?.high ? payload.high.toFixed(4) : '—' }}</span>
        </div>
        <div class="fc-overlay__stat-item">
          <span class="fc-overlay__stat-label">{{ t("holdings.navHistory.low", "Low").toUpperCase() }}</span>
          <span class="fc-overlay__stat-value">{{ payload?.low ? payload.low.toFixed(4) : '—' }}</span>
        </div>
        <div class="fc-overlay__stat-item">
          <span class="fc-overlay__stat-label">{{ t("holdings.navHistory.latest", "Latest").toUpperCase() }}</span>
          <span class="fc-overlay__stat-value">{{ payload?.latest ? payload.latest.toFixed(4) : '—' }}</span>
        </div>
        <div class="fc-overlay__stat-item">
          <span class="fc-overlay__stat-label">{{ t("holdings.chartOverlay.rangeDelta", "RANGE DELTA") }}</span>
          <span 
            class="fc-overlay__stat-value font-semibold"
            :class="isPositive ? 'text-emerald-600 dark:text-emerald-500' : 'text-rose-600 dark:text-rose-500'"
          >
            {{ payload ? formatSignedPercent(payload.delta_pct) : '—' }}
          </span>
        </div>
        <div class="fc-overlay__stat-item">
          <span class="fc-overlay__stat-label">{{ t("holdings.chartOverlay.dataFeed", "DATA FEED") }}</span>
          <span class="fc-overlay__stat-value">{{ t("holdings.chartOverlay.accountingNav", "Accounting Closed NAV") }}</span>
        </div>
      </div>

      <!-- Professional financial terminal status footer -->
      <footer class="fc-overlay__footer">
        <div class="fc-overlay__footer-left">
          <span>{{ t("holdings.chartOverlay.businessDate", "BUSINESS DATE:") }} <strong>{{ summary?.business_date || '—' }}</strong></span>
          <span class="fc-overlay__footer-separator">•</span>
          <span>{{ t("holdings.chartOverlay.lastUpdate", "LAST UPDATE:") }} <strong>{{ footer?.last_priced_at ? formatDate(footer.last_priced_at.split('T')[0]) + ' ' + footer.last_priced_at.split('T')[1].substring(0, 5) : '—' }}</strong></span>
          <span class="fc-overlay__footer-separator">•</span>
          <span>{{ t("holdings.chartOverlay.provider", "PROVIDER:") }} <strong>{{ t("holdings.chartOverlay.providerValue", "TH-IMS Accounting Engine") }}</strong></span>
        </div>
        <div class="fc-overlay__footer-right">
          <span>{{ t("holdings.chartOverlay.irgRuleset", "IRG RULESET:") }} <strong>{{ footer?.irg_rule_version || '—' }}</strong></span>
          <span class="fc-overlay__footer-separator">•</span>
          <span>{{ t("holdings.chartOverlay.auditHash", "AUDIT HASH:") }} <code>{{ footer?.audit_hash || '—' }}</code></span>
        </div>
      </footer>
    </div>
  </Teleport>
</template>

<style scoped>
.fc-overlay {
  position: fixed;
  inset: 0;
  height: 100vh;
  width: 100vw;
  display: flex;
  flex-direction: column;
  background: var(--bg-card, #ffffff);
  z-index: var(--z-modal, 1100);
  font-family: var(--font-sans, Inter, sans-serif);
  color: var(--text-primary, #24292f);
}

.fc-overlay:focus {
  outline: none;
}

/* Header bar styling */
.fc-overlay__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-bottom: 1px solid var(--border-subtle, #eaeef2);
  background: var(--bg-card, #ffffff);
}

.fc-overlay__header-left {
  display: flex;
  align-items: center;
  gap: var(--space-4, 16px);
}

.fc-overlay__brand {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.fc-overlay__title {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary, #24292f);
  letter-spacing: -0.01em;
}

.fc-overlay__meta-badge {
  font-size: 11px;
  color: var(--text-secondary, #57606a);
  font-weight: 500;
}

.fc-overlay__separator {
  width: 1px;
  height: 20px;
  background: var(--border-subtle, #eaeef2);
}

.fc-overlay__stats-nav {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.fc-overlay__nav-val {
  font-size: 18px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary, #24292f);
}

.fc-overlay__nav-delta {
  font-size: 12px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  font-variant-numeric: tabular-nums;
}

.fc-overlay__nav-delta.is-positive {
  background: var(--status-approved-bg, #ecfdf3);
  color: var(--status-approved-text, #027a48);
}

.fc-overlay__nav-delta.is-negative {
  background: var(--status-rejected-bg, #fef3f2);
  color: var(--status-rejected-text, #b42318);
}

.fc-overlay__as-of {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}

.fc-overlay__as-of-label {
  color: var(--text-tertiary, #6e7781);
}

.fc-overlay__as-of-value {
  color: var(--text-secondary, #57606a);
  font-weight: 500;
}

.fc-overlay__status-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  display: inline-block;
}

.fc-overlay__status-dot.is-fresh {
  background: var(--color-success-500, #12b76a);
}

.fc-overlay__status-dot.is-stale {
  background: var(--color-warning-500, #f59e0b);
}

/* Close action */
.fc-overlay__close-button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 28px;
  padding: 0 10px;
  border: 1px solid var(--border-default, #d0d7de);
  background: var(--bg-card, #ffffff);
  color: var(--text-primary, #24292f);
  font-size: 12px;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.fc-overlay__close-button:hover {
  background: var(--bg-row-hover, #f6f8fa);
  border-color: var(--border-strong, #8c959f);
}

.fc-overlay__close-icon {
  width: 14px;
  height: 14px;
}

/* Toolbar & ranges selection */
.fc-overlay__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: var(--bg-table-header, #ffffff);
  border-bottom: 1px solid var(--border-subtle, #eaeef2);
}

.fc-overlay__ranges {
  display: flex;
  gap: 2px;
}

.fc-overlay__range-tab {
  height: 26px;
  padding: 0 10px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-secondary, #57606a);
  font-size: 11px;
  font-weight: 600;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.1s ease;
}

.fc-overlay__range-tab:hover {
  background: var(--bg-row-hover, #f6f8fa);
  color: var(--text-primary, #24292f);
}

.fc-overlay__range-tab.is-active {
  background: var(--bg-card, #ffffff);
  border-color: var(--border-default, #d0d7de);
  color: var(--color-primary-600, #0969da);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.fc-overlay__toolbar-meta {
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
}

/* Chart work area */
.fc-overlay__chart-workspace {
  flex: 1;
  min-height: 0;
  position: relative;
  background: var(--bg-input, #ffffff);
  display: flex;
  flex-direction: column;
}

.fc-overlay__graph-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px;
  min-height: 0;
}

:deep(.ht-nav__chart-container) {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

:deep(.ht-nav__chart) {
  flex: 1;
  min-height: 0;
}

/* States styling */
.fc-overlay__error-state,
.fc-overlay__empty-state {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-8, 32px);
  text-align: center;
}

.fc-overlay__state-icon {
  width: 32px;
  height: 32px;
  color: var(--text-placeholder, #8c959f);
  margin-bottom: var(--space-3, 12px);
}

.fc-overlay__state-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #24292f);
  margin: 0 0 4px 0;
}

.fc-overlay__state-desc {
  font-size: 12px;
  color: var(--text-secondary, #57606a);
  margin: 0;
  max-width: 360px;
}

/* Statistics Strip */
.fc-overlay__stats-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 1px;
  background: var(--border-subtle, #eaeef2);
  border-top: 1px solid var(--border-subtle, #eaeef2);
  border-bottom: 1px solid var(--border-subtle, #eaeef2);
}

.fc-overlay__stat-item {
  flex: 1;
  min-width: 140px;
  background: var(--bg-card, #ffffff);
  padding: 10px 16px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.fc-overlay__stat-label {
  font-size: 9px;
  font-weight: 700;
  color: var(--text-tertiary, #6e7781);
  letter-spacing: 0.05em;
}

.fc-overlay__stat-value {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary, #24292f);
  font-variant-numeric: tabular-nums;
}

/* Footer status bar */
.fc-overlay__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: var(--bg-card, #ffffff);
  font-size: 10px;
  color: var(--text-tertiary, #6e7781);
  letter-spacing: 0.02em;
}

.fc-overlay__footer-left,
.fc-overlay__footer-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.fc-overlay__footer-separator {
  color: var(--border-strong, #8c959f);
  opacity: 0.5;
}

.fc-overlay__footer code {
  font-family: var(--font-mono, monospace);
  background: var(--bg-row-hover, #f6f8fa);
  padding: 1px 4px;
  border-radius: 3px;
  border: 1px solid var(--border-subtle, #eaeef2);
}

@media (max-width: 768px) {
  .fc-overlay__header-left {
    flex-wrap: wrap;
    gap: 8px;
  }
  .fc-overlay__separator {
    display: none;
  }
  .fc-overlay__stats-strip {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
