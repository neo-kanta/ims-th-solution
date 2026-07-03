<script setup lang="ts">
/**
 * Intraday-aware KPI strip — replaces the single "Today NAV" tile with two
 * tiles ("Official NAV" + "Estimated NAV") so the user sees the accounting
 * figure and the live market estimate side-by-side. Official NAV is always
 * the most recent ACCOUNTING_CLOSED snapshot; estimated NAV is the live
 * market data calculation.
 *
 * Falls back to placeholders ("—") when no live data is available — the page
 * never displays a fabricated number.
 */
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import { formatBangkokTime, formatSignedPercent } from "../lib/holdingsFormat";
import type { IntradayValuation, MarketDataStatus } from "../services/intradayValuationApi";

const props = defineProps<{
  valuation: IntradayValuation | null;
  status: MarketDataStatus | null;
  lastRefreshAt: string;
  workflowLabel: string;
  lastSettledDate: string;
}>();

const { t } = useI18n();

const PLACEHOLDER_DASH = "—";

function parseNum(s: string | undefined | null): number {
  if (!s) return 0;
  const n = Number(s);
  return Number.isFinite(n) ? n : 0;
}

function fmtMoney(s: string | undefined | null, digits = 2): string {
  if (!s) return PLACEHOLDER_DASH;
  const n = Number(s);
  if (!Number.isFinite(n)) return PLACEHOLDER_DASH;
  return n.toLocaleString("en-US", {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  });
}

const ccy = computed(() => props.valuation?.valuation_ccy || "THB");

const officialAumDisplay = computed(() =>
  props.valuation?.has_official ? fmtMoney(props.valuation.official_aum, 2) : PLACEHOLDER_DASH,
);

const estimatedAumDisplay = computed(() =>
  props.valuation && parseNum(props.valuation.estimated_aum) !== 0
    ? fmtMoney(props.valuation.estimated_aum, 2)
    : PLACEHOLDER_DASH,
);

const officialNavDisplay = computed(() =>
  props.valuation?.official_nav_per_unit
    ? fmtMoney(props.valuation.official_nav_per_unit, 5)
    : PLACEHOLDER_DASH,
);

const estimatedNavDisplay = computed(() =>
  props.valuation?.estimated_nav_per_unit
    ? fmtMoney(props.valuation.estimated_nav_per_unit, 5)
    : PLACEHOLDER_DASH,
);

const unitsOutDisplay = computed(() =>
  props.valuation?.units_outstanding
    ? fmtMoney(props.valuation.units_outstanding, 2)
    : PLACEHOLDER_DASH,
);

const deltaPctValue = computed(() => parseNum(props.valuation?.delta_pct_vs_last_close));

const deltaTone = computed(() => {
  const v = deltaPctValue.value;
  if (v === 0) return "neutral";
  return v > 0 ? "positive" : "negative";
});

const feedHealthy = computed(() => {
  if (!props.status) return !props.valuation?.is_stale;
  return props.status.healthy;
});

const feedLabel = computed(() => {
  if (feedHealthy.value) return t("holdings.kpis.feedOk", "Feeds OK");
  if (props.status?.note) return props.status.note;
  return t("holdings.kpis.feedStale", "Provider stale");
});

const providerName = computed(() => {
  const p = props.valuation?.primary_provider || props.status?.primary_provider;
  if (!p) return "";
  return p.replace(/_/g, " ");
});

const lastQuoteDisplay = computed(() => {
  if (props.status?.last_quote_at) return formatBangkokTime(props.status.last_quote_at);
  if (props.valuation?.as_of) return formatBangkokTime(props.valuation.as_of);
  return formatBangkokTime(props.lastRefreshAt);
});
</script>

<template>
  <section class="ihc">
    <article class="ihc__card ihc__card--accent-blue">
      <header class="ihc__label">
        <span class="ihc__dot ihc__dot--blue" aria-hidden="true" />
        {{ t("holdings.kpis.officialAum", "Official AUM") }}
      </header>
      <div class="ihc__value">{{ officialAumDisplay }}</div>
      <div class="ihc__meta ihc__meta-muted">{{ ccy }} · {{ t("holdings.kpis.accountingClose", "accounting close") }}</div>
    </article>

    <article class="ihc__card ihc__card--accent-green">
      <header class="ihc__label">
        <span class="ihc__dot ihc__dot--green" aria-hidden="true" />
        {{ t("holdings.kpis.estimatedAum", "Estimated AUM") }}
      </header>
      <div class="ihc__value">{{ estimatedAumDisplay }}</div>
      <div class="ihc__meta" :class="`tone-${deltaTone}`">
        {{ formatSignedPercent(deltaPctValue) }}
        <span class="ihc__meta-muted">{{ t("holdings.kpis.vsLastClose", "vs. last close") }}</span>
      </div>
    </article>

    <article class="ihc__card">
      <header class="ihc__label">{{ t("holdings.kpis.officialUnitNav", "Official Unit NAV") }}</header>
      <div class="ihc__value">{{ officialNavDisplay }}</div>
      <div class="ihc__meta ihc__meta-muted">
        {{ valuation?.official_as_of || lastSettledDate || "—" }}
      </div>
    </article>

    <article class="ihc__card">
      <header class="ihc__label">{{ t("holdings.kpis.estimatedUnitNav", "Estimated Unit NAV") }}</header>
      <div class="ihc__value">{{ estimatedNavDisplay }}</div>
      <div class="ihc__meta ihc__meta-muted">
        {{
          !valuation?.has_units
            ? t("holdings.kpis.fundNotUnitised", "fund is not unitised")
            : valuation?.units_indicative
              ? t("holdings.kpis.unitsIndicative", "indicative · units at par")
              : t("holdings.kpis.fromLivePrices", "from live prices")
        }}
      </div>
    </article>

    <article class="ihc__card">
      <header class="ihc__label">{{ t("holdings.kpis.todayUnitsOut", "Units Outstanding") }}</header>
      <div class="ihc__value">{{ unitsOutDisplay }}</div>
      <div class="ihc__meta ihc__meta-muted">
        {{
          unitsOutDisplay === "—"
            ? t("holdings.kpis.noUnits", "no units")
            : valuation?.units_indicative
              ? t("holdings.kpis.unitsAtPar", "derived at par")
              : t("holdings.kpis.noMovement", "no movement")
        }}
      </div>
    </article>

    <article class="ihc__card">
      <header class="ihc__label">{{ t("holdings.kpis.lastSettled", "Last Settled") }}</header>
      <div class="ihc__value ihc__value--date">{{ lastSettledDate || "—" }}</div>
      <div class="ihc__meta ihc__meta-muted">{{ workflowLabel || "—" }}</div>
    </article>

    <article
      class="ihc__card ihc__card--full"
      :class="feedHealthy ? 'ihc__card--ok' : 'ihc__card--warning'"
    >
      <header class="ihc__label">
        <span
          class="ihc__dot"
          :class="feedHealthy ? 'ihc__dot--success' : 'ihc__dot--warning'"
          aria-hidden="true"
        />
        {{ t("holdings.kpis.feedTitle", "Data Freshness") }}
      </header>
      <div class="ihc__value ihc__value--sm">{{ feedLabel }}</div>
      <div class="ihc__meta ihc__meta-muted">
        <template v-if="providerName">{{ providerName }} · </template>
        {{ t("holdings.kpis.lastQuote", "last quote") }} {{ lastQuoteDisplay }}
      </div>
    </article>
  </section>
</template>

<style scoped>
.ihc {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: var(--space-2);
}

.ihc__card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 14px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  min-width: 0;
}

.ihc__card--accent-blue {
  border-left: 3px solid var(--color-primary-500, #2563eb);
}
.ihc__card--accent-green {
  border-left: 3px solid var(--state-success, #12b76a);
}
.ihc__card--warning {
  background: var(--alert-warning-bg);
  border-color: var(--alert-warning-border);
}
.ihc__card--ok {
  background: var(--alert-success-bg);
  border-color: var(--alert-success-border);
}
.ihc__card--full {
  grid-column: span 1;
}

.ihc__label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}

.ihc__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}
.ihc__dot--blue {
  background: var(--color-primary-500, #2563eb);
}
.ihc__dot--green {
  background: var(--state-success, #12b76a);
}
.ihc__dot--warning {
  background: var(--color-warning-500, #f59e0b);
}
.ihc__dot--success {
  background: var(--color-success-500, #12b76a);
}

.ihc__value {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.ihc__value--date,
.ihc__value--sm {
  font-size: 1.05rem;
}

.ihc__meta {
  font-size: 11px;
  color: var(--text-primary);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.ihc__meta-muted {
  color: var(--text-tertiary);
  font-weight: 400;
}

.tone-positive {
  color: var(--state-success);
}
.tone-negative {
  color: var(--state-danger);
}
.tone-neutral {
  color: var(--text-primary);
}

@media (max-width: 1400px) {
  .ihc {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
@media (max-width: 900px) {
  .ihc {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 480px) {
  .ihc {
    grid-template-columns: 1fr;
  }
}
</style>
