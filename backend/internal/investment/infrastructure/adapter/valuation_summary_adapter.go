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
// Day P&L is derived in reporting currency for each portfolio:
//
//	(LocalCumulativePnL_t - LocalCumulativePnL_t-1) * FX_t
//	  + LocalAUM_t-1 * (FX_t - FX_t-1)
//
// The first term translates local daily performance at closing FX; the second
// captures reporting-currency FX movement on opening net assets. Current values
// use an exact-date persisted FX snapshot for the current valuation business
// date. Opening assets and the return denominator use the same exact-date
// policy at the previous snapshot's own business date. External cash flows are
// therefore treated at closing FX: the ledger has a business date but no
// authoritative intraday valuation/FX observation per flow, so this endpoint
// does not claim transaction-time or GIPS/time-weighted performance.
//
// Same-currency values use the identity rate without a provider call; a
// different currency is never assigned a fabricated 1:1 rate.
//
// Any missing/stale/wrong-date/invalid FX input, stale valuation, missing
// valuation, or inconsistent current valuation date makes the overall result
// INCOMPLETE. Coverage remains available for diagnostics, but DataAvailable is
// false and no normal UI total may be presented. SIMULATION and MODEL
// portfolios are intentionally excluded from official reporting and do not by
// themselves make an otherwise complete LIVE total incomplete.
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
	previous   *entity.ValuationSnapshot
	currentFX  decimal.Decimal
	previousFX decimal.Decimal
}

type fundCoverageState struct {
	includedLive int
	blocking     bool
}

type fxLookupKey struct {
	from string
	date string
}

type fxLookupResult struct {
	rate   decimal.Decimal
	reason contract.ValuationSummaryExclusionReason
}

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
	result.Coverage.TotalFundCount = len(funds)
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

	var candidates []scopedValuation
	for _, fund := range funds {
		fundStates[fund.ID] = &fundCoverageState{}
		portfolios, _, listErr := a.portfolios.List(ctx, domain.PortfolioListFilter{
			FundID: &fund.ID,
			Status: &activePortfolio,
			Limit:  10_000,
		})
		if listErr != nil {
			return nil, fmt.Errorf("listing portfolios for fund %s: %w", fund.ID, listErr)
		}
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
		if latest.HasStaleInputs {
			addExclusion(
				candidate.fund,
				candidate.portfolio,
				latest.ValuationCcy,
				latest.BusinessDate,
				latest.BusinessDate,
				contract.ValuationSummaryExclusionStaleValuation,
				true,
			)
			continue
		}
		if !sameBusinessDate(latest.BusinessDate, result.BusinessDate) {
			addExclusion(
				candidate.fund,
				candidate.portfolio,
				latest.ValuationCcy,
				latest.BusinessDate,
				result.BusinessDate,
				contract.ValuationSummaryExclusionValuationDate,
				true,
			)
			continue
		}

		previous, previousErr := a.previousInternalSnapshot(ctx, latest.PortfolioID, latest.BusinessDate)
		if previousErr != nil {
			return nil, fmt.Errorf("loading previous valuation for portfolio %s: %w", latest.PortfolioID, previousErr)
		}
		if previous != nil && previous.HasStaleInputs {
			addExclusion(
				candidate.fund,
				candidate.portfolio,
				previous.ValuationCcy,
				previous.BusinessDate,
				previous.BusinessDate,
				contract.ValuationSummaryExclusionStaleValuation,
				true,
			)
			continue
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

		currentFX := a.reportingFXRate(ctx, latest.ValuationCcy, latest.BusinessDate, fxCache)
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

		previousFX := fxLookupResult{rate: decimal.NewFromInt(1)}
		if previous != nil {
			previousFX = a.reportingFXRate(ctx, previous.ValuationCcy, previous.BusinessDate, fxCache)
			if previousFX.reason != "" {
				addExclusion(
					candidate.fund,
					candidate.portfolio,
					previous.ValuationCcy,
					previous.BusinessDate,
					previous.BusinessDate,
					previousFX.reason,
					true,
				)
				continue
			}
		}

		prepared = append(prepared, preparedValuation{
			scopedValuation: candidate,
			previous:        previous,
			currentFX:       currentFX.rate,
			previousFX:      previousFX.rate,
		})
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
		openingFXPnL := item.previous.AUM.Mul(item.currentFX.Sub(item.previousFX))
		todayPnL = todayPnL.Add(localDailyPnL).Add(openingFXPnL)
		previousAUM = previousAUM.Add(item.previous.AUM.Mul(item.previousFX))
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

func (a *ValuationSummaryAdapter) reportingFXRate(
	ctx context.Context,
	fromCurrency string,
	businessDate time.Time,
	cache map[fxLookupKey]fxLookupResult,
) fxLookupResult {
	fromCurrency = strings.TrimSpace(fromCurrency)
	if fromCurrency == "" || a.reportingCurrency == "" || businessDate.IsZero() {
		return fxLookupResult{reason: contract.ValuationSummaryExclusionInvalidFX}
	}
	if fromCurrency == a.reportingCurrency {
		return fxLookupResult{rate: decimal.NewFromInt(1)}
	}

	key := fxLookupKey{from: fromCurrency, date: businessDate.UTC().Format("2006-01-02")}
	if cached, ok := cache[key]; ok {
		return cached
	}

	missing := fxLookupResult{reason: contract.ValuationSummaryExclusionMissingFX}
	if a.fx == nil {
		cache[key] = missing
		return missing
	}

	symbol := "FX_" + fromCurrency + a.reportingCurrency
	quote, err := a.fx.GetQuoteAsOf(ctx, symbol, businessDate)
	if err != nil || quote == nil {
		cache[key] = missing
		return missing
	}

	result := fxLookupResult{rate: quote.Price}
	switch {
	case !sameBusinessDate(quote.EffectiveAt, businessDate):
		result = fxLookupResult{reason: contract.ValuationSummaryExclusionWrongDateFX}
	case quote.Stale:
		result = fxLookupResult{reason: contract.ValuationSummaryExclusionStaleFX}
	case quote.Symbol != symbol:
		result = fxLookupResult{reason: contract.ValuationSummaryExclusionFXSymbol}
	case quote.Currency != a.reportingCurrency:
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
