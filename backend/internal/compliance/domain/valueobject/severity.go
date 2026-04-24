package valueobject

// Severity is the action-level assigned at the binding, not the rule type.
// The same rule type can be BLOCK in one mandate and WARN in another.
type Severity string

const (
	SeverityBlock           Severity = "BLOCK"
	SeverityWarn            Severity = "WARN"
	SeverityRequireApproval Severity = "REQUIRE_APPROVAL"
	SeverityMonitor         Severity = "MONITOR"
)

// IsValid returns true if the severity is a known value.
func (s Severity) IsValid() bool {
	_, ok := severityToMaxVerdict[s]
	return ok
}

// CapVerdict applies binding severity to a raw rule verdict.
// If the binding says MONITOR, even a rule BLOCK becomes PASS (logged only).
// If the binding says WARN, a rule BLOCK is downgraded to WARN.
func (s Severity) CapVerdict(raw Verdict) Verdict {
	maxVerdict, ok := severityToMaxVerdict[s]
	if !ok {
		return raw
	}
	if raw.Dominates(maxVerdict) {
		return maxVerdict
	}
	return raw
}

var severityToMaxVerdict = map[Severity]Verdict{
	SeverityBlock:           VerdictBlock,
	SeverityWarn:            VerdictWarn,
	SeverityRequireApproval: VerdictWarn,
	SeverityMonitor:         VerdictPass,
}
