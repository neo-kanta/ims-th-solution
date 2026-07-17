/** Builds a Portfolio Compliance V2 workspace route from a business code. Never accepts a UUID. */
export function buildPortfolioComplianceHref(code: string): string {
  return `/portfolios/${encodeURIComponent(code)}/compliance`;
}
