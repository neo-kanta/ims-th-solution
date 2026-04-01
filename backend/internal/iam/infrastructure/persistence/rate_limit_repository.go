package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRateLimitRepository implements domain.RateLimitRepository.
type PostgresRateLimitRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRateLimitRepository creates a new rate limit repository.
func NewPostgresRateLimitRepository(pool *pgxpool.Pool) *PostgresRateLimitRepository {
	return &PostgresRateLimitRepository{pool: pool}
}

// RecordAttempt inserts a login attempt record.
func (r *PostgresRateLimitRepository) RecordAttempt(ctx context.Context, ipAddress string, username string, success bool) error {
	query := `INSERT INTO iam_login_attempts (ip_address, username, success) VALUES ($1::inet, $2, $3)`
	_, err := r.pool.Exec(ctx, query, nullableString(ipAddress), nullableString(username), success)
	if err != nil {
		return fmt.Errorf("recording login attempt: %w", err)
	}
	return nil
}

// CountRecentFailures counts failed login attempts from an IP within the given interval.
func (r *PostgresRateLimitRepository) CountRecentFailures(ctx context.Context, ipAddress string, window string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM iam_login_attempts
		 WHERE ip_address = $1::inet AND success = false
		 AND attempted_at > NOW() - $2::interval`,
		ipAddress, window,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting recent failures: %w", err)
	}
	return count, nil
}

// CountRecentFailuresByUser counts failed login attempts from an IP+username within the given interval.
func (r *PostgresRateLimitRepository) CountRecentFailuresByUser(ctx context.Context, ipAddress string, username string, window string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM iam_login_attempts
		 WHERE ip_address = $1::inet AND username = $2 AND success = false
		 AND attempted_at > NOW() - $3::interval`,
		ipAddress, username, window,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting recent failures by user: %w", err)
	}
	return count, nil
}

// PurgeOld removes login attempts older than the given interval.
func (r *PostgresRateLimitRepository) PurgeOld(ctx context.Context, olderThan string) (int64, error) {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM iam_login_attempts WHERE attempted_at < NOW() - $1::interval`,
		olderThan,
	)
	if err != nil {
		return 0, fmt.Errorf("purging old login attempts: %w", err)
	}
	return result.RowsAffected(), nil
}
