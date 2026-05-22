/**
 * Static fixture for the Holdings workspace. ONE fund only: fund-alpha.
 *
 * Numbers below are internally consistent: every "% of NAV" was calculated
 * against `TOTAL_NAV`, and subtotals equal the sum of their child rows.
 * If you tweak a row, recompute the percentages or the asset table will
 * show numbers that disagree with the visual bars.
 *
 * Business date is anchored to a recent business day so the page works
 * regardless of when the build is shipped. The toolbar's as-of picker
 * filters this mock to a single business date — older dates produce the
 * stale-warning variant.
 */

import type {
  AllocationPayload,
  FundCard,
  FundFooter,
  HoldingsAssets,
  HoldingsSummary,
  NavHistoryPayload,
  NavHistoryRange,
  RatiosPayload,
} from "../types";

export const MOCK_FUND_SLUG = "fund-alpha";

/** Today's business date in the mock world. */
export const MOCK_AS_OF = "2026-05-21";

/** When the data is filtered to a date older than this, we go stale. */
export const STALE_THRESHOLD_DAYS = 3;

/** Anchor NAV used to drive % calculations across the mock. */
const TOTAL_NAV: number = 2_638_865_022.97;
const CCY = "THB";

// ─── My funds list ───────────────────────────────────────────────────────────

export const mockMyFunds: FundCard[] = [
  {
    fund_id: MOCK_FUND_SLUG,
    fund_uuid: "00000000-0000-0000-0000-00000000a02f",
    code: "TH-GOV-LTF",
    short_name: "fund-alpha",
    contract_code: "A02",
    privacy: "PRIVATE",
    role: "MANAGER",
    base_currency: CCY,
    has_units: true,
    status: "ACTIVE",
  },
];

// ─── Summary (header + KPIs + freshness) ─────────────────────────────────────

export function buildSummary(asOf: string): HoldingsSummary {
  const isStale = isAsOfStale(asOf);
  return {
    fund: {
      code: "TH-GOV-LTF",
      short_name: "fund-alpha",
      contract_code: "A02",
      privacy: "PRIVATE",
      watch_count: 14,
      star_count: 38,
      subscribed: false,
      tab_counts: {
        holdings: undefined,
        stages: 7,
        decisions: 23,
        compliance: 2,
      },
    },
    business_date: asOf,
    as_of: asOf,
    kpis: {
      today_nav: {
        value: "2,638,865,022.97",
        currency: CCY,
        delta_pct_vs_last_close: -0.18,
      },
      today_unit_nav: {
        value: "10.12310",
        delta_vs_yesterday: -0.002,
      },
      today_units_outstanding: {
        value: "260,677,580.00",
        movement_label: "no movement",
      },
      net_subscription: {
        value: "0.00",
        currency: CCY,
        flow_label: "0 units / no flows today",
      },
      last_settled_date: {
        date: "2026-05-19",
        label: "acctg closing",
      },
      stale_warning: {
        is_stale: isStale,
        age_label: isStale ? "stale" : "all fresh",
        last_refresh_at: `${asOf}T11:42:00+07:00`,
      },
    },
    freshness: {
      is_stale: isStale,
      label: isStale ? "Stale data" : "Feeds OK",
    },
  };
}

// ─── Assets (liquid + exposed) ───────────────────────────────────────────────

export function buildAssets(asOf: string): HoldingsAssets {
  const liquidRows = [
    row("cashOnHand", "Cash on hand", "", 0),
    row("timeDeposits", "Time deposits", "", 600_000_000),
    row("shortTermNotes", "Short-term notes", "", 0),
    row("repo", "Repo (RP)", "", 0),
    row("receivables", "Receivables", "", 56_700_000),
    row("payables", "Payables", "", 0),
  ];
  const liquidSum = liquidRows.reduce((a, r) => a + r.asset_value_numeric, 0);

  const exposedRows = [
    exposedRow("listedEquity", "Listed equity", "", 98_551_080.13, -91_438.0, -91_438.0),
    exposedRow("otcEquity", "OTC equity", "", 1_725_080.08, null, null),
    exposedRow("etf", "ETF", "", 388_889_628.78, null, null),
    exposedRow("bonds", "Bonds", "", 506_333_551.78, -1_211_488.0, -1_211_488.0),
    exposedRow("mutualFunds", "Mutual funds", "", 6_700_757.38, null, null),
    exposedRow("derivatives", "Derivatives", "", 6_850_000.00, -35_411_807.74, -35_411_807.74),
  ];
  const exposedSum = exposedRows.reduce((a, r) => a + r.asset_value_numeric, 0);
  const exposedPnlT = exposedRows.reduce(
    (a, r) => a + (r.today_pnl_numeric ?? 0),
    0,
  );
  const exposedPnlYtd = exposedRows.reduce(
    (a, r) => a + (r.ytd_pnl_numeric ?? 0),
    0,
  );

  return {
    as_of: asOf,
    total_nav: formatNumber(TOTAL_NAV),
    total_nav_numeric: TOTAL_NAV,
    liquid: {
      rows: liquidRows,
      subtotal: {
        instrument_key: "subtotal",
        instrument_label: "Subtotal",
        asset_value: formatNumber(liquidSum),
        asset_value_numeric: liquidSum,
        pct_of_nav: pct(liquidSum),
        is_subtotal: true,
      },
    },
    exposed: {
      rows: exposedRows,
      subtotal: {
        asset_class_key: "subtotal",
        asset_class_label: "Total exposure",
        asset_value: formatNumber(exposedSum),
        asset_value_numeric: exposedSum,
        pct_of_nav: pct(exposedSum),
        today_pnl: formatSigned(exposedPnlT),
        today_pnl_numeric: exposedPnlT,
        ytd_pnl: formatSigned(exposedPnlYtd),
        ytd_pnl_numeric: exposedPnlYtd,
        is_subtotal: true,
      },
      pnl_sensitive_note: "On balance sheet exposure across asset classes.",
    },
  };
}

function row(
  key: string,
  label: string,
  secondary: string,
  value: number,
) {
  return {
    instrument_key: key,
    instrument_label: label,
    instrument_label_secondary: secondary,
    asset_value: value === 0 ? "—" : formatNumber(value),
    asset_value_numeric: value,
    pct_of_nav: pct(value),
  };
}

function exposedRow(
  key: string,
  label: string,
  secondary: string,
  value: number,
  todayPnl: number | null,
  ytdPnl: number | null,
) {
  return {
    asset_class_key: key,
    asset_class_label: label,
    asset_class_label_secondary: secondary,
    asset_value: value === 0 ? "—" : formatNumber(value),
    asset_value_numeric: value,
    pct_of_nav: pct(value),
    today_pnl: todayPnl === null ? null : formatSigned(todayPnl),
    today_pnl_numeric: todayPnl,
    ytd_pnl: ytdPnl === null ? null : formatSigned(ytdPnl),
    ytd_pnl_numeric: ytdPnl,
  };
}

// ─── Allocation (all four dimensions in one payload) ─────────────────────────

export function buildAllocation(asOf: string): AllocationPayload {
  return {
    as_of: asOf,
    dimensions: {
      country: [
        { key: "CN", label: "China (Mainland)", value_numeric: 19.0, pct_of_nav: 19.0 },
        { key: "OTHERS", label: "Others", value_numeric: 12.6, pct_of_nav: 12.6 },
        { key: "US", label: "United States", value_numeric: 4.728, pct_of_nav: 4.728 },
        { key: "TW", label: "Taiwan, ROC", value_numeric: 0.385, pct_of_nav: 0.385 },
        { key: "TOTAL", label: "Total", value_numeric: 36.737, pct_of_nav: 36.737 },
      ],
      category: [
        { key: "EQ_LARGE", label: "Large-cap equity", value_numeric: 22.4, pct_of_nav: 22.4 },
        { key: "EQ_MID", label: "Mid-cap equity", value_numeric: 8.1, pct_of_nav: 8.1 },
        { key: "GOV_BOND", label: "Government bonds", value_numeric: 14.9, pct_of_nav: 14.9 },
        { key: "CORP_BOND", label: "Corporate bonds", value_numeric: 4.3, pct_of_nav: 4.3 },
        { key: "CASH", label: "Cash & equivalents", value_numeric: 22.85, pct_of_nav: 22.85 },
      ],
      industry: [
        { key: "TECH", label: "Technology", value_numeric: 14.7, pct_of_nav: 14.7 },
        { key: "FIN", label: "Financials", value_numeric: 8.9, pct_of_nav: 8.9 },
        { key: "INDU", label: "Industrials", value_numeric: 5.2, pct_of_nav: 5.2 },
        { key: "CONS", label: "Consumer", value_numeric: 4.6, pct_of_nav: 4.6 },
        { key: "OTHER", label: "Other", value_numeric: 3.3, pct_of_nav: 3.3 },
      ],
      currency: [
        { key: "THB", label: "THB", value_numeric: 63.3, pct_of_nav: 63.3 },
        { key: "USD", label: "USD", value_numeric: 22.1, pct_of_nav: 22.1 },
        { key: "CNY", label: "CNY", value_numeric: 9.1, pct_of_nav: 9.1 },
        { key: "JPY", label: "JPY", value_numeric: 3.2, pct_of_nav: 3.2 },
        { key: "OTHERS", label: "Others", value_numeric: 2.3, pct_of_nav: 2.3 },
      ],
    },
  };
}

// ─── Ratios — warning/blocker logic ──────────────────────────────────────────

export function buildRatios(asOf: string): RatiosPayload {
  return {
    as_of: asOf,
    irg_rule_version: "v2.14",
    ratios: [
      {
        code: "FOREIGN_EXPOSURE",
        label_key: "ratios.foreignExposure",
        label_fallback: "Total foreign exposure",
        value: 0.066,
        policy_ceiling: 30.0,
        ceiling_label: "≤ 30%",
        status: "OK",
      },
      {
        code: "EQUITY_BANK_TIER_DEBT",
        label_key: "ratios.equityBankTierDebt",
        label_fallback: "Equity + bank-tier debt concentration",
        value: 0.062,
        policy_ceiling: 60.0,
        ceiling_label: "≤ 60%",
        status: "OK",
      },
    ],
  };
}

// ─── NAV history ─────────────────────────────────────────────────────────────

export function buildNavHistory(range: NavHistoryRange): NavHistoryPayload {
  const days = pointsForRange(range);
  const series = generateNavSeries(days);
  const values = series.map((p) => p.nav_per_unit);
  const high = Math.max(...values);
  const low = Math.min(...values);
  const latest = values[values.length - 1] ?? 0;
  const first = values[0] ?? 0;
  const deltaPct = first === 0 ? 0 : ((latest - first) / first) * 100;
  return { range, series, high, low, latest, delta_pct: deltaPct };
}

function pointsForRange(range: NavHistoryRange): number {
  switch (range) {
    case "1D": return 2;
    case "5D": return 5;
    case "1M": return 22;
    case "3M": return 66;
    case "6M": return 132;
    case "YTD": return 100;
    case "1Y": return 252;
    case "5Y": return 1260;
    case "All": return 2520;
    default: return 66;
  }
}

function generateNavSeries(n: number) {
  // Deterministic noisy walk resembling a real financial asset price trend.
  const out: { business_date: string; nav_per_unit: number }[] = [];
  const end = parseDate(MOCK_AS_OF);
  for (let i = n - 1; i >= 0; i--) {
    const d = subBusinessDays(end, i);
    const t = (n - i) / n;
    
    // Multi-frequency wave components to mimic market volatility:
    // Base trend + mid-term cycle + short-term noise + micro-wobbles
    const baseTrend = t * 0.25;
    const cycle = Math.sin(t * Math.PI * 1.5) * 0.15;
    const midFreq = Math.sin(t * Math.PI * 12) * 0.03;
    const highFreq = Math.sin(t * Math.PI * 40) * 0.012;
    const microNoise = Math.sin(t * Math.PI * 120) * 0.004;
    
    const nav = 9.92 + baseTrend + cycle + midFreq + highFreq + microNoise;
    out.push({ business_date: toIso(d), nav_per_unit: Number(nav.toFixed(5)) });
  }
  return out;
}

// ─── Footer ──────────────────────────────────────────────────────────────────

export function buildFooter(asOf: string, viewLanguage: "en" | "th" | "zh"): FundFooter {
  return {
    is_active: true,
    last_priced_at: `${asOf}T14:31:00+07:00`,
    irg_rule_version: "v2.14",
    audit_hash: "a8f0c14",
    reference_date: asOf,
    view_language: viewLanguage,
  };
}

// ─── Internal helpers (mock-only) ────────────────────────────────────────────

function isAsOfStale(asOf: string): boolean {
  const ref = parseDate(MOCK_AS_OF);
  const picked = parseDate(asOf);
  const diffDays = Math.floor((ref.getTime() - picked.getTime()) / 86_400_000);
  return diffDays > STALE_THRESHOLD_DAYS;
}

function pct(value: number): number {
  if (TOTAL_NAV === 0) return 0;
  return Number(((value / TOTAL_NAV) * 100).toFixed(3));
}

function formatNumber(n: number): string {
  return n.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatSigned(n: number): string {
  if (n === 0) return "0.00";
  const sign = n < 0 ? "-" : "";
  return `${sign}${formatNumber(Math.abs(n))}`;
}

function parseDate(iso: string): Date {
  const [y, m, d] = iso.split("-").map(Number) as [number, number, number];
  return new Date(Date.UTC(y, m - 1, d));
}

function toIso(d: Date): string {
  const y = d.getUTCFullYear();
  const m = String(d.getUTCMonth() + 1).padStart(2, "0");
  const day = String(d.getUTCDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function subBusinessDays(end: Date, days: number): Date {
  let cur = new Date(end.getTime());
  let remaining = days;
  while (remaining > 0) {
    cur = new Date(cur.getTime() - 86_400_000);
    const dow = cur.getUTCDay();
    if (dow !== 0 && dow !== 6) remaining--;
  }
  return cur;
}
