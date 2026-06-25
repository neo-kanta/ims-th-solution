package adapter

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
)

// PostgresUserLookupAdapter resolves user display info from iam_users.
//
// This is a read-only cross-module projection — the same pattern used by the
// approval module's PostgresUserDirectory. It does NOT import internal/iam.
type PostgresUserLookupAdapter struct {
	pool *pgxpool.Pool
}

func NewPostgresUserLookupAdapter(pool *pgxpool.Pool) *PostgresUserLookupAdapter {
	return &PostgresUserLookupAdapter{pool: pool}
}

func (a *PostgresUserLookupAdapter) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*watchlistdomain.UserInfo, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]*watchlistdomain.UserInfo{}, nil
	}
	rows, err := a.pool.Query(ctx,
		`SELECT id, COALESCE(display_name, username, '') FROM iam_users WHERE id = ANY($1)`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("looking up user display names: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*watchlistdomain.UserInfo, len(ids))
	for rows.Next() {
		var info watchlistdomain.UserInfo
		if err := rows.Scan(&info.UserID, &info.DisplayName); err != nil {
			return nil, fmt.Errorf("scanning user info: %w", err)
		}
		result[info.UserID] = &info
	}
	return result, rows.Err()
}

var _ watchlistdomain.UserLookupPort = (*PostgresUserLookupAdapter)(nil)
