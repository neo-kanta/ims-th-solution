package valueobject

// ComplianceStatus describes whether the compliance pipeline produced a
// decision from configured controls. It is deliberately separate from
// Verdict: PASS/WARN/BLOCK describes evaluated rules, while these statuses
// identify a control-plane gap that a LIVE portfolio must treat as fail-closed.
type ComplianceStatus string

const (
	ComplianceStatusEvaluated     ComplianceStatus = "COMPLIANCE_EVALUATED"
	ComplianceStatusNotConfigured ComplianceStatus = "COMPLIANCE_NOT_CONFIGURED"
	ComplianceStatusUnavailable   ComplianceStatus = "COMPLIANCE_UNAVAILABLE"
)
