package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
)

// AlertEventFilter narrows alert event list queries.
type AlertEventFilter struct {
	OwnerUserID  *uuid.UUID
	PortfolioIDs []uuid.UUID
	// FundIDs restricts portfolio-scoped alerts to those whose portfolio belongs
	// to one of the given fund IDs. Applied via an EXISTS subquery on portfolios.
	FundIDs      []uuid.UUID
	ScopeType    *entity.ScopeType
	SecurityID   *uuid.UUID
	RuleID       *uuid.UUID
	Acknowledged *bool
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	Limit        int
	Offset       int
	// UnionPersonalOwnerID switches to OR-based visibility for unscoped queries:
	// personal alerts owned by this user OR portfolio alerts (restricted by FundIDs,
	// or all portfolio alerts when UnionPortfolioAll is true).
	UnionPersonalOwnerID *uuid.UUID
	UnionPortfolioAll    bool
}

// AlertEventRepository persists alert events.
type AlertEventRepository interface {
	// Insert inserts a new alert event. Returns ErrAlertIdempotencyConflict when the
	// idempotency key already exists, which the caller treats as suppression.
	Insert(ctx context.Context, tx pgx.Tx, event *entity.AlertEvent) error
	// GetByID loads one alert event by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AlertEvent, error)
	// UpdateNotificationStatus persists the notification outcome.
	// This is always a best-effort post-commit update; no transaction is accepted.
	UpdateNotificationStatus(ctx context.Context, id uuid.UUID, status entity.NotificationStatus, notifErr *string) error
	// Acknowledge sets acknowledgement fields atomically. Returns ErrAlreadyAcknowledged
	// when already acknowledged.
	Acknowledge(ctx context.Context, tx pgx.Tx, id uuid.UUID, by uuid.UUID, note *string) error
	// List returns alert events matching the filter.
	List(ctx context.Context, filter AlertEventFilter) ([]*entity.AlertEvent, int, error)
}
