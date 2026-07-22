/**
 * Portfolio-workspace route builders — mirrors
 * `features/portfolio-decision/lib/decisionRoutes.ts`'s pattern so every
 * navigation call encodes `portfolioCode` consistently.
 */
export function portfolioOverviewPath(portfolioCode: string): string {
  return `/portfolios/${encodeURIComponent(portfolioCode)}/overview`;
}
