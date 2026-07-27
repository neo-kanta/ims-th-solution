package persistence

import (
	"context"
	"fmt"
	"os"
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

func TestIntegrationPositionRepositoryConcurrentFirstInsertRetry(t *testing.T) {
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

	schema := "inv_it_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
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
	instrumentID := uuid.New()
	businessDate := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)
	actor := uuid.New()

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- withTestTx(ctx, pool, func(tx pgx.Tx) error {
				qty := decimal.NewFromInt(10)
				price := decimal.NewFromInt(5)
				txn := &entity.PortfolioTransaction{
					ID:              uuid.New(),
					PortfolioID:     portfolioID,
					FundID:          func() *uuid.UUID { v := uuid.New(); return &v }(),
					InstrumentID:    &instrumentID,
					TransactionType: vo.TransactionTypeBuy,
					Quantity:        &qty,
					Price:           &price,
					Currency:        "THB",
					GrossAmount:     decimal.NewFromInt(50),
					Fees:            decimal.Zero,
					NetAmount:       decimal.NewFromInt(-50),
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

	var got entity.PortfolioPosition
	err = pool.QueryRow(ctx, `
		SELECT quantity, average_cost, cost_basis, version
		  FROM investment__portfolio_positions
		 WHERE portfolio_id = $1 AND instrument_id = $2`,
		portfolioID, instrumentID,
	).Scan(&got.Quantity, &got.AverageCost, &got.CostBasis, &got.Version)
	require.NoError(t, err)
	require.True(t, decimal.NewFromInt(20).Equal(got.Quantity), "quantity = %s", got.Quantity)
}

func integrationDSN() string {
	if dsn := os.Getenv("IMS_TEST_DATABASE_DSN"); dsn != "" {
		return dsn
	}
	return "host=localhost port=5437 user=ims_app password=ims_dev_password dbname=ims_dev sslmode=disable"
}

func withTestTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
