/**
 * Pure derivations for the Portfolio Compliance V2 workspace posture summary
 * and breach ordering. No HTTP — callers supply already-fetched data.
 */
import { severityRank } from "./formatters";
import { deriveBindingState } from "../../portfolio-workspace/lib/complianceBindingState";
import type {
  ApiPortfolioBreachView,
  ApiPortfolioRuleCatalogEntry,
} from "../../portfolio-workspace/services/portfolioComplianceApi";

export interface PortfolioCompliancePosture {
  openBreachCount: number;
  effectiveCount: number;
  scheduledCount: number;
  /** Highest-ranked severity among currently EFFECTIVE bindings, or null when none are effective. */
  strongestEffectiveSeverity: string | null;
}

export function derivePortfolioPosture(
  rules: ApiPortfolioRuleCatalogEntry[],
  breaches: ApiPortfolioBreachView[],
  todayIso: string,
): PortfolioCompliancePosture {
  let effectiveCount = 0;
  let scheduledCount = 0;
  let strongestEffectiveSeverity: string | null = null;

  for (const entry of rules) {
    const state = deriveBindingState(entry.binding, todayIso);
    if (state === "EFFECTIVE") {
      effectiveCount += 1;
      const severity = entry.binding?.severity ?? null;
      if (
        severity &&
        (!strongestEffectiveSeverity ||
          severityRank(severity) > severityRank(strongestEffectiveSeverity))
      ) {
        strongestEffectiveSeverity = severity;
      }
    } else if (state === "SCHEDULED") {
      scheduledCount += 1;
    }
  }

  const openBreachCount = breaches.filter((b) => b.status === "OPEN").length;

  return { openBreachCount, effectiveCount, scheduledCount, strongestEffectiveSeverity };
}

/** Orders portfolio breaches by severity strength (BLOCK first), then most recent first. */
export function sortPortfolioBreaches(
  breaches: ApiPortfolioBreachView[],
): ApiPortfolioBreachView[] {
  return [...breaches].sort((a, b) => {
    const rankDiff = severityRank(b.severity ?? "") - severityRank(a.severity ?? "");
    if (rankDiff !== 0) return rankDiff;
    return (b.created_at || "").localeCompare(a.created_at || "");
  });
}
