package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// InvestmentPositionSnapshotAdapter reads current investment positions and
// instrument identifiers for compliance exposure checks.
type InvestmentPositionSnapshotAdapter struct {
	pool *pgxpool.Pool
}

func NewInvestmentPositionSnapshotAdapter(pool *pgxpool.Pool) *InvestmentPositionSnapshotAdapter {
	return &InvestmentPositionSnapshotAdapter{pool: pool}
}

func (a *InvestmentPositionSnapshotAdapter) GetSnapshot(
	ctx context.Context,
	portfolioID uuid.UUID,
	asOf time.Time,
) (*spi.PositionSnapshot, error) {
	if a == nil || a.pool == nil {
		return nil, fmt.Errorf("investment position adapter not initialised")
	}

	rows, err := a.pool.Query(ctx, `
		SELECT
			i.primary_ticker,
			COALESCE(isin.id_value, '') AS isin,
			p.quantity,
			p.cost_basis,
			COALESCE(p.quantity * lp.price, p.cost_basis) AS market_value
		FROM investment__portfolio_positions p
		JOIN investment__instruments i
			ON i.id = p.instrument_id
		LEFT JOIN LATERAL (
			SELECT id_value
			FROM investment__instrument_identifiers
			WHERE instrument_id = i.id
			  AND id_type = 'ISIN'
			ORDER BY is_primary DESC, created_at DESC
			LIMIT 1
		) isin ON true
		LEFT JOIN LATERAL (
			SELECT price
			FROM investment__price_snapshots
			WHERE instrument_id = p.instrument_id
			  AND business_date <= $2
			ORDER BY business_date DESC, captured_at DESC
			LIMIT 1
		) lp ON true
		WHERE p.portfolio_id = $1
		  AND i.deleted_at IS NULL
		ORDER BY i.primary_ticker`,
		portfolioID, asOf,
	)
	if err != nil {
		return nil, fmt.Errorf("querying investment positions: %w", err)
	}
	defer rows.Close()

	snap := &spi.PositionSnapshot{
		PortfolioID: portfolioID,
		AsOf:        asOf,
		Holdings:    []spi.Holding{},
	}
	for rows.Next() {
		var h spi.Holding
		if err := rows.Scan(&h.Ticker, &h.ISIN, &h.Quantity, &h.CostBasis, &h.MarketValue); err != nil {
			return nil, fmt.Errorf("scanning investment position: %w", err)
		}
		snap.Holdings = append(snap.Holdings, h)
	}
	return snap, rows.Err()
}

// InvestmentMarketDataAdapter reads current cash balances, latest valuation
// snapshots, and instrument prices from investment tables.
type InvestmentMarketDataAdapter struct {
	pool *pgxpool.Pool
}

func NewInvestmentMarketDataAdapter(pool *pgxpool.Pool) *InvestmentMarketDataAdapter {
	return &InvestmentMarketDataAdapter{pool: pool}
}

func (a *InvestmentMarketDataAdapter) GetNAV(
	ctx context.Context,
	portfolioID uuid.UUID,
	asOf time.Time,
) (*spi.NAVSnapshot, error) {
	return a.getNAV(ctx, portfolioID, asOf, "")
}

func (a *InvestmentMarketDataAdapter) GetNAVForCheck(
	ctx context.Context,
	input spi.CheckInput,
) (*spi.NAVSnapshot, error) {
	currency := ""
	if input.ProposedOrder != nil {
		currency = input.ProposedOrder.Currency
	}
	return a.getNAV(ctx, input.PortfolioID, input.BusinessDate, currency)
}

func (a *InvestmentMarketDataAdapter) getNAV(
	ctx context.Context,
	portfolioID uuid.UUID,
	asOf time.Time,
	cashCurrency string,
) (*spi.NAVSnapshot, error) {
	if a == nil || a.pool == nil {
		return nil, fmt.Errorf("investment market data adapter not initialised")
	}

	baseCurrency := "THB"
	if err := a.pool.QueryRow(ctx, `
		SELECT base_currency
		FROM investment__portfolios
		WHERE id = $1
		  AND deleted_at IS NULL`,
		portfolioID,
	).Scan(&baseCurrency); err != nil {
		if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("querying portfolio currency: %w", err)
		}
	}
	if cashCurrency == "" {
		cashCurrency = baseCurrency
	}

	cashBalance := decimal.Zero
	if err := a.pool.QueryRow(ctx, `
		SELECT balance
		FROM investment__cash_balances
		WHERE portfolio_id = $1
		  AND currency = $2`,
		portfolioID, cashCurrency,
	).Scan(&cashBalance); err != nil {
		if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("querying cash balance: %w", err)
		}
	}

	nav := decimal.Zero
	if err := a.pool.QueryRow(ctx, `
		SELECT aum
		FROM investment__valuation_snapshots
		WHERE portfolio_id = $1
		  AND business_date <= $2
		ORDER BY business_date DESC,
		         CASE source WHEN 'INTERNAL' THEN 0 WHEN 'PAM' THEN 1 ELSE 2 END,
		         created_at DESC
		LIMIT 1`,
		portfolioID, asOf,
	).Scan(&nav); err != nil {
		if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("querying valuation snapshot: %w", err)
		}
	}

	if nav.IsZero() {
		positionMV := decimal.Zero
		if err := a.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(p.quantity * COALESCE(lp.price, p.average_cost)), 0)
			FROM investment__portfolio_positions p
			LEFT JOIN LATERAL (
				SELECT price
				FROM investment__price_snapshots
				WHERE instrument_id = p.instrument_id
				  AND business_date <= $2
				ORDER BY business_date DESC, captured_at DESC
				LIMIT 1
			) lp ON true
			WHERE p.portfolio_id = $1`,
			portfolioID, asOf,
		).Scan(&positionMV); err != nil {
			return nil, fmt.Errorf("estimating portfolio market value: %w", err)
		}
		nav = cashBalance.Add(positionMV)
	}

	return &spi.NAVSnapshot{
		PortfolioID:  portfolioID,
		AsOf:         asOf,
		NAV:          nav,
		CashBalance:  cashBalance,
		ReservedCash: decimal.Zero,
	}, nil
}

func (a *InvestmentMarketDataAdapter) GetPrices(
	ctx context.Context,
	tickers []string,
	asOf time.Time,
) (*spi.MarketPriceSnapshot, error) {
	if a == nil || a.pool == nil {
		return nil, fmt.Errorf("investment market data adapter not initialised")
	}
	if len(tickers) == 0 {
		return &spi.MarketPriceSnapshot{Prices: map[string]decimal.Decimal{}, AsOf: asOf}, nil
	}

	rows, err := a.pool.Query(ctx, `
		SELECT DISTINCT ON (i.primary_ticker)
			i.primary_ticker,
			ps.price
		FROM investment__instruments i
		JOIN investment__price_snapshots ps
			ON ps.instrument_id = i.id
		WHERE i.primary_ticker = ANY($1::text[])
		  AND i.deleted_at IS NULL
		  AND ps.business_date <= $2
		ORDER BY i.primary_ticker, ps.business_date DESC, ps.captured_at DESC`,
		uniqueStrings(tickers), asOf,
	)
	if err != nil {
		return nil, fmt.Errorf("querying investment prices: %w", err)
	}
	defer rows.Close()

	prices := make(map[string]decimal.Decimal, len(tickers))
	for rows.Next() {
		var ticker string
		var price decimal.Decimal
		if err := rows.Scan(&ticker, &price); err != nil {
			return nil, fmt.Errorf("scanning investment price: %w", err)
		}
		prices[ticker] = price
	}
	return &spi.MarketPriceSnapshot{Prices: prices, AsOf: asOf}, rows.Err()
}

func (a *InvestmentMarketDataAdapter) GetFXRates(
	ctx context.Context,
	baseCurrency string,
	asOf time.Time,
) (*spi.FXRateSnapshot, error) {
	return &spi.FXRateSnapshot{
		BaseCurrency: baseCurrency,
		Rates:        map[string]decimal.Decimal{baseCurrency: decimal.NewFromInt(1)},
		AsOf:         asOf,
	}, nil
}

// InvestmentInstrumentClassificationAdapter maps tickers to taxonomy codes
// from investment instrument master data.
type InvestmentInstrumentClassificationAdapter struct {
	pool *pgxpool.Pool
}

func NewInvestmentInstrumentClassificationAdapter(pool *pgxpool.Pool) *InvestmentInstrumentClassificationAdapter {
	return &InvestmentInstrumentClassificationAdapter{pool: pool}
}

func (a *InvestmentInstrumentClassificationAdapter) GetClassifications(
	ctx context.Context,
	tickers []string,
) (*spi.ClassificationSnapshot, error) {
	if a == nil || a.pool == nil {
		return nil, fmt.Errorf("investment classification adapter not initialised")
	}
	if len(tickers) == 0 {
		return spi.NewClassificationSnapshot(nil), nil
	}

	rows, err := a.pool.Query(ctx, `
		SELECT
			i.primary_ticker,
			COALESCE(isin.id_value, '') AS isin,
			COALESCE(NULLIF(i.attributes->>'issuer', ''), i.primary_ticker) AS issuer,
			COALESCE(NULLIF(i.attributes->>'parent_entity', ''), NULLIF(i.attributes->>'issuer', ''), i.primary_ticker) AS parent_entity,
			COALESCE(s.code, '') AS sector,
			COALESCE(ac.code, '') AS asset_class,
			COALESCE(c.iso_code, '') AS country,
			COALESCE(i.primary_exchange, '') AS exchange,
			COALESCE((i.attributes->>'is_government')::boolean, false) AS is_government
		FROM investment__instruments i
		JOIN investment__asset_classes ac ON ac.id = i.asset_class_id
		JOIN investment__countries c ON c.id = i.country_id
		LEFT JOIN investment__sectors s ON s.id = i.sector_id
		LEFT JOIN LATERAL (
			SELECT id_value
			FROM investment__instrument_identifiers
			WHERE instrument_id = i.id
			  AND id_type = 'ISIN'
			ORDER BY is_primary DESC, created_at DESC
			LIMIT 1
		) isin ON true
		WHERE i.primary_ticker = ANY($1::text[])
		  AND i.deleted_at IS NULL`,
		uniqueStrings(tickers),
	)
	if err != nil {
		return nil, fmt.Errorf("querying investment classifications: %w", err)
	}
	defer rows.Close()

	items := []spi.InstrumentClassification{}
	for rows.Next() {
		var item spi.InstrumentClassification
		if err := rows.Scan(
			&item.Ticker,
			&item.ISIN,
			&item.Issuer,
			&item.ParentEntity,
			&item.Sector,
			&item.AssetClass,
			&item.Country,
			&item.Exchange,
			&item.IsGovernment,
		); err != nil {
			return nil, fmt.Errorf("scanning investment classification: %w", err)
		}
		items = append(items, item)
	}
	return spi.NewClassificationSnapshot(items), rows.Err()
}

// InvestmentPortfolioMetadataAdapter reads portfolio-level metadata used by
// compliance rules that depend on base currency or contract identity.
type InvestmentPortfolioMetadataAdapter struct {
	pool *pgxpool.Pool
}

func NewInvestmentPortfolioMetadataAdapter(pool *pgxpool.Pool) *InvestmentPortfolioMetadataAdapter {
	return &InvestmentPortfolioMetadataAdapter{pool: pool}
}

func (a *InvestmentPortfolioMetadataAdapter) GetMetadata(
	ctx context.Context,
	portfolioID uuid.UUID,
) (*spi.PortfolioMetadata, error) {
	if a == nil || a.pool == nil {
		return nil, fmt.Errorf("investment portfolio metadata adapter not initialised")
	}

	meta := &spi.PortfolioMetadata{PortfolioID: portfolioID, Jurisdiction: "TH"}
	err := a.pool.QueryRow(ctx, `
		SELECT p.fund_id, p.base_currency, p.inception_date, COALESCE(fc.code, '')
		FROM investment__portfolios p
		LEFT JOIN investment__funds f ON f.id = p.fund_id
		LEFT JOIN investment__fund_categories fc ON fc.id = f.fund_category_id
		WHERE p.id = $1
		  AND p.deleted_at IS NULL`,
		portfolioID,
	).Scan(&meta.ContractID, &meta.BaseCurrency, &meta.InceptionDate, &meta.MandateType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("portfolio %s not found", portfolioID)
		}
		return nil, fmt.Errorf("querying portfolio metadata: %w", err)
	}
	return meta, nil
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
