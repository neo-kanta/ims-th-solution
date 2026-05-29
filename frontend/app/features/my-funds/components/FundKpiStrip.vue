<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";

import { formatMoneyCompact } from "../lib/format";
import type { MyFundsKpiStrip } from "../types";

interface Props {
  kpis: MyFundsKpiStrip;
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
    return t("myFunds.kpis.noBreach", "No open breaches");
  }
  if (props.kpis.worst_breach_severity === "BLOCK") {
    return t("myFunds.kpis.blockerBreach", "Blocker open");
  }
  if (props.kpis.worst_breach_severity === "WARN") {
    return t("myFunds.kpis.warningBreach", "Warnings open");
  }
  return t("myFunds.kpis.infoBreach", "Notices open");
});

const breachVariant = computed(() => {
  if (!props.kpis.open_breach_count) return "muted";
  if (props.kpis.worst_breach_severity === "BLOCK") return "danger";
  if (props.kpis.worst_breach_severity === "WARN") return "warn";
  return "info";
});

const staleLabel = computed(() => {
  if (!props.kpis.stale_count) {
    return t("myFunds.kpis.dataFresh", "All feeds fresh");
  }
  return t(
    "myFunds.kpis.dataStaleCount",
    { count: props.kpis.stale_count },
    `${props.kpis.stale_count} stale`,
  );
});

const pendingLabel = computed(() =>
  props.kpis.pending_workflow_count
    ? t(
      "myFunds.kpis.pendingWithCount",
      { count: props.kpis.pending_workflow_count },
      `${props.kpis.pending_workflow_count} pending`,
    )
    : t("myFunds.kpis.noPending", "No pending stages"),
);
</script>

<template>
  <div class="kpi-strip" :class="{ 'is-loading': loading }">
    <div class="kpi-strip__cell">
      <div class="kpi-strip__label">
        {{ t("myFunds.kpis.totalAum", "Total AUM") }}
      </div>
      <div class="kpi-strip__value">{{ totalAum }}</div>
      <div class="kpi-strip__hint">
        {{
          t(
            "myFunds.kpis.totalAumHint",
            { count: kpis.total_count },
            `Across ${kpis.total_count} accessible fund(s)`,
          )
        }}
      </div>
    </div>

    <div class="kpi-strip__cell">
      <div class="kpi-strip__label">
        {{ t("myFunds.kpis.unrealisedPnl", "Unrealised P&L") }}
      </div>
      <div class="kpi-strip__value" :data-trend="pnlTrend">{{ totalPnl }}</div>
      <div class="kpi-strip__hint">
        {{ t("myFunds.kpis.unrealisedHint", "Sum of latest valuations") }}
      </div>
    </div>

    <div class="kpi-strip__cell">
      <div class="kpi-strip__label">
        {{ t("myFunds.kpis.activeContracts", "Active Contracts") }}
      </div>
      <div class="kpi-strip__value">
        {{ kpis.active_count }}
        <span class="kpi-strip__sub">/ {{ kpis.total_count }}</span>
      </div>
      <div class="kpi-strip__hint">
        {{ pendingLabel }}
      </div>
    </div>

    <div class="kpi-strip__cell">
      <div class="kpi-strip__label">
        {{ t("myFunds.kpis.openBreaches", "Open Breaches") }}
      </div>
      <div class="kpi-strip__value" :data-variant="breachVariant">
        {{ kpis.open_breach_count }}
      </div>
      <div class="kpi-strip__hint" :data-variant="breachVariant">
        {{ breachLabel }}
      </div>
    </div>

    <div class="kpi-strip__cell">
      <div class="kpi-strip__label">
        {{ t("myFunds.kpis.dataHealth", "Data Health") }}
      </div>
      <div class="kpi-strip__value">
        {{ kpis.stale_count ? kpis.stale_count : "—" }}
      </div>
      <div class="kpi-strip__hint">
        {{ staleLabel }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.kpi-strip {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: var(--space-3);
}

.kpi-strip__cell {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 6px);
  padding: var(--space-3) var(--space-4);
  background: var(--bg-card, #ffffff);
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-height: 84px;
}

.kpi-strip.is-loading .kpi-strip__cell {
  opacity: 0.6;
}

.kpi-strip__label {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-secondary, #57606a);
}

.kpi-strip__value {
  font-size: 22px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary, #1f2328);
  line-height: 1.15;
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
  color: var(--state-warning, #9a6700);
}

.kpi-strip__value[data-variant="info"] {
  color: var(--state-info, #1f6feb);
}

.kpi-strip__sub {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-tertiary, #6e7781);
}

.kpi-strip__hint {
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
}

.kpi-strip__hint[data-variant="danger"] {
  color: var(--state-danger, #cf222e);
}

.kpi-strip__hint[data-variant="warn"] {
  color: var(--state-warning, #9a6700);
}

.kpi-strip__hint[data-variant="info"] {
  color: var(--state-info, #1f6feb);
}
</style>
