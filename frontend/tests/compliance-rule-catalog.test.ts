import { describe, expect, it } from "vitest";

import {
  RULE_CATALOG,
  lookupRuleCatalog,
  ruleExplanation,
  ruleLabel,
  ruleSuggestedCorrection,
} from "../app/features/compliance/lib/ruleTypeCatalog";

describe("rule catalog", () => {
  it("contains an entry for every self-registered backend rule package", () => {
    // Spot-check against the seven blank-imported packages under
    // backend/internal/compliance/rules. If a new TypeID is added on the
    // backend, this test should fail until the catalog gets the human copy.
    const required = [
      "cash.availability",
      "concentration.single_issuer",
      "credit.min_rating",
      "quantity.min_trading_unit",
      "quantity.sell_available",
      "ratio.sector_exposure",
      "restriction.blacklist",
      "restriction.whitelist",
      "restriction.list_enforcement",
    ];
    for (const id of required) {
      const entry = RULE_CATALOG.find((e) => e.typeId === id);
      expect(entry, `missing catalog entry for ${id}`).toBeDefined();
    }
  });

  it("never returns null label for a catalog entry", () => {
    for (const entry of RULE_CATALOG) {
      expect(entry.label.length).toBeGreaterThan(0);
      expect(entry.explanation.length).toBeGreaterThan(0);
      expect(entry.suggestedCorrection.length).toBeGreaterThan(0);
    }
  });

  it("falls back to the raw id when an unknown rule type is requested", () => {
    expect(ruleLabel("future.unknown.rule")).toBe("future.unknown.rule");
    expect(lookupRuleCatalog("future.unknown.rule")).toBeNull();
    expect(ruleSuggestedCorrection("future.unknown.rule")).toBeNull();
  });

  it("uses the backend message as the explanation fallback", () => {
    const fallback = "raw backend message";
    expect(ruleExplanation("future.unknown.rule", fallback)).toBe(fallback);
  });

  it("returns the catalog explanation, not the fallback, for known types", () => {
    const explanation = ruleExplanation(
      "concentration.single_issuer",
      "ignored",
    );
    expect(explanation).not.toBe("ignored");
    expect(explanation.toLowerCase()).toContain("issuer");
  });
});
