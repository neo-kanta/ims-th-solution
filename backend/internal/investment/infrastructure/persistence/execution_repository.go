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

type PostgresExecutionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresExecutionRepository(pool *pgxpool.Pool) *PostgresExecutionRepository {
	return &PostgresExecutionRepository{pool: pool}
}

const executionSelect = `
	SELECT id, decision_id, fund_id, portfolio_id, contract_id,
	       instrument_id, instrument_code, business_date,
	       side, ordered_quantity, ordered_amount, executed_quantity, executed_amount, execution_price,
	       currency, status, trader_user_id, broker_reference,
	       executed_at, cancelled_at, cancelled_by, cancellation_reason,
	       created_at, created_by, updated_at, updated_by
	FROM investment__executions`

func (r *PostgresExecutionRepository) Create(ctx context.Context, tx pgx.Tx, e *entity.Execution) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__executions (
			id, decision_id, fund_id, portfolio_id, contract_id,
			instrument_id, instrument_code, business_date,
			side, ordered_quantity, ordered_amount, executed_quantity, executed_amount, execution_price,
			currency, status, trader_user_id, broker_reference,
			executed_at, cancelled_at, cancelled_by, cancellation_reason,
			created_at, created_by, updated_at, updated_by
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18,
			$19, $20, $21, $22,
			$23, $24, $25, $26
		)`,
		e.ID, e.DecisionID, e.FundID, e.PortfolioID, e.ContractID,
		e.InstrumentID, e.InstrumentCode, e.BusinessDate,
		string(e.Side), e.OrderedQuantity, e.OrderedAmount, e.ExecutedQuantity, e.ExecutedAmount, e.ExecutionPrice,
		e.Currency, string(e.Status), e.TraderUserID, e.BrokerReference,
		e.ExecutedAt, e.CancelledAt, e.CancelledBy, e.CancellationReason,
		e.CreatedAt, e.CreatedBy, e.UpdatedAt, e.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert execution: %w", err)
	}
	return nil
}

func (r *PostgresExecutionRepository) Update(ctx context.Context, tx pgx.Tx, e *entity.Execution) error {
	_, err := tx.Exec(ctx, `
		UPDATE investment__executions SET
			executed_quantity = $2, executed_amount = $3, execution_price = $4,
			status = $5, trader_user_id = $6, broker_reference = $7,
			executed_at = $8, cancelled_at = $9, cancelled_by = $10, cancellation_reason = $11,
			updated_at = $12, updated_by = $13
		WHERE id = $1`,
		e.ID, e.ExecutedQuantity, e.ExecutedAmount, e.ExecutionPrice,
		string(e.Status), e.TraderUserID, e.BrokerReference,
		e.ExecutedAt, e.CancelledAt, e.CancelledBy, e.CancellationReason,
		e.UpdatedAt, e.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("update execution: %w", err)
	}
	return nil
}

func (r *PostgresExecutionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Execution, error) {
	row := r.pool.QueryRow(ctx, executionSelect+` WHERE id = $1`, id)
	e, err := scanExecution(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return e, nil
}

func (r *PostgresExecutionRepository) ListByDecision(ctx context.Context, decisionID uuid.UUID) ([]*entity.Execution, error) {
	rows, err := r.pool.Query(ctx, executionSelect+` WHERE decision_id = $1 ORDER BY created_at`, decisionID)
	if err != nil {
		return nil, fmt.Errorf("list executions by decision: %w", err)
	}
	defer rows.Close()
	out := []*entity.Execution{}
	for rows.Next() {
		e, err := scanExecution(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *PostgresExecutionRepository) ListByContractDate(ctx context.Context, contractID uuid.UUID, businessDate time.Time) ([]*entity.Execution, error) {
	rows, err := r.pool.Query(ctx,
		executionSelect+` WHERE contract_id = $1 AND business_date = $2 ORDER BY created_at`,
		contractID, businessDate)
	if err != nil {
		return nil, fmt.Errorf("list executions by contract: %w", err)
	}
	defer rows.Close()
	out := []*entity.Execution{}
	for rows.Next() {
		e, err := scanExecution(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func scanExecution(row rowScanner) (*entity.Execution, error) {
	e := &entity.Execution{}
	var (
		side, status                                string
		instrument                                  *uuid.UUID
		ordQty, ordAmt, execQty, execAmt, execPrice *decimal.Decimal
		trader                                      *uuid.UUID
		broker                                      string
		executedAt                                  *time.Time
		cancelledAt                                 *time.Time
		cancelledBy                                 *uuid.UUID
		cancellationReason                          string
	)
	if err := row.Scan(
		&e.ID, &e.DecisionID, &e.FundID, &e.PortfolioID, &e.ContractID,
		&instrument, &e.InstrumentCode, &e.BusinessDate,
		&side, &ordQty, &ordAmt, &execQty, &execAmt, &execPrice,
		&e.Currency, &status, &trader, &broker,
		&executedAt, &cancelledAt, &cancelledBy, &cancellationReason,
		&e.CreatedAt, &e.CreatedBy, &e.UpdatedAt, &e.UpdatedBy,
	); err != nil {
		return nil, err
	}
	e.Side = vo.OrderSide(side)
	e.Status = vo.ExecutionStatus(status)
	e.InstrumentID = instrument
	e.OrderedQuantity = ordQty
	e.OrderedAmount = ordAmt
	e.ExecutedQuantity = execQty
	e.ExecutedAmount = execAmt
	e.ExecutionPrice = execPrice
	e.TraderUserID = trader
	e.BrokerReference = broker
	e.ExecutedAt = executedAt
	e.CancelledAt = cancelledAt
	e.CancelledBy = cancelledBy
	e.CancellationReason = cancellationReason
	return e, nil
}

var _ domain.ExecutionRepository = (*PostgresExecutionRepository)(nil)
