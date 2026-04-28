package valueobject

// ProcessStepKey identifies one step in the stock investment lifecycle.
type ProcessStepKey string

const (
	ProcessStepAnalysisReport      ProcessStepKey = "ANALYSIS_REPORT"
	ProcessStepInvestmentDecision  ProcessStepKey = "INVESTMENT_DECISION"
	ProcessStepInvestmentExecution ProcessStepKey = "INVESTMENT_EXECUTION"
	ProcessStepInvestmentReview    ProcessStepKey = "INVESTMENT_REVIEW"
)

// IsValid reports whether s is a known investment process step.
func (s ProcessStepKey) IsValid() bool {
	switch s {
	case ProcessStepAnalysisReport,
		ProcessStepInvestmentDecision,
		ProcessStepInvestmentExecution,
		ProcessStepInvestmentReview:
		return true
	default:
		return false
	}
}
