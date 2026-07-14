import { describe, expect, it } from "vitest";

import { RULE_PARAMETER_SAMPLES } from "../app/features/compliance/lib/ruleParameterSamples";
import { RULE_CATALOG } from "../app/features/compliance/lib/ruleTypeCatalog";

// Required parameter keys per backend spi.ParameterSchema(), verified by
// reading backend/internal/compliance/rules/**. Corrects prior drift:
// cash.availability used min_balance (should be min_cash_buffer_pct),
// ratio.sector_exposure used sector_code (should be sector),
// quantity.min_trading_unit used min_units (should be default_lot_size).
const REQUIRED_PARAM_KEYS: Record<string, string[]> = {
  "allocation.asset_class_max": ["asset_class", "max_percent_nav"],
  "allocation.asset_class_min": ["asset_class", "min_percent_nav"],
  "amount.minimum_trade": ["min_amount"],
  "cash.availability": [], // min_cash_buffer_pct is optional
  "concentration.single_issuer": ["max_pct"],
  "credit.min_rating": ["min_rating"],
  "exposure.max_order_percent_aum": ["max_percent_aum"],
  "quantity.min_trading_unit": ["default_lot_size"],
  "quantity.sell_available": [], // allow_short_sell is optional
  "ratio.sector_exposure": ["sector", "max_pct"],
  "restriction.blacklist": [],
  "restriction.whitelist": [],
  "restriction.list_enforcement": [], // enforced_types is optional
  "valuation.min_nav": ["min_nav"],
};

// Keys the sample must NOT use — the specific drift found during review.
const FORBIDDEN_PARAM_KEYS: Record<string, string[]> = {
  "cash.availability": ["min_balance"],
  "ratio.sector_exposure": ["sector_code"],
  "quantity.min_trading_unit": ["min_units"],
  "restriction.list_enforcement": ["list_code"],
  "restriction.blacklist": ["list_code"],
  "restriction.whitelist": ["list_code"],
};

const STUB_TYPE_IDS = ["credit_rating.minimum", "regulatory.thai_sec"];

describe("rule builder parameter samples", () => {
  it("every selectable production rule type has a sample", () => {
    for (const entry of RULE_CATALOG) {
      if (!entry.selectable) continue;
      expect(
        RULE_PARAMETER_SAMPLES[entry.typeId],
        `missing sample for ${entry.typeId}`,
      ).toBeDefined();
    }
  });

  it("provides no creation sample for an unimplemented regulatory/stub rule", () => {
    for (const stubId of STUB_TYPE_IDS) {
      expect(RULE_PARAMETER_SAMPLES[stubId]).toBeUndefined();
    }
  });

  it("every sample parses as a JSON object", () => {
    for (const [typeId, json] of Object.entries(RULE_PARAMETER_SAMPLES)) {
      const parsed: unknown = JSON.parse(json);
      expect(parsed && typeof parsed === "object" && !Array.isArray(parsed), typeId).toBe(
        true,
      );
    }
  });

  it("each sample includes every key the backend ParameterSchema requires", () => {
    for (const [typeId, requiredKeys] of Object.entries(REQUIRED_PARAM_KEYS)) {
      const json = RULE_PARAMETER_SAMPLES[typeId];
      expect(json, `no sample for ${typeId}`).toBeDefined();
      const parsed = JSON.parse(json!) as Record<string, unknown>;
      for (const key of requiredKeys) {
        expect(Object.prototype.hasOwnProperty.call(parsed, key), `${typeId}.${key}`).toBe(
          true,
        );
      }
    }
  });

  it("never regresses to a known-wrong parameter key", () => {
    for (const [typeId, forbiddenKeys] of Object.entries(FORBIDDEN_PARAM_KEYS)) {
      const json = RULE_PARAMETER_SAMPLES[typeId];
      expect(json, `no sample for ${typeId}`).toBeDefined();
      const parsed = JSON.parse(json!) as Record<string, unknown>;
      for (const key of forbiddenKeys) {
        expect(
          Object.prototype.hasOwnProperty.call(parsed, key),
          `${typeId} sample regressed to using '${key}'`,
        ).toBe(false);
      }
    }
  });
});
