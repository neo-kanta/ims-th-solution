package valueobject

// WorkflowAction is the discriminator for a state transition request.
// It is sent by the caller in POST /workflow/day-states/{id}/transitions.
type WorkflowAction string

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
