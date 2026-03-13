package command

import (
	"context"
	"fmt"

	"github.com/your-org/ims-th-solution/backend/internal/iam/application"
	"github.com/your-org/ims-th-solution/backend/internal/iam/application/dto"
	"github.com/your-org/ims-th-solution/backend/internal/iam/domain"
	"github.com/your-org/ims-th-solution/backend/internal/iam/domain/valueobject"
	apperrors "github.com/your-org/ims-th-solution/backend/platform/errors"
)

// LoginCommand handles user authentication.
type LoginCommand struct {
	userRepo     domain.UserRepository
	tokenService *application.TokenService
	permsFetcher PermissionsFetcher
}

// PermissionsFetcher provides user permissions for the login response.
// This is a local interface satisfied by the permissions infrastructure adapter.
type PermissionsFetcher interface {
	GetUserFunctionPermissions(ctx context.Context, userID string) ([]string, error)
	GetUserDataPermissions(ctx context.Context, userID string) ([]string, error)
}

// NewLoginCommand creates a LoginCommand with its dependencies.
func NewLoginCommand(
	userRepo domain.UserRepository,
	tokenService *application.TokenService,
	permsFetcher PermissionsFetcher,
) *LoginCommand {
	return &LoginCommand{
		userRepo:     userRepo,
		tokenService: tokenService,
		permsFetcher: permsFetcher,
	}
}

// LoginInput is the input for the login use case.
type LoginInput struct {
	Username string
	Password string
}

// Execute authenticates the user and returns a JWT token with profile and permissions.
func (c *LoginCommand) Execute(ctx context.Context, input LoginInput) (*dto.LoginResult, error) {
	// 1. Find user by username
	user, err := c.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "invalid username or password")
	}

	// 2. Verify password
	if err := valueobject.CheckPassword(user.PasswordHash, input.Password); err != nil {
		return nil, apperrors.NewBusinessError(apperrors.CodeUnauthorized, "invalid username or password")
	}

	// 3. Check if login is allowed (active, not on leave)
	if allowed, reason := user.IsLoginAllowed(); !allowed {
		return nil, apperrors.NewBusinessError(apperrors.CodeForbidden, reason)
	}

	// 4. Generate JWT token
	token, expiresAt, err := c.tokenService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}

	// 5. Fetch permissions
	functions, err := c.permsFetcher.GetUserFunctionPermissions(ctx, user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("fetching function permissions: %w", err)
	}

	contracts, err := c.permsFetcher.GetUserDataPermissions(ctx, user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("fetching data permissions: %w", err)
	}

	// 6. Build result
	return &dto.LoginResult{
		Token:     token,
		ExpiresAt: expiresAt,
		User: dto.UserProfile{
			ID:          user.ID.String(),
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			Groups:      user.Groups,
			IsOnLeave:   user.IsOnLeave,
		},
		Permissions: dto.UserPermissions{
			Functions: functions,
			Contracts: contracts,
		},
	}, nil
}
