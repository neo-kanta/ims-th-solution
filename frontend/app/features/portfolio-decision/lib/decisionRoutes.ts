/**
 * Portfolio-scoped decision route builders. Centralised so every
 * navigation call in this feature encodes portfolioCode/decisionId
 * consistently (docs/frontend/portfolio-v2-frontend-ddd.md: "URL-encode the
 * value when building API paths" applies equally to frontend route links).
 */
export function decisionsListPath(portfolioCode: string): string {
  return `/portfolios/${encodeURIComponent(portfolioCode)}/decisions`;
}

export function newDecisionPath(portfolioCode: string): string {
  return `${decisionsListPath(portfolioCode)}/new`;
}

export function decisionDetailPath(portfolioCode: string, decisionId: string): string {
  return `${decisionsListPath(portfolioCode)}/${encodeURIComponent(decisionId)}`;
}

export function executionsListPath(portfolioCode: string): string {
  return `/portfolios/${encodeURIComponent(portfolioCode)}/executions`;
}
