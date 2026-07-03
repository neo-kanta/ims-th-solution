package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/entity"
)

// SessionRepository persists Session aggregates.
type SessionRepository interface {
	Create(ctx context.Context, s *entity.Session) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Session, error)
	Touch(ctx context.Context, id uuid.UUID) error
	UpdateModel(ctx context.Context, id uuid.UUID, provider, model string) error
	// ListByUser returns one page of a user's sessions (most-recent first) plus
	// the total count for pagination.
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Session, int, error)
}

// MessageRepository persists individual messages in append-only fashion.
type MessageRepository interface {
	Append(ctx context.Context, m *entity.Message) error
	ListBySession(ctx context.Context, sessionID uuid.UUID, limit int) ([]entity.Message, error)
	// ListBySessionPaged returns one page of a session's messages
	// (chronological) plus the total count for pagination.
	ListBySessionPaged(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]entity.Message, int, error)
}

// ToolInvocationRepository persists tool-call provenance records (append-only).
type ToolInvocationRepository interface {
	Append(ctx context.Context, inv *entity.ToolInvocation) error
}
