// Package adapter holds cross-module adapters for the approval module. They
// translate the approval module's required ports into concrete implementations
// sourced from other modules without leaking those modules' internal types.
package adapter

import (
	"context"

	"github.com/google/uuid"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
)

// AuditLoggerAdapter bridges the shared audit Recorder to approval's AuditPort.
type AuditLoggerAdapter struct {
	recorder auditdomain.Recorder
}

// NewAuditLoggerAdapter wires the adapter; a nil recorder degrades to a no-op.
func NewAuditLoggerAdapter(recorder auditdomain.Recorder) *AuditLoggerAdapter {
	if recorder == nil {
		recorder = auditdomain.NopRecorder{}
	}
	return &AuditLoggerAdapter{recorder: recorder}
}

// Record implements domain.AuditPort.
func (a *AuditLoggerAdapter) Record(ctx context.Context, actorID *uuid.UUID, eventType, targetType, targetID string, metadata map[string]any) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.Record(ctx, actorID, eventType, targetType, targetID, "", "", metadata)
}

var _ domain.AuditPort = (*AuditLoggerAdapter)(nil)
