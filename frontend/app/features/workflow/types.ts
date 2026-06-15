/**
 * Workflow module — shared types.
 *
 * Wire DTO types come from the generated OpenAPI client
 * (`~/api/ims-api`). The unions below pin the string literal sets the
 * frontend recognises so we get exhaustive switches and i18n lookups.
 * Keep them aligned with `backend/internal/workflow` action / state
 * enums; the canonical map lives in `NAMING.md`.
 */
import type { components } from "~/api/ims-api";

/** Canonical backend action codes (Phase 2 Daily API). */
export type WorkflowAction =
  | "START_INVESTMENT_DAY"
  | "CANCEL_INVESTMENT_DAY"
  | "MANAGER_APPROVE"
  | "CANCEL_MANAGER_APPROVAL"
  | "CLOSE_TRANSACTION"
  | "CANCEL_TRANSACTION_CLOSE"
  | "CLOSE_ACCOUNTING"
  | "CANCEL_ACCOUNTING_CLOSE";

export const WORKFLOW_ACTIONS: readonly WorkflowAction[] = [
  "START_INVESTMENT_DAY",
  "CANCEL_INVESTMENT_DAY",
  "MANAGER_APPROVE",
  "CANCEL_MANAGER_APPROVAL",
  "CLOSE_TRANSACTION",
  "CANCEL_TRANSACTION_CLOSE",
  "CLOSE_ACCOUNTING",
  "CANCEL_ACCOUNTING_CLOSE",
];

/** Workflow day state. NOT_STARTED is a synthetic state for empty days. */
export type WorkflowStateCode =
  | "NOT_STARTED"
  | "INVESTMENT_DAY_STARTED"
  | "MANAGER_APPROVED"
  | "TRANSACTION_CLOSED"
  | "ACCOUNTING_CLOSED";

export const WORKFLOW_STATES: readonly WorkflowStateCode[] = [
  "NOT_STARTED",
  "INVESTMENT_DAY_STARTED",
  "MANAGER_APPROVED",
  "TRANSACTION_CLOSED",
  "ACCOUNTING_CLOSED",
];

/** Linear rank used for stage progress comparisons. */
export const STATE_RANK: Record<WorkflowStateCode, number> = {
  NOT_STARTED: 0,
  INVESTMENT_DAY_STARTED: 1,
  MANAGER_APPROVED: 2,
  TRANSACTION_CLOSED: 3,
  ACCOUNTING_CLOSED: 4,
};

/** Stage groups shown on the tracker. */
export type WorkflowStage =
  | "INVESTMENT_DAY_STARTED"
  | "MANAGER_APPROVED"
  | "TRANSACTION_CLOSED"
  | "ACCOUNTING_CLOSED";

export const WORKFLOW_STAGES: readonly WorkflowStage[] = [
  "INVESTMENT_DAY_STARTED",
  "MANAGER_APPROVED",
  "TRANSACTION_CLOSED",
  "ACCOUNTING_CLOSED",
];

export type WorkflowStageStatus =
  | "complete"
  | "active"
  | "upcoming"
  | "blocked";

export type WorkflowDialogTone = "primary" | "warning" | "danger";

/** Re-exports from generated OpenAPI for ergonomic consumption. */
export type WorkflowStateResponse =
  components["schemas"]["DailyWorkflowResponse"];
export type WorkflowHistoryResponse =
  components["schemas"]["DailyTransitionsResponse"];
export type WorkflowTransitionEntry =
  components["schemas"]["DailyTimelineEntry"];
export type WorkflowExecuteRequest =
  components["schemas"]["DailyExecuteRequest"];
export type WorkflowExecuteResponse =
  components["schemas"]["DailyWorkflowResponse"];
export type WorkflowBlockingReason = components["schemas"]["DailyBlockingReason"];
export type WorkflowModuleReadiness = components["schemas"]["DailyWorkflowResponse"]["moduleReadiness"];

export function isWorkflowAction(value: unknown): value is WorkflowAction {
  return (
    typeof value === "string" &&
    (WORKFLOW_ACTIONS as readonly string[]).includes(value)
  );
}

export function isWorkflowState(value: unknown): value is WorkflowStateCode {
  return (
    typeof value === "string" &&
    (WORKFLOW_STATES as readonly string[]).includes(value)
  );
}
