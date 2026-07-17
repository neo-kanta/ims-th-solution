import type { components } from "~/api/ims-api";

import type {
  ValuationSummaryCoverageDTO,
  ValuationSummaryDTO,
  ValuationSummaryStatus,
} from "../types";

// Kept in its own module (no runtime `~/api/openapi` import) so it can be
// unit tested in plain Vitest — see tests/dashboard-valuation-summary.test.ts.
// dashboardApi.ts re-exports this for callers that need the wire mapper
// alongside the live client.
type GeneratedValuationSummary = components["schemas"]["ValuationSummaryDTO"];
type GeneratedValuationSummaryCoverage = components["schemas"]["ValuationSummaryCoverageDTO"];

const EMPTY_COVERAGE: ValuationSummaryCoverageDTO = {
  totalFundCount: 0,
  includedFundCount: 0,
  excludedFundCount: 0,
  totalPortfolioCount: 0,
  includedPortfolioCount: 0,
  excludedPortfolioCount: 0,
  excludedCurrencies: [],
  excludedBusinessDates: [],
  exclusionReasons: [],
  exclusions: [],
};

function normalizeStatus(payload: GeneratedValuationSummary | undefined): ValuationSummaryStatus {
  if (payload?.status === "AVAILABLE" || payload?.status === "INCOMPLETE") {
    return payload.status;
  }
  return "NO_DATA";
}

function normalizeCoverage(
  coverage: GeneratedValuationSummaryCoverage | undefined,
): ValuationSummaryCoverageDTO {
  if (!coverage) return { ...EMPTY_COVERAGE };

  return {
    totalFundCount: coverage.total_fund_count ?? 0,
    includedFundCount: coverage.included_fund_count ?? 0,
    excludedFundCount: coverage.excluded_fund_count ?? 0,
    totalPortfolioCount: coverage.total_portfolio_count ?? 0,
    includedPortfolioCount: coverage.included_portfolio_count ?? 0,
    excludedPortfolioCount: coverage.excluded_portfolio_count ?? 0,
    excludedCurrencies: [...(coverage.excluded_currencies ?? [])],
    excludedBusinessDates: [...(coverage.excluded_business_dates ?? [])],
    exclusionReasons: [...(coverage.exclusion_reasons ?? [])],
    exclusions: (coverage.exclusions ?? []).map((item) => ({
      businessDate: item.business_date ?? null,
      currency: item.currency ?? null,
      fundCode: item.fund_code ?? null,
      portfolioCode: item.portfolio_code ?? null,
      reason: item.reason ?? "",
      requiredBusinessDate: item.required_business_date ?? null,
    })),
  };
}

export function normalizeValuationSummary(
  payload: GeneratedValuationSummary | undefined,
): ValuationSummaryDTO {
  const status = normalizeStatus(payload);
  const dataAvailable = status === "AVAILABLE" && payload?.data_available === true;

  return {
    scope: payload?.scope === "mine" ? "mine" : "company",
    username: payload?.username ?? null,
    status,
    businessDate: payload?.business_date ?? null,
    currency: payload?.currency ?? "",
    aumToday: dataAvailable ? payload?.aum_today ?? null : null,
    todayPnl: dataAvailable ? payload?.today_pnl ?? null : null,
    todayPnlPercent: dataAvailable ? payload?.today_pnl_percent ?? null : null,
    asOf: dataAvailable ? payload?.as_of ?? null : null,
    dataAvailable,
    coverage: normalizeCoverage(payload?.coverage),
  };
}
