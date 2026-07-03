package command

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
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// fundAUMFixture provides in-memory stubs sufficient to exercise every
// rejection path of ComputeFundAUMHandler. Happy-path insert is exercised
// by the E2E lifecycle test which runs against a live database; trying to
// fake the pool here would require a wholesale refactor of the existing
// `withTransaction` pattern that other commands share.
type fundAUMFixture struct {
	fund         *entity.Fund
	portfolios   []*entity.Portfolio
	snapshots    map[uuid.UUID]*entity.ValuationSnapshot
	existingAUM  *entity.AUMSnapshot
	businessDate time.Time
}

func newFundAUMFixture() *fundAUMFixture {
	fund := &entity.Fund{
		ID:           uuid.New(),
		BaseCurrency: "THB",
		Status:       vo.FundStatusActive,
	}
	portfolioA := &entity.Portfolio{ID: uuid.New(), FundID: fund.ID, Status: vo.PortfolioStatusActive, BaseCurrency: "THB", ValuationCurrency: "THB"}
	portfolioB := &entity.Portfolio{ID: uuid.New(), FundID: fund.ID, Status: vo.PortfolioStatusActive, BaseCurrency: "THB", ValuationCurrency: "THB"}

	bd := time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC)

	return &fundAUMFixture{
		fund:         fund,
		portfolios:   []*entity.Portfolio{portfolioA, portfolioB},
		snapshots:    map[uuid.UUID]*entity.ValuationSnapshot{},
		businessDate: bd,
	}
}

func (f *fundAUMFixture) seedSnapshot(p *entity.Portfolio, aum decimal.Decimal, hash, ccy string) {
	f.snapshots[p.ID] = &entity.ValuationSnapshot{
		ID:           uuid.New(),
		PortfolioID:  p.ID,
		BusinessDate: f.businessDate,
		ValuationCcy: ccy,
		AUM:          aum,
		PriceSetHash: hash,
		Source:       vo.ValuationSourceInternal,
		IsIndicative: true,
	}
}

func (f *fundAUMFixture) handler() *ComputeFundAUMHandler {
	return NewComputeFundAUMHandler(
		nil, // pool: rejection paths return before withTransaction.
		fundAUMFundRepo{fund: f.fund},
		fundAUMPortfolioRepo{portfolios: f.portfolios},
		&fundAUMValuationRepo{snapshots: f.snapshots, existing: f.existingAUM},
		nilAudit{},
		func() time.Time { return f.businessDate.Add(8 * time.Hour) },
	)
}

func TestComputeFundAUM_RejectsMissingArgs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		req  ComputeFundAUMRequest
		want string
	}{
		{"missing fund_id", ComputeFundAUMRequest{BusinessDate: time.Now(), ActorID: uuid.New()}, "fund_id"},
		{"missing business_date", ComputeFundAUMRequest{FundID: uuid.New(), ActorID: uuid.New()}, "business_date"},
		{"missing actor_id", ComputeFundAUMRequest{FundID: uuid.New(), BusinessDate: time.Now()}, "actor_id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := newFundAUMFixture().handler()
			_, err := h.Handle(context.Background(), tc.req)
			var invalid *domain.ErrInvalidDecisionRequest
			require.ErrorAs(t, err, &invalid)
			require.Equal(t, tc.want, invalid.Field)
		})
	}
}

func TestComputeFundAUM_RejectsUnknownFund(t *testing.T) {
	t.Parallel()
	f := newFundAUMFixture()
	f.fund = nil // GetByID returns nil
	h := f.handler()

	_, err := h.Handle(context.Background(), ComputeFundAUMRequest{
		FundID:       uuid.New(),
		BusinessDate: f.businessDate,
		ActorID:      uuid.New(),
	})
	var nf *domain.ErrFundNotFound
	require.ErrorAs(t, err, &nf)
}

func TestComputeFundAUM_IdempotentReturnsExistingWithoutInsert(t *testing.T) {
	t.Parallel()
	f := newFundAUMFixture()
	existing := &entity.AUMSnapshot{
		ID:           uuid.New(),
		ScopeType:    vo.AumScopeFund,
		ScopeID:      f.fund.ID,
		BusinessDate: f.businessDate,
		AUM:          decimal.NewFromInt(123),
		ValuationCcy: "THB",
		Source:       vo.ValuationSourceInternal,
	}
	f.existingAUM = existing
	h := f.handler()

	res, err := h.Handle(context.Background(), ComputeFundAUMRequest{
		FundID:       f.fund.ID,
		BusinessDate: f.businessDate,
		ActorID:      uuid.New(),
	})
	require.NoError(t, err)
	require.True(t, res.Idempotent)
	require.Equal(t, existing.ID, res.Snapshot.ID)
}

func TestComputeFundAUM_RejectsFundWithNoPortfolios(t *testing.T) {
	t.Parallel()
	f := newFundAUMFixture()
	f.portfolios = nil
	h := f.handler()

	_, err := h.Handle(context.Background(), ComputeFundAUMRequest{
		FundID:       f.fund.ID,
		BusinessDate: f.businessDate,
		ActorID:      uuid.New(),
	})
	var incomplete *domain.ErrIncompleteFundValuation
	require.ErrorAs(t, err, &incomplete)
	require.Equal(t, "MISSING_PORTFOLIO_SNAPSHOT", incomplete.Reason)
}

func TestComputeFundAUM_RejectsMissingPortfolioSnapshot(t *testing.T) {
	t.Parallel()
	f := newFundAUMFixture()
	// Only seed one out of two portfolios.
	f.seedSnapshot(f.portfolios[0], decimal.NewFromInt(100), "hash-A", "THB")
	h := f.handler()

	_, err := h.Handle(context.Background(), ComputeFundAUMRequest{
		FundID:       f.fund.ID,
		BusinessDate: f.businessDate,
		ActorID:      uuid.New(),
	})
	var incomplete *domain.ErrIncompleteFundValuation
	require.ErrorAs(t, err, &incomplete)
	require.Equal(t, "MISSING_PORTFOLIO_SNAPSHOT", incomplete.Reason)
}

func TestComputeFundAUM_RejectsCurrencyMismatch(t *testing.T) {
	t.Parallel()
	f := newFundAUMFixture()
	f.seedSnapshot(f.portfolios[0], decimal.NewFromInt(100), "hash-A", "THB")
	f.seedSnapshot(f.portfolios[1], decimal.NewFromInt(200), "hash-A", "USD") // different from fund.BaseCurrency
	h := f.handler()

	_, err := h.Handle(context.Background(), ComputeFundAUMRequest{
		FundID:       f.fund.ID,
		BusinessDate: f.businessDate,
		ActorID:      uuid.New(),
	})
	var incomplete *domain.ErrIncompleteFundValuation
	require.ErrorAs(t, err, &incomplete)
	require.Equal(t, "VALUATION_CCY_MISMATCH", incomplete.Reason)
}

func TestComputeFundAUM_RejectsHashMismatch(t *testing.T) {
	t.Parallel()
	f := newFundAUMFixture()
	f.seedSnapshot(f.portfolios[0], decimal.NewFromInt(100), "hash-A", "THB")
	f.seedSnapshot(f.portfolios[1], decimal.NewFromInt(200), "hash-B", "THB")
	h := f.handler()

	_, err := h.Handle(context.Background(), ComputeFundAUMRequest{
		FundID:       f.fund.ID,
		BusinessDate: f.businessDate,
		ActorID:      uuid.New(),
	})
	var incomplete *domain.ErrIncompleteFundValuation
	require.ErrorAs(t, err, &incomplete)
	require.Equal(t, "PRICE_SET_HASH_MISMATCH", incomplete.Reason)
}

func TestComputeFundAUM_PropagatesRepositoryErrors(t *testing.T) {
	t.Parallel()
	f := newFundAUMFixture()
	wantErr := errors.New("db down")
	h := NewComputeFundAUMHandler(
		nil,
		fundAUMFundRepo{fund: f.fund},
		fundAUMPortfolioRepo{portfolios: f.portfolios},
		&fundAUMValuationRepo{getAUMErr: wantErr},
		nilAudit{},
		nil,
	)

	_, err := h.Handle(context.Background(), ComputeFundAUMRequest{
		FundID:       f.fund.ID,
		BusinessDate: f.businessDate,
		ActorID:      uuid.New(),
	})
	require.ErrorIs(t, err, wantErr)
}

// ─── stubs ───────────────────────────────────────────────────────────────────

type fundAUMFundRepo struct{ fund *entity.Fund }

func (r fundAUMFundRepo) Create(context.Context, pgx.Tx, *entity.Fund) error { return nil }
func (r fundAUMFundRepo) GetByID(context.Context, uuid.UUID) (*entity.Fund, error) {
	return r.fund, nil
}
func (r fundAUMFundRepo) GetByCode(context.Context, string) (*entity.Fund, error) { return nil, nil }
func (r fundAUMFundRepo) GetByContractCode(context.Context, string) (*entity.Fund, error) {
	return nil, nil
}
func (r fundAUMFundRepo) List(context.Context, domain.FundListFilter) ([]*entity.Fund, int, error) {
	return nil, 0, nil
}
func (r fundAUMFundRepo) Update(context.Context, pgx.Tx, *entity.Fund) error { return nil }
func (r fundAUMFundRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r fundAUMFundRepo) CountActivePortfolios(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

type fundAUMPortfolioRepo struct{ portfolios []*entity.Portfolio }

func (r fundAUMPortfolioRepo) Create(context.Context, pgx.Tx, *entity.Portfolio) error { return nil }
func (r fundAUMPortfolioRepo) GetByID(context.Context, uuid.UUID) (*entity.Portfolio, error) {
	return nil, nil
}
func (r fundAUMPortfolioRepo) GetByFundCode(context.Context, uuid.UUID, string) (*entity.Portfolio, error) {
	return nil, nil
}
func (r fundAUMPortfolioRepo) List(_ context.Context, _ domain.PortfolioListFilter) ([]*entity.Portfolio, int, error) {
	return r.portfolios, len(r.portfolios), nil
}
func (r fundAUMPortfolioRepo) Update(context.Context, pgx.Tx, *entity.Portfolio) error { return nil }
func (r fundAUMPortfolioRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r fundAUMPortfolioRepo) HasOpenActivity(context.Context, uuid.UUID, time.Time) (bool, string, error) {
	return false, "", nil
}

type fundAUMValuationRepo struct {
	snapshots map[uuid.UUID]*entity.ValuationSnapshot
	existing  *entity.AUMSnapshot
	getAUMErr error
}

func (r *fundAUMValuationRepo) Insert(context.Context, pgx.Tx, *entity.ValuationSnapshot) error {
	return nil
}
func (r *fundAUMValuationRepo) GetLatest(context.Context, uuid.UUID, vo.ValuationSource) (*entity.ValuationSnapshot, error) {
	return nil, nil
}
func (r *fundAUMValuationRepo) GetByID(context.Context, uuid.UUID) (*entity.ValuationSnapshot, error) {
	return nil, nil
}
func (r *fundAUMValuationRepo) GetByPortfolioBusinessDate(_ context.Context, portfolioID uuid.UUID, _ time.Time, _ vo.ValuationSource) (*entity.ValuationSnapshot, error) {
	return r.snapshots[portfolioID], nil
}
func (r *fundAUMValuationRepo) List(context.Context, uuid.UUID, time.Time, time.Time, int, int) ([]*entity.ValuationSnapshot, int, error) {
	return nil, 0, nil
}
func (r *fundAUMValuationRepo) InsertNAV(context.Context, pgx.Tx, *entity.NAVSnapshot) error {
	return nil
}
func (r *fundAUMValuationRepo) ListNAV(context.Context, uuid.UUID, time.Time, time.Time, int, int) ([]*entity.NAVSnapshot, int, error) {
	return nil, 0, nil
}
func (r *fundAUMValuationRepo) InsertAUM(context.Context, pgx.Tx, *entity.AUMSnapshot) error {
	return nil
}
func (r *fundAUMValuationRepo) GetAUM(context.Context, vo.AumScopeType, uuid.UUID, time.Time, vo.ValuationSource) (*entity.AUMSnapshot, error) {
	return r.existing, r.getAUMErr
}
func (r *fundAUMValuationRepo) ListAUM(context.Context, vo.AumScopeType, uuid.UUID, time.Time, time.Time, int, int) ([]*entity.AUMSnapshot, int, error) {
	return nil, 0, nil
}
