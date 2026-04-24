package valueobject

// Evidence is the structured proof of how a rule arrived at its verdict.
// All numeric values are strings (decimal-formatted) to prevent float64 drift.
type Evidence struct {
	// Metrics are measured values (e.g., "proposed_concentration" -> "12.30").
	Metrics map[string]string `json:"metrics,omitempty"`

	// ThresholdBreached is the specific threshold that was violated, if any.
	ThresholdBreached *ThresholdBreach `json:"threshold_breached,omitempty"`

	// References are identifiers of related entities (issuer IDs, tickers, etc.).
	References map[string]string `json:"references,omitempty"`
}

// ThresholdBreach records the specific limit violation.
type ThresholdBreach struct {
	MetricName string `json:"metric_name"`
	Actual     string `json:"actual"`
	Limit      string `json:"limit"`
	Operator   string `json:"operator"` // "<=", ">=", "<", ">", "==", "!="
	Unit       string `json:"unit"`     // "percent", "amount", "count", "lots"
}
