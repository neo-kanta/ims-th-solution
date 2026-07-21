import { describe, expect, it } from "vitest";

import type { components } from "../app/api/ims-api";

import {
  normalizeComplianceBreach,
  normalizeComplianceCheckGroupResult,
  normalizeComplianceEvidence,
  normalizeComplianceOverride,
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

describe("normalizeComplianceEvidence", () => {
  it("normalizes metrics, references, and a threshold breach together", () => {
    const evidence = normalizeComplianceEvidence({
      metrics: { single_issuer_pct: "11.2" },
      references: { issuer: "KBANK" },
      threshold_breached: {
        actual: "11.2",
        limit: "10",
        metric_name: "single_issuer_pct",
        operator: ">",
        unit: "%",
      },
    });

    expect(evidence).toEqual({
      metrics: { single_issuer_pct: "11.2" },
      references: { issuer: "KBANK" },
      threshold_breached: {
        actual: "11.2",
        limit: "10",
        metric_name: "single_issuer_pct",
        operator: ">",
        unit: "%",
      },
    });
  });

  it("returns undefined rather than an empty object for missing/absent evidence", () => {
    expect(normalizeComplianceEvidence(undefined)).toBeUndefined();
    expect(normalizeComplianceEvidence(null)).toBeUndefined();
    expect(normalizeComplianceEvidence({})).toBeUndefined();
  });

  it("degrades a malformed payload to 'no evidence' instead of throwing", () => {
    expect(normalizeComplianceEvidence({ metrics: "not-an-object" })).toBeUndefined();
    expect(normalizeComplianceEvidence("just a string")).toBeUndefined();
  });

  it("keeps only string-valued metric entries", () => {
    const evidence = normalizeComplianceEvidence({
      metrics: { good: "1", bad: 2 },
    });
    expect(evidence?.metrics).toEqual({ good: "1" });
  });
});

describe("normalizeComplianceCheckGroupResult", () => {
  it("normalizes nested records and breaches under the check group", () => {
    const result = normalizeComplianceCheckGroupResult({
      check_group_id: "group-1",
      records: [
        {
          id: "record-1",
          checkGroupID: "group-1",
          portfolioID: "portfolio-1",
          ruleTypeID: "cash.availability",
          ruleInstanceID: "rule-1",
          verdict: "BLOCK",
          effectiveSeverity: "BLOCK",
          finalVerdict: "BLOCK",
          businessDate: "2026-07-20",
          createdAt: "2026-07-20T10:12:00Z",
        },
      ],
      breaches: [
        {
          id: "breach-1",
          checkRecordID: "record-1",
          checkGroupID: "group-1",
          portfolioID: "portfolio-1",
          ruleTypeID: "cash.availability",
          ruleInstanceID: "rule-1",
          severity: "BLOCK",
          verdict: "BLOCK",
          status: "OPEN",
          businessDate: "2026-07-20",
          createdAt: "2026-07-20T10:12:00Z",
        },
      ],
    });

    expect(result.check_group_id).toBe("group-1");
    expect(result.records).toHaveLength(1);
    expect(result.records[0]?.finalVerdict).toBe("BLOCK");
    expect(result.breaches).toHaveLength(1);
    expect(result.breaches[0]?.id).toBe("breach-1");
  });

  it("defaults missing records/breaches arrays to empty rather than throwing", () => {
    const result = normalizeComplianceCheckGroupResult({ check_group_id: "group-2" });
    expect(result.records).toEqual([]);
    expect(result.breaches).toEqual([]);
  });
});

describe("normalizeComplianceOverride", () => {
  it("normalizes the generated Override DTO without exposing internals beyond the typed shape", () => {
    const override = normalizeComplianceOverride({
      id: "override-1",
      breachID: "breach-1",
      reason: "Manager pre-approved this concentration limit breach.",
      overriddenBy: "user-1",
      createdAt: "2026-07-20T10:20:00Z",
    });

    expect(override).toEqual({
      id: "override-1",
      breachID: "breach-1",
      reason: "Manager pre-approved this concentration limit breach.",
      overriddenBy: "user-1",
      delegatedFrom: undefined,
      approvedBy: undefined,
      createdAt: "2026-07-20T10:20:00Z",
    });
  });
});
