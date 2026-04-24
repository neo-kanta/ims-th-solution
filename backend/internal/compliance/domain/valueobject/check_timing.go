package valueobject

// CheckTiming identifies when a compliance check is triggered.
type CheckTiming string

const (
	TimingPreTrade  CheckTiming = "PRE_TRADE"
	TimingPostTrade CheckTiming = "POST_TRADE"
	TimingPeriodic  CheckTiming = "PERIODIC"
)

// IsValid returns true if the timing is a known value.
func (t CheckTiming) IsValid() bool {
	switch t {
	case TimingPreTrade, TimingPostTrade, TimingPeriodic:
		return true
	}
	return false
}
