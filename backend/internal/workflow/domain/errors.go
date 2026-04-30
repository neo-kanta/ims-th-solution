package domain

import "fmt"

// ErrInvalidTransition is returned when a state transition is structurally
// forbidden (wrong current state) or blocked by a guard condition.
// The transport layer maps this to HTTP 422 Unprocessable Entity.
type ErrInvalidTransition struct {
	Code            string // e.g. "WORKFLOW_NOT_BUSINESS_DAY"
	CurrentState    string
	AttemptedAction string
	Reason          string
	Details         map[string]any // included in the JSON error body
}

func (e *ErrInvalidTransition) Error() string {
	return fmt.Sprintf("[%s] cannot %s from %s: %s",
		e.Code, e.AttemptedAction, e.CurrentState, e.Reason)
}

// ErrWorkflowDayExists is returned when OPEN_DAY is attempted for a
// (contract_id, business_date) that already has a persisted record.
// Indicates a concurrent duplicate request or a logic error in the caller.
// The transport layer maps this to HTTP 409 Conflict.
type ErrWorkflowDayExists struct {
	ContractID   string
	BusinessDate string
	CurrentState string
}

func (e *ErrWorkflowDayExists) Error() string {
	return fmt.Sprintf(
		"workflow day already exists for contract %s on %s (current state: %s)",
		e.ContractID, e.BusinessDate, e.CurrentState,
	)
}

// ErrVersionConflict is returned when the optimistic lock check fails during
// an UPDATE. The caller must re-read the current state and retry if appropriate.
// The transport layer maps this to HTTP 409 Conflict.
type ErrVersionConflict struct{}

func (e *ErrVersionConflict) Error() string {
	return "workflow day state was modified concurrently; re-read and retry"
}

// ErrNotFound is returned when a required workflow day record does not exist
// for a command that expects one (e.g. APPROVE on a day that was never opened).
// The transport layer maps this to HTTP 422 (not 404, because the business date
// itself is valid — the day simply hasn't been started).
type ErrNotFound struct {
	ContractID   string
	BusinessDate string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf(
		"no workflow day found for contract %s on %s; open the day first",
		e.ContractID, e.BusinessDate,
	)
}

// ErrPostTradeBreachesBlockClose is returned when CLOSE_TRANSACTIONS is attempted
// while the IRG post-trade sweep still shows at least one BLOCK-level breach.
// The transport layer maps this to HTTP 422 Unprocessable Entity — the operator
// must resolve or override the breach(es) before retrying.
type ErrPostTradeBreachesBlockClose struct {
	ContractID   string
	BusinessDate string
	CheckGroupID string
	BreachCount  int
}

func (e *ErrPostTradeBreachesBlockClose) Error() string {
	return fmt.Sprintf(
		"cannot close transactions for contract %s on %s: %d IRG BLOCK breach(es) "+
			"(check group %s) must be resolved or overridden first",
		e.ContractID, e.BusinessDate, e.BreachCount, e.CheckGroupID,
	)
}

// ErrSelfApprovalForbidden is returned when the approver is the same
// user who opened the day. Maker and checker must differ.
// Transport layer maps to HTTP 403.
type ErrSelfApprovalForbidden struct {
	ContractID   string
	BusinessDate string
	ActorID      string
}

func (e *ErrSelfApprovalForbidden) Error() string {
	return fmt.Sprintf(
		"self-approval forbidden for contract %s on %s: "+
			"approver and day-opener must differ",
		e.ContractID, e.BusinessDate,
	)
}

// ErrCancelBlockedByTransactions is returned when CANCEL_DAY_START is attempted
// for a contract+date that already has investment transactions. Cancelling would
// orphan or destroy those transactions, so the operator must void them first.
// The transport layer maps this to HTTP 422 Unprocessable Entity.
type ErrCancelBlockedByTransactions struct {
	ContractID       string
	BusinessDate     string
	TransactionCount int
}

func (e *ErrCancelBlockedByTransactions) Error() string {
	return fmt.Sprintf(
		"cannot cancel day start for contract %s on %s: %d transaction(s) exist; "+
			"void all transactions before cancelling",
		e.ContractID, e.BusinessDate, e.TransactionCount,
	)
}
