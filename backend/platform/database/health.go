package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthCheck verifies the database connection is alive.
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
