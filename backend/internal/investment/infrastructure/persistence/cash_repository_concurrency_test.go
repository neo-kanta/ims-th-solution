package persistence

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// TestIntegrationCashRepositoryConcurrentFirstInsertRetry asserts the cash
// balance projection survives the same B1-class race as the position
// projection: two goroutines see an empty (portfolio, currency) row,
// each tries to insert version=1 with expected_version=0, only one
// physically succeeds, the loser detects the conflict via RETURNING-zero-
// rows and retries successfully — neither silently no-ops.
//
// Mirrors TestIntegrationPositionRepositoryConcurrentFirstInsertRetry on
// purpose so the two repositories are exercised through the same shape.
func TestIntegrationCashRepositoryConcurrentFirstInsertRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	ctx := context.Background()
	dsn := integrationDSN()
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Skipf("Postgres integration database unavailable: %v", err)
	}
	defer admin.Close(ctx)

	schema := "inv_cash_it_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	_, err = admin.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", schema))
	require.NoError(t, err)
	defer admin.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))

	_, err = admin.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE %s.investment__portfolio_positions (
			id UUID PRIMARY KEY,
			portfolio_id UUID NOT NULL,
			instrument_id UUID NOT NULL,
			quantity DECIMAL(28,8) NOT NULL DEFAULT 0,
			average_cost DECIMAL(28,8) NOT NULL DEFAULT 0,
			cost_basis DECIMAL(28,8) NOT NULL DEFAULT 0,
			last_transaction_id UUID,
			last_business_date DATE,
			version INTEGER NOT NULL DEFAULT 1,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (portfolio_id, instrument_id)
		);
		CREATE TABLE %s.investment__cash_movements (
			id UUID PRIMARY KEY,
			portfolio_id UUID NOT NULL,
			currency CHAR(3) NOT NULL,
			amount DECIMAL(28,8) NOT NULL,
			business_date DATE NOT NULL,
			transaction_id UUID,
			movement_type VARCHAR(20) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			created_by UUID NOT NULL
		);
		CREATE TABLE %s.investment__cash_balances (
			id UUID PRIMARY KEY,
			portfolio_id UUID NOT NULL,
			currency CHAR(3) NOT NULL,
			balance DECIMAL(28,8) NOT NULL DEFAULT 0,
			last_movement_id UUID,
			last_business_date DATE,
			version INTEGER NOT NULL DEFAULT 1,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (portfolio_id, currency)
		);`, schema, schema, schema))
	require.NoError(t, err)

	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	cfg.MaxConns = 4
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO "+schema)
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)
	defer pool.Close()

	positions := NewPostgresPositionRepository(pool)
	cash := NewPostgresCashLedgerRepository(pool)
	projector := service.NewPortfolioProjector(positions, cash)

	portfolioID := uuid.New()
	businessDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)
	actor := uuid.New()

	// Two CASH_IN transactions of 100 each, racing into the same
	// (portfolio, THB) cell. Both start with no row → expected_version=0.
	// Exactly one INSERT physically succeeds; the loser must retry and
	// converge on a non-conflicting UPDATE rather than silently dropping.
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- withTestTx(ctx, pool, func(tx pgx.Tx) error {
				txn := &entity.PortfolioTransaction{
					ID:              uuid.New(),
					PortfolioID:     portfolioID,
					FundID:          uuid.New(),
					InstrumentID:    nil,
					TransactionType: vo.TransactionTypeCashIn,
					Quantity:        nil,
					Price:           nil,
					Currency:        "THB",
					GrossAmount:     decimal.NewFromInt(100),
					Fees:            decimal.Zero,
					NetAmount:       decimal.NewFromInt(100),
					BusinessDate:    businessDate,
					CreatedAt:       time.Now().UTC(),
					CreatedBy:       actor,
				}
				return projector.Apply(ctx, tx, txn, 1)
			})
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	// Both transactions committed. The cash balance must reflect the sum
	// of both deposits — proving neither silently no-op'd.
	var got entity.CashBalance
	err = pool.QueryRow(ctx, `
		SELECT balance, version
		  FROM investment__cash_balances
		 WHERE portfolio_id = $1 AND currency = $2`,
		portfolioID, "THB",
	).Scan(&got.Balance, &got.Version)
	require.NoError(t, err)
	require.True(t, decimal.NewFromInt(200).Equal(got.Balance), "balance = %s, want 200", got.Balance)
	require.GreaterOrEqual(t, got.Version, 2, "version must have advanced at least twice (one per successful upsert)")

	// Cash movements must contain exactly two rows — one per transaction.
	// If a goroutine had silently no-op'd, we'd see only one row.
	var movementCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM investment__cash_movements
		 WHERE portfolio_id = $1 AND currency = $2`,
		portfolioID, "THB",
	).Scan(&movementCount)
	require.NoError(t, err)
	require.Equal(t, 2, movementCount, "expected two cash_movements rows, got %d", movementCount)
}
