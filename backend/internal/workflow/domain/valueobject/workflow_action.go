package valueobject

// WorkflowAction is the discriminator for a state transition request.
// It is sent by the caller in POST /workflow/day-states/{id}/transitions.
type WorkflowAction string

// Internal action constants (legacy API and command layer).
const (
	ActionOpenDay                 WorkflowAction = "OPEN_DAY"
	ActionCancelDayStart          WorkflowAction = "CANCEL_DAY_START"
	ActionApprove                 WorkflowAction = "APPROVE"
	ActionCancelApproval          WorkflowAction = "CANCEL_APPROVAL"
	ActionCloseTransactions       WorkflowAction = "CLOSE_TRANSACTIONS"
	ActionCancelTransactionClose  WorkflowAction = "CANCEL_TRANSACTION_CLOSE"
	ActionCloseAccounting         WorkflowAction = "CLOSE_ACCOUNTING"
	ActionRollbackAccountingClose WorkflowAction = "ROLLBACK_ACCOUNTING_CLOSE"
)

// API-facing operation type names accepted by the new daily endpoints.
// These are the canonical names exposed to the frontend; mapped internally to
// the action constants above by ParseAPIOperationType.
const (
	APIOpStartInvestmentDay    = "START_INVESTMENT_DAY"
	APIOpCancelInvestmentDay   = "CANCEL_INVESTMENT_DAY"
	APIOpManagerApprove        = "MANAGER_APPROVE"
	APIOpCancelManagerApproval = "CANCEL_MANAGER_APPROVAL"
	APIOpCloseTransaction      = "CLOSE_TRANSACTION"
	APIOpCancelTransactionClose = "CANCEL_TRANSACTION_CLOSE"
	APIOpCloseAccounting       = "CLOSE_ACCOUNTING"
	APIOpCancelAccountingClose = "CANCEL_ACCOUNTING_CLOSE"
)

// ParseAPIOperationType maps an API-facing operationType string to the internal
// WorkflowAction constant. Returns ("", false) for unknown values.
func ParseAPIOperationType(op string) (WorkflowAction, bool) {
	switch op {
	case APIOpStartInvestmentDay:
		return ActionOpenDay, true
	case APIOpCancelInvestmentDay:
		return ActionCancelDayStart, true
	case APIOpManagerApprove:
		return ActionApprove, true
	case APIOpCancelManagerApproval:
		return ActionCancelApproval, true
	case APIOpCloseTransaction:
		return ActionCloseTransactions, true
	case APIOpCancelTransactionClose:
		return ActionCancelTransactionClose, true
	case APIOpCloseAccounting:
		return ActionCloseAccounting, true
	case APIOpCancelAccountingClose:
		return ActionRollbackAccountingClose, true
	}
	return "", false
}

// ToAPIName returns the canonical API-facing operationType name for this action.
// Used when building transition history responses for the daily endpoints.
func (a WorkflowAction) ToAPIName() string {
	switch a {
	case ActionOpenDay:
		return APIOpStartInvestmentDay
	case ActionCancelDayStart:
		return APIOpCancelInvestmentDay
	case ActionApprove:
		return APIOpManagerApprove
	case ActionCancelApproval:
		return APIOpCancelManagerApproval
	case ActionCloseTransactions:
		return APIOpCloseTransaction
	case ActionCancelTransactionClose:
		return APIOpCancelTransactionClose
	case ActionCloseAccounting:
		return APIOpCloseAccounting
	case ActionRollbackAccountingClose:
		return APIOpCancelAccountingClose
	}
	return string(a)
}

// IsValid reports whether a is a recognised action value.
func (a WorkflowAction) IsValid() bool {
	switch a {
	case ActionOpenDay, ActionCancelDayStart,
		ActionApprove, ActionCancelApproval,
		ActionCloseTransactions, ActionCancelTransactionClose,
		ActionCloseAccounting, ActionRollbackAccountingClose:
		return true
	}
	return false
}

// RequiresReason reports whether a mandatory reason must accompany this action.
// Used in transport validation before the command handler is invoked.
func (a WorkflowAction) RequiresReason() bool {
	switch a {
	case ActionCancelDayStart, ActionCancelApproval,
		ActionCancelTransactionClose, ActionRollbackAccountingClose:
		return true
	}
	return false
}
