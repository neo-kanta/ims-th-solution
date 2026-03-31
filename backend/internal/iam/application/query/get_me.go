package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/dto"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	apperrors "github.com/neo-kanta/ims-th-solution/backend/platform/errors"
)

// GetMeQuery retrieves the current user's profile by their ID.
type GetMeQuery struct {
	userRepo     domain.UserRepository
	permsFetcher domain.PermissionsFetcher
}

// NewGetMeQuery creates a GetMeQuery with its dependencies.
func NewGetMeQuery(userRepo domain.UserRepository, permsFetcher domain.PermissionsFetcher) *GetMeQuery {
	return &GetMeQuery{
		userRepo:     userRepo,
		permsFetcher: permsFetcher,
	}
}

// Execute returns the user profile and permissions for the given user ID.
func (q *GetMeQuery) Execute(ctx context.Context, userID string) (*dto.LoginResult, error) {
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

	functions, err := q.permsFetcher.GetUserFunctionPermissions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching function permissions: %w", err)
	}

	contracts, err := q.permsFetcher.GetUserDataPermissions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching data permissions: %w", err)
	}

	return &dto.LoginResult{
		User: dto.UserProfile{
			ID:          user.ID.String(),
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			Groups:      user.Groups,
		},
		Permissions: dto.UserPermissions{
			Functions: functions,
			Contracts: contracts,
		},
	}, nil
}
