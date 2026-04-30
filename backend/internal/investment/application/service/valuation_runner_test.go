package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
)

func TestLookupFxAcceptsDocumentedPairKey(t *testing.T) {
	rate := decimal.RequireFromString("36.25")
	got, err := lookupFx(map[string]decimal.Decimal{"USD->THB": rate}, "USD", "THB")
	require.NoError(t, err)
	require.True(t, rate.Equal(got), "rate = %s", got)
}

func TestIsPriceStaleExactThresholdIsFresh(t *testing.T) {
	valuationDate := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)
	priceDate := valuationDate.AddDate(0, 0, -1)

	require.False(t, isPriceStale(priceDate, valuationDate, 1))
	require.True(t, isPriceStale(priceDate.AddDate(0, 0, -1), valuationDate, 1))
}

func TestValuationRunnerRejectsMissingPriceSnapshot(t *testing.T) {
	portfolioID := uuid.New()
	fundID := uuid.New()
	instrumentID := uuid.New()
	actorID := uuid.New()
	businessDate := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)

	runner := NewValuationRunner(
		nil,
		&valuationPortfolioRepo{portfolio: &entity.Portfolio{
			ID:                portfolioID,
			FundID:            fundID,
			BaseCurrency:      "USD",
			ValuationCurrency: "USD",
			HasUnits:          false,
		}},
		&valuationPositionRepo{positions: []*entity.PortfolioPosition{{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			InstrumentID: instrumentID,
			Quantity:     decimal.NewFromInt(10),
			CostBasis:    decimal.NewFromInt(1000),
		}}},
		&valuationCashRepo{},
		&valuationPriceRepo{},
		&valuationInstrumentRepo{instrument: &entity.Instrument{
			ID:       instrumentID,
			Currency: "USD",
		}},
		nil,
		nil,
		func() time.Time { return businessDate },
	)

	_, err := runner.Run(context.Background(), RunRequest{
		PortfolioID:  portfolioID,
		BusinessDate: businessDate,
		FxRates:      map[string]decimal.Decimal{},
		ActorID:      actorID,
	})

	require.Error(t, err)
	var invalid *domain.ErrInvalidDecisionRequest
	require.True(t, errors.As(err, &invalid))
	require.Equal(t, "price_snapshot", invalid.Field)
	require.Contains(t, invalid.Detail, instrumentID.String())
}

type valuationPortfolioRepo struct {
	portfolio *entity.Portfolio
}

func (r *valuationPortfolioRepo) Create(ctx context.Context, tx pgx.Tx, p *entity.Portfolio) error {
	return nil
}

func (r *valuationPortfolioRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Portfolio, error) {
	if r.portfolio != nil && r.portfolio.ID == id {
		return r.portfolio, nil
	}
	return nil, nil
}

func (r *valuationPortfolioRepo) GetByFundCode(ctx context.Context, fundID uuid.UUID, code string) (*entity.Portfolio, error) {
	return nil, nil
}

func (r *valuationPortfolioRepo) List(ctx context.Context, filter domain.PortfolioListFilter) ([]*entity.Portfolio, int, error) {
	return nil, 0, nil
}

func (r *valuationPortfolioRepo) Update(ctx context.Context, tx pgx.Tx, p *entity.Portfolio) error {
	return nil
}

func (r *valuationPortfolioRepo) SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, expectedVersion int, deletedBy uuid.UUID) error {
	return nil
}

func (r *valuationPortfolioRepo) HasOpenActivity(ctx context.Context, portfolioID uuid.UUID, asOf time.Time) (bool, string, error) {
	return false, "", nil
}

type valuationPositionRepo struct {
	positions []*entity.PortfolioPosition
}

func (r *valuationPositionRepo) GetForUpdate(ctx context.Context, tx pgx.Tx, portfolioID, instrumentID uuid.UUID) (*entity.PortfolioPosition, error) {
	return nil, nil
}

func (r *valuationPositionRepo) Upsert(ctx context.Context, tx pgx.Tx, pos *entity.PortfolioPosition, expectedVersion int) error {
	return nil
}

func (r *valuationPositionRepo) ListByPortfolio(ctx context.Context, portfolioID uuid.UUID) ([]*entity.PortfolioPosition, error) {
	return r.positions, nil
}

type valuationCashRepo struct{}

func (r *valuationCashRepo) InsertMovement(ctx context.Context, tx pgx.Tx, m *entity.CashMovement) error {
	return nil
}

func (r *valuationCashRepo) GetBalanceForUpdate(ctx context.Context, tx pgx.Tx, portfolioID uuid.UUID, currency string) (*entity.CashBalance, error) {
	return nil, nil
}

func (r *valuationCashRepo) UpsertBalance(ctx context.Context, tx pgx.Tx, b *entity.CashBalance, expectedVersion int) error {
	return nil
}

func (r *valuationCashRepo) ListBalances(ctx context.Context, portfolioID uuid.UUID) ([]*entity.CashBalance, error) {
	return nil, nil
}

type valuationPriceRepo struct {
	latest *entity.PriceSnapshot
}

func (r *valuationPriceRepo) Insert(ctx context.Context, tx pgx.Tx, p *entity.PriceSnapshot) error {
	return nil
}

func (r *valuationPriceRepo) GetLatest(ctx context.Context, instrumentID uuid.UUID, asOf time.Time) (*entity.PriceSnapshot, error) {
	return r.latest, nil
}

func (r *valuationPriceRepo) ListByInstrument(ctx context.Context, instrumentID uuid.UUID, from, to time.Time, limit int) ([]*entity.PriceSnapshot, error) {
	return nil, nil
}

type valuationInstrumentRepo struct {
	instrument *entity.Instrument
}

func (r *valuationInstrumentRepo) Create(ctx context.Context, tx pgx.Tx, inst *entity.Instrument) error {
	return nil
}

func (r *valuationInstrumentRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Instrument, error) {
	if r.instrument != nil && r.instrument.ID == id {
		return r.instrument, nil
	}
	return nil, nil
}

func (r *valuationInstrumentRepo) List(ctx context.Context, filter domain.InstrumentListFilter) ([]*entity.Instrument, int, error) {
	return nil, 0, nil
}

func (r *valuationInstrumentRepo) Update(ctx context.Context, tx pgx.Tx, inst *entity.Instrument) error {
	return nil
}

func (r *valuationInstrumentRepo) SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, deletedBy uuid.UUID) error {
	return nil
}

var (
	_ domain.PortfolioRepository         = (*valuationPortfolioRepo)(nil)
	_ domain.PortfolioPositionRepository = (*valuationPositionRepo)(nil)
	_ domain.CashLedgerRepository        = (*valuationCashRepo)(nil)
	_ domain.PriceSnapshotRepository     = (*valuationPriceRepo)(nil)
	_ domain.InstrumentRepository        = (*valuationInstrumentRepo)(nil)
)
