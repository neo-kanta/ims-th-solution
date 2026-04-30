package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
)

// PostgresPositionRepository implements domain.PortfolioPositionRepository.
type PostgresPositionRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPositionRepository wires the repo.
func NewPostgresPositionRepository(pool *pgxpool.Pool) *PostgresPositionRepository {
	return &PostgresPositionRepository{pool: pool}
}

const positionSelect = `
	SELECT id, portfolio_id, instrument_id,
	       quantity, average_cost, cost_basis,
	       last_transaction_id, last_business_date,
	       version, updated_at
	FROM investment__portfolio_positions`

// GetForUpdate locks the (portfolio, instrument) row for the duration of the
// transaction. Returns nil (no error) when no row exists yet.
func (r *PostgresPositionRepository) GetForUpdate(ctx context.Context, tx pgx.Tx, portfolioID, instrumentID uuid.UUID) (*entity.PortfolioPosition, error) {
	row := tx.QueryRow(ctx,
		positionSelect+" WHERE portfolio_id = $1 AND instrument_id = $2 FOR UPDATE",
		portfolioID, instrumentID,
	)
	return scanPosition(row)
}

// Upsert applies a position projection with strict version-guarded semantics:
//
//   - When no row exists yet (caller passes expectedVersion=0 and pos.Version=1),
//     a fresh INSERT lands.
//   - When a row exists with the expected version, the row is updated and the
//     new version is returned.
//   - When a concurrent writer changed the version between the caller's
//     GetForUpdate and this Upsert, the WHERE guard fails, RETURNING is empty,
//     and *domain.ErrPortfolioVersionMismatch is raised. Callers (the projector)
//     re-read with GetForUpdate and retry once.
//
// One SQL round-trip handles both branches atomically; ON CONFLICT DO NOTHING
// is intentionally NOT used here because it would silently drop the second
// concurrent first-inserter.
func (r *PostgresPositionRepository) Upsert(ctx context.Context, tx pgx.Tx, pos *entity.PortfolioPosition, expectedVersion int) error {
	if pos.ID == uuid.Nil {
		pos.ID = uuid.New()
	}

	const upsertSQL = `
		INSERT INTO investment__portfolio_positions (
			id, portfolio_id, instrument_id,
			quantity, average_cost, cost_basis,
			last_transaction_id, last_business_date,
			version, updated_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6,
			$7, $8,
			$9, NOW()
		)
		ON CONFLICT (portfolio_id, instrument_id) DO UPDATE
		   SET quantity            = EXCLUDED.quantity,
		       average_cost        = EXCLUDED.average_cost,
		       cost_basis          = EXCLUDED.cost_basis,
		       last_transaction_id = EXCLUDED.last_transaction_id,
		       last_business_date  = EXCLUDED.last_business_date,
		       version             = EXCLUDED.version,
		       updated_at          = NOW()
		 WHERE investment__portfolio_positions.version = $10
		RETURNING version`

	var newVersion int
	err := tx.QueryRow(ctx, upsertSQL,
		pos.ID, pos.PortfolioID, pos.InstrumentID,
		pos.Quantity, pos.AverageCost, pos.CostBasis,
		pos.LastTransactionID, pos.LastBusinessDate,
		pos.Version, expectedVersion,
	).Scan(&newVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.ErrPortfolioVersionMismatch{
				PortfolioID:     pos.PortfolioID.String(),
				ExpectedVersion: expectedVersion,
				ActualVersion:   -1,
			}
		}
		return fmt.Errorf("upserting position: %w", err)
	}
	return nil
}

// ListByPortfolio returns all current positions for a portfolio (including zero ones).
func (r *PostgresPositionRepository) ListByPortfolio(ctx context.Context, portfolioID uuid.UUID) ([]*entity.PortfolioPosition, error) {
	rows, err := r.pool.Query(ctx,
		positionSelect+" WHERE portfolio_id = $1 ORDER BY instrument_id",
		portfolioID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing positions: %w", err)
	}
	defer rows.Close()

	out := []*entity.PortfolioPosition{}
	for rows.Next() {
		p, err := scanPosition(rows)
		if err != nil {
			return nil, err
		}
		if p != nil {
			out = append(out, p)
		}
	}
	return out, rows.Err()
}

func scanPosition(s scanner) (*entity.PortfolioPosition, error) {
	var p entity.PortfolioPosition
	err := s.Scan(
		&p.ID, &p.PortfolioID, &p.InstrumentID,
		&p.Quantity, &p.AverageCost, &p.CostBasis,
		&p.LastTransactionID, &p.LastBusinessDate,
		&p.Version, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning position: %w", err)
	}
	return &p, nil
}

var _ domain.PortfolioPositionRepository = (*PostgresPositionRepository)(nil)
