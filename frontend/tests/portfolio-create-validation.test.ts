/**
 * Create Portfolio form validation/normalization
 * (`CreatePortfolioV2Request` — see `ims-api.d.ts`). Pure logic, no Nuxt
 * imports, so it loads under bare `vitest run`.
 */
import { describe, expect, it } from "vitest";

import {
  buildPortfolioCreateRequest,
  emptyPortfolioCreateFormValues,
  normalizePortfolioCreateValues,
  validatePortfolioCreateForm,
  type PortfolioCreateFormValues,
} from "../app/features/portfolio-workspace/lib/portfolioCreateValidation";

function values(overrides: Partial<PortfolioCreateFormValues> = {}): PortfolioCreateFormValues {
  return {
    ...emptyPortfolioCreateFormValues(),
    fund_code: "TH-FUND-01",
    portfolio_type: "LIVE",
    code: "PF-001",
    name: "Growth Portfolio",
    base_currency: "thb",
    valuation_currency: "thb",
    inception_date: "2026-01-01",
    ...overrides,
  };
}

describe("normalizePortfolioCreateValues", () => {
  it("trims text fields and uppercases portfolio_type and currency codes", () => {
    const normalized = normalizePortfolioCreateValues(
      values({
        code: "  pf-001  ",
        name: "  Growth Portfolio  ",
        base_currency: " thb ",
        valuation_currency: "thb",
        portfolio_type: "live",
      }),
    );

    expect(normalized.code).toBe("pf-001");
    expect(normalized.name).toBe("Growth Portfolio");
    expect(normalized.base_currency).toBe("THB");
    expect(normalized.valuation_currency).toBe("THB");
    expect(normalized.portfolio_type).toBe("LIVE");
  });
});

describe("validatePortfolioCreateForm", () => {
  it("passes for a fully valid form", () => {
    const result = validatePortfolioCreateForm(values());
    expect(result.valid).toBe(true);
    expect(result.errors).toEqual({});
  });

  it("flags every required field as missing on a blank form", () => {
    const result = validatePortfolioCreateForm(emptyPortfolioCreateFormValues());
    expect(result.valid).toBe(false);
    expect(result.errors.fund_code).toBe("required");
    expect(result.errors.portfolio_type).toBe("required");
    expect(result.errors.code).toBe("required");
    expect(result.errors.name).toBe("required");
    expect(result.errors.base_currency).toBe("required");
    expect(result.errors.valuation_currency).toBe("required");
    expect(result.errors.inception_date).toBe("required");
  });

  it("rejects a portfolio_type outside LIVE/SIMULATION/MODEL", () => {
    const result = validatePortfolioCreateForm(values({ portfolio_type: "DEMO" }));
    expect(result.valid).toBe(false);
    expect(result.errors.portfolio_type).toBe("invalid_portfolio_type");
  });

  it("accepts every real portfolio type", () => {
    for (const type of ["LIVE", "SIMULATION", "MODEL", "live", "simulation", "model"]) {
      const result = validatePortfolioCreateForm(values({ portfolio_type: type }));
      expect(result.valid).toBe(true);
    }
  });

  it("rejects a currency code that is not 3 letters", () => {
    const result = validatePortfolioCreateForm(values({ base_currency: "TH", valuation_currency: "THBX" }));
    expect(result.valid).toBe(false);
    expect(result.errors.base_currency).toBe("invalid_currency");
    expect(result.errors.valuation_currency).toBe("invalid_currency");
  });

  it("rejects an invalid inception_date", () => {
    expect(validatePortfolioCreateForm(values({ inception_date: "not-a-date" })).errors.inception_date).toBe(
      "invalid_date",
    );
    // Overflowed calendar date (Date would otherwise silently roll this into March).
    expect(validatePortfolioCreateForm(values({ inception_date: "2026-02-30" })).errors.inception_date).toBe(
      "invalid_date",
    );
  });

  it("accepts a valid ISO calendar date", () => {
    expect(validatePortfolioCreateForm(values({ inception_date: "2026-02-28" })).valid).toBe(true);
  });
});

describe("buildPortfolioCreateRequest", () => {
  it("builds exactly the required fields when optionals are blank", () => {
    const body = buildPortfolioCreateRequest(values());
    expect(body).toEqual({
      fund_code: "TH-FUND-01",
      portfolio_type: "LIVE",
      code: "PF-001",
      name: "Growth Portfolio",
      base_currency: "THB",
      valuation_currency: "THB",
      inception_date: "2026-01-01",
    });
    expect(body).not.toHaveProperty("description");
    expect(body).not.toHaveProperty("strategy_code");
    expect(body).not.toHaveProperty("benchmark");
    expect(body).not.toHaveProperty("risk_profile");
  });

  it("includes optional fields only when non-blank", () => {
    const body = buildPortfolioCreateRequest(
      values({
        description: "  A growth-focused strategy  ",
        strategy_code: "EQUITY_GROWTH",
        benchmark: "SET50",
        risk_profile: "MODERATE",
      }),
    );
    expect(body.description).toBe("A growth-focused strategy");
    expect(body.strategy_code).toBe("EQUITY_GROWTH");
    expect(body.benchmark).toBe("SET50");
    expect(body.risk_profile).toBe("MODERATE");
  });

  it("never includes style_id or manager_user_id — no typed picker exists for either", () => {
    const body = buildPortfolioCreateRequest(values());
    expect(body).not.toHaveProperty("style_id");
    expect(body).not.toHaveProperty("manager_user_id");
  });
});
