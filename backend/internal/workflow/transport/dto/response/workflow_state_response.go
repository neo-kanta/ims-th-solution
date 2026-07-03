package response

import "time"

// WorkflowStateResponse is returned by GET /workflow/day-states/{contractId}
type WorkflowStateResponse struct {
	ContractID   string `json:"contractId"`
	BusinessDate string `json:"businessDate"` // YYYY-MM-DD

	// CurrentState is one of: NOT_STARTED, DAY_OPEN, MANAGER_APPROVED,
	// TRANSACTION_CLOSED, ACCOUNTING_CLOSED
	CurrentState string `json:"currentState"`

	// Persisted is false when no workflow__day_states row exists yet.
	// Clients can use this to distinguish a genuinely-started day from
	// a cancelled/never-started synthetic NOT_STARTED response.
	Persisted bool `json:"persisted"`

	TransactionsLocked bool `json:"transactionsLocked"`
	PendingReclose     bool `json:"pendingReclose"`
	RecloseCount       int  `json:"recloseCount"`

	// Per-stage timestamps (null until that stage is reached)
	OpenedAt            *time.Time `json:"openedAt"`
	OpenedBy            *string    `json:"openedBy"`
	ManagerApprovedAt   *time.Time `json:"managerApprovedAt"`
	ManagerApprovedBy   *string    `json:"managerApprovedBy"`
	TransactionClosedAt *time.Time `json:"transactionClosedAt"`
	AccountingClosedAt  *time.Time `json:"accountingClosedAt"`

	// AccountingDate / PrevAccountingDate surface the NAV cycle date and
	// the prior cycle stashed during a rollback. Format YYYY-MM-DD.
	AccountingDate     *string `json:"accountingDate,omitempty"`
	PrevAccountingDate *string `json:"prevAccountingDate,omitempty"`

	// AllowedActions is computed server-side from the current state.
	// The UI must use this list to decide which buttons to enable.
	AllowedActions []string `json:"allowedActions"`

	// Version is the optimistic lock counter. Clients should pass this
	// back in write requests (future: when version field is added to the
	// transition request). Nil when Persisted=false.
	Version *int `json:"version,omitempty"`

	// PreviousDay is populated only when Persisted=false (NOT_STARTED).
	// It lets the UI display the previous-day status before the operator
	// clicks "Open Day".
	PreviousDay *PreviousDayStatus `json:"previousDay,omitempty"`

	// BlockingReasons explains why AllowedActions may be empty even though
	// a transition appears structurally valid (e.g. previous day still open).
	BlockingReasons []BlockingReason `json:"blockingReasons,omitempty"`
}

// PreviousDayStatus summarises the state of the prior business day.
type PreviousDayStatus struct {
	BusinessDate string `json:"businessDate"`
	CurrentState string `json:"currentState"`
	Persisted    bool   `json:"persisted"`
}

// BlockingReason explains a specific guard-condition failure that prevents
// an otherwise structurally-valid action from being executed.
type BlockingReason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TransitionResponse is returned by POST /workflow/day-states/{contractId}/transitions
type TransitionResponse struct {
	TransitionID  string `json:"transitionId"`
	WorkflowDayID string `json:"workflowDayId"`
	ContractID    string `json:"contractId"`
	BusinessDate  string `json:"businessDate"`
	FromState     string `json:"fromState"`
	ToState       string `json:"toState"`
	OccurredAt    string `json:"occurredAt"` // RFC3339 UTC

	// ApprovalID is set when the transition was an APPROVE action.
	ApprovalID *string `json:"approvalId,omitempty"`
	// IsZeroTransaction is set when the APPROVE was on a zero-transaction day.
	IsZeroTransaction *bool `json:"isZeroTransaction,omitempty"`

	// AccountingDate is set when the transition was CLOSE_ACCOUNTING. Format
	// YYYY-MM-DD; distinct from BusinessDate because operators can backdate
	// or delay accounting cycles.
	AccountingDate string `json:"accountingDate,omitempty"`
}

// HistoryResponse is returned by GET /workflow/day-states/{contractId}/history
type HistoryResponse struct {
	ContractID   string            `json:"contractId"`
	BusinessDate string            `json:"businessDate"`
	Transitions  []TransitionEntry `json:"transitions"`
}

// TransitionEntry is one row in the history response.
type TransitionEntry struct {
	ID            string    `json:"id"`
	FromState     string    `json:"fromState"`
	ToState       string    `json:"toState"`
	Action        string    `json:"action"`
	ActorType     string    `json:"actorType"`
	ActorID       *string   `json:"actorId,omitempty"`
	ActorUsername string    `json:"actorUsername"`
	Reason        *string   `json:"reason,omitempty"`
	OccurredAt    time.Time `json:"occurredAt"`
	RequestID     string    `json:"requestId"`
}
