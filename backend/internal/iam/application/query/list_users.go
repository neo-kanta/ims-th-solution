package query

import (
	"context"
	"fmt"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

// ListUsersQuery retrieves paginated users for admin management.
type ListUsersQuery struct {
	userRepo domain.UserRepository
}

// NewListUsersQuery creates a new ListUsersQuery.
func NewListUsersQuery(userRepo domain.UserRepository) *ListUsersQuery {
	return &ListUsersQuery{userRepo: userRepo}
}

// Execute returns paginated users matching the filter.
func (q *ListUsersQuery) Execute(ctx context.Context, filter domain.UserFilter) ([]entity.User, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}

	users, total, err := q.userRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}
	return users, total, nil
}
