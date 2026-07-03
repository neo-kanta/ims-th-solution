package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// ToolInvocation is one persisted tool call: its arguments, the verbatim raw
// result, and its lifecycle state. It is the provenance source of record —
// every figure the assistant cites traces back to a row like this.
type ToolInvocation struct {
	ID         uuid.UUID
	SessionID  uuid.UUID
	MessageID  *uuid.UUID // assistant message this call grounded; set after the turn persists
	ToolCallID string     // provider-assigned id
	ToolName   string
	MCPServer  string
	Arguments  json.RawMessage
	RawResult  json.RawMessage
	IsError    bool
	State      valueobject.ToolState
	Error      string
	// CorrelationID ties this tool call to the request/turn that issued it.
	CorrelationID string
	StartedAt     time.Time
	FinishedAt    *time.Time
}
