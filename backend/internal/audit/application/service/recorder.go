package service

import (
	"context"
	"log/slog"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain/entity"
)

// Recorder persists immutable audit events through the audit repository.
type Recorder struct {
	repo domain.AuditRepository
}

// NewRecorder creates an audit recorder backed by the given repository.
func NewRecorder(repo domain.AuditRepository) *Recorder {
	return &Recorder{repo: repo}
}

// Record writes a standardized audit event and logs failures without interrupting callers.
// Use for low-risk CRUD that should not be blocked by an audit-store outage.
func (s *Recorder) Record(
	ctx context.Context,
	actorID *uuid.UUID,
	eventType string,
	targetType string,
	targetID string,
	ipAddress string,
	userAgent string,
	metadata map[string]interface{},
) {
	if err := s.RecordStrict(ctx, actorID, eventType, targetType, targetID, ipAddress, userAgent, metadata); err != nil {
		slog.Error("failed to record audit event", "error", err, "event_type", eventType, "target_type", targetType, "target_id", targetID)
	}
}

// RecordStrict persists the audit event synchronously and returns the
// underlying repository error. Use for financial-grade actions where audit
// loss is unacceptable.
func (s *Recorder) RecordStrict(
	ctx context.Context,
	actorID *uuid.UUID,
	eventType string,
	targetType string,
	targetID string,
	ipAddress string,
	userAgent string,
	metadata map[string]interface{},
) error {
	metadata = enrichMetadata(ctx, metadata)

	event := &entity.AuditEvent{
		ID:         uuid.New(),
		ActorID:    actorID,
		EventType:  eventType,
		TargetType: targetType,
		TargetID:   targetID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Metadata:   metadata,
	}
	return s.repo.Record(ctx, event)
}

func enrichMetadata(ctx context.Context, metadata map[string]interface{}) map[string]interface{} {
	requestID := chimw.GetReqID(ctx)
	if metadata == nil && requestID == "" {
		return nil
	}

	enriched := cloneMetadata(metadata)
	if enriched == nil {
		enriched = map[string]interface{}{}
	}
	if requestID != "" {
		if _, exists := enriched["request_id"]; !exists {
			enriched["request_id"] = requestID
		}
	}
	return enriched
}

func cloneMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return nil
	}

	cloned := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		cloned[k] = v
	}
	return cloned
}
