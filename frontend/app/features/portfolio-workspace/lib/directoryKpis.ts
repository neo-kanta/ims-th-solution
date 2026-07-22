import type { MyPortfolioCard, MyPortfoliosKpiStrip } from "../types";
import type {
  ValuationSummaryCoverageDTO,
  ValuationSummaryDTO,
} from "~/features/dashboard/types";

const EMPTY_COVERAGE: ValuationSummaryCoverageDTO = {
  totalFundCount: 0,
  includedFundCount: 0,
  excludedFundCount: 0,
  totalPortfolioCount: 0,
  includedPortfolioCount: 0,
  excludedPortfolioCount: 0,
  latestAvailablePortfolioCount: 0,
  oldestIncludedBusinessDate: null,
  excludedCurrencies: [],
  excludedBusinessDates: [],
  exclusionReasons: [],
  exclusions: [],
};

/**
 * Combines local non-monetary directory signals with the authoritative,
 * reporting-currency-normalized aggregate returned by the backend. Monetary
 * values are never derived from per-card portfolio snapshots.
 */
export function buildPortfolioDirectoryKpis(
  cards: readonly MyPortfolioCard[],
  valuationSummary: ValuationSummaryDTO | null,
  valuationSummaryError = false,
): MyPortfoliosKpiStrip {
  let active = 0;
  let breachCount = 0;
  let complianceUnavailableCount = 0;
  let staleCount = 0;
  let worstSeverity: MyPortfoliosKpiStrip["worst_breach_severity"] = null;

  const aumValue = Number(valuationSummary?.aumToday);
  const todayPnlValue = Number(valuationSummary?.todayPnl);
  const valuationAvailable =
    valuationSummary?.status === "AVAILABLE" &&
    valuationSummary.dataAvailable &&
    valuationSummary.aumToday !== null &&
    valuationSummary.todayPnl !== null &&
    Number.isFinite(aumValue) &&
    Number.isFinite(todayPnlValue);
  const todayPnl = valuationAvailable ? todayPnlValue : null;
  const todayPnlTrend =
    todayPnl === null || !Number.isFinite(todayPnl)
      ? null
      : todayPnl > 0
        ? "up"
        : todayPnl < 0
          ? "down"
          : "flat";

  for (const card of cards) {
    if (card.compliance.available) {
      breachCount += card.compliance.open_count;
    } else {
      complianceUnavailableCount += 1;
    }
    if (card.status === "ACTIVE" || card.status === "BREACH") active += 1;
    if (card.valuation.has_stale_inputs || !card.valuation.available) staleCount += 1;

    if (
      card.compliance.worst_severity === "BLOCK" ||
      (card.compliance.worst_severity === "WARN" && worstSeverity !== "BLOCK") ||
      (card.compliance.worst_severity === "INFO" && worstSeverity === null)
    ) {
      worstSeverity = card.compliance.worst_severity;
    }
  }

  return {
    valuation_summary_status: valuationSummaryError
      ? "ERROR"
      : valuationSummary?.status === "AVAILABLE" && !valuationAvailable
        ? "INCOMPLETE"
        : valuationSummary?.status ?? "NO_DATA",
    total_aum: valuationAvailable ? valuationSummary.aumToday : null,
    today_pnl: valuationAvailable ? valuationSummary.todayPnl : null,
    today_pnl_percent: valuationAvailable
      ? valuationSummary.todayPnlPercent
      : null,
    today_pnl_trend: todayPnlTrend,
    valuation_ccy: valuationSummary?.currency || null,
    valuation_business_date: valuationSummary?.businessDate ?? null,
    valuation_as_of: valuationSummary?.asOf ?? null,
    valuation_coverage: valuationSummary?.coverage ?? { ...EMPTY_COVERAGE },
    active_count: active,
    total_count: cards.length,
    open_breach_count: complianceUnavailableCount === 0 ? breachCount : null,
    worst_breach_severity: worstSeverity,
    compliance_unavailable_count: complianceUnavailableCount,
    stale_count: staleCount,
  };
}
