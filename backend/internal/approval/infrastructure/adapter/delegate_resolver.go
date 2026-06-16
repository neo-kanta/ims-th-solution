package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
)

// PostgresDelegateResolver looks up the active delegation for a user in the
// approval__delegations table. It satisfies domain.DelegateResolver.
type PostgresDelegateResolver struct {
	pool *pgxpool.Pool
}

// NewPostgresDelegateResolver creates a resolver backed by the delegations table.
func NewPostgresDelegateResolver(pool *pgxpool.Pool) *PostgresDelegateResolver {
	return &PostgresDelegateResolver{pool: pool}
}

// ResolveDelegate returns the active delegate for originalUserID at asOf, or
// (nil, nil) if no active delegation exists. When contractID is non-nil, rows
// with a matching contract_id take precedence over wildcard (NULL) rows.
func (r *PostgresDelegateResolver) ResolveDelegate(
	ctx context.Context,
	originalUserID uuid.UUID,
	contractID *uuid.UUID,
	asOf time.Time,
) (*domain.Delegate, error) {
	const q = `
		SELECT to_user_id
		FROM   approval__delegations
		WHERE  from_user_id = $1
		  AND  is_active     = TRUE
		  AND  active_from  <= $2
		  AND  active_until  > $2
		  AND  (contract_id IS NULL OR contract_id = $3)
		ORDER BY
		  -- prefer contract-specific over wildcard rows
		  CASE WHEN contract_id IS NOT NULL THEN 0 ELSE 1 END,
		  active_from DESC
		LIMIT 1
	`
	var cid interface{}
	if contractID != nil {
		cid = *contractID
	}
	var toUserID uuid.UUID
	err := r.pool.QueryRow(ctx, q, originalUserID, asOf, cid).Scan(&toUserID)
	if err != nil {
		// pgx returns pgx.ErrNoRows when no row is found; treat as no-delegate.
		return nil, nil //nolint:nilerr
	}
	return &domain.Delegate{DelegateUserID: toUserID}, nil
}
