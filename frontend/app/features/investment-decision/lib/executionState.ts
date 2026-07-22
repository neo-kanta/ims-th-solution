/**
 * Derives an honest execution-lifecycle state for one investment decision
 * from two REAL backend fields only:
 *
 *   - `DecisionResponse.status` (`DecisionLifecycleStatus` — see
 *     backend/internal/investment/domain/valueobject/decision_lifecycle.go):
 *     DRAFT | PENDING_APPROVAL | APPROVED | REJECTED | CANCELLED |
 *     READY_FOR_EXECUTION | EXECUTED | PENDING_COMPLIANCE_RELEASE
 *   - the matched `ExecutionResponse.status` (`ExecutionStatus`, same file):
 *     PENDING | EXECUTED | PARTIALLY_EXECUTED | CANCELLED
 *
 * There is no backend "FAILED" or "BLOCKED" execution status — this module
 * never invents one. A rejected/compliance-held decision is surfaced with
 * its own accurate label instead of being forced into a fabricated "failed"
 * bucket.
 */

export type ExecutionLifecycleState =
  | "draft"
  | "pending_approval"
  | "executable"
  | "rejected"
  | "cancelled"
  | "compliance_blocked"
  | "execution_pending"
  | "partially_filled"
  | "filled"
  | "execution_cancelled"
  | "unknown";

/**
 * `decisionStatus` is the decision's own lifecycle status. `executionStatus`
 * is the status of the execution matched to this decision (via
 * `ExecutionResponse.decision_id`), if one has been created yet. When an
 * execution exists, its status is authoritative for fill progress; when none
 * exists yet, the decision's own status tells the operator what's blocking
 * (or ready for) execution.
 */
export function deriveExecutionLifecycleState(
  decisionStatus: string | null | undefined,
  executionStatus: string | null | undefined,
): ExecutionLifecycleState {
  if (executionStatus) {
    switch (executionStatus) {
      case "PENDING":
        return "execution_pending";
      case "EXECUTED":
        return "filled";
      case "PARTIALLY_EXECUTED":
        return "partially_filled";
      case "CANCELLED":
        return "execution_cancelled";
      default:
        return "unknown";
    }
  }

  switch (decisionStatus) {
    case "DRAFT":
      return "draft";
    case "PENDING_APPROVAL":
      return "pending_approval";
    case "APPROVED":
    case "READY_FOR_EXECUTION":
      return "executable";
    case "REJECTED":
      return "rejected";
    case "CANCELLED":
      return "cancelled";
    case "PENDING_COMPLIANCE_RELEASE":
      return "compliance_blocked";
    case "EXECUTED":
      // Decision-level marker of completion when no execution row could be
      // matched (e.g. matched execution list did not include it) — trust the
      // decision rather than showing "unknown".
      return "filled";
    default:
      return "unknown";
  }
}

/**
 * Status-badge tone bucket, matching the tone vocabulary already used by
 * `AppStatusBadge` in `DecisionApprovalGrid.vue`/`DecisionDetailDrawer.vue`
 * (`pending` | `approved` | `rejected` | `inactive` | `success` | `neutral`).
 */
export type ExecutionStateTone =
  | "pending"
  | "approved"
  | "success"
  | "rejected"
  | "inactive"
  | "neutral";

export function executionStateTone(state: ExecutionLifecycleState): ExecutionStateTone {
  switch (state) {
    case "filled":
      return "success";
    case "executable":
    case "execution_pending":
      return "approved";
    case "partially_filled":
    case "pending_approval":
    case "draft":
      return "pending";
    case "rejected":
    case "compliance_blocked":
      return "rejected";
    case "cancelled":
    case "execution_cancelled":
      return "inactive";
    default:
      return "neutral";
  }
}
