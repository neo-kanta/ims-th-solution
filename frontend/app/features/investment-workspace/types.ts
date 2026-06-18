/**
 * Type definitions for the Single Fund Investment Workspace / Holdings page.
 *
 * The page reads every value through the generated OpenAPI client
 * (`myFundsApi`, `useFundDetail`, etc.). These types describe only the
 * local presentation shapes (header chips, KPI tiles, sub-tabs) that the
 * UI components still consume after mapping the backend payloads.
 */

export type WorkspaceTab =
  | "holdings"
  | "stages"
  | "decisions"
  | "compliance"
  | "audit"
  | "reviewers"
  | "settings";

export type HoldingsSubTab =
  | "overview"
  | "positions"
  | "equities"
  | "fixed_income"
  | "funds"
  | "futures"
  | "short_notes"
  | "repos"
  | "fx_forwards";

export type AllocationDimension = "country" | "category" | "industry" | "currency";

export type RatioStatus = "OK" | "WARNING" | "BLOCKER";

export type FundPrivacy = "PRIVATE" | "PUBLIC";

/** One row in the "My funds" list. */
export interface FundCard {
  fund_id: string;          // slug used in the URL
  fund_uuid?: string;       // present once backend wires real UUIDs
  code: string;             // e.g. "TH-GOV-LTF"
  short_name: string;       // e.g. "fund-alpha"
  contract_code: string;    // e.g. "A02"
  privacy: FundPrivacy;
  role: "MANAGER" | "DELEGATE" | "APPROVER" | "RISK_VIEWER" | "AUDITOR";
  base_currency: string;
  has_units: boolean;
  status: "ACTIVE" | "SUSPENDED" | "CLOSED";
}

/** Header block: drives the breadcrumb + badges row at the top of the page. */
export interface FundHeader {
  code: string;
  short_name: string;
  contract_code: string;
  privacy: FundPrivacy;
  watch_count: number;
  star_count: number;
  subscribed: boolean;
  tab_counts: Partial<Record<WorkspaceTab, number>>;
}

/** Six KPI tiles. Numeric fields are pre-formatted decimals from the API. */
export interface HoldingsKpis {
  today_nav: {
    value: string;
    currency: string;
    delta_pct_vs_last_close: number; // signed
  };
  today_unit_nav: {
    value: string;
    delta_vs_yesterday: number; // signed
  };
  today_units_outstanding: {
    value: string;
    movement_label: string; // e.g. "no movement"
  };
  net_subscription: {
    value: string;
    currency: string;
    flow_label: string; // e.g. "0 units / no flows today"
  };
  last_settled_date: {
    date: string;          // YYYY-MM-DD
    label: string;         // e.g. "acctg closing"
  };
  stale_warning: {
    is_stale: boolean;
    age_label: string;     // e.g. "all fresh", "stale 3d"
    last_refresh_at: string; // ISO timestamp
  };
}

/** Per-row freshness banner shown in the toolbar. */
export interface Freshness {
  is_stale: boolean;
  label: string;
}

/** Liquid asset row — Instrument · Asset Value · % of NAV (no P&L). */
export interface LiquidAssetRow {
  instrument_key: string;        // i18n key fragment, e.g. "cashOnHand"
  instrument_label: string;      // English label (fallback)
  instrument_label_secondary?: string; // optional secondary script (e.g. Chinese)
  asset_value: string;           // decimal as string
  asset_value_numeric: number;   // for bar math
  pct_of_nav: number;            // 0..100
  is_subtotal?: boolean;
}

/** Exposed asset row — Asset Class · Asset Value · % of NAV · Today/YTD P&L. */
export interface ExposedAssetRow {
  asset_class_key: string;
  asset_class_label: string;
  asset_class_label_secondary?: string;
  asset_value: string;
  asset_value_numeric: number;
  pct_of_nav: number;
  today_pnl: string | null;
  today_pnl_numeric: number | null;
  ytd_pnl: string | null;
  ytd_pnl_numeric: number | null;
  is_subtotal?: boolean;
}

/** Composite payload for the assets section. */
export interface HoldingsAssets {
  as_of: string;
  total_nav: string;
  total_nav_numeric: number;
  liquid: {
    rows: LiquidAssetRow[];
    subtotal: LiquidAssetRow;
  };
  exposed: {
    rows: ExposedAssetRow[];
    subtotal: ExposedAssetRow;
    pnl_sensitive_note: string;
  };
}

/** Summary payload combining header + KPIs + freshness. */
export interface HoldingsSummary {
  fund: FundHeader;
  business_date: string;
  as_of: string;
  kpis: HoldingsKpis;
  freshness: Freshness;
}

export interface AllocationBar {
  key: string;
  label: string;
  value_numeric: number;   // raw weight 0..100
  pct_of_nav: number;      // 0..100, may equal value_numeric
}

export interface AllocationPayload {
  as_of: string;
  dimensions: Record<AllocationDimension, AllocationBar[]>;
}

export interface SpecialRatio {
  code: string;                 // e.g. "FOREIGN_EXPOSURE"
  label_key: string;            // i18n key
  label_fallback: string;       // English label
  value: number;                // 0..100
  policy_floor?: number;        // % floor (if any)
  policy_ceiling?: number;      // % ceiling (if any)
  ceiling_label?: string;       // e.g. "≤ 50%"
  status: RatioStatus;
  /** Caption shown under the bar — already pre-formatted by the API. */
  caption?: string;
}

export interface RatiosPayload {
  as_of: string;
  irg_rule_version: string;
  ratios: SpecialRatio[];
}

export interface NavHistoryPoint {
  business_date: string; // YYYY-MM-DD
  nav_per_unit: number;
}

export interface NavHistoryPayload {
  range: NavHistoryRange;
  series: NavHistoryPoint[];
  high: number;
  low: number;
  latest: number;
  delta_pct: number; // signed pct change over the range
}

export type NavHistoryRange = "1D" | "5D" | "1M" | "3M" | "6M" | "YTD" | "1Y" | "5Y" | "All";

export interface FundFooter {
  is_active: boolean;
  last_priced_at: string;       // ISO timestamp
  irg_rule_version: string;     // e.g. "v2.14"
  audit_hash: string;           // short hex
  reference_date: string;       // YYYY-MM-DD
  view_language: "en" | "th" | "zh";
}
