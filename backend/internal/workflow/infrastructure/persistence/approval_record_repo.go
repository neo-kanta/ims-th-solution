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
)

// PostgresApprovalRecordRepository implements domain.ApprovalRecordRepository.
type PostgresApprovalRecordRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresApprovalRecordRepository creates the repository.
func NewPostgresApprovalRecordRepository(pool *pgxpool.Pool) *PostgresApprovalRecordRepository {
	return &PostgresApprovalRecordRepository{pool: pool}
}

// Insert persists a new approval record within an existing transaction.
func (r *PostgresApprovalRecordRepository) Insert(
	ctx context.Context,
	tx pgx.Tx,
	rec *entity.ApprovalRecord,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO workflow__approval_records (
			id, workflow_day_id, contract_id, business_date,
			approver_id, approver_username, approver_role,
			approval_status,
			is_zero_transaction, attestation_reason,
			approved_at, notes
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8,
			$9, $10,
			$11, $12
		)`,
		rec.ID, rec.WorkflowDayID, rec.ContractID, rec.BusinessDate,
		rec.ApproverID, rec.ApproverUsername, rec.ApproverRole,
		string(rec.ApprovalStatus),
		rec.IsZeroTransaction, rec.AttestationReason,
		rec.ApprovedAt, rec.Notes,
	)
	if err != nil {
		return fmt.Errorf("inserting approval record: %w", err)
	}
	return nil
}

// ListByWorkflowDay returns all approval records for a given workflow day.
func (r *PostgresApprovalRecordRepository) ListByWorkflowDay(
	ctx context.Context,
	workflowDayID uuid.UUID,
) ([]*entity.ApprovalRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id, workflow_day_id, contract_id, business_date,
			approver_id, approver_username, approver_role,
			approval_status,
			is_zero_transaction, attestation_reason,
			approved_at, revoked_at, revoked_by, notes
		FROM workflow__approval_records
		WHERE workflow_day_id = $1
		ORDER BY approved_at ASC`,
		workflowDayID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying approval records: %w", err)
	}
	defer rows.Close()

	var results []*entity.ApprovalRecord
	for rows.Next() {
		rec, err := scanApprovalRecord(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	return results, rows.Err()
}

func scanApprovalRecord(rows pgx.Rows) (*entity.ApprovalRecord, error) {
	var rec entity.ApprovalRecord
	var statusStr string

	err := rows.Scan(
		&rec.ID, &rec.WorkflowDayID, &rec.ContractID, &rec.BusinessDate,
		&rec.ApproverID, &rec.ApproverUsername, &rec.ApproverRole,
		&statusStr,
		&rec.IsZeroTransaction, &rec.AttestationReason,
		&rec.ApprovedAt, &rec.RevokedAt, &rec.RevokedBy, &rec.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning approval record: %w", err)
	}

	rec.ApprovalStatus = entity.ApprovalStatus(statusStr)
	rec.ApprovedAt = rec.ApprovedAt.UTC()
	if rec.RevokedAt != nil {
		t := rec.RevokedAt.UTC()
		rec.RevokedAt = &t
	}
	return &rec, nil
}

// RevokeLatestActive marks the most recent APPROVED record for a workflow day as REVOKED.
//
// Implementation note: a single-row UPDATE filtered to APPROVED status is used
// rather than a sub-select, so the FOR UPDATE-locked workflow_day row already
// serialises concurrent writers — there can only be one active approval at any
// moment under the current single-approver policy.
func (r *PostgresApprovalRecordRepository) RevokeLatestActive(
	ctx context.Context,
	tx pgx.Tx,
	workflowDayID uuid.UUID,
	revokedBy uuid.UUID,
	revokedAt time.Time,
) error {
	tag, err := tx.Exec(ctx, `
		UPDATE workflow__approval_records
		SET approval_status = 'REVOKED',
		    revoked_at      = $2,
		    revoked_by      = $3
		WHERE id = (
		    SELECT id FROM workflow__approval_records
		    WHERE workflow_day_id = $1 AND approval_status = 'APPROVED'
		    ORDER BY approved_at DESC
		    LIMIT 1
		)`,
		workflowDayID, revokedAt, revokedBy,
	)
	if err != nil {
		return fmt.Errorf("revoking latest approval: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no active approval found for workflow day %s", workflowDayID)
	}
	return nil
}

// compile-time interface check
var _ domain.ApprovalRecordRepository = (*PostgresApprovalRecordRepository)(nil)
