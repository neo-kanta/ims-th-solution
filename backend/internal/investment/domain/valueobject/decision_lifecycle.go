package valueobject

// DecisionLifecycleStatus is the full lifecycle status of an investment
// decision aggregate. The narrower DecisionStatus type in decision_status.go
// is kept for backwards compatibility with the legacy pre-trade pipeline.
type DecisionLifecycleStatus string

const (
	DecisionLifecycleDraft                    DecisionLifecycleStatus = "DRAFT"
	DecisionLifecyclePendingApproval          DecisionLifecycleStatus = "PENDING_APPROVAL"
	DecisionLifecycleApproved                 DecisionLifecycleStatus = "APPROVED"
	DecisionLifecycleRejected                 DecisionLifecycleStatus = "REJECTED"
	DecisionLifecycleCancelled                DecisionLifecycleStatus = "CANCELLED"
	DecisionLifecycleReadyForExecution        DecisionLifecycleStatus = "READY_FOR_EXECUTION"
	DecisionLifecycleExecuted                 DecisionLifecycleStatus = "EXECUTED"
	DecisionLifecyclePendingComplianceRelease DecisionLifecycleStatus = "PENDING_COMPLIANCE_RELEASE"
)

// IsValid reports whether the lifecycle status is one of the accepted values.
func (s DecisionLifecycleStatus) IsValid() bool {
	switch s {
	case DecisionLifecycleDraft, DecisionLifecyclePendingApproval,
		DecisionLifecycleApproved, DecisionLifecycleRejected,
		DecisionLifecycleCancelled, DecisionLifecycleReadyForExecution,
		DecisionLifecycleExecuted, DecisionLifecyclePendingComplianceRelease:
		return true
	}
	return false
}

// IsTerminal reports whether the lifecycle status disallows further
// transitions other than execution bookkeeping.
func (s DecisionLifecycleStatus) IsTerminal() bool {
	switch s {
	case DecisionLifecycleRejected, DecisionLifecycleCancelled,
		DecisionLifecycleExecuted:
		return true
	}
	return false
}

// ExecutionStatus tracks where an execution sits in its lifecycle.
type ExecutionStatus string

const (
	ExecutionStatusPending           ExecutionStatus = "PENDING"
	ExecutionStatusExecuted          ExecutionStatus = "EXECUTED"
	ExecutionStatusPartiallyExecuted ExecutionStatus = "PARTIALLY_EXECUTED"
	ExecutionStatusCancelled         ExecutionStatus = "CANCELLED"
)

// IsValid reports whether the execution status is one of the accepted values.
func (s ExecutionStatus) IsValid() bool {
	switch s {
	case ExecutionStatusPending, ExecutionStatusExecuted,
		ExecutionStatusPartiallyExecuted, ExecutionStatusCancelled:
		return true
	}
	return false
}

// TradeConfirmationStatus tracks where a confirmation sits in its lifecycle.
type TradeConfirmationStatus string

const (
	TradeConfirmationPendingReview TradeConfirmationStatus = "PENDING_REVIEW"
	TradeConfirmationMatched       TradeConfirmationStatus = "MATCHED"
	TradeConfirmationMismatched    TradeConfirmationStatus = "MISMATCHED"
	TradeConfirmationReviewed      TradeConfirmationStatus = "REVIEWED"
)

// IsValid reports whether the confirmation status is one of the accepted
// values.
func (s TradeConfirmationStatus) IsValid() bool {
	switch s {
	case TradeConfirmationPendingReview, TradeConfirmationMatched,
		TradeConfirmationMismatched, TradeConfirmationReviewed:
		return true
	}
	return false
}
