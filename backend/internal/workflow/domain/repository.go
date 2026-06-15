package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
)

// WorkflowDayRepository manages the current-state snapshot table.
//
// Read methods accept a plain context and use the repository's connection pool
// (READ COMMITTED, no lock). Write methods accept a pgx.Tx so they participate
// in transactions managed by the command layer.
type WorkflowDayRepository interface {
	// GetByBusinessDate returns the global workflow day for a date.
	// Returns nil (no error) when no row exists; caller interprets nil as NOT_STARTED.
	GetByBusinessDate(ctx context.Context, businessDate time.Time) (*entity.WorkflowDay, error)

	// GetByContractDate returns the record for a contract+date.
	// Returns nil (no error) when no row exists — caller interprets nil as NOT_STARTED.
	GetByContractDate(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (*entity.WorkflowDay, error)

	// GetForUpdateByBusinessDate acquires SELECT FOR UPDATE for the global day row.
	// Returns nil (no error) when no row exists.
	GetForUpdateByBusinessDate(ctx context.Context, tx pgx.Tx, businessDate time.Time) (*entity.WorkflowDay, error)

	// GetForUpdate acquires SELECT … FOR UPDATE within an existing transaction.
	// Returns nil (no error) when no row exists.
	// Must only be called inside a database.WithTransaction closure.
	GetForUpdate(ctx context.Context, tx pgx.Tx, contractID uuid.UUID, businessDate time.Time) (*entity.WorkflowDay, error)

	// Insert creates a new row via INSERT … ON CONFLICT DO NOTHING.
	// Returns ErrWorkflowDayExists when a concurrent request already created the row.
	Insert(ctx context.Context, tx pgx.Tx, day *entity.WorkflowDay) error

	// UpdateState persists a state change with an optimistic version check.
	// Returns ErrVersionConflict when the row's version has changed since last read.
	UpdateState(ctx context.Context, tx pgx.Tx, day *entity.WorkflowDay) error

	// ResetToNotStarted reverts a DAY_OPEN row to NOT_STARTED, clearing the
	// opened_at and opened_by stamps. Must only be called inside a
	// database.WithTransaction closure.
	ResetToNotStarted(ctx context.Context, tx pgx.Tx, day *entity.WorkflowDay) error

	// ListByState returns all workflow days in a given state on a given business date.
	ListByState(ctx context.Context, state vo.WorkflowState, businessDate time.Time, offset, limit int) ([]*entity.WorkflowDay, int64, error)
}

// TransitionLogRepository is append-only: no Update or Delete methods, ever.
type TransitionLogRepository interface {
	// Append inserts one immutable transition record within a tx.
	Append(ctx context.Context, tx pgx.Tx, t *entity.WorkflowTransition) error

	// ListByBusinessDate returns all transitions for a business date in chronological order.
	ListByBusinessDate(ctx context.Context, businessDate time.Time) ([]*entity.WorkflowTransition, error)

	// ListByContractDate returns all transitions for a contract+date in chronological order.
	ListByContractDate(ctx context.Context, contractID uuid.UUID, businessDate time.Time) ([]*entity.WorkflowTransition, error)

	// ListByBusinessDatePaginated returns a page of transitions ordered by occurred_at ASC,
	// plus the total count matching the date filter.
	ListByBusinessDatePaginated(ctx context.Context, req TransitionLogPageRequest) (*TransitionLogPageResult, error)

	// ListByContractDatePaginated returns a page of transitions ordered by occurred_at ASC,
	// plus the total count matching the filter.
	ListByContractDatePaginated(ctx context.Context, req TransitionLogPageRequest) (*TransitionLogPageResult, error)
}

// ApprovalRecordRepository manages approval records for the maker-checker model.
type ApprovalRecordRepository interface {
	// Insert persists a new approval record within a tx.
	Insert(ctx context.Context, tx pgx.Tx, rec *entity.ApprovalRecord) error

	// ListByWorkflowDay returns all approval records for a given workflow day.
	ListByWorkflowDay(ctx context.Context, workflowDayID uuid.UUID) ([]*entity.ApprovalRecord, error)

	// RevokeLatestActive marks the most recent APPROVED record for a workflow day
	// as REVOKED, stamping revoked_at and revoked_by. Returns ErrNotFound when no
	// active approval exists (defensive — the policy layer should have already
	// rejected the call in that case).
	RevokeLatestActive(ctx context.Context, tx pgx.Tx, workflowDayID uuid.UUID, revokedBy uuid.UUID, revokedAt time.Time) error
}

// WorkflowApprovalSettingRepository manages per-operationType approver configuration.
type WorkflowApprovalSettingRepository interface {
	// ListByOperationType returns all active settings for the given operation type.
	// Returns an empty slice (no error) when none are configured.
	ListByOperationType(ctx context.Context, operationType string) ([]*entity.WorkflowApprovalSetting, error)

	// ListAll returns every setting (active and inactive), ordered by operation_type, updated_at.
	ListAll(ctx context.Context) ([]*entity.WorkflowApprovalSetting, error)

	// Upsert replaces all active settings for the given operationType with the supplied list.
	// Runs inside the provided transaction: deactivates previous rows, then inserts new ones.
	// Passing an empty slice deactivates all settings for that operationType.
	Upsert(ctx context.Context, tx pgx.Tx, operationType string, settings []*entity.WorkflowApprovalSetting) error
}

// TransitionLogPageRequest carries pagination params for listing transitions.
type TransitionLogPageRequest struct {
	ContractID   uuid.UUID
	BusinessDate time.Time
	Page         int // 1-based
	PageSize     int
}

// TransitionLogPageResult is the paginated response.
type TransitionLogPageResult struct {
	Transitions []*entity.WorkflowTransition
	Total       int64
	Page        int
	PageSize    int
}

// SchedulerRepository manages workflow scheduler rules and audit rows.
type SchedulerRepository interface {
	// TryAcquireSchedulerLock attempts to acquire a cross-process scheduler lock
	// using the backing store. The returned lock must be released when the tick
	// finishes. When acquired is false, another scheduler instance is running.
	TryAcquireSchedulerLock(ctx context.Context, lockKey int64) (lock SchedulerLock, acquired bool, err error)

	// ListActiveScheduleRules returns enabled rules whose effective period
	// contains businessDate. The scheduler still decides whether each rule is
	// due based on the current local time.
	ListActiveScheduleRules(ctx context.Context, businessDate time.Time) ([]*entity.ScheduleRule, error)

	// CreateSchedulerRun inserts a RUNNING audit header.
	CreateSchedulerRun(ctx context.Context, run *entity.SchedulerRun) error

	// FinishSchedulerRun updates an audit header with its terminal result.
	FinishSchedulerRun(ctx context.Context, run *entity.SchedulerRun) error

	// InsertSchedulerRunItem inserts one per-contract audit detail row.
	InsertSchedulerRunItem(ctx context.Context, item *entity.SchedulerRunItem) error
}

// SchedulerLock is a held cross-process scheduler lock.
type SchedulerLock interface {
	Release(ctx context.Context) error
}
