import { describe, expect, it } from "vitest";

import { resolveMessage } from "../app/shared/i18n/core";
import { messages } from "../app/shared/i18n/messages";
import {
  RULE_CATALOG,
  lookupRuleCatalog,
  ruleExplanation,
  ruleLabel,
  ruleSuggestedCorrection,
} from "../app/features/compliance/lib/ruleTypeCatalog";

// Ground truth: every backend rule package under
// backend/internal/compliance/rules/**, and the Category each Metadata()
// implementation actually declares (verified by reading source, not guessed).
// module.go blank-imports allocation, amount, cash, concentration, credit,
// exposure, quantity, ratio, restriction, stub, valuation — all 16 type IDs
// below are live in production.
const BACKEND_RULE_TYPES: Record<string, string> = {
  "allocation.asset_class_max": "MANDATE",
  "allocation.asset_class_min": "MANDATE",
  "amount.minimum_trade": "MANDATE",
  "cash.availability": "MANDATE",
  "concentration.single_issuer": "MANDATE",
  "credit.min_rating": "MANDATE",
  "credit_rating.minimum": "MANDATE", // stub
  "exposure.max_order_percent_aum": "MANDATE",
  "quantity.min_trading_unit": "MANDATE",
  "quantity.sell_available": "MANDATE",
  "ratio.sector_exposure": "MANDATE",
  "regulatory.thai_sec": "REGULATORY", // stub
  "restriction.blacklist": "RESTRICTION",
  "restriction.whitelist": "RESTRICTION",
  "restriction.list_enforcement": "RESTRICTION",
  "valuation.min_nav": "MANDATE",
};

const STUB_TYPE_IDS = ["credit_rating.minimum", "regulatory.thai_sec"];

// A fake `t` that resolves purely against the real message catalogs — same
// mechanism the app uses, without needing Nuxt's useI18n() runtime context.
function tFor(locale: "en" | "th" | "zh") {
  return (key: string): string => resolveMessage(messages[locale], key) ?? key;
}

describe("rule catalog coverage", () => {
  it("contains an entry for every registered backend rule type", () => {
    for (const id of Object.keys(BACKEND_RULE_TYPES)) {
      const entry = RULE_CATALOG.find((e) => e.typeId === id);
      expect(entry, `missing catalog entry for ${id}`).toBeDefined();
    }
  });

  it("has no catalog entry for a rule type the backend does not register", () => {
    for (const entry of RULE_CATALOG) {
      expect(
        BACKEND_RULE_TYPES[entry.typeId],
        `catalog entry ${entry.typeId} has no matching backend rule package`,
      ).toBeDefined();
    }
  });

  it("matches verified backend Metadata().Category for every entry", () => {
    for (const entry of RULE_CATALOG) {
      expect(entry.category, entry.typeId).toBe(BACKEND_RULE_TYPES[entry.typeId]);
    }
  });

  // Regression guard for the drift found during review: credit.min_rating and
  // credit_rating.minimum are backend MANDATE (not RESTRICTION), and
  // ratio.sector_exposure is backend MANDATE (not RATIO).
  it("fixes the specific category drift found during review", () => {
    expect(lookupRuleCatalog("credit.min_rating")?.category).toBe("MANDATE");
    expect(lookupRuleCatalog("credit_rating.minimum")?.category).toBe("MANDATE");
    expect(lookupRuleCatalog("ratio.sector_exposure")?.category).toBe("MANDATE");
  });

  it("marks only the two backend stubs as non-selectable", () => {
    for (const entry of RULE_CATALOG) {
      const expectSelectable = !STUB_TYPE_IDS.includes(entry.typeId);
      expect(entry.selectable, entry.typeId).toBe(expectSelectable);
    }
  });

  it("falls back to the raw id when an unknown rule type is requested", () => {
    const t = tFor("en");
    expect(ruleLabel("future.unknown.rule", t)).toBe("future.unknown.rule");
    expect(lookupRuleCatalog("future.unknown.rule")).toBeNull();
    expect(ruleSuggestedCorrection("future.unknown.rule", t)).toBeNull();
  });

  it("uses the backend message as the explanation fallback for unknown types", () => {
    const t = tFor("en");
    const fallback = "raw backend message";
    expect(ruleExplanation("future.unknown.rule", fallback, t)).toBe(fallback);
  });
});

describe("rule catalog i18n resolution", () => {
  const locales = ["en", "th", "zh"] as const;

  it.each(locales)("resolves label/explanation/suggestedCorrection in %s", (locale) => {
    const t = tFor(locale);
    for (const entry of RULE_CATALOG) {
      const label = ruleLabel(entry.typeId, t);
      const explanation = ruleExplanation(entry.typeId, "unused fallback", t);
      const correction = ruleSuggestedCorrection(entry.typeId, t);

      // A resolved string must not equal the raw translation key path (that
      // would mean the locale is missing the entry and translateMessage fell
      // through to returning the key itself).
      expect(label, `${locale} ${entry.typeId} label`).not.toBe(entry.labelKey);
      expect(explanation, `${locale} ${entry.typeId} explanation`).not.toBe(
        entry.explanationKey,
      );
      expect(correction, `${locale} ${entry.typeId} suggestedCorrection`).not.toBe(
        entry.suggestedCorrectionKey,
      );
      expect(label.length).toBeGreaterThan(0);
      expect(explanation.length).toBeGreaterThan(0);
      expect(correction && correction.length).toBeGreaterThan(0);
    }
  });

  it("keeps stub labels disambiguated as stubs in every locale", () => {
    for (const locale of locales) {
      const t = tFor(locale);
      // Each locale marks the stub with its own "(stub)"-equivalent suffix;
      // assert non-empty and distinct from the non-stub credit.min_rating
      // label rather than asserting one literal substring per locale.
      const stubLabel = ruleLabel("credit_rating.minimum", t);
      const realLabel = ruleLabel("credit.min_rating", t);
      expect(stubLabel).not.toBe(realLabel);
    }
  });
});
