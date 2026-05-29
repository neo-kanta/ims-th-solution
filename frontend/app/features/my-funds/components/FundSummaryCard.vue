<script setup lang="ts">
import { computed } from "vue";

import AppBadge from "~/shared/ui/AppBadge.vue";
import { useI18n } from "~/composables/useI18n";

import { formatMoneyCompact, formatPercent, relativeTime } from "../lib/format";
import type { MyFundCard } from "../types";

import ComplianceBadge from "./ComplianceBadge.vue";
import FundWorkflowBar from "./FundWorkflowBar.vue";
import RoleActionMenu from "./RoleActionMenu.vue";

interface Props {
  card: MyFundCard;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), { loading: false });

const emit = defineEmits<{
  (e: "open", fundId: string): void;
  (e: "preTrade", fundId: string): void;
  (e: "viewBreaches", fundId: string): void;
  (e: "writeResearch", fundId: string): void;
}>();

const { t } = useI18n();

const navLabel = computed(() =>
  formatMoneyCompact(props.card.valuation.nav_numeric, props.card.valuation.valuation_ccy),
);

const aumLabel = computed(() =>
  formatMoneyCompact(props.card.valuation.aum_numeric, props.card.valuation.valuation_ccy),
);

const unrealisedLabel = computed(() =>
  formatMoneyCompact(
    props.card.valuation.unrealised_pnl_numeric,
    props.card.valuation.valuation_ccy,
    true,
  ),
);

const unrealisedTrend = computed<"up" | "down" | "flat">(() => {
  const v = props.card.valuation.unrealised_pnl_numeric;
  if (v > 0) return "up";
  if (v < 0) return "down";
  return "flat";
});

const cashBufferLabel = computed(() => formatPercent(props.card.valuation.cash_buffer_pct, 2));

const roiLabel = computed(() => {
  if (!props.card.valuation.roi) return "—";
  const n = Number(props.card.valuation.roi);
  if (!Number.isFinite(n)) return "—";
  // ROI from valuation snapshots is already a fraction (e.g. 0.0184 → 1.84%).
  return formatPercent(n * 100, 2, true);
});

const updatedLabel = computed(() => relativeTime(props.card.updated_at));

const statusVariant = computed(() => {
  switch (props.card.status) {
    case "ACTIVE":
      return "success";
    case "LOCKED":
      return "locked";
    case "CLOSED":
      return "neutral";
    case "BREACH":
      return "error";
    default:
      return "neutral";
  }
});

const statusLabel = computed(() => {
  switch (props.card.status) {
    case "ACTIVE":
      return t("myFunds.status.active", "Active");
    case "LOCKED":
      return t("myFunds.status.locked", "Locked");
    case "CLOSED":
      return t("myFunds.status.closed", "Closed");
    case "BREACH":
      return t("myFunds.status.breach", "Breach");
    default:
      return props.card.fund_status_raw || "—";
  }
});

const roleVariant = computed(() =>
  props.card.role === "MANAGER" ? "info" : "neutral",
);

const roleLabel = computed(() =>
  props.card.role === "MANAGER"
    ? t("myFunds.roles.manager", "Manager")
    : t("myFunds.roles.member", "Member"),
);

const cashCaption = computed(() => {
  if (props.card.valuation.cash_buffer_pct === null) return t("myFunds.card.cashUnavailable", "no valuation yet");
  const cashStr = formatMoneyCompact(Number(props.card.valuation.cash_balance), props.card.valuation.valuation_ccy);
  return t(
    "myFunds.card.cashHint",
    { value: cashStr },
    `Cash on hand ${cashStr}`,
  );
});

const initials = computed(() => {
  const seed = (props.card.code || props.card.short_name || props.card.fund_id || "?").trim();
  const parts = seed.split(/[-_\s]/).filter(Boolean);
  if (parts.length >= 2) {
    return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
  }
  return seed.slice(0, 2).toUpperCase();
});

function onOpen() {
  emit("open", props.card.fund_id);
}
</script>

<template>
  <article class="fund-card" :class="{ 'is-loading': loading, [`fund-card--${card.status.toLowerCase()}`]: true }" @click="onOpen">
    <header class="fund-card__header">
      <div class="fund-card__identity">
        <span class="fund-card__avatar" :data-status="card.status">{{ initials }}</span>
        <div class="fund-card__id-block">
          <div class="fund-card__codes">
            <span class="fund-card__code">{{ card.code || "—" }}</span>
            <span class="fund-card__sep">/</span>
            <span class="fund-card__slug">{{ card.short_name || card.fund_id.slice(0, 8) }}</span>
          </div>
          <div class="fund-card__name">{{ card.name || card.short_name || card.code }}</div>
          <div class="fund-card__chips">
            <AppBadge size="sm" variant="neutral">{{ t("myFunds.badges.private", "Private") }}</AppBadge>
            <AppBadge size="sm" :variant="roleVariant">{{ roleLabel }}</AppBadge>
            <AppBadge v-if="card.base_currency" size="sm" variant="neutral">{{ card.base_currency }}</AppBadge>
          </div>
        </div>
      </div>
      <AppBadge size="sm" :variant="statusVariant" :dot="card.status === 'BREACH' || card.status === 'LOCKED'">
        {{ statusLabel }}
      </AppBadge>
    </header>

    <section class="fund-card__nav-block">
      <div class="fund-card__nav-row">
        <div>
          <div class="fund-card__metric-label">
            {{ t("myFunds.card.nav", "Net Asset Value") }}
          </div>
          <div class="fund-card__nav-value" :title="card.valuation.nav">
            {{ navLabel }}
          </div>
          <div class="fund-card__nav-hint">
            {{
              card.valuation.business_date
                ? t(
                  "myFunds.card.asOf",
                  { date: card.valuation.business_date },
                  `as of ${card.valuation.business_date}`,
                )
                : t("myFunds.card.noValuation", "No valuation yet")
            }}
          </div>
        </div>
        <div class="fund-card__nav-delta" :data-trend="unrealisedTrend">
          <span class="fund-card__delta-amount">{{ unrealisedLabel }}</span>
          <span class="fund-card__delta-hint">{{
            t("myFunds.card.unrealised", "unrealised P&L")
          }}</span>
        </div>
      </div>
      <div v-if="card.valuation.has_stale_inputs || card.valuation.is_indicative" class="fund-card__warn">
        <span aria-hidden="true">⚠</span>
        {{
          card.valuation.has_stale_inputs
            ? t("myFunds.card.staleInputs", "Stale price inputs detected — values indicative")
            : t("myFunds.card.indicative", "Indicative valuation — not all inputs confirmed")
        }}
      </div>
    </section>

    <section class="fund-card__metrics">
      <div class="fund-card__metric">
        <span class="fund-card__metric-label">{{ t("myFunds.card.aum", "AUM") }}</span>
        <span class="fund-card__metric-value">{{ aumLabel }}</span>
        <span class="fund-card__metric-hint">{{
          card.valuation.portfolio_count
            ? t(
              "myFunds.card.portfolioCount",
              { count: card.valuation.portfolio_count },
              `${card.valuation.portfolio_count} portfolio(s)`,
            )
            : t("myFunds.card.noPortfolios", "No portfolios mapped")
        }}</span>
      </div>
      <div class="fund-card__metric">
        <span class="fund-card__metric-label">{{ t("myFunds.card.cashBuffer", "Cash buffer") }}</span>
        <span class="fund-card__metric-value">{{ cashBufferLabel }}</span>
        <span class="fund-card__metric-hint">{{ cashCaption }}</span>
      </div>
      <div class="fund-card__metric">
        <span class="fund-card__metric-label">{{ t("myFunds.card.roi", "ROI") }}</span>
        <span class="fund-card__metric-value">{{ roiLabel }}</span>
        <span class="fund-card__metric-hint">{{ t("myFunds.card.roiHint", "Latest valuation snapshot") }}</span>
      </div>
    </section>

    <FundWorkflowBar
      :current-state="card.workflow.current_state"
      :available="card.workflow.available"
      :business-date="card.workflow.business_date"
    />

    <footer class="fund-card__footer">
      <div class="fund-card__footer-left">
        <ComplianceBadge :compliance="card.compliance" />
        <span class="fund-card__updated">
          {{ t("myFunds.card.updated", { value: updatedLabel }, `Updated ${updatedLabel}`) }}
        </span>
      </div>
      <RoleActionMenu
        :card="card"
        @open="emit('open', $event)"
        @pre-trade="emit('preTrade', $event)"
        @view-breaches="emit('viewBreaches', $event)"
        @write-research="emit('writeResearch', $event)"
      />
    </footer>
  </article>
</template>

<style scoped>
.fund-card {
  background: var(--bg-card, #ffffff);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 8px);
  padding: var(--space-4) var(--space-4);
  display: grid;
  gap: var(--space-3);
  cursor: pointer;
  transition: border-color 0.12s ease, box-shadow 0.12s ease;
}

.fund-card:hover {
  border-color: var(--state-info, #1f6feb);
  box-shadow: 0 0 0 1px rgba(31, 111, 235, 0.12);
}

.fund-card--breach {
  border-color: rgba(207, 34, 46, 0.45);
}

.fund-card.is-loading {
  opacity: 0.7;
  pointer-events: none;
}

.fund-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-3);
}

.fund-card__identity {
  display: flex;
  gap: var(--space-3);
  align-items: flex-start;
  min-width: 0;
}

.fund-card__avatar {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: var(--bg-card-muted, #eaeef2);
  color: var(--text-primary, #1f2328);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.fund-card__avatar[data-status="ACTIVE"] {
  background: #e6f4ea;
  color: #1f883d;
}

.fund-card__avatar[data-status="LOCKED"] {
  background: #f1f3f5;
  color: #57606a;
}

.fund-card__avatar[data-status="BREACH"] {
  background: #fde6e8;
  color: #cf222e;
}

.fund-card__avatar[data-status="CLOSED"] {
  background: #f1f3f5;
  color: #6e7781;
}

.fund-card__id-block {
  min-width: 0;
}

.fund-card__codes {
  display: inline-flex;
  align-items: baseline;
  gap: 6px;
  font-size: 12px;
  color: var(--state-info, #1f6feb);
  font-weight: 600;
}

.fund-card__sep {
  color: var(--text-tertiary, #6e7781);
}

.fund-card__slug {
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, monospace);
  color: var(--text-secondary, #57606a);
  font-weight: 500;
}

.fund-card__name {
  margin-top: 2px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #1f2328);
  line-height: 1.25;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.fund-card__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 6px;
}

.fund-card__nav-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.fund-card__nav-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: var(--space-3);
}

.fund-card__nav-value {
  font-size: 28px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary, #1f2328);
  line-height: 1.05;
}

.fund-card__nav-hint {
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
}

.fund-card__nav-delta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  font-size: 12px;
  font-weight: 600;
}

.fund-card__nav-delta[data-trend="up"] .fund-card__delta-amount {
  color: var(--state-success, #1f883d);
}

.fund-card__nav-delta[data-trend="down"] .fund-card__delta-amount {
  color: var(--state-danger, #cf222e);
}

.fund-card__nav-delta[data-trend="flat"] .fund-card__delta-amount {
  color: var(--text-tertiary, #6e7781);
}

.fund-card__delta-amount {
  font-size: 14px;
  font-variant-numeric: tabular-nums;
}

.fund-card__delta-hint {
  font-size: 10px;
  font-weight: 500;
  color: var(--text-tertiary, #6e7781);
}

.fund-card__warn {
  font-size: 11px;
  color: var(--state-warning, #9a6700);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.fund-card__metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-3);
}

.fund-card__metric {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.fund-card__metric-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  font-weight: 600;
}

.fund-card__metric-value {
  font-size: 15px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary, #1f2328);
}

.fund-card__metric-hint {
  font-size: 10px;
  color: var(--text-tertiary, #6e7781);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fund-card__footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
  padding-top: var(--space-2);
  border-top: 1px solid var(--border-subtle, #d0d7de);
}

.fund-card__footer-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
  flex-wrap: wrap;
}

.fund-card__updated {
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
}
</style>
