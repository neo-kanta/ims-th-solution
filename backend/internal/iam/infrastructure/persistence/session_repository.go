package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

// PostgresSessionRepository implements domain.SessionRepository.
type PostgresSessionRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresSessionRepository creates a new session repository.
func NewPostgresSessionRepository(pool *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{pool: pool}
}

// Create inserts a new refresh token session.
func (r *PostgresSessionRepository) Create(ctx context.Context, session *entity.Session) error {
	query := `
		INSERT INTO iam_sessions (id, user_id, refresh_token_hash, token_family,
		                          ip_address, user_agent, expires_at,
		                          absolute_expires_at, last_activity_at, device_fingerprint)
		VALUES ($1, $2, $3, $4, $5::inet, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		session.ID, session.UserID, session.RefreshTokenHash, session.TokenFamily,
		nullableString(session.IPAddress), session.UserAgent, session.ExpiresAt,
		session.AbsoluteExpiresAt, session.LastActivityAt,
		nullableString(session.DeviceFingerprint),
	)
	if err != nil {
		return fmt.Errorf("inserting session: %w", err)
	}
	return nil
}

// FindByTokenHash retrieves a session by its refresh token hash.
func (r *PostgresSessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Session, error) {
	return r.scanSession(ctx, `
		SELECT id, user_id, refresh_token_hash, token_family,
		       host(ip_address), user_agent, is_revoked, expires_at,
		       absolute_expires_at, last_activity_at, device_fingerprint, revoke_reason,
		       created_at, rotated_at
		FROM iam_sessions
		WHERE refresh_token_hash = $1
	`, tokenHash)
}

// FindByID retrieves a session by its UUID.
func (r *PostgresSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Session, error) {
	return r.scanSession(ctx, `
		SELECT id, user_id, refresh_token_hash, token_family,
		       host(ip_address), user_agent, is_revoked, expires_at,
		       absolute_expires_at, last_activity_at, device_fingerprint, revoke_reason,
		       created_at, rotated_at
		FROM iam_sessions
		WHERE id = $1
	`, id)
}

func (r *PostgresSessionRepository) scanSession(ctx context.Context, query string, arg interface{}) (*entity.Session, error) {
	var s entity.Session
	var ipAddr *string
	var deviceFp *string
	var revokeReason *string
	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&s.ID, &s.UserID, &s.RefreshTokenHash, &s.TokenFamily,
		&ipAddr, &s.UserAgent, &s.IsRevoked, &s.ExpiresAt,
		&s.AbsoluteExpiresAt, &s.LastActivityAt, &deviceFp, &revokeReason,
		&s.CreatedAt, &s.RotatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying session: %w", err)
	}
	if ipAddr != nil {
		s.IPAddress = *ipAddr
	}
	if deviceFp != nil {
		s.DeviceFingerprint = *deviceFp
	}
	if revokeReason != nil {
		s.RevokeReason = *revokeReason
	}
	return &s, nil
}

// RevokeByID revokes a single session.
func (r *PostgresSessionRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE iam_sessions SET is_revoked = true, rotated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	return nil
}

// RevokeByIDWithReason revokes a single session with a reason.
func (r *PostgresSessionRepository) RevokeByIDWithReason(ctx context.Context, id uuid.UUID, reason string) error {
	query := `UPDATE iam_sessions SET is_revoked = true, rotated_at = NOW(), revoke_reason = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, reason)
	if err != nil {
		return fmt.Errorf("revoking session: %w", err)
	}
	return nil
}

// RevokeByFamily revokes all sessions in a token family (breach detection).
func (r *PostgresSessionRepository) RevokeByFamily(ctx context.Context, family uuid.UUID) error {
	query := `UPDATE iam_sessions SET is_revoked = true, rotated_at = NOW(), revoke_reason = 'breach'
	          WHERE token_family = $1 AND is_revoked = false`
	_, err := r.pool.Exec(ctx, query, family)
	if err != nil {
		return fmt.Errorf("revoking session family: %w", err)
	}
	return nil
}

// RevokeAllForUser revokes all sessions for a user (force logout).
func (r *PostgresSessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE iam_sessions SET is_revoked = true, rotated_at = NOW()
	          WHERE user_id = $1 AND is_revoked = false`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoking all sessions for user: %w", err)
	}
	return nil
}

// RevokeAllForUserWithReason revokes all sessions for a user with a reason.
func (r *PostgresSessionRepository) RevokeAllForUserWithReason(ctx context.Context, userID uuid.UUID, reason string) error {
	query := `UPDATE iam_sessions SET is_revoked = true, rotated_at = NOW(), revoke_reason = $2
	          WHERE user_id = $1 AND is_revoked = false`
	_, err := r.pool.Exec(ctx, query, userID, reason)
	if err != nil {
		return fmt.Errorf("revoking all sessions for user: %w", err)
	}
	return nil
}

// DeleteExpired removes expired sessions to keep the table lean.
func (r *PostgresSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.pool.Exec(ctx, `DELETE FROM iam_sessions WHERE expires_at < NOW()`)
	if err != nil {
		return 0, fmt.Errorf("deleting expired sessions: %w", err)
	}
	return result.RowsAffected(), nil
}

// ListActiveForUser returns all active (non-revoked, non-expired) sessions for a user.
func (r *PostgresSessionRepository) ListActiveForUser(ctx context.Context, userID uuid.UUID) ([]entity.Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, token_family,
		       host(ip_address), user_agent, is_revoked, expires_at,
		       absolute_expires_at, last_activity_at, device_fingerprint, revoke_reason,
		       created_at, rotated_at
		FROM iam_sessions
		WHERE user_id = $1 AND is_revoked = false AND expires_at > NOW()
		ORDER BY last_activity_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("listing active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []entity.Session
	for rows.Next() {
		var s entity.Session
		var ipAddr, deviceFp, revokeReason *string
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.RefreshTokenHash, &s.TokenFamily,
			&ipAddr, &s.UserAgent, &s.IsRevoked, &s.ExpiresAt,
			&s.AbsoluteExpiresAt, &s.LastActivityAt, &deviceFp, &revokeReason,
			&s.CreatedAt, &s.RotatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning session: %w", err)
		}
		if ipAddr != nil {
			s.IPAddress = *ipAddr
		}
		if deviceFp != nil {
			s.DeviceFingerprint = *deviceFp
		}
		if revokeReason != nil {
			s.RevokeReason = *revokeReason
		}
		sessions = append(sessions, s)
	}
	if sessions == nil {
		sessions = []entity.Session{}
	}
	return sessions, nil
}

// CountActiveForUser returns the count of active sessions for a user.
func (r *PostgresSessionRepository) CountActiveForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM iam_sessions WHERE user_id = $1 AND is_revoked = false AND expires_at > NOW()`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting active sessions: %w", err)
	}
	return count, nil
}

// RevokeOldestForUser revokes the oldest active session for a user.
func (r *PostgresSessionRepository) RevokeOldestForUser(ctx context.Context, userID uuid.UUID, reason string) error {
	query := `
		UPDATE iam_sessions SET is_revoked = true, rotated_at = NOW(), revoke_reason = $2
		WHERE id = (
			SELECT id FROM iam_sessions
			WHERE user_id = $1 AND is_revoked = false AND expires_at > NOW()
			ORDER BY created_at ASC
			LIMIT 1
		)
	`
	_, err := r.pool.Exec(ctx, query, userID, reason)
	if err != nil {
		return fmt.Errorf("revoking oldest session: %w", err)
	}
	return nil
}

// UpdateLastActivity updates the last_activity_at timestamp for a session.
func (r *PostgresSessionRepository) UpdateLastActivity(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE iam_sessions SET last_activity_at = NOW() WHERE id = $1 AND is_revoked = false AND (last_activity_at IS NULL OR last_activity_at < NOW() - interval '60 seconds')`,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("updating last activity: %w", err)
	}
	return nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
