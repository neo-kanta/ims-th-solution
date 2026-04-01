package adapter

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PermissionsFetcher queries the permissions tables to get a user's effective permissions.
// This is a read-only adapter used by the IAM module's login flow.
type PermissionsFetcher struct {
	pool *pgxpool.Pool
}

// NewPermissionsFetcher creates a new permissions fetcher.
func NewPermissionsFetcher(pool *pgxpool.Pool) *PermissionsFetcher {
	return &PermissionsFetcher{pool: pool}
}

// GetUserFunctionPermissions returns function permission codes granted to the user
// through their group memberships.
func (f *PermissionsFetcher) GetUserFunctionPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT fr.permission_code
		FROM permissions_function_rights fr
		INNER JOIN permissions_accounts_groups ag ON ag.group_id = fr.group_id
		WHERE ag.user_id = $1 AND fr.is_granted = true
		ORDER BY fr.permission_code
	`

	rows, err := f.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying function permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("scanning permission code: %w", err)
		}
		permissions = append(permissions, code)
	}
	if permissions == nil {
		permissions = []string{}
	}

	return permissions, nil
}

// GetUserDataPermissions returns the contract IDs the user has data access to.
func (f *PermissionsFetcher) GetUserDataPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT contract_id
		FROM permissions_data_rights
		WHERE user_id = $1 AND is_granted = true
		ORDER BY contract_id
	`

	rows, err := f.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying data permissions: %w", err)
	}
	defer rows.Close()

	var contracts []string
	for rows.Next() {
		var contractID string
		if err := rows.Scan(&contractID); err != nil {
			return nil, fmt.Errorf("scanning contract id: %w", err)
		}
		contracts = append(contracts, contractID)
	}
	if contracts == nil {
		contracts = []string{}
	}

	return contracts, nil
}
