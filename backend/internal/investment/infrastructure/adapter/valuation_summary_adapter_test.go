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
	return &entity.Portfolio{
		ID:                uuid.New(),
		FundID:            &fundID,
		PortfolioType:     vo.PortfolioTypeLive,
		Status:            vo.PortfolioStatusActive,
		BaseCurrency:      "THB",
		ValuationCurrency: "THB",
	}
}

func newTestValuationSummaryAdapter(
	funds domain.FundRepository,
	portfolios domain.PortfolioRepository,
	valuation domain.ValuationRepository,
) *ValuationSummaryAdapter {
	return NewValuationSummaryAdapter(funds, portfolios, valuation, "THB", nil)
}

// snapshotPair seeds today's and yesterday's INTERNAL snapshot for a
// portfolio (newest first, matching the real repository's ORDER BY
// business_date DESC).
func snapshotPair(portfolioID uuid.UUID, today time.Time, todayAUM, todayUPnL, todayRPnL, prevAUM, prevUPnL, prevRPnL decimal.Decimal) []*entity.ValuationSnapshot {
	return []*entity.ValuationSnapshot{
		{
			PortfolioID: portfolioID, BusinessDate: today, AUM: todayAUM,
			ValuationCcy: "THB", UnrealisedPnL: todayUPnL, RealisedPnL: todayRPnL,
			Source: vo.ValuationSourceInternal, CreatedAt: today,
		},
		{
			PortfolioID: portfolioID, BusinessDate: today.AddDate(0, 0, -1), AUM: prevAUM,
			ValuationCcy: "THB", UnrealisedPnL: prevUPnL, RealisedPnL: prevRPnL,
			Source: vo.ValuationSourceInternal, CreatedAt: today.AddDate(0, 0, -1),
		},
	}
}

func withValuationCurrency(snapshots []*entity.ValuationSnapshot, currency string) []*entity.ValuationSnapshot {
	for _, snapshot := range snapshots {
		snapshot.ValuationCcy = currency
	}
	return snapshots
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
	a := newTestValuationSummaryAdapter(
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
	require.Equal(t, 2, res.Coverage.IncludedFundCount)
	require.NotNil(t, res.TodayPnLPercent)
}

func TestGetValuationSummary_MineScopeFiltersByPortfolioManager(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	userA := uuid.New()
	userB := uuid.New()

	// Deliberately cross fund and portfolio ownership. Mine scope must follow
	// the portfolio manager, not the fund manager.
	fundMine := newFund("THB", &userB)
	fundOther := newFund("THB", &userA)
	pMine := newPortfolio(fundMine.ID)
	pOther := newPortfolio(fundOther.ID)
	pMine.ManagerUserID = &userA
	pOther.ManagerUserID = &userB

	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		pMine.ID:  snapshotPair(pMine.ID, today, decimal.NewFromInt(500), decimal.NewFromInt(5), decimal.Zero, decimal.NewFromInt(490), decimal.NewFromInt(3), decimal.Zero),
		pOther.ID: snapshotPair(pOther.ID, today, decimal.NewFromInt(9000), decimal.NewFromInt(500), decimal.Zero, decimal.NewFromInt(8000), decimal.NewFromInt(100), decimal.Zero),
	}}
	a := newTestValuationSummaryAdapter(
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
	require.Equal(t, 1, res.Coverage.TotalFundCount)
	require.Equal(t, 1, res.Coverage.TotalPortfolioCount)
	require.Equal(t, 1, res.Coverage.IncludedFundCount)
	require.True(t, decimal.NewFromInt(500).Equal(res.AUM), "must only include portfolios managed by the caller, got %s", res.AUM)
}

func TestGetValuationSummary_CompanyScopeIgnoresAccessibleFundFilter(t *testing.T) {
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
	a := newTestValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fundAllowed, fundBlocked}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{pAllowed, pBlocked}},
		valRepo,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{
		Scope:             contract.ValuationSummaryScopeCompany,
		AccessibleFundIDs: []uuid.UUID{fundAllowed.ID}, // ignored for authoritative company scope
	})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.Equal(t, 2, res.Coverage.IncludedFundCount)
	require.True(t, decimal.NewFromInt(1000099).Equal(res.AUM), "company scope must ignore AccessibleFundIDs, got %s", res.AUM)
}

func TestGetValuationSummary_MineScopeIntersectsManagerAndAccessibleFunds(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	manager := uuid.New()
	otherManager := uuid.New()

	fundIncluded := newFund("THB", &manager)
	fundNotAccessible := newFund("THB", &manager)
	fundNotManaged := newFund("THB", &otherManager)
	pIncluded := newPortfolio(fundIncluded.ID)
	pNotAccessible := newPortfolio(fundNotAccessible.ID)
	pNotManaged := newPortfolio(fundNotManaged.ID)
	pIncluded.ManagerUserID = &manager
	pNotAccessible.ManagerUserID = &manager
	pNotManaged.ManagerUserID = &otherManager

	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		pIncluded.ID:      snapshotPair(pIncluded.ID, today, decimal.NewFromInt(100), decimal.Zero, decimal.Zero, decimal.NewFromInt(100), decimal.Zero, decimal.Zero),
		pNotAccessible.ID: snapshotPair(pNotAccessible.ID, today, decimal.NewFromInt(200), decimal.Zero, decimal.Zero, decimal.NewFromInt(200), decimal.Zero, decimal.Zero),
		pNotManaged.ID:    snapshotPair(pNotManaged.ID, today, decimal.NewFromInt(300), decimal.Zero, decimal.Zero, decimal.NewFromInt(300), decimal.Zero, decimal.Zero),
	}}
	a := newTestValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fundIncluded, fundNotAccessible, fundNotManaged}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{pIncluded, pNotAccessible, pNotManaged}},
		valRepo,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{
		Scope:             contract.ValuationSummaryScopeMine,
		UserID:            manager,
		AccessibleFundIDs: []uuid.UUID{fundIncluded.ID, fundNotManaged.ID},
	})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.Equal(t, 1, res.Coverage.TotalFundCount)
	require.Equal(t, 1, res.Coverage.TotalPortfolioCount)
	require.Equal(t, 1, res.Coverage.IncludedFundCount)
	require.True(t, decimal.NewFromInt(100).Equal(res.AUM), "mine scope must intersect portfolio-manager ownership and fund access, got %s", res.AUM)
}

func TestGetValuationSummary_ExhaustsFundAndPortfolioPages(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	fundA := newFund("THB", nil)
	fundB := newFund("THB", nil)
	pA1 := newPortfolio(fundA.ID)
	pA2 := newPortfolio(fundA.ID)
	pB1 := newPortfolio(fundB.ID)
	pB2 := newPortfolio(fundB.ID)
	portfolios := []*entity.Portfolio{pA1, pA2, pB1, pB2}
	history := make(map[uuid.UUID][]*entity.ValuationSnapshot, len(portfolios))
	for _, portfolio := range portfolios {
		history[portfolio.ID] = snapshotPair(portfolio.ID, today, decimal.NewFromInt(100), decimal.Zero, decimal.Zero, decimal.NewFromInt(100), decimal.Zero, decimal.Zero)
	}

	a := newTestValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fundA, fundB}, pageSize: 1},
		&fakePortfolioRepo{portfolios: portfolios, pageSize: 1},
		&fakeValuationRepo{history: history},
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{
		Scope: contract.ValuationSummaryScopeCompany,
	})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.Equal(t, 2, res.Coverage.TotalFundCount)
	require.Equal(t, 4, res.Coverage.TotalPortfolioCount)
	require.True(t, decimal.NewFromInt(400).Equal(res.AUM), "all paginated valuations must be included, got %s", res.AUM)
}

func TestGetValuationSummary_NoDataReturnsUnavailable(t *testing.T) {
	t.Parallel()

	t.Run("no funds in scope", func(t *testing.T) {
		t.Parallel()
		a := newTestValuationSummaryAdapter(
			&fakeFundRepo{},
			&fakePortfolioRepo{},
			&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{}},
		)
		res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{
			Scope: contract.ValuationSummaryScopeCompany,
		})
		require.NoError(t, err)
		require.False(t, res.DataAvailable)
		require.Equal(t, contract.ValuationSummaryStatusNoData, res.Status)
		require.Equal(t, 0, res.Coverage.TotalFundCount)
	})

	t.Run("funds exist but no valuation snapshot yet", func(t *testing.T) {
		t.Parallel()
		fund := newFund("THB", nil)
		p := newPortfolio(fund.ID)
		a := newTestValuationSummaryAdapter(
			&fakeFundRepo{funds: []*entity.Fund{fund}},
			&fakePortfolioRepo{portfolios: []*entity.Portfolio{p}},
			&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{}},
		)
		res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{
			Scope: contract.ValuationSummaryScopeCompany,
		})
		require.NoError(t, err)
		require.False(t, res.DataAvailable)
		require.Equal(t, contract.ValuationSummaryStatusIncomplete, res.Status)
		require.True(t, res.AUM.IsZero(), "unavailable result must not report a numeric AUM")
		require.Equal(t, 1, res.Coverage.TotalFundCount)
		require.Equal(t, 0, res.Coverage.IncludedFundCount)
		require.Equal(t, 1, res.Coverage.ExcludedFundCount)
		require.Equal(t, 1, res.Coverage.TotalPortfolioCount)
		require.Equal(t, 0, res.Coverage.IncludedPortfolioCount)
		require.Equal(t, 1, res.Coverage.ExcludedPortfolioCount)
		require.Equal(t, []contract.ValuationSummaryExclusionReason{contract.ValuationSummaryExclusionMissingValuation}, res.Coverage.ExclusionReasons)
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
		ValuationCcy:  "THB",
		UnrealisedPnL: decimal.NewFromInt(60), RealisedPnL: decimal.NewFromInt(15),
		Source: vo.ValuationSourceInternal, CreatedAt: today,
	}
	a := newTestValuationSummaryAdapter(
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

func TestGetValuationSummary_MixedCurrencyConvertsAtLatestFX(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	thbFundA := newFund("THB", nil)
	thbFundB := newFund("THB", nil)
	usdFund := newFund("USD", nil)
	pThbA := newPortfolio(thbFundA.ID)
	pThbB := newPortfolio(thbFundB.ID)
	pUsd := &entity.Portfolio{ID: uuid.New(), FundID: &usdFund.ID, PortfolioType: vo.PortfolioTypeLive, Status: vo.PortfolioStatusActive, BaseCurrency: "USD", ValuationCurrency: "USD"}

	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		pThbA.ID: snapshotPair(pThbA.ID, today, decimal.NewFromInt(100), decimal.Zero, decimal.Zero, decimal.NewFromInt(100), decimal.Zero, decimal.Zero),
		pThbB.ID: snapshotPair(pThbB.ID, today, decimal.NewFromInt(200), decimal.Zero, decimal.Zero, decimal.NewFromInt(200), decimal.Zero, decimal.Zero),
		pUsd.ID: withValuationCurrency(
			snapshotPair(pUsd.ID, today, decimal.NewFromInt(10), decimal.NewFromInt(2), decimal.Zero, decimal.NewFromInt(8), decimal.NewFromInt(1), decimal.Zero),
			"USD",
		),
	}}
	fx := &fakeQuoteProvider{latestQuotes: map[string]*contract.MarketQuote{
		"FX_USDTHB": exactFXQuote("FX_USDTHB", "THB", today, decimal.NewFromInt(36)),
	}}
	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{thbFundA, thbFundB, usdFund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{pThbA, pThbB, pUsd}},
		valRepo,
		"THB",
		fx,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.Equal(t, contract.ValuationSummaryStatusAvailable, res.Status)
	require.Equal(t, "THB", res.Currency)
	require.True(t, decimal.NewFromInt(660).Equal(res.AUM), "got %s", res.AUM)
	require.True(t, decimal.NewFromInt(36).Equal(res.TodayPnL), "got %s", res.TodayPnL)
	require.NotNil(t, res.TodayPnLPercent)
	require.True(t, decimal.RequireFromString("6.1224").Equal(*res.TodayPnLPercent), "got %s", res.TodayPnLPercent)
	require.Equal(t, 3, res.Coverage.TotalFundCount)
	require.Equal(t, 3, res.Coverage.IncludedFundCount)
	require.Equal(t, 3, res.Coverage.TotalPortfolioCount)
	require.Equal(t, 3, res.Coverage.IncludedPortfolioCount)
	require.Empty(t, res.Coverage.Exclusions)
	require.Equal(t, []string{"FX_USDTHB"}, fx.latestCalls, "one latest FX lookup must serve current and previous values")
}

func TestGetValuationSummary_LatestFXDoesNotInventHistoricalFXMovement(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	fund := newFund("USD", nil)
	portfolio := &entity.Portfolio{
		ID:                uuid.New(),
		FundID:            &fund.ID,
		PortfolioType:     vo.PortfolioTypeLive,
		Status:            vo.PortfolioStatusActive,
		BaseCurrency:      "USD",
		ValuationCurrency: "USD",
	}
	valuations := withValuationCurrency(
		snapshotPair(portfolio.ID, today, decimal.NewFromInt(100), decimal.Zero, decimal.Zero, decimal.NewFromInt(100), decimal.Zero, decimal.Zero),
		"USD",
	)
	fx := &fakeQuoteProvider{latestQuotes: map[string]*contract.MarketQuote{
		"FX_USDTHB": exactFXQuote("FX_USDTHB", "THB", today, decimal.NewFromInt(36)),
	}}
	a := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{portfolio}},
		&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{portfolio.ID: valuations}},
		"THB",
		fx,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.Equal(t, contract.ValuationSummaryStatusAvailable, res.Status)
	require.True(t, decimal.NewFromInt(3600).Equal(res.AUM), "got AUM %s", res.AUM)
	require.True(t, res.TodayPnL.IsZero(), "one latest rate must not invent unavailable historical FX movement; got %s", res.TodayPnL)
	require.NotNil(t, res.TodayPnLPercent)
	require.True(t, res.TodayPnLPercent.IsZero(), "got %s", res.TodayPnLPercent)
}

func TestGetValuationSummary_ExternalFlowsUseClosingFXConvention(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	for _, tt := range []struct {
		name       string
		currentAUM decimal.Decimal
	}{
		{name: "cash in", currentAUM: decimal.NewFromInt(125)},
		{name: "cash out", currentAUM: decimal.NewFromInt(75)},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fund := newFund("USD", nil)
			portfolio := &entity.Portfolio{
				ID:                uuid.New(),
				FundID:            &fund.ID,
				PortfolioType:     vo.PortfolioTypeLive,
				Status:            vo.PortfolioStatusActive,
				BaseCurrency:      "USD",
				ValuationCurrency: "USD",
			}
			// Cumulative local P&L is unchanged while AUM moves by an external
			// flow. The flow is excluded at today's closing FX, leaving only
			// the FX movement on the USD 100 opening net assets.
			valuations := withValuationCurrency(
				snapshotPair(portfolio.ID, today, tt.currentAUM, decimal.NewFromInt(10), decimal.Zero, decimal.NewFromInt(100), decimal.NewFromInt(10), decimal.Zero),
				"USD",
			)
			fx := &fakeQuoteProvider{latestQuotes: map[string]*contract.MarketQuote{
				"FX_USDTHB": exactFXQuote("FX_USDTHB", "THB", today, decimal.NewFromInt(36)),
			}}
			a := NewValuationSummaryAdapter(
				&fakeFundRepo{funds: []*entity.Fund{fund}},
				&fakePortfolioRepo{portfolios: []*entity.Portfolio{portfolio}},
				&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{portfolio.ID: valuations}},
				"THB",
				fx,
			)

			res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
			require.NoError(t, err)
			require.Equal(t, contract.ValuationSummaryStatusAvailable, res.Status)
			require.True(t, res.TodayPnL.IsZero(), "flow must not be reported as P&L; got %s", res.TodayPnL)
			require.NotNil(t, res.TodayPnLPercent)
			require.True(t, res.TodayPnLPercent.IsZero(), "got %s", res.TodayPnLPercent)
		})
	}
}

func TestGetValuationSummary_PreviousSnapshotErrorPropagates(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	fund := newFund("THB", nil)
	p := newPortfolio(fund.ID)
	wantErr := errors.New("db down")

	a := newTestValuationSummaryAdapter(
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

func TestGetValuationSummary_InconsistentBusinessDateUsesLatestAvailable(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	stale := today.AddDate(0, 0, -3)

	fundFresh := newFund("THB", nil)
	fundStale := newFund("THB", nil)
	pFresh := newPortfolio(fundFresh.ID)
	pStale := newPortfolio(fundStale.ID)

	valRepo := &fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
		pFresh.ID: snapshotPair(pFresh.ID, today, decimal.NewFromInt(1000), decimal.NewFromInt(50), decimal.NewFromInt(10), decimal.NewFromInt(940), decimal.NewFromInt(40), decimal.NewFromInt(5)),
		// pStale's most recent valuation run is 3 days old and is included under
		// the latest-available owner policy.
		pStale.ID: snapshotPair(pStale.ID, stale, decimal.NewFromInt(999999), decimal.NewFromInt(999), decimal.Zero, decimal.NewFromInt(999999), decimal.NewFromInt(999), decimal.Zero),
	}}
	a := newTestValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fundFresh, fundStale}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{pFresh, pStale}},
		valRepo,
	)

	res, err := a.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.Equal(t, contract.ValuationSummaryStatusAvailable, res.Status)
	require.True(t, today.Equal(res.BusinessDate))
	require.True(t, decimal.NewFromInt(1000999).Equal(res.AUM), "latest available snapshots must all contribute")
	require.Equal(t, 2, res.Coverage.TotalFundCount)
	require.Equal(t, 2, res.Coverage.IncludedFundCount)
	require.Equal(t, 0, res.Coverage.ExcludedFundCount)
	require.Equal(t, 2, res.Coverage.TotalPortfolioCount)
	require.Equal(t, 2, res.Coverage.IncludedPortfolioCount)
	require.Equal(t, 0, res.Coverage.ExcludedPortfolioCount)
	require.Equal(t, 1, res.Coverage.LatestAvailablePortfolioCount)
	require.True(t, stale.Equal(res.Coverage.OldestIncludedBusinessDate))
	require.Empty(t, res.Coverage.ExclusionReasons)
	require.Empty(t, res.Coverage.Exclusions)
}

func TestGetValuationSummary_ValuationCurrencyChangeIsIncomplete(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	fund := newFund("USD", nil)
	portfolio := &entity.Portfolio{
		ID:                uuid.New(),
		FundID:            &fund.ID,
		PortfolioType:     vo.PortfolioTypeLive,
		Status:            vo.PortfolioStatusActive,
		BaseCurrency:      "USD",
		ValuationCurrency: "USD",
	}
	valuations := snapshotPair(portfolio.ID, today, decimal.NewFromInt(100), decimal.NewFromInt(5), decimal.Zero, decimal.NewFromInt(90), decimal.NewFromInt(4), decimal.Zero)
	valuations[0].ValuationCcy = "USD"
	valuations[1].ValuationCcy = "EUR"
	adapter := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{portfolio}},
		&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{portfolio.ID: valuations}},
		"THB",
		&fakeQuoteProvider{},
	)

	res, err := adapter.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.Equal(t, contract.ValuationSummaryStatusIncomplete, res.Status)
	require.False(t, res.DataAvailable)
	require.True(t, res.AUM.IsZero())
	require.True(t, res.TodayPnL.IsZero())
	require.Equal(t, []contract.ValuationSummaryExclusionReason{contract.ValuationSummaryExclusionValuationCurrency}, res.Coverage.ExclusionReasons)
	require.Len(t, res.Coverage.Exclusions, 1)
	require.Equal(t, "EUR", res.Coverage.Exclusions[0].Currency)
}

func TestGetValuationSummary_FXFailuresAreTypedIncompleteWithoutPartialTotals(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		quote  *contract.MarketQuote
		err    error
		reason contract.ValuationSummaryExclusionReason
	}{
		{name: "provider error", err: contract.ErrMarketDataUnavailable, reason: contract.ValuationSummaryExclusionMissingFX},
		{name: "nil quote", reason: contract.ValuationSummaryExclusionMissingFX},
		{name: "wrong symbol", quote: exactFXQuote("FX_THBUSD", "THB", today, decimal.NewFromInt(36)), reason: contract.ValuationSummaryExclusionFXSymbol},
		{name: "quote currency mismatch", quote: exactFXQuote("FX_USDTHB", "USD", today, decimal.NewFromInt(36)), reason: contract.ValuationSummaryExclusionFXCurrency},
		{name: "zero rate", quote: exactFXQuote("FX_USDTHB", "THB", today, decimal.Zero), reason: contract.ValuationSummaryExclusionInvalidFX},
		{name: "negative rate", quote: exactFXQuote("FX_USDTHB", "THB", today, decimal.NewFromInt(-36)), reason: contract.ValuationSummaryExclusionInvalidFX},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fund := newFund("USD", nil)
			fund.Code = "USD-FUND"
			portfolio := newPortfolio(fund.ID)
			portfolio.Code = "USD-LIVE"
			portfolio.ValuationCurrency = "USD"
			provider := &fakeQuoteProvider{
				latestQuotes: map[string]*contract.MarketQuote{"FX_USDTHB": tt.quote},
				latestErrs:   map[string]error{"FX_USDTHB": tt.err},
			}
			adapter := NewValuationSummaryAdapter(
				&fakeFundRepo{funds: []*entity.Fund{fund}},
				&fakePortfolioRepo{portfolios: []*entity.Portfolio{portfolio}},
				&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
					portfolio.ID: withValuationCurrency(snapshotPair(portfolio.ID, today, decimal.NewFromInt(10), decimal.NewFromInt(1), decimal.Zero, decimal.NewFromInt(9), decimal.Zero, decimal.Zero), "USD"),
				}},
				"THB",
				provider,
			)

			res, err := adapter.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
			require.NoError(t, err)
			require.False(t, res.DataAvailable)
			require.Equal(t, contract.ValuationSummaryStatusIncomplete, res.Status)
			require.True(t, res.AUM.IsZero())
			require.True(t, res.TodayPnL.IsZero())
			require.Nil(t, res.TodayPnLPercent)
			require.Equal(t, []contract.ValuationSummaryExclusionReason{tt.reason}, res.Coverage.ExclusionReasons)
			require.Equal(t, []string{"USD"}, res.Coverage.ExcludedCurrencies)
			require.Equal(t, 1, res.Coverage.TotalFundCount)
			require.Equal(t, 0, res.Coverage.IncludedFundCount)
			require.Equal(t, 1, res.Coverage.ExcludedFundCount)
			require.Equal(t, 1, res.Coverage.TotalPortfolioCount)
			require.Equal(t, 0, res.Coverage.IncludedPortfolioCount)
			require.Equal(t, 1, res.Coverage.ExcludedPortfolioCount)
			require.Len(t, res.Coverage.Exclusions, 1)
			require.Equal(t, "USD-FUND", res.Coverage.Exclusions[0].FundCode)
			require.Equal(t, "USD-LIVE", res.Coverage.Exclusions[0].PortfolioCode)
			require.Equal(t, []string{"FX_USDTHB"}, provider.latestCalls)
		})
	}
}

func TestGetValuationSummary_LatestFXMayBeOlderOrCached(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	fund := newFund("USD", nil)
	portfolio := newPortfolio(fund.ID)
	latest := exactFXQuote("FX_USDTHB", "THB", today.AddDate(0, 0, -1), decimal.NewFromInt(36))
	latest.Stale = true
	provider := &fakeQuoteProvider{latestQuotes: map[string]*contract.MarketQuote{"FX_USDTHB": latest}}
	adapter := NewValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{portfolio}},
		&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
			portfolio.ID: withValuationCurrency(snapshotPair(portfolio.ID, today, decimal.NewFromInt(10), decimal.NewFromInt(1), decimal.Zero, decimal.NewFromInt(9), decimal.Zero, decimal.Zero), "USD"),
		}},
		"THB",
		provider,
	)

	res, err := adapter.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.Equal(t, contract.ValuationSummaryStatusAvailable, res.Status)
	require.True(t, res.DataAvailable)
	require.True(t, decimal.NewFromInt(360).Equal(res.AUM))
	require.Equal(t, []string{"FX_USDTHB"}, provider.latestCalls)
}

func TestGetValuationSummary_StaleValuationUsesLatestAvailable(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	fund := newFund("THB", nil)
	portfolio := newPortfolio(fund.ID)
	snapshots := snapshotPair(portfolio.ID, today, decimal.NewFromInt(100), decimal.Zero, decimal.Zero, decimal.NewFromInt(99), decimal.Zero, decimal.Zero)
	snapshots[0].HasStaleInputs = true
	adapter := newTestValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{portfolio}},
		&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{portfolio.ID: snapshots}},
	)

	res, err := adapter.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.Equal(t, contract.ValuationSummaryStatusAvailable, res.Status)
	require.True(t, res.DataAvailable)
	require.True(t, decimal.NewFromInt(100).Equal(res.AUM))
	require.Equal(t, 1, res.Coverage.LatestAvailablePortfolioCount)
	require.True(t, today.Equal(res.Coverage.OldestIncludedBusinessDate))
	require.Empty(t, res.Coverage.ExclusionReasons)
}

func TestGetValuationSummary_NonOfficialPortfoliosAreExcludedFromOfficialAUM(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	fund := newFund("THB", nil)
	live := newPortfolio(fund.ID)
	simulation := newPortfolio(fund.ID)
	simulation.PortfolioType = vo.PortfolioTypeSimulation
	model := newPortfolio(fund.ID)
	model.PortfolioType = vo.PortfolioTypeModel
	adapter := newTestValuationSummaryAdapter(
		&fakeFundRepo{funds: []*entity.Fund{fund}},
		&fakePortfolioRepo{portfolios: []*entity.Portfolio{live, simulation, model}},
		&fakeValuationRepo{history: map[uuid.UUID][]*entity.ValuationSnapshot{
			live.ID:       snapshotPair(live.ID, today, decimal.NewFromInt(100), decimal.Zero, decimal.Zero, decimal.NewFromInt(100), decimal.Zero, decimal.Zero),
			simulation.ID: snapshotPair(simulation.ID, today, decimal.NewFromInt(900), decimal.Zero, decimal.Zero, decimal.NewFromInt(900), decimal.Zero, decimal.Zero),
			model.ID:      snapshotPair(model.ID, today, decimal.NewFromInt(9000), decimal.Zero, decimal.Zero, decimal.NewFromInt(9000), decimal.Zero, decimal.Zero),
		}},
	)

	res, err := adapter.GetValuationSummary(context.Background(), contract.ValuationSummaryRequest{Scope: contract.ValuationSummaryScopeCompany})
	require.NoError(t, err)
	require.Equal(t, contract.ValuationSummaryStatusAvailable, res.Status)
	require.True(t, res.DataAvailable)
	require.True(t, decimal.NewFromInt(100).Equal(res.AUM))
	require.Equal(t, 1, res.Coverage.TotalFundCount)
	require.Equal(t, 1, res.Coverage.IncludedFundCount)
	require.Equal(t, 3, res.Coverage.TotalPortfolioCount)
	require.Equal(t, 1, res.Coverage.IncludedPortfolioCount)
	require.Equal(t, 2, res.Coverage.ExcludedPortfolioCount)
	require.Equal(t, []contract.ValuationSummaryExclusionReason{contract.ValuationSummaryExclusionNonOfficial}, res.Coverage.ExclusionReasons)
	require.Len(t, res.Coverage.Exclusions, 2)
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
	a := newTestValuationSummaryAdapter(
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
	a := newTestValuationSummaryAdapter(
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

type fakeFundRepo struct {
	funds    []*entity.Fund
	pageSize int
}

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
	total := len(out)
	return paginateTestSlice(out, filter.Page, filter.Limit, r.pageSize), total, nil
}
func (r *fakeFundRepo) Update(context.Context, pgx.Tx, *entity.Fund) error { return nil }
func (r *fakeFundRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r *fakeFundRepo) CountActivePortfolios(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

type fakePortfolioRepo struct {
	portfolios []*entity.Portfolio
	pageSize   int
}

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
		if filter.FundID != nil && (p.FundID == nil || *p.FundID != *filter.FundID) {
			continue
		}
		if filter.Status != nil && p.Status != *filter.Status {
			continue
		}
		if filter.ManagerUserID != nil && (p.ManagerUserID == nil || *p.ManagerUserID != *filter.ManagerUserID) {
			continue
		}
		if filter.AccessibleFundIDs != nil {
			allowed := false
			for _, id := range filter.AccessibleFundIDs {
				if p.FundID != nil && id == *p.FundID {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		out = append(out, p)
	}
	total := len(out)
	return paginateTestSlice(out, filter.Page, filter.Limit, r.pageSize), total, nil
}
func (r *fakePortfolioRepo) Update(context.Context, pgx.Tx, *entity.Portfolio) error { return nil }
func (r *fakePortfolioRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r *fakePortfolioRepo) HasOpenActivity(context.Context, uuid.UUID, time.Time) (bool, string, error) {
	return false, "", nil
}

func paginateTestSlice[T any](items []T, page, requestedLimit, forcedPageSize int) []T {
	limit := requestedLimit
	if forcedPageSize > 0 && (limit <= 0 || forcedPageSize < limit) {
		limit = forcedPageSize
	}
	if limit <= 0 {
		limit = len(items)
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * limit
	if start >= len(items) {
		return []T{}
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

// fakeValuationRepo's history maps portfolioID to snapshots ordered newest
// first, mirroring the real repository's ORDER BY business_date DESC.
// listErr, when set, is returned by List for every call (used to simulate
// an infrastructure failure while resolving a portfolio's prior snapshot).
type fakeValuationRepo struct {
	history map[uuid.UUID][]*entity.ValuationSnapshot
	listErr error
}

type fakeQuoteProvider struct {
	latestQuotes map[string]*contract.MarketQuote
	latestErrs   map[string]error
	latestCalls  []string
}

func exactFXQuote(symbol, currency string, businessDate time.Time, rate decimal.Decimal) *contract.MarketQuote {
	return &contract.MarketQuote{
		Symbol:      symbol,
		Price:       rate,
		Currency:    currency,
		EffectiveAt: businessDate,
		FetchedAt:   businessDate,
		Source:      contract.QuoteSourceMarketDataSnapshot,
	}
}

func (p *fakeQuoteProvider) GetLatestQuote(_ context.Context, symbol string) (*contract.MarketQuote, error) {
	p.latestCalls = append(p.latestCalls, symbol)
	if err := p.latestErrs[symbol]; err != nil {
		return nil, err
	}
	quote := p.latestQuotes[symbol]
	if quote == nil {
		return nil, nil
	}
	copy := *quote
	return &copy, nil
}

func (p *fakeQuoteProvider) GetQuoteAsOf(context.Context, string, time.Time) (*contract.MarketQuote, error) {
	return nil, contract.ErrMarketDataUnavailable
}

func (p *fakeQuoteProvider) PrimaryProviderName() string    { return "fake" }
func (p *fakeQuoteProvider) ProviderConfigured(string) bool { return true }

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
