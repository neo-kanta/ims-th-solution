/**
 * Pure derivations from backend DTOs to My Funds view models.
 *
 * No HTTP, no Vue runtime — these helpers stay framework-agnostic so they
 * can be unit-tested and reused by both the list and the detail page.
 */
import { parseDecimal, parseDecimalOrNull } from "./format";
import type {
  ApiBreach,
  ApiCashBalance,
  ApiFund,
  ApiPortfolio,
  ApiValuation,
  ApiWorkflowState,
  FundComplianceSummary,
  FundValuationSummary,
  FundWorkflowSummary,
  MyFundRole,
  MyFundStatus,
} from "../types";

/**
 * Aggregates per-portfolio valuations into a fund-level snapshot.
 *
 * NOTE: This is a display-only aggregation. The backend's
 * `POST /investment/funds/{id}/aum/compute` writes an authoritative fund-level
 * AUM snapshot — the My Funds list deliberately avoids that side effect on
 * read and instead sums latest valuations. When portfolios have mixed
 * valuation currencies the aggregate is reported under the first portfolio's
 * currency with `is_indicative=true` so the UI can warn.
 */
export function aggregateValuations(
  portfolios: ApiPortfolio[],
  latestByPortfolio: Map<string, ApiValuation | null>,
  cashByPortfolio: Map<string, ApiCashBalance[]>,
): FundValuationSummary {
  const contributing: ApiValuation[] = [];
  for (const portfolio of portfolios) {
    const id = portfolio.id;
    if (!id) continue;
    const val = latestByPortfolio.get(id);
    if (val) contributing.push(val);
  }

  if (contributing.length === 0) {
    return emptyValuationSummary();
  }

  const valuationCcy = contributing[0]!.valuation_ccy ?? "";
  const mixedCcy = contributing.some(
    (v) => (v.valuation_ccy ?? "") !== valuationCcy,
  );

  let aum = 0;
  let nav = 0;
  let unrealised = 0;
  let realised = 0;
  let cash = 0;
  let hasStale = false;
  let isIndicative = mixedCcy;
  let latestBusinessDate: string | null = null;

  for (const v of contributing) {
    aum += parseDecimal(v.aum);
    nav += parseDecimal(v.market_value);
    unrealised += parseDecimal(v.unrealised_pnl);
    realised += parseDecimal(v.realised_pnl);
    cash += parseDecimal(v.cash_balance);
    if (v.has_stale_inputs) hasStale = true;
    if (v.is_indicative) isIndicative = true;
    if (v.business_date && (!latestBusinessDate || v.business_date > latestBusinessDate)) {
      latestBusinessDate = v.business_date;
    }
  }

  // Cash buffer percentage uses the cash balances endpoint, not the valuation
  // snapshot — cash balances are the live ledger, while valuation.cash_balance
  // is the cash as of the snapshot moment.
  let liveCash = 0;
  let haveLiveCash = false;
  for (const portfolio of portfolios) {
    const id = portfolio.id;
    if (!id) continue;
    const balances = cashByPortfolio.get(id) ?? [];
    for (const b of balances) {
      const amount = parseDecimal(b.balance);
      if (Number.isFinite(amount)) {
        liveCash += amount;
        haveLiveCash = true;
      }
    }
  }

  const effectiveCash = haveLiveCash ? liveCash : cash;
  const cashBufferPct = nav > 0 ? (effectiveCash / nav) * 100 : null;

  // Pick a representative ROI: the average ROI weighted by AUM.
  let weightedRoi = 0;
  let weightDenom = 0;
  for (const v of contributing) {
    const r = parseDecimalOrNull(v.roi);
    const w = parseDecimal(v.aum);
    if (r !== null && w > 0) {
      weightedRoi += r * w;
      weightDenom += w;
    }
  }
  const roi = weightDenom > 0 ? (weightedRoi / weightDenom).toString() : null;

  return {
    available: true,
    nav: nav.toString(),
    nav_numeric: nav,
    aum: aum.toString(),
    aum_numeric: aum,
    unrealised_pnl: unrealised.toString(),
    unrealised_pnl_numeric: unrealised,
    realised_pnl: realised.toString(),
    realised_pnl_numeric: realised,
    roi,
    cash_balance: effectiveCash.toString(),
    cash_buffer_pct: cashBufferPct,
    valuation_ccy: valuationCcy,
    business_date: latestBusinessDate,
    has_stale_inputs: hasStale,
    is_indicative: isIndicative,
    portfolio_count: contributing.length,
  };
}

export function emptyValuationSummary(): FundValuationSummary {
  return {
    available: false,
    nav: "0",
    nav_numeric: 0,
    aum: "0",
    aum_numeric: 0,
    unrealised_pnl: "0",
    unrealised_pnl_numeric: 0,
    realised_pnl: "0",
    realised_pnl_numeric: 0,
    roi: null,
    cash_balance: "0",
    cash_buffer_pct: null,
    valuation_ccy: "",
    business_date: null,
    has_stale_inputs: false,
    is_indicative: false,
    portfolio_count: 0,
  };
}

export function mapWorkflowSummary(
  state: ApiWorkflowState | null | undefined,
  fundId: string,
  businessDate: string,
): FundWorkflowSummary {
  if (!state) {
    return {
      available: false,
      contract_id: fundId,
      business_date: businessDate,
      current_state: "NOT_STARTED",
      allowed_actions: [],
      opened_at: null,
      manager_approved_at: null,
      transaction_closed_at: null,
      accounting_closed_at: null,
    };
  }
  return {
    available: true,
    contract_id: state.contractId ?? fundId,
    business_date: state.businessDate ?? businessDate,
    current_state: state.currentState ?? "NOT_STARTED",
    allowed_actions: state.allowedActions ?? [],
    opened_at: state.openedAt ?? null,
    manager_approved_at: state.managerApprovedAt ?? null,
    transaction_closed_at: state.transactionClosedAt ?? null,
    accounting_closed_at: state.accountingClosedAt ?? null,
  };
}

const SEVERITY_RANK: Record<string, number> = {
  INFO: 1,
  WARN: 2,
  BLOCK: 3,
};

export function summarizeBreaches(breaches: ApiBreach[]): FundComplianceSummary {
  if (!breaches.length) {
    return {
      available: true,
      open_count: 0,
      warning_count: 0,
      worst_severity: null,
      latest_breach_id: null,
      latest_message: null,
    };
  }

  let openCount = 0;
  let warningCount = 0;
  let worstRank = 0;
  let worstSeverity: "INFO" | "WARN" | "BLOCK" | null = null;
  let latest: ApiBreach | null = null;

  for (const breach of breaches) {
    const status = (breach.status ?? "").toUpperCase();
    const severity = (breach.severity ?? "").toUpperCase();
    if (status === "OPEN") openCount += 1;
    if (severity === "WARN") warningCount += 1;
    const rank = SEVERITY_RANK[severity] ?? 0;
    if (rank > worstRank) {
      worstRank = rank;
      worstSeverity = severity as "INFO" | "WARN" | "BLOCK";
    }
    if (!latest || (breach.createdAt ?? "") > (latest.createdAt ?? "")) {
      latest = breach;
    }
  }

  return {
    available: true,
    open_count: openCount,
    warning_count: warningCount,
    worst_severity: worstSeverity,
    latest_breach_id: latest?.id ?? null,
    latest_message: latest?.message ?? null,
  };
}

export function deriveRole(fund: ApiFund, currentUserId: string | null): MyFundRole {
  if (currentUserId && fund.manager_user_id === currentUserId) {
    return "MANAGER";
  }
  return "MEMBER";
}

const LOCKING_STATES = new Set([
  "MANAGER_APPROVED",
  "TRANSACTION_CLOSED",
  "ACCOUNTING_CLOSED",
]);

export function deriveStatus(
  fund: ApiFund,
  workflow: FundWorkflowSummary,
  compliance: FundComplianceSummary,
): MyFundStatus {
  if (compliance.open_count > 0) return "BREACH";
  if ((fund.status ?? "").toUpperCase() !== "ACTIVE") return "CLOSED";
  if (LOCKING_STATES.has(workflow.current_state)) return "LOCKED";
  return "ACTIVE";
}

/**
 * Workflow stage list in the order the cockpit bar renders them. Names
 * mirror the i18n keys under `myFunds.workflow.stages`.
 */
export const WORKFLOW_STAGES = [
  "DAY_START",
  "ANALYSIS",
  "DECISION",
  "MANAGER_APPROVAL",
  "EXECUTION",
  "TRANSACTION_CLOSED",
  "ACCOUNTING_CLOSED",
] as const;

export type WorkflowStageKey = (typeof WORKFLOW_STAGES)[number];

/**
 * Maps the backend's three persisted stages plus implicit phases into the
 * seven-stage cockpit bar. Backend remains the source of truth; this just
 * paints the timeline.
 */
export function activeStageIndex(currentState: string): number {
  switch (currentState) {
    case "NOT_STARTED":
      return -1;
    case "DAY_OPEN":
      return 0;
    case "MANAGER_APPROVED":
      return 3;
    case "TRANSACTION_CLOSED":
      return 5;
    case "ACCOUNTING_CLOSED":
      return 6;
    default:
      return -1;
  }
}

/** Asia/Bangkok calendar date (YYYY-MM-DD) for the workflow business day. */
export function todayBangkokIso(now: Date = new Date()): string {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Bangkok",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(now);
  const lookup = Object.fromEntries(parts.map((p) => [p.type, p.value]));
  return `${lookup.year}-${lookup.month}-${lookup.day}`;
}
