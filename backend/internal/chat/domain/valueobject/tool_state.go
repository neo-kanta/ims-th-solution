package valueobject

// ToolState is the lifecycle state of a single tool invocation.
type ToolState string

const (
	ToolStateRequested ToolState = "requested"
	ToolStateSucceeded ToolState = "succeeded"
	ToolStateFailed    ToolState = "failed"
	ToolStateDenied    ToolState = "denied" // blocked by permission or write-policy
)

// String returns the enum value as a string.
func (s ToolState) String() string { return string(s) }
