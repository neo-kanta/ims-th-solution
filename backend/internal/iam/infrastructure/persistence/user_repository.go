package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/your-org/ims-th-solution/backend/internal/iam/domain/entity"
)

// PostgresUserRepository implements domain.UserRepository using PostgreSQL.
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository creates a new PostgreSQL-backed user repository.
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

// FindByUsername retrieves a user by their login username, including group memberships.
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	return r.findUser(ctx, "u.username = $1", username)
}

// FindByID retrieves a user by their UUID, including group memberships.
func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return r.findUser(ctx, "u.id = $1", id)
}

// findUser is a shared query method that loads a user with their group memberships.
func (r *PostgresUserRepository) findUser(ctx context.Context, whereClause string, arg interface{}) (*entity.User, error) {
	query := fmt.Sprintf(`
		SELECT u.id, u.username, u.display_name, u.email, u.password_hash,
		       u.is_active, u.is_on_leave, u.created_at, u.updated_at,
		       u.created_by, u.updated_by
		FROM iam_users u
		WHERE %s
	`, whereClause)

	var user entity.User
	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.PasswordHash,
		&user.IsActive, &user.IsOnLeave, &user.CreatedAt, &user.UpdatedAt,
		&user.CreatedBy, &user.UpdatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying user: %w", err)
	}

	// Load group memberships
	groups, err := r.loadUserGroups(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("loading user groups: %w", err)
	}
	user.Groups = groups

	return &user, nil
}

// loadUserGroups retrieves the group names for a given user.
func (r *PostgresUserRepository) loadUserGroups(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT g.name
		FROM permissions_groups g
		INNER JOIN permissions_accounts_groups ag ON ag.group_id = g.id
		WHERE ag.user_id = $1 AND g.is_active = true
		ORDER BY g.name
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying user groups: %w", err)
	}
	defer rows.Close()

	var groups []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scanning group name: %w", err)
		}
		groups = append(groups, name)
	}

	if groups == nil {
		groups = []string{}
	}

	return groups, nil
}
