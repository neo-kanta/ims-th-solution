// Package postgres holds the Postgres-backed implementation of the approval
// module's repositories. One PostgresRepository type satisfies the aggregate
// domain.Repository interface; read methods use the pool, write methods accept
// the caller's pgx.Tx so the application layer controls transaction boundaries.
package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
)

// PostgresRepository implements domain.Repository.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository wires the repository to a pgx pool.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Compile-time assertion that the Postgres repo satisfies the aggregate port.
var _ domain.Repository = (*PostgresRepository)(nil)
