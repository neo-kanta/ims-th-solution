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

/** Canonical backend action codes. Do not invent new ones on the frontend. */
export type WorkflowAction =
  | "OPEN_DAY"
  | "APPROVE"
  | "CANCEL_DAY_START"
  | "CANCEL_APPROVAL"
  | "CLOSE_TRANSACTIONS"
  | "CANCEL_TRANSACTION_CLOSE"
  | "CLOSE_ACCOUNTING"
  | "ROLLBACK_ACCOUNTING_CLOSE";

export const WORKFLOW_ACTIONS: readonly WorkflowAction[] = [
  "OPEN_DAY",
  "APPROVE",
  "CANCEL_DAY_START",
  "CANCEL_APPROVAL",
  "CLOSE_TRANSACTIONS",
  "CANCEL_TRANSACTION_CLOSE",
  "CLOSE_ACCOUNTING",
  "ROLLBACK_ACCOUNTING_CLOSE",
];

/** Workflow day state. NOT_STARTED is a synthetic state for empty days. */
export type WorkflowStateCode =
  | "NOT_STARTED"
  | "DAY_OPEN"
  | "MANAGER_APPROVED"
  | "TRANSACTION_CLOSED"
  | "ACCOUNTING_CLOSED";

export const WORKFLOW_STATES: readonly WorkflowStateCode[] = [
  "NOT_STARTED",
  "DAY_OPEN",
  "MANAGER_APPROVED",
  "TRANSACTION_CLOSED",
  "ACCOUNTING_CLOSED",
];

/** Linear rank used for stage progress comparisons. */
export const STATE_RANK: Record<WorkflowStateCode, number> = {
  NOT_STARTED: 0,
  DAY_OPEN: 1,
  MANAGER_APPROVED: 2,
  TRANSACTION_CLOSED: 3,
  ACCOUNTING_CLOSED: 4,
};

/** Stage groups shown on the tracker. */
export type WorkflowStage =
  | "DAY_OPEN"
  | "MANAGER_APPROVED"
  | "TRANSACTION_CLOSED"
  | "ACCOUNTING_CLOSED";

export const WORKFLOW_STAGES: readonly WorkflowStage[] = [
  "DAY_OPEN",
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
  components["schemas"]["WorkflowStateResponse"];
export type WorkflowHistoryResponse =
  components["schemas"]["HistoryResponse"];
export type WorkflowTransitionEntry =
  components["schemas"]["TransitionEntry"];
export type WorkflowExecuteRequest =
  components["schemas"]["ExecuteTransitionRequest"];
export type WorkflowExecuteResponse =
  components["schemas"]["TransitionResponse"];
export type WorkflowBlockingReason = components["schemas"]["BlockingReason"];
export type WorkflowPreviousDayStatus =
  components["schemas"]["PreviousDayStatus"];

export function isWorkflowAction(value: unknown): value is WorkflowAction {
  return (
    typeof value === "string"
    && (WORKFLOW_ACTIONS as readonly string[]).includes(value)
  );
}

export function isWorkflowState(value: unknown): value is WorkflowStateCode {
  return (
    typeof value === "string"
    && (WORKFLOW_STATES as readonly string[]).includes(value)
  );
}
