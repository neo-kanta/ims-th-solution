/**
 * Pure validation for the Portfolio Compliance V2 bind-rule drawer. Returns
 * error codes (not localized copy) so the component maps them through `t()`.
 */
import type { BindDraft } from "../../portfolio-workspace/lib/complianceBindingState";

export type BindDraftFieldError = "REQUIRED" | "INVALID_DATE" | "BEFORE_FROM";

export interface BindDraftValidation {
  effectiveFromError: BindDraftFieldError | null;
  effectiveToError: BindDraftFieldError | null;
  isValid: boolean;
}

function isValidIsoDate(value: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  const d = new Date(`${value}T00:00:00Z`);
  return !Number.isNaN(d.getTime());
}

export function validateBindDraft(draft: BindDraft): BindDraftValidation {
  let effectiveFromError: BindDraftFieldError | null = null;
  let effectiveToError: BindDraftFieldError | null = null;

  if (!draft.effectiveFrom.trim()) {
    effectiveFromError = "REQUIRED";
  } else if (!isValidIsoDate(draft.effectiveFrom)) {
    effectiveFromError = "INVALID_DATE";
  }

  if (draft.effectiveTo.trim()) {
    if (!isValidIsoDate(draft.effectiveTo)) {
      effectiveToError = "INVALID_DATE";
    } else if (!effectiveFromError && draft.effectiveTo < draft.effectiveFrom) {
      effectiveToError = "BEFORE_FROM";
    }
  }

  return {
    effectiveFromError,
    effectiveToError,
    isValid: effectiveFromError === null && effectiveToError === null,
  };
}
