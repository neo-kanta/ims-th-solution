/**
 * Workflow action → operator-UX metadata lookup.
 *
 * Permission codes mirror `backend/internal/workflow/permission/policies.go`
 * (the catalog the backend's middleware checks). The frontend enforcement is
 * UX only — backend `RequirePermission(...)` plus `HasFunctionPermission` is
 * authoritative. Keep this file the single source of truth for WORKFLOW_*
 * strings used in templates and stores.
 */
import type {
  WorkflowAction,
  WorkflowDialogTone,
} from "./types";

import type { AppTranslationKey } from "~/shared/i18n/messages";

export const WORKFLOW_PERMISSIONS = {
  VIEW: "WORKFLOW_VIEW",
  OPEN_DAY: "WORKFLOW_OPEN_DAY",
  APPROVE: "WORKFLOW_APPROVE",
  CANCEL_DAY_START: "WORKFLOW_CANCEL_DAY_START",
  CANCEL_APPROVAL: "WORKFLOW_CANCEL_APPROVAL",
  CLOSE_TRANSACTIONS: "WORKFLOW_CLOSE_TRANSACTIONS",
  CANCEL_TRANSACTION_CLOSE: "WORKFLOW_CANCEL_TRANSACTION_CLOSE",
  CLOSE_ACCOUNTING: "WORKFLOW_CLOSE_ACCOUNTING",
  ROLLBACK_ACCOUNTING_CLOSE: "WORKFLOW_ROLLBACK_ACCOUNTING_CLOSE",
  RUN_SCHEDULER: "WORKFLOW_RUN_SCHEDULER",
} as const;

export type WorkflowPermissionCode =
  (typeof WORKFLOW_PERMISSIONS)[keyof typeof WORKFLOW_PERMISSIONS];

export interface ActionOption {
  value: WorkflowAction;
  permission: WorkflowPermissionCode;
  tone: WorkflowDialogTone;
  /** Backend enforces ≥20 chars on cancel/rollback reasons. */
  requiresReason: boolean;
  /** i18n key for the human label; fallback shown in code. */
  labelKey: AppTranslationKey;
  labelFallback: string;
}

export const ACTION_CATALOG: Record<WorkflowAction, ActionOption> = {
  OPEN_DAY: {
    value: "OPEN_DAY",
    permission: WORKFLOW_PERMISSIONS.OPEN_DAY,
    tone: "primary",
    requiresReason: false,
    labelKey: "workflow.action.OPEN_DAY",
    labelFallback: "Investment Day Start",
  },
  APPROVE: {
    value: "APPROVE",
    permission: WORKFLOW_PERMISSIONS.APPROVE,
    tone: "primary",
    requiresReason: false,
    labelKey: "workflow.action.APPROVE",
    labelFallback: "Manager Approval",
  },
  CLOSE_TRANSACTIONS: {
    value: "CLOSE_TRANSACTIONS",
    permission: WORKFLOW_PERMISSIONS.CLOSE_TRANSACTIONS,
    tone: "primary",
    requiresReason: false,
    labelKey: "workflow.action.CLOSE_TRANSACTIONS",
    labelFallback: "Transaction Closing",
  },
  CLOSE_ACCOUNTING: {
    value: "CLOSE_ACCOUNTING",
    permission: WORKFLOW_PERMISSIONS.CLOSE_ACCOUNTING,
    tone: "primary",
    requiresReason: false,
    labelKey: "workflow.action.CLOSE_ACCOUNTING",
    labelFallback: "Accounting Closing",
  },
  CANCEL_DAY_START: {
    value: "CANCEL_DAY_START",
    permission: WORKFLOW_PERMISSIONS.CANCEL_DAY_START,
    tone: "warning",
    requiresReason: true,
    labelKey: "workflow.action.CANCEL_DAY_START",
    labelFallback: "Cancel Investment Day Start",
  },
  CANCEL_APPROVAL: {
    value: "CANCEL_APPROVAL",
    permission: WORKFLOW_PERMISSIONS.CANCEL_APPROVAL,
    tone: "warning",
    requiresReason: true,
    labelKey: "workflow.action.CANCEL_APPROVAL",
    labelFallback: "Cancel Manager Approval",
  },
  CANCEL_TRANSACTION_CLOSE: {
    value: "CANCEL_TRANSACTION_CLOSE",
    permission: WORKFLOW_PERMISSIONS.CANCEL_TRANSACTION_CLOSE,
    tone: "warning",
    requiresReason: true,
    labelKey: "workflow.action.CANCEL_TRANSACTION_CLOSE",
    labelFallback: "Cancel Transaction Closing",
  },
  ROLLBACK_ACCOUNTING_CLOSE: {
    value: "ROLLBACK_ACCOUNTING_CLOSE",
    permission: WORKFLOW_PERMISSIONS.ROLLBACK_ACCOUNTING_CLOSE,
    tone: "danger",
    requiresReason: true,
    labelKey: "workflow.action.ROLLBACK_ACCOUNTING_CLOSE",
    labelFallback: "Rollback Accounting Closing",
  },
};
