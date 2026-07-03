package persistence

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func asOfIntegrationDSN() string {
	if dsn := os.Getenv("IMS_TEST_DATABASE_DSN"); dsn != "" {
		return dsn
	}
	return "host=localhost port=5437 user=ims_app password=ims_dev_password dbname=ims_dev sslmode=disable"
}

// TestIntegrationGetSnapshotAsOfPicksLatestOnOrBeforeBusinessDate proves the
// real SQL behind GetSnapshotAsOf: given snapshots on 06-28 and 07-01 for the
// same symbol, a request for 06-30 (a date with no snapshot of its own) must
// carry forward the 06-28 row — never the later 07-01 row, and never the
// unconditionally-latest row a naive "ORDER BY date DESC LIMIT 1" without a
// WHERE bound would return.
func TestIntegrationGetSnapshotAsOfPicksLatestOnOrBeforeBusinessDate(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, asOfIntegrationDSN())
	if err != nil {
		t.Skipf("Postgres integration database unavailable: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Postgres integration database unavailable: %v", err)
	}

	repo := NewPostgresRepository(pool)
	symbol := "IT-ASOF-" + strings.ToUpper(uuid.NewString()[:8])

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM market_data_snapshots WHERE symbol_id IN (SELECT id FROM market_symbols WHERE symbol = $1)", symbol)
		_, _ = pool.Exec(ctx, "DELETE FROM market_symbols WHERE symbol = $1", symbol)
	}()

	var symbolID string
	err = pool.QueryRow(ctx, `
		INSERT INTO market_symbols (symbol, asset_type) VALUES ($1, 'UNKNOWN') RETURNING id`,
		symbol,
	).Scan(&symbolID)
	require.NoError(t, err)

	insertSnapshot := func(dataType string, priceDate time.Time, price string) {
		_, err := pool.Exec(ctx, `
			INSERT INTO market_data_snapshots (
				symbol_id, provider_name, data_type, price_date, as_of, price, close_price, captured_at
			) VALUES ($1, 'internal_demo', $2, $3, $4, $5, $5, NOW())`,
			symbolID, dataType, priceDate, priceDate, decimal.RequireFromString(price),
		)
		require.NoError(t, err)
	}

	insertSnapshot("DAILY", time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC), "1030.00")
	insertSnapshot("DAILY", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), "1032.00")

	// Requesting a date strictly between the two snapshots must carry the
	// earlier (06-28) row forward — not the later, not "no data".
	q, err := repo.GetSnapshotAsOf(ctx, symbol, time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NotNil(t, q)
	require.True(t, decimal.RequireFromString("1030.00").Equal(q.Price), "price = %s", q.Price)
	require.Equal(t, "2026-06-28", q.AsOf.Format("2006-01-02"))

	// Requesting exactly the later date must return that day's own snapshot.
	q, err = repo.GetSnapshotAsOf(ctx, symbol, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NotNil(t, q)
	require.True(t, decimal.RequireFromString("1032.00").Equal(q.Price), "price = %s", q.Price)

	// Requesting a date before any snapshot must return (nil, nil), not an error.
	q, err = repo.GetSnapshotAsOf(ctx, symbol, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.Nil(t, q)
}

// TestIntegrationGetSnapshotAsOfPrefersDailyOverQuoteOnSameDate proves the
// DAILY-vs-QUOTE tiebreak: an end-of-day close is a cleaner mark-to-market
// value than an intraday QUOTE captured earlier the same day.
func TestIntegrationGetSnapshotAsOfPrefersDailyOverQuoteOnSameDate(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, asOfIntegrationDSN())
	if err != nil {
		t.Skipf("Postgres integration database unavailable: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Postgres integration database unavailable: %v", err)
	}

	repo := NewPostgresRepository(pool)
	symbol := "IT-TIE-" + strings.ToUpper(uuid.NewString()[:8])

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM market_data_snapshots WHERE symbol_id IN (SELECT id FROM market_symbols WHERE symbol = $1)", symbol)
		_, _ = pool.Exec(ctx, "DELETE FROM market_symbols WHERE symbol = $1", symbol)
	}()

	var symbolID string
	err = pool.QueryRow(ctx, `
		INSERT INTO market_symbols (symbol, asset_type) VALUES ($1, 'UNKNOWN') RETURNING id`,
		symbol,
	).Scan(&symbolID)
	require.NoError(t, err)

	priceDate := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	_, err = pool.Exec(ctx, `
		INSERT INTO market_data_snapshots (symbol_id, provider_name, data_type, price_date, as_of, price, close_price, captured_at)
		VALUES ($1, 'internal_demo', 'QUOTE', $2, $3, $4, $4, NOW())`,
		symbolID, priceDate, priceDate, decimal.RequireFromString("999.00"),
	)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO market_data_snapshots (symbol_id, provider_name, data_type, price_date, as_of, price, close_price, captured_at)
		VALUES ($1, 'internal_demo', 'DAILY', $2, $3, $4, $4, NOW())`,
		symbolID, priceDate, priceDate, decimal.RequireFromString("1000.00"),
	)
	require.NoError(t, err)

	q, err := repo.GetSnapshotAsOf(ctx, symbol, priceDate)
	require.NoError(t, err)
	require.NotNil(t, q)
	require.True(t, decimal.RequireFromString("1000.00").Equal(q.Price), "expected DAILY close to win the tiebreak, got %s", q.Price)
}
