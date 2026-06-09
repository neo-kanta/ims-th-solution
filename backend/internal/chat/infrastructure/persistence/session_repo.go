// Package persistence holds the chat module's Postgres-backed repositories.
package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	chatdomain "github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// PostgresSessionRepository persists Session aggregates in chat_sessions.
type PostgresSessionRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresSessionRepository constructs the repository.
func NewPostgresSessionRepository(pool *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{pool: pool}
}

// compile-time interface conformance
var _ chatdomain.SessionRepository = (*PostgresSessionRepository)(nil)

// Create inserts a new session row.
func (r *PostgresSessionRepository) Create(ctx context.Context, s *entity.Session) error {
	const q = `
		INSERT INTO chat_sessions (id, user_id, provider, model, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, q, s.ID, s.UserID, string(s.Provider), s.Model, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert chat_session: %w", err)
	}
	return nil
}

// FindByID returns the session or nil when not found.
func (r *PostgresSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Session, error) {
	const q = `
		SELECT id, user_id, provider, model, created_at, updated_at
		FROM chat_sessions
		WHERE id = $1
	`
	var s entity.Session
	var provider string
	row := r.pool.QueryRow(ctx, q, id)
	if err := row.Scan(&s.ID, &s.UserID, &provider, &s.Model, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select chat_session: %w", err)
	}
	s.Provider = valueobject.ProviderID(provider)
	return &s, nil
}

// Touch updates updated_at = now() for the given session.
func (r *PostgresSessionRepository) Touch(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE chat_sessions SET updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("touch chat_session: %w", err)
	}
	return nil
}

// UpdateModel updates the provider and model for a given session.
func (r *PostgresSessionRepository) UpdateModel(ctx context.Context, id uuid.UUID, provider, model string) error {
	const q = `UPDATE chat_sessions SET provider = $2, model = $3, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id, provider, model)
	if err != nil {
		return fmt.Errorf("update model chat_session: %w", err)
	}
	return nil
}

// ListByUser returns one page of a user's sessions (most-recently-updated first)
// plus the total count.
func (r *PostgresSessionRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]entity.Session, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM chat_sessions WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count chat_sessions: %w", err)
	}

	const q = `
		SELECT id, user_id, provider, model, created_at, updated_at
		FROM chat_sessions
		WHERE user_id = $1
		ORDER BY updated_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("select chat_sessions by user: %w", err)
	}
	defer rows.Close()

	var out []entity.Session
	for rows.Next() {
		var s entity.Session
		var provider string
		if err := rows.Scan(&s.ID, &s.UserID, &provider, &s.Model, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan chat_session: %w", err)
		}
		s.Provider = valueobject.ProviderID(provider)
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate chat_sessions: %w", err)
	}
	return out, total, nil
}
