<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import { clampPct, formatPercent, pnlTone } from "../lib/holdingsFormat";
import type { ExposedAssetRow, LiquidAssetRow } from "../types";

type Variant = "liquid" | "exposed";

const props = defineProps<{
  variant: Variant;
  rows: LiquidAssetRow[] | ExposedAssetRow[];
  subtotal: LiquidAssetRow | ExposedAssetRow;
  showBar?: boolean;
}>();

const { t } = useI18n();

// The % of NAV value comes pre-computed from the API. We do NOT recompute it
// on the client — that would invite drift between row total and visible bar.
function rowPct(r: LiquidAssetRow | ExposedAssetRow): number {
  return r.pct_of_nav;
}

function isLiquid(row: LiquidAssetRow | ExposedAssetRow): row is LiquidAssetRow {
  return "instrument_key" in row;
}

function rowLabelKey(row: LiquidAssetRow | ExposedAssetRow): string {
  if (isLiquid(row)) {
    return `holdings.liquid.rows.${row.instrument_key}`;
  }
  return `holdings.exposed.rows.${row.asset_class_key}`;
}

function rowLabelFallback(row: LiquidAssetRow | ExposedAssetRow): string {
  return isLiquid(row) ? row.instrument_label : row.asset_class_label;
}

function rowLabelSecondary(row: LiquidAssetRow | ExposedAssetRow): string | undefined {
  return isLiquid(row) ? row.instrument_label_secondary : row.asset_class_label_secondary;
}
</script>

<template>
  <div class="ht-table">
    <table>
      <thead>
        <tr>
          <th class="ht-table__col-name">
            {{
              variant === "liquid"
                ? t("holdings.liquid.columns.instrument", "Instrument")
                : t("holdings.exposed.columns.assetClass", "Asset Class")
            }}
          </th>
          <th class="ht-table__col-num">{{ t("holdings.liquid.columns.assetValue", "Asset Value") }}</th>
          <th class="ht-table__col-num">{{ t("holdings.liquid.columns.pctOfNav", "% of NAV") }}</th>
          <th v-if="variant === 'exposed'" class="ht-table__col-num">
            {{ t("holdings.exposed.columns.todayPnl", "Today P&L") }}
          </th>
          <th v-if="variant === 'exposed'" class="ht-table__col-num">
            {{ t("holdings.exposed.columns.ytdPnl", "YTD P&L") }}
          </th>
          <th v-if="variant === 'liquid' && (showBar ?? true)" class="ht-table__col-bar" aria-hidden="true" />
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, idx) in rows" :key="idx">
          <td class="ht-table__col-name">
            <span class="ht-table__row-label">{{ t(rowLabelKey(row) as any, rowLabelFallback(row)) }}</span>
            <span v-if="rowLabelSecondary(row)" class="ht-table__row-secondary">
              {{ rowLabelSecondary(row) }}
            </span>
          </td>
          <td class="ht-table__col-num">{{ row.asset_value }}</td>
          <td class="ht-table__col-num">{{ formatPercent(rowPct(row)) }}</td>
          <td v-if="variant === 'exposed'" class="ht-table__col-num" :class="`tone-${pnlTone((row as ExposedAssetRow).today_pnl_numeric)}`">
            {{ (row as ExposedAssetRow).today_pnl ?? "—" }}
          </td>
          <td v-if="variant === 'exposed'" class="ht-table__col-num" :class="`tone-${pnlTone((row as ExposedAssetRow).ytd_pnl_numeric)}`">
            {{ (row as ExposedAssetRow).ytd_pnl ?? "—" }}
          </td>
          <td v-if="variant === 'liquid' && (showBar ?? true)" class="ht-table__col-bar">
            <div
              class="ht-table__bar"
              :style="{ width: `${clampPct(rowPct(row) * 2)}%` }"
              aria-hidden="true"
            />
          </td>
        </tr>
      </tbody>
      <tfoot>
        <tr class="ht-table__subtotal">
          <td class="ht-table__col-name">
            {{
              variant === "liquid"
                ? t("holdings.liquid.rows.subtotal", "Subtotal")
                : t("holdings.exposed.rows.total", "Total exposure")
            }}
          </td>
          <td class="ht-table__col-num">{{ subtotal.asset_value }}</td>
          <td class="ht-table__col-num">{{ formatPercent(subtotal.pct_of_nav) }}</td>
          <td
            v-if="variant === 'exposed'"
            class="ht-table__col-num"
            :class="`tone-${pnlTone((subtotal as ExposedAssetRow).today_pnl_numeric)}`"
          >
            {{ (subtotal as ExposedAssetRow).today_pnl ?? "—" }}
          </td>
          <td
            v-if="variant === 'exposed'"
            class="ht-table__col-num"
            :class="`tone-${pnlTone((subtotal as ExposedAssetRow).ytd_pnl_numeric)}`"
          >
            {{ (subtotal as ExposedAssetRow).ytd_pnl ?? "—" }}
          </td>
          <td v-if="variant === 'liquid' && (showBar ?? true)" class="ht-table__col-bar">
            <div
              class="ht-table__bar ht-table__bar--subtotal"
              :style="{ width: `${clampPct(subtotal.pct_of_nav * 2)}%` }"
              aria-hidden="true"
            />
          </td>
        </tr>
      </tfoot>
    </table>
  </div>
</template>

<style scoped>
.ht-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

th,
td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-subtle);
  text-align: left;
  vertical-align: middle;
}

thead th {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-secondary);
  background: var(--bg-table-header);
  border-bottom: 1px solid var(--border-subtle);
}

.ht-table__col-num {
  text-align: right;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.ht-table__col-bar {
  width: 140px;
  text-align: right;
  padding-right: 14px;
}

.ht-table__row-label {
  font-weight: 500;
  color: var(--text-primary);
}

.ht-table__row-secondary {
  display: inline-block;
  margin-left: 8px;
  font-size: 11px;
  color: var(--text-tertiary);
}

.ht-table__bar {
  display: inline-block;
  height: 6px;
  background: var(--color-primary-500, #2563eb);
  border-radius: 3px;
  min-width: 1px;
}

.ht-table__bar--subtotal {
  background: var(--color-primary-700, #1e40af);
}

.ht-table__subtotal td {
  border-top: 1px solid var(--border-subtle);
  border-bottom: none;
  font-weight: 600;
  color: var(--text-primary);
  background: var(--bg-row-hover);
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
</style>
