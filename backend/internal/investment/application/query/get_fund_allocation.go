package query

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// GetFundAllocationRequest names the fund to break down.
//
// BusinessDate scopes the price-selection tier to the same as-of date used by
// the holdings mark-to-market valuation. A zero value defaults to today.
type GetFundAllocationRequest struct {
	FundID       uuid.UUID
	BusinessDate time.Time
}

// AllocationBucket is one entry in a breakdown — e.g. "EQUITY" / "ENERGY" / "TH".
type AllocationBucket struct {
	Key         string          // stable machine key (asset class code, ISO country code, sector code, currency)
	Label       string          // human label
	MarketValue decimal.Decimal // valuation-currency market value
	PctOfNAV    decimal.Decimal // 0..100 (one decimal precision is plenty for UI bars)
}

// GetFundAllocationResult is the aggregated breakdown across all portfolios
// under a fund. AsOf is the latest contributing valuation business_date.
type GetFundAllocationResult struct {
	FundID         uuid.UUID
	AsOf           time.Time
	ValuationCcy   string
	TotalNAV       decimal.Decimal
	TotalCash      decimal.Decimal
	ByAssetClass   []AllocationBucket
	BySector       []AllocationBucket
	ByCountry      []AllocationBucket
	ByCurrency     []AllocationBucket
	PortfolioCount int
	// HasAnySnapshot is false when the fund has no portfolios with valuations
	// (and therefore no NAV to slice). The handler maps this to a 404.
	HasAnySnapshot bool
}

// GetFundAllocationHandler computes asset-class / sector / country / currency
// allocations for a single fund from the live ledger projection.
//
// The handler joins positions × instruments × latest prices to produce the
// market value of each holding, then groups those market values across each
// taxonomy axis. Cash balances flow into the "Cash" asset-class bucket and
// the corresponding currency bucket; they do not contribute to sector or
// country breakdowns since cash has no issuer.
//
// All arithmetic is performed in the fund's reported valuation currency.
// Cross-currency positions are not FX-converted in this read path — the
// caller is expected to know the fund's reported currency from its latest
// valuation snapshot (returned in ValuationCcy).
type GetFundAllocationHandler struct {
	funds      domain.FundRepository
	portfolios domain.PortfolioRepository
	valuation  domain.ValuationRepository
	cash       domain.CashLedgerRepository
	pool       *pgxpool.Pool

	// intraday is wired after construction (see SetIntradayService) because
	// it depends on the market_data quote provider, which is constructed
	// after this handler in module.go. When set, its per-instrument market
	// values REPLACE the SQL-only price CTE below so allocation always
	// agrees with the holdings mark-to-market view — same live quote /
	// market-data snapshot / official price / cost-carry tier per
	// instrument, not a second independent price lookup. When nil (e.g.
	// market_data not wired, or unit tests), the SQL CTE — now bounded by
	// business_date — is used as a still-correct but live-quote-blind
	// fallback.
	intraday *service.IntradayValuationService
}

// SetIntradayService wires the holdings valuation service so allocation
// reuses its exact resolved market values instead of re-deriving prices
// independently. Nil-safe; idempotent.
func (h *GetFundAllocationHandler) SetIntradayService(intraday *service.IntradayValuationService) {
	if h == nil {
		return
	}
	h.intraday = intraday
}

func NewGetFundAllocationHandler(
	funds domain.FundRepository,
	portfolios domain.PortfolioRepository,
	valuation domain.ValuationRepository,
	cash domain.CashLedgerRepository,
	pool *pgxpool.Pool,
) *GetFundAllocationHandler {
	return &GetFundAllocationHandler{
		funds:      funds,
		portfolios: portfolios,
		valuation:  valuation,
		cash:       cash,
		pool:       pool,
	}
}

// rawAllocationRow mirrors the SQL aggregation result. Labels are looked up at
// the taxonomy level (Asset class / sector / country names) so we don't leak
// UUIDs to the API surface.
type rawAllocationRow struct {
	instrumentID   uuid.UUID
	assetClassCode string
	assetClassName string
	sectorCode     *string
	sectorName     *string
	countryCode    string
	countryName    string
	instrumentCcy  string
	marketValue    decimal.Decimal
}

const fundAllocationSQL = `
WITH portfolio_set AS (
    SELECT id FROM investment__portfolios
    WHERE fund_id = $1 AND deleted_at IS NULL
),
-- Bounded by business_date so a historical allocation query agrees with the
-- same-date holdings mark-to-market valuation (never the unconditional
-- latest-ever snapshot).
latest_price AS (
    SELECT DISTINCT ON (ps.instrument_id) ps.instrument_id, ps.price
    FROM investment__price_snapshots ps
    WHERE ps.business_date <= $2
    ORDER BY ps.instrument_id, ps.business_date DESC, ps.captured_at DESC
)
SELECT
    inst.id                                            AS instrument_id,
    ac.code                                            AS asset_class_code,
    ac.name                                            AS asset_class_name,
    sec.code                                           AS sector_code,
    sec.name                                           AS sector_name,
    co.iso_code                                        AS country_code,
    co.name                                            AS country_name,
    inst.currency                                      AS instrument_ccy,
    (pos.quantity * COALESCE(lp.price, 0))             AS market_value
FROM investment__portfolio_positions pos
JOIN portfolio_set ps                  ON ps.id = pos.portfolio_id
JOIN investment__instruments inst      ON inst.id = pos.instrument_id
JOIN investment__asset_classes ac      ON ac.id = inst.asset_class_id
LEFT JOIN investment__sectors sec      ON sec.id = inst.sector_id
JOIN investment__countries co          ON co.id = inst.country_id
LEFT JOIN latest_price lp              ON lp.instrument_id = pos.instrument_id
WHERE pos.quantity > 0
`

func (h *GetFundAllocationHandler) Handle(
	ctx context.Context,
	req GetFundAllocationRequest,
) (*GetFundAllocationResult, error) {
	if h == nil || h.funds == nil || h.portfolios == nil || h.pool == nil {
		return nil, errors.New("fund allocation handler not initialised")
	}

	fund, err := h.funds.GetByID(ctx, req.FundID)
	if err != nil {
		return nil, fmt.Errorf("loading fund: %w", err)
	}
	if fund == nil {
		return nil, &domain.ErrFundNotFound{FundID: req.FundID.String()}
	}

	// Resolve the valuation currency + total NAV by reusing the per-portfolio
	// latest valuation, which already accounts for FX in the snapshot layer.
	pf, _, err := h.portfolios.List(ctx, domain.PortfolioListFilter{
		FundID: &req.FundID,
		Page:   1,
		Limit:  200,
	})
	if err != nil {
		return nil, fmt.Errorf("loading portfolios: %w", err)
	}

	res := &GetFundAllocationResult{
		FundID:       req.FundID,
		ValuationCcy: fund.BaseCurrency,
	}

	var sumCash decimal.Decimal
	cashByCcy := map[string]decimal.Decimal{}

	for _, p := range pf {
		val, vErr := h.valuation.GetLatest(ctx, p.ID, vo.ValuationSourceInternal)
		if vErr != nil {
			return nil, fmt.Errorf("loading latest valuation for portfolio %s: %w", p.ID, vErr)
		}
		if val != nil {
			res.HasAnySnapshot = true
			res.PortfolioCount++
			res.TotalNAV = res.TotalNAV.Add(val.AUM)
			if res.AsOf.IsZero() || val.BusinessDate.After(res.AsOf) {
				res.AsOf = val.BusinessDate
			}
			if res.ValuationCcy == "" {
				res.ValuationCcy = val.ValuationCcy
			}
		}

		if h.cash != nil {
			balances, cErr := h.cash.ListBalances(ctx, p.ID)
			if cErr != nil {
				return nil, fmt.Errorf("loading cash balances for portfolio %s: %w", p.ID, cErr)
			}
			for _, b := range balances {
				sumCash = sumCash.Add(b.Balance)
				ccy := strings.ToUpper(b.Currency)
				cashByCcy[ccy] = cashByCcy[ccy].Add(b.Balance)
			}
		}
	}
	res.TotalCash = sumCash

	if !res.HasAnySnapshot {
		// 404 path — caller renders an empty state instead of zeroes.
		return res, nil
	}

	businessDate := req.BusinessDate
	if businessDate.IsZero() {
		businessDate = time.Now().UTC()
	}

	// Reuse the exact same per-instrument market values the holdings
	// mark-to-market view resolved (live quote / market-data snapshot /
	// official price / cost-carry), so allocation cannot silently disagree
	// with the holdings page for the same fund and business_date. Best
	// effort: an error here still leaves the SQL-only price CTE below as a
	// (business_date-bounded) fallback rather than failing the whole request.
	var mtmMarketValue map[uuid.UUID]decimal.Decimal
	if h.intraday != nil {
		if mtm, mtmErr := h.intraday.ComputeFundValuation(ctx, req.FundID, businessDate); mtmErr == nil && mtm != nil {
			mtmMarketValue = make(map[uuid.UUID]decimal.Decimal, len(mtm.Positions))
			for _, p := range mtm.Positions {
				mtmMarketValue[p.InstrumentID] = mtmMarketValue[p.InstrumentID].Add(p.MarketValue)
			}
		}
	}

	// Aggregate per (asset_class, sector, country, currency) from raw rows.
	rows, err := h.pool.Query(ctx, fundAllocationSQL, req.FundID, businessDate)
	if err != nil {
		return nil, fmt.Errorf("aggregating allocation: %w", err)
	}
	defer rows.Close()

	var allocRows []rawAllocationRow
	for rows.Next() {
		var r rawAllocationRow
		if scanErr := rows.Scan(
			&r.instrumentID,
			&r.assetClassCode,
			&r.assetClassName,
			&r.sectorCode,
			&r.sectorName,
			&r.countryCode,
			&r.countryName,
			&r.instrumentCcy,
			&r.marketValue,
		); scanErr != nil {
			return nil, fmt.Errorf("scanning allocation row: %w", scanErr)
		}
		allocRows = append(allocRows, r)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterating allocation rows: %w", rows.Err())
	}
	allocRows = applyMarketValueOverride(allocRows, mtmMarketValue)
	pctDenominator := allocationPctDenominator(res.TotalNAV, sumCash, allocRows, mtmMarketValue)

	byClass := map[string]*AllocationBucket{}
	bySector := map[string]*AllocationBucket{}
	byCountry := map[string]*AllocationBucket{}
	byCurrency := map[string]*AllocationBucket{}

	for _, r := range allocRows {
		// Asset class
		ackey := r.assetClassCode
		if b, ok := byClass[ackey]; ok {
			b.MarketValue = b.MarketValue.Add(r.marketValue)
		} else {
			byClass[ackey] = &AllocationBucket{
				Key:         ackey,
				Label:       r.assetClassName,
				MarketValue: r.marketValue,
			}
		}

		// Sector (positions without a sector are bucketed as "UNCLASSIFIED")
		secKey := "UNCLASSIFIED"
		secLabel := "Unclassified"
		if r.sectorCode != nil && r.sectorName != nil && *r.sectorCode != "" {
			secKey = *r.sectorCode
			secLabel = *r.sectorName
		}
		if b, ok := bySector[secKey]; ok {
			b.MarketValue = b.MarketValue.Add(r.marketValue)
		} else {
			bySector[secKey] = &AllocationBucket{Key: secKey, Label: secLabel, MarketValue: r.marketValue}
		}

		// Country
		coKey := r.countryCode
		if b, ok := byCountry[coKey]; ok {
			b.MarketValue = b.MarketValue.Add(r.marketValue)
		} else {
			byCountry[coKey] = &AllocationBucket{Key: coKey, Label: r.countryName, MarketValue: r.marketValue}
		}

		// Currency (per-instrument; cash currencies are folded in below)
		ccy := strings.ToUpper(r.instrumentCcy)
		if b, ok := byCurrency[ccy]; ok {
			b.MarketValue = b.MarketValue.Add(r.marketValue)
		} else {
			byCurrency[ccy] = &AllocationBucket{Key: ccy, Label: ccy, MarketValue: r.marketValue}
		}
	}

	// Add cash buckets to the relevant breakdowns. Cash contributes to
	// asset-class "Cash" and to its native currency, but not to sector
	// or country (it has no issuer).
	if sumCash.GreaterThan(decimal.Zero) {
		if b, ok := byClass["CASH"]; ok {
			b.MarketValue = b.MarketValue.Add(sumCash)
		} else {
			byClass["CASH"] = &AllocationBucket{Key: "CASH", Label: "Cash & Equivalent", MarketValue: sumCash}
		}
		for ccy, amount := range cashByCcy {
			if b, ok := byCurrency[ccy]; ok {
				b.MarketValue = b.MarketValue.Add(amount)
			} else {
				byCurrency[ccy] = &AllocationBucket{Key: ccy, Label: ccy, MarketValue: amount}
			}
		}
	}

	// Materialise each map → slice in descending market-value order with
	// pct-of-NAV computed against the fund AUM (preserves a clean ≤ 100% total).
	res.ByAssetClass = materialise(byClass, pctDenominator)
	res.BySector = materialise(bySector, pctDenominator)
	res.ByCountry = materialise(byCountry, pctDenominator)
	res.ByCurrency = materialise(byCurrency, pctDenominator)

	return res, nil
}

// applyMarketValueOverride replaces each row's SQL-derived market value with
// the holdings mark-to-market value for the same instrument, when available.
// This is the mechanism that guarantees allocation cannot silently disagree
// with the holdings valuation for the same fund/business_date: both read the
// exact same resolved price per instrument (live quote / market-data
// snapshot / official price / cost-carry) — allocation just re-buckets it by
// taxonomy dimension instead of by instrument. Rows for instruments the
// mark-to-market view didn't resolve (or when it is unavailable entirely)
// keep their SQL-computed, business_date-bounded value untouched.
func applyMarketValueOverride(rows []rawAllocationRow, mtmMarketValue map[uuid.UUID]decimal.Decimal) []rawAllocationRow {
	if len(mtmMarketValue) == 0 {
		return rows
	}
	for i := range rows {
		if mv, ok := mtmMarketValue[rows[i].instrumentID]; ok {
			rows[i].marketValue = mv
		}
	}
	return rows
}

// allocationPctDenominator picks the total that bucket percentages are
// computed against. Percentages must be computed against the same total the
// bucket values were drawn from: when MTM overrides were applied, the
// buckets no longer sum to the official accounting NAV (officialTotalNAV) —
// they sum to the mark-to-market book value instead — so the denominator has
// to switch to that same MTM total (allocation rows + cash), or allocation
// would show market values from one view (MTM) divided by a total from a
// different view (official), producing percentages that don't foot to the
// displayed figures or agree with the holdings valuation's own totals.
func allocationPctDenominator(
	officialTotalNAV decimal.Decimal,
	sumCash decimal.Decimal,
	allocRows []rawAllocationRow,
	mtmMarketValue map[uuid.UUID]decimal.Decimal,
) decimal.Decimal {
	if len(mtmMarketValue) == 0 {
		return officialTotalNAV
	}
	mtmTotal := sumCash
	for _, r := range allocRows {
		mtmTotal = mtmTotal.Add(r.marketValue)
	}
	if mtmTotal.Sign() > 0 {
		return mtmTotal
	}
	return officialTotalNAV
}

// materialise turns a bucket map into a stable, sorted slice with pct filled in.
func materialise(m map[string]*AllocationBucket, total decimal.Decimal) []AllocationBucket {
	out := make([]AllocationBucket, 0, len(m))
	for _, b := range m {
		if total.GreaterThan(decimal.Zero) {
			b.PctOfNAV = b.MarketValue.Mul(decimal.NewFromInt(100)).Div(total).Round(2)
		}
		out = append(out, *b)
	}
	// Stable order: largest first, then alphabetical key for ties.
	sortBucketsDesc(out)
	return out
}

func sortBucketsDesc(buckets []AllocationBucket) {
	for i := 1; i < len(buckets); i++ {
		j := i
		for j > 0 {
			cur := buckets[j]
			prev := buckets[j-1]
			if cur.MarketValue.GreaterThan(prev.MarketValue) ||
				(cur.MarketValue.Equal(prev.MarketValue) && cur.Key < prev.Key) {
				buckets[j], buckets[j-1] = prev, cur
				j--
				continue
			}
			break
		}
	}
}
