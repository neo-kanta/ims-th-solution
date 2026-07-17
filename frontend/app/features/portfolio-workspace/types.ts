import type { components } from "~/api/ims-api";
import type {
  ValuationSummaryCoverageDTO,
  ValuationSummaryStatus,
} from "~/features/dashboard/types";

export type ApiPortfolio = components["schemas"]["PortfolioResponse"];
export type ApiValuation = components["schemas"]["ValuationResponse"];
export type ApiCashBalance = components["schemas"]["CashBalanceResponse"];
export type ApiBreach = components["schemas"]["Breach"];

export type MyPortfolioRole = "MANAGER" | "MEMBER";
export type MyPortfolioStatus = "ACTIVE" | "CLOSED" | "BREACH";

export interface PortfolioValuationSummary {
  available: boolean;
  nav: string;
  nav_numeric: number;
  aum: string;
  aum_numeric: number;
  unrealised_pnl: string;
  unrealised_pnl_numeric: number;
  realised_pnl: string;
  realised_pnl_numeric: number;
  roi: string | null;
  cash_balance: string;
  cash_buffer_pct: number | null;
  valuation_ccy: string;
  business_date: string | null;
  has_stale_inputs: boolean;
  is_indicative: boolean;
}

export interface PortfolioComplianceSummary {
  available: boolean;
  open_count: number;
  warning_count: number;
  worst_severity: "INFO" | "WARN" | "BLOCK" | null;
  latest_breach_id: string | null;
  latest_message: string | null;
}

export interface MyPortfolioCard {
  portfolio_id: string;
  code: string;
  name: string;
  base_currency: string;
  valuation_currency: string;
  portfolio_type: string;
  risk_profile: string;
  role: MyPortfolioRole;
  status: MyPortfolioStatus;
  portfolio_status_raw: string;
  manager_user_id: string | null;
  valuation: PortfolioValuationSummary;
  compliance: PortfolioComplianceSummary;
  updated_at: string;
  fund_id: string | null;
}

export type MyPortfoliosFilter =
  | "all"
  | "managed"
  | "breached"
  | "locked"
  | "stale";

export type MyPortfoliosSort = "name" | "breach" | "updated";

export interface MyPortfoliosKpiStrip {
  valuation_summary_status: ValuationSummaryStatus | "ERROR";
  total_aum: string | null;
  today_pnl: string | null;
  today_pnl_percent: string | null;
  today_pnl_trend: "up" | "down" | "flat" | null;
  valuation_ccy: string | null;
  valuation_business_date: string | null;
  valuation_as_of: string | null;
  valuation_coverage: ValuationSummaryCoverageDTO;
  active_count: number;
  total_count: number;
  open_breach_count: number | null;
  worst_breach_severity: "INFO" | "WARN" | "BLOCK" | null;
  compliance_unavailable_count: number;
  stale_count: number;
}
