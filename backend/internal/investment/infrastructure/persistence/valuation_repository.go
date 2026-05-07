package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PostgresValuationRepository implements domain.ValuationRepository.
//
// All snapshots persisted by this repo are append-only at the storage layer.
type PostgresValuationRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresValuationRepository wires the repo.
func NewPostgresValuationRepository(pool *pgxpool.Pool) *PostgresValuationRepository {
	return &PostgresValuationRepository{pool: pool}
}

const valuationSelect = `
	SELECT id, portfolio_id, business_date, valuation_ccy,
	       market_value, cost_basis, unrealised_pnl, realised_pnl, roi,
	       aum, cash_balance,
	       price_set_hash, has_stale_inputs, is_indicative, source,
	       created_at, created_by
	FROM investment__valuation_snapshots`

// Insert persists the parent valuation row and all holding-line children.
func (r *PostgresValuationRepository) Insert(ctx context.Context, tx pgx.Tx, snap *entity.ValuationSnapshot) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__valuation_snapshots (
			id, portfolio_id, business_date, valuation_ccy,
			market_value, cost_basis, unrealised_pnl, realised_pnl, roi,
			aum, cash_balance,
			price_set_hash, has_stale_inputs, is_indicative, source,
			created_at, created_by
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11,
			$12, $13, $14, $15,
			$16, $17
		)`,
		snap.ID, snap.PortfolioID, snap.BusinessDate, snap.ValuationCcy,
		snap.MarketValue, snap.CostBasis, snap.UnrealisedPnL, snap.RealisedPnL, snap.ROI,
		snap.AUM, snap.CashBalance,
		snap.PriceSetHash, snap.HasStaleInputs, snap.IsIndicative, string(snap.Source),
		snap.CreatedAt, snap.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting valuation snapshot: %w", err)
	}

	for i := range snap.HoldingLines {
		ln := &snap.HoldingLines[i]
		_, err := tx.Exec(ctx, `
			INSERT INTO investment__valuation_holding_lines (
				id, valuation_snapshot_id, instrument_id, price_snapshot_id,
				quantity, price_in_quote_ccy, quote_currency, fx_rate_to_valuation_ccy,
				market_value, cost_basis, unrealised_pnl, is_stale,
				created_at
			) VALUES (
				$1, $2, $3, $4,
				$5, $6, $7, $8,
				$9, $10, $11, $12,
				$13
			)`,
			ln.ID, ln.ValuationSnapshotID, ln.InstrumentID, ln.PriceSnapshotID,
			ln.Quantity, ln.PriceInQuoteCcy, ln.QuoteCurrency, ln.FxRateToValuationCcy,
			ln.MarketValue, ln.CostBasis, ln.UnrealisedPnL, ln.IsStale,
			ln.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("inserting valuation holding line: %w", err)
		}
	}

	return nil
}

func (r *PostgresValuationRepository) GetLatest(ctx context.Context, portfolioID uuid.UUID, source vo.ValuationSource) (*entity.ValuationSnapshot, error) {
	row := r.pool.QueryRow(ctx,
		valuationSelect+`
		 WHERE portfolio_id = $1 AND source = $2
		 ORDER BY business_date DESC, created_at DESC LIMIT 1`,
		portfolioID, string(source),
	)
	v, err := scanValuation(row)
	if err != nil || v == nil {
		return v, err
	}
	return v, r.loadHoldingLines(ctx, v)
}

func (r *PostgresValuationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ValuationSnapshot, error) {
	row := r.pool.QueryRow(ctx, valuationSelect+" WHERE id = $1", id)
	v, err := scanValuation(row)
	if err != nil || v == nil {
		return v, err
	}
	return v, r.loadHoldingLines(ctx, v)
}

// GetByPortfolioBusinessDate returns the valuation snapshot for the exact
// (portfolio_id, business_date, source) tuple, or nil-no-error when missing.
// Used by ComputeFundAUMHandler to assert all portfolios under a fund have a
// snapshot for the same business date.
func (r *PostgresValuationRepository) GetByPortfolioBusinessDate(
	ctx context.Context,
	portfolioID uuid.UUID,
	businessDate time.Time,
	source vo.ValuationSource,
) (*entity.ValuationSnapshot, error) {
	row := r.pool.QueryRow(ctx,
		valuationSelect+`
		 WHERE portfolio_id = $1 AND business_date = $2 AND source = $3
		 ORDER BY created_at DESC LIMIT 1`,
		portfolioID, businessDate, string(source),
	)
	v, err := scanValuation(row)
	if err != nil || v == nil {
		return v, err
	}
	return v, r.loadHoldingLines(ctx, v)
}

func (r *PostgresValuationRepository) List(ctx context.Context, portfolioID uuid.UUID, from, to time.Time, page, limit int) ([]*entity.ValuationSnapshot, int, error) {
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}

	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM investment__valuation_snapshots
		 WHERE portfolio_id = $1 AND business_date BETWEEN $2 AND $3`,
		portfolioID, from, to,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting valuations: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		valuationSelect+`
		 WHERE portfolio_id = $1 AND business_date BETWEEN $2 AND $3
		 ORDER BY business_date DESC LIMIT $4 OFFSET $5`,
		portfolioID, from, to, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing valuations: %w", err)
	}
	defer rows.Close()

	out := []*entity.ValuationSnapshot{}
	for rows.Next() {
		v, err := scanValuation(rows)
		if err != nil {
			return nil, 0, err
		}
		if v != nil {
			out = append(out, v)
		}
	}
	return out, total, rows.Err()
}

func (r *PostgresValuationRepository) loadHoldingLines(ctx context.Context, v *entity.ValuationSnapshot) error {
	rows, err := r.pool.Query(ctx, `
		SELECT id, valuation_snapshot_id, instrument_id, price_snapshot_id,
		       quantity, price_in_quote_ccy, quote_currency, fx_rate_to_valuation_ccy,
		       market_value, cost_basis, unrealised_pnl, is_stale,
		       created_at
		  FROM investment__valuation_holding_lines
		 WHERE valuation_snapshot_id = $1
		 ORDER BY instrument_id`,
		v.ID,
	)
	if err != nil {
		return fmt.Errorf("loading valuation holding lines: %w", err)
	}
	defer rows.Close()

	v.HoldingLines = nil
	for rows.Next() {
		var ln entity.ValuationHoldingLine
		if err := rows.Scan(
			&ln.ID, &ln.ValuationSnapshotID, &ln.InstrumentID, &ln.PriceSnapshotID,
			&ln.Quantity, &ln.PriceInQuoteCcy, &ln.QuoteCurrency, &ln.FxRateToValuationCcy,
			&ln.MarketValue, &ln.CostBasis, &ln.UnrealisedPnL, &ln.IsStale,
			&ln.CreatedAt,
		); err != nil {
			return fmt.Errorf("scanning valuation holding line: %w", err)
		}
		v.HoldingLines = append(v.HoldingLines, ln)
	}
	return rows.Err()
}

func (r *PostgresValuationRepository) InsertNAV(ctx context.Context, tx pgx.Tx, nav *entity.NAVSnapshot) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__nav_snapshots (
			id, portfolio_id, business_date,
			total_units, nav_per_unit,
			valuation_snapshot_id, is_indicative,
			created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		nav.ID, nav.PortfolioID, nav.BusinessDate,
		nav.TotalUnits, nav.NAVPerUnit,
		nav.ValuationSnapshotID, nav.IsIndicative,
		nav.CreatedAt, nav.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting nav snapshot: %w", err)
	}
	return nil
}

const navSelect = `
	SELECT id, portfolio_id, business_date, total_units, nav_per_unit,
	       valuation_snapshot_id, is_indicative, created_at, created_by
	FROM investment__nav_snapshots`

func (r *PostgresValuationRepository) ListNAV(ctx context.Context, portfolioID uuid.UUID, from, to time.Time, page, limit int) ([]*entity.NAVSnapshot, int, error) {
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}

	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM investment__nav_snapshots
		 WHERE portfolio_id = $1 AND business_date BETWEEN $2 AND $3`,
		portfolioID, from, to,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting navs: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		navSelect+`
		 WHERE portfolio_id = $1 AND business_date BETWEEN $2 AND $3
		 ORDER BY business_date DESC LIMIT $4 OFFSET $5`,
		portfolioID, from, to, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing navs: %w", err)
	}
	defer rows.Close()

	out := []*entity.NAVSnapshot{}
	for rows.Next() {
		var n entity.NAVSnapshot
		if err := rows.Scan(
			&n.ID, &n.PortfolioID, &n.BusinessDate, &n.TotalUnits, &n.NAVPerUnit,
			&n.ValuationSnapshotID, &n.IsIndicative, &n.CreatedAt, &n.CreatedBy,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning nav: %w", err)
		}
		out = append(out, &n)
	}
	return out, total, rows.Err()
}

func (r *PostgresValuationRepository) InsertAUM(ctx context.Context, tx pgx.Tx, aum *entity.AUMSnapshot) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO investment__aum_snapshots (
			id, scope_type, scope_id, business_date,
			aum, valuation_ccy, source,
			created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		aum.ID, string(aum.ScopeType), aum.ScopeID, aum.BusinessDate,
		aum.AUM, aum.ValuationCcy, string(aum.Source),
		aum.CreatedAt, aum.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting aum snapshot: %w", err)
	}
	return nil
}

const aumSelect = `
	SELECT id, scope_type, scope_id, business_date,
	       aum, valuation_ccy, source,
	       created_at, created_by
	FROM investment__aum_snapshots`

// GetAUM returns the AUM snapshot for an exact (scope, business_date, source)
// tuple, or nil-no-error when missing. The fund-AUM compute handler uses this
// for idempotency: if a snapshot already exists for today, return it as-is.
func (r *PostgresValuationRepository) GetAUM(
	ctx context.Context,
	scopeType vo.AumScopeType,
	scopeID uuid.UUID,
	businessDate time.Time,
	source vo.ValuationSource,
) (*entity.AUMSnapshot, error) {
	row := r.pool.QueryRow(ctx,
		aumSelect+`
		 WHERE scope_type = $1 AND scope_id = $2 AND business_date = $3 AND source = $4
		 LIMIT 1`,
		string(scopeType), scopeID, businessDate, string(source),
	)
	var a entity.AUMSnapshot
	var scopeStr, sourceStr string
	err := row.Scan(
		&a.ID, &scopeStr, &a.ScopeID, &a.BusinessDate,
		&a.AUM, &a.ValuationCcy, &sourceStr,
		&a.CreatedAt, &a.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning aum: %w", err)
	}
	a.ScopeType = vo.AumScopeType(scopeStr)
	a.Source = vo.ValuationSource(sourceStr)
	return &a, nil
}

func (r *PostgresValuationRepository) ListAUM(ctx context.Context, scopeType vo.AumScopeType, scopeID uuid.UUID, from, to time.Time, page, limit int) ([]*entity.AUMSnapshot, int, error) {
	if limit <= 0 {
		limit = 50
	}
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}

	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM investment__aum_snapshots
		 WHERE scope_type = $1 AND scope_id = $2 AND business_date BETWEEN $3 AND $4`,
		string(scopeType), scopeID, from, to,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting aum: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		aumSelect+`
		 WHERE scope_type = $1 AND scope_id = $2 AND business_date BETWEEN $3 AND $4
		 ORDER BY business_date DESC LIMIT $5 OFFSET $6`,
		string(scopeType), scopeID, from, to, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listing aum: %w", err)
	}
	defer rows.Close()

	out := []*entity.AUMSnapshot{}
	for rows.Next() {
		var a entity.AUMSnapshot
		var scopeStr, sourceStr string
		if err := rows.Scan(
			&a.ID, &scopeStr, &a.ScopeID, &a.BusinessDate,
			&a.AUM, &a.ValuationCcy, &sourceStr,
			&a.CreatedAt, &a.CreatedBy,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning aum: %w", err)
		}
		a.ScopeType = vo.AumScopeType(scopeStr)
		a.Source = vo.ValuationSource(sourceStr)
		out = append(out, &a)
	}
	return out, total, rows.Err()
}

func scanValuation(s scanner) (*entity.ValuationSnapshot, error) {
	var v entity.ValuationSnapshot
	var sourceStr string
	err := s.Scan(
		&v.ID, &v.PortfolioID, &v.BusinessDate, &v.ValuationCcy,
		&v.MarketValue, &v.CostBasis, &v.UnrealisedPnL, &v.RealisedPnL, &v.ROI,
		&v.AUM, &v.CashBalance,
		&v.PriceSetHash, &v.HasStaleInputs, &v.IsIndicative, &sourceStr,
		&v.CreatedAt, &v.CreatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning valuation: %w", err)
	}
	v.Source = vo.ValuationSource(sourceStr)
	return &v, nil
}

var _ domain.ValuationRepository = (*PostgresValuationRepository)(nil)
