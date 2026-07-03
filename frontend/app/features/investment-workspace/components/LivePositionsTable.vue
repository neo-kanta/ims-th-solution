<script setup lang="ts">
/**
 * Live positions table — driven by the intraday valuation endpoint
 *   GET /investment/funds/{id}/holdings/valuation
 *
 * Adds latest price, market value, and unrealised P&L columns to the
 * existing positions view. The provider freshness badge on each row tells
 * the user whether the displayed number came from a fresh provider quote or
 * a stale cached snapshot.
 *
 * Cost basis, quantity, and average cost still come from the live ledger
 * projection — exactly the same fields the original RealPositionsTable
 * rendered, so this component is a drop-in replacement.
 */
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import type {
  IntradayCashRow,
  IntradayPosition,
} from "../services/intradayValuationApi";

const props = defineProps<{
  positions: IntradayPosition[];
  cashRows: IntradayCashRow[];
  valuationCcy: string;
  loading?: boolean;
}>();

const { t } = useI18n();

function parseNum(v: string | undefined | null): number {
  if (!v) return 0;
  const n = Number(v);
  return Number.isFinite(n) ? n : 0;
}

function fmtQty(v: number): string {
  return v.toLocaleString("en-US", {
    maximumFractionDigits: 2,
    minimumFractionDigits: 0,
  });
}

function fmtMoney(v: number, digits = 2): string {
  return v.toLocaleString("en-US", {
    maximumFractionDigits: digits,
    minimumFractionDigits: digits,
  });
}

function fmtSignedPct(raw: string | undefined): string {
  const n = parseNum(raw);
  if (n === 0) return "0.00%";
  const sign = n > 0 ? "+" : "";
  return `${sign}${n.toFixed(2)}%`;
}

const totalCashAmount = computed(() => {
  return props.cashRows.reduce((sum, c) => sum + parseNum(c.balance), 0);
});

const totalCostBasis = computed(() =>
  props.positions.reduce((sum, p) => sum + parseNum(p.cost_basis), 0) + totalCashAmount.value,
);

const totalMarketValue = computed(() =>
  props.positions.reduce((sum, p) => sum + parseNum(p.market_value), 0) + totalCashAmount.value,
);

const totalUnrealisedPnL = computed(() =>
  props.positions.reduce((sum, p) => sum + parseNum(p.unrealised_pnl), 0),
);

const totalCashByCurrency = computed(() => {
  const m = new Map<string, number>();
  for (const c of props.cashRows) {
    const cur = c.currency || "";
    if (!cur) continue;
    m.set(cur, (m.get(cur) ?? 0) + parseNum(c.balance));
  }
  return m;
});

function pnlTone(raw: string | undefined): string {
  const n = parseNum(raw);
  if (n === 0) return "neutral";
  return n > 0 ? "positive" : "negative";
}

const sourceLabels: Record<string, string> = {
  live_quote: "Live quote",
  market_data_snapshot: "Market data",
  official_price_snapshot: "Official price",
  cost_carry: "Carried at cost",
};

function sourceLabel(source: string | undefined): string {
  if (!source) return "";
  return sourceLabels[source] ?? source;
}
</script>

<template>
  <div class="lp-wrap">
    <table v-if="positions.length || cashRows.length" class="lp-table">
      <thead>
        <tr class="lp-th-group-row">
          <th colspan="3" class="lp-th lp-th--group lp-th--left">Instrument</th>
          <th colspan="3" class="lp-th lp-th--group lp-th--right">Position</th>
          <th colspan="2" class="lp-th lp-th--group lp-th--right">Valuation</th>
          <th colspan="2" class="lp-th lp-th--group lp-th-group-last">P&L / Feed</th>
        </tr>
        <tr>
          <th class="lp-th lp-th--left">{{ t("holdings.positions.ticker", "Ticker") }}</th>
          <th class="lp-th lp-th--left">{{ t("holdings.positions.name", "Name") }}</th>
          <th class="lp-th lp-th--left">{{ t("holdings.positions.assetClass", "Asset class") }}</th>
          <th class="lp-th lp-th--right">{{ t("holdings.positions.quantity", "Quantity") }}</th>
          <th class="lp-th lp-th--right">{{ t("holdings.positions.avgCost", "Avg cost") }}</th>
          <th class="lp-th lp-th--right">{{ t("holdings.positions.latestPrice", "Latest price") }}</th>
          <th class="lp-th lp-th--right">{{ t("holdings.positions.marketValue", "Market value") }}</th>
          <th class="lp-th lp-th--right">{{ t("holdings.positions.costBasis", "Cost basis") }}</th>
          <th class="lp-th lp-th--right">{{ t("holdings.positions.unrealisedPnL", "Unrealised P&L") }}</th>
          <th class="lp-th lp-th--center">{{ t("holdings.positions.feed", "Feed") }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in positions" :key="row.instrument_id" class="lp-row">
          <td class="lp-td lp-td--mono">
            <span class="lp-ticker-text">{{ row.ticker || "—" }}</span>
            <span v-if="row.ticker && row.ticker !== 'CASH'" class="lp-ticker-badge">SET</span>
          </td>
          <td class="lp-td">{{ row.name || "—" }}</td>
          <td class="lp-td">
            <span class="lp-chip">{{ row.asset_class_label || "—" }}</span>
          </td>
          <td class="lp-td lp-td--num">{{ fmtQty(parseNum(row.quantity)) }}</td>
          <td class="lp-td lp-td--num">{{ fmtMoney(parseNum(row.average_cost), 2) }}</td>
          <td class="lp-td lp-td--num">{{ fmtMoney(parseNum(row.latest_price), 2) }}</td>
          <td class="lp-td lp-td--num">{{ fmtMoney(parseNum(row.market_value), 2) }}</td>
          <td class="lp-td lp-td--num">{{ fmtMoney(parseNum(row.cost_basis), 2) }}</td>
          <td class="lp-td lp-td--num" :class="`tone-${pnlTone(row.unrealised_pnl)}`">
            <span class="lp-pnl-val">{{ fmtMoney(parseNum(row.unrealised_pnl), 2) }}</span>
            <span v-if="row.unrealised_pnl_pct" class="lp-td__sub">
              {{ fmtSignedPct(row.unrealised_pnl_pct) }}
            </span>
          </td>
          <td class="lp-td lp-td--center">
            <span
              class="lp-feed"
              :class="row.is_stale ? 'lp-feed--stale' : 'lp-feed--live'"
              :title="row.stale_reason || row.provider || ''"
            >
              <span class="lp-feed-dot" />
              {{
                row.is_stale
                  ? t("holdings.positions.feedStale", "stale")
                  : row.provider || t("holdings.positions.feedLive", "live")
              }}
            </span>
            <span v-if="sourceLabel(row.source)" class="lp-td__sub">{{ sourceLabel(row.source) }}</span>
            <span v-if="row.price_effective_date" class="lp-td__sub">{{ row.price_effective_date }}</span>
          </td>
        </tr>

        <tr
          v-for="[ccy, amount] in [...totalCashByCurrency.entries()]"
          :key="`cash-${ccy}`"
          class="lp-row lp-row--cash"
        >
          <td class="lp-td lp-td--mono">
            <span class="lp-ticker-text">{{ t("holdings.positions.cashLabel", "CASH") }}</span>
          </td>
          <td class="lp-td">{{ t("holdings.positions.cashEquivalents", "Cash & equivalents") }}</td>
          <td class="lp-td">
            <span class="lp-chip lp-chip--cash">{{ ccy }}</span>
          </td>
          <td class="lp-td lp-td--num">—</td>
          <td class="lp-td lp-td--num">—</td>
          <td class="lp-td lp-td--num">—</td>
          <td class="lp-td lp-td--num">{{ fmtMoney(amount, 2) }}</td>
          <td class="lp-td lp-td--num">{{ fmtMoney(amount, 2) }}</td>
          <td class="lp-td lp-td--num">—</td>
          <td class="lp-td lp-td--center">
            <span class="lp-feed lp-feed--live">
              <span class="lp-feed-dot lp-feed-dot--settled" />
              {{ t("holdings.positions.settled", "Settled") }}
            </span>
          </td>
        </tr>

        <tr class="lp-row lp-row--subtotal">
          <td class="lp-td lp-td--mono" colspan="3">
            {{ t("holdings.positions.totalInclCash", "TOTAL INCL. CASH") }}
          </td>
          <td class="lp-td lp-td--num" colspan="3"></td>
          <td class="lp-td lp-td--num">{{ fmtMoney(totalMarketValue, 2) }}</td>
          <td class="lp-td lp-td--num">{{ fmtMoney(totalCostBasis, 2) }}</td>
          <td class="lp-td lp-td--num" :class="`tone-${totalUnrealisedPnL >= 0 ? 'positive' : 'negative'}`">
            {{ totalUnrealisedPnL >= 0 ? '+' : '' }}{{ fmtMoney(totalUnrealisedPnL, 2) }}
          </td>
          <td class="lp-td"></td>
        </tr>
      </tbody>
    </table>

    <p v-else-if="!loading" class="lp-empty">
      {{
        t(
          "holdings.positions.empty",
          "No positions have been posted to this portfolio yet.",
        )
      }}
    </p>

    <p v-else class="lp-empty lp-empty--loading">
      {{ t("holdings.positions.loading", "Loading positions…") }}
    </p>

    <p class="lp-note">
      {{
        t(
          "holdings.positions.liveNote",
          "Market value, latest price, and unrealised P&L are live estimates from a market data provider. Official accounting NAV is the closing valuation and remains the authoritative figure.",
        )
      }}
    </p>
  </div>
</template>

<style scoped>
.lp-wrap {
  display: grid;
  gap: 8px;
}

.lp-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  font-variant-numeric: tabular-nums;
}

.lp-th-group-row {
  border-bottom: 0;
}

.lp-th {
  padding: 8px 10px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-tertiary);
  border-bottom: 1px solid var(--border-subtle);
  background: transparent;
}

.lp-th--group {
  padding-bottom: 4px;
  font-size: 9px;
  color: var(--text-placeholder);
  border-bottom: 0;
}

.lp-th-group-last {
  text-align: left;
  padding-left: 20px;
}

.lp-th--left {
  text-align: left;
}
.lp-th--right {
  text-align: right;
}
.lp-th--center {
  text-align: center;
}

.lp-row {
  border-bottom: 1px solid var(--border-subtle);
}
.lp-row:hover {
  background: var(--bg-row-hover);
}

.lp-row--subtotal {
  background: transparent;
  font-weight: 700;
  border-top: 2px solid var(--border-subtle);
  border-bottom: 2px solid var(--border-subtle);
}

.lp-row--subtotal .lp-td {
  color: var(--text-primary);
  padding: 12px 10px;
}

.lp-row--cash .lp-td--mono {
  color: var(--text-secondary);
}

.lp-td {
  padding: 10px 10px;
  color: var(--text-primary);
  vertical-align: middle;
}

.lp-td--mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.02em;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 6px;
}

.lp-ticker-badge {
  display: inline-flex;
  align-items: center;
  height: 16px;
  padding: 0 4px;
  background: var(--bg-sidebar-subtle);
  border: 1px solid var(--border-subtle);
  border-radius: 4px;
  font-size: 9px;
  font-weight: 700;
  color: var(--text-tertiary);
  letter-spacing: 0.02em;
}

.lp-td--num {
  text-align: right;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}
.lp-td--center {
  text-align: center;
}

.lp-pnl-val {
  font-weight: 600;
}

.lp-td__sub {
  display: block;
  font-size: 10px;
  font-weight: 500;
  color: inherit;
  margin-top: 1px;
}

.lp-chip {
  display: inline-flex;
  align-items: center;
  height: 20px;
  padding: 0 8px;
  background: var(--status-executed-bg);
  border: 1px solid var(--border-focus);
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  color: var(--status-executed-text);
  white-space: nowrap;
}
.lp-chip--cash {
  background: var(--status-in-review-bg);
  border-color: var(--border-subtle);
  color: var(--status-in-review-text);
}

.lp-feed {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.02em;
  text-transform: capitalize;
}

.lp-feed-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}

.lp-feed--live {
  color: var(--state-success);
}

.lp-feed--live .lp-feed-dot {
  background: var(--state-success);
}

.lp-feed--stale {
  color: var(--state-warning);
}

.lp-feed--stale .lp-feed-dot {
  background: var(--state-warning);
}

.lp-feed-dot--settled {
  background: var(--state-success);
}

.lp-empty {
  margin: 4px 0 0;
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
  border: 1px dashed var(--border-subtle);
  border-radius: 6px;
}
.lp-empty--loading {
  color: var(--text-secondary);
}

.lp-note {
  margin: 0;
  font-size: 11px;
  color: var(--text-tertiary);
  line-height: 1.5;
}

.tone-positive {
  color: var(--state-success);
}
.tone-negative {
  color: var(--state-danger);
}
.tone-neutral {
  color: var(--text-secondary);
}
</style>
