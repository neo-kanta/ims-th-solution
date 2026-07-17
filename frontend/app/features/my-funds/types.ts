/**
 * Fund DTOs and derived view models shared by portfolio-workspace and the
 * investment operator page. View models centralise the mapping from backend
 * DTOs (FundResponse, ValuationResponse, WorkflowStateResponse, Breach) into
 * a flat shape the UI consumes. All API access goes through the generated
 * OpenAPI client — see services/myFundsApi.ts.
 */
import type { components } from "~/api/ims-api";

export type ApiFund = components["schemas"]["FundResponse"];
export type ApiPortfolio = components["schemas"]["PortfolioResponse"];
export type ApiValuation = components["schemas"]["ValuationResponse"];
export type ApiCashBalance = components["schemas"]["CashBalanceResponse"];
export type ApiBreach = components["schemas"]["Breach"];
export type ApiWorkflowState = components["schemas"]["WorkflowStateResponse"];

/**
 * Derived role of the current user against one fund.
 *
 * `MANAGER` — fund.manager_user_id matches me.
 * `MEMBER`  — user has data access via permissions.contracts[] but is not the
 *             manager. Covers reviewer/analyst/trader/compliance until per-fund
 *             role assignments land in the backend.
 */
export type MyFundRole = "MANAGER" | "MEMBER";

/**
 * Operational status derived from fund.status + workflow.currentState.
 *
 * `ACTIVE`  — fund.status=ACTIVE and trading allowed (DAY_OPEN or NOT_STARTED).
 * `LOCKED`  — workflow in MANAGER_APPROVED / TRANSACTION_CLOSED / ACCOUNTING_CLOSED.
 * `CLOSED`  — fund.status != ACTIVE.
 * `BREACH`  — any open compliance breach. Wins over ACTIVE/LOCKED for badge color.
 */
export type MyFundStatus = "ACTIVE" | "LOCKED" | "CLOSED" | "BREACH";

/**
 * Aggregated NAV / AUM for a fund computed from its portfolios' latest
 * valuations. Values are decimal strings to preserve precision; numeric
 * twins (`*_numeric`) are derived once for sorting / bar math.
 *
 * When the fund has no portfolios or all portfolio valuations are missing,
 * `available` is false and the UI shows a "no valuation" placeholder.
 */
export interface FundValuationSummary {
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
  /** Number of portfolios actually contributing to the aggregate. */
  portfolio_count: number;
}

/**
 * Workflow snapshot for the fund's current Thai business date.
 * `currentState` mirrors the backend enum.
 */
export interface FundWorkflowSummary {
  available: boolean;
  contract_id: string;
  business_date: string;
  current_state: string;
  allowed_actions: string[];
  opened_at: string | null;
  manager_approved_at: string | null;
  transaction_closed_at: string | null;
  accounting_closed_at: string | null;
}

export interface FundComplianceSummary {
  available: boolean;
  open_count: number;
  warning_count: number;
  worst_severity: "INFO" | "WARN" | "BLOCK" | null;
  latest_breach_id: string | null;
  latest_message: string | null;
}
