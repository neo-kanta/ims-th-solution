package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
)

// WatchlistItemFilter narrows item list queries.
type WatchlistItemFilter struct {
	OwnerUserID     *uuid.UUID
	PortfolioIDs    []uuid.UUID
	// FundIDs restricts portfolio-scoped items to those whose portfolio belongs
	// to one of the given fund IDs. Applied via an EXISTS subquery on portfolios.
	FundIDs         []uuid.UUID
	ScopeType       *entity.ScopeType
	SecurityID      *uuid.UUID
	IncludeDisabled bool
	Limit           int
	Offset          int
	// UnionPersonalOwnerID switches to OR-based visibility for unscoped queries:
	// personal items owned by this user OR portfolio items (restricted by FundIDs,
	// or all portfolio items when UnionPortfolioAll is true).
	UnionPersonalOwnerID *uuid.UUID
	UnionPortfolioAll    bool
}

// WatchlistItemRepository persists watchlist items.
type WatchlistItemRepository interface {
	// Create inserts a new item. Returns ErrDuplicateWatchlistItem on constraint violation.
	Create(ctx context.Context, tx pgx.Tx, item *entity.WatchlistItem) error
	// GetByID loads one item by ID. Returns nil, nil when not found or deleted.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.WatchlistItem, error)
	// Update persists item field changes.
	Update(ctx context.Context, tx pgx.Tx, item *entity.WatchlistItem) error
	// SoftDelete marks the item as deleted.
	SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, deletedBy uuid.UUID) error
	// List returns items visible to the caller after scope filtering.
	List(ctx context.Context, filter WatchlistItemFilter) ([]*entity.WatchlistItem, int, error)
}
