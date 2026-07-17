/**
 * Pure page-state resolution for the New Decision order ticket, kept
 * separate from PortfolioDecisionNewView.vue so it is unit-testable without
 * mounting a Vue component (this project's Vitest setup has no Nuxt/component
 * runtime).
 *
 * `usePortfolioContext` (portfolio-workspace, shared by unrelated views) only
 * exposes an error *message*, not an HTTP status, and is out of scope to
 * edit here. The backend's resolvePortfolioByCode
 * (backend/internal/investment/transport/handler/portfolio_v2_handler.go)
 * returns exactly two literal messages for the cases this page must tell
 * apart — "portfolio not found" (404) and "no access to this portfolio"
 * (403) — so those are matched directly instead of guessing at a status code
 * that was never surfaced.
 */

export type PortfolioContextErrorKind = "not-found" | "permission-denied" | "unknown";

export function classifyPortfolioContextError(
  message: string | null | undefined,
): PortfolioContextErrorKind {
  if (!message) return "unknown";
  const normalized = message.toLowerCase();
  if (normalized.includes("not found")) return "not-found";
  if (normalized.includes("no access") || normalized.includes("forbidden")) {
    return "permission-denied";
  }
  return "unknown";
}

export type DecisionNewPageState =
  | { kind: "loading" }
  | { kind: "not-found" }
  | { kind: "permission-denied"; message: string }
  | { kind: "error"; message: string }
  | { kind: "ready" };

export interface DecisionNewPageStateInput {
  portfolioCode: string;
  /** The currently loaded portfolio descriptor's own code, or null/undefined
   * when nothing has loaded yet. Comparing this against `portfolioCode`
   * catches a stale response landing after the user has already switched
   * portfolios — a slow load for the old code must never be rendered as the
   * new code's data. */
  loadedPortfolioCode: string | null | undefined;
  error: string | null;
}

export function resolveDecisionNewPageState(
  input: DecisionNewPageStateInput,
): DecisionNewPageState {
  if (input.loadedPortfolioCode && input.loadedPortfolioCode === input.portfolioCode) {
    return { kind: "ready" };
  }
  if (input.error) {
    const kind = classifyPortfolioContextError(input.error);
    if (kind === "not-found") return { kind: "not-found" };
    if (kind === "permission-denied") return { kind: "permission-denied", message: input.error };
    return { kind: "error", message: input.error };
  }
  return { kind: "loading" };
}
