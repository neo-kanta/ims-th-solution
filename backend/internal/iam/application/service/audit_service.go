package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

// AuditService centralizes IAM audit event creation and persistence.
type AuditService struct {
	repo domain.AuditRepository
}

// NewAuditService creates a new audit service.
func NewAuditService(repo domain.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Record persists a standardized IAM audit event.
func (s *AuditService) Record(
	ctx context.Context,
	actorID *uuid.UUID,
	eventType string,
	targetType string,
	targetID string,
	ipAddress string,
	userAgent string,
	metadata map[string]interface{},
) {
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
	if err := s.repo.Record(ctx, event); err != nil {
		slog.Error("failed to record IAM audit event", "error", err, "event_type", eventType, "target_type", targetType, "target_id", targetID)
	}
}
