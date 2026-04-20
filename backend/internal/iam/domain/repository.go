package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

// UserRepository defines persistence operations for User entities.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	List(ctx context.Context, filter UserFilter) ([]entity.User, int, error)
}

// UserFilter defines filtering options for listing users.
type UserFilter struct {
	IsActive *bool
	IsLocked *bool
	Search   string
	Offset   int
	Limit    int
}

// SessionRepository defines persistence operations for refresh token sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Session, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	RevokeByIDWithReason(ctx context.Context, id uuid.UUID, reason string) error
	RevokeByFamily(ctx context.Context, family uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	RevokeAllForUserWithReason(ctx context.Context, userID uuid.UUID, reason string) error
	DeleteExpired(ctx context.Context) (int64, error)
	// Session listing and management
	ListActiveForUser(ctx context.Context, userID uuid.UUID) ([]entity.Session, error)
	CountActiveForUser(ctx context.Context, userID uuid.UUID) (int, error)
	RevokeOldestForUser(ctx context.Context, userID uuid.UUID, reason string) error
	// Idle tracking
	UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error
}

// MFARepository defines persistence operations for MFA enrollments and recovery codes.
type MFARepository interface {
	// Enrollment
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.MFAEnrollment, error)
	CreateEnrollment(ctx context.Context, enrollment *entity.MFAEnrollment) error
	EnableEnrollment(ctx context.Context, userID uuid.UUID) error
	DisableEnrollment(ctx context.Context, userID uuid.UUID) error
	DeleteEnrollment(ctx context.Context, userID uuid.UUID) error
	// Recovery codes
	StoreRecoveryCodes(ctx context.Context, codes []entity.MFARecoveryCode) error
	FindUnusedRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]entity.MFARecoveryCode, error)
	UseRecoveryCode(ctx context.Context, codeID uuid.UUID) error
	DeleteRecoveryCodes(ctx context.Context, userID uuid.UUID) error
}

// PermissionsFetcher provides user permissions from the permissions module.
type PermissionsFetcher interface {
	GetUserFunctionPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetUserDataPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
}
