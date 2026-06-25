package adapter

import (
	"context"

	"github.com/google/uuid"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
)

// AuditAdapter adapts audit.Recorder to the watchlist WatchlistAuditRecorder.
type AuditAdapter struct {
	recorder auditdomain.Recorder
}

func NewAuditAdapter(recorder auditdomain.Recorder) *AuditAdapter {
	return &AuditAdapter{recorder: recorder}
}

func (a *AuditAdapter) Record(ctx context.Context, actorID *uuid.UUID, eventType, targetType, targetID, ipAddress, userAgent string, metadata map[string]interface{}) {
	if a == nil || a.recorder == nil {
		return
	}
	a.recorder.Record(ctx, actorID, eventType, targetType, targetID, ipAddress, userAgent, metadata)
}
