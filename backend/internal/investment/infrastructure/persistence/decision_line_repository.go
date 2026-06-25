package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PostgresDecisionLineRepository implements domain.DecisionLineRepository.
type PostgresDecisionLineRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresDecisionLineRepository wires the repo.
func NewPostgresDecisionLineRepository(pool *pgxpool.Pool) *PostgresDecisionLineRepository {
	return &PostgresDecisionLineRepository{pool: pool}
}

const lineSelect = `
	SELECT id, decision_id, line_number,
	       instrument_id, instrument_code, product_type,
	       side, quantity, amount, target_weight, limit_price, currency,
	       notes, created_at, created_by, updated_at, updated_by
	FROM investment__decision_lines`

func (r *PostgresDecisionLineRepository) ListByDecision(ctx context.Context, decisionID uuid.UUID) ([]*entity.DecisionLine, error) {
	rows, err := r.pool.Query(ctx, lineSelect+` WHERE decision_id = $1 ORDER BY line_number`, decisionID)
	if err != nil {
		return nil, fmt.Errorf("list decision lines: %w", err)
	}
	defer rows.Close()
	var out []*entity.DecisionLine
	for rows.Next() {
		l, err := scanLine(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *PostgresDecisionLineRepository) CreateLines(ctx context.Context, tx pgx.Tx, lines []*entity.DecisionLine) error {
	for _, l := range lines {
		if err := insertLine(ctx, tx, l); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresDecisionLineRepository) ReplaceLines(ctx context.Context, tx pgx.Tx, decisionID uuid.UUID, lines []*entity.DecisionLine) error {
	if _, err := tx.Exec(ctx, `DELETE FROM investment__decision_lines WHERE decision_id = $1`, decisionID); err != nil {
		return fmt.Errorf("delete decision lines: %w", err)
	}
	for _, l := range lines {
		if err := insertLine(ctx, tx, l); err != nil {
			return err
		}
	}
	return nil
}

func insertLine(ctx context.Context, tx pgx.Tx, l *entity.DecisionLine) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__decision_lines (
			id, decision_id, line_number,
			instrument_id, instrument_code, product_type,
			side, quantity, amount, target_weight, limit_price, currency,
			notes, created_at, created_by, updated_at, updated_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		l.ID, l.DecisionID, l.LineNumber,
		l.InstrumentID, l.InstrumentCode, string(l.ProductType),
		string(l.Side), l.Quantity, l.Amount, l.TargetWeight, l.LimitPrice, l.Currency,
		l.Notes, l.CreatedAt, l.CreatedBy, l.UpdatedAt, l.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert decision line %d: %w", l.LineNumber, err)
	}
	return nil
}

func scanLine(row rowScanner) (*entity.DecisionLine, error) {
	l := &entity.DecisionLine{}
	var (
		productType string
		side        string
		quantity, amount, targetWeight, limitPrice *decimal.Decimal
	)
	if err := row.Scan(
		&l.ID, &l.DecisionID, &l.LineNumber,
		&l.InstrumentID, &l.InstrumentCode, &productType,
		&side, &quantity, &amount, &targetWeight, &limitPrice, &l.Currency,
		&l.Notes, &l.CreatedAt, &l.CreatedBy, &l.UpdatedAt, &l.UpdatedBy,
	); err != nil {
		return nil, fmt.Errorf("scan decision line: %w", err)
	}
	l.ProductType = vo.DecisionProductType(productType)
	l.Side = vo.OrderSide(side)
	l.Quantity = quantity
	l.Amount = amount
	l.TargetWeight = targetWeight
	l.LimitPrice = limitPrice
	return l, nil
}

var _ domain.DecisionLineRepository = (*PostgresDecisionLineRepository)(nil)
