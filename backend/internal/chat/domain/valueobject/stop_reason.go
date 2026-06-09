package valueobject

// StopReason is the canonical reason a provider stopped generating. Each
// adapter maps its native stop-reason strings to one of these values; the
// agent loop branches on the canonical value, not the vendor's.
type StopReason string

const (
	StopReasonEndTurn      StopReason = "end_turn"
	StopReasonToolUse      StopReason = "tool_use"
	StopReasonMaxTokens    StopReason = "max_tokens"
	StopReasonStopSequence StopReason = "stop_sequence"
	StopReasonError        StopReason = "error"
	StopReasonUnknown      StopReason = "unknown"
)

// String returns the enum value as a string.
func (s StopReason) String() string { return string(s) }
