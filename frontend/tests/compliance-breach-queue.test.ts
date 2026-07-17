import { describe, expect, it } from "vitest";

import {
  buildPortfolioComplianceHref,
  resolvePortfolioLink,
  sortBreachesForQueue,
  type PortfolioLinkOption,
} from "../app/features/compliance/lib/breachQueue";
import type { ComplianceBreach } from "../app/features/compliance/types";

function makeBreach(overrides: Partial<ComplianceBreach> = {}): ComplianceBreach {
  return {
    id: "b1",
    checkRecordID: "cr1",
    checkGroupID: "cg1",
    portfolioID: "p1",
    contractID: "c1",
    ruleTypeID: "cash.availability",
    ruleInstanceID: "ri1",
    severity: "WARN",
    verdict: "WARN",
    status: "OPEN",
    message: "test",
    businessDate: "2026-07-15",
    createdAt: "2026-07-15T10:00:00Z",
    ...overrides,
  };
}

describe("sortBreachesForQueue", () => {
  it("orders BLOCK before WARN/REQUIRE_APPROVAL before MONITOR", () => {
    const warn = makeBreach({ id: "warn", severity: "WARN", createdAt: "2026-07-15T10:00:00Z" });
    const block = makeBreach({ id: "block", severity: "BLOCK", createdAt: "2026-07-15T09:00:00Z" });
    const monitor = makeBreach({ id: "monitor", severity: "MONITOR", createdAt: "2026-07-15T11:00:00Z" });

    const sorted = sortBreachesForQueue([warn, block, monitor]);
    expect(sorted.map((b) => b.id)).toEqual(["block", "warn", "monitor"]);
  });

  it("breaks ties within the same severity by most-recent first", () => {
    const older = makeBreach({ id: "older", severity: "BLOCK", createdAt: "2026-07-14T00:00:00Z" });
    const newer = makeBreach({ id: "newer", severity: "BLOCK", createdAt: "2026-07-15T00:00:00Z" });

    const sorted = sortBreachesForQueue([older, newer]);
    expect(sorted.map((b) => b.id)).toEqual(["newer", "older"]);
  });

  it("does not mutate the input array", () => {
    const items = [makeBreach({ id: "a" }), makeBreach({ id: "b" })];
    const copy = [...items];
    sortBreachesForQueue(items);
    expect(items).toEqual(copy);
  });
});

describe("resolvePortfolioLink", () => {
  const byId = new Map<string, PortfolioLinkOption>([
    ["p1", { id: "p1", code: "PF-001", name: "Growth Fund" }],
  ]);

  it("resolves a known portfolio id to a code + combined label", () => {
    const result = resolvePortfolioLink("p1", byId, "Portfolio unavailable");
    expect(result).toEqual({ label: "PF-001 — Growth Fund", code: "PF-001" });
  });

  it("never falls back to the raw UUID — uses the neutral label when unresolved", () => {
    const unknownId = "99999999-9999-9999-9999-999999999999";
    const result = resolvePortfolioLink(unknownId, byId, "Portfolio unavailable");
    expect(result.label).toBe("Portfolio unavailable");
    expect(result.label).not.toContain(unknownId);
    expect(result.code).toBeNull();
  });

  it("falls back to the neutral label for a null/undefined id", () => {
    expect(resolvePortfolioLink(null, byId, "Portfolio unavailable").code).toBeNull();
    expect(resolvePortfolioLink(undefined, byId, "Portfolio unavailable").code).toBeNull();
  });
});

describe("buildPortfolioComplianceHref", () => {
  it("builds a portfolio-code route, URL-encoded", () => {
    expect(buildPortfolioComplianceHref("PF 001")).toBe("/portfolios/PF%20001/compliance");
  });
});
