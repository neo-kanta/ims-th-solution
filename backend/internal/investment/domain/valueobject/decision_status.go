package valueobject

// DecisionStatus tracks where an investment decision sits in the 4-step flow.
type DecisionStatus string

const (
	// DecisionStatusDraft — decision has been drafted but not yet submitted to OMS.
	DecisionStatusDraft DecisionStatus = "DRAFT"

	// DecisionStatusSubmitted — cleared by IRG pre-trade and queued for execution.
	DecisionStatusSubmitted DecisionStatus = "SUBMITTED"

	// DecisionStatusBlocked — rejected by IRG pre-trade (BLOCK verdict).
	DecisionStatusBlocked DecisionStatus = "BLOCKED"

	// DecisionStatusExecuted — order filled at the broker.
	DecisionStatusExecuted DecisionStatus = "EXECUTED"

	// DecisionStatusCancelled — cancelled by the trader before execution.
	DecisionStatusCancelled DecisionStatus = "CANCELLED"
)

// IsValid reports whether the status is a known transition state.
func (s DecisionStatus) IsValid() bool {
	switch s {
	case DecisionStatusDraft, DecisionStatusSubmitted, DecisionStatusBlocked,
		DecisionStatusExecuted, DecisionStatusCancelled:
		return true
	default:
		return false
	}
}
