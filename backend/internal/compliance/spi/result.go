package spi

import vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"

// EvalResult is the output of a single rule evaluation.
type EvalResult struct {
	Status   vo.ComplianceStatus `json:"status"`
	Verdict  vo.Verdict          `json:"verdict"`
	Message  string              `json:"message"`
	Evidence vo.Evidence         `json:"evidence"`
}

// NotImplementedResult returns a standard PASS result for Phase 2 stubs.
func NotImplementedResult(typeID string) EvalResult {
	return EvalResult{
		Verdict: vo.VerdictPass,
		Message: typeID + ": not implemented in PoC — passing by default",
		Evidence: vo.Evidence{
			Metrics: map[string]string{"status": "NOT_IMPLEMENTED"},
		},
	}
}

// Explanation is the output of the Explain method.
type Explanation struct {
	PlainText  string      `json:"plain_text"`
	Structured vo.Evidence `json:"structured"`
}
