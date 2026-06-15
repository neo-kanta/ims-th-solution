package valueobject

// WorkflowState represents a position in the day-level workflow state machine.
// The zero value ("") is not a valid state; use StateNotStarted explicitly.
type WorkflowState string

const (
	// StateNotStarted means no workflow record has been persisted for this
	// contract + business date. It is a synthetic state — no row exists in
	// workflow__day_states. The system never stores this value in the database.
	StateNotStarted WorkflowState = "NOT_STARTED"

	// StateInvestmentDayStarted means Investment Day Start has been executed.
	// Transactions are permitted from this point. Canonical name (Phase 2).
	StateInvestmentDayStarted WorkflowState = "INVESTMENT_DAY_STARTED"

	// StateDayOpen is the Phase 1 backward-compat alias for StateInvestmentDayStarted.
	// Existing database rows written before migration 20260613000007 carry this value.
	// Will be removed in the Phase 2 cleanup migration after all rows are renamed.
	StateDayOpen WorkflowState = "DAY_OPEN"

	// StateManagerApprovedEOD means the fund manager has approved all transactions
	// for the day. Transactions are locked; no new trades are accepted. Canonical name (Phase 2).
	StateManagerApprovedEOD WorkflowState = "MANAGER_APPROVED_END_OF_DAY"

	// StateManagerApproved is the Phase 1 backward-compat alias for StateManagerApprovedEOD.
	// Will be removed in the Phase 2 cleanup migration after all rows are renamed.
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
	case StateNotStarted,
		StateInvestmentDayStarted, StateDayOpen,
		StateManagerApprovedEOD, StateManagerApproved,
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

// IsOpenForTrading reports whether new investment transactions are permitted.
// True for both the canonical INVESTMENT_DAY_STARTED and the Phase 1 compat alias DAY_OPEN.
func (s WorkflowState) IsOpenForTrading() bool {
	return s == StateInvestmentDayStarted || s == StateDayOpen
}

// IsManagerApproved reports whether the day is in either of the two manager-approved
// state representations: the canonical MANAGER_APPROVED_END_OF_DAY and the Phase 1
// backward-compat alias MANAGER_APPROVED.
func (s WorkflowState) IsManagerApproved() bool {
	return s == StateManagerApprovedEOD || s == StateManagerApproved
}

// IsTransactionLocked reports whether new trades are blocked.
// True from manager approval onwards (covers both canonical and compat state names).
func (s WorkflowState) IsTransactionLocked() bool {
	return s.IsManagerApproved() || s == StateTransactionClosed || s == StateAccountingClosed
}

// ToAPIName normalises the stored state string to the canonical API name exposed
// by the new daily endpoints. Phase 1 compat aliases are mapped to their Phase 2
// equivalents so that the API surface is consistent regardless of which migration
// epoch wrote the row.
func (s WorkflowState) ToAPIName() string {
	switch s {
	case StateDayOpen:
		return string(StateInvestmentDayStarted)
	case StateManagerApprovedEOD:
		return string(StateManagerApproved)
	default:
		return string(s)
	}
}

// AllowedActions returns every WorkflowAction that is structurally permitted
// from state s. Permission and guard conditions are checked separately; this
// function answers only the state-machine topology question.
func AllowedActions(s WorkflowState) []WorkflowAction {
	switch s {
	case StateNotStarted:
		return []WorkflowAction{ActionOpenDay}
	case StateInvestmentDayStarted, StateDayOpen:
		return []WorkflowAction{ActionApprove, ActionCancelDayStart}
	case StateManagerApprovedEOD, StateManagerApproved:
		return []WorkflowAction{ActionCloseTransactions, ActionCancelApproval}
	case StateTransactionClosed:
		return []WorkflowAction{ActionCloseAccounting, ActionCancelTransactionClose}
	case StateAccountingClosed:
		return []WorkflowAction{ActionRollbackAccountingClose}
	default:
		return nil
	}
}
