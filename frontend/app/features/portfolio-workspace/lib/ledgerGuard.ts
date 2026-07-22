/**
 * Pure guard logic for the portfolio-code-routed ledger/new (simulate + post
 * transaction) page. Extracted so the MODEL-portfolio block can be unit
 * tested without a Vue rendering harness.
 *
 * IMPORTANT (frontend-only enforcement — see manager handoff): as of this
 * session, `backend/internal/investment/application/command/post_transaction.go`
 * and the legacy `RunValuation` service contain NO check against
 * `portfolio_type == "MODEL"`. This guard is a UX affordance only — it hides
 * the ledger-entry form for a MODEL portfolio in this workspace, it does not
 * and cannot enforce the "MODEL has no ledger" business rule at the API
 * level. A MODEL portfolio's ledger is still writable through any other
 * authenticated caller of `POST /portfolios/{portfolioCode}/transactions`
 * until that backend rule is implemented.
 */
export type LedgerPortfolioType = "LIVE" | "SIMULATION" | "MODEL" | string;

/** True when this workspace should offer the ledger simulate/post form at all. */
export function canEnterLedgerTransaction(
  portfolioType: LedgerPortfolioType,
): boolean {
  return portfolioType !== "MODEL";
}

/**
 * Distinguishes an official (LIVE) ledger post from a paper/non-official
 * (SIMULATION) one so the UI copy never implies a SIMULATION posting is a
 * real trade or an official accounting result (docs/MANAGER/MEMORY.md).
 */
export type LedgerIntent = "OFFICIAL" | "PAPER" | "BLOCKED";

export function ledgerIntentFor(portfolioType: LedgerPortfolioType): LedgerIntent {
  if (portfolioType === "MODEL") return "BLOCKED";
  if (portfolioType === "SIMULATION") return "PAPER";
  return "OFFICIAL";
}
