/**
 * Pure query-building for the portfolio-scoped Watchlists page. Locks the
 * existing (user/personal-or-portfolio-scoped) watchlist feature to exactly
 * this workspace's resolved portfolio, so the page never lets the operator
 * pick a different portfolio or fall back to a personal watchlist — see
 * frontend/app/features/watchlist for the reused list/alert/drawer
 * components this feeds.
 */
export interface LockedWatchlistItemQuery {
  scope_type: "PORTFOLIO";
  portfolio_id: string;
  include_thresholds: true;
}

export interface LockedWatchlistAlertQuery {
  scope_type: "PORTFOLIO";
  portfolio_id: string;
}

/** Returns null when the portfolio's internal id has not resolved yet — callers must not query without it. */
export function buildLockedWatchlistItemQuery(
  portfolioId: string | null | undefined,
): LockedWatchlistItemQuery | null {
  if (!portfolioId) return null;
  return { scope_type: "PORTFOLIO", portfolio_id: portfolioId, include_thresholds: true };
}

export function buildLockedWatchlistAlertQuery(
  portfolioId: string | null | undefined,
): LockedWatchlistAlertQuery | null {
  if (!portfolioId) return null;
  return { scope_type: "PORTFOLIO", portfolio_id: portfolioId };
}
