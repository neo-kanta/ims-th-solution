import { describe, expect, it } from "vitest";

import {
  buildPortfolioComplianceHref,
  filterPortfolioOptions,
} from "../app/features/compliance/lib/portfolioFinder";
import type { CompliancePortfolioOption } from "../app/features/compliance/types";

function option(overrides: Partial<CompliancePortfolioOption> = {}): CompliancePortfolioOption {
  return {
    id: "id-1",
    code: "PF-001",
    name: "Growth Fund",
    base_currency: "THB",
    fund_id: "f1",
    ...overrides,
  };
}

describe("filterPortfolioOptions", () => {
  const options = [
    option({ id: "1", code: "PF-001", name: "Growth Fund" }),
    option({ id: "2", code: "PF-002", name: "Income Fund" }),
    option({ id: "3", code: "SIM-900", name: "Growth Simulation" }),
  ];

  it("returns nothing for an empty query — no results dropdown without input", () => {
    expect(filterPortfolioOptions(options, "")).toEqual([]);
    expect(filterPortfolioOptions(options, "   ")).toEqual([]);
  });

  it("matches by code, case-insensitively", () => {
    const result = filterPortfolioOptions(options, "pf-002");
    expect(result.map((o) => o.id)).toEqual(["2"]);
  });

  it("matches by name substring", () => {
    const result = filterPortfolioOptions(options, "growth");
    expect(result.map((o) => o.id).sort()).toEqual(["1", "3"]);
  });

  it("respects the result limit", () => {
    const many = Array.from({ length: 20 }, (_, i) => option({ id: String(i), code: `PF-${i}` }));
    expect(filterPortfolioOptions(many, "PF", 5)).toHaveLength(5);
  });
});

describe("buildPortfolioComplianceHref", () => {
  it("routes by portfolio code, URL-encoded, never by internal id", () => {
    expect(buildPortfolioComplianceHref("PF-001")).toBe("/portfolios/PF-001/compliance");
    expect(buildPortfolioComplianceHref("PF 001")).toBe("/portfolios/PF%20001/compliance");
  });
});
