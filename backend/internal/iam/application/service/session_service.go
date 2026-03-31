package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
)

// SessionService centralizes session lifecycle validation and activity tracking.
type SessionService struct {
	sessionRepo domain.SessionRepository
}

// NewSessionService creates a new session service.
func NewSessionService(sessionRepo domain.SessionRepository) *SessionService {
	return &SessionService{sessionRepo: sessionRepo}
}

// ValidateActiveSession checks whether a session belongs to the user and is still active.
func (s *SessionService) ValidateActiveSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, now time.Time) (bool, error) {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return false, fmt.Errorf("finding session: %w", err)
	}
	if session == nil {
		return false, nil
	}
	if session.UserID != userID {
		return false, nil
	}
	if !session.IsValid(now) || session.IsAbsoluteExpired(now) {
		return false, nil
	}
	return true, nil
}

// TouchSessionActivity updates the session's last activity marker.
func (s *SessionService) TouchSessionActivity(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.sessionRepo.UpdateLastActivity(ctx, sessionID); err != nil {
		return fmt.Errorf("updating session activity: %w", err)
	}
	return nil
}
