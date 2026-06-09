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
//
// Record is the legacy fire-and-forget path: implementations log errors
// internally but never propagate them to the caller. Use it for low-risk
// CRUD events that should not be allowed to fail the operation.
//
// RecordStrict is the synchronous, error-returning path required for
// financial-grade actions. Callers must check the returned error and
// surface it; the operation may be retried or alerted on failure.
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
	RecordStrict(
		ctx context.Context,
		actorID *uuid.UUID,
		eventType string,
		targetType string,
		targetID string,
		ipAddress string,
		userAgent string,
		metadata map[string]interface{},
	) error
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

// RecordStrict implements Recorder. Returns nil so an unwired audit recorder
// never blocks the operation in tests, while production wiring should always
// supply a real Recorder so failures are surfaced.
func (NopRecorder) RecordStrict(
	context.Context,
	*uuid.UUID,
	string,
	string,
	string,
	string,
	string,
	map[string]interface{},
) error {
	return nil
}
