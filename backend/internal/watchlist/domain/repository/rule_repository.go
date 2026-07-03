package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
)

// RuleStateUpdate carries the fields written after each evaluation cycle.
type RuleStateUpdate struct {
	ID                 uuid.UUID
	LastState          entity.RuleState
	LastObservedPrice  *decimal.Decimal
	LastObservedAt     *time.Time
	LastEvaluatedAt    time.Time
	LastStateChangedAt *time.Time
	LastAlertedAt      *time.Time
	LastQuoteStale     bool
	LastStaleReason    *string
	UpdatedBy          *uuid.UUID
}

// EvaluatorRuleFilter filters rules returned to the evaluator.
type EvaluatorRuleFilter struct {
	ScopeType   *entity.ScopeType
	PortfolioID *uuid.UUID
	SecurityID  *uuid.UUID
	ItemID      *uuid.UUID
	RuleID      *uuid.UUID
}

// ThresholdRuleRepository persists threshold rules.
type ThresholdRuleRepository interface {
	// Create inserts a new rule. Returns ErrDuplicateThresholdRule on constraint violation.
	Create(ctx context.Context, tx pgx.Tx, rule *entity.ThresholdRule) error
	// GetByID loads one rule by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ThresholdRule, error)
	// GetByIDForUpdate loads a rule inside a transaction and holds a row-level lock
	// (SELECT … FOR UPDATE). Used by the evaluator to prevent duplicate alerts.
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.ThresholdRule, error)
	// ListByItemID returns active rules for an item, ordered by created_at.
	ListByItemID(ctx context.Context, itemID uuid.UUID) ([]*entity.ThresholdRule, error)
	// Update persists rule field changes.
	Update(ctx context.Context, tx pgx.Tx, rule *entity.ThresholdRule) error
	// UpdateState persists evaluation-cycle state changes with a lightweight update.
	UpdateState(ctx context.Context, tx pgx.Tx, upd RuleStateUpdate) error
	// SoftDisable marks the rule disabled (soft delete for threshold rules).
	// Returns ErrRuleDisabled when no matching active row is found.
	SoftDisable(ctx context.Context, tx pgx.Tx, id uuid.UUID, deletedBy uuid.UUID) error
	// ListForEvaluation returns ENABLED rules whose items are ACTIVE, with the given filter.
	// Results are ordered deterministically for batch processing.
	ListForEvaluation(ctx context.Context, filter EvaluatorRuleFilter) ([]*entity.ThresholdRule, error)
}
