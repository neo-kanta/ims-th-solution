package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
)

// PostgresOverrideRepository implements domain.OverrideRepository.
// Append-only: overrides are inserted via CommitOverride inside an atomic
// transaction, and never updated or deleted thereafter.
type PostgresOverrideRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresOverrideRepository creates a new instance.
func NewPostgresOverrideRepository(pool *pgxpool.Pool) *PostgresOverrideRepository {
	return &PostgresOverrideRepository{pool: pool}
}

// pgUniqueViolation is the SQLSTATE for unique_violation (integrity constraint).
// Documented: https://www.postgresql.org/docs/current/errcodes-appendix.html
const pgUniqueViolation = "23505"

// CommitOverride atomically locks the target breach, verifies its state, inserts
// the override row, and updates the breach to OVERRIDDEN — all inside one tx.
//
// Race safety is provided by the combination of:
//
//   - SELECT ... FOR UPDATE serialises concurrent attempts on the same breach row.
//   - UNIQUE(breach_id) on compliance_overrides is the database-level backstop:
//     if a second transaction somehow slips past the row lock, the INSERT fails
//     with SQLSTATE 23505 and is mapped to *domain.ErrOverrideAlreadyExists.
//
// Any error causes a rollback; no partial writes are ever persisted.
func (r *PostgresOverrideRepository) CommitOverride(
	ctx context.Context,
	o *entity.Override,
) error {
	if o == nil {
		return fmt.Errorf("override is nil")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin override tx: %w", err)
	}
	// Rollback is a no-op after a successful commit.
	defer func() { _ = tx.Rollback(ctx) }()

	// Step 1 — lock the target breach row.
	var currentStatus string
	err = tx.QueryRow(ctx, `
		SELECT status
		FROM compliance_breaches
		WHERE id = $1
		FOR UPDATE
	`, o.BreachID).Scan(&currentStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return &domain.ErrBreachNotFound{BreachID: o.BreachID.String()}
	}
	if err != nil {
		return fmt.Errorf("locking breach: %w", err)
	}

	// Step 2 — verify breach state.
	if currentStatus != string(entity.BreachStatusOpen) {
		return &domain.ErrBreachNotOpen{
			BreachID:      o.BreachID.String(),
			CurrentStatus: currentStatus,
		}
	}

	// Step 3 — insert the immutable override record.
	// UNIQUE(breach_id) maps to ErrOverrideAlreadyExists on collision.
	_, err = tx.Exec(ctx, `
		INSERT INTO compliance_overrides
			(id, breach_id, reason, overridden_by, delegated_from, approved_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`,
		o.ID, o.BreachID, o.Reason,
		o.OverriddenBy, o.DelegatedFrom, o.ApprovedBy, o.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return &domain.ErrOverrideAlreadyExists{BreachID: o.BreachID.String()}
		}
		return fmt.Errorf("inserting override: %w", err)
	}

	// Step 4 — transition the breach to OVERRIDDEN with a shared UTC timestamp.
	resolvedBy := o.OverriddenBy
	resolvedAt := o.CreatedAt
	if _, err := tx.Exec(ctx, `
		UPDATE compliance_breaches
		SET status = $2, resolved_by = $3, resolved_at = $4, updated_at = $4
		WHERE id = $1
	`, o.BreachID, string(entity.BreachStatusOverridden), resolvedBy, resolvedAt); err != nil {
		return fmt.Errorf("updating breach status: %w", err)
	}

	// Step 5 — commit.
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing override tx: %w", err)
	}
	return nil
}

// GetByBreachID returns the override for a specific breach (at most one,
// enforced by UNIQUE(breach_id)).
func (r *PostgresOverrideRepository) GetByBreachID(ctx context.Context, breachID uuid.UUID) (*entity.Override, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, breach_id, reason, overridden_by, delegated_from, approved_by, created_at
		FROM compliance_overrides
		WHERE breach_id = $1
		LIMIT 1
	`, breachID)
	if err != nil {
		return nil, fmt.Errorf("querying override: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return scanOverride(rows)
}

// List returns overrides for a slice of breach IDs (used for bulk display).
func (r *PostgresOverrideRepository) List(ctx context.Context, breachIDs []uuid.UUID) ([]entity.Override, error) {
	if len(breachIDs) == 0 {
		return nil, nil
	}

	// Build $1,$2,... placeholder list.
	placeholders := make([]string, len(breachIDs))
	args := make([]interface{}, len(breachIDs))
	for i, id := range breachIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, breach_id, reason, overridden_by, delegated_from, approved_by, created_at
		FROM compliance_overrides
		WHERE breach_id IN (%s)
		ORDER BY created_at DESC
	`, joinStrings(placeholders, ",")), args...)
	if err != nil {
		return nil, fmt.Errorf("listing overrides: %w", err)
	}
	defer rows.Close()

	var results []entity.Override
	for rows.Next() {
		o, err := scanOverride(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning override: %w", err)
		}
		results = append(results, *o)
	}
	return results, rows.Err()
}

// --- row scanner ---

func scanOverride(s pgx.Rows) (*entity.Override, error) {
	var o entity.Override
	err := s.Scan(
		&o.ID, &o.BreachID, &o.Reason,
		&o.OverriddenBy, &o.DelegatedFrom, &o.ApprovedBy, &o.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	o.CreatedAt = o.CreatedAt.UTC()
	return &o, nil
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// ensure interface satisfaction at compile time
var _ domain.OverrideRepository = (*PostgresOverrideRepository)(nil)
