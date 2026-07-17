/**
 * Pure resolution logic shared by the legacy fund-scoped "new decision"
 * entry points (frontend/app/pages/investment/funds/[fundId]/operation/new.vue,
 * frontend/app/pages/investment/operator/[fundId]/operation/new.vue).
 *
 * Portfolio V2 (docs/frontend/portfolio-v2-frontend-ddd.md) replaces the old
 * fund-scoped decision form with a single portfolio-scoped one at
 * /portfolios/{portfolioCode}/decisions/new. These legacy routes no longer
 * render a form themselves — they resolve the fund's accessible portfolio(s)
 * and forward the user, never silently guessing when the choice is
 * ambiguous.
 */
import { newDecisionPath } from "./decisionRoutes";

export interface LegacyRedirectPortfolio {
  code: string;
  name: string;
}

export type LegacyRedirectResult =
  | { kind: "redirect"; path: string; portfolio: LegacyRedirectPortfolio }
  | { kind: "choose"; portfolios: LegacyRedirectPortfolio[] }
  | { kind: "empty" };

export function resolveLegacyFundRedirect(
  portfolios: ReadonlyArray<{ code?: string | null; name?: string | null }>,
): LegacyRedirectResult {
  const withCode: LegacyRedirectPortfolio[] = portfolios
    .filter((p): p is { code: string; name?: string | null } => Boolean(p.code))
    .map((p) => ({ code: p.code, name: p.name?.trim() || p.code }));

  if (withCode.length === 0) return { kind: "empty" };
  if (withCode.length === 1) {
    const portfolio = withCode[0]!;
    return { kind: "redirect", path: newDecisionPath(portfolio.code), portfolio };
  }
  return { kind: "choose", portfolios: withCode };
}
