package request

// DailyExecuteRequest is the JSON body for
// POST /api/v1/workflow/daily/execute
type DailyExecuteRequest struct {
	// BusinessDate is mandatory; format YYYY-MM-DD (Asia/Bangkok calendar).
	BusinessDate string `json:"businessDate"`

	// OperationType is the canonical API operation name.
	// Valid values: START_INVESTMENT_DAY, CANCEL_INVESTMENT_DAY,
	//   MANAGER_APPROVE, CANCEL_MANAGER_APPROVAL,
	//   CLOSE_TRANSACTION, CANCEL_TRANSACTION_CLOSE,
	//   CLOSE_ACCOUNTING, CANCEL_ACCOUNTING_CLOSE
	OperationType string `json:"operationType"`

	// Reason is mandatory for cancel/rollback operations (min 20 chars).
	Reason *string `json:"reason,omitempty"`

	// Remark is optional free text from the operator. For cancel/rollback
	// operations it is accepted as an alias for reason.
	Remark string `json:"remark,omitempty"`

	// ZeroTransactionAttestation must be true when approving a day with no
	// investment transactions. Applies to MANAGER_APPROVE only.
	ZeroTransactionAttestation bool `json:"zeroTransactionAttestation,omitempty"`

	// AttestationReason is required when ZeroTransactionAttestation is true (min 30 chars).
	AttestationReason string `json:"attestationReason,omitempty"`

	// Notes is an optional free-text remark for MANAGER_APPROVE.
	Notes *string `json:"notes,omitempty"`

	// AccountingDate applies to CLOSE_ACCOUNTING. Format YYYY-MM-DD.
	// When omitted, defaults to BusinessDate.
	AccountingDate *string `json:"accountingDate,omitempty"`
}

// DailySettingsUpdateRequest is the JSON body for
// PUT /api/v1/workflow/settings
type DailySettingsUpdateRequest struct {
	// OperationType is the operation type to configure approvers for.
	OperationType string `json:"operationType"`

	// Approvers replaces the entire active approver list for this operation type.
	// Pass an empty array to remove all approvers.
	Approvers []ApproverInput `json:"approvers"`
}

// ApproverInput specifies one approver entry.
type ApproverInput struct {
	// AccountCode is the approver's account code. Required.
	AccountCode string `json:"accountCode"`
	// Username is a snapshot of the approver's display name. Optional.
	Username string `json:"username,omitempty"`
	// Role is a descriptive label (e.g. "FundManager"). Optional.
	Role string `json:"role,omitempty"`
}
