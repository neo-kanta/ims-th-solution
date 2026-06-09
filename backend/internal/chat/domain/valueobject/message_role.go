package valueobject

// MessageRole is the role of a single conversation message. Tool results
// arrive as role=tool when Slice C wires the agent loop; Slice A only emits
// user and assistant.
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// IsValid reports whether r is one of the known roles.
func (r MessageRole) IsValid() bool {
	switch r {
	case RoleSystem, RoleUser, RoleAssistant, RoleTool:
		return true
	}
	return false
}

// String returns the enum value as a string.
func (r MessageRole) String() string { return string(r) }
