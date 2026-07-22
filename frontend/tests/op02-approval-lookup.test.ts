import { describe, expect, it } from "vitest";

import {
  approvalFiltersFromQuery,
  approvalFiltersToQuery,
  cleanApprovalFilters,
} from "../app/features/investment-decision/lib/approvalLookup";

describe("OP-02 approval lookup", () => {
  it("maps the business decision number to the exact decision_no filter", () => {
    expect(
      cleanApprovalFilters({
        decision_no: "  DEC-20260715-0042  ",
        search: "  ",
      }),
    ).toEqual({ decision_no: "DEC-20260715-0042" });
  });

  it("round-trips only supported approval lookup filters", () => {
    const filters = approvalFiltersFromQuery({
      decision_no: ["DEC-42", "ignored"],
      product_type: "BOND",
      business_date_from: "2026-07-15",
      status: "APPROVED",
      portfolio_id: "internal-id-must-not-enter-the-op02-query",
    });

    expect(approvalFiltersToQuery(filters)).toEqual({
      decision_no: "DEC-42",
      product_type: "BOND",
      business_date_from: "2026-07-15",
    });
  });

  it("round-trips the business portfolio_code filter through the URL query", () => {
    const filters = approvalFiltersFromQuery({ portfolio_code: "PF-001" });

    expect(filters).toEqual({ portfolio_code: "PF-001" });
    expect(approvalFiltersToQuery(filters)).toEqual({ portfolio_code: "PF-001" });
  });

  it("never resolves portfolio_code into the internal portfolio_id filter by itself", () => {
    const filters = cleanApprovalFilters({ portfolio_code: "PF-001" });
    expect(filters).not.toHaveProperty("portfolio_id");
  });
});
