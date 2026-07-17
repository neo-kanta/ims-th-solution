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

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PostgresPortfolioRepository implements domain.PortfolioRepository.
type PostgresPortfolioRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPortfolioRepository wires the repo.
func NewPostgresPortfolioRepository(pool *pgxpool.Pool) *PostgresPortfolioRepository {
	return &PostgresPortfolioRepository{pool: pool}
}

const portfolioSelect = `
	SELECT id, fund_id, portfolio_type, code, name, COALESCE(description, ''),
	       base_currency, valuation_currency,
	       COALESCE(strategy_code, ''), style_id,
	       manager_user_id, COALESCE(benchmark, ''), COALESCE(risk_profile, ''),
	       inception_date, status, has_units, tax_lot_method, version,
	       created_at, updated_at, created_by, updated_by, deleted_at
	FROM investment__portfolios`

func (r *PostgresPortfolioRepository) Create(ctx context.Context, tx pgx.Tx, p *entity.Portfolio) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__portfolios (
			id, fund_id, portfolio_type, code, name, description,
			base_currency, valuation_currency, strategy_code, style_id,
			manager_user_id, benchmark, risk_profile, inception_date,
			status, has_units, tax_lot_method, version,
			created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, NULLIF($6,''),
			$7, $8, NULLIF($9,''), $10,
			$11, NULLIF($12,''), NULLIF($13,''), $14,
			$15, $16, $17, $18,
			$19, $20, $21, $22
		)`,
		p.ID, p.FundID, string(p.PortfolioType), p.Code, p.Name, p.Description,
		p.BaseCurrency, p.ValuationCurrency, p.StrategyCode, p.StyleID,
		p.ManagerUserID, p.Benchmark, string(p.RiskProfile), p.InceptionDate,
		string(p.Status), p.HasUnits, string(p.TaxLotMethod), p.Version,
		p.CreatedAt, p.UpdatedAt, p.CreatedBy, p.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting portfolio: %w", err)
	}
	return nil
}

func (r *PostgresPortfolioRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Portfolio, error) {
	row := r.pool.QueryRow(ctx, portfolioSelect+" WHERE id = $1 AND deleted_at IS NULL", id)
	return scanPortfolio(row)
}

func (r *PostgresPortfolioRepository) GetByFundCode(ctx context.Context, fundID uuid.UUID, code string) (*entity.Portfolio, error) {
	row := r.pool.QueryRow(ctx,
		portfolioSelect+" WHERE fund_id = $1 AND code = $2 AND deleted_at IS NULL",
		fundID, code,
	)
	return scanPortfolio(row)
}

// GetByCode fetches every alive row for the code (not just one) so it can
// detect the fund-scoped-uniqueness gap described on the interface doc
// comment: two alive portfolios under different funds may legally share a
// code today. Returning an arbitrary match in that case could resolve to
// the wrong portfolio — a data-access issue, not just a display nit — so
// this refuses to guess and returns *domain.ErrAmbiguousPortfolioCode
// instead.
func (r *PostgresPortfolioRepository) GetByCode(ctx context.Context, code string) (*entity.Portfolio, error) {
	rows, err := r.pool.Query(ctx,
		portfolioSelect+" WHERE code = $1 AND deleted_at IS NULL",
		code,
	)
	if err != nil {
		return nil, fmt.Errorf("querying portfolio by code: %w", err)
	}
	defer rows.Close()

	var matches []*entity.Portfolio
	for rows.Next() {
		p, err := scanPortfolio(rows)
		if err != nil {
			return nil, err
		}
		if p != nil {
			matches = append(matches, p)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanning portfolios by code: %w", err)
	}

	switch len(matches) {
	case 0:
		return nil, nil
	case 1:
		return matches[0], nil
	default:
		return nil, &domain.ErrAmbiguousPortfolioCode{Code: code, Count: len(matches)}
	}
}

func (r *PostgresPortfolioRepository) List(ctx context.Context, filter domain.PortfolioListFilter) ([]*entity.Portfolio, int, error) {
	conds := []string{"deleted_at IS NULL"}
	args := []any{}
	idx := 1
	if filter.FundID != nil {
		conds = append(conds, fmt.Sprintf("fund_id = $%d", idx))
		args = append(args, *filter.FundID)
		idx++
	}
	if filter.Status != nil {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(*filter.Status))
		idx++
	}
	if filter.ManagerUserID != nil {
		conds = append(conds, fmt.Sprintf("manager_user_id = $%d", idx))
		args = append(args, *filter.ManagerUserID)
		idx++
	}
	if filter.AccessibleFundIDs != nil {
		// Empty slice means "no access" — return zero rows.
		if len(filter.AccessibleFundIDs) == 0 {
			return []*entity.Portfolio{}, 0, nil
		}
		conds = append(conds, fmt.Sprintf("fund_id = ANY($%d)", idx))
		args = append(args, filter.AccessibleFundIDs)
		idx++
	}
	where := "WHERE " + strings.Join(conds, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM investment__portfolios "+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting portfolios: %w", err)
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
		portfolioSelect+" "+where+
			fmt.Sprintf(" ORDER BY fund_id, code LIMIT $%d OFFSET $%d", idx, idx+1),
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing portfolios: %w", err)
	}
	defer rows.Close()

	out := make([]*entity.Portfolio, 0, limit)
	for rows.Next() {
		p, err := scanPortfolio(rows)
		if err != nil {
			return nil, 0, err
		}
		if p != nil {
			out = append(out, p)
		}
	}
	return out, total, rows.Err()
}

func (r *PostgresPortfolioRepository) Update(ctx context.Context, tx pgx.Tx, p *entity.Portfolio) error {
	tag, err := tx.Exec(ctx, `
		UPDATE investment__portfolios
		   SET name              = $3,
		       description       = NULLIF($4,''),
		       strategy_code     = NULLIF($5,''),
		       style_id          = $6,
		       manager_user_id   = $7,
		       benchmark         = NULLIF($8,''),
		       risk_profile      = NULLIF($9,''),
		       status            = $10,
		       version           = $11,
		       updated_at        = $12,
		       updated_by        = $13
		 WHERE id = $1 AND version = $2 AND deleted_at IS NULL`,
		p.ID, p.Version-1,
		p.Name, p.Description, p.StrategyCode, p.StyleID,
		p.ManagerUserID, p.Benchmark, string(p.RiskProfile),
		string(p.Status), p.Version, p.UpdatedAt, p.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("updating portfolio: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ErrPortfolioVersionMismatch{
			PortfolioID: p.ID.String(), ExpectedVersion: p.Version - 1, ActualVersion: -1,
		}
	}
	return nil
}

func (r *PostgresPortfolioRepository) SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, expectedVersion int, deletedBy uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
		UPDATE investment__portfolios
		   SET deleted_at = NOW(),
		       updated_at = NOW(),
		       updated_by = $3,
		       version    = version + 1
		 WHERE id = $1 AND version = $2 AND deleted_at IS NULL`,
		id, expectedVersion, deletedBy,
	)
	if err != nil {
		return fmt.Errorf("soft-deleting portfolio: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return &domain.ErrPortfolioVersionMismatch{
			PortfolioID: id.String(), ExpectedVersion: expectedVersion, ActualVersion: -1,
		}
	}
	return nil
}

// HasOpenActivity refuses portfolio deletion when any non-zero position or
// any same-day transaction exists.
func (r *PostgresPortfolioRepository) HasOpenActivity(ctx context.Context, portfolioID uuid.UUID, asOf time.Time) (bool, string, error) {
	var nzPos int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM investment__portfolio_positions
		 WHERE portfolio_id = $1 AND quantity <> 0`,
		portfolioID,
	).Scan(&nzPos); err != nil {
		return false, "", fmt.Errorf("checking non-zero positions: %w", err)
	}
	if nzPos > 0 {
		return true, fmt.Sprintf("%d non-zero position(s) remain", nzPos), nil
	}

	var sameDayTxns int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM investment__portfolio_transactions
		 WHERE portfolio_id = $1 AND business_date = $2`,
		portfolioID, asOf,
	).Scan(&sameDayTxns); err != nil {
		return false, "", fmt.Errorf("checking same-day transactions: %w", err)
	}
	if sameDayTxns > 0 {
		return true, fmt.Sprintf("%d transaction(s) on the current business date", sameDayTxns), nil
	}

	return false, "", nil
}

func scanPortfolio(s scanner) (*entity.Portfolio, error) {
	var p entity.Portfolio
	var typeStr, statusStr, riskStr, taxStr string
	err := s.Scan(
		&p.ID, &p.FundID, &typeStr, &p.Code, &p.Name, &p.Description,
		&p.BaseCurrency, &p.ValuationCurrency,
		&p.StrategyCode, &p.StyleID,
		&p.ManagerUserID, &p.Benchmark, &riskStr,
		&p.InceptionDate, &statusStr, &p.HasUnits, &taxStr, &p.Version,
		&p.CreatedAt, &p.UpdatedAt, &p.CreatedBy, &p.UpdatedBy, &p.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning portfolio: %w", err)
	}
	p.PortfolioType = vo.PortfolioType(typeStr)
	p.Status = vo.PortfolioStatus(statusStr)
	p.RiskProfile = vo.RiskProfile(riskStr)
	p.TaxLotMethod = vo.TaxLotMethod(taxStr)
	return &p, nil
}

var _ domain.PortfolioRepository = (*PostgresPortfolioRepository)(nil)
