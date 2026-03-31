package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/dto"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
)

// ListSessionsQuery retrieves active sessions for a user.
type ListSessionsQuery struct {
	sessionRepo domain.SessionRepository
}

// NewListSessionsQuery creates a new ListSessionsQuery.
func NewListSessionsQuery(sessionRepo domain.SessionRepository) *ListSessionsQuery {
	return &ListSessionsQuery{sessionRepo: sessionRepo}
}

// Execute returns all active sessions for the given user.
func (q *ListSessionsQuery) Execute(ctx context.Context, userID uuid.UUID) ([]dto.SessionInfo, error) {
	sessions, err := q.sessionRepo.ListActiveForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}

	result := make([]dto.SessionInfo, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, dto.SessionInfo{
			ID:             s.ID.String(),
			IPAddress:      s.IPAddress,
			UserAgent:      s.UserAgent,
			LastActivityAt: s.LastActivityAt,
			CreatedAt:      s.CreatedAt,
			ExpiresAt:      s.ExpiresAt,
		})
	}
	return result, nil
}
