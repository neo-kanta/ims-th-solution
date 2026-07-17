/**
 * Presentation helpers for the portfolio decision order ticket and lifecycle
 * views. Quantity/amount/price formatting reuses the ledger's number
 * formatters so the two investment surfaces stay visually consistent.
 */
export {
  formatMoney,
  formatQuantity,
  shortenId,
} from "../../investment-ledger/lib/ledgerFormat";
import type { AppTranslationKey } from "~/composables/useI18n";

export type DecisionLifecycleStatus =
  | "DRAFT"
  | "PENDING_COMPLIANCE_RELEASE"
  | "PENDING_APPROVAL"
  | "APPROVED"
  | "READY_FOR_EXECUTION"
  | "EXECUTED"
  | "REJECTED"
  | "CANCELLED";

export const DECISION_STATUS_KEYS: Readonly<Record<string, AppTranslationKey>> = {
  DRAFT: "portfolio.decisionNew.status.draft",
  SUBMITTED: "portfolio.decisionNew.status.submitted",
  BLOCKED: "portfolio.decisionNew.status.blocked",
  PENDING_COMPLIANCE_RELEASE: "portfolio.decisionNew.status.pendingComplianceRelease",
  PENDING_APPROVAL: "portfolio.decisionNew.status.pendingApproval",
  APPROVED: "portfolio.decisionNew.status.approved",
  READY_FOR_EXECUTION: "portfolio.decisionNew.status.readyForExecution",
  EXECUTED: "portfolio.decisionNew.status.executed",
  REJECTED: "portfolio.decisionNew.status.rejected",
  CANCELLED: "portfolio.decisionNew.status.cancelled",
};

export function decisionStatusKey(status: string | null | undefined): AppTranslationKey | null {
  if (!status) return null;
  return DECISION_STATUS_KEYS[status] ?? null;
}

export type WorkflowStageState = "complete" | "active" | "upcoming" | "blocked";

export interface DecisionWorkflowStage {
  key: string;
  labelKey: AppTranslationKey;
  state: WorkflowStageState;
}

const STAGE_KEYS = [
  "draft",
  "compliance_approval",
  "execution",
  "confirmation",
  "holdings_updated",
] as const;

const STAGE_LABEL_KEYS: Record<(typeof STAGE_KEYS)[number], AppTranslationKey> = {
  draft: "portfolio.decisionNew.workflow.draft",
  compliance_approval: "portfolio.decisionNew.workflow.complianceApproval",
  execution: "portfolio.decisionNew.workflow.execution",
  confirmation: "portfolio.decisionNew.workflow.confirmation",
  holdings_updated: "portfolio.decisionNew.workflow.holdingsUpdated",
};

/**
 * Maps a decision's lifecycle status onto the 5-stage conceptual workflow
 * shown to the user. This is illustrative, not authoritative — execution,
 * confirmation, and the holdings/cash update itself happen in the
 * execution/confirmation/ledger modules, which this screen does not touch.
 */
export function decisionWorkflowStages(
  status: string | null | undefined,
): DecisionWorkflowStage[] {
  const terminalBlocked = status === "REJECTED" || status === "CANCELLED";

  const rank: Record<string, number> = {
    DRAFT: 0,
    PENDING_COMPLIANCE_RELEASE: 1,
    PENDING_APPROVAL: 1,
    APPROVED: 2,
    READY_FOR_EXECUTION: 2,
    EXECUTED: 3,
  };
  const currentRank = status ? (rank[status] ?? (terminalBlocked ? 0 : -1)) : -1;

  return STAGE_KEYS.map((key, index) => {
    let state: WorkflowStageState = "upcoming";
    if (currentRank < 0) {
      state = index === 0 ? "upcoming" : "upcoming";
    } else if (terminalBlocked) {
      state = index === 0 ? "complete" : index === 1 ? "blocked" : "upcoming";
    } else if (index < currentRank) {
      state = "complete";
    } else if (index === currentRank) {
      state = "active";
    }
    return { key, labelKey: STAGE_LABEL_KEYS[key], state };
  });
}

export function estimatedConsideration(
  quantity: string,
  limitPrice: string,
  amount = "",
): number | null {
  if (amount.trim()) {
    const explicitAmount = Number(amount);
    return Number.isFinite(explicitAmount) && explicitAmount > 0
      ? explicitAmount
      : null;
  }
  const qty = Number(quantity);
  const price = Number(limitPrice);
  if (!Number.isFinite(qty) || !Number.isFinite(price) || qty <= 0 || price <= 0) {
    return null;
  }
  return qty * price;
}
