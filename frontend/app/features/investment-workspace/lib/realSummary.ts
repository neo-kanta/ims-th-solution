/**
 * Build a {@link HoldingsSummary} from real backend data — the My Funds
 * `MyFundCard` plus the fund-level NAV snapshot — so the Holdings header /
 * breadcrumb / KPI strip render the actual selected fund instead of the
 * legacy "fund-alpha" mock.
 *
 * Sections without backend coverage yet (asset table rows, allocation,
 * special ratios, NAV history) keep their mock data; this adapter only
 * fills in identity + valuation + stale-flag from the live database.
 */
import type {
  ApiPortfolio,
  MyFundCard,
} from "~/features/my-funds";
import { parseDecimalOrNull } from "~/features/my-funds";
import type { FundNavSnapshot } from "~/features/my-funds/services/myFundsApi";

import type {
  FundHeader,
  Freshness,
  HoldingsKpis,
  HoldingsSummary,
} from "../types";

const PLACEHOLDER_DASH = "—";

/**
 * Derive a "Contract XYZ" badge value from the fund's primary portfolio.
 *
 * Portfolios in the demo seed follow the pattern `<contract>-<role>`
 * (e.g. `A02-CORE`, `B14-CORE`); we drop the trailing `-CORE` so the badge
 * reads "Contract A02". For shapes outside that convention we fall back to
 * the full portfolio code so the badge is still meaningful.
 */
function deriveContractCode(portfolios: ApiPortfolio[]): string {
  const first = portfolios[0];
  const code = first?.code ?? "";
  if (!code) return "";
  const m = code.match(/^([A-Z0-9]+)[-_]CORE$/i);
  return m?.[1] ?? code;
}

function buildStaleLabel(hasStale: boolean, businessDate: string | null): string {
  if (!hasStale) return "all fresh";
  if (!businessDate) return "stale";
  const todayMs = Date.now();
  const dateMs = Date.parse(businessDate + "T00:00:00Z");
  if (Number.isNaN(dateMs)) return "stale";
  const days = Math.max(0, Math.round((todayMs - dateMs) / 86_400_000));
  if (days <= 0) return "stale";
  if (days === 1) return "stale 1d";
  return `stale ${days}d`;
}

function buildFreshness(card: MyFundCard): Freshness {
  const stale = card.valuation.has_stale_inputs;
  return {
    is_stale: stale,
    label: stale
      ? `Stale inputs · ${buildStaleLabel(true, card.valuation.business_date)}`
      : "Inputs fresh",
  };
}

/**
 * Map a workflow state to a short label for the "Last settled date" KPI.
 * Matches the language used elsewhere on the cockpit cards.
 */
function settledLabelForState(currentState: string): string {
  switch (currentState) {
    case "ACCOUNTING_CLOSED":
      return "acctg closing";
    case "TRANSACTION_CLOSED":
      return "tx closing";
    case "MANAGER_APPROVED":
      return "mgr approved";
    case "DAY_OPEN":
      return "day open";
    case "NOT_STARTED":
      return "not started";
    default:
      return "valuation";
  }
}

function fmtMoney(value: number | null | undefined, fractionDigits = 2): string {
  if (value == null || !Number.isFinite(value)) return PLACEHOLDER_DASH;
  return value.toLocaleString("en-US", {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  });
}

function buildKpis(
  card: MyFundCard,
  nav: FundNavSnapshot | null,
): HoldingsKpis {
  const navNumeric = card.valuation.aum_numeric;
  const ccy = card.valuation.valuation_ccy || card.base_currency || "THB";

  const unitNavStr = nav?.nav_per_unit;
  const unitNavNum = parseDecimalOrNull(unitNavStr ?? null);

  const unitsOutStr = nav?.total_units;
  const unitsOutNum = parseDecimalOrNull(unitsOutStr ?? null);

  const lastSettledDate =
    card.valuation.business_date ?? card.workflow.business_date ?? "";

  return {
    today_nav: {
      value: card.valuation.available ? fmtMoney(navNumeric, 2) : PLACEHOLDER_DASH,
      currency: ccy,
      // No intraday delta yet — backend currently exposes daily snapshots only.
      delta_pct_vs_last_close: 0,
    },
    today_unit_nav: {
      value: unitNavNum != null ? fmtMoney(unitNavNum, 5) : PLACEHOLDER_DASH,
      delta_vs_yesterday: 0,
    },
    today_units_outstanding: {
      value: unitsOutNum != null ? fmtMoney(unitsOutNum, 2) : PLACEHOLDER_DASH,
      movement_label: unitsOutNum != null ? "no movement" : "no units",
    },
    net_subscription: {
      value: fmtMoney(0, 2),
      currency: ccy,
      flow_label: "0 units / no flows today",
    },
    last_settled_date: {
      date: lastSettledDate,
      label: settledLabelForState(card.workflow.current_state),
    },
    stale_warning: {
      is_stale: card.valuation.has_stale_inputs,
      age_label: buildStaleLabel(
        card.valuation.has_stale_inputs,
        card.valuation.business_date,
      ),
      last_refresh_at: card.updated_at,
    },
  };
}

function buildHeader(
  card: MyFundCard,
  portfolios: ApiPortfolio[],
): FundHeader {
  return {
    code: card.code,
    short_name: card.short_name || card.code,
    contract_code: deriveContractCode(portfolios),
    privacy: "PRIVATE",
    // Watch / star counts come from the notification module which is still a
    // scaffold in the backend. Stubbed at 0 so the buttons render but show
    // no fake numbers.
    watch_count: 0,
    star_count: 0,
    subscribed: false,
    tab_counts: {},
  };
}

export interface BuildRealSummaryInput {
  card: MyFundCard;
  portfolios: ApiPortfolio[];
  nav: FundNavSnapshot | null;
  asOf: string;
}

/**
 * Convert real backend data into the shape the Holdings header / KPI
 * components already expect. Returns null when the card has no
 * valuation snapshot yet — let the caller render its "no NAV" empty state.
 */
export function buildRealSummary(
  input: BuildRealSummaryInput,
): HoldingsSummary | null {
  const { card, portfolios, nav, asOf } = input;
  if (!card) return null;

  return {
    fund: buildHeader(card, portfolios),
    business_date: card.workflow.business_date || asOf,
    as_of: asOf,
    kpis: buildKpis(card, nav),
    freshness: buildFreshness(card),
  };
}
