package adapter

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ValuationSummaryAdapter implements contract.ValuationSummaryProvider by
// aggregating authoritative INTERNAL valuation snapshots into the configured
// reporting currency.
//
// Latest-snapshot P&L is derived in reporting currency for each portfolio:
//
//	(LocalCumulativePnL_latest - LocalCumulativePnL_previous) * FX_latest
//
// The newest available FX observation is used consistently for both the latest
// and previous local snapshots. This produces a current reporting-currency
// translation without inventing historical FX movement when only the latest
// rate is requested. External cash flows therefore remain outside P&L because
// the calculation differences cumulative valuation P&L rather than AUM.
//
// Same-currency values use the identity rate without a provider call; a
// different currency is never assigned a fabricated 1:1 rate.
//
// By explicit owner policy, each LIVE portfolio contributes its latest
// available valuation even when valuation dates differ or the snapshot reports
// stale inputs. Coverage exposes the carried-forward count and oldest included
// date so consumers can label the mixed-date total honestly. Missing
// valuations, invalid/missing FX, or valuation-currency changes remain blocking.
// SIMULATION and MODEL portfolios are intentionally excluded from official
// reporting and do not by themselves make an otherwise complete LIVE total
// incomplete.
type ValuationSummaryAdapter struct {
	funds             domain.FundRepository
	portfolios        domain.PortfolioRepository
	valuation         domain.ValuationRepository
	reportingCurrency string
	fx                contract.MarketQuoteProvider
}

// NewValuationSummaryAdapter wires the adapter.
func NewValuationSummaryAdapter(
	funds domain.FundRepository,
	portfolios domain.PortfolioRepository,
	valuation domain.ValuationRepository,
	reportingCurrency string,
	fx contract.MarketQuoteProvider,
) *ValuationSummaryAdapter {
	return &ValuationSummaryAdapter{
		funds:             funds,
		portfolios:        portfolios,
		valuation:         valuation,
		reportingCurrency: strings.TrimSpace(reportingCurrency),
		fx:                fx,
	}
}

// ReportingCurrency exposes immutable presentation metadata without reading
// any valuation data. Integration uses this only to label permission-denied
// or otherwise unavailable responses without guessing a currency.
func (a *ValuationSummaryAdapter) ReportingCurrency() string {
	if a == nil {
		return ""
	}
	return a.reportingCurrency
}

type scopedValuation struct {
	fund      *entity.Fund
	portfolio *entity.Portfolio
	latest    *entity.ValuationSnapshot
}

type preparedValuation struct {
	scopedValuation
	previous  *entity.ValuationSnapshot
	currentFX decimal.Decimal
}

type fundCoverageState struct {
	includedLive int
	blocking     bool
}

type fxLookupKey struct {
	from string
}

type fxLookupResult struct {
	rate   decimal.Decimal
	reason contract.ValuationSummaryExclusionReason
}

const valuationSummaryPageSize = 1_000

// GetValuationSummary implements contract.ValuationSummaryProvider.
func (a *ValuationSummaryAdapter) GetValuationSummary(
	ctx context.Context,
	req contract.ValuationSummaryRequest,
) (*contract.ValuationSummaryResult, error) {
	if a == nil || a.funds == nil || a.portfolios == nil || a.valuation == nil {
		return nil, fmt.Errorf("valuation summary adapter not initialised")
	}

	result := &contract.ValuationSummaryResult{
		Currency: a.reportingCurrency,
		Status:   contract.ValuationSummaryStatusNoData,
	}

	activeStatus := vo.FundStatusActive
	filter := domain.FundListFilter{
		Status: &activeStatus,
	}
	var portfolioManagerID *uuid.UUID
	switch req.Scope {
	case contract.ValuationSummaryScopeCompany:
		// Company scope is authoritative and company-wide by policy. Enforce that
		// invariant here as well as in the integration caller so a future internal
		// caller cannot accidentally label a permission-filtered subset "company".
	case contract.ValuationSummaryScopeMine:
		uid := req.UserID
		filter.AccessibleFundIDs = req.AccessibleFundIDs
		portfolioManagerID = &uid
	default:
		return nil, fmt.Errorf("invalid valuation summary scope %q", req.Scope)
	}

	funds, err := a.listAllFunds(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing scoped funds: %w", err)
	}
	if len(funds) == 0 {
		return result, nil
	}

	activePortfolio := vo.PortfolioStatusActive
	fundStates := make(map[uuid.UUID]*fundCoverageState, len(funds))
	excludedPortfolioIDs := make(map[uuid.UUID]struct{})
	includedPortfolioIDs := make(map[uuid.UUID]struct{})
	excludedCurrencies := make(map[string]struct{})
	excludedDates := make(map[string]time.Time)
	exclusionReasons := make(map[contract.ValuationSummaryExclusionReason]struct{})
	blockingCoverage := false
	scopedFunds := make([]*entity.Fund, 0, len(funds))

	addExclusion := func(
		fund *entity.Fund,
		portfolio *entity.Portfolio,
		currency string,
		businessDate time.Time,
		requiredBusinessDate time.Time,
		reason contract.ValuationSummaryExclusionReason,
		blocking bool,
	) {
		exclusion := contract.ValuationSummaryExclusion{
			Currency:             currency,
			BusinessDate:         businessDate,
			RequiredBusinessDate: requiredBusinessDate,
			Reason:               reason,
		}
		if fund != nil {
			exclusion.FundCode = fund.Code
			state := fundStates[fund.ID]
			if state == nil {
				state = &fundCoverageState{}
				fundStates[fund.ID] = state
			}
			state.blocking = state.blocking || blocking
		}
		if portfolio != nil {
			exclusion.PortfolioCode = portfolio.Code
			excludedPortfolioIDs[portfolio.ID] = struct{}{}
		}
		if currency != "" {
			excludedCurrencies[currency] = struct{}{}
		}
		if !businessDate.IsZero() {
			key := businessDate.UTC().Format("2006-01-02")
			excludedDates[key] = businessDate
		}
		exclusionReasons[reason] = struct{}{}
		result.Coverage.Exclusions = append(result.Coverage.Exclusions, exclusion)
		blockingCoverage = blockingCoverage || blocking
	}

	// One paginated listing covers every scoped fund; grouping by FundID in
	// memory preserves the per-fund iteration order below. The repository sorts
	// by (fund_id, code), so each fund's group keeps the same relative order a
	// per-fund listing would have produced. Portfolios of funds outside the
	// scoped fund set (e.g. inactive funds) are ignored by the grouping lookup.
	portfolioFilter := domain.PortfolioListFilter{
		Status:        &activePortfolio,
		ManagerUserID: portfolioManagerID,
	}
	if req.Scope == contract.ValuationSummaryScopeMine {
		// Company scope must never be narrowed by an accessible-fund filter;
		// mine scope stays bounded by the caller's accessible funds.
		portfolioFilter.AccessibleFundIDs = req.AccessibleFundIDs
	}
	scopedPortfolios, err := a.listAllPortfolios(ctx, portfolioFilter)
	if err != nil {
		return nil, fmt.Errorf("listing scoped portfolios: %w", err)
	}
	portfoliosByFund := make(map[uuid.UUID][]*entity.Portfolio, len(funds))
	for _, portfolio := range scopedPortfolios {
		// Fund-less portfolios (FundID == nil) have no fund AUM to roll into
		// and are intentionally excluded from this fund-centric summary —
		// they group under uuid.Nil, which no real fund.ID ever matches.
		if portfolio.FundID == nil {
			continue
		}
		portfoliosByFund[*portfolio.FundID] = append(portfoliosByFund[*portfolio.FundID], portfolio)
	}

	var candidates []scopedValuation
	for _, fund := range funds {
		portfolios := portfoliosByFund[fund.ID]
		// A mine-scope fund with no portfolio managed by the caller is outside the
		// requested scope. It must not appear as an excluded/no-active fund and
		// must not make the caller's otherwise complete portfolio total unavailable.
		if req.Scope == contract.ValuationSummaryScopeMine && len(portfolios) == 0 {
			continue
		}

		scopedFunds = append(scopedFunds, fund)
		fundStates[fund.ID] = &fundCoverageState{}
		result.Coverage.TotalFundCount++
		result.Coverage.TotalPortfolioCount += len(portfolios)
		if len(portfolios) == 0 {
			addExclusion(fund, nil, fund.BaseCurrency, time.Time{}, time.Time{}, contract.ValuationSummaryExclusionNoActivePortfolio, true)
			continue
		}

		for _, portfolio := range portfolios {
			if portfolio.PortfolioType != vo.PortfolioTypeLive {
				addExclusion(
					fund,
					portfolio,
					portfolio.ValuationCurrency,
					time.Time{},
					time.Time{},
					contract.ValuationSummaryExclusionNonOfficial,
					false,
				)
				continue
			}

			latest, latestErr := a.valuation.GetLatest(ctx, portfolio.ID, vo.ValuationSourceInternal)
			if latestErr != nil {
				return nil, fmt.Errorf("loading latest valuation for portfolio %s: %w", portfolio.ID, latestErr)
			}
			if latest == nil {
				addExclusion(
					fund,
					portfolio,
					portfolio.ValuationCurrency,
					time.Time{},
					time.Time{},
					contract.ValuationSummaryExclusionMissingValuation,
					true,
				)
				continue
			}
			candidates = append(candidates, scopedValuation{fund: fund, portfolio: portfolio, latest: latest})
			if result.BusinessDate.IsZero() || latest.BusinessDate.After(result.BusinessDate) {
				result.BusinessDate = latest.BusinessDate
			}
		}
	}
	funds = scopedFunds

	if len(candidates) == 0 {
		if blockingCoverage {
			result.Status = contract.ValuationSummaryStatusIncomplete
		}
		finalizeCoverage(result, funds, fundStates, includedPortfolioIDs, excludedPortfolioIDs, excludedCurrencies, excludedDates, exclusionReasons)
		return result, nil
	}

	fxCache := make(map[fxLookupKey]fxLookupResult)
	prepared := make([]preparedValuation, 0, len(candidates))
	for _, candidate := range candidates {
		latest := candidate.latest
		previous, previousErr := a.previousInternalSnapshot(ctx, latest.PortfolioID, latest.BusinessDate)
		if previousErr != nil {
			return nil, fmt.Errorf("loading previous valuation for portfolio %s: %w", latest.PortfolioID, previousErr)
		}
		if previous != nil && previous.ValuationCcy != latest.ValuationCcy {
			addExclusion(
				candidate.fund,
				candidate.portfolio,
				previous.ValuationCcy,
				previous.BusinessDate,
				latest.BusinessDate,
				contract.ValuationSummaryExclusionValuationCurrency,
				true,
			)
			continue
		}

		currentFX := a.reportingFXRate(ctx, latest.ValuationCcy, fxCache)
		if currentFX.reason != "" {
			addExclusion(
				candidate.fund,
				candidate.portfolio,
				latest.ValuationCcy,
				latest.BusinessDate,
				latest.BusinessDate,
				currentFX.reason,
				true,
			)
			continue
		}

		prepared = append(prepared, preparedValuation{
			scopedValuation: candidate,
			previous:        previous,
			currentFX:       currentFX.rate,
		})
		if result.Coverage.OldestIncludedBusinessDate.IsZero() || latest.BusinessDate.Before(result.Coverage.OldestIncludedBusinessDate) {
			result.Coverage.OldestIncludedBusinessDate = latest.BusinessDate
		}
		if latest.HasStaleInputs || !sameBusinessDate(latest.BusinessDate, result.BusinessDate) || (previous != nil && previous.HasStaleInputs) {
			result.Coverage.LatestAvailablePortfolioCount++
		}
		includedPortfolioIDs[candidate.portfolio.ID] = struct{}{}
		delete(excludedPortfolioIDs, candidate.portfolio.ID)
		fundStates[candidate.fund.ID].includedLive++
	}

	finalizeCoverage(result, funds, fundStates, includedPortfolioIDs, excludedPortfolioIDs, excludedCurrencies, excludedDates, exclusionReasons)
	if blockingCoverage {
		result.Status = contract.ValuationSummaryStatusIncomplete
		return result, nil
	}
	if len(prepared) == 0 {
		result.Status = contract.ValuationSummaryStatusNoData
		return result, nil
	}

	aum := decimal.Zero
	previousAUM := decimal.Zero
	todayPnL := decimal.Zero
	haveBaseline := true
	for _, item := range prepared {
		latest := item.latest
		aum = aum.Add(latest.AUM.Mul(item.currentFX))
		if latest.CreatedAt.After(result.AsOf) {
			result.AsOf = latest.CreatedAt
		}

		currentLocalPnL := latest.UnrealisedPnL.Add(latest.RealisedPnL)
		if item.previous == nil {
			todayPnL = todayPnL.Add(currentLocalPnL.Mul(item.currentFX))
			haveBaseline = false
			continue
		}

		previousLocalPnL := item.previous.UnrealisedPnL.Add(item.previous.RealisedPnL)
		localDailyPnL := currentLocalPnL.Sub(previousLocalPnL).Mul(item.currentFX)
		todayPnL = todayPnL.Add(localDailyPnL)
		previousAUM = previousAUM.Add(item.previous.AUM.Mul(item.currentFX))
	}

	result.AUM = aum
	result.TodayPnL = todayPnL
	result.DataAvailable = true
	result.Status = contract.ValuationSummaryStatusAvailable

	// Percent base is the prior-snapshot aggregate AUM in reporting currency.
	// Fall back to today's converted AUM when any portfolio has no baseline or
	// the converted prior AUM is zero/negative.
	denominator := previousAUM
	if !haveBaseline || denominator.Sign() <= 0 {
		denominator = aum
	}
	if denominator.Sign() > 0 {
		pct := todayPnL.Div(denominator).Mul(decimal.NewFromInt(100)).Round(4)
		result.TodayPnLPercent = &pct
	}

	return result, nil
}

// listAllPages exhausts a paginated repository listing while enforcing the
// fail-closed aggregation guards: the reported total must stay stable across
// pages, pagination metadata must remain internally consistent, and paging
// must reach exactly the reported total — a silently truncated listing would
// otherwise understate an official company/mine aggregate.
func listAllPages[T any](
	kind string,
	list func(page int) ([]T, int, error),
) ([]T, error) {
	items := make([]T, 0)
	expectedTotal := -1

	for page := 1; ; page++ {
		batch, total, err := list(page)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", page, err)
		}
		if expectedTotal == -1 {
			expectedTotal = total
		} else if total != expectedTotal {
			return nil, fmt.Errorf("%s total changed while aggregating: was %d, now %d", kind, expectedTotal, total)
		}
		if total < 0 || len(items)+len(batch) > expectedTotal {
			return nil, fmt.Errorf("invalid %s pagination metadata on page %d", kind, page)
		}

		items = append(items, batch...)
		if len(items) == expectedTotal {
			return items, nil
		}
		if len(batch) == 0 {
			return nil, fmt.Errorf("%s pagination ended at %d of %d rows", kind, len(items), expectedTotal)
		}
	}
}

func (a *ValuationSummaryAdapter) listAllFunds(
	ctx context.Context,
	filter domain.FundListFilter,
) ([]*entity.Fund, error) {
	filter.Limit = valuationSummaryPageSize
	return listAllPages("fund", func(page int) ([]*entity.Fund, int, error) {
		filter.Page = page
		return a.funds.List(ctx, filter)
	})
}

func (a *ValuationSummaryAdapter) listAllPortfolios(
	ctx context.Context,
	filter domain.PortfolioListFilter,
) ([]*entity.Portfolio, error) {
	filter.Limit = valuationSummaryPageSize
	return listAllPages("portfolio", func(page int) ([]*entity.Portfolio, int, error) {
		filter.Page = page
		return a.portfolios.List(ctx, filter)
	})
}

func (a *ValuationSummaryAdapter) reportingFXRate(
	ctx context.Context,
	fromCurrency string,
	cache map[fxLookupKey]fxLookupResult,
) fxLookupResult {
	fromCurrency = strings.ToUpper(strings.TrimSpace(fromCurrency))
	reportingCurrency := strings.ToUpper(strings.TrimSpace(a.reportingCurrency))
	if fromCurrency == "" || reportingCurrency == "" {
		return fxLookupResult{reason: contract.ValuationSummaryExclusionInvalidFX}
	}
	if fromCurrency == reportingCurrency {
		return fxLookupResult{rate: decimal.NewFromInt(1)}
	}

	key := fxLookupKey{from: fromCurrency}
	if cached, ok := cache[key]; ok {
		return cached
	}

	missing := fxLookupResult{reason: contract.ValuationSummaryExclusionMissingFX}
	if a.fx == nil {
		cache[key] = missing
		return missing
	}

	symbol := "FX_" + fromCurrency + reportingCurrency
	quote, err := a.fx.GetLatestQuote(ctx, symbol)
	if err != nil || quote == nil {
		cache[key] = missing
		return missing
	}

	result := fxLookupResult{rate: quote.Price}
	switch {
	case quote.Symbol != symbol:
		result = fxLookupResult{reason: contract.ValuationSummaryExclusionFXSymbol}
	case strings.ToUpper(strings.TrimSpace(quote.Currency)) != reportingCurrency:
		result = fxLookupResult{reason: contract.ValuationSummaryExclusionFXCurrency}
	case quote.Price.Sign() <= 0:
		result = fxLookupResult{reason: contract.ValuationSummaryExclusionInvalidFX}
	}
	cache[key] = result
	return result
}

func finalizeCoverage(
	result *contract.ValuationSummaryResult,
	funds []*entity.Fund,
	fundStates map[uuid.UUID]*fundCoverageState,
	includedPortfolioIDs map[uuid.UUID]struct{},
	excludedPortfolioIDs map[uuid.UUID]struct{},
	excludedCurrencies map[string]struct{},
	excludedDates map[string]time.Time,
	exclusionReasons map[contract.ValuationSummaryExclusionReason]struct{},
) {
	result.Coverage.IncludedPortfolioCount = len(includedPortfolioIDs)
	result.Coverage.ExcludedPortfolioCount = result.Coverage.TotalPortfolioCount - len(includedPortfolioIDs)
	if result.Coverage.ExcludedPortfolioCount < len(excludedPortfolioIDs) {
		result.Coverage.ExcludedPortfolioCount = len(excludedPortfolioIDs)
	}

	for _, fund := range funds {
		state := fundStates[fund.ID]
		if state != nil && state.includedLive > 0 && !state.blocking {
			result.Coverage.IncludedFundCount++
		}
	}
	result.Coverage.ExcludedFundCount = result.Coverage.TotalFundCount - result.Coverage.IncludedFundCount

	for currency := range excludedCurrencies {
		result.Coverage.ExcludedCurrencies = append(result.Coverage.ExcludedCurrencies, currency)
	}
	sort.Strings(result.Coverage.ExcludedCurrencies)

	dateKeys := make([]string, 0, len(excludedDates))
	for date := range excludedDates {
		dateKeys = append(dateKeys, date)
	}
	sort.Strings(dateKeys)
	for _, date := range dateKeys {
		result.Coverage.ExcludedBusinessDates = append(result.Coverage.ExcludedBusinessDates, excludedDates[date])
	}

	reasons := make([]string, 0, len(exclusionReasons))
	for reason := range exclusionReasons {
		reasons = append(reasons, string(reason))
	}
	sort.Strings(reasons)
	for _, reason := range reasons {
		result.Coverage.ExclusionReasons = append(
			result.Coverage.ExclusionReasons,
			contract.ValuationSummaryExclusionReason(reason),
		)
	}
}

// previousInternalSnapshot returns the most recent INTERNAL valuation
// snapshot strictly before businessDate, or nil when none exists. Repository
// failures propagate so infrastructure errors cannot masquerade as a missing
// baseline and materially overstate daily P&L.
func (a *ValuationSummaryAdapter) previousInternalSnapshot(
	ctx context.Context,
	portfolioID uuid.UUID,
	businessDate time.Time,
) (*entity.ValuationSnapshot, error) {
	from := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	to := businessDate.AddDate(0, 0, -1)
	list, _, err := a.valuation.List(ctx, portfolioID, from, to, 1, 5)
	if err != nil {
		return nil, err
	}
	for _, valuation := range list {
		if valuation.Source == vo.ValuationSourceInternal {
			return valuation, nil
		}
	}
	return nil, nil
}

func sameBusinessDate(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return false
	}
	ay, am, ad := a.UTC().Date()
	by, bm, bd := b.UTC().Date()
	return ay == by && am == bm && ad == bd
}

var _ contract.ValuationSummaryProvider = (*ValuationSummaryAdapter)(nil)
