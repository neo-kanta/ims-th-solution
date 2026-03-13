package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/your-org/ims-th-solution/backend/internal/iam/application/dto"
	"github.com/your-org/ims-th-solution/backend/internal/iam/domain"
	apperrors "github.com/your-org/ims-th-solution/backend/platform/errors"
)

// GetMeQuery retrieves the current user's profile by their ID.
type GetMeQuery struct {
	userRepo domain.UserRepository
}

// NewGetMeQuery creates a GetMeQuery with its dependencies.
func NewGetMeQuery(userRepo domain.UserRepository) *GetMeQuery {
	return &GetMeQuery{userRepo: userRepo}
}

// Execute returns the user profile for the given user ID (from JWT claims).
func (q *GetMeQuery) Execute(ctx context.Context, userID string) (*dto.UserProfile, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.NewBusinessError(apperrors.CodeInvalidInput, "invalid user ID")
	}

	user, err := q.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return nil, apperrors.NewNotFoundError("user", userID)
	}

	return &dto.UserProfile{
		ID:          user.ID.String(),
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Groups:      user.Groups,
		IsOnLeave:   user.IsOnLeave,
	}, nil
}
