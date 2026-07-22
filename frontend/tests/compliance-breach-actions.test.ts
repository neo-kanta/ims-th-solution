import { describe, expect, it } from "vitest";

import {
  canOverrideBreach,
  overrideBlockedReason,
} from "../app/features/compliance/lib/breachActions";
import type { ComplianceBreach } from "../app/features/compliance/types";

function makeBreach(overrides: Partial<ComplianceBreach> = {}): ComplianceBreach {
  return {
    id: "breach-1",
    checkRecordID: "record-1",
    checkGroupID: "group-1",
    portfolioID: "portfolio-1",
    ruleTypeID: "concentration.single_issuer",
    ruleInstanceID: "rule-1",
    severity: "BLOCK",
    verdict: "BLOCK",
    status: "OPEN",
    message: "KBANK 11.2% single-issuer concentration (limit 10%).",
    businessDate: "2026-07-20",
    createdAt: "2026-07-20T10:12:00Z",
    ...overrides,
  };
}

describe("canOverrideBreach — permission + OPEN-status gating", () => {
  it("allows override when the breach is OPEN and the user holds the permission", () => {
    expect(canOverrideBreach(makeBreach({ status: "OPEN" }), true)).toBe(true);
  });

  it("blocks override when the user lacks IRG_OVERRIDE_BREACH, even on an OPEN breach", () => {
    expect(canOverrideBreach(makeBreach({ status: "OPEN" }), false)).toBe(false);
  });

  it("blocks override once the breach is OVERRIDDEN, even with permission", () => {
    expect(canOverrideBreach(makeBreach({ status: "OVERRIDDEN" }), true)).toBe(false);
  });

  it("blocks override once the breach is RESOLVED, even with permission", () => {
    expect(canOverrideBreach(makeBreach({ status: "RESOLVED" }), true)).toBe(false);
  });
});

describe("overrideBlockedReason", () => {
  it("returns null when override is available", () => {
    expect(overrideBlockedReason(makeBreach({ status: "OPEN" }), true)).toBeNull();
  });

  it("reports NOT_OPEN before NO_PERMISSION when both would apply", () => {
    expect(overrideBlockedReason(makeBreach({ status: "RESOLVED" }), false)).toBe("NOT_OPEN");
  });

  it("reports NO_PERMISSION for an OPEN breach without the permission", () => {
    expect(overrideBlockedReason(makeBreach({ status: "OPEN" }), false)).toBe("NO_PERMISSION");
  });
});
