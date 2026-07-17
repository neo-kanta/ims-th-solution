/**
 * Pure helpers for the /compliance control center's portfolio finder.
 * Routes are always built from `portfolioCode`, never from `portfolio.id`.
 */
import type { CompliancePortfolioOption } from "../types";

export function filterPortfolioOptions(
  options: CompliancePortfolioOption[],
  query: string,
  limit = 8,
): CompliancePortfolioOption[] {
  const term = query.trim().toLowerCase();
  if (!term) return [];
  return options
    .filter((option) => {
      const haystack = `${option.code} ${option.name}`.toLowerCase();
      return haystack.includes(term);
    })
    .slice(0, limit);
}

export function buildPortfolioComplianceHref(code: string): string {
  return `/portfolios/${encodeURIComponent(code)}/compliance`;
}
