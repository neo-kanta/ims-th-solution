package adapter

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ValuationSummaryAdapter implements contract.ValuationSummaryProvider by
// aggregating the latest INTERNAL valuation snapshot of every portfolio
// under each in-scope, active fund.
//
// Day P&L is derived rather than stored: for each portfolio it is
//
//	(UnrealisedPnL_t + RealisedPnL_t) − (UnrealisedPnL_t-1 + RealisedPnL_t-1)
//
// RealisedPnL on a ValuationSnapshot is cumulative-to-date (see
// PostgresTransactionRepository.SumRealisedPnLBase, which sums
// business_date <= asOf), so subtracting yesterday's cumulative total from
// today's isolates exactly today's contribution and stays correct in the
// presence of cash flows (a deposit/withdrawal moves AUM and cost basis but
// not unrealised/realised P&L). When a portfolio has no prior snapshot the
// whole cumulative total is attributed to today — there is no earlier
// baseline to subtract (typically the portfolio's first valuation day).
//
// A scope can span funds in different base currencies (this system seeds
// mostly THB funds plus a USD thematic fund). Rather than invent an FX
// conversion here, the adapter aggregates only the funds sharing the
// *predominant* currency within the resolved scope and reports that value
// as Currency; funds in a different currency are excluded from the sum.
type ValuationSummaryAdapter struct {
	funds      domain.FundRepository
	portfolios domain.PortfolioRepository
	valuation  domain.ValuationRepository
}

// NewValuationSummaryAdapter wires the adapter.
func NewValuationSummaryAdapter(
	funds domain.FundRepository,
	portfolios domain.PortfolioRepository,
	valuation domain.ValuationRepository,
) *ValuationSummaryAdapter {
	return &ValuationSummaryAdapter{funds: funds, portfolios: portfolios, valuation: valuation}
}

// GetValuationSummary implements contract.ValuationSummaryProvider.
func (a *ValuationSummaryAdapter) GetValuationSummary(
	ctx context.Context,
	req contract.ValuationSummaryRequest,
) (*contract.ValuationSummaryResult, error) {
	if a == nil {
		return nil, fmt.Errorf("valuation summary adapter not initialised")
	}

	activeStatus := vo.FundStatusActive
	filter := domain.FundListFilter{
		Status:            &activeStatus,
		AccessibleFundIDs: req.AccessibleFundIDs,
		Limit:             10_000,
	}
	if req.Scope == contract.ValuationSummaryScopeMine {
		uid := req.UserID
		filter.ManagerUserID = &uid
	}

	funds, _, err := a.funds.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing scoped funds: %w", err)
	}
	if len(funds) == 0 {
		return &contract.ValuationSummaryResult{DataAvailable: false}, nil
	}

	currency := predominantCurrency(funds)
	activePortfolio := vo.PortfolioStatusActive

	// First pass: collect every in-scope, currency-matching portfolio's
	// latest INTERNAL snapshot without aggregating yet. Portfolios are
	// valued independently (the nightly valuation run can finish for some
	// portfolios before others, or skip a portfolio with no activity), so
	// summing them blind would silently mix snapshots from different
	// business dates under one reported business_date/as_of label.
	var latestSnapshots []*entity.ValuationSnapshot
	for _, f := range funds {
		if f.BaseCurrency != currency {
			continue
		}
		portfolios, _, err := a.portfolios.List(ctx, domain.PortfolioListFilter{
			FundID: &f.ID,
			Status: &activePortfolio,
			Limit:  10_000,
		})
		if err != nil {
			return nil, fmt.Errorf("listing portfolios for fund %s: %w", f.ID, err)
		}

		for _, p := range portfolios {
			latest, err := a.valuation.GetLatest(ctx, p.ID, vo.ValuationSourceInternal)
			if err != nil {
				return nil, fmt.Errorf("loading latest valuation for portfolio %s: %w", p.ID, err)
			}
			if latest == nil {
				continue
			}
			latestSnapshots = append(latestSnapshots, latest)
		}
	}

	if len(latestSnapshots) == 0 {
		return &contract.ValuationSummaryResult{Currency: currency, DataAvailable: false}, nil
	}

	// The reported business date is the most recent one any in-scope
	// portfolio was valued on. Only portfolios whose latest snapshot is
	// exactly on that date are aggregated; a portfolio last valued earlier
	// is stale relative to the scope and is excluded rather than folded
	// into a total labelled with a newer date.
	businessDate := latestSnapshots[0].BusinessDate
	for _, s := range latestSnapshots {
		if s.BusinessDate.After(businessDate) {
			businessDate = s.BusinessDate
		}
	}

	aum := decimal.Zero
	prevAUM := decimal.Zero
	todayPnL := decimal.Zero
	haveBaseline := true
	includedCount := 0
	var asOf time.Time

	for _, latest := range latestSnapshots {
		if !latest.BusinessDate.Equal(businessDate) {
			continue
		}
		includedCount++
		aum = aum.Add(latest.AUM)
		if latest.CreatedAt.After(asOf) {
			asOf = latest.CreatedAt
		}

		todayTotal := latest.UnrealisedPnL.Add(latest.RealisedPnL)
		prev, err := a.previousInternalSnapshot(ctx, latest.PortfolioID, latest.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("loading previous valuation for portfolio %s: %w", latest.PortfolioID, err)
		}
		if prev != nil {
			prevTotal := prev.UnrealisedPnL.Add(prev.RealisedPnL)
			todayPnL = todayPnL.Add(todayTotal.Sub(prevTotal))
			prevAUM = prevAUM.Add(prev.AUM)
		} else {
			todayPnL = todayPnL.Add(todayTotal)
			haveBaseline = false
		}
	}

	if includedCount == 0 {
		return &contract.ValuationSummaryResult{Currency: currency, DataAvailable: false}, nil
	}

	result := &contract.ValuationSummaryResult{
		Currency:      currency,
		BusinessDate:  businessDate,
		AsOf:          asOf,
		AUM:           aum,
		TodayPnL:      todayPnL,
		DataAvailable: true,
		FundCount:     len(funds),
	}

	// Percent base is the prior day's aggregate AUM (the value the day's
	// P&L is a return *on*). Fall back to today's AUM when there is no
	// usable baseline (first valuation day, or a zero/negative prior AUM).
	denom := prevAUM
	if !haveBaseline || denom.Sign() <= 0 {
		denom = aum
	}
	if denom.Sign() > 0 {
		pct := todayPnL.Div(denom).Mul(decimal.NewFromInt(100)).Round(4)
		result.TodayPnLPercent = &pct
	}

	return result, nil
}

// previousInternalSnapshot returns the most recent INTERNAL valuation
// snapshot strictly before businessDate, or nil when none exists.
//
// A repository error is returned to the caller rather than treated as "no
// previous snapshot" — conflating the two would attribute a portfolio's
// entire cumulative P&L to today whenever the lookup merely failed,
// silently producing a materially wrong number instead of failing loudly.
//
// ValuationRepository.List does not filter by source (it only bounds by
// portfolio + date range), so a handful of rows are pulled and the first
// INTERNAL one is picked — cheap in practice since a portfolio gets at most
// one valuation run per business date.
func (a *ValuationSummaryAdapter) previousInternalSnapshot(ctx context.Context, portfolioID uuid.UUID, businessDate time.Time) (*entity.ValuationSnapshot, error) {
	from := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	to := businessDate.AddDate(0, 0, -1)
	list, _, err := a.valuation.List(ctx, portfolioID, from, to, 1, 5)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if v.Source == vo.ValuationSourceInternal {
			return v, nil
		}
	}
	return nil, nil
}

// predominantCurrency returns the most common Fund.BaseCurrency in funds,
// breaking ties alphabetically for determinism.
func predominantCurrency(funds []*entity.Fund) string {
	counts := map[string]int{}
	for _, f := range funds {
		counts[f.BaseCurrency]++
	}
	best := ""
	bestCount := -1
	for ccy, c := range counts {
		if c > bestCount || (c == bestCount && ccy < best) {
			best = ccy
			bestCount = c
		}
	}
	return best
}

var _ contract.ValuationSummaryProvider = (*ValuationSummaryAdapter)(nil)
