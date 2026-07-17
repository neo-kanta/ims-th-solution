import { describe, expect, it } from "vitest";

import type { components } from "../app/api/ims-api";

import {
  normalizeComplianceBreach,
  normalizeCompliancePortfolio,
  normalizeComplianceRule,
} from "../app/features/compliance/lib/formatters";

describe("compliance generated API normalization", () => {
  it("normalizes a generated rule DTO into the feature model", () => {
    const rule = normalizeComplianceRule({
      id: "rule-1",
      ruleTypeID: "cash.availability",
      name: "Minimum cash",
      description: "",
      currentVersion: 2,
      isActive: true,
      createdBy: "user-1",
      createdAt: "2026-07-16T01:00:00Z",
      updatedAt: "2026-07-16T02:00:00Z",
      type_metadata: {
        type_id: "cash.availability",
        version: "1.0.0",
        category: "MANDATE",
        default_severity: "BLOCK",
        supported_timings: ["PRE_TRADE"],
        supported_scopes: ["PORTFOLIO"],
        overridable: false,
        description: "Cash control",
      },
    });

    expect(rule.id).toBe("rule-1");
    expect(rule.type_metadata?.default_severity).toBe("BLOCK");
  });

  it("normalizes the runtime effective-window field names", () => {
    const raw: components["schemas"]["RuleInstanceDetail"] = {
      id: "rule-2",
      ruleTypeID: "valuation.min_nav",
      name: "Minimum NAV",
      currentVersion: 1,
      isActive: true,
      createdBy: "user-1",
      createdAt: "2026-07-16T01:00:00Z",
      updatedAt: "2026-07-16T02:00:00Z",
    };
    Reflect.set(raw, "effectiveWindow", {
      valid_from: "2026-08-01T00:00:00Z",
      valid_to: "2026-12-31T00:00:00Z",
    });

    expect(normalizeComplianceRule(raw).effectiveWindow).toEqual({
      from: "2026-08-01T00:00:00Z",
      to: "2026-12-31T00:00:00Z",
    });
  });

  it("keeps portfolio-only breaches free of a fabricated contract id", () => {
    const breach = normalizeComplianceBreach({
      id: "breach-1",
      checkRecordID: "record-1",
      checkGroupID: "group-1",
      portfolioID: "portfolio-1",
      ruleTypeID: "cash.availability",
      ruleInstanceID: "rule-1",
      severity: "BLOCK",
      verdict: "BLOCK",
      status: "OPEN",
      message: "Cash below threshold",
      businessDate: "2026-07-16",
      createdAt: "2026-07-16T03:00:00Z",
    });

    expect(breach.contractID).toBeUndefined();
  });

  it("requires the business identity used by portfolio links", () => {
    expect(
      normalizeCompliancePortfolio({
        id: "portfolio-1",
        code: "A02-CORE",
        name: "Core Portfolio",
        base_currency: "THB",
        fund_id: "fund-1",
        status: "ACTIVE",
      }),
    ).toMatchObject({ code: "A02-CORE", name: "Core Portfolio" });
  });
});
