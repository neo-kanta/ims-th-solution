package valueobject

// Verdict is the outcome of a single rule evaluation.
type Verdict string

const (
	VerdictPass  Verdict = "PASS"
	VerdictWarn  Verdict = "WARN"
	VerdictBlock Verdict = "BLOCK"
)

// Dominates returns true if v takes precedence over other.
// BLOCK > WARN > PASS.
func (v Verdict) Dominates(other Verdict) bool {
	return verdictRank[v] > verdictRank[other]
}

// IsValid returns true if the verdict is a known value.
func (v Verdict) IsValid() bool {
	_, ok := verdictRank[v]
	return ok
}

var verdictRank = map[Verdict]int{
	VerdictPass:  0,
	VerdictWarn:  1,
	VerdictBlock: 2,
}
