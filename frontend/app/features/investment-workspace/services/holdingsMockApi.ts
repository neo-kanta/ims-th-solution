/**
 * Mock holdings service. Async function signatures mirror the future REST
 * client so the real implementation can be dropped in without composable or
 * component changes.
 *
 * MOCK REPLACEMENT POINT
 * ----------------------
 * When the backend lands, create services/holdingsApi.ts with the same
 * exported names but back each call with useApi().apiFetch(...). Then flip
 * the imports in:
 *   - features/investment-workspace/composables/useFundWorkspace.ts
 *   - features/investment-workspace/composables/useHoldings.ts
 * and delete this file plus mocks/fundHoldings.mock.ts.
 */

import {
  buildAllocation,
  buildAssets,
  buildFooter,
  buildNavHistory,
  buildRatios,
  buildSummary,
  mockMyFunds,
} from "../mocks/fundHoldings.mock";
import type {
  AllocationPayload,
  FundCard,
  FundFooter,
  HoldingsAssets,
  HoldingsSubTab,
  HoldingsSummary,
  NavHistoryPayload,
  NavHistoryRange,
  RatiosPayload,
} from "../types";

// Simulated latency window. Keeps loading skeletons visible without dragging
// the UX. Real API will replace this with actual network time.
const MIN_LATENCY_MS = 120;
const MAX_LATENCY_MS = 320;

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function withLatency<T>(value: T): Promise<T> {
  const span = MAX_LATENCY_MS - MIN_LATENCY_MS;
  const ms = MIN_LATENCY_MS + Math.random() * span;
  await sleep(ms);
  return value;
}

export const holdingsMockApi = {
  /** Funds the current user can open. Real endpoint: GET /investment/funds/me. */
  listMyFunds(): Promise<FundCard[]> {
    return withLatency(mockMyFunds);
  },

  /** Header + KPI tiles + freshness chip. */
  getSummary(_fundId: string, asOf: string): Promise<HoldingsSummary> {
    return withLatency(buildSummary(asOf));
  },

  /**
   * Liquid + exposed tables. The mock returns both regardless of sub-tab —
   * the page filters/selects which slice to show. The real API will likely
   * accept ?group=<sub-tab> and return only what's needed.
   */
  getAssets(
    _fundId: string,
    asOf: string,
    _group: HoldingsSubTab = "overview",
  ): Promise<HoldingsAssets> {
    return withLatency(buildAssets(asOf));
  },

  /** All four dimensions in one payload for cheap client-side switching. */
  getAllocation(_fundId: string, asOf: string): Promise<AllocationPayload> {
    return withLatency(buildAllocation(asOf));
  },

  /** Compliance-sensitive ratios with policy bounds and OK/WARNING/BLOCKER. */
  getRatios(_fundId: string, asOf: string): Promise<RatiosPayload> {
    return withLatency(buildRatios(asOf));
  },

  /** Sparkline series + high/low/latest stats for the small NAV chart. */
  getNavHistory(_fundId: string, range: NavHistoryRange): Promise<NavHistoryPayload> {
    return withLatency(buildNavHistory(range));
  },

  /** Status-strip metadata for the page footer. */
  getFooter(
    _fundId: string,
    asOf: string,
    viewLanguage: "en" | "th" | "zh",
  ): Promise<FundFooter> {
    return withLatency(buildFooter(asOf, viewLanguage));
  },

  /**
   * Mock CSV builder. Produces an actual download via Blob so the export
   * button is functional in the mock environment.
   *
   * MOCK REPLACEMENT POINT — the real endpoint is
   *   GET /investment/funds/{id}/holdings/export?asOf=...&format=csv
   * returning text/csv directly. Replace this body with a Blob fetched
   * through useApi() with responseType: "blob".
   */
  async exportCsv(fundId: string, asOf: string): Promise<{ blob: Blob; filename: string }> {
    const assets = await holdingsMockApi.getAssets(fundId, asOf);
    const lines: string[] = [];
    lines.push(`Fund,${fundId}`);
    lines.push(`As of,${asOf}`);
    lines.push("");
    lines.push("Section,Instrument,Asset Value,% of NAV");
    for (const r of assets.liquid.rows) {
      lines.push(
        csvRow(["LIQUID", r.instrument_label, String(r.asset_value_numeric), String(r.pct_of_nav)]),
      );
    }
    lines.push(
      csvRow([
        "LIQUID",
        "Subtotal",
        String(assets.liquid.subtotal.asset_value_numeric),
        String(assets.liquid.subtotal.pct_of_nav),
      ]),
    );
    lines.push("");
    lines.push("Section,Asset Class,Asset Value,% of NAV,Today P&L,YTD P&L");
    for (const r of assets.exposed.rows) {
      lines.push(
        csvRow([
          "EXPOSED",
          r.asset_class_label,
          String(r.asset_value_numeric),
          String(r.pct_of_nav),
          String(r.today_pnl_numeric ?? ""),
          String(r.ytd_pnl_numeric ?? ""),
        ]),
      );
    }
    lines.push(
      csvRow([
        "EXPOSED",
        "Total exposure",
        String(assets.exposed.subtotal.asset_value_numeric),
        String(assets.exposed.subtotal.pct_of_nav),
        String(assets.exposed.subtotal.today_pnl_numeric ?? ""),
        String(assets.exposed.subtotal.ytd_pnl_numeric ?? ""),
      ]),
    );
    const bom = "﻿";
    const blob = new Blob([bom + lines.join("\r\n")], {
      type: "text/csv;charset=utf-8",
    });
    const filename = `holdings_${fundId}_${asOf}.csv`;
    return { blob, filename };
  },
};

function csvRow(cols: string[]): string {
  return cols.map(csvCell).join(",");
}

function csvCell(value: string): string {
  if (/[",\r\n]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`;
  }
  return value;
}
