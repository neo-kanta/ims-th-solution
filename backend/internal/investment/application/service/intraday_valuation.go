// Package service contains application services used by the investment
// module's HTTP handlers and commands.
//
// IntradayValuationService bridges live market data (via the
// contract.MarketQuoteProvider port) with the accounting-grade holdings
// already persisted in investment__portfolio_positions.
//
// IMPORTANT BUSINESS RULE
// Provider market price is NEVER written to investment__price_snapshots and
// NEVER overwrites the official accounting NAV. It is used only to compute a
// real-time estimate (estimated AUM, estimated NAV, unrealised P&L,
// allocation, dashboard freshness). Official NAV remains the latest internal
// valuation snapshot.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// IntradayValuationService computes a fund-level estimated valuation from
// live provider quotes joined onto the live portfolio projection.
type IntradayValuationService struct {
	pool       *pgxpool.Pool
	quotes     contract.MarketQuoteProvider
	funds      domain.FundRepository
	portfolios domain.PortfolioRepository
	cash       domain.CashLedgerRepository
	valuation  domain.ValuationRepository

	staleAfter time.Duration
	parValue   decimal.Decimal
	now        func() time.Time
}

// IntradayConfig configures freshness thresholds and provider behaviour for
// the intraday valuation service. Zero values use sensible defaults.
type IntradayConfig struct {
	// StaleAfter is the maximum age of a quote before it is flagged stale on
	// the dashboard. Defaults to 15 minutes.
	StaleAfter time.Duration

	// ParValue is the issuance par value used to derive units outstanding for
	// a unitised fund that has no accounting NAV snapshot yet (units =
	// official AUM / par). Thai mutual funds conventionally IPO at 10.00, so
	// that is the default. Set <= 0 to disable the par fallback (units then
	// stay "—" until a NAV snapshot exists).
	ParValue decimal.Decimal

	// Now is the clock injection used by tests. Defaults to time.Now().UTC.
	Now func() time.Time
}

// NewIntradayValuationService wires the service.
func NewIntradayValuationService(
	pool *pgxpool.Pool,
	quotes contract.MarketQuoteProvider,
	funds domain.FundRepository,
	portfolios domain.PortfolioRepository,
	cash domain.CashLedgerRepository,
	valuation domain.ValuationRepository,
	cfg IntradayConfig,
) *IntradayValuationService {
	if cfg.StaleAfter <= 0 {
		cfg.StaleAfter = 15 * time.Minute
	}
	if cfg.Now == nil {
		cfg.Now = func() time.Time { return time.Now().UTC() }
	}
	if cfg.ParValue.IsZero() {
		cfg.ParValue = decimal.NewFromInt(10) // Thai mutual-fund IPO par
	}
	return &IntradayValuationService{
		pool:       pool,
		quotes:     quotes,
		funds:      funds,
		portfolios: portfolios,
		cash:       cash,
		valuation:  valuation,
		staleAfter: cfg.StaleAfter,
		parValue:   cfg.ParValue,
		now:        cfg.Now,
	}
}

// -----------------------------------------------------------------------------
// Public types
// -----------------------------------------------------------------------------

// IntradayPosition is one row in the live positions table.
type IntradayPosition struct {
	InstrumentID     uuid.UUID
	Ticker           string
	Name             string
	AssetClassCode   string
	AssetClassName   string
	Currency         string
	Quantity         decimal.Decimal
	AverageCost      decimal.Decimal
	CostBasis        decimal.Decimal
	LatestPrice      decimal.Decimal
	MarketValue      decimal.Decimal
	UnrealisedPnL    decimal.Decimal
	UnrealisedPnLPct *decimal.Decimal // nil when cost basis is zero
	Provider         string
	PriceAt          time.Time
	IsStale          bool
	StaleReason      string
}

// IntradayCashRow surfaces cash balances split by currency.
type IntradayCashRow struct {
	Currency string
	Balance  decimal.Decimal
}

// IntradayAllocationBucket is one entry in a breakdown.
type IntradayAllocationBucket struct {
	Key         string
	Label       string
	MarketValue decimal.Decimal
	PctOfTotal  decimal.Decimal // 0..100, 2 decimal places
}

// IntradayValuationResult is the full payload returned by ComputeFundValuation.
//
// OfficialAUM / OfficialNAVPerUnit come from the most recent accounting
// valuation snapshot. EstimatedAUM / EstimatedNAVPerUnit are computed live
// from provider quotes.
type IntradayValuationResult struct {
	FundID         uuid.UUID
	FundCode       string
	ValuationCcy   string
	AsOf           time.Time
	BusinessDate   time.Time
	PortfolioCount int

	OfficialAUM        decimal.Decimal
	OfficialNAVPerUnit *decimal.Decimal
	OfficialAsOf       time.Time

	EstimatedAUM        decimal.Decimal
	EstimatedNAVPerUnit *decimal.Decimal
	UnitsOutstanding    *decimal.Decimal
	DeltaPctVsLastClose decimal.Decimal
	UnrealisedPnL       decimal.Decimal
	CashBalance         decimal.Decimal

	HasUnits      bool
	HasOfficial   bool
	HasLivePrices bool
	// UnitsIndicative is true when units outstanding were derived from the
	// par-value fallback (no accounting NAV snapshot existed). The UI should
	// label the estimated NAV as indicative in that case.
	UnitsIndicative bool
	IsStale         bool
	StaleReason     string

	PrimaryProvider string
	ProvidersUsed   []string

	Positions       []IntradayPosition
	CashRows        []IntradayCashRow
	Allocation      []IntradayAllocationBucket
	UnmappedSymbols []string
}

// RefreshResult is the summary returned by RefreshFundQuotes.
type RefreshResult struct {
	RequestedSymbols int
	SuccessSymbols   int
	FailedSymbols    int
	StaleSymbols     int
	UnmappedSymbols  []string
	Errors           []string
	UsedProvider     string
	StartedAt        time.Time
	CompletedAt      time.Time
}

// ProviderStatus is the payload returned by GetFeedStatus.
type ProviderStatus struct {
	FundID          uuid.UUID
	PrimaryProvider string
	Healthy         bool
	StaleAfter      time.Duration
	StalePositions  int
	TotalPositions  int
	LastQuoteAt     time.Time
	UnmappedSymbols []string
	Note            string
}

// -----------------------------------------------------------------------------
// Read path
// -----------------------------------------------------------------------------

// ComputeFundValuation builds the intraday holdings view for a fund.
//
// Failure modes:
//   - Fund missing → *domain.ErrFundNotFound
//   - Provider chain down → result still returned; IsStale=true, individual
//     positions carry IsStale=true, and StaleReason explains why. No HTTP 500.
func (s *IntradayValuationService) ComputeFundValuation(
	ctx context.Context,
	fundID uuid.UUID,
) (*IntradayValuationResult, error) {
	if s == nil {
		return nil, errors.New("intraday valuation service not initialised")
	}
	fund, err := s.funds.GetByID(ctx, fundID)
	if err != nil {
		return nil, fmt.Errorf("loading fund: %w", err)
	}
	if fund == nil {
		return nil, &domain.ErrFundNotFound{FundID: fundID.String()}
	}

	pf, _, err := s.portfolios.List(ctx, domain.PortfolioListFilter{
		FundID: &fundID,
		Page:   1,
		Limit:  200,
	})
	if err != nil {
		return nil, fmt.Errorf("loading portfolios: %w", err)
	}

	now := s.now()
	result := &IntradayValuationResult{
		FundID:         fundID,
		FundCode:       fund.Code,
		ValuationCcy:   fund.BaseCurrency,
		AsOf:           now,
		HasUnits:       fund.HasUnits,
		PortfolioCount: len(pf),
	}

	if s.quotes != nil {
		result.PrimaryProvider = s.quotes.PrimaryProviderName()
	}

	// 1. Accounting-side official AUM + NAV.
	//    We reuse the existing per-portfolio internal valuation snapshot so the
	//    "official" column matches every other page that reads from it.
	var officialBD time.Time
	var topPortfolioID uuid.UUID
	topAUM := decimal.Zero
	for _, p := range pf {
		val, vErr := s.valuation.GetLatest(ctx, p.ID, vo.ValuationSourceInternal)
		if vErr != nil {
			return nil, fmt.Errorf("loading latest valuation for portfolio %s: %w", p.ID, vErr)
		}
		if val == nil {
			continue
		}
		result.HasOfficial = true
		result.OfficialAUM = result.OfficialAUM.Add(val.AUM)
		if val.BusinessDate.After(officialBD) {
			officialBD = val.BusinessDate
		}
		if val.AUM.GreaterThan(topAUM) {
			topAUM = val.AUM
			topPortfolioID = p.ID
		}
	}
	if !officialBD.IsZero() {
		result.OfficialAsOf = officialBD
		result.BusinessDate = officialBD
	}

	// Units outstanding + official NAV-per-unit. Source priority:
	//  1. The latest accounting NAV snapshot of the dominant (highest-AUM)
	//     portfolio — the authoritative units count.
	//  2. Par-value fallback: a unitised fund with an official AUM but no NAV
	//     snapshot yet → units = official AUM / par. Flagged indicative.
	if fund.HasUnits && topPortfolioID != uuid.Nil {
		if nav := s.latestNAV(ctx, topPortfolioID); nav != nil {
			if nav.TotalUnits.Sign() > 0 {
				tu := nav.TotalUnits
				result.UnitsOutstanding = &tu
			}
			if nav.NAVPerUnit.Sign() > 0 {
				np := nav.NAVPerUnit
				result.OfficialNAVPerUnit = &np
			}
			if nav.BusinessDate.After(result.OfficialAsOf) {
				result.OfficialAsOf = nav.BusinessDate
			}
		}
		if result.UnitsOutstanding == nil && result.HasOfficial &&
			result.OfficialAUM.Sign() > 0 && s.parValue.Sign() > 0 {
			units := result.OfficialAUM.Div(s.parValue).Round(4)
			if units.Sign() > 0 {
				result.UnitsOutstanding = &units
				result.UnitsIndicative = true
				if result.OfficialNAVPerUnit == nil {
					par := s.parValue
					result.OfficialNAVPerUnit = &par
				}
			}
		}
	}

	// 2. Live cash buffer (always from the ledger, not the snapshot).
	cashByCcy := map[string]decimal.Decimal{}
	for _, p := range pf {
		balances, cErr := s.cash.ListBalances(ctx, p.ID)
		if cErr != nil {
			return nil, fmt.Errorf("loading cash balances for portfolio %s: %w", p.ID, cErr)
		}
		for _, b := range balances {
			ccy := strings.ToUpper(b.Currency)
			cashByCcy[ccy] = cashByCcy[ccy].Add(b.Balance)
			result.CashBalance = result.CashBalance.Add(b.Balance)
		}
	}
	result.CashRows = make([]IntradayCashRow, 0, len(cashByCcy))
	for ccy, bal := range cashByCcy {
		result.CashRows = append(result.CashRows, IntradayCashRow{Currency: ccy, Balance: bal})
	}

	// 3. Per-instrument holdings projection. The query joins positions to the
	//    instrument master and asset-class taxonomy so we can render the table
	//    without a second round-trip per row.
	rawPositions, err := s.loadPositions(ctx, fundID)
	if err != nil {
		return nil, err
	}

	// 4. Quote each unique instrument through the provider port. Failures
	//    degrade to stale=true at the position level — they do not abort the
	//    whole valuation.
	quoteByInstrument, unmapped, providersUsed, anyLive, anyStale := s.quoteHoldings(ctx, rawPositions)

	assembleValuation(result, fund.HasUnits, rawPositions, quoteByInstrument, unmapped, providersUsed, anyLive, anyStale, s.isStaleByAge)
	return result, nil
}

// assembleValuation is the pure assembly step: takes accounting/cash inputs
// already populated on `result`, plus the per-position quote lookup, and
// fills in positions, allocation, estimated AUM/NAV, delta and stale flag.
//
// Extracted so unit tests can drive the math without a database.
func assembleValuation(
	result *IntradayValuationResult,
	hasUnits bool,
	positions []rawPosition,
	quoteByInstrument map[uuid.UUID]*contract.MarketQuote,
	unmapped []string,
	providersUsed []string,
	anyLive, anyStale bool,
	isStaleByAge func(time.Time) bool,
) {
	result.UnmappedSymbols = unmapped
	result.ProvidersUsed = providersUsed
	result.HasLivePrices = anyLive

	totalMV := decimal.Zero
	unrealised := decimal.Zero
	allocByClass := map[string]*IntradayAllocationBucket{}

	for _, p := range positions {
		row := IntradayPosition{
			InstrumentID:   p.instrumentID,
			Ticker:         p.ticker,
			Name:           p.name,
			AssetClassCode: p.assetClassCode,
			AssetClassName: p.assetClassName,
			Currency:       p.currency,
			Quantity:       p.quantity,
			AverageCost:    p.averageCost,
			CostBasis:      p.costBasis,
		}

		// Pricing priority for the estimate:
		//  1. A live provider quote with a positive price (true intraday value).
		//  2. The instrument's last official accounting price (carry at last
		//     close). This keeps un-quotable instruments — e.g. Thai government
		//     bonds that no equity provider quotes — at their book value instead
		//     of collapsing the estimated AUM to zero.
		//  3. Average cost (carry at book) when no official price exists either.
		// Cases 2 and 3 are flagged stale because they are not a live quote.
		q, ok := quoteByInstrument[p.instrumentID]
		if ok && q != nil && q.Price.Sign() > 0 {
			row.LatestPrice = q.Price
			row.Provider = q.Provider
			row.PriceAt = q.EffectiveAt
			row.IsStale = q.Stale || (isStaleByAge != nil && isStaleByAge(q.EffectiveAt))
			row.StaleReason = q.StaleReason
		} else {
			row.IsStale = true
			switch {
			case p.lastOfficialPrice.Sign() > 0:
				row.LatestPrice = p.lastOfficialPrice
				row.StaleReason = "no live quote — valued at last official price"
			case p.averageCost.Sign() > 0:
				row.LatestPrice = p.averageCost
				row.StaleReason = "no live quote — valued at carrying cost"
			default:
				row.StaleReason = "no provider quote available"
			}
		}
		row.MarketValue = row.Quantity.Mul(row.LatestPrice)
		row.UnrealisedPnL = row.MarketValue.Sub(row.CostBasis)
		if row.CostBasis.Sign() != 0 {
			pct := row.UnrealisedPnL.Div(row.CostBasis).Mul(decimal.NewFromInt(100)).Round(4)
			row.UnrealisedPnLPct = &pct
		}

		totalMV = totalMV.Add(row.MarketValue)
		unrealised = unrealised.Add(row.UnrealisedPnL)

		key := p.assetClassCode
		label := p.assetClassName
		if key == "" {
			key = "UNCLASSIFIED"
			label = "Unclassified"
		}
		if b, exists := allocByClass[key]; exists {
			b.MarketValue = b.MarketValue.Add(row.MarketValue)
		} else {
			allocByClass[key] = &IntradayAllocationBucket{Key: key, Label: label, MarketValue: row.MarketValue}
		}

		result.Positions = append(result.Positions, row)
		if row.IsStale {
			anyStale = true
		}
	}

	estimatedAUM := totalMV.Add(result.CashBalance)
	if result.CashBalance.GreaterThan(decimal.Zero) {
		if b, exists := allocByClass["CASH"]; exists {
			b.MarketValue = b.MarketValue.Add(result.CashBalance)
		} else {
			allocByClass["CASH"] = &IntradayAllocationBucket{
				Key: "CASH", Label: "Cash & Equivalent", MarketValue: result.CashBalance,
			}
		}
	}

	result.EstimatedAUM = estimatedAUM
	result.UnrealisedPnL = unrealised

	// Estimated NAV per unit — only when the fund is unitised AND units
	// outstanding came through from the official NAV snapshot.
	if hasUnits && result.UnitsOutstanding != nil && result.UnitsOutstanding.Sign() > 0 {
		nav := estimatedAUM.Div(*result.UnitsOutstanding).Round(6)
		result.EstimatedNAVPerUnit = &nav
	}

	if result.HasOfficial && result.OfficialAUM.Sign() > 0 {
		delta := estimatedAUM.Sub(result.OfficialAUM).Div(result.OfficialAUM).Mul(decimal.NewFromInt(100)).Round(4)
		result.DeltaPctVsLastClose = delta
	}

	for _, b := range allocByClass {
		if estimatedAUM.Sign() > 0 {
			b.PctOfTotal = b.MarketValue.Mul(decimal.NewFromInt(100)).Div(estimatedAUM).Round(2)
		}
		result.Allocation = append(result.Allocation, *b)
	}

	if anyStale {
		result.IsStale = true
		result.StaleReason = "one or more positions show stale or missing market data"
	}
}

// -----------------------------------------------------------------------------
// Write path (refresh)
// -----------------------------------------------------------------------------

// RefreshFundQuotes forces a provider fetch for every instrument held by the
// fund. Successes are persisted by the market_data service; failures are
// counted and surfaced. The caller is expected to audit the action.
func (s *IntradayValuationService) RefreshFundQuotes(
	ctx context.Context,
	fundID uuid.UUID,
) (*RefreshResult, error) {
	if s == nil {
		return nil, errors.New("intraday valuation service not initialised")
	}
	if s.quotes == nil {
		return &RefreshResult{
			StartedAt:    s.now(),
			CompletedAt:  s.now(),
			UsedProvider: "",
			Errors:       []string{"no market data provider configured"},
		}, nil
	}

	fund, err := s.funds.GetByID(ctx, fundID)
	if err != nil {
		return nil, fmt.Errorf("loading fund: %w", err)
	}
	if fund == nil {
		return nil, &domain.ErrFundNotFound{FundID: fundID.String()}
	}

	rawPositions, err := s.loadPositions(ctx, fundID)
	if err != nil {
		return nil, err
	}

	res := &RefreshResult{
		StartedAt:    s.now(),
		UsedProvider: s.quotes.PrimaryProviderName(),
	}

	type fetchOutcome struct {
		instID     uuid.UUID
		stale      bool
		unmapped   string
		errMessage string
	}

	// Deduplicate by provider symbol so a fund holding the same instrument
	// across portfolios does not double-charge the provider rate limit.
	type job struct {
		instrumentID   uuid.UUID
		providerSymbol string
	}
	seen := map[string]bool{}
	jobs := make([]job, 0, len(rawPositions))
	for _, p := range rawPositions {
		sym := pickProviderSymbol(p, s.quotes.PrimaryProviderName())
		if sym == "" {
			res.UnmappedSymbols = append(res.UnmappedSymbols, p.ticker)
			continue
		}
		if seen[sym] {
			continue
		}
		seen[sym] = true
		jobs = append(jobs, job{p.instrumentID, sym})
	}
	res.RequestedSymbols = len(jobs)

	const concurrency = 4
	sem := make(chan struct{}, concurrency)
	outcomes := make(chan fetchOutcome, len(jobs))
	var wg sync.WaitGroup

	for _, j := range jobs {
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			q, err := s.quotes.GetLatestQuote(ctx, j.providerSymbol)
			o := fetchOutcome{instID: j.instrumentID}
			if err != nil {
				o.errMessage = fmt.Sprintf("%s: %s", j.providerSymbol, err.Error())
				outcomes <- o
				return
			}
			if q == nil {
				o.errMessage = fmt.Sprintf("%s: no data", j.providerSymbol)
				outcomes <- o
				return
			}
			if q.Stale {
				o.stale = true
			}
			outcomes <- o
		}(j)
	}
	wg.Wait()
	close(outcomes)

	for o := range outcomes {
		switch {
		case o.errMessage != "":
			res.FailedSymbols++
			res.Errors = append(res.Errors, o.errMessage)
		case o.stale:
			res.StaleSymbols++
		default:
			res.SuccessSymbols++
		}
	}
	res.CompletedAt = s.now()
	return res, nil
}

// GetFeedStatus produces the provider-health snapshot the dashboard uses to
// render the "Feeds OK" / "Feeds stale" badge.
func (s *IntradayValuationService) GetFeedStatus(
	ctx context.Context,
	fundID uuid.UUID,
) (*ProviderStatus, error) {
	if s == nil {
		return nil, errors.New("intraday valuation service not initialised")
	}
	primary := ""
	if s.quotes != nil {
		primary = s.quotes.PrimaryProviderName()
	}
	status := &ProviderStatus{
		FundID:          fundID,
		PrimaryProvider: primary,
		StaleAfter:      s.staleAfter,
	}
	if s.quotes == nil {
		status.Note = "no market data provider configured"
		return status, nil
	}

	rawPositions, err := s.loadPositions(ctx, fundID)
	if err != nil {
		return nil, err
	}
	status.TotalPositions = len(rawPositions)

	now := s.now()
	var latest time.Time
	for _, p := range rawPositions {
		sym := pickProviderSymbol(p, primary)
		if sym == "" {
			status.UnmappedSymbols = append(status.UnmappedSymbols, p.ticker)
			status.StalePositions++
			continue
		}
		q, qErr := s.quotes.GetLatestQuote(ctx, sym)
		if qErr != nil || q == nil {
			status.StalePositions++
			continue
		}
		if q.Stale || now.Sub(q.EffectiveAt) > s.staleAfter {
			status.StalePositions++
			continue
		}
		if q.EffectiveAt.After(latest) {
			latest = q.EffectiveAt
		}
	}
	status.LastQuoteAt = latest
	status.Healthy = status.StalePositions == 0 && status.TotalPositions > 0
	if status.StalePositions > 0 {
		status.Note = fmt.Sprintf("%d of %d positions stale", status.StalePositions, status.TotalPositions)
	} else if status.TotalPositions == 0 {
		status.Note = "no live positions"
	}
	return status, nil
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// rawPosition is the per-row read model used internally. Lives alongside the
// service so neither the entity nor the instrument repository need to grow a
// new field set.
type rawPosition struct {
	instrumentID      uuid.UUID
	ticker            string
	name              string
	currency          string
	primaryExchange   string
	assetClassCode    string
	assetClassName    string
	providerYahoo     string
	providerAlpha     string
	quantity          decimal.Decimal
	averageCost       decimal.Decimal
	costBasis         decimal.Decimal
	lastOfficialPrice decimal.Decimal // latest accounting price snapshot; 0 when none
}

const positionReadSQL = `
WITH portfolio_set AS (
    SELECT id FROM investment__portfolios
    WHERE fund_id = $1 AND deleted_at IS NULL
),
agg AS (
    SELECT
        pos.instrument_id,
        SUM(pos.quantity)            AS quantity,
        SUM(pos.cost_basis)          AS cost_basis,
        SUM(pos.quantity * pos.average_cost) AS weighted_cost
    FROM investment__portfolio_positions pos
    JOIN portfolio_set ps ON ps.id = pos.portfolio_id
    WHERE pos.quantity > 0
    GROUP BY pos.instrument_id
),
latest_official_price AS (
    SELECT DISTINCT ON (ps.instrument_id) ps.instrument_id, ps.price
    FROM investment__price_snapshots ps
    ORDER BY ps.instrument_id, ps.business_date DESC, ps.captured_at DESC
)
SELECT
    a.instrument_id,
    inst.primary_ticker,
    inst.name,
    inst.currency,
    COALESCE(inst.primary_exchange, '') AS primary_exchange,
    COALESCE(ac.code, '')               AS asset_class_code,
    COALESCE(ac.name, '')               AS asset_class_name,
    COALESCE(inst.provider_symbol_yahoo, '')         AS provider_symbol_yahoo,
    COALESCE(inst.provider_symbol_alpha_vantage, '') AS provider_symbol_alpha_vantage,
    a.quantity,
    a.cost_basis,
    CASE WHEN a.quantity > 0 THEN a.weighted_cost / a.quantity ELSE 0 END AS average_cost,
    COALESCE(lop.price, 0)              AS last_official_price
FROM agg a
JOIN investment__instruments inst ON inst.id = a.instrument_id
LEFT JOIN investment__asset_classes ac ON ac.id = inst.asset_class_id
LEFT JOIN latest_official_price lop ON lop.instrument_id = a.instrument_id
WHERE inst.deleted_at IS NULL
ORDER BY a.cost_basis DESC, inst.primary_ticker
`

func (s *IntradayValuationService) loadPositions(ctx context.Context, fundID uuid.UUID) ([]rawPosition, error) {
	rows, err := s.pool.Query(ctx, positionReadSQL, fundID)
	if err != nil {
		return nil, fmt.Errorf("loading positions: %w", err)
	}
	defer rows.Close()

	out := []rawPosition{}
	for rows.Next() {
		var r rawPosition
		if err := rows.Scan(
			&r.instrumentID,
			&r.ticker,
			&r.name,
			&r.currency,
			&r.primaryExchange,
			&r.assetClassCode,
			&r.assetClassName,
			&r.providerYahoo,
			&r.providerAlpha,
			&r.quantity,
			&r.costBasis,
			&r.averageCost,
			&r.lastOfficialPrice,
		); err != nil {
			return nil, fmt.Errorf("scanning position row: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// quoteHoldings fetches a quote per unique instrument. Returns:
//   - quoteByInstrument: instrument id → latest quote (nil when unresolved)
//   - unmappedTickers: tickers without a provider symbol
//   - providersUsed: distinct provider tags observed
//   - anyLive: true when at least one quote was returned non-stale
//   - anyStale: true when at least one quote was stale or missing
func (s *IntradayValuationService) quoteHoldings(
	ctx context.Context,
	positions []rawPosition,
) (map[uuid.UUID]*contract.MarketQuote, []string, []string, bool, bool) {
	quoteByInstrument := make(map[uuid.UUID]*contract.MarketQuote, len(positions))
	var unmapped []string
	providerSet := map[string]bool{}
	primary := ""
	if s.quotes != nil {
		primary = s.quotes.PrimaryProviderName()
	}

	// Deduplicate fetches per provider symbol — repeating a symbol does not
	// help us and burns provider quota.
	type pending struct {
		instrumentIDs []uuid.UUID
		symbol        string
	}
	bySymbol := map[string]*pending{}
	for _, p := range positions {
		sym := pickProviderSymbol(p, primary)
		if sym == "" {
			unmapped = append(unmapped, p.ticker)
			quoteByInstrument[p.instrumentID] = nil
			continue
		}
		if cur, ok := bySymbol[sym]; ok {
			cur.instrumentIDs = append(cur.instrumentIDs, p.instrumentID)
			continue
		}
		bySymbol[sym] = &pending{symbol: sym, instrumentIDs: []uuid.UUID{p.instrumentID}}
	}

	const concurrency = 4
	sem := make(chan struct{}, concurrency)
	type result struct {
		instrumentIDs []uuid.UUID
		quote         *contract.MarketQuote
	}
	resCh := make(chan result, len(bySymbol))
	var wg sync.WaitGroup

	for _, j := range bySymbol {
		wg.Add(1)
		go func(j *pending) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if s.quotes == nil {
				resCh <- result{instrumentIDs: j.instrumentIDs, quote: nil}
				return
			}
			q, err := s.quotes.GetLatestQuote(ctx, j.symbol)
			if err != nil {
				q = nil
			}
			resCh <- result{instrumentIDs: j.instrumentIDs, quote: q}
		}(j)
	}
	wg.Wait()
	close(resCh)

	anyLive, anyStale := false, false
	for r := range resCh {
		for _, instID := range r.instrumentIDs {
			quoteByInstrument[instID] = r.quote
		}
		if r.quote == nil {
			anyStale = true
			continue
		}
		providerSet[r.quote.Provider] = true
		if r.quote.Stale || s.isStaleByAge(r.quote.EffectiveAt) {
			anyStale = true
		} else {
			anyLive = true
		}
	}

	providers := make([]string, 0, len(providerSet))
	for p := range providerSet {
		providers = append(providers, p)
	}
	return quoteByInstrument, unmapped, providers, anyLive, anyStale
}

func (s *IntradayValuationService) isStaleByAge(t time.Time) bool {
	if t.IsZero() {
		return true
	}
	return s.now().Sub(t) > s.staleAfter
}

// latestNAV returns the most recent accounting NAV snapshot for a portfolio,
// or nil when none exists. ListNAV filters on business_date BETWEEN, so we
// pass a wide window (epoch .. now+1y) rather than zero times — zero/zero
// would match nothing.
func (s *IntradayValuationService) latestNAV(ctx context.Context, portfolioID uuid.UUID) *entity.NAVSnapshot {
	if s.valuation == nil {
		return nil
	}
	from := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	to := s.now().AddDate(1, 0, 0)
	list, _, err := s.valuation.ListNAV(ctx, portfolioID, from, to, 1, 1)
	if err != nil || len(list) == 0 {
		return nil
	}
	return list[0]
}

// pickProviderSymbol resolves the best provider symbol for a position.
// Order:
//  1. The mapping column matching the configured primary provider.
//  2. The other mapping column (so a Yahoo-mapped instrument still gets a
//     quote when the primary is Alpha Vantage but it falls back to Yahoo).
//  3. Deterministic fallback for Thai SET equities: "<ticker>.BK".
//  4. The bare ticker.
//
// Returns "" when no symbol could be derived (rare; only happens when ticker
// is empty too).
func pickProviderSymbol(p rawPosition, primary string) string {
	switch strings.ToLower(primary) {
	case "alpha_vantage":
		if p.providerAlpha != "" {
			return p.providerAlpha
		}
		if p.providerYahoo != "" {
			return p.providerYahoo
		}
	case "yahoo":
		if p.providerYahoo != "" {
			return p.providerYahoo
		}
		if p.providerAlpha != "" {
			return p.providerAlpha
		}
	default:
		if p.providerYahoo != "" {
			return p.providerYahoo
		}
		if p.providerAlpha != "" {
			return p.providerAlpha
		}
	}
	// SET equities use ".BK" on every provider we ship today.
	if strings.EqualFold(p.primaryExchange, "SET") && p.ticker != "" {
		return strings.ToUpper(p.ticker) + ".BK"
	}
	if p.ticker != "" {
		return strings.ToUpper(p.ticker)
	}
	return ""
}
