package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// TestIntegrationGetByCode_AmbiguousCodeReturnsExplicitError exercises the
// real DB gap described on PortfolioRepository.GetByCode: the schema only
// enforces code uniqueness per fund (uq_inv_portfolios_fund_code_alive), so
// two alive portfolios under different funds can legally share a code.
// GetByCode must refuse to guess in that case rather than silently
// returning one of them.
func TestIntegrationGetByCode_AmbiguousCodeReturnsExplicitError(t *testing.T) {
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

	schema := "inv_code_it_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	_, err = admin.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", schema))
	require.NoError(t, err)
	defer admin.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))

	_, err = admin.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE %s.investment__portfolios (
			id                  UUID PRIMARY KEY,
			fund_id             UUID NOT NULL,
			code                VARCHAR(40) NOT NULL,
			name                VARCHAR(200) NOT NULL,
			description         TEXT,
			base_currency       CHAR(3) NOT NULL,
			valuation_currency  CHAR(3) NOT NULL,
			strategy_code       VARCHAR(40),
			style_id            UUID,
			manager_user_id     UUID,
			benchmark           VARCHAR(100),
			risk_profile        VARCHAR(20),
			inception_date      DATE NOT NULL,
			status              VARCHAR(20) NOT NULL,
			has_units           BOOLEAN NOT NULL DEFAULT false,
			tax_lot_method      VARCHAR(20) NOT NULL DEFAULT 'AVERAGE',
			version             INTEGER NOT NULL DEFAULT 1,
			created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by          UUID,
			updated_by          UUID,
			deleted_at          TIMESTAMPTZ
		);`, schema))
	require.NoError(t, err)

	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO "+schema)
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewPostgresPortfolioRepository(pool)

	fundA, fundB := uuid.New(), uuid.New()
	actor := uuid.New()
	inception := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	const sharedCode = "DUP-CODE"

	makePortfolio := func(fundID uuid.UUID) *entity.Portfolio {
		return &entity.Portfolio{
			ID:                uuid.New(),
			FundID:            fundID,
			Code:              sharedCode,
			Name:              "Test Portfolio",
			BaseCurrency:      "THB",
			ValuationCurrency: "THB",
			InceptionDate:     inception,
			Status:            vo.PortfolioStatusActive,
			TaxLotMethod:      vo.TaxLotMethod("AVERAGE"),
			Version:           1,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
			CreatedBy:         &actor,
			UpdatedBy:         &actor,
		}
	}

	require.NoError(t, withTestTx(ctx, pool, func(tx pgx.Tx) error {
		return repo.Create(ctx, tx, makePortfolio(fundA))
	}))

	// Only one alive portfolio has this code so far — must resolve cleanly.
	got, err := repo.GetByCode(ctx, sharedCode)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, fundA, got.FundID)

	// A second, different fund reuses the same code — the DB allows this
	// because uniqueness is only enforced per (fund_id, code).
	require.NoError(t, withTestTx(ctx, pool, func(tx pgx.Tx) error {
		return repo.Create(ctx, tx, makePortfolio(fundB))
	}))

	got, err = repo.GetByCode(ctx, sharedCode)
	require.Nil(t, got)
	require.Error(t, err)
	var ambiguous *domain.ErrAmbiguousPortfolioCode
	require.True(t, errors.As(err, &ambiguous), "expected *domain.ErrAmbiguousPortfolioCode, got %T: %v", err, err)
	require.Equal(t, sharedCode, ambiguous.Code)
	require.Equal(t, 2, ambiguous.Count)
}

// TestIntegrationGetByCode_IgnoresSoftDeletedDuplicates confirms a
// soft-deleted portfolio sharing a code with an alive one is not treated as
// an ambiguity — only alive rows compete for a code.
func TestIntegrationGetByCode_IgnoresSoftDeletedDuplicates(t *testing.T) {
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

	schema := "inv_code_it_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	_, err = admin.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", schema))
	require.NoError(t, err)
	defer admin.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))

	_, err = admin.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE %s.investment__portfolios (
			id                  UUID PRIMARY KEY,
			fund_id             UUID NOT NULL,
			code                VARCHAR(40) NOT NULL,
			name                VARCHAR(200) NOT NULL,
			description         TEXT,
			base_currency       CHAR(3) NOT NULL,
			valuation_currency  CHAR(3) NOT NULL,
			strategy_code       VARCHAR(40),
			style_id            UUID,
			manager_user_id     UUID,
			benchmark           VARCHAR(100),
			risk_profile        VARCHAR(20),
			inception_date      DATE NOT NULL,
			status              VARCHAR(20) NOT NULL,
			has_units           BOOLEAN NOT NULL DEFAULT false,
			tax_lot_method      VARCHAR(20) NOT NULL DEFAULT 'AVERAGE',
			version             INTEGER NOT NULL DEFAULT 1,
			created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by          UUID,
			updated_by          UUID,
			deleted_at          TIMESTAMPTZ
		);`, schema))
	require.NoError(t, err)

	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO "+schema)
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)
	defer pool.Close()

	repo := NewPostgresPortfolioRepository(pool)

	actor := uuid.New()
	inception := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	const sharedCode = "SOFT-DUP"

	alive := &entity.Portfolio{
		ID: uuid.New(), FundID: uuid.New(), Code: sharedCode, Name: "Alive",
		BaseCurrency: "THB", ValuationCurrency: "THB", InceptionDate: inception,
		Status: vo.PortfolioStatusActive, TaxLotMethod: vo.TaxLotMethod("AVERAGE"),
		Version: 1, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		CreatedBy: &actor, UpdatedBy: &actor,
	}
	deleted := &entity.Portfolio{
		ID: uuid.New(), FundID: uuid.New(), Code: sharedCode, Name: "Deleted",
		BaseCurrency: "THB", ValuationCurrency: "THB", InceptionDate: inception,
		Status: vo.PortfolioStatusClosed, TaxLotMethod: vo.TaxLotMethod("AVERAGE"),
		Version: 1, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
		CreatedBy: &actor, UpdatedBy: &actor,
	}

	require.NoError(t, withTestTx(ctx, pool, func(tx pgx.Tx) error {
		return repo.Create(ctx, tx, alive)
	}))
	require.NoError(t, withTestTx(ctx, pool, func(tx pgx.Tx) error {
		return repo.Create(ctx, tx, deleted)
	}))
	require.NoError(t, withTestTx(ctx, pool, func(tx pgx.Tx) error {
		return repo.SoftDelete(ctx, tx, deleted.ID, deleted.Version, actor)
	}))

	got, err := repo.GetByCode(ctx, sharedCode)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, alive.ID, got.ID)
}
