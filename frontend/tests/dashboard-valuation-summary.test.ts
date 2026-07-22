import { describe, expect, it } from "vitest";

import { normalizeTranslateArgs, translateMessage } from "../app/shared/i18n/core";
import { messages } from "../app/shared/i18n/messages";
import {
  buildAumMetric,
  buildPnlMetric,
  buildScopeOptions,
  defaultAumScope,
} from "../app/features/dashboard/lib/dashboard";
import { normalizeValuationSummary } from "../app/features/dashboard/lib/valuationSummaryMapping";
import type {
  ValuationSummaryCoverageDTO,
  ValuationSummaryDTO,
} from "../app/features/dashboard/types";

// Real translate function backed by the actual EN message catalog, so these
// tests also catch a typo'd or missing i18n key for the new dashboard
// scope/metric copy.
function t(
  key: string,
  paramsOrFallback?: Record<string, unknown> | string,
  fallback?: string,
): string {
  const normalized = normalizeTranslateArgs(paramsOrFallback, fallback);
  return translateMessage({
    key,
    locale: "en",
    messages,
    fallback: normalized.fallback,
    params: normalized.params,
  });
}

function makeCoverage(
  overrides: Partial<ValuationSummaryCoverageDTO> = {},
): ValuationSummaryCoverageDTO {
  return {
    totalFundCount: 1,
    includedFundCount: 1,
    excludedFundCount: 0,
    totalPortfolioCount: 1,
    includedPortfolioCount: 1,
    excludedPortfolioCount: 0,
    latestAvailablePortfolioCount: 0,
    oldestIncludedBusinessDate: "2026-07-14",
    excludedCurrencies: [],
    excludedBusinessDates: [],
    exclusionReasons: [],
    exclusions: [],
    ...overrides,
  };
}

function makeSummary(overrides: Partial<ValuationSummaryDTO> = {}): ValuationSummaryDTO {
  return {
    scope: "company",
    username: null,
    status: "AVAILABLE",
    businessDate: "2026-07-14",
    currency: "THB",
    aumToday: "1234567.89",
    todayPnl: "12345.67",
    todayPnlPercent: "1.01",
    asOf: "2026-07-14T14:30:00Z",
    dataAvailable: true,
    coverage: makeCoverage(),
    ...overrides,
  };
}

describe("normalizeValuationSummary (API integration mapping)", () => {
  it("maps the snake_case wire shape to the camelCase DTO", () => {
    const dto = normalizeValuationSummary({
      scope: "mine",
      username: "neo",
      business_date: "2026-07-14",
      currency: "THB",
      aum_today: "1234567.89",
      today_pnl: "12345.67",
      today_pnl_percent: "1.01",
      as_of: "2026-07-14T14:30:00Z",
      data_available: true,
      status: "AVAILABLE",
      coverage: {
        total_fund_count: 2,
        included_fund_count: 1,
        excluded_fund_count: 1,
        total_portfolio_count: 2,
        included_portfolio_count: 1,
        excluded_portfolio_count: 1,
        latest_available_portfolio_count: 1,
        oldest_included_business_date: "2026-07-13",
        excluded_currencies: ["USD"],
        excluded_business_dates: ["2026-07-13"],
        exclusion_reasons: ["MISSING_FX"],
        exclusions: [
          {
            fund_code: "FUND-USD",
            portfolio_code: "PF-USD",
            currency: "USD",
            business_date: "2026-07-13",
            required_business_date: "2026-07-14",
            reason: "MISSING_FX",
          },
        ],
      },
    });

    expect(dto).toEqual({
      scope: "mine",
      username: "neo",
      status: "AVAILABLE",
      businessDate: "2026-07-14",
      currency: "THB",
      aumToday: "1234567.89",
      todayPnl: "12345.67",
      todayPnlPercent: "1.01",
      asOf: "2026-07-14T14:30:00Z",
      dataAvailable: true,
      coverage: {
        totalFundCount: 2,
        includedFundCount: 1,
        excludedFundCount: 1,
        totalPortfolioCount: 2,
        includedPortfolioCount: 1,
        excludedPortfolioCount: 1,
        latestAvailablePortfolioCount: 1,
        oldestIncludedBusinessDate: "2026-07-13",
        excludedCurrencies: ["USD"],
        excludedBusinessDates: ["2026-07-13"],
        exclusionReasons: ["MISSING_FX"],
        exclusions: [
          {
            fundCode: "FUND-USD",
            portfolioCode: "PF-USD",
            currency: "USD",
            businessDate: "2026-07-13",
            requiredBusinessDate: "2026-07-14",
            reason: "MISSING_FX",
          },
        ],
      },
    });
  });

  it("defaults to an explicit unavailable state when the payload is missing, never a fabricated zero", () => {
    const dto = normalizeValuationSummary(undefined);
    expect(dto.dataAvailable).toBe(false);
    expect(dto.status).toBe("NO_DATA");
    expect(dto.scope).toBe("company");
    expect(dto.username).toBeNull();
    expect(dto.aumToday).toBeNull();
    expect(dto.todayPnl).toBeNull();
    expect(dto.coverage.totalPortfolioCount).toBe(0);
  });

  it("preserves INCOMPLETE coverage while withholding all numeric totals", () => {
    const dto = normalizeValuationSummary({
      status: "INCOMPLETE",
      data_available: false,
      currency: "THB",
      aum_today: "999999",
      today_pnl: "123",
      coverage: {
        total_portfolio_count: 3,
        included_portfolio_count: 2,
        excluded_portfolio_count: 1,
        exclusion_reasons: ["STALE_VALUATION"],
      },
    });

    expect(dto.status).toBe("INCOMPLETE");
    expect(dto.dataAvailable).toBe(false);
    expect(dto.aumToday).toBeNull();
    expect(dto.todayPnl).toBeNull();
    expect(dto.coverage).toMatchObject({
      totalPortfolioCount: 3,
      includedPortfolioCount: 2,
      excludedPortfolioCount: 1,
      exclusionReasons: ["STALE_VALUATION"],
    });
  });

});

describe("defaultAumScope", () => {
  it("defaults every authenticated dashboard session to company", () => {
    expect(defaultAumScope()).toBe("company");
  });
});

describe("buildScopeOptions", () => {
  it("offers company and mine to every authenticated dashboard user", () => {
    const options = buildScopeOptions(t);
    expect(options.map((o) => o.key)).toEqual(["company", "mine"]);
  });
});

describe("buildAumMetric", () => {
  it("renders an explicit not-available state (not a zero) when no valuation snapshot exists", () => {
    const metric = buildAumMetric(null, false, t);
    expect(metric.value).toBe("-");
    expect(metric.changeTone).toBe("neutral");
    expect(metric.helperText).toBe("Not yet available");
  });

  it("renders an explicit not-available state when dataAvailable is false", () => {
    const metric = buildAumMetric(makeSummary({ dataAvailable: false }), false, t);
    expect(metric.value).toBe("-");
  });

  it("renders a distinct load-error state on fetch failure", () => {
    const metric = buildAumMetric(null, true, t);
    expect(metric.value).toBe("-");
    expect(metric.helperText).toBe("Could not load");
  });

  it("formats the aggregate AUM with currency and an as-of timestamp when data is available", () => {
    const metric = buildAumMetric(makeSummary(), false, t);
    expect(metric.value).not.toBe("-");
    expect(metric.value).toContain("฿");
    // asOf "2026-07-14T14:30:00Z" rendered in Asia/Bangkok (UTC+7) is 21:30,
    // not the raw UTC hour — the dashboard must display local Bangkok time.
    expect(metric.helperText).toContain("21:30");
  });

  it("shows the oldest valuation date when latest available portfolio data is used", () => {
    const metric = buildAumMetric(makeSummary({
      coverage: makeCoverage({
        totalPortfolioCount: 7,
        includedPortfolioCount: 7,
        latestAvailablePortfolioCount: 1,
        oldestIncludedBusinessDate: "2026-07-17",
      }),
    }), false, t);

    expect(metric.value).not.toBe("-");
    expect(metric.helperText).toContain("Latest available data used for 1 of 7 portfolios");
    expect(metric.helperText).toContain("Jul 17, 2026");
    expect(metric.helperText).toContain("THB");
  });

  it("explains incomplete coverage instead of rendering a partial AUM", () => {
    const metric = buildAumMetric(
      makeSummary({
        status: "INCOMPLETE",
        dataAvailable: false,
        aumToday: null,
        todayPnl: null,
        coverage: makeCoverage({
          totalPortfolioCount: 4,
          includedPortfolioCount: 3,
          excludedPortfolioCount: 1,
        }),
      }),
      false,
      t,
    );

    expect(metric.value).toBe("-");
    expect(metric.helperText).toContain("3 of 4 in-scope portfolios");
  });

  it("explains when stale valuation data keeps the company total unavailable", () => {
    const metric = buildAumMetric(
      makeSummary({
        status: "INCOMPLETE",
        dataAvailable: false,
        aumToday: null,
        coverage: makeCoverage({
          totalPortfolioCount: 7,
          includedPortfolioCount: 6,
          excludedPortfolioCount: 1,
          excludedBusinessDates: ["2026-07-17"],
          exclusionReasons: ["STALE_VALUATION"],
        }),
      }),
      false,
      t,
    );

    expect(metric.value).toBe("-");
    expect(metric.helperText).toContain("Stale valuation data from Jul 17, 2026");
    expect(metric.helperText).toContain("company total unavailable");
  });
});

describe("buildPnlMetric", () => {
  it("renders success tone and a signed positive value for positive P&L", () => {
    const metric = buildPnlMetric(
      makeSummary({ todayPnl: "12345.67", todayPnlPercent: "1.01" }),
      false,
      t,
    );
    expect(metric.tone).toBe("success");
    expect(metric.changeTone).toBe("success");
    expect(metric.value.startsWith("+")).toBe(true);
    expect(metric.changeLabel).toContain("+");
  });

  it("renders danger tone and a signed negative value for negative P&L", () => {
    const metric = buildPnlMetric(
      makeSummary({ todayPnl: "-500", todayPnlPercent: "-0.5" }),
      false,
      t,
    );
    expect(metric.tone).toBe("danger");
    expect(metric.changeTone).toBe("danger");
    expect(metric.value.startsWith("+")).toBe(false);
  });

  it("does not look like a zero when data is unavailable", () => {
    const metric = buildPnlMetric(null, false, t);
    expect(metric.value).toBe("-");
    expect(metric.value).not.toBe("0");
    expect(metric.changeTone).toBe("neutral");
  });

  it("renders a distinct load-error state on fetch failure", () => {
    const metric = buildPnlMetric(null, true, t);
    expect(metric.helperText).toBe("Could not load");
  });
});
