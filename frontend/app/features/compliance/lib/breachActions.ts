/**
 * Pure override-eligibility rules shared by the breach table and detail
 * drawer, so "can this row be overridden, and why not" is defined once.
 * Override is available only when the breach is OPEN and the caller holds
 * IRG_OVERRIDE_BREACH — this mirrors (but does not replace) backend
 * authorization, which remains the source of truth.
 */
import type { ComplianceBreach } from "../types";

export type OverrideBlockedReason = "NOT_OPEN" | "NO_PERMISSION" | null;

export function canOverrideBreach(breach: ComplianceBreach, canOverride: boolean): boolean {
  return canOverride && breach.status === "OPEN";
}

export function overrideBlockedReason(
  breach: ComplianceBreach,
  canOverride: boolean,
): OverrideBlockedReason {
  if (breach.status !== "OPEN") return "NOT_OPEN";
  if (!canOverride) return "NO_PERMISSION";
  return null;
}
