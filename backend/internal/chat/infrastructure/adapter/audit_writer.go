// Package adapter contains chat's outbound adapters — implementations of
// chat-owned ports that bridge to platform or other-module services. The
// chat module imports audit/domain for the Recorder interface (the same
// pattern the iam and investment modules use); it does NOT import any other
// module's internal package.
package adapter

import (
	"context"

	"github.com/google/uuid"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/service"
)

// NewRecorderAuditWriter adapts audit/domain.Recorder to chat's AuditWriter
// port. It uses RecordStrict so audit failures surface as turn failures —
// chat turns touch financial data once Slice C lands, and audit loss in
// those flows is unacceptable.
func NewRecorderAuditWriter(recorder auditdomain.Recorder) service.AuditWriter {
	if recorder == nil {
		recorder = auditdomain.NopRecorder{}
	}
	return &recorderAuditWriter{recorder: recorder}
}

type recorderAuditWriter struct {
	recorder auditdomain.Recorder
}

func (w *recorderAuditWriter) RecordTurnEvent(
	ctx context.Context,
	eventType string,
	actorID *uuid.UUID,
	sessionID string,
	ipAddress string,
	userAgent string,
	metadata map[string]interface{},
) error {
	return w.recorder.RecordStrict(
		ctx,
		actorID,
		eventType,
		"CHAT_SESSION",
		sessionID,
		ipAddress,
		userAgent,
		metadata,
	)
}
