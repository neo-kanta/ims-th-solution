import { describe, expect, it } from "vitest";

import {
  isCurrencyCode,
  isIsoDate,
  isStrictPositiveDecimal,
  validateDecisionDraft,
  type DecisionDraftLike,
} from "../app/features/portfolio-decision/lib/decisionValidation";
import {
  decisionDetailPath,
  decisionsListPath,
  newDecisionPath,
} from "../app/features/portfolio-decision/lib/decisionRoutes";

function draft(overrides: Partial<DecisionDraftLike> = {}): DecisionDraftLike {
  return {
    side: "BUY",
    instrumentId: "11111111-1111-1111-1111-111111111111",
    instrumentCode: "PTT",
    quantity: "1000",
    amount: "",
    limitPrice: "35.00",
    currency: "THB",
    businessDate: "2026-07-14",
    ...overrides,
  };
}

describe("isStrictPositiveDecimal", () => {
  it("accepts plain positive integers and decimals", () => {
    expect(isStrictPositiveDecimal("1000")).toBe(true);
    expect(isStrictPositiveDecimal("35.5")).toBe(true);
    expect(isStrictPositiveDecimal("0.01")).toBe(true);
  });

  it("rejects zero, negative, NaN, and scientific notation", () => {
    expect(isStrictPositiveDecimal("0")).toBe(false);
    expect(isStrictPositiveDecimal("-5")).toBe(false);
    expect(isStrictPositiveDecimal("abc")).toBe(false);
    expect(isStrictPositiveDecimal("1e5")).toBe(false);
    expect(isStrictPositiveDecimal("1E5")).toBe(false);
    expect(isStrictPositiveDecimal("")).toBe(false);
    expect(isStrictPositiveDecimal("Infinity")).toBe(false);
  });
});

describe("isIsoDate / isCurrencyCode", () => {
  it("accepts a real calendar date in YYYY-MM-DD", () => {
    expect(isIsoDate("2026-07-14")).toBe(true);
  });

  it("rejects malformed or non-existent calendar dates", () => {
    expect(isIsoDate("2026-13-01")).toBe(false);
    expect(isIsoDate("2026-02-30")).toBe(false);
    expect(isIsoDate("07/14/2026")).toBe(false);
    expect(isIsoDate("")).toBe(false);
  });

  it("requires exactly 3 uppercase letters for currency", () => {
    expect(isCurrencyCode("THB")).toBe(true);
    expect(isCurrencyCode("thb")).toBe(false);
    expect(isCurrencyCode("TH")).toBe(false);
    expect(isCurrencyCode("THBB")).toBe(false);
  });
});

describe("validateDecisionDraft", () => {
  it("accepts a fully valid BUY draft", () => {
    expect(validateDecisionDraft(draft())).toEqual({});
  });

  it("requires an instrument to be selected", () => {
    const errors = validateDecisionDraft(draft({ instrumentCode: "" }));
    expect(errors.instrument).toBeDefined();
  });

  it("requires quantity or amount, mirroring the backend's validateCreateDecision rule", () => {
    const errors = validateDecisionDraft(draft({ quantity: "", amount: "" }));
    expect(errors.quantity).toMatch(/quantity or an amount/i);
  });

  it("accepts amount-only orders without requiring quantity", () => {
    const errors = validateDecisionDraft(draft({ quantity: "", amount: "50000" }));
    expect(errors.quantity).toBeUndefined();
    expect(errors.amount).toBeUndefined();
  });

  it("does not require limit_price", () => {
    const errors = validateDecisionDraft(draft({ limitPrice: "" }));
    expect(errors.limitPrice).toBeUndefined();
  });

  it("rejects scientific-notation quantity even though Number() would accept it", () => {
    const errors = validateDecisionDraft(draft({ quantity: "1e5" }));
    expect(errors.quantity).toMatch(/positive number/i);
  });

  it("rejects a non-3-letter currency", () => {
    const errors = validateDecisionDraft(draft({ currency: "thb" }));
    expect(errors.currency).toBeDefined();
  });

  it("rejects a malformed business date", () => {
    const errors = validateDecisionDraft(draft({ businessDate: "14-07-2026" }));
    expect(errors.businessDate).toBeDefined();
  });

  it("flags a SELL quantity that exceeds the known available holding", () => {
    const errors = validateDecisionDraft(draft({ side: "SELL", quantity: "500" }), {
      availableQuantity: 100,
    });
    expect(errors.quantity).toMatch(/exceeds available holding/i);
  });

  it("allows a SELL quantity within the available holding", () => {
    const errors = validateDecisionDraft(draft({ side: "SELL", quantity: "100" }), {
      availableQuantity: 100,
    });
    expect(errors.quantity).toBeUndefined();
  });

  it("does not apply the oversell check to BUY orders", () => {
    const errors = validateDecisionDraft(draft({ side: "BUY", quantity: "999999" }), {
      availableQuantity: 10,
    });
    expect(errors.quantity).toBeUndefined();
  });
});

describe("validateDecisionDraft SELL non-owned-instrument prevention", () => {
  it("blocks a SELL of an instrument the portfolio does not hold, even before a quantity is entered", () => {
    const errors = validateDecisionDraft(draft({ side: "SELL", quantity: "" }), {
      isOwnedInstrument: false,
    });
    expect(errors.instrument).toMatch(/does not hold this instrument/i);
  });

  it("blocks a SELL of a non-owned instrument independently of the oversell/quantity check", () => {
    // A tiny quantity that would otherwise pass the oversell check must
    // still be rejected because the instrument itself isn't owned.
    const errors = validateDecisionDraft(draft({ side: "SELL", quantity: "1" }), {
      availableQuantity: 100,
      isOwnedInstrument: false,
    });
    expect(errors.instrument).toBeDefined();
  });

  it("allows a SELL once the instrument is confirmed owned", () => {
    const errors = validateDecisionDraft(draft({ side: "SELL", quantity: "50" }), {
      availableQuantity: 100,
      isOwnedInstrument: true,
    });
    expect(errors.instrument).toBeUndefined();
  });

  it("does not block on ownership when it is unknown (holdings not loaded yet) — only an explicit false blocks", () => {
    const errors = validateDecisionDraft(draft({ side: "SELL", quantity: "50" }), {
      availableQuantity: 100,
      isOwnedInstrument: null,
    });
    expect(errors.instrument).toBeUndefined();
  });

  it("does not apply the ownership gate to BUY orders", () => {
    const errors = validateDecisionDraft(draft({ side: "BUY", quantity: "50" }), {
      isOwnedInstrument: false,
    });
    expect(errors.instrument).toBeUndefined();
  });
});

describe("portfolio decision route builders", () => {
  it("URL-encodes portfolio codes containing special characters", () => {
    expect(decisionsListPath("TH/EQ 01")).toBe("/portfolios/TH%2FEQ%2001/decisions");
    expect(newDecisionPath("TH EQ#01")).toBe("/portfolios/TH%20EQ%2301/decisions/new");
  });

  it("URL-encodes both portfolioCode and decisionId in the detail path", () => {
    expect(decisionDetailPath("TH-EQ 01", "dec/123")).toBe(
      "/portfolios/TH-EQ%2001/decisions/dec%2F123",
    );
  });

  it("round-trips a normal portfolio code unchanged", () => {
    expect(decisionsListPath("TH-EQ-01")).toBe("/portfolios/TH-EQ-01/decisions");
  });
});
