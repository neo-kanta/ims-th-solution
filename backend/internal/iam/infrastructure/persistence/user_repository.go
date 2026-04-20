package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

// PostgresUserRepository implements domain.UserRepository using PostgreSQL.
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository creates a new PostgreSQL-backed user repository.
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

// FindByUsername retrieves a user by their login username.
func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	return r.findUser(ctx, "u.username = $1 AND u.deleted_at IS NULL", username)
}

// FindByID retrieves a user by their UUID.
func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return r.findUser(ctx, "u.id = $1 AND u.deleted_at IS NULL", id)
}

func (r *PostgresUserRepository) findUser(ctx context.Context, whereClause string, arg interface{}) (*entity.User, error) {
	query := fmt.Sprintf(`
		SELECT u.id, u.username, u.display_name, u.email, u.password_hash,
		       u.is_active, u.force_password_change, u.failed_login_attempts,
		       u.locked_until, u.last_login_at, u.password_changed_at, u.version,
		       u.created_at, u.updated_at, u.created_by, u.updated_by, u.deleted_at
		FROM iam_users u
		WHERE %s
	`, whereClause)

	var user entity.User
	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.PasswordHash,
		&user.IsActive, &user.ForcePasswordChange, &user.FailedLoginAttempts,
		&user.LockedUntil, &user.LastLoginAt, &user.PasswordChangedAt, &user.Version,
		&user.CreatedAt, &user.UpdatedAt, &user.CreatedBy, &user.UpdatedBy, &user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying user: %w", err)
	}

	groups, err := r.loadUserGroups(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("loading user groups: %w", err)
	}
	user.Groups = groups

	return &user, nil
}

// Create inserts a new user.
func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO iam_users (id, username, display_name, email, password_hash,
		                       is_active, force_password_change, password_changed_at,
		                       created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Username, user.DisplayName, user.Email, user.PasswordHash,
		user.IsActive, user.ForcePasswordChange, user.PasswordChangedAt,
		user.CreatedBy, user.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

// Update saves changes to an existing user using optimistic locking.
func (r *PostgresUserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE iam_users
		SET username = $1, display_name = $2, email = $3, password_hash = $4,
		    is_active = $5, force_password_change = $6, failed_login_attempts = $7,
		    locked_until = $8, last_login_at = $9, password_changed_at = $10,
		    updated_by = $11, version = version + 1
		WHERE id = $12 AND version = $13 AND deleted_at IS NULL
	`
	result, err := r.pool.Exec(ctx, query,
		user.Username, user.DisplayName, user.Email, user.PasswordHash,
		user.IsActive, user.ForcePasswordChange, user.FailedLoginAttempts,
		user.LockedUntil, user.LastLoginAt, user.PasswordChangedAt,
		user.UpdatedBy,
		user.ID, user.Version,
	)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or version conflict")
	}
	user.Version++
	return nil
}

// SoftDelete marks a user as deleted.
func (r *PostgresUserRepository) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	query := `
		UPDATE iam_users SET deleted_at = NOW(), updated_by = $1
		WHERE id = $2 AND deleted_at IS NULL
	`
	result, err := r.pool.Exec(ctx, query, deletedBy, id)
	if err != nil {
		return fmt.Errorf("soft deleting user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// List returns paginated users matching the filter.
func (r *PostgresUserRepository) List(ctx context.Context, filter domain.UserFilter) ([]entity.User, int, error) {
	where := "u.deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if filter.IsActive != nil {
		where += fmt.Sprintf(" AND u.is_active = $%d", argIdx)
		args = append(args, *filter.IsActive)
		argIdx++
	}
	if filter.IsLocked != nil {
		if *filter.IsLocked {
			where += fmt.Sprintf(" AND u.locked_until IS NOT NULL AND u.locked_until > $%d", argIdx)
		} else {
			where += fmt.Sprintf(" AND (u.locked_until IS NULL OR u.locked_until <= $%d)", argIdx)
		}
		args = append(args, time.Now().UTC())
		argIdx++
	}
	if filter.Search != "" {
		where += fmt.Sprintf(" AND (u.username ILIKE $%d OR u.display_name ILIKE $%d OR u.email ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM iam_users u WHERE %s", where)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}

	// Fetch page
	dataQuery := fmt.Sprintf(`
		SELECT u.id, u.username, u.display_name, u.email, u.password_hash,
		       u.is_active, u.force_password_change, u.failed_login_attempts,
		       u.locked_until, u.last_login_at, u.password_changed_at, u.version,
		       u.created_at, u.updated_at, u.created_by, u.updated_by, u.deleted_at
		FROM iam_users u
		WHERE %s
		ORDER BY u.username
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.DisplayName, &u.Email, &u.PasswordHash,
			&u.IsActive, &u.ForcePasswordChange, &u.FailedLoginAttempts,
			&u.LockedUntil, &u.LastLoginAt, &u.PasswordChangedAt, &u.Version,
			&u.CreatedAt, &u.UpdatedAt, &u.CreatedBy, &u.UpdatedBy, &u.DeletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}
	if users == nil {
		users = []entity.User{}
	}

	// Load groups for each user
	for i := range users {
		groups, err := r.loadUserGroups(ctx, users[i].ID)
		if err != nil {
			return nil, 0, fmt.Errorf("loading groups for user %s: %w", users[i].ID, err)
		}
		users[i].Groups = groups
	}

	return users, total, nil
}

func (r *PostgresUserRepository) loadUserGroups(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT g.name
		FROM permissions_groups g
		INNER JOIN permissions_accounts_groups ag ON ag.group_id = g.id
		WHERE ag.user_id = $1 AND g.is_active = true AND g.deleted_at IS NULL
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
