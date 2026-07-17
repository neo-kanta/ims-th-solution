/**
 * Pure helpers for Portfolio Compliance V2 binding display — extracted from
 * PortfolioComplianceView.vue so effective-state derivation and the
 * duplicate-submission guard can be unit tested without a Vue rendering
 * harness (matches the project's existing pattern of testing the pure logic
 * layer a component delegates to, not the component itself — see
 * tests/approval-components.test.ts).
 */
import type {
  ApiBindRuleRequest,
  ApiPortfolioRuleBindingView,
} from "../services/portfolioComplianceApi";

/** Local draft state for the "bind rule to portfolio" form. */
export interface BindDraft {
  severity: string;
  effectiveFrom: string;
  effectiveTo: string;
}

/**
 * Builds the Portfolio Compliance V2 bind-rule request body. Portfolio code
 * is carried in the URL and the rule instance ID in the path — this payload
 * carries only `severity`/`priority`/`effective_from`/`effective_to`
 * (matching the generated `BindRuleRequest` schema and the backend's
 * `handler.BindRuleRequest` struct) and must never gain `fund_id`,
 * `FundID`, `contract_id`, `ContractID`, `portfolio_id`, `scope_id`, or
 * `scope_type` — those are resolved/derived entirely on the backend.
 */
export function buildBindRulePayload(draft: BindDraft): ApiBindRuleRequest {
  return {
    severity: draft.severity,
    effective_from: draft.effectiveFrom,
    effective_to: draft.effectiveTo || undefined,
  };
}

export type BindingEffectiveState =
  | "SCHEDULED"
  | "EFFECTIVE"
  | "EXPIRED"
  | "DEACTIVATED";

/**
 * Derives a binding's real-world effective state from `is_active` plus the
 * effective_from/effective_to window, evaluated against `todayIso`
 * (YYYY-MM-DD). `is_active === true` alone does not mean "currently
 * enforced" — the backend's ResolveApplicable additionally gates on the
 * effective window against the business date, so a binding scheduled in the
 * future or past its effective_to must not display as effective.
 *
 * Boundary semantics mirror compliance/lib/formatters.ts' deriveRuleStatus:
 * exactly on effective_from or effective_to counts as effective, not
 * scheduled/expired.
 */
export function deriveBindingState(
  binding: ApiPortfolioRuleBindingView | null | undefined,
  todayIso: string,
): BindingEffectiveState | null {
  if (!binding) return null;
  if (!binding.is_active) return "DEACTIVATED";
  const from = binding.effective_from ? binding.effective_from.slice(0, 10) : null;
  const to = binding.effective_to ? binding.effective_to.slice(0, 10) : null;
  if (from && from > todayIso) return "SCHEDULED";
  if (to && to < todayIso) return "EXPIRED";
  return "EFFECTIVE";
}

/** True when a bind/deactivate submission for `id` is already in flight. */
export function isSubmissionInFlight(
  busy: Record<string, boolean>,
  id: string,
): boolean {
  return busy[id] === true;
}
