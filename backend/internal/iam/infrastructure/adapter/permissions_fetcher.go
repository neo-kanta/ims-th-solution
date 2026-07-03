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
		SELECT DISTINCT permission_code
		FROM (
			SELECT fr.permission_code
			FROM permissions_function_rights fr
			INNER JOIN permissions_accounts_groups ag ON ag.group_id = fr.group_id
			WHERE ag.user_id = $1 AND fr.is_granted = true

			UNION

			SELECT fr.permission_code
			FROM permission_function_rights fr
			WHERE (fr.can_view OR fr.can_search OR fr.can_add OR fr.can_edit OR fr.can_delete
			       OR fr.can_approve OR fr.can_revoke_approval OR fr.can_export OR fr.can_configure)
			  AND (
				(fr.subject_type = 'USER' AND fr.subject_id = $1)
				OR (fr.subject_type = 'GROUP' AND EXISTS (
					SELECT 1
					FROM permissions_accounts_groups ag
					WHERE ag.group_id = fr.subject_id AND ag.user_id = $1
				))
				OR (fr.subject_type = 'ROLE' AND EXISTS (
					SELECT 1
					FROM permission_user_role_assignments ura
					WHERE ura.role_id = fr.subject_id
					  AND ura.user_id = $1
					  AND ura.status = 'APPROVED'
					  AND (ura.effective_from IS NULL OR ura.effective_from <= NOW())
					  AND (ura.effective_to IS NULL OR ura.effective_to > NOW())
				))
			  )
		) effective_permissions
		ORDER BY permission_code
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
		SELECT DISTINCT contract_id
		FROM (
			SELECT contract_id
			FROM permissions_data_rights
			WHERE user_id = $1 AND is_granted = true

			UNION

			SELECT COALESCE(dr.contract_id, dr.fund_id, dr.portfolio_id)::text AS contract_id
			FROM permission_data_rights dr
			WHERE (
				(dr.subject_type = 'USER' AND dr.subject_id = $1)
				OR (dr.subject_type = 'GROUP' AND EXISTS (
					SELECT 1
					FROM permissions_accounts_groups ag
					WHERE ag.group_id = dr.subject_id AND ag.user_id = $1
				))
				OR (dr.subject_type = 'ROLE' AND EXISTS (
					SELECT 1
					FROM permission_user_role_assignments ura
					WHERE ura.role_id = dr.subject_id
					  AND ura.user_id = $1
					  AND ura.status = 'APPROVED'
					  AND (ura.effective_from IS NULL OR ura.effective_from <= NOW())
					  AND (ura.effective_to IS NULL OR ura.effective_to > NOW())
				))
			)
		) effective_data_permissions
		WHERE contract_id IS NOT NULL
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
