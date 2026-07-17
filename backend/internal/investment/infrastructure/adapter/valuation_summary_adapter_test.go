package adapter

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
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ─── fixture builders ───────────────────────────────────────────────────────

func newFund(ccy string, manager *uuid.UUID) *entity.Fund {
	return &entity.Fund{ID: uuid.New(), BaseCurrency: ccy, Status: vo.FundStatusActive, ManagerUserID: manager}
}

func newPortfolio(fundID uuid.UUID) *entity.Portfolio {
	return &entity.Portfolio{ID: uuid.New(), FundID: fundID, Status: vo.PortfolioStatusActive, BaseCurrency: "THB", ValuationCurrency: "THB"}
}

// snapshotPair seeds today's and yesterday's INTERNAL snapshot for a
// portfolio (newest first, matching the real repository's ORDER BY
// business_date DESC).
func snapshotPair(portfolioID uuid.UUID, today time.Time, todayAUM, todayUPnL, todayRPnL, prevAUM, prevUPnL, prevRPnL decimal.Decimal) []*entity.ValuationSnapshot {
	return []*entity.ValuationSnapshot{
		{
			PortfolioID: portfolioID, BusinessDate: today, AUM: todayAUM,
			UnrealisedPnL: todayUPnL, RealisedPnL: todayRPnL, Source: vo.ValuationSourceInternal, CreatedAt: today,
		},
		{
			PortfolioID: portfolioID, BusinessDate: today.AddDate(0, 0, -1), AUM: prevAUM,
			UnrealisedPnL: prevUPnL, RealisedPnL: prevRPnL, Source: vo.ValuationSourceInternal, CreatedAt: today.AddDate(0, 0, -1),
		},
	}
}

// ─── tests ──────────────────────────────────────────────────────────────────

func TestGetValuationSummary_CompanyScopeAggregatesAcrossFunds(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	fundA := newFund("THB", nil)
	fundB := newFund("THB", nil)
	pA := newPortfolio(fundA.ID)
	pB := newPortfolio(fundB.ID)

	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		pA.ID: snapshotPair(pA.ID, today, decimal.NewFromInt(1000), decimal.NewFromInt(50), decimal.NewFromInt(10), decimal.NewFromInt(940), decimal.NewFromInt(40), decimal.NewFromInt(5)),
		pB.ID: snapshotPair(pB.ID, today, decimal.NewFromInt(2000), decimal.NewFromInt(80), decimal.NewFromInt(0), decimal.NewFromInt(1985), decimal.NewFromInt(70), decimal.NewFromInt(0)),
	}}
	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fundA, fundB}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{pA, pB}},
		valRepo,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{
		Scope: contract.ValuationSummaryScopeCompany,
	})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.Equal(t, "THB", res.Currency)
	require.True(t, decimal.NewFromInt(3000).Equal(res.AUM), "got %s", res.AUM)
	// portfolio A day pnl: (50+10)-(40+5)=15 ; portfolio B: (80+0)-(70+0)=10 → 25
	require.True(t, decimal.NewFromInt(25).Equal(res.TodayPnL), "got %s", res.TodayPnL)
	require.Equal(t, 2, res.FundCount)
	require.NotNil(t, res.TodayPnLPercent)
}

func TestGetValuationSummary_MineScopeFiltersByManager(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	userA := uuid.New()
	userB := uuid.New()

	fundMine := newFund("THB", &userA)
	fundOther := newFund("THB", &userB)
	pMine := newPortfolio(fundMine.ID)
	pOther := newPortfolio(fundOther.ID)

	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		pMine.ID:  snapshotPair(pMine.ID, today, decimal.NewFromInt(500), decimal.NewFromInt(5), decimal.Zero, decimal.NewFromInt(490), decimal.NewFromInt(3), decimal.Zero),
		pOther.ID: snapshotPair(pOther.ID, today, decimal.NewFromInt(9000), decimal.NewFromInt(500), decimal.Zero, decimal.NewFromInt(8000), decimal.NewFromInt(100), decimal.Zero),
	}}
	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fundMine, fundOther}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{pMine, pOther}},
		valRepo,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{
		Scope:  contract.ValuationSummaryScopeMine,
		UserID: userA,
	})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.Equal(t, 1, res.FundCount)
	require.True(t, decimal.NewFromInt(500).Equal(res.AUM), "must only include the manager's own fund, got %s", res.AUM)
}

func TestGetValuationSummary_DataScopeRestrictsAccessibleFunds(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	fundAllowed := newFund("THB", nil)
	fundBlocked := newFund("THB", nil)
	pAllowed := newPortfolio(fundAllowed.ID)
	pBlocked := newPortfolio(fundBlocked.ID)

	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		pAllowed.ID: snapshotPair(pAllowed.ID, today, decimal.NewFromInt(100), decimal.Zero, decimal.Zero, decimal.NewFromInt(100), decimal.Zero, decimal.Zero),
		pBlocked.ID: snapshotPair(pBlocked.ID, today, decimal.NewFromInt(999999), decimal.Zero, decimal.Zero, decimal.NewFromInt(999999), decimal.Zero, decimal.Zero),
	}}
	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fundAllowed, fundBlocked}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{pAllowed, pBlocked}},
		valRepo,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{
		Scope:             contract.ValuationSummaryScopeCompany,
		AccessibleFundIDs: []uuid.UUID{fundAllowed.ID}, // non-wildcard: fundBlocked must be excluded
	})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.Equal(t, 1, res.FundCount)
	require.True(t, decimal.NewFromInt(100).Equal(res.AUM), "data scope must exclude funds outside AccessibleFundIDs, got %s", res.AUM)
}

func TestGetValuationSummary_NoDataReturnsUnavailable(t *testing.T) {
	t.Parallel()

	t.Run("no funds in scope", func(t *testing.T) {
		t.Parallel()
		a := NewValuationSummaryAdapter(
			&fakeFundRepo{},
			&fakePortfolioRepo{},
			&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{}},
		)
		res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{
			Scope: contract.ValuationSummaryScopeCompany,
		})
		require.NoError(t, err)
		require.False(t, res.DataAvailable)
	})

	t.Run("funds exist but no valuation snapshot yet", func(t *testing.T) {
		t.Parallel()
		fund := newFund("THB", nil)
		p := newPortfolio(fund.ID)
		a := NewValuationSummaryAdapter(
			&fakeFundRepo{funds: []*entity.Fund{fund}},
			&fakePortfolioRepo{portfolios: []*entity.Portfolio{p}},
			&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{}},
		)
		res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{
			Scope: contract.ValuationSummaryScopeCompany,
		})
		require.NoError(t, err)
		require.False(t, res.DataAvailable)
		require.True(t, res.AUM.IsZero(), "unavailable result must not report a numeric AUM")
	})
}

func TestGetValuationSummary_DayPnLNoBaselineUsesFullCumulative(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	fund := newFund("THB", nil)
	p := newPortfolio(fund.ID)
	// Only one snapshot exists — no prior business day to diff against.
	only := &entity.ValuationSnapshot{
		PortfolioID: p.ID, BusinessDate: today, AUM: decimal.NewFromInt(1000),
		UnrealisedPnL: decimal.NewFromInt(60), RealisedPnL: decimal.NewFromInt(15),
		Source: vo.ValuationSourceInternal, CreatedAt: today,
	}
	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{p}},
		&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{p.ID: {only}}},
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.True(t, decimal.NewFromInt(75).Equal(res.TodayPnL), "with no baseline the whole cumulative total is today's P&L, got %s", res.TodayPnL)
	require.NotNil(t, res.TodayPnLPercent, "percent should fall back to today's AUM as the denominator")
}

func TestGetValuationSummary_MixedCurrencyExcludesMinority(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	thbFundA := newFund("THB", nil)
	thbFundB := newFund("THB", nil)
	usdFund := newFund("USD", nil)
	pThbA := newPortfolio(thbFundA.ID)
	pThbB := newPortfolio(thbFundB.ID)
	pUsd := &entity.Portfolio{ID: uuid.New(), FundID: usdFund.ID, Status: vo.PortfolioStatusActive, BaseCurrency: "USD", ValuationCurrency: "USD"}

	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		pThbA.ID: snapshotPair(pThbA.ID, today, decimal.NewFromInt(100), decimal.Zero, decimal.Zero, decimal.NewFromInt(100), decimal.Zero, decimal.Zero),
		pThbB.ID: snapshotPair(pThbB.ID, today, decimal.NewFromInt(200), decimal.Zero, decimal.Zero, decimal.NewFromInt(200), decimal.Zero, decimal.Zero),
		pUsd.ID:  snapshotPair(pUsd.ID, today, decimal.NewFromInt(999999), decimal.Zero, decimal.Zero, decimal.NewFromInt(999999), decimal.Zero, decimal.Zero),
	}}
	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{thbFundA, thbFundB, usdFund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{pThbA, pThbB, pUsd}},
		valRepo,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.Equal(t, "THB", res.Currency, "THB funds outnumber the single USD fund")
	require.True(t, decimal.NewFromInt(300).Equal(res.AUM), "USD fund must be excluded from the THB aggregate, got %s", res.AUM)
}

func TestGetValuationSummary_PreviousSnapshotErrorPropagates(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	fund := newFund("THB", nil)
	p := newPortfolio(fund.ID)
	wantErr := errors.New("db down")

	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{p}},
		&fakeValuationRepo{
			history: snapshotHistoryFor(p.ID, snapshotPair(p.ID, today, decimal.NewFromInt(1000), decimal.NewFromInt(50), decimal.NewFromInt(10), decimal.NewFromInt(940), decimal.NewFromInt(40), decimal.NewFromInt(5))),
			listErr: wantErr,
		},
	)

	_, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.Error(t, err, "an infrastructure failure resolving the previous snapshot must not be silently treated as \"no baseline\"")
	require.ErrorIs(t, err, wantErr)
}

func TestGetValuationSummary_StalePortfolioExcludedFromInconsistentBusinessDate(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	stale := today.AddDate(0, 0, -3)

	fundFresh := newFund("THB", nil)
	fundStale := newFund("THB", nil)
	pFresh := newPortfolio(fundFresh.ID)
	pStale := newPortfolio(fundStale.ID)

	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		pFresh.ID: snapshotPair(pFresh.ID, today, decimal.NewFromInt(1000), decimal.NewFromInt(50), decimal.NewFromInt(10), decimal.NewFromInt(940), decimal.NewFromInt(40), decimal.NewFromInt(5)),
		// pStale's most recent valuation run is 3 days old — it must not be
		// summed into a total labelled with today's business date.
		pStale.ID: snapshotPair(pStale.ID, stale, decimal.NewFromInt(999999), decimal.NewFromInt(999), decimal.Zero, decimal.NewFromInt(999999), decimal.NewFromInt(999), decimal.Zero),
	}}
	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fundFresh, fundStale}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{pFresh, pStale}},
		valRepo,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.True(t, today.Equal(res.BusinessDate))
	require.True(t, decimal.NewFromInt(1000).Equal(res.AUM), "the stale portfolio's AUM must not be mixed into a total dated today, got %s", res.AUM)
	require.True(t, decimal.NewFromInt(15).Equal(res.TodayPnL), "got %s", res.TodayPnL)
}

func TestGetValuationSummary_NegativeDayPnL(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	fund := newFund("THB", nil)
	p := newPortfolio(fund.ID)
	// Yesterday's cumulative total (80) exceeds today's (60) — a losing day.
	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		p.ID: snapshotPair(p.ID, today, decimal.NewFromInt(900), decimal.NewFromInt(60), decimal.Zero, decimal.NewFromInt(1000), decimal.NewFromInt(80), decimal.Zero),
	}}
	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{p}},
		valRepo,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.True(t, decimal.NewFromInt(-20).Equal(res.TodayPnL), "got %s", res.TodayPnL)
	require.NotNil(t, res.TodayPnLPercent)
	require.True(t, res.TodayPnLPercent.IsNegative())
}

func TestGetValuationSummary_ZeroDayPnL(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	fund := newFund("THB", nil)
	p := newPortfolio(fund.ID)
	// Today's cumulative total equals yesterday's — a flat day.
	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		p.ID: snapshotPair(p.ID, today, decimal.NewFromInt(1000), decimal.NewFromInt(50), decimal.NewFromInt(10), decimal.NewFromInt(1000), decimal.NewFromInt(50), decimal.NewFromInt(10)),
	}}
	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{p}},
		valRepo,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.True(t, res.TodayPnL.IsZero(), "got %s", res.TodayPnL)
}

// snapshotHistoryFor wraps a single portfolio's pre-built snapshot slice
// (from snapshotPair) into the map fakeValuationRepo.history expects.
func snapshotHistoryFor(portfolioID uuid.UUID, snaps []*entity.ValuationSnapshot) map[uuid.UUID][]*entity.ValuationSnapshot {
	return map[uuid.UUID][]*entity.ValuationSnapshot{portfolioID: snaps}
}

// ─── fakes ──────────────────────────────────────────────────────────────────

type fakeFundRepo struct{ funds []*entity.Fund }

func (r *fakeFundRepo) Create(context.Context, pgx.Tx, *entity.Fund) error       { return nil }
func (r *fakeFundRepo) GetByID(context.Context, uuid.UUID) (*entity.Fund, error) { return nil, nil }
func (r *fakeFundRepo) GetByCode(context.Context, string) (*entity.Fund, error)  { return nil, nil }
func (r *fakeFundRepo) GetByContractCode(context.Context, string) (*entity.Fund, error) {
	return nil, nil
}
func (r *fakeFundRepo) List(_ context.Context, filter domain.FundListFilter) ([]*entity.Fund, int, error) {
	out := []*entity.Fund{}
	for _, f := range r.funds {
		if filter.Status != nil && f.Status != *filter.Status {
			continue
		}
		if filter.ManagerUserID != nil && (f.ManagerUserID == nil || *f.ManagerUserID != *filter.ManagerUserID) {
			continue
		}
		if filter.AccessibleFundIDs != nil {
			allowed := false
			for _, id := range filter.AccessibleFundIDs {
				if id == f.ID {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		out = append(out, f)
	}
	return out, len(out), nil
}
func (r *fakeFundRepo) Update(context.Context, pgx.Tx, *entity.Fund) error { return nil }
func (r *fakeFundRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r *fakeFundRepo) CountActivePortfolios(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

type fakePortfolioRepo struct{ portfolios []*entity.Portfolio }

func (r *fakePortfolioRepo) Create(context.Context, pgx.Tx, *entity.Portfolio) error { return nil }
func (r *fakePortfolioRepo) GetByID(context.Context, uuid.UUID) (*entity.Portfolio, error) {
	return nil, nil
}
func (r *fakePortfolioRepo) GetByFundCode(context.Context, uuid.UUID, string) (*entity.Portfolio, error) {
	return nil, nil
}
func (r *fakePortfolioRepo) GetByCode(context.Context, string) (*entity.Portfolio, error) {
	return nil, nil
}
func (r *fakePortfolioRepo) List(_ context.Context, filter domain.PortfolioListFilter) ([]*entity.Portfolio, int, error) {
	out := []*entity.Portfolio{}
	for _, p := range r.portfolios {
		if filter.FundID != nil && p.FundID != *filter.FundID {
			continue
		}
		if filter.Status != nil && p.Status != *filter.Status {
			continue
		}
		out = append(out, p)
	}
	return out, len(out), nil
}
func (r *fakePortfolioRepo) Update(context.Context, pgx.Tx, *entity.Portfolio) error { return nil }
func (r *fakePortfolioRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r *fakePortfolioRepo) HasOpenActivity(context.Context, uuid.UUID, time.Time) (bool, string, error) {
	return false, "", nil
}

// fakeValuationRepo's history maps portfolioID to snapshots ordered newest
// first, mirroring the real repository's ORDER BY business_date DESC.
// listErr, when set, is returned by List for every call (used to simulate
// an infrastructure failure while resolving a portfolio's prior snapshot).
type fakeValuationRepo struct {
	history map[uuid.UUID][]*entity.ValuationSnapshot
	listErr error
}

func (r *fakeValuationRepo) Insert(context.Context, pgx.Tx, *entity.ValuationSnapshot) error {
	return nil
}
func (r *fakeValuationRepo) GetLatest(_ context.Context, portfolioID uuid.UUID, source vo.ValuationSource) (*entity.ValuationSnapshot, error) {
	for _, v := range r.history[portfolioID] {
		if v.Source == source {
			return v, nil
		}
	}
	return nil, nil
}
func (r *fakeValuationRepo) GetByID(context.Context, uuid.UUID) (*entity.ValuationSnapshot, error) {
	return nil, nil
}
func (r *fakeValuationRepo) GetByPortfolioBusinessDate(context.Context, uuid.UUID, time.Time, vo.ValuationSource) (*entity.ValuationSnapshot, error) {
	return nil, nil
}
func (r *fakeValuationRepo) List(_ context.Context, portfolioID uuid.UUID, from, to time.Time, _ int, limit int) ([]*entity.ValuationSnapshot, int, error) {
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	var out []*entity.ValuationSnapshot
	for _, v := range r.history[portfolioID] {
		if (v.BusinessDate.Equal(from) || v.BusinessDate.After(from)) && (v.BusinessDate.Equal(to) || v.BusinessDate.Before(to)) {
			out = append(out, v)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, len(out), nil
}
func (r *fakeValuationRepo) InsertNAV(context.Context, pgx.Tx, *entity.NAVSnapshot) error { return nil }
func (r *fakeValuationRepo) ListNAV(context.Context, uuid.UUID, time.Time, time.Time, int, int) ([]*entity.NAVSnapshot, int, error) {
	return nil, 0, nil
}
func (r *fakeValuationRepo) InsertAUM(context.Context, pgx.Tx, *entity.AUMSnapshot) error { return nil }
func (r *fakeValuationRepo) GetAUM(context.Context, vo.AumScopeType, uuid.UUID, time.Time, vo.ValuationSource) (*entity.AUMSnapshot, error) {
	return nil, nil
}
func (r *fakeValuationRepo) ListAUM(context.Context, vo.AumScopeType, uuid.UUID, time.Time, time.Time, int, int) ([]*entity.AUMSnapshot, int, error) {
	return nil, 0, nil
}
