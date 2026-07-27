/**
 * Pure guard logic for the portfolio-code-routed ledger/new (simulate + post
 * transaction) page. Extracted so the MODEL-portfolio block can be unit
 * tested without a Vue rendering harness.
 *
 * IMPORTANT (frontend-only enforcement for BUY/SELL — see manager handoff):
 * `backend/internal/investment/application/command/post_transaction.go` now
 * DOES reject a MODEL portfolio's cash movement (CASH_IN/CASH_OUT/FEE/
 * DIVIDEND) — see the LIVE cash-transaction approval gate (Stage 2,
 * `docs/investment/live-cash-transaction-approval-design.md`). BUY/SELL and
 * the legacy `RunValuation` service still have NO check against
 * `portfolio_type == "MODEL"` (BUY/SELL gating was explicitly out of Stage
 * 2's scope). This guard remains a UX affordance for the whole ledger-entry
 * form — it hides the form for a MODEL portfolio in this workspace, but a
 * MODEL portfolio's BUY/SELL ledger is still writable through any other
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
