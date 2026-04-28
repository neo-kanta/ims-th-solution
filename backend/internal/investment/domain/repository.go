package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// DecisionRepository persists investment Decision aggregates.
// Implementations live in infrastructure/persistence/.
type DecisionRepository interface {
	// GetByID returns the decision or (*ErrDecisionNotFound) when missing.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Decision, error)

	// UpdateStatus atomically transitions the decision to the given status.
	// Implementations should use optimistic concurrency / compare-and-set so
	// that two concurrent submits cannot both land.
	UpdateStatus(
		ctx context.Context,
		id uuid.UUID,
		to vo.DecisionStatus,
		checkGroupID uuid.UUID,
		updatedBy uuid.UUID,
		updatedAt time.Time,
	) error
}

// InvestmentProcessGuardRepository provides read models required to decide
// whether a user may execute one investment process step for a contract/date.
type InvestmentProcessGuardRepository interface {
	// GetActiveDaySetting resolves the effective workflow day setting for the
	// contract/date. Contract-specific settings should outrank GLOBAL settings.
	GetActiveDaySetting(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (*entity.WorkflowDaySetting, error)

	// FindBlockingControlDecision returns an active rejection that blocks either
	// the whole day or the requested process step. Nil means no active block.
	FindBlockingControlDecision(
		ctx context.Context,
		contractID uuid.UUID,
		businessDate time.Time,
		processStep vo.ProcessStepKey,
	) (*entity.BlockingControlDecision, error)

	// FindUserProcessAssignment returns the best active assignment authorizing
	// userID for the process step. Nil means the user is not assigned.
	FindUserProcessAssignment(
		ctx context.Context,
		userID uuid.UUID,
		contractID uuid.UUID,
		businessDate time.Time,
		processStep vo.ProcessStepKey,
	) (*entity.ProcessAssignmentMatch, error)
}
