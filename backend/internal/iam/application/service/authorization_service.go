package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
)

// AuthorizationService provides server-side function and data-scope evaluation.
type AuthorizationService struct {
	permsFetcher domain.PermissionsFetcher
}

// NewAuthorizationService creates a new authorization service.
func NewAuthorizationService(permsFetcher domain.PermissionsFetcher) *AuthorizationService {
	return &AuthorizationService{permsFetcher: permsFetcher}
}

// HasFunctionPermission evaluates whether a user has an effective function permission.
func (s *AuthorizationService) HasFunctionPermission(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	perms, err := s.permsFetcher.GetUserFunctionPermissions(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("fetching function permissions: %w", err)
	}
	for _, perm := range perms {
		if perm == code {
			return true, nil
		}
	}
	return false, nil
}

// HasDataPermission evaluates whether a user has access to a specific contract/fund scope.
func (s *AuthorizationService) HasDataPermission(ctx context.Context, userID uuid.UUID, scopeID string) (bool, error) {
	scopes, err := s.permsFetcher.GetUserDataPermissions(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("fetching data permissions: %w", err)
	}
	for _, scope := range scopes {
		if scope == scopeID {
			return true, nil
		}
	}
	return false, nil
}

// GetAccessibleScopes returns every effective data scope for a user.
func (s *AuthorizationService) GetAccessibleScopes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	scopes, err := s.permsFetcher.GetUserDataPermissions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching data permissions: %w", err)
	}
	return scopes, nil
}
