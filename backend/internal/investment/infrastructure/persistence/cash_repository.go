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

// PostgresCashLedgerRepository implements domain.CashLedgerRepository.
type PostgresCashLedgerRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresCashLedgerRepository wires the repo.
func NewPostgresCashLedgerRepository(pool *pgxpool.Pool) *PostgresCashLedgerRepository {
	return &PostgresCashLedgerRepository{pool: pool}
}

func (r *PostgresCashLedgerRepository) InsertMovement(ctx context.Context, tx pgx.Tx, m *entity.CashMovement) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__cash_movements (
			id, portfolio_id, currency, amount, business_date,
			transaction_id, movement_type,
			created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		m.ID, m.PortfolioID, m.Currency, m.Amount, m.BusinessDate,
		m.TransactionID, m.MovementType,
		m.CreatedAt, m.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting cash movement: %w", err)
	}
	return nil
}

const cashBalSelect = `
	SELECT id, portfolio_id, currency, balance,
	       last_movement_id, last_business_date,
	       version, updated_at
	FROM investment__cash_balances`

func (r *PostgresCashLedgerRepository) GetBalanceForUpdate(ctx context.Context, tx pgx.Tx, portfolioID uuid.UUID, currency string) (*entity.CashBalance, error) {
	row := tx.QueryRow(ctx,
		cashBalSelect+" WHERE portfolio_id = $1 AND currency = $2 FOR UPDATE",
		portfolioID, currency,
	)
	return scanCashBalance(row)
}

// UpsertBalance applies a cash-balance projection with the same version-guarded
// semantics as PostgresPositionRepository.Upsert. See its docstring for the
// concurrency contract — both repos share the same shape so the projector can
// retry uniformly.
func (r *PostgresCashLedgerRepository) UpsertBalance(ctx context.Context, tx pgx.Tx, b *entity.CashBalance, expectedVersion int) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}

	const upsertSQL = `
		INSERT INTO investment__cash_balances (
			id, portfolio_id, currency, balance,
			last_movement_id, last_business_date,
			version, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6,
			$7, NOW()
		)
		ON CONFLICT (portfolio_id, currency) DO UPDATE
		   SET balance            = EXCLUDED.balance,
		       last_movement_id   = EXCLUDED.last_movement_id,
		       last_business_date = EXCLUDED.last_business_date,
		       version            = EXCLUDED.version,
		       updated_at         = NOW()
		 WHERE investment__cash_balances.version = $8
		RETURNING version`

	var newVersion int
	err := tx.QueryRow(ctx, upsertSQL,
		b.ID, b.PortfolioID, b.Currency, b.Balance,
		b.LastMovementID, b.LastBusinessDate,
		b.Version, expectedVersion,
	).Scan(&newVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.ErrPortfolioVersionMismatch{
				PortfolioID:     b.PortfolioID.String(),
				ExpectedVersion: expectedVersion,
				ActualVersion:   -1,
			}
		}
		return fmt.Errorf("upserting cash balance: %w", err)
	}
	return nil
}

func (r *PostgresCashLedgerRepository) ListBalances(ctx context.Context, portfolioID uuid.UUID) ([]*entity.CashBalance, error) {
	rows, err := r.pool.Query(ctx,
		cashBalSelect+" WHERE portfolio_id = $1 ORDER BY currency",
		portfolioID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing cash balances: %w", err)
	}
	defer rows.Close()

	out := []*entity.CashBalance{}
	for rows.Next() {
		b, err := scanCashBalance(rows)
		if err != nil {
			return nil, err
		}
		if b != nil {
			out = append(out, b)
		}
	}
	return out, rows.Err()
}

func scanCashBalance(s scanner) (*entity.CashBalance, error) {
	var b entity.CashBalance
	err := s.Scan(
		&b.ID, &b.PortfolioID, &b.Currency, &b.Balance,
		&b.LastMovementID, &b.LastBusinessDate,
		&b.Version, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning cash balance: %w", err)
	}
	return &b, nil
}

var _ domain.CashLedgerRepository = (*PostgresCashLedgerRepository)(nil)
