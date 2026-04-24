package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain/entity"
)

// AuditRepository defines persistence operations for audit events.
type AuditRepository interface {
	Record(ctx context.Context, event *entity.AuditEvent) error
	List(ctx context.Context, filter AuditFilter) ([]entity.AuditEvent, int, error)
}

// AuditFilter defines filtering options for listing audit events.
type AuditFilter struct {
	ActorID    *uuid.UUID
	EventType  string
	TargetType string
	TargetID   string
	Since      *time.Time
	Until      *time.Time
	Offset     int
	Limit      int
}

// Recorder is the write port other modules use to emit immutable audit events.
type Recorder interface {
	Record(
		ctx context.Context,
		actorID *uuid.UUID,
		eventType string,
		targetType string,
		targetID string,
		ipAddress string,
		userAgent string,
		metadata map[string]interface{},
	)
}

// NopRecorder safely discards audit writes when no implementation is wired.
type NopRecorder struct{}

// Record implements Recorder.
func (NopRecorder) Record(
	context.Context,
	*uuid.UUID,
	string,
	string,
	string,
	string,
	string,
	map[string]interface{},
) {
}
