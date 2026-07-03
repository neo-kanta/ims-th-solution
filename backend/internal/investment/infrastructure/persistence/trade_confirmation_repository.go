package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

type PostgresTradeConfirmationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTradeConfirmationRepository(pool *pgxpool.Pool) *PostgresTradeConfirmationRepository {
	return &PostgresTradeConfirmationRepository{pool: pool}
}

const confirmationSelect = `
	SELECT id, execution_id, decision_id, fund_id, portfolio_id, contract_id, business_date,
	       confirmed_quantity, confirmed_amount, confirmed_price, currency,
	       broker_reference, import_batch_id,
	       status, discrepancy_reason,
	       reviewed_at, reviewed_by,
	       created_at, created_by, updated_at, updated_by
	FROM investment__trade_confirmations`

func (r *PostgresTradeConfirmationRepository) Create(ctx context.Context, tx pgx.Tx, c *entity.TradeConfirmation) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__trade_confirmations (
			id, execution_id, decision_id, fund_id, portfolio_id, contract_id, business_date,
			confirmed_quantity, confirmed_amount, confirmed_price, currency,
			broker_reference, import_batch_id,
			status, discrepancy_reason,
			reviewed_at, reviewed_by,
			created_at, created_by, updated_at, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13,
			$14, $15,
			$16, $17,
			$18, $19, $20, $21
		)`,
		c.ID, c.ExecutionID, c.DecisionID, c.FundID, c.PortfolioID, c.ContractID, c.BusinessDate,
		c.ConfirmedQuantity, c.ConfirmedAmount, c.ConfirmedPrice, c.Currency,
		c.BrokerReference, c.ImportBatchID,
		string(c.Status), c.DiscrepancyReason,
		c.ReviewedAt, c.ReviewedBy,
		c.CreatedAt, c.CreatedBy, c.UpdatedAt, c.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert trade confirmation: %w", err)
	}
	return nil
}

func (r *PostgresTradeConfirmationRepository) Update(ctx context.Context, tx pgx.Tx, c *entity.TradeConfirmation) error {
	_, err := tx.Exec(ctx, `
		UPDATE investment__trade_confirmations SET
			confirmed_quantity = $2, confirmed_amount = $3, confirmed_price = $4,
			broker_reference = $5, status = $6, discrepancy_reason = $7,
			reviewed_at = $8, reviewed_by = $9,
			updated_at = $10, updated_by = $11
		WHERE id = $1`,
		c.ID, c.ConfirmedQuantity, c.ConfirmedAmount, c.ConfirmedPrice,
		c.BrokerReference, string(c.Status), c.DiscrepancyReason,
		c.ReviewedAt, c.ReviewedBy,
		c.UpdatedAt, c.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("update trade confirmation: %w", err)
	}
	return nil
}

func (r *PostgresTradeConfirmationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.TradeConfirmation, error) {
	row := r.pool.QueryRow(ctx, confirmationSelect+` WHERE id = $1`, id)
	c, err := scanConfirmation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}

func (r *PostgresTradeConfirmationRepository) ListByExecution(ctx context.Context, executionID uuid.UUID) ([]*entity.TradeConfirmation, error) {
	rows, err := r.pool.Query(ctx, confirmationSelect+` WHERE execution_id = $1 ORDER BY created_at`, executionID)
	if err != nil {
		return nil, fmt.Errorf("list confirmations by execution: %w", err)
	}
	defer rows.Close()
	out := []*entity.TradeConfirmation{}
	for rows.Next() {
		c, err := scanConfirmation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *PostgresTradeConfirmationRepository) ListByContractDate(ctx context.Context, contractID uuid.UUID, businessDate time.Time) ([]*entity.TradeConfirmation, error) {
	rows, err := r.pool.Query(ctx,
		confirmationSelect+` WHERE contract_id = $1 AND business_date = $2 ORDER BY created_at`,
		contractID, businessDate)
	if err != nil {
		return nil, fmt.Errorf("list confirmations by contract: %w", err)
	}
	defer rows.Close()
	out := []*entity.TradeConfirmation{}
	for rows.Next() {
		c, err := scanConfirmation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetByBrokerReference returns the existing confirmation for a non-empty
// broker reference, or nil. Used by the batch importer to dedupe rows
// against confirmations already persisted by earlier batches.
//
// Empty broker references are not deduped — many brokers reuse blank fields
// for cash adjustments and IPOs, and a false dedupe hit there would silently
// drop a real confirmation.
func (r *PostgresTradeConfirmationRepository) GetByBrokerReference(ctx context.Context, brokerReference string) (*entity.TradeConfirmation, error) {
	if brokerReference == "" {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, confirmationSelect+` WHERE broker_reference = $1 ORDER BY created_at DESC LIMIT 1`, brokerReference)
	c, err := scanConfirmation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}

func scanConfirmation(row rowScanner) (*entity.TradeConfirmation, error) {
	c := &entity.TradeConfirmation{}
	var (
		status                      string
		confQty, confAmt, confPrice *decimal.Decimal
		broker                      string
		batch                       *uuid.UUID
		discrepancy                 string
		reviewedAt                  *time.Time
		reviewedBy                  *uuid.UUID
	)
	if err := row.Scan(
		&c.ID, &c.ExecutionID, &c.DecisionID, &c.FundID, &c.PortfolioID, &c.ContractID, &c.BusinessDate,
		&confQty, &confAmt, &confPrice, &c.Currency,
		&broker, &batch,
		&status, &discrepancy,
		&reviewedAt, &reviewedBy,
		&c.CreatedAt, &c.CreatedBy, &c.UpdatedAt, &c.UpdatedBy,
	); err != nil {
		return nil, err
	}
	c.Status = vo.TradeConfirmationStatus(status)
	c.ConfirmedQuantity = confQty
	c.ConfirmedAmount = confAmt
	c.ConfirmedPrice = confPrice
	c.BrokerReference = broker
	c.ImportBatchID = batch
	c.DiscrepancyReason = discrepancy
	c.ReviewedAt = reviewedAt
	c.ReviewedBy = reviewedBy
	return c, nil
}

var _ domain.TradeConfirmationRepository = (*PostgresTradeConfirmationRepository)(nil)
