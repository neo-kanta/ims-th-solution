import { describe, expect, it } from "vitest";

import { buildPortfolioDirectoryKpis } from "../app/features/portfolio-workspace/lib/directoryKpis";
import type { ValuationSummaryDTO } from "../app/features/dashboard/types";
import type { MyPortfolioCard } from "../app/features/portfolio-workspace/types";

function summary(
  overrides: Partial<ValuationSummaryDTO> = {},
): ValuationSummaryDTO {
  return {
    scope: "company",
    username: null,
    status: "AVAILABLE",
    businessDate: "2026-07-17",
    currency: "THB",
    aumToday: "777777.77",
    todayPnl: "123.45",
    todayPnlPercent: "0.02",
    asOf: "2026-07-17T08:00:00Z",
    dataAvailable: true,
    coverage: {
      totalFundCount: 2,
      includedFundCount: 2,
      excludedFundCount: 0,
      totalPortfolioCount: 2,
      includedPortfolioCount: 2,
      excludedPortfolioCount: 0,
      latestAvailablePortfolioCount: 0,
      oldestIncludedBusinessDate: "2026-07-17",
      excludedCurrencies: [],
      excludedBusinessDates: [],
      exclusionReasons: [],
      exclusions: [],
    },
    ...overrides,
  };
}

function card(
  currency: string,
  aum: number,
  overrides: Partial<MyPortfolioCard> = {},
): MyPortfolioCard {
  return {
    portfolio_id: `internal-${currency}`,
    code: `${currency}-01`,
    name: `${currency} portfolio`,
    base_currency: currency,
    valuation_currency: currency,
    portfolio_type: "DISCRETIONARY",
    risk_profile: "MODERATE",
    role: "MANAGER",
    status: "ACTIVE",
    portfolio_status_raw: "ACTIVE",
    manager_user_id: null,
    valuation: {
      available: true,
      nav: String(aum),
      nav_numeric: aum,
      aum: String(aum),
      aum_numeric: aum,
      unrealised_pnl: "10",
      unrealised_pnl_numeric: 10,
      realised_pnl: "0",
      realised_pnl_numeric: 0,
      roi: null,
      cash_balance: "0",
      cash_buffer_pct: 0,
      valuation_ccy: currency,
      business_date: "2026-07-17",
      has_stale_inputs: false,
      is_indicative: false,
    },
    compliance: {
      available: true,
      open_count: 0,
      warning_count: 0,
      worst_severity: null,
      latest_breach_id: null,
      latest_message: null,
    },
    updated_at: "2026-07-17T00:00:00Z",
    fund_id: null,
    ...overrides,
  };
}

describe("buildPortfolioDirectoryKpis", () => {
  it("uses only the server-normalized aggregate and never adds mixed-currency card values", () => {
    const result = buildPortfolioDirectoryKpis([
      card("THB", 1_000_000),
      card("USD", 2_000_000),
    ], summary());

    expect(result.total_aum).toBe("777777.77");
    expect(result.today_pnl).toBe("123.45");
    expect(result.today_pnl_trend).toBe("up");
    expect(result.valuation_ccy).toBe("THB");
    expect(result.valuation_summary_status).toBe("AVAILABLE");
    expect(result.active_count).toBe(2);
  });

  it("withholds all money for INCOMPLETE while preserving coverage", () => {
    const result = buildPortfolioDirectoryKpis(
      [card("THB", 1_000_000), card("USD", 2_000_000)],
      summary({
        status: "INCOMPLETE",
        dataAvailable: false,
        aumToday: null,
        todayPnl: null,
        todayPnlPercent: null,
        asOf: null,
        coverage: {
          ...summary().coverage,
          includedPortfolioCount: 1,
          excludedPortfolioCount: 1,
          exclusionReasons: ["MISSING_FX"],
          excludedCurrencies: ["USD"],
        },
      }),
    );

    expect(result.valuation_summary_status).toBe("INCOMPLETE");
    expect(result.total_aum).toBeNull();
    expect(result.today_pnl).toBeNull();
    expect(result.today_pnl_trend).toBeNull();
    expect(result.valuation_ccy).toBe("THB");
    expect(result.valuation_coverage).toMatchObject({
      includedPortfolioCount: 1,
      excludedPortfolioCount: 1,
      excludedCurrencies: ["USD"],
    });
  });

  it("derives the trend only from the server's today's P&L value", () => {
    const result = buildPortfolioDirectoryKpis(
      [card("THB", 1_000_000)],
      summary({ todayPnl: "-42.50", todayPnlPercent: "-0.01" }),
    );

    expect(result.today_pnl).toBe("-42.50");
    expect(result.today_pnl_trend).toBe("down");
  });

  it("distinguishes NO_DATA and request error from a real zero", () => {
    const noData = buildPortfolioDirectoryKpis([], summary({
      status: "NO_DATA",
      dataAvailable: false,
      aumToday: null,
      todayPnl: null,
      todayPnlPercent: null,
      asOf: null,
    }));
    const failed = buildPortfolioDirectoryKpis([], null, true);

    expect(noData.valuation_summary_status).toBe("NO_DATA");
    expect(noData.total_aum).toBeNull();
    expect(failed.valuation_summary_status).toBe("ERROR");
    expect(failed.total_aum).toBeNull();
  });

  it("reports compliance as unavailable instead of treating a failed lookup as clear", () => {
    const unavailable = card("THB", 100, {
      compliance: {
        available: false,
        open_count: 0,
        warning_count: 0,
        worst_severity: null,
        latest_breach_id: null,
        latest_message: null,
      },
    });

    const result = buildPortfolioDirectoryKpis([unavailable], summary());
    expect(result.open_breach_count).toBeNull();
    expect(result.compliance_unavailable_count).toBe(1);
  });
});
