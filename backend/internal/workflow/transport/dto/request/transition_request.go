package request

// ExecuteTransitionRequest is the JSON body for
// POST /api/v1/workflow/day-states/{contractId}/transitions
type ExecuteTransitionRequest struct {

	// BusinessDate is mandatory; format YYYY-MM-DD (Asia/Bangkok calendar).
	BusinessDate string `json:"businessDate"`

	// Action identifies the transition to execute.
	// Valid values: OPEN_DAY, APPROVE, CANCEL_DAY_START, CANCEL_APPROVAL,
	// CLOSE_TRANSACTIONS, CANCEL_TRANSACTION_CLOSE, CLOSE_ACCOUNTING,
	// ROLLBACK_ACCOUNTING_CLOSE
	Action string `json:"action"`

	// Reason is mandatory for cancel and rollback actions (min 20 chars).
	// Optional for forward transitions.
	Reason *string `json:"reason,omitempty"`

	// ZeroTransactionAttestation must be true when approving a day with no
	// investment transactions. Ignored for all other actions.
	ZeroTransactionAttestation bool `json:"zeroTransactionAttestation,omitempty"`

	// AttestationReason is required when ZeroTransactionAttestation is true.
	// Must be at least 30 characters.
	AttestationReason string `json:"attestationReason,omitempty"`

	// Notes is an optional free-text remark stored on the approval record.
	// Applies to APPROVE only.
	Notes *string `json:"notes,omitempty"`

	// IdempotencyKey is an optional client-generated UUID. When provided, a
	// duplicate request with the same key returns the original result without
	// re-executing the transition. (Full idempotency store is Batch 2.)
	IdempotencyKey *string `json:"idempotencyKey,omitempty"`
}
