<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import { formatMoneyCompact } from "~/features/my-funds/lib/format";
import type { MyPortfoliosKpiStrip } from "../types";

interface Props {
  kpis: MyPortfoliosKpiStrip;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), { loading: false });
const { t } = useI18n();

const totalAum = computed(() =>
  formatMoneyCompact(props.kpis.total_aum_numeric, props.kpis.valuation_ccy),
);

const totalPnl = computed(() =>
  formatMoneyCompact(
    props.kpis.total_unrealised_pnl_numeric,
    props.kpis.valuation_ccy,
    true,
  ),
);

const pnlTrend = computed(() => props.kpis.unrealised_pnl_trend);

const breachLabel = computed(() => {
  if (!props.kpis.open_breach_count) {
    return t("portfolio.kpis.noBreach", "No open breaches");
  }
  if (props.kpis.worst_breach_severity === "BLOCK") {
    return t("portfolio.kpis.blockerBreach", "Blocker open");
  }
  if (props.kpis.worst_breach_severity === "WARN") {
    return t("portfolio.kpis.warningBreach", "Warnings open");
  }
  return t("portfolio.kpis.infoBreach", "Notices open");
});

const breachVariant = computed(() => {
  if (!props.kpis.open_breach_count) return "muted";
  if (props.kpis.worst_breach_severity === "BLOCK") return "danger";
  if (props.kpis.worst_breach_severity === "WARN") return "warn";
  return "info";
});

const staleLabel = computed(() => {
  if (!props.kpis.stale_count) {
    return t("portfolio.kpis.dataFresh", "All feeds fresh");
  }
  return t(
    "portfolio.kpis.dataStaleCount",
    { count: props.kpis.stale_count },
    `${props.kpis.stale_count} stale`,
  );
});
</script>

<template>
  <div class="kpi-strip" :class="{ 'is-loading': loading }">
    <!-- Total AUM -->
    <div class="kpi-strip__cell">
      <div class="kpi-strip__label">
        <svg class="kpi-strip__label-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="20" x2="18" y2="10"></line>
          <line x1="12" y1="20" x2="12" y2="4"></line>
          <line x1="6" y1="20" x2="6" y2="14"></line>
        </svg>
        <span>{{ t("portfolio.kpis.totalAum", "Total AUM") }}</span>
      </div>
      <div class="kpi-strip__value">{{ totalAum }}</div>
      <div class="kpi-strip__hint">
        {{
          t(
            "portfolio.kpis.totalAumHint",
            { count: kpis.total_count },
            `Across ${kpis.total_count} portfolios`,
          )
        }}
      </div>
    </div>

    <!-- Unrealised P&L -->
    <div class="kpi-strip__cell">
      <div class="kpi-strip__label">
        <svg class="kpi-strip__label-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="23 6 13.5 15.5 8.5 10.5 1 18"></polyline>
          <polyline points="17 6 23 6 23 12"></polyline>
        </svg>
        <span>{{ t("portfolio.kpis.unrealisedPnl", "Unrealised P&L") }}</span>
      </div>
      <div class="kpi-strip__value" :data-trend="pnlTrend">{{ totalPnl }}</div>
      <div class="kpi-strip__hint">
        {{ t("portfolio.kpis.unrealisedHint", "Sum of latest valuations") }}
      </div>
    </div>

    <!-- Active Portfolios -->
    <div class="kpi-strip__cell">
      <div class="kpi-strip__label">
        <svg class="kpi-strip__label-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"></circle>
          <circle cx="12" cy="12" r="6"></circle>
          <circle cx="12" cy="12" r="2"></circle>
        </svg>
        <span>{{ t("portfolio.kpis.activePortfolios", "Active Portfolios") }}</span>
      </div>
      <div class="kpi-strip__value">
        {{ kpis.active_count }}<span class="kpi-strip__sub">/{{ kpis.total_count }}</span>
      </div>
      <div class="kpi-strip__hint">
        {{ kpis.active_count === kpis.total_count ? t("portfolio.kpis.allActive", "All active today") : t("portfolio.kpis.someInactive", "Some inactive/closed") }}
      </div>
    </div>

    <!-- Open Breaches -->
    <div class="kpi-strip__cell">
      <div class="kpi-strip__label">
        <svg class="kpi-strip__label-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
          <line x1="12" y1="9" x2="12" y2="13"></line>
          <line x1="12" y1="17" x2="12.01" y2="17"></line>
        </svg>
        <span>{{ t("portfolio.kpis.openBreaches", "Open Breaches") }}</span>
      </div>
      <div class="kpi-strip__value" :data-variant="breachVariant">
        {{ kpis.open_breach_count }}
      </div>
      <div class="kpi-strip__hint" :data-variant="breachVariant">
        {{ breachLabel }}
      </div>
    </div>

    <!-- Data Health -->
    <div class="kpi-strip__cell">
      <div class="kpi-strip__label">
        <svg class="kpi-strip__label-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
        </svg>
        <span>{{ t("portfolio.kpis.dataHealth", "Data Health") }}</span>
      </div>
      <div class="kpi-strip__value" :data-variant="kpis.stale_count ? 'warn' : 'muted'">
        {{ kpis.stale_count ? kpis.stale_count : "—" }}
      </div>
      <div class="kpi-strip__hint">
        {{ kpis.stale_count ? t("portfolio.kpis.dataStaleCount", { count: kpis.stale_count }, `${kpis.stale_count} stale`) : staleLabel }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.kpi-strip {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: var(--space-3, 12px);
}

.kpi-strip__cell {
  border: 1px solid var(--border-subtle, #e1e8ed);
  border-radius: 8px;
  padding: 12px 16px;
  background: var(--bg-card, #ffffff);
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-height: 84px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
  transition: border-color 0.15s ease, background-color 0.15s ease;
}

.kpi-strip.is-loading .kpi-strip__cell {
  opacity: 0.6;
}

.kpi-strip__label {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--text-secondary, #6e7781);
  display: flex;
  align-items: center;
  gap: 6px;
}

.kpi-strip__label-icon {
  flex-shrink: 0;
  color: var(--text-secondary, #6e7781);
}

.kpi-strip__value {
  font-size: 22px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary, #1f2328);
  line-height: 1.15;
  margin-top: 2px;
}

.kpi-strip__value[data-trend="up"] {
  color: var(--state-success, #1f883d);
}

.kpi-strip__value[data-trend="down"] {
  color: var(--state-danger, #cf222e);
}

.kpi-strip__value[data-variant="danger"] {
  color: var(--state-danger, #cf222e);
}

.kpi-strip__value[data-variant="warn"] {
  color: var(--state-warning, #f08800);
}

.kpi-strip__sub {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary, #6e7781);
}

.kpi-strip__hint {
  font-size: 11px;
  color: var(--text-secondary, #6e7781);
  margin-top: 1px;
}

.kpi-strip__hint[data-variant="danger"] {
  color: var(--state-danger, #cf222e);
}

.kpi-strip__hint[data-variant="warn"] {
  color: var(--state-warning, #d97706);
}

.kpi-strip__hint[data-variant="info"] {
  color: var(--state-info, #1f6feb);
}
</style>
