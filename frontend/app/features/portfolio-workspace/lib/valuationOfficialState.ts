/**
 * Pure derivation of the Valuations page's official-vs-indicative display
 * state. Extracted so the LIVE/SIMULATION/MODEL distinction (docs/MANAGER/
 * MEMORY.md: "SIMULATION is indicative/non-official and must never imply a
 * real trade or official accounting result"; "MODEL has no ... official
 * valuation") can be unit tested without a Vue rendering harness.
 *
 * `is_indicative`/`has_stale_inputs` are the backend's own signals on
 * `ValuationResponse` (see ims-api.d.ts). SIMULATION is additionally always
 * treated as indicative here as a UI-level safety net — the durable product
 * rule that a SIMULATION portfolio's numbers are never official must hold
 * even if a particular snapshot's `is_indicative` flag were ever false.
 */
export type PortfolioTypeForValuation = "LIVE" | "SIMULATION" | "MODEL" | string;

export interface ValuationLike {
  is_indicative?: boolean;
  has_stale_inputs?: boolean;
}

export type ValuationOfficialState =
  | "OFFICIAL"
  | "INDICATIVE"
  | "STALE_INDICATIVE"
  | "UNAVAILABLE";

export function deriveValuationOfficialState(
  valuation: ValuationLike | null | undefined,
  portfolioType: PortfolioTypeForValuation,
): ValuationOfficialState {
  // MODEL portfolios have no official valuation lifecycle at all — never
  // present a MODEL number as official or indicative, only unavailable.
  if (portfolioType === "MODEL") return "UNAVAILABLE";
  if (!valuation) return "UNAVAILABLE";

  const indicative = portfolioType === "SIMULATION" || valuation.is_indicative === true;
  if (!indicative) return "OFFICIAL";
  return valuation.has_stale_inputs ? "STALE_INDICATIVE" : "INDICATIVE";
}

/** True only when the Valuations page may render a number as an official figure. */
export function isOfficialValuationState(state: ValuationOfficialState): boolean {
  return state === "OFFICIAL";
}
