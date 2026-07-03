package adapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
)

// PostgresUserDirectory resolves user display info directly from iam_users.
//
// This is a read-only projection across the IAM table (the same pattern the
// permissions module uses). It does NOT import internal/iam, keeping the module
// boundary intact while letting signature/stamp records carry real names.
type PostgresUserDirectory struct {
	pool *pgxpool.Pool
}

// NewPostgresUserDirectory wires the directory to a pgx pool.
func NewPostgresUserDirectory(pool *pgxpool.Pool) *PostgresUserDirectory {
	return &PostgresUserDirectory{pool: pool}
}

// GetUser returns the user's display name (and best-effort title) or nil.
func (d *PostgresUserDirectory) GetUser(ctx context.Context, id uuid.UUID) (*domain.UserInfo, error) {
	if d == nil || d.pool == nil {
		return nil, nil
	}
	var info domain.UserInfo
	info.ID = id
	err := d.pool.QueryRow(ctx, `
		SELECT COALESCE(display_name, username, ''), COALESCE(username, '')
		FROM iam_users WHERE id = $1`, id).Scan(&info.DisplayName, &info.Title)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("resolving user info: %w", err)
	}
	// iam_users has no job-title column in the PoC schema; expose username as a
	// stable secondary label rather than inventing data.
	info.Title = ""
	return &info, nil
}

var _ domain.UserDirectory = (*PostgresUserDirectory)(nil)
