package domain

import (
	"context"

	"github.com/google/uuid"

	"github.com/your-org/ims-th-solution/backend/internal/iam/domain/entity"
)

// UserRepository defines the persistence interface for User entities.
// Implementations live in infrastructure/persistence/.
type UserRepository interface {
	// FindByUsername retrieves a user by their login username.
	// Returns nil if no user is found.
	FindByUsername(ctx context.Context, username string) (*entity.User, error)

	// FindByID retrieves a user by their UUID.
	// Returns nil if no user is found.
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}
