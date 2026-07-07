<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import { formatMoneyCompact, formatPercent, relativeTime } from "~/features/my-funds/lib/format";
import type { MyPortfolioCard } from "../types";
import PortfolioRoleActionMenu from "./PortfolioRoleActionMenu.vue";

interface Props {
  card: MyPortfolioCard;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), { loading: false });

const emit = defineEmits<{
  (e: "open", code: string): void;
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
  return formatPercent(n * 100, 2, true);
});

const updatedLabel = computed(() => relativeTime(props.card.updated_at));

const statusVariant = computed(() => {
  switch (props.card.status) {
    case "ACTIVE":
      return "success";
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
      return t("portfolio.status.active", "Active");
    case "CLOSED":
      return t("portfolio.status.closed", "Closed");
    case "BREACH":
      return t("portfolio.status.breach", "Breach");
    default:
      return props.card.portfolio_status_raw || "—";
  }
});

const roleVariant = computed(() =>
  props.card.role === "MANAGER" ? "info" : "neutral",
);

const roleLabel = computed(() =>
  props.card.role === "MANAGER"
    ? t("portfolio.roles.manager", "Manager")
    : t("portfolio.roles.member", "Member"),
);

const initials = computed(() => {
  const seed = (props.card.code || props.card.name || props.card.portfolio_id || "?").trim();
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
    return t("portfolio.filters.stale", "Stale NAV");
  }
  if (props.card.compliance.open_count > 0) {
    return t("portfolio.status.breach", "Breached");
  }
  return t("portfolio.compliance.clearShort", "Clear");
});

function onOpen() {
  emit("open", props.card.code);
}
</script>

<template>
  <article
    class="portfolio-card"
    :class="{
      'is-loading': loading,
      [`portfolio-card--${card.status.toLowerCase()}`]: true,
      'portfolio-card--stale': card.valuation.has_stale_inputs
    }"
    @click="onOpen"
  >
    <header class="portfolio-card__header">
      <div class="portfolio-card__identity">
        <div class="portfolio-card__codes">
          <span class="portfolio-card__code">{{ (card.code || "—").toUpperCase() }}</span>
          <span v-if="card.portfolio_type" class="portfolio-card__sep">/</span>
          <span v-if="card.portfolio_type" class="portfolio-card__slug">{{ (card.portfolio_type || "").toLowerCase() }}</span>
        </div>
        <div class="portfolio-card__name">{{ card.name || card.code }}</div>
        <div class="portfolio-card__chips">
          <span class="portfolio-card__badge-outline">{{ card.risk_profile ? t(`portfolio.risk.${card.risk_profile.toLowerCase()}`, card.risk_profile) : t("portfolio.badges.private", "Private") }}</span>
          <span class="portfolio-card__badge-outline">{{ roleLabel }}</span>
          <span v-if="card.base_currency" class="portfolio-card__badge-outline">{{ card.base_currency }}</span>
        </div>
      </div>
      <div class="portfolio-card__status-badge" :data-status="card.status">
        <span class="portfolio-card__status-dot"></span>
        <span>{{ statusLabel }}</span>
      </div>
    </header>

    <section class="portfolio-card__metrics-grid">
      <div class="portfolio-card__grid-col">
        <div class="portfolio-card__grid-label">{{ t("portfolio.card.nav", "Net Asset Value") }}</div>
        <div class="portfolio-card__grid-value" :title="card.valuation.nav">{{ navLabel }}</div>
        <div class="portfolio-card__grid-hint">
          {{
            card.valuation.business_date
              ? t("portfolio.card.asOf", { date: card.valuation.business_date }, `as of ${card.valuation.business_date}`)
              : t("portfolio.card.noValuation", "No valuation yet")
          }}
        </div>
      </div>

      <div class="portfolio-card__grid-col">
        <div class="portfolio-card__grid-label">{{ t("portfolio.card.virtualPnl", "Virtual P&L") }}</div>
        <div class="portfolio-card__grid-value" :data-trend="unrealisedTrend">{{ unrealisedLabel }}</div>
        <div class="portfolio-card__grid-hint">{{ t("portfolio.card.vsYest", "vs yest") }}</div>
      </div>

      <div class="portfolio-card__grid-col">
        <div class="portfolio-card__grid-label">{{ t("portfolio.card.cashBuffer", "Cash Buffer") }}</div>
        <div class="portfolio-card__grid-value">{{ cashBufferLabel }}</div>
        <div class="portfolio-card__grid-hint">{{ cashBufferHint }}</div>
      </div>

      <div class="portfolio-card__grid-col">
        <div class="portfolio-card__grid-label">{{ t("portfolio.card.roi", "ROI") }}</div>
        <div class="portfolio-card__grid-value" :data-trend="roiTrend">{{ roiValueLabel }}</div>
        <div class="portfolio-card__grid-hint">{{ t("portfolio.card.unadjusted", "unadjusted") }}</div>
      </div>
    </section>

    <div v-if="card.valuation.has_stale_inputs || card.valuation.is_indicative" class="portfolio-card__warn">
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
            ? t("portfolio.card.staleInputs", "Stale price inputs — values indicative")
            : t("portfolio.card.indicative", "Indicative valuation — not all inputs confirmed")
        }}
      </span>
    </div>

    <footer class="portfolio-card__footer">
      <div class="portfolio-card__footer-left">
        <div class="portfolio-card__compliance-badge" :data-tone="complianceBadgeTone">
          <span v-if="complianceBadgeTone === 'success'" class="portfolio-card__badge-check">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
          </span>
          <span v-else-if="complianceBadgeTone === 'stale'" class="portfolio-card__badge-warn">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"></circle>
              <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
          </span>
          <span v-else class="portfolio-card__badge-danger">!</span>
          <span>{{ complianceBadgeLabel }}</span>
        </div>
        <span class="portfolio-card__updated">
          {{ t("portfolio.card.updated", { value: updatedLabel }, `Updated ${updatedLabel}`) }}
        </span>
      </div>
      <PortfolioRoleActionMenu
        :card="card"
        @open="emit('open', $event)"
        @view-breaches="emit('viewBreaches', $event)"
        @write-research="emit('writeResearch', $event)"
      />
    </footer>
  </article>
</template>

<style scoped>
.portfolio-card {
  background: var(--bg-card, #ffffff);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-top: 4px solid var(--state-success, #1a7f37); /* Default green top border */
  border-radius: 8px;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  cursor: pointer;
  transition: all 0.15s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02);
}

.portfolio-card:hover {
  border-color: #afb8c1;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transform: translateY(-1px);
}

.portfolio-card--stale {
  border-top-color: var(--state-warning, #f08800); /* Stale orange top border */
}

.portfolio-card--breach {
  border-top-color: var(--state-danger, #cf222e); /* Breach red top border */
}

.portfolio-card--locked,
.portfolio-card--closed {
  border-top-color: var(--text-secondary, #8c959f); /* Closed grey top border */
}

.portfolio-card.is-loading {
  opacity: 0.7;
  pointer-events: none;
}

.portfolio-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.portfolio-card__identity {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.portfolio-card__codes {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--text-secondary, #57606a);
  font-weight: 700;
  letter-spacing: 0.02em;
}

.portfolio-card__sep {
  color: #afb8c1;
}

.portfolio-card__slug {
  color: var(--text-secondary, #6e7781);
  font-weight: 500;
}

.portfolio-card__name {
  margin-top: 4px;
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary, #0f172a);
  line-height: 1.25;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
}

.portfolio-card__chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.portfolio-card__badge-outline {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary, #475569);
  border: 1px solid var(--border-subtle, #cbd5e1);
  background: var(--bg-card-muted, #f8fafc);
  padding: 1px 7px;
  border-radius: 4px;
}

.portfolio-card__status-badge {
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

.portfolio-card__status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #137333;
}

.portfolio-card__status-badge[data-status="BREACH"] {
  background: #fce8e6;
  color: #c5221f;
}
.portfolio-card__status-badge[data-status="BREACH"] .portfolio-card__status-dot {
  background: #c5221f;
}

.portfolio-card__status-badge[data-status="CLOSED"] {
  background: #f1f3f4;
  color: #5f6368;
}
.portfolio-card__status-badge[data-status="CLOSED"] .portfolio-card__status-dot {
  background: #5f6368;
}

.portfolio-card__metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
  border-top: 1px dashed var(--border-subtle, #e2e8f0);
  border-bottom: 1px dashed var(--border-subtle, #e2e8f0);
  padding: 12px 0;
}

.portfolio-card__grid-col {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
}

.portfolio-card__grid-label {
  font-size: 9px;
  font-weight: 700;
  color: var(--text-secondary, #64748b);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.portfolio-card__grid-value {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary, #0f172a);
  line-height: 1.15;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.portfolio-card__grid-value[data-trend="up"] {
  color: var(--state-success, #1f883d);
}

.portfolio-card__grid-value[data-trend="down"] {
  color: var(--state-danger, #cf222e);
}

.portfolio-card__grid-hint {
  font-size: 10px;
  color: var(--text-secondary, #64748b);
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.portfolio-card__warn {
  font-size: 11px;
  font-weight: 500;
  color: var(--state-warning, #d97706);
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 0;
}

.portfolio-card__footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  border-top: 1px solid var(--border-subtle, #e2e8f0);
  padding-top: 12px;
  margin-top: 4px;
}

.portfolio-card__footer-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.portfolio-card__compliance-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 999px;
  border: 1px solid transparent;
}

.portfolio-card__compliance-badge[data-tone="success"] {
  background: #e2f0d9;
  color: #385723;
  border-color: #c5e0b4;
}

.portfolio-card__compliance-badge[data-tone="stale"] {
  background: #fdf6e2;
  color: #b25e00;
  border-color: #fce1a6;
}

.portfolio-card__compliance-badge[data-tone="danger"] {
  background: #fce8e6;
  color: #c5221f;
  border-color: #f5b4ad;
}

.portfolio-card__badge-check,
.portfolio-card__badge-warn {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.portfolio-card__badge-danger {
  font-weight: 900;
  line-height: 1;
}

.portfolio-card__updated {
  font-size: 11px;
  color: var(--text-secondary, #8c959f);
  font-weight: 500;
}
</style>
