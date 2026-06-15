package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
)

// PostgresWorkflowDayRepository implements domain.WorkflowDayRepository.
type PostgresWorkflowDayRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresWorkflowDayRepository creates the repository.
func NewPostgresWorkflowDayRepository(pool *pgxpool.Pool) *PostgresWorkflowDayRepository {
	return &PostgresWorkflowDayRepository{pool: pool}
}

// ─── Read (pool-based, no lock) ───────────────────────────────────────────────

// GetByBusinessDate returns the global workflow day for a date, or nil if none exists.
func (r *PostgresWorkflowDayRepository) GetByBusinessDate(
	ctx context.Context,
	businessDate time.Time,
) (*entity.WorkflowDay, error) {
	row := r.pool.QueryRow(ctx, selectDaySQL+" WHERE business_date = $1 LIMIT 1",
		businessDate)
	return scanWorkflowDay(row)
}

// GetByContractDate returns the record for a contract+date, or nil if none exists.
func (r *PostgresWorkflowDayRepository) GetByContractDate(
	ctx context.Context,
	contractID uuid.UUID,
	businessDate time.Time,
) (*entity.WorkflowDay, error) {
	row := r.pool.QueryRow(ctx, selectDaySQL+" WHERE contract_id = $1 AND business_date = $2",
		contractID, businessDate)
	return scanWorkflowDay(row)
}

// ListByState returns paginated workflow days in a given state on a date.
func (r *PostgresWorkflowDayRepository) ListByState(
	ctx context.Context,
	state vo.WorkflowState,
	businessDate time.Time,
	offset, limit int,
) ([]*entity.WorkflowDay, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int64
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM workflow__day_states WHERE current_state = $1 AND business_date = $2",
		string(state), businessDate,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting workflow days by state: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		selectDaySQL+" WHERE current_state = $1 AND business_date = $2 ORDER BY contract_id LIMIT $3 OFFSET $4",
		string(state), businessDate, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing workflow days by state: %w", err)
	}
	defer rows.Close()

	var results []*entity.WorkflowDay
	for rows.Next() {
		d, err := scanWorkflowDay(rows)
		if err != nil {
			return nil, 0, err
		}
		if d != nil {
			results = append(results, d)
		}
	}
	return results, total, rows.Err()
}

// ─── Write (tx-based) ─────────────────────────────────────────────────────────

// GetForUpdate acquires SELECT … FOR UPDATE within a transaction.
// GetForUpdateByBusinessDate acquires SELECT FOR UPDATE for the global day row.
func (r *PostgresWorkflowDayRepository) GetForUpdateByBusinessDate(
	ctx context.Context,
	tx pgx.Tx,
	businessDate time.Time,
) (*entity.WorkflowDay, error) {
	row := tx.QueryRow(ctx,
		selectDaySQL+" WHERE business_date = $1 FOR UPDATE",
		businessDate,
	)
	return scanWorkflowDay(row)
}

func (r *PostgresWorkflowDayRepository) GetForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	contractID uuid.UUID,
	businessDate time.Time,
) (*entity.WorkflowDay, error) {
	row := tx.QueryRow(ctx,
		selectDaySQL+" WHERE contract_id = $1 AND business_date = $2 FOR UPDATE",
		contractID, businessDate,
	)
	return scanWorkflowDay(row)
}

// Insert creates a new workflow day record.
// Uses ON CONFLICT DO NOTHING; returns ErrWorkflowDayExists when a row already exists.
func (r *PostgresWorkflowDayRepository) Insert(
	ctx context.Context,
	tx pgx.Tx,
	d *entity.WorkflowDay,
) error {
	tag, err := tx.Exec(ctx, `
		INSERT INTO workflow__day_states (
			id, contract_id, business_date, current_state,
			opened_at, opened_by,
			transactions_locked_at, pending_reclose, reclose_count,
			version, created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4,
			$5, $6,
			$7, $8, $9,
			$10, $11, $12, $13, $14
		)
		ON CONFLICT (business_date) DO NOTHING`,
		d.ID, d.ContractID, d.BusinessDate, string(d.CurrentState),
		d.OpenedAt, d.OpenedBy,
		d.TransactionsLockedAt, d.PendingReclose, d.RecloseCount,
		d.Version, d.CreatedAt, d.UpdatedAt, d.CreatedBy, d.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting workflow day: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Concurrent OPEN_DAY won the race; read back to populate the error.
		existing, readErr := r.GetByBusinessDate(ctx, d.BusinessDate)
		state := "UNKNOWN"
		if readErr == nil && existing != nil {
			state = string(existing.CurrentState)
		}
		return &domain.ErrWorkflowDayExists{
			ContractID:   d.ContractID.String(),
			BusinessDate: d.BusinessDate.Format("2006-01-02"),
			CurrentState: state,
		}
	}
	return nil
}

// UpdateState applies a state change using an optimistic version check.
// Returns ErrVersionConflict when the row was concurrently modified.
func (r *PostgresWorkflowDayRepository) UpdateState(
	ctx context.Context,
	tx pgx.Tx,
	d *entity.WorkflowDay,
) error {
	tag, err := tx.Exec(ctx, `
		UPDATE workflow__day_states SET
			current_state           = $3,
			manager_approved_at     = $4,
			manager_approved_by     = $5,
			transactions_locked_at  = $6,
			transaction_closed_at   = $7,
			transaction_closed_by   = $8,
			accounting_closed_at    = $9,
			accounting_closed_by    = $10,
			pending_reclose         = $11,
			reclose_count           = $12,
			accounting_date         = $15,
			prev_accounting_date    = $16,
			version                 = version + 1,
			updated_at              = $13,
			updated_by              = $14
		WHERE id = $1 AND version = $2`,
		d.ID, d.Version,
		string(d.CurrentState),
		d.ManagerApprovedAt, d.ManagerApprovedBy,
		d.TransactionsLockedAt,
		d.TransactionClosedAt, d.TransactionClosedBy,
		d.AccountingClosedAt, d.AccountingClosedBy,
		d.PendingReclose, d.RecloseCount,
		d.UpdatedAt, d.UpdatedBy,
		d.AccountingDate, d.PrevAccountingDate,
	)
	if err != nil {
		return fmt.Errorf("updating workflow day state: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ErrVersionConflict{}
	}
	return nil
}

// ResetToNotStarted reverts a DAY_OPEN row to NOT_STARTED, clearing opened_at
// and opened_by. Uses the same optimistic version check as UpdateState.
func (r *PostgresWorkflowDayRepository) ResetToNotStarted(
	ctx context.Context,
	tx pgx.Tx,
	d *entity.WorkflowDay,
) error {
	tag, err := tx.Exec(ctx, `
		UPDATE workflow__day_states SET
			current_state = 'NOT_STARTED',
			opened_at     = NULL,
			opened_by     = NULL,
			version       = version + 1,
			updated_at    = $3,
			updated_by    = $4
		WHERE id = $1 AND version = $2`,
		d.ID, d.Version,
		d.UpdatedAt, d.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("resetting workflow day to not-started: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ErrVersionConflict{}
	}
	return nil
}

// ─── Scanning ─────────────────────────────────────────────────────────────────

const selectDaySQL = `
	SELECT
		id, contract_id, business_date, current_state,
		opened_at, opened_by,
		manager_approved_at, manager_approved_by,
		transactions_locked_at,
		transaction_closed_at, transaction_closed_by,
		accounting_closed_at, accounting_closed_by,
		pending_reclose, reclose_count,
		accounting_date, prev_accounting_date,
		version, created_at, updated_at, created_by, updated_by
	FROM workflow__day_states`

// scanner is satisfied by both pgx.Row and pgx.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanWorkflowDay(s scanner) (*entity.WorkflowDay, error) {
	var d entity.WorkflowDay
	var stateStr string

	err := s.Scan(
		&d.ID, &d.ContractID, &d.BusinessDate, &stateStr,
		&d.OpenedAt, &d.OpenedBy,
		&d.ManagerApprovedAt, &d.ManagerApprovedBy,
		&d.TransactionsLockedAt,
		&d.TransactionClosedAt, &d.TransactionClosedBy,
		&d.AccountingClosedAt, &d.AccountingClosedBy,
		&d.PendingReclose, &d.RecloseCount,
		&d.AccountingDate, &d.PrevAccountingDate,
		&d.Version, &d.CreatedAt, &d.UpdatedAt, &d.CreatedBy, &d.UpdatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning workflow day: %w", err)
	}

	d.CurrentState = vo.WorkflowState(stateStr)
	d.CreatedAt = d.CreatedAt.UTC()
	d.UpdatedAt = d.UpdatedAt.UTC()
	if d.OpenedAt != nil {
		t := d.OpenedAt.UTC()
		d.OpenedAt = &t
	}
	if d.ManagerApprovedAt != nil {
		t := d.ManagerApprovedAt.UTC()
		d.ManagerApprovedAt = &t
	}
	if d.TransactionsLockedAt != nil {
		t := d.TransactionsLockedAt.UTC()
		d.TransactionsLockedAt = &t
	}
	return &d, nil
}

// compile-time interface check
var _ domain.WorkflowDayRepository = (*PostgresWorkflowDayRepository)(nil)
