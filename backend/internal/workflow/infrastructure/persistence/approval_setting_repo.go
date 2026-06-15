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

// PostgresApprovalSettingRepository implements domain.WorkflowApprovalSettingRepository.
type PostgresApprovalSettingRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresApprovalSettingRepository creates the repository.
func NewPostgresApprovalSettingRepository(pool *pgxpool.Pool) *PostgresApprovalSettingRepository {
	return &PostgresApprovalSettingRepository{pool: pool}
}

// ListByOperationType returns all active settings for the given operation type.
func (r *PostgresApprovalSettingRepository) ListByOperationType(
	ctx context.Context,
	operationType string,
) ([]*entity.WorkflowApprovalSetting, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, operation_type, approval_mode, approver_account_code,
		       approver_username, approver_role, is_active, updated_by, updated_at
		FROM workflow__approval_settings
		WHERE operation_type = $1 AND is_active = TRUE
		ORDER BY updated_at ASC`,
		operationType,
	)
	if err != nil {
		return nil, fmt.Errorf("querying approval settings: %w", err)
	}
	defer rows.Close()
	return scanApprovalSettings(rows)
}

// ListAll returns every setting ordered by operation_type, updated_at.
func (r *PostgresApprovalSettingRepository) ListAll(
	ctx context.Context,
) ([]*entity.WorkflowApprovalSetting, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, operation_type, approval_mode, approver_account_code,
		       approver_username, approver_role, is_active, updated_by, updated_at
		FROM workflow__approval_settings
		ORDER BY operation_type ASC, updated_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("querying all approval settings: %w", err)
	}
	defer rows.Close()
	return scanApprovalSettings(rows)
}

// Upsert replaces all active settings for operationType within the provided tx.
// Deactivates existing rows, then inserts the new ones.
func (r *PostgresApprovalSettingRepository) Upsert(
	ctx context.Context,
	tx pgx.Tx,
	operationType string,
	settings []*entity.WorkflowApprovalSetting,
) error {
	now := time.Now().UTC()

	_, err := tx.Exec(ctx, `
		UPDATE workflow__approval_settings
		SET is_active = FALSE, updated_at = $1
		WHERE operation_type = $2 AND is_active = TRUE`,
		now, operationType,
	)
	if err != nil {
		return fmt.Errorf("deactivating existing approval settings: %w", err)
	}

	for _, s := range settings {
		if s.ID == uuid.Nil {
			s.ID = uuid.New()
		}
		s.UpdatedAt = now
		_, err := tx.Exec(ctx, `
			INSERT INTO workflow__approval_settings (
				id, operation_type, approval_mode,
				approver_account_code, approver_username, approver_role,
				is_active, updated_by, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, TRUE, $7, $8)`,
			s.ID, s.OperationType, string(s.ApprovalMode),
			s.ApproverAccountCode, s.ApproverUsername, s.ApproverRole,
			s.UpdatedBy, s.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("inserting approval setting: %w", err)
		}
	}
	return nil
}

func scanApprovalSettings(rows pgx.Rows) ([]*entity.WorkflowApprovalSetting, error) {
	var results []*entity.WorkflowApprovalSetting
	for rows.Next() {
		var s entity.WorkflowApprovalSetting
		var modeStr string
		err := rows.Scan(
			&s.ID, &s.OperationType, &modeStr,
			&s.ApproverAccountCode, &s.ApproverUsername, &s.ApproverRole,
			&s.IsActive, &s.UpdatedBy, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning approval setting: %w", err)
		}
		s.ApprovalMode = entity.ApprovalMode(modeStr)
		results = append(results, &s)
	}
	return results, rows.Err()
}

// compile-time interface check
var _ domain.WorkflowApprovalSettingRepository = (*PostgresApprovalSettingRepository)(nil)
