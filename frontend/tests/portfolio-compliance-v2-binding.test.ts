import { describe, expect, it } from "vitest";

import {
  buildBindRulePayload,
  deriveBindingState,
  isSubmissionInFlight,
  type BindDraft,
} from "../app/features/portfolio-workspace/lib/complianceBindingState";
import type { ApiPortfolioRuleBindingView } from "../app/features/portfolio-workspace/services/portfolioComplianceApi";

// Fix 1 — Portfolio Compliance V2 identity contract. portfolioCode travels in
// the URL and the backend resolves portfolio_id/fund_id/actor from it and
// from the JWT; the bind-rule request body must carry none of those.
describe("Portfolio Compliance V2 bind payload", () => {
  const FORBIDDEN_KEYS = [
    "fund_id",
    "FundID",
    "contract_id",
    "ContractID",
    "portfolio_id",
    "scope_id",
    "scope_type",
  ];

  function draft(overrides: Partial<BindDraft> = {}): BindDraft {
    return {
      severity: "BLOCK",
      effectiveFrom: "2026-07-13",
      effectiveTo: "",
      ...overrides,
    };
  }

  it("sends only severity, effective_from, and effective_to", () => {
    const payload = buildBindRulePayload(
      draft({ effectiveTo: "2026-12-31" }),
    );
    expect(Object.keys(payload).sort()).toEqual(
      ["effective_from", "effective_to", "severity"].sort(),
    );
  });

  it("never includes a forbidden identity or scope field", () => {
    const payload = buildBindRulePayload(draft({ effectiveTo: "2026-12-31" }));
    for (const key of FORBIDDEN_KEYS) {
      expect(Object.prototype.hasOwnProperty.call(payload, key), key).toBe(false);
    }
  });

  it("omits effective_to (via undefined) when the draft leaves it blank", () => {
    const payload = buildBindRulePayload(draft({ effectiveTo: "" }));
    expect(payload.effective_to).toBeUndefined();
    // JSON.stringify drops undefined-valued keys, so this never reaches the
    // wire as `"effective_to": null` or an empty string.
    expect(JSON.parse(JSON.stringify(payload))).not.toHaveProperty("effective_to");
  });

  it("carries the portfolio code and rule instance id via the URL, not the body", () => {
    // buildBindRulePayload's signature takes only a BindDraft — there is no
    // parameter through which a caller could smuggle portfolio/fund identity
    // into the body even by mistake.
    expect(buildBindRulePayload.length).toBe(1);
  });
});

// Fix 5 — effective binding state. is_active alone does not mean "currently
// enforced"; the backend's ResolveApplicable also gates on effective_from/
// effective_to against the business date.
describe("deriveBindingState", () => {
  const TODAY = "2026-07-13";

  function binding(
    overrides: Partial<ApiPortfolioRuleBindingView> = {},
  ): ApiPortfolioRuleBindingView {
    return {
      binding_id: "b1",
      severity: "BLOCK",
      priority: 100,
      is_active: true,
      effective_from: "2026-01-01",
      effective_to: undefined,
      ...overrides,
    };
  }

  it("returns null when there is no binding at all", () => {
    expect(deriveBindingState(undefined, TODAY)).toBeNull();
    expect(deriveBindingState(null, TODAY)).toBeNull();
  });

  it("returns DEACTIVATED when is_active is false, regardless of window", () => {
    expect(
      deriveBindingState(
        binding({ is_active: false, effective_from: "2020-01-01", effective_to: "2099-01-01" }),
        TODAY,
      ),
    ).toBe("DEACTIVATED");
  });

  it("returns SCHEDULED strictly before effective_from", () => {
    expect(deriveBindingState(binding({ effective_from: "2026-07-14" }), TODAY)).toBe(
      "SCHEDULED",
    );
  });

  it("returns EFFECTIVE exactly on effective_from", () => {
    expect(deriveBindingState(binding({ effective_from: "2026-07-13" }), TODAY)).toBe(
      "EFFECTIVE",
    );
  });

  it("returns EFFECTIVE strictly between effective_from and effective_to", () => {
    expect(
      deriveBindingState(
        binding({ effective_from: "2026-01-01", effective_to: "2026-12-31" }),
        TODAY,
      ),
    ).toBe("EFFECTIVE");
  });

  it("returns EFFECTIVE exactly on effective_to", () => {
    expect(
      deriveBindingState(
        binding({ effective_from: "2026-01-01", effective_to: "2026-07-13" }),
        TODAY,
      ),
    ).toBe("EFFECTIVE");
  });

  it("returns EXPIRED strictly after effective_to", () => {
    expect(
      deriveBindingState(
        binding({ effective_from: "2026-01-01", effective_to: "2026-07-12" }),
        TODAY,
      ),
    ).toBe("EXPIRED");
  });

  it("returns EFFECTIVE when effective_to is null/absent (open-ended)", () => {
    expect(
      deriveBindingState(binding({ effective_from: "2026-01-01", effective_to: undefined }), TODAY),
    ).toBe("EFFECTIVE");
  });
});

// Fix 6 — duplicate bind/deactivate submission prevention.
describe("isSubmissionInFlight", () => {
  it("is false when the id has no busy entry", () => {
    expect(isSubmissionInFlight({}, "rule-1")).toBe(false);
  });

  it("is true only while the specific id's submission is busy", () => {
    const busy = { "rule-1": true, "rule-2": false };
    expect(isSubmissionInFlight(busy, "rule-1")).toBe(true);
    expect(isSubmissionInFlight(busy, "rule-2")).toBe(false);
    expect(isSubmissionInFlight(busy, "rule-3")).toBe(false);
  });
});
