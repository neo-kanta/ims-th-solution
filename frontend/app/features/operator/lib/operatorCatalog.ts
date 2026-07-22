import type { AppTranslationKey } from "~/composables/useI18n";

export type CapabilityState = "ready" | "limited" | "denied";

export interface OperatorWorkflowDefinition {
  code: "OP-01" | "OP-02" | "OP-03";
  titleKey: AppTranslationKey;
  descriptionKey: AppTranslationKey;
  /** Function permission gating this workflow's entry point. */
  requiredPermission: string;
  /**
   * True when the workflow is fully API-backed end to end; false when a
   * documented backend contract gap means only part of the workflow is
   * available today. A "limited" workflow still opens — it discloses what's
   * missing rather than claiming Ready.
   */
  fullyAvailable: boolean;
  limitedReasonKey?: AppTranslationKey;
  routePath: string;
}

/**
 * The Operations Directory catalog. Single source of truth for the three
 * operator workflows — see docs/MANAGER for the backend contract audit that
 * determined `fullyAvailable` for OP-02/OP-03.
 */
export const OPERATOR_CATALOG: readonly OperatorWorkflowDefinition[] = [
  {
    code: "OP-01",
    titleKey: "operator.directory.workflows.op01.title",
    descriptionKey: "operator.directory.workflows.op01.description",
    requiredPermission: "INVESTMENT_DECISION_MANAGE",
    fullyAvailable: true,
    routePath: "/investment/operator/decision/new",
  },
  {
    code: "OP-02",
    titleKey: "operator.directory.workflows.op02.title",
    descriptionKey: "operator.directory.workflows.op02.description",
    requiredPermission: "INVESTMENT_DECISION_APPROVE",
    // Execution eligibility/lifecycle state now renders from the typed
    // GET /portfolios/{portfolioCode}/executions contract (matched by
    // decision_id) — see DecisionApprovalGrid.vue + lib/executionState.ts.
    fullyAvailable: true,
    routePath: "/investment/decision",
  },
  {
    code: "OP-03",
    titleKey: "operator.directory.workflows.op03.title",
    descriptionKey: "operator.directory.workflows.op03.description",
    requiredPermission: "INVESTMENT_DECISION_VIEW",
    // Execution/fill and trade-confirmation status now render from the
    // typed executions/confirmations contract — see
    // usePrintSummarySections.ts.
    fullyAvailable: true,
    routePath: "/investment/operator/review",
  },
];

/**
 * Derives real capability from permission + contract availability. Never
 * returns "ready" for a workflow the caller cannot open, and never returns
 * "ready" for a workflow whose backend contract is incomplete.
 */
export function deriveCapability(
  def: OperatorWorkflowDefinition,
  hasPermission: (code: string) => boolean,
): CapabilityState {
  if (!hasPermission(def.requiredPermission)) return "denied";
  return def.fullyAvailable ? "ready" : "limited";
}

export interface OperatorCatalogRow extends OperatorWorkflowDefinition {
  title: string;
  description: string;
  capability: CapabilityState;
  statusLabel: string;
}

/**
 * Filters a list of resolved catalog rows by a free-text query against code,
 * title, description, and status label — used by the directory search box.
 */
export function filterOperatorRows<T extends { code: string; title: string; description: string; statusLabel: string }>(
  rows: readonly T[],
  query: string,
): T[] {
  const q = query.toLowerCase().trim();
  if (!q) return [...rows];
  return rows.filter(
    (row) =>
      row.code.toLowerCase().includes(q) ||
      row.title.toLowerCase().includes(q) ||
      row.description.toLowerCase().includes(q) ||
      row.statusLabel.toLowerCase().includes(q),
  );
}
