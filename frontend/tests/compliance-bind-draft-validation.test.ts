import { describe, expect, it } from "vitest";

import { validateBindDraft } from "../app/features/compliance/lib/bindDraftValidation";
import type { BindDraft } from "../app/features/portfolio-workspace/lib/complianceBindingState";

function draft(overrides: Partial<BindDraft> = {}): BindDraft {
  return {
    severity: "BLOCK",
    effectiveFrom: "2026-07-15",
    effectiveTo: "",
    ...overrides,
  };
}

describe("validateBindDraft", () => {
  it("is valid with only a required effective-from date", () => {
    const result = validateBindDraft(draft());
    expect(result.isValid).toBe(true);
    expect(result.effectiveFromError).toBeNull();
    expect(result.effectiveToError).toBeNull();
  });

  it("requires effective_from", () => {
    const result = validateBindDraft(draft({ effectiveFrom: "" }));
    expect(result.isValid).toBe(false);
    expect(result.effectiveFromError).toBe("REQUIRED");
  });

  it("rejects a malformed effective_from date", () => {
    const result = validateBindDraft(draft({ effectiveFrom: "15/07/2026" }));
    expect(result.effectiveFromError).toBe("INVALID_DATE");
  });

  it("allows a blank optional effective_to", () => {
    const result = validateBindDraft(draft({ effectiveTo: "" }));
    expect(result.effectiveToError).toBeNull();
  });

  it("rejects an effective_to before effective_from", () => {
    const result = validateBindDraft(
      draft({ effectiveFrom: "2026-07-15", effectiveTo: "2026-07-01" }),
    );
    expect(result.isValid).toBe(false);
    expect(result.effectiveToError).toBe("BEFORE_FROM");
  });

  it("accepts an effective_to equal to effective_from (single-day window)", () => {
    const result = validateBindDraft(
      draft({ effectiveFrom: "2026-07-15", effectiveTo: "2026-07-15" }),
    );
    expect(result.isValid).toBe(true);
  });

  it("accepts an effective_to after effective_from", () => {
    const result = validateBindDraft(
      draft({ effectiveFrom: "2026-07-15", effectiveTo: "2026-12-31" }),
    );
    expect(result.isValid).toBe(true);
  });

  it("rejects a malformed effective_to date", () => {
    const result = validateBindDraft(draft({ effectiveTo: "not-a-date" }));
    expect(result.effectiveToError).toBe("INVALID_DATE");
  });
});
