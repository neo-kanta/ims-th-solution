<script setup lang="ts">
/**
 * Real positions table — driven directly by the backend's
 *   GET /investment/portfolios/{id}/holdings
 * endpoint and joined client-side with the instrument master + asset-class
 * reference so it shows tickers, names and category labels.
 *
 * Only authoritative numbers are rendered: quantity, average cost, and cost
 * basis. Mark-to-market values are intentionally omitted on a per-row basis
 * because the per-instrument latest-price endpoint isn't exposed yet — the
 * aggregate market value already lives on the KPI strip from the valuation
 * snapshot, which is the authoritative figure.
 */
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import type { ApiHolding, ApiInstrument, ApiAssetClass } from "~/features/my-funds/services/myFundsApi";

const props = defineProps<{
  holdings: ApiHolding[];
  instruments: ApiInstrument[];
  assetClasses: ApiAssetClass[];
  cashRows: { currency: string; balance: string }[];
  valuationCcy: string;
  loading?: boolean;
}>();

const { t } = useI18n();

const instrumentById = computed(() => {
  const m = new Map<string, ApiInstrument>();
  for (const i of props.instruments) {
    if (i.id) m.set(i.id, i);
  }
  return m;
});

const assetClassById = computed(() => {
  const m = new Map<string, ApiAssetClass>();
  for (const c of props.assetClasses) {
    if (c.id) m.set(c.id, c);
  }
  return m;
});

interface PositionRow {
  key: string;
  ticker: string;
  name: string;
  asset_class_label: string;
  currency: string;
  quantity: number;
  average_cost: number;
  cost_basis: number;
}

const positionRows = computed<PositionRow[]>(() => {
  const rows: PositionRow[] = [];
  for (const h of props.holdings) {
    const id = h.instrument_id ?? "";
    const inst = instrumentById.value.get(id);
    const ac = inst?.asset_class_id ? assetClassById.value.get(inst.asset_class_id) : undefined;

    let assetClassLabel = ac?.name ?? "";
    if (!assetClassLabel) {
      const kind = inst?.attributes?.asset_kind;
      if (typeof kind === "string" && kind.length > 0) {
        assetClassLabel = kind;
      }
    }
    if (!assetClassLabel) assetClassLabel = "—";

    rows.push({
      key: id || `row-${rows.length}`,
      ticker: inst?.primary_ticker ?? "—",
      name: inst?.name ?? "(unknown instrument)",
      asset_class_label: assetClassLabel,
      currency: inst?.currency ?? "",
      quantity: parseNum(h.quantity),
      average_cost: parseNum(h.average_cost),
      cost_basis: parseNum(h.cost_basis),
    });
  }
  // Sort largest cost-basis first so the user lands on the most material rows.
  rows.sort((a, b) => b.cost_basis - a.cost_basis);
  return rows;
});

const totalCostBasis = computed(() =>
  positionRows.value.reduce((sum, r) => sum + r.cost_basis, 0),
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

function parseNum(v: string | null | undefined): number {
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
</script>

<template>
  <div class="rp-wrap">
    <table v-if="positionRows.length || cashRows.length" class="rp-table">
      <thead>
        <tr>
          <th class="rp-th rp-th--left">{{ t("holdings.positions.ticker", "Ticker") }}</th>
          <th class="rp-th rp-th--left">{{ t("holdings.positions.name", "Name") }}</th>
          <th class="rp-th rp-th--left">{{ t("holdings.positions.assetClass", "Asset class") }}</th>
          <th class="rp-th rp-th--right">{{ t("holdings.positions.quantity", "Quantity") }}</th>
          <th class="rp-th rp-th--right">{{ t("holdings.positions.avgCost", "Avg cost") }}</th>
          <th class="rp-th rp-th--right">{{ t("holdings.positions.costBasis", "Cost basis") }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in positionRows" :key="row.key" class="rp-row">
          <td class="rp-td rp-td--mono">{{ row.ticker }}</td>
          <td class="rp-td">{{ row.name }}</td>
          <td class="rp-td">
            <span class="rp-chip">{{ row.asset_class_label }}</span>
          </td>
          <td class="rp-td rp-td--num">{{ fmtQty(row.quantity) }}</td>
          <td class="rp-td rp-td--num">{{ fmtMoney(row.average_cost, 2) }}</td>
          <td class="rp-td rp-td--num">{{ fmtMoney(row.cost_basis, 2) }}</td>
        </tr>

        <tr
          v-for="[ccy, amount] in [...totalCashByCurrency.entries()]"
          :key="`cash-${ccy}`"
          class="rp-row rp-row--cash"
        >
          <td class="rp-td rp-td--mono">{{ t("holdings.positions.cashLabel", "CASH") }}</td>
          <td class="rp-td">{{ t("holdings.positions.cashBalance", "Cash balance") }}</td>
          <td class="rp-td">
            <span class="rp-chip rp-chip--cash">{{ ccy }}</span>
          </td>
          <td class="rp-td rp-td--num">—</td>
          <td class="rp-td rp-td--num">—</td>
          <td class="rp-td rp-td--num">{{ fmtMoney(amount, 2) }}</td>
        </tr>

        <tr class="rp-row rp-row--subtotal">
          <td class="rp-td" colspan="5">
            {{ t("holdings.positions.subtotal", "Subtotal — cost basis (excludes cash)") }}
          </td>
          <td class="rp-td rp-td--num">{{ fmtMoney(totalCostBasis, 2) }}</td>
        </tr>
      </tbody>
    </table>

    <p v-else-if="!loading" class="rp-empty">
      {{
        t(
          "holdings.positions.empty",
          "No positions have been posted to this portfolio yet.",
        )
      }}
    </p>

    <p v-else class="rp-empty rp-empty--loading">
      {{ t("holdings.positions.loading", "Loading positions…") }}
    </p>

    <p class="rp-note">
      {{
        t(
          "holdings.positions.note",
          "Quantity, average cost and cost basis come from the live portfolio projection. Mark-to-market and per-row unrealised P&L will appear once the per-instrument price-read endpoint lands.",
        )
      }}
    </p>
  </div>
</template>

<style scoped>
.rp-wrap {
  display: grid;
  gap: 8px;
}

.rp-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  font-variant-numeric: tabular-nums;
}

.rp-th {
  padding: 8px 10px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-table-header, transparent);
}

.rp-th--left {
  text-align: left;
}
.rp-th--right {
  text-align: right;
}

.rp-row {
  border-bottom: 1px solid var(--border-subtle);
}
.rp-row:hover {
  background: var(--bg-row-hover, var(--surface-1));
}

.rp-row--subtotal {
  background: var(--surface-1);
  font-weight: 600;
}

.rp-row--cash .rp-td--mono {
  color: var(--text-tertiary);
}

.rp-td {
  padding: 8px 10px;
  color: var(--text-primary);
  vertical-align: middle;
}
.rp-td--mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  letter-spacing: 0.02em;
  white-space: nowrap;
}
.rp-td--num {
  text-align: right;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.rp-chip {
  display: inline-flex;
  align-items: center;
  height: 20px;
  padding: 0 8px;
  background: var(--action-secondary);
  border: 1px solid var(--border-default);
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
  color: var(--text-secondary);
  white-space: nowrap;
}
.rp-chip--cash {
  background: var(--status-in-review-bg);
  border-color: var(--alert-info-border);
  color: var(--status-in-review-text);
}

.rp-empty {
  margin: 4px 0 0;
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: var(--text-tertiary);
  border: 1px dashed var(--border-subtle);
  border-radius: 6px;
}

.rp-empty--loading {
  color: var(--text-secondary);
}

.rp-note {
  margin: 0;
  font-size: 11px;
  color: var(--text-tertiary);
  line-height: 1.5;
}
</style>
