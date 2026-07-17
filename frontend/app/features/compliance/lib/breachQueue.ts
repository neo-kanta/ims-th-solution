/**
 * Pure helpers for the /compliance control center's open-breach queue.
 * No HTTP — callers supply already-fetched data.
 */
import { severityRank } from "./formatters";
import type { ComplianceBreach } from "../types";

/** Orders breaches by severity strength (BLOCK first), then most recent first. */
export function sortBreachesForQueue(
  breaches: ComplianceBreach[],
): ComplianceBreach[] {
  return [...breaches].sort((a, b) => {
    const rankDiff = severityRank(b.severity) - severityRank(a.severity);
    if (rankDiff !== 0) return rankDiff;
    return (b.createdAt || "").localeCompare(a.createdAt || "");
  });
}

export interface PortfolioLinkOption {
  id: string;
  code: string;
  name: string;
}

export interface ResolvedPortfolioLink {
  /** Human label to render. Never a raw UUID. */
  label: string;
  /** Route-safe portfolio code, or null when the portfolio could not be resolved. */
  code: string | null;
}

/**
 * Resolves a breach's portfolio id to a display label and a routable code
 * using an already-loaded, single-call directory (never one lookup call per
 * breach). Falls back to a neutral, caller-supplied label — never the raw
 * UUID — when the directory has no matching entry.
 */
export function resolvePortfolioLink(
  portfolioId: string | null | undefined,
  byId: Map<string, PortfolioLinkOption>,
  unavailableLabel: string,
): ResolvedPortfolioLink {
  if (!portfolioId) return { label: unavailableLabel, code: null };
  const hit = byId.get(portfolioId);
  if (!hit || !hit.code) return { label: unavailableLabel, code: null };
  const label = hit.name ? `${hit.code} — ${hit.name}` : hit.code;
  return { label, code: hit.code };
}

export { buildPortfolioComplianceHref } from "./portfolioLinks";
