package valueobject

// WorkflowState represents a position in the day-level workflow state machine.
// The zero value ("") is not a valid state; use StateNotStarted explicitly.
type WorkflowState string

const (
	// StateNotStarted means no workflow record has been persisted for this
	// contract + business date. It is a synthetic state — no row exists in
	// workflow__day_states. The system never stores this value in the database.
	StateNotStarted WorkflowState = "NOT_STARTED"

	// StateDayOpen means Investment Day Start has been executed.
	// Transactions are permitted from this point.
	StateDayOpen WorkflowState = "DAY_OPEN"

	// StateManagerApproved means the fund manager has approved all transactions
	// for the day. Transactions are locked; no new trades are accepted.
	StateManagerApproved WorkflowState = "MANAGER_APPROVED"

	// StateTransactionClosed means inventory reconciliation is confirmed.
	// Awaiting NAV / accounting postback.
	StateTransactionClosed WorkflowState = "TRANSACTION_CLOSED"

	// StateAccountingClosed means NAV has been posted and the day is fully closed.
	StateAccountingClosed WorkflowState = "ACCOUNTING_CLOSED"
)

// IsValid reports whether s is a known state value.
func (s WorkflowState) IsValid() bool {
	switch s {
	case StateNotStarted, StateDayOpen, StateManagerApproved,
		StateTransactionClosed, StateAccountingClosed:
		return true
	}
	return false
}

// IsPersisted reports whether this state corresponds to an actual database row.
// StateNotStarted is synthetic and never persisted.
func (s WorkflowState) IsPersisted() bool {
	return s != StateNotStarted
}

// AllowedActions returns every WorkflowAction that is structurally permitted
// from state s. Permission and guard conditions are checked separately; this
// function answers only the state-machine topology question.
func AllowedActions(s WorkflowState) []WorkflowAction {
	switch s {
	case StateNotStarted:
		return []WorkflowAction{ActionOpenDay}
	case StateDayOpen:
		return []WorkflowAction{ActionApprove, ActionCancelDayStart}
	case StateManagerApproved:
		return []WorkflowAction{ActionCloseTransactions, ActionCancelApproval}
	case StateTransactionClosed:
		return []WorkflowAction{ActionCloseAccounting, ActionCancelTransactionClose}
	case StateAccountingClosed:
		return []WorkflowAction{ActionRollbackAccountingClose}
	default:
		return nil
	}
}
