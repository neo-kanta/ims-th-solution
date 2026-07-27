package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

const syncFailureSelect = `
	SELECT id, approval_request_id, subject_type, subject_id, outcome, COALESCE(reason, ''),
	       attempt_count, max_attempts, last_error, status,
	       created_at, updated_at, resolved_at, resolved_by
	FROM approval__sync_failures`

// CreateSyncFailure inserts a new PENDING (or, if the caller pre-set it,
// whatever status/attempt_count it was given) sync-failure record.
func (r *PostgresRepository) CreateSyncFailure(ctx context.Context, tx pgx.Tx, f *entity.SyncFailure) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO approval__sync_failures (
			id, approval_request_id, subject_type, subject_id, outcome, reason,
			attempt_count, max_attempts, last_error, status,
			created_at, updated_at, resolved_at, resolved_by
		) VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,$10,$11,$12,$13,$14)`,
		f.ID, f.ApprovalRequestID, string(f.SubjectType), f.SubjectID, string(f.Outcome), f.Reason,
		f.AttemptCount, f.MaxAttempts, f.LastError, string(f.Status),
		f.CreatedAt, f.UpdatedAt, f.ResolvedAt, f.ResolvedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting approval sync failure: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetSyncFailure(ctx context.Context, id uuid.UUID) (*entity.SyncFailure, error) {
	row := r.pool.QueryRow(ctx, syncFailureSelect+" WHERE id = $1", id)
	return scanSyncFailure(row)
}

func (r *PostgresRepository) GetSyncFailureForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.SyncFailure, error) {
	row := tx.QueryRow(ctx, syncFailureSelect+" WHERE id = $1 FOR UPDATE", id)
	return scanSyncFailure(row)
}

func (r *PostgresRepository) ListSyncFailures(ctx context.Context, f domain.SyncFailureFilter) ([]*entity.SyncFailure, int, error) {
	page, limit := f.Page, f.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	where := ""
	args := []any{}
	if f.Status != "" {
		where = " WHERE status = $1"
		args = append(args, string(f.Status))
	}

	var total int
	countArgs := append([]any{}, args...)
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM approval__sync_failures"+where, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting approval sync failures: %w", err)
	}

	args = append(args, limit, (page-1)*limit)
	rows, err := r.pool.Query(ctx, syncFailureSelect+where+
		fmt.Sprintf(" ORDER BY created_at ASC LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing approval sync failures: %w", err)
	}
	defer rows.Close()

	var out []*entity.SyncFailure
	for rows.Next() {
		sf, err := scanSyncFailureRow(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, sf)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepository) UpdateSyncFailure(ctx context.Context, tx pgx.Tx, f *entity.SyncFailure) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__sync_failures
		   SET attempt_count = $2,
		       last_error    = $3,
		       status        = $4,
		       updated_at    = $5,
		       resolved_at   = $6,
		       resolved_by   = $7
		 WHERE id = $1`,
		f.ID, f.AttemptCount, f.LastError, string(f.Status), f.UpdatedAt, f.ResolvedAt, f.ResolvedBy,
	)
	if err != nil {
		return fmt.Errorf("updating approval sync failure %s: %w", f.ID, err)
	}
	return nil
}

type syncFailureScanner interface {
	Scan(dest ...any) error
}

func scanSyncFailure(s syncFailureScanner) (*entity.SyncFailure, error) {
	var f entity.SyncFailure
	var subjectType, outcome, status string
	err := s.Scan(
		&f.ID, &f.ApprovalRequestID, &subjectType, &f.SubjectID, &outcome, &f.Reason,
		&f.AttemptCount, &f.MaxAttempts, &f.LastError, &status,
		&f.CreatedAt, &f.UpdatedAt, &f.ResolvedAt, &f.ResolvedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning approval sync failure: %w", err)
	}
	f.SubjectType = vo.SubjectType(subjectType)
	f.Outcome = vo.SyncFailureOutcome(outcome)
	f.Status = vo.SyncFailureStatus(status)
	return &f, nil
}

func scanSyncFailureRow(rows pgx.Rows) (*entity.SyncFailure, error) {
	return scanSyncFailure(rows)
}
