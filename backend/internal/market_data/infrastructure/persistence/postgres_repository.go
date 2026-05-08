package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) UpsertSymbol(ctx context.Context, mapping domain.SymbolMapping) error {
	if r == nil || r.pool == nil {
		return nil
	}
	symbol := normalizeSymbol(mapping.Symbol)
	if symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	assetType := mapping.AssetType
	if assetType == "" {
		assetType = domain.AssetTypeUnknown
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO market_symbols (
			symbol, asset_type, name, currency, alpha_vantage_symbol, yahoo_symbol
		) VALUES (
			$1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, '')
		)
		ON CONFLICT (symbol) DO UPDATE
		   SET asset_type = CASE
		           WHEN EXCLUDED.asset_type <> 'UNKNOWN' THEN EXCLUDED.asset_type
		           ELSE market_symbols.asset_type
		       END,
		       name = COALESCE(EXCLUDED.name, market_symbols.name),
		       currency = COALESCE(EXCLUDED.currency, market_symbols.currency),
		       alpha_vantage_symbol = COALESCE(EXCLUDED.alpha_vantage_symbol, market_symbols.alpha_vantage_symbol),
		       yahoo_symbol = COALESCE(EXCLUDED.yahoo_symbol, market_symbols.yahoo_symbol),
		       updated_at = NOW()`,
		symbol,
		string(assetType),
		strings.TrimSpace(mapping.Name),
		strings.ToUpper(strings.TrimSpace(mapping.Currency)),
		strings.TrimSpace(mapping.AlphaVantageSymbol),
		strings.TrimSpace(mapping.YahooFinanceSymbol),
	)
	if err != nil {
		return fmt.Errorf("upserting market symbol: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetSymbolMapping(ctx context.Context, symbol string) (*domain.SymbolMapping, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, `
		SELECT symbol, asset_type, COALESCE(name, ''), COALESCE(currency, ''),
		       COALESCE(alpha_vantage_symbol, ''), COALESCE(yahoo_symbol, '')
		  FROM market_symbols
		 WHERE symbol = $1 AND is_active = true`,
		normalizeSymbol(symbol),
	)
	var mapping domain.SymbolMapping
	var assetType string
	if err := row.Scan(
		&mapping.Symbol,
		&assetType,
		&mapping.Name,
		&mapping.Currency,
		&mapping.AlphaVantageSymbol,
		&mapping.YahooFinanceSymbol,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("loading market symbol mapping: %w", err)
	}
	mapping.AssetType = domain.AssetType(assetType)
	return &mapping, nil
}

func (r *PostgresRepository) SaveQuote(ctx context.Context, quote domain.Quote) error {
	if r == nil || r.pool == nil {
		return nil
	}
	symbolID, err := r.ensureSymbolID(ctx, quote.Symbol)
	if err != nil {
		return err
	}
	asOf := quote.AsOf
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	capturedAt := quote.CapturedAt
	if capturedAt.IsZero() {
		capturedAt = time.Now().UTC()
	}
	raw, _ := json.Marshal(quote)
	_, err = r.pool.Exec(ctx, `
		INSERT INTO market_data_snapshots (
			symbol_id, provider_name, data_type, price_date, as_of,
			price, open_price, high_price, low_price, close_price, previous_close,
			change_amount, change_percent, volume, currency, raw_payload, captured_at
		) VALUES (
			$1, $2, 'QUOTE', $3, $4,
			$5, $6, $7, $8, $5, $9,
			$10, $11, $12, NULLIF($13, ''), $14, $15
		)
		ON CONFLICT (symbol_id, provider_name, data_type, price_date) DO UPDATE
		   SET as_of = EXCLUDED.as_of,
		       price = EXCLUDED.price,
		       open_price = EXCLUDED.open_price,
		       high_price = EXCLUDED.high_price,
		       low_price = EXCLUDED.low_price,
		       close_price = EXCLUDED.close_price,
		       previous_close = EXCLUDED.previous_close,
		       change_amount = EXCLUDED.change_amount,
		       change_percent = EXCLUDED.change_percent,
		       volume = EXCLUDED.volume,
		       currency = COALESCE(EXCLUDED.currency, market_data_snapshots.currency),
		       raw_payload = EXCLUDED.raw_payload,
		       captured_at = NOW()`,
		symbolID,
		quote.Provider,
		asOf,
		asOf,
		quote.Price,
		nullableDecimal(quote.Open),
		nullableDecimal(quote.High),
		nullableDecimal(quote.Low),
		nullableDecimal(quote.PreviousClose),
		nullableDecimal(quote.Change),
		nullableDecimal(quote.ChangePercent),
		nullableInt64(quote.Volume),
		strings.ToUpper(strings.TrimSpace(quote.Currency)),
		raw,
		capturedAt,
	)
	if err != nil {
		return fmt.Errorf("saving market quote snapshot: %w", err)
	}
	return nil
}

func (r *PostgresRepository) SaveDailyPrices(ctx context.Context, symbol string, provider string, bars []domain.PriceBar) error {
	if r == nil || r.pool == nil || len(bars) == 0 {
		return nil
	}
	symbolID, err := r.ensureSymbolID(ctx, symbol)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning market data snapshot transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, bar := range bars {
		raw, _ := json.Marshal(bar)
		asOf := bar.Date
		if asOf.IsZero() {
			asOf = time.Now().UTC()
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO market_data_snapshots (
				symbol_id, provider_name, data_type, price_date, as_of,
				price, open_price, high_price, low_price, close_price, adjusted_close,
				volume, currency, raw_payload, captured_at
			) VALUES (
				$1, $2, 'DAILY', $3, $4,
				$5, $6, $7, $8, $5, $9,
				$10, NULLIF($11, ''), $12, NOW()
			)
			ON CONFLICT (symbol_id, provider_name, data_type, price_date) DO UPDATE
			   SET as_of = EXCLUDED.as_of,
			       price = EXCLUDED.price,
			       open_price = EXCLUDED.open_price,
			       high_price = EXCLUDED.high_price,
			       low_price = EXCLUDED.low_price,
			       close_price = EXCLUDED.close_price,
			       adjusted_close = EXCLUDED.adjusted_close,
			       volume = EXCLUDED.volume,
			       currency = COALESCE(EXCLUDED.currency, market_data_snapshots.currency),
			       raw_payload = EXCLUDED.raw_payload,
			       captured_at = NOW()`,
			symbolID,
			provider,
			asOf,
			asOf,
			bar.Close,
			bar.Open,
			bar.High,
			bar.Low,
			bar.AdjustedClose,
			nullableInt64(bar.Volume),
			strings.ToUpper(strings.TrimSpace(bar.Currency)),
			raw,
		)
		if err != nil {
			return fmt.Errorf("saving daily market data snapshot: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing market data snapshots: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetLatestQuote(ctx context.Context, symbol string) (*domain.Quote, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, `
		SELECT s.symbol, d.provider_name, s.asset_type,
		       d.price, COALESCE(d.open_price, 0), COALESCE(d.high_price, 0), COALESCE(d.low_price, 0),
		       COALESCE(d.previous_close, 0), COALESCE(d.change_amount, 0), COALESCE(d.change_percent, 0),
		       COALESCE(d.volume, 0), COALESCE(d.currency, ''), d.as_of, d.captured_at
		  FROM market_data_snapshots d
		  JOIN market_symbols s ON s.id = d.symbol_id
		 WHERE s.symbol = $1 AND d.data_type = 'QUOTE'
		 ORDER BY d.as_of DESC, d.captured_at DESC
		 LIMIT 1`,
		normalizeSymbol(symbol),
	)
	var q domain.Quote
	var assetType string
	err := row.Scan(
		&q.Symbol, &q.Provider, &assetType,
		&q.Price, &q.Open, &q.High, &q.Low,
		&q.PreviousClose, &q.Change, &q.ChangePercent,
		&q.Volume, &q.Currency, &q.AsOf, &q.CapturedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("loading latest market quote snapshot: %w", err)
	}
	q.AssetType = domain.AssetType(assetType)
	return &q, nil
}

func (r *PostgresRepository) ListDailyPrices(ctx context.Context, symbol string, limit int) ([]domain.PriceBar, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 250
	}
	rows, err := r.pool.Query(ctx, `
		SELECT s.symbol, d.provider_name, d.price_date,
		       COALESCE(d.open_price, d.close_price), COALESCE(d.high_price, d.close_price),
		       COALESCE(d.low_price, d.close_price), d.close_price, d.adjusted_close,
		       COALESCE(d.volume, 0), COALESCE(d.currency, '')
		  FROM market_data_snapshots d
		  JOIN market_symbols s ON s.id = d.symbol_id
		 WHERE s.symbol = $1 AND d.data_type = 'DAILY'
		 ORDER BY d.price_date DESC, d.captured_at DESC
		 LIMIT $2`,
		normalizeSymbol(symbol), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("listing cached daily market prices: %w", err)
	}
	defer rows.Close()

	out := []domain.PriceBar{}
	for rows.Next() {
		var bar domain.PriceBar
		var adjusted *decimal.Decimal
		if err := rows.Scan(
			&bar.Symbol, &bar.Provider, &bar.Date,
			&bar.Open, &bar.High, &bar.Low, &bar.Close, &adjusted,
			&bar.Volume, &bar.Currency,
		); err != nil {
			return nil, fmt.Errorf("scanning cached daily market price: %w", err)
		}
		bar.AdjustedClose = adjusted
		out = append(out, bar)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) LogProviderRequest(ctx context.Context, log domain.ProviderRequestLog) error {
	if r == nil || r.pool == nil {
		return nil
	}
	if log.RequestedAt.IsZero() {
		log.RequestedAt = time.Now().UTC()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO provider_requests_log (
			provider_name, symbol, operation, status, status_code,
			error_code, error_message, duration_ms, requested_at
		) VALUES (
			$1, $2, $3, $4, NULLIF($5, 0),
			NULLIF($6, ''), NULLIF($7, ''), $8, $9
		)`,
		log.ProviderName,
		normalizeSymbol(log.Symbol),
		log.Operation,
		log.Status,
		log.StatusCode,
		log.ErrorCode,
		log.ErrorMessage,
		log.Duration.Milliseconds(),
		log.RequestedAt,
	)
	if err != nil {
		return fmt.Errorf("logging market data provider request: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListLatestProviderRequests(ctx context.Context, limit int) ([]domain.ProviderRequestLog, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT provider_name, symbol, operation, status, COALESCE(status_code, 0),
		       COALESCE(error_code, ''), COALESCE(error_message, ''),
		       duration_ms, requested_at
		  FROM provider_requests_log
		 ORDER BY requested_at DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("listing provider request logs: %w", err)
	}
	defer rows.Close()

	out := []domain.ProviderRequestLog{}
	for rows.Next() {
		var log domain.ProviderRequestLog
		var durationMS int64
		if err := rows.Scan(
			&log.ProviderName, &log.Symbol, &log.Operation, &log.Status, &log.StatusCode,
			&log.ErrorCode, &log.ErrorMessage, &durationMS, &log.RequestedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning provider request log: %w", err)
		}
		log.Duration = time.Duration(durationMS) * time.Millisecond
		out = append(out, log)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) ensureSymbolID(ctx context.Context, symbol string) (string, error) {
	symbol = normalizeSymbol(symbol)
	if symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO market_symbols (symbol, asset_type)
		VALUES ($1, 'UNKNOWN')
		ON CONFLICT (symbol) DO UPDATE
		   SET updated_at = NOW()
		RETURNING id`,
		symbol,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("ensuring market symbol: %w", err)
	}
	return id, nil
}

func normalizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}

func nullableDecimal(v decimal.Decimal) any {
	if v.IsZero() {
		return nil
	}
	return v
}

func nullableInt64(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}

var _ domain.SnapshotRepository = (*PostgresRepository)(nil)
var _ domain.ProviderRequestLogger = (*PostgresRepository)(nil)
