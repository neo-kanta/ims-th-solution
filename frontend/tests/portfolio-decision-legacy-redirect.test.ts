/**
 * Tests for the legacy fund-scoped "new decision" entry-point resolver,
 * shared by frontend/app/pages/investment/funds/[fundId]/operation/new.vue
 * and frontend/app/pages/investment/operator/[fundId]/operation/new.vue —
 * both must resolve accessible portfolios and forward the user instead of
 * rendering the old V1 fund_id/portfolio_id form, and must never silently
 * pick an arbitrary portfolio when the choice is ambiguous.
 */
import { describe, expect, it } from "vitest";

import { resolveLegacyFundRedirect } from "../app/features/portfolio-decision/lib/legacyEntry";

describe("resolveLegacyFundRedirect", () => {
  it("redirects automatically when exactly one accessible portfolio exists", () => {
    const result = resolveLegacyFundRedirect([{ code: "TH-EQ-01", name: "Thailand Equity" }]);

    expect(result).toEqual({
      kind: "redirect",
      path: "/portfolios/TH-EQ-01/decisions/new",
      portfolio: { code: "TH-EQ-01", name: "Thailand Equity" },
    });
  });

  it("never silently picks a portfolio when several are accessible — returns a chooser instead", () => {
    const result = resolveLegacyFundRedirect([
      { code: "TH-EQ-01", name: "Thailand Equity" },
      { code: "TH-EQ-02", name: "Thailand Equity Growth" },
    ]);

    expect(result.kind).toBe("choose");
    if (result.kind === "choose") {
      expect(result.portfolios).toEqual([
        { code: "TH-EQ-01", name: "Thailand Equity" },
        { code: "TH-EQ-02", name: "Thailand Equity Growth" },
      ]);
    }
  });

  it("returns an empty result when the fund has no portfolios", () => {
    expect(resolveLegacyFundRedirect([])).toEqual({ kind: "empty" });
  });

  it("ignores portfolios missing a code (cannot build a portfolio-code route for them)", () => {
    const result = resolveLegacyFundRedirect([
      { code: null, name: "No code" },
      { code: "TH-EQ-01", name: "Thailand Equity" },
    ]);

    expect(result).toEqual({
      kind: "redirect",
      path: "/portfolios/TH-EQ-01/decisions/new",
      portfolio: { code: "TH-EQ-01", name: "Thailand Equity" },
    });
  });

  it("falls back to the code as the display name when name is blank", () => {
    const result = resolveLegacyFundRedirect([{ code: "TH-EQ-01", name: "" }]);

    expect(result).toEqual({
      kind: "redirect",
      path: "/portfolios/TH-EQ-01/decisions/new",
      portfolio: { code: "TH-EQ-01", name: "TH-EQ-01" },
    });
  });

  it("URL-encodes the redirect path for a portfolio code with special characters", () => {
    const result = resolveLegacyFundRedirect([{ code: "TH/EQ 01", name: "Special" }]);

    expect(result.kind).toBe("redirect");
    if (result.kind === "redirect") {
      expect(result.path).toBe("/portfolios/TH%2FEQ%2001/decisions/new");
    }
  });
});
