package service

import (
	"context"

	"github.com/google/uuid"
)

// AuditWriter is the chat module's audit port. It is intentionally narrower
// than the central audit.Recorder so chat's tests can stub a single method
// and so any future tightening (e.g. mandatory correlation keys) lives in
// the chat-side adapter without touching every module that consumes audit.
//
// metadata MUST carry at least session_id; turn events MUST carry message_id;
// Slice C will add tool_invocation_id. The central audit table is the spine —
// metadata supplies the pointers back into the chat tables (correction 2).
type AuditWriter interface {
	RecordTurnEvent(
		ctx context.Context,
		eventType string,
		actorID *uuid.UUID,
		sessionID string,
		ipAddress string,
		userAgent string,
		metadata map[string]interface{},
	) error
}
