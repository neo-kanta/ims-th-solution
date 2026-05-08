package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PostgresTransactionRepository implements domain.PortfolioTransactionRepository.
//
// The Postgres RULE on investment__portfolio_transactions blocks UPDATE and
// DELETE; this repo only exposes Insert / Read operations to mirror that
// invariant at the code level.
type PostgresTransactionRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTransactionRepository wires the repo.
func NewPostgresTransactionRepository(pool *pgxpool.Pool) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{pool: pool}
}

const txnSelect = `
	SELECT id, portfolio_id, fund_id, instrument_id,
	       transaction_type, side,
	       quantity, price, currency, gross_amount, fees, net_amount, realised_pnl_base, fx_rate_to_base,
	       business_date, settlement_date,
	       source_decision_id, source_execution_id, reverses_transaction_id,
	       COALESCE(external_ref, ''), COALESCE(reason, ''), status,
	       created_at, created_by
	FROM investment__portfolio_transactions`

func (r *PostgresTransactionRepository) Insert(ctx context.Context, tx pgx.Tx, t *entity.PortfolioTransaction) error {
	var sideStr *string
	if t.Side != nil {
		s := string(*t.Side)
		sideStr = &s
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__portfolio_transactions (
			id, portfolio_id, fund_id, instrument_id,
			transaction_type, side,
			quantity, price, currency, gross_amount, fees, net_amount, realised_pnl_base, fx_rate_to_base,
			business_date, settlement_date,
			source_decision_id, source_execution_id, reverses_transaction_id,
			external_ref, reason, status,
			created_at, created_by
		) VALUES (
			$1, $2, $3, $4,
			$5, $6,
			$7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16,
			$17, $18, $19,
			NULLIF($20,''), NULLIF($21,''), $22,
			$23, $24
		)`,
		t.ID, t.PortfolioID, t.FundID, t.InstrumentID,
		string(t.TransactionType), sideStr,
		t.Quantity, t.Price, t.Currency, t.GrossAmount, t.Fees, t.NetAmount, t.RealisedPnLBase, t.FxRateToBase,
		t.BusinessDate, t.SettlementDate,
		t.SourceDecisionID, t.SourceExecutionID, t.ReversesTransactionID,
		t.ExternalRef, t.Reason, string(t.Status),
		t.CreatedAt, t.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting portfolio transaction: %w", err)
	}
	return nil
}

func (r *PostgresTransactionRepository) ListForPositionReplay(ctx context.Context, tx pgx.Tx, portfolioID, instrumentID uuid.UUID) ([]*entity.PortfolioTransaction, error) {
	rows, err := tx.Query(ctx,
		txnSelect+`
		 WHERE portfolio_id = $1 AND instrument_id = $2
		 ORDER BY business_date ASC, created_at ASC, id ASC`,
		portfolioID, instrumentID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing transactions for replay: %w", err)
	}
	defer rows.Close()

	out := []*entity.PortfolioTransaction{}
	for rows.Next() {
		t, err := scanTxn(rows)
		if err != nil {
			return nil, err
		}
		if t != nil {
			out = append(out, t)
		}
	}
	return out, rows.Err()
}

func (r *PostgresTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.PortfolioTransaction, error) {
	row := r.pool.QueryRow(ctx, txnSelect+" WHERE id = $1", id)
	return scanTxn(row)
}

func (r *PostgresTransactionRepository) List(ctx context.Context, filter domain.TransactionListFilter) ([]*entity.PortfolioTransaction, int, error) {
	conds := []string{"1=1"}
	args := []any{}
	idx := 1
	if filter.PortfolioID != nil {
		conds = append(conds, fmt.Sprintf("portfolio_id = $%d", idx))
		args = append(args, *filter.PortfolioID)
		idx++
	}
	if filter.FundID != nil {
		conds = append(conds, fmt.Sprintf("fund_id = $%d", idx))
		args = append(args, *filter.FundID)
		idx++
	}
	if filter.InstrumentID != nil {
		conds = append(conds, fmt.Sprintf("instrument_id = $%d", idx))
		args = append(args, *filter.InstrumentID)
		idx++
	}
	if filter.Type != nil {
		conds = append(conds, fmt.Sprintf("transaction_type = $%d", idx))
		args = append(args, string(*filter.Type))
		idx++
	}
	if filter.BusinessFrom != nil {
		conds = append(conds, fmt.Sprintf("business_date >= $%d", idx))
		args = append(args, *filter.BusinessFrom)
		idx++
	}
	if filter.BusinessTo != nil {
		conds = append(conds, fmt.Sprintf("business_date <= $%d", idx))
		args = append(args, *filter.BusinessTo)
		idx++
	}
	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM investment__portfolio_transactions "+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting transactions: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if filter.Page > 1 {
		offset = (filter.Page - 1) * limit
	}
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx,
		txnSelect+" "+where+
			fmt.Sprintf(" ORDER BY business_date DESC, created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing transactions: %w", err)
	}
	defer rows.Close()

	out := make([]*entity.PortfolioTransaction, 0, limit)
	for rows.Next() {
		t, err := scanTxn(rows)
		if err != nil {
			return nil, 0, err
		}
		if t != nil {
			out = append(out, t)
		}
	}
	return out, total, rows.Err()
}

// HasReversal returns true when a REVERSAL row already exists for the
// supplied original. Backed by the partial unique index in the migration.
func (r *PostgresTransactionRepository) HasReversal(ctx context.Context, originalID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM investment__portfolio_transactions
			 WHERE reverses_transaction_id = $1
		)`,
		originalID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking existing reversal: %w", err)
	}
	return exists, nil
}

func (r *PostgresTransactionRepository) SumRealisedPnLBase(ctx context.Context, portfolioID uuid.UUID, asOf time.Time) (decimal.Decimal, error) {
	var total decimal.Decimal
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(realised_pnl_base), 0)
		  FROM investment__portfolio_transactions
		 WHERE portfolio_id = $1
		   AND business_date <= $2
		   AND status = 'POSTED'`,
		portfolioID, asOf,
	).Scan(&total)
	if err != nil {
		return decimal.Zero, fmt.Errorf("summing realised pnl: %w", err)
	}
	return total, nil
}

func scanTxn(s scanner) (*entity.PortfolioTransaction, error) {
	var t entity.PortfolioTransaction
	var typeStr, statusStr string
	var sideStr *string
	err := s.Scan(
		&t.ID, &t.PortfolioID, &t.FundID, &t.InstrumentID,
		&typeStr, &sideStr,
		&t.Quantity, &t.Price, &t.Currency, &t.GrossAmount, &t.Fees, &t.NetAmount, &t.RealisedPnLBase, &t.FxRateToBase,
		&t.BusinessDate, &t.SettlementDate,
		&t.SourceDecisionID, &t.SourceExecutionID, &t.ReversesTransactionID,
		&t.ExternalRef, &t.Reason, &statusStr,
		&t.CreatedAt, &t.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning transaction: %w", err)
	}
	t.TransactionType = vo.TransactionType(typeStr)
	t.Status = vo.TransactionStatus(statusStr)
	if sideStr != nil {
		s := vo.OrderSide(*sideStr)
		t.Side = &s
	}
	return &t, nil
}

var _ domain.PortfolioTransactionRepository = (*PostgresTransactionRepository)(nil)
