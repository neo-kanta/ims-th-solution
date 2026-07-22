package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// RuleInstanceRepository manages rule instance persistence.
type RuleInstanceRepository interface {
	Create(ctx context.Context, instance *entity.RuleInstance) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.RuleInstance, error)
	List(ctx context.Context, filter RuleInstanceFilter) ([]entity.RuleInstance, int64, error)
	UpdateActive(ctx context.Context, id uuid.UUID, isActive bool) error

	// Version management — append-only
	CreateVersion(ctx context.Context, version *entity.RuleInstanceVersion) error
	GetCurrentVersion(ctx context.Context, instanceID uuid.UUID) (*entity.RuleInstanceVersion, error)
	// GetCurrentVersions batch-loads the current version for many instances in
	// one query. Instances without a current version are absent from the map.
	GetCurrentVersions(ctx context.Context, instanceIDs []uuid.UUID) (map[uuid.UUID]*entity.RuleInstanceVersion, error)
	ListVersions(ctx context.Context, instanceID uuid.UUID) ([]entity.RuleInstanceVersion, error)
}

// RuleInstanceFilter defines filtering options for listing rule instances.
type RuleInstanceFilter struct {
	RuleTypeID *string
	IsActive   *bool
	Offset     int
	Limit      int
}

// RuleBindingRepository manages rule binding persistence.
type RuleBindingRepository interface {
	Create(ctx context.Context, binding *entity.RuleBinding) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.RuleBinding, error)
	List(ctx context.Context, filter BindingFilter) ([]entity.RuleBinding, int64, error)
	Deactivate(ctx context.Context, id uuid.UUID) error

	// Hot-path query: resolve all active bindings applicable to a set of scopes on a date.
	// Used by the evaluation pipeline on every check — must be fast.
	ResolveApplicable(ctx context.Context, scopes []vo.Scope, date time.Time) ([]ResolvedBinding, error)
}

// ResolvedBinding is the result of resolving applicable bindings.
// It includes the binding, its associated rule instance, and the current parameter version.
type ResolvedBinding struct {
	Binding        entity.RuleBinding
	RuleInstance   entity.RuleInstance
	CurrentVersion entity.RuleInstanceVersion
}

// BindingFilter defines filtering options for listing bindings.
type BindingFilter struct {
	ScopeType      *vo.ScopeType
	ScopeID        *uuid.UUID
	RuleInstanceID *uuid.UUID
	IsActive       *bool
	Offset         int
	Limit          int
}

// RuleSetRepository manages rule set persistence.
type RuleSetRepository interface {
	Create(ctx context.Context, set *entity.RuleSet) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.RuleSet, error)
	List(ctx context.Context, offset, limit int) ([]entity.RuleSet, int64, error)
}

// CheckRecordRepository is append-only. No Update or Delete methods.
type CheckRecordRepository interface {
	Create(ctx context.Context, record *entity.CheckRecord) error
	CreateBatch(ctx context.Context, records []entity.CheckRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.CheckRecord, error)
	GetByGroupID(ctx context.Context, groupID uuid.UUID) ([]entity.CheckRecord, error)
	List(ctx context.Context, filter CheckRecordFilter) ([]entity.CheckRecord, int64, error)
}

// CheckRecordFilter defines filtering options for check record queries.
type CheckRecordFilter struct {
	OrderID     *uuid.UUID
	PortfolioID *uuid.UUID
	ContractID  *uuid.UUID
	Ticker      string
	Timing      *vo.CheckTiming
	Verdict     *vo.Verdict
	DateFrom    *time.Time
	DateTo      *time.Time
	Offset      int
	Limit       int
}

// BreachRepository manages breach persistence.
type BreachRepository interface {
	Create(ctx context.Context, breach *entity.Breach) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Breach, error)
	List(ctx context.Context, filter BreachFilter) ([]entity.Breach, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.BreachStatus,
		resolvedBy *uuid.UUID, resolvedAt *time.Time) error
}

// BreachFilter defines filtering options for breach queries.
type BreachFilter struct {
	PortfolioID *uuid.UUID
	ContractID  *uuid.UUID
	Status      *entity.BreachStatus
	RuleTypeID  string
	DateFrom    *time.Time
	DateTo      *time.Time
	Offset      int
	Limit       int
}

// OverrideRepository manages override persistence.
type OverrideRepository interface {
	// CommitOverride atomically commits a breach override in a single transaction:
	//   1. SELECT ... FOR UPDATE locks the breach row.
	//   2. Verifies the breach exists and is in OPEN status.
	//   3. Inserts the override record.
	//   4. Updates the breach status to OVERRIDDEN with resolved_by = override.OverriddenBy
	//      and resolved_at = override.CreatedAt (shared UTC timestamp).
	//   5. Commits only if every step succeeds.
	//
	// Returns a typed domain error when the invariant is violated:
	//   - *ErrBreachNotFound        — no breach with that ID
	//   - *ErrBreachNotOpen         — breach status is not OPEN
	//   - *ErrOverrideAlreadyExists — a concurrent request won the race
	//                                 (UNIQUE(breach_id) constraint violation, SQLSTATE 23505)
	CommitOverride(ctx context.Context, override *entity.Override) error

	// GetByBreachID returns the override for a specific breach (at most one).
	GetByBreachID(ctx context.Context, breachID uuid.UUID) (*entity.Override, error)

	// List returns overrides for a slice of breach IDs (used for bulk display).
	List(ctx context.Context, breachIDs []uuid.UUID) ([]entity.Override, error)
}
