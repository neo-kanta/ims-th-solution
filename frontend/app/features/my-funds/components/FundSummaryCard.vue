<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";

import { formatMoneyCompact, formatPercent, relativeTime } from "../lib/format";
import type { MyFundCard } from "../types";

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

const cashBufferHint = computed(() => {
  if (props.card.valuation.cash_buffer_pct === null) return "—";
  const cashStr = formatMoneyCompact(Number(props.card.valuation.cash_balance), props.card.valuation.valuation_ccy);
  return `Cash ${cashStr}`;
});

const roiValueLabel = computed(() => roiLabel.value);
const roiTrend = computed(() => {
  if (!props.card.valuation.roi) return "flat";
  const n = Number(props.card.valuation.roi);
  if (n > 0) return "up";
  if (n < 0) return "down";
  return "flat";
});

const complianceBadgeTone = computed(() => {
  if (props.card.valuation.has_stale_inputs) return "stale";
  if (props.card.compliance.open_count > 0) return "danger";
  return "success";
});

const complianceBadgeLabel = computed(() => {
  if (props.card.valuation.has_stale_inputs) {
    return t("myFunds.filters.stale", "Stale NAV");
  }
  if (props.card.compliance.open_count > 0) {
    return t("myFunds.status.breach", "Breached");
  }
  return t("myFunds.compliance.clearShort", "Clear");
});

function onOpen() {
  emit("open", props.card.fund_id);
}
</script>

<template>
  <article
    class="fund-card"
    :class="{
      'is-loading': loading,
      [`fund-card--${card.status.toLowerCase()}`]: true,
      'fund-card--stale': card.valuation.has_stale_inputs
    }"
    @click="onOpen"
  >
    <header class="fund-card__header">
      <div class="fund-card__identity">
        <div class="fund-card__codes">
          <span class="fund-card__code">{{ (card.code || "—").toUpperCase() }}</span>
          <span class="fund-card__sep">/</span>
          <span class="fund-card__slug">{{ (card.short_name || "").toLowerCase() }}</span>
        </div>
        <div class="fund-card__name">{{ card.name || card.short_name || card.code }}</div>
        <div class="fund-card__chips">
          <span class="fund-card__badge-outline">{{ t("myFunds.badges.private", "Private") }}</span>
          <span class="fund-card__badge-outline">{{ roleLabel }}</span>
          <span v-if="card.base_currency" class="fund-card__badge-outline">{{ card.base_currency }}</span>
        </div>
      </div>
      <div class="fund-card__status-badge" :data-status="card.status">
        <span class="fund-card__status-dot"></span>
        <span>{{ statusLabel }}</span>
      </div>
    </header>

    <section class="fund-card__metrics-grid">
      <div class="fund-card__grid-col">
        <div class="fund-card__grid-label">{{ t("myFunds.card.nav", "Net Asset Value") }}</div>
        <div class="fund-card__grid-value" :title="card.valuation.nav">{{ navLabel }}</div>
        <div class="fund-card__grid-hint">
          {{
            card.valuation.business_date
              ? t("myFunds.card.asOf", { date: card.valuation.business_date }, `as of ${card.valuation.business_date}`)
              : t("myFunds.card.noValuation", "No valuation yet")
          }}
        </div>
      </div>

      <div class="fund-card__grid-col">
        <div class="fund-card__grid-label">{{ t("myFunds.card.virtualPnl", "Virtual P&L") }}</div>
        <div class="fund-card__grid-value" :data-trend="unrealisedTrend">{{ unrealisedLabel }}</div>
        <div class="fund-card__grid-hint">{{ t("myFunds.card.vsYest", "vs yest") }}</div>
      </div>

      <div class="fund-card__grid-col">
        <div class="fund-card__grid-label">{{ t("myFunds.card.cashBuffer", "Cash Buffer") }}</div>
        <div class="fund-card__grid-value">{{ cashBufferLabel }}</div>
        <div class="fund-card__grid-hint">{{ cashBufferHint }}</div>
      </div>

      <div class="fund-card__grid-col">
        <div class="fund-card__grid-label">{{ t("myFunds.card.roi", "ROI") }}</div>
        <div class="fund-card__grid-value" :data-trend="roiTrend">{{ roiValueLabel }}</div>
        <div class="fund-card__grid-hint">{{ t("myFunds.card.unadjusted", "unadjusted") }}</div>
      </div>
    </section>

    <div v-if="card.valuation.has_stale_inputs || card.valuation.is_indicative" class="fund-card__warn">
      <span aria-hidden="true">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
          <line x1="12" y1="9" x2="12" y2="13"></line>
          <line x1="12" y1="17" x2="12.01" y2="17"></line>
        </svg>
      </span>
      <span>
        {{
          card.valuation.has_stale_inputs
            ? t("myFunds.card.staleInputs", "Stale price inputs — values indicative")
            : t("myFunds.card.indicative", "Indicative valuation — not all inputs confirmed")
        }}
      </span>
    </div>

    <footer class="fund-card__footer">
      <div class="fund-card__footer-left">
        <div class="fund-card__compliance-badge" :data-tone="complianceBadgeTone">
          <span v-if="complianceBadgeTone === 'success'" class="fund-card__badge-check">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
          </span>
          <span v-else-if="complianceBadgeTone === 'stale'" class="fund-card__badge-warn">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"></circle>
              <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
          </span>
          <span v-else class="fund-card__badge-danger">!</span>
          <span>{{ complianceBadgeLabel }}</span>
        </div>
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
  border: 1px solid #d0d7de;
  border-top: 4px solid #1a7f37; /* Default green top border */
  border-radius: 8px;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  cursor: pointer;
  transition: all 0.15s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
}

.fund-card:hover {
  border-color: #afb8c1;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transform: translateY(-1px);
}

.fund-card--stale {
  border-top-color: #f08800; /* Stale orange top border */
}

.fund-card--breach {
  border-top-color: #cf222e; /* Breach red top border */
}

.fund-card--locked,
.fund-card--closed {
  border-top-color: #8c959f; /* Locked/closed grey top border */
}

.fund-card.is-loading {
  opacity: 0.7;
  pointer-events: none;
}

.fund-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.fund-card__identity {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.fund-card__codes {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #57606a;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.fund-card__sep {
  color: #afb8c1;
}

.fund-card__slug {
  color: #6e7781;
  font-weight: 500;
}

.fund-card__name {
  margin-top: 4px;
  font-size: 16px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.25;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
}

.fund-card__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

/* Badges styling */
.fund-card__badge-outline {
  font-size: 11px;
  font-weight: 600;
  color: #475569;
  border: 1px solid #cbd5e1;
  background: #f8fafc;
  padding: 1px 7px;
  border-radius: 4px;
}

/* Status badge (top right) */
.fund-card__status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 999px;
  background: #e6f4ea;
  color: #137333;
  flex-shrink: 0;
}

.fund-card__status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #137333;
}

.fund-card__status-badge[data-status="BREACH"] {
  background: #fce8e6;
  color: #c5221f;
}
.fund-card__status-badge[data-status="BREACH"] .fund-card__status-dot {
  background: #c5221f;
}

.fund-card__status-badge[data-status="LOCKED"],
.fund-card__status-badge[data-status="CLOSED"] {
  background: #f1f3f4;
  color: #5f6368;
}
.fund-card__status-badge[data-status="LOCKED"] .fund-card__status-dot,
.fund-card__status-badge[data-status="CLOSED"] .fund-card__status-dot {
  background: #5f6368;
}

/* Metrics 4-column Grid */
.fund-card__metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  border-top: 1px dashed #e2e8f0;
  border-bottom: 1px dashed #e2e8f0;
  padding: 12px 0;
}

.fund-card__grid-col {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.fund-card__grid-label {
  font-size: 9px;
  font-weight: 700;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fund-card__grid-value {
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.15;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fund-card__grid-value[data-trend="up"] {
  color: var(--state-success, #1f883d);
}

.fund-card__grid-value[data-trend="down"] {
  color: var(--state-danger, #cf222e);
}

.fund-card__grid-hint {
  font-size: 10px;
  color: #64748b;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Warn/Indicative Banner */
.fund-card__warn {
  font-size: 11px;
  font-weight: 500;
  color: #d97706; /* amber-700 */
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 0;
}

/* Footer layout */
.fund-card__footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  border-top: 1px solid #e2e8f0;
  padding-top: 12px;
  margin-top: 4px;
}

.fund-card__footer-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

/* Compliance Badge (footer left) */
.fund-card__compliance-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 999px;
  border: 1px solid transparent;
}

.fund-card__compliance-badge[data-tone="success"] {
  background: #e2f0d9;
  color: #385723;
  border-color: #c5e0b4;
}

.fund-card__compliance-badge[data-tone="stale"] {
  background: #fdf6e2;
  color: #b25e00;
  border-color: #fce1a6;
}

.fund-card__compliance-badge[data-tone="danger"] {
  background: #fce8e6;
  color: #c5221f;
  border-color: #f5b4ad;
}

.fund-card__badge-check,
.fund-card__badge-warn {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.fund-card__badge-danger {
  font-weight: 900;
  line-height: 1;
}

.fund-card__updated {
  font-size: 11px;
  color: #8c959f;
  font-weight: 500;
}
</style>
