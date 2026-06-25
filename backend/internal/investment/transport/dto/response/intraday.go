package response

import (
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
)

// IntradayPositionResponse is one row in the live positions table.
//
// All money / quantity values are emitted as JSON strings to preserve decimal
// precision through the wire — clients MUST parse them with a decimal library.
type IntradayPositionResponse struct {
	InstrumentID     uuid.UUID `json:"instrument_id"`
	Ticker           string    `json:"ticker"`
	Name             string    `json:"name"`
	AssetClassCode   string    `json:"asset_class_code,omitempty"`
	AssetClassLabel  string    `json:"asset_class_label,omitempty"`
	Currency         string    `json:"currency"`
	Quantity         string    `json:"quantity"`
	AverageCost      string    `json:"average_cost"`
	CostBasis        string    `json:"cost_basis"`
	LatestPrice      string    `json:"latest_price"`
	MarketValue      string    `json:"market_value"`
	UnrealisedPnL    string    `json:"unrealised_pnl"`
	UnrealisedPnLPct string    `json:"unrealised_pnl_pct,omitempty"`
	Provider         string    `json:"provider,omitempty"`
	PriceAt          string    `json:"price_at,omitempty"`
	IsStale          bool      `json:"is_stale"`
	StaleReason      string    `json:"stale_reason,omitempty"`
}

// IntradayCashRowResponse is one cash row in the live positions table.
type IntradayCashRowResponse struct {
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
}

// IntradayAllocationBucketResponse is one slice of the live allocation chart.
type IntradayAllocationBucketResponse struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	MarketValue string `json:"market_value"`
	PctOfTotal  string `json:"pct_of_total"`
}

// IntradayValuationResponse is the body returned by
// GET /investment/funds/{id}/holdings/valuation.
type IntradayValuationResponse struct {
	FundID         uuid.UUID `json:"fund_id"`
	FundCode       string    `json:"fund_code,omitempty"`
	ValuationCcy   string    `json:"valuation_ccy"`
	AsOf           time.Time `json:"as_of"`
	BusinessDate   string    `json:"business_date,omitempty"`
	PortfolioCount int       `json:"portfolio_count"`

	OfficialAUM        string `json:"official_aum"`
	OfficialNAVPerUnit string `json:"official_nav_per_unit,omitempty"`
	OfficialAsOf       string `json:"official_as_of,omitempty"`

	EstimatedAUM        string `json:"estimated_aum"`
	EstimatedNAVPerUnit string `json:"estimated_nav_per_unit,omitempty"`
	UnitsOutstanding    string `json:"units_outstanding,omitempty"`
	DeltaPctVsLastClose string `json:"delta_pct_vs_last_close"`
	UnrealisedPnL       string `json:"unrealised_pnl"`
	CashBalance         string `json:"cash_balance"`

	HasUnits        bool `json:"has_units"`
	HasOfficial     bool `json:"has_official"`
	HasLivePrices   bool `json:"has_live_prices"`
	UnitsIndicative bool `json:"units_indicative"`
	IsStale         bool `json:"is_stale"`

	StaleReason     string   `json:"stale_reason,omitempty"`
	PrimaryProvider string   `json:"primary_provider,omitempty"`
	ProvidersUsed   []string `json:"providers_used,omitempty"`

	Positions       []IntradayPositionResponse         `json:"positions"`
	Cash            []IntradayCashRowResponse          `json:"cash"`
	Allocation      []IntradayAllocationBucketResponse `json:"allocation"`
	UnmappedSymbols []string                           `json:"unmapped_symbols,omitempty"`
}

// MarketDataRefreshResponse is the body returned by
// POST /investment/funds/{id}/market-data/refresh.
type MarketDataRefreshResponse struct {
	FundID           uuid.UUID `json:"fund_id"`
	RequestedSymbols int       `json:"requested_symbols"`
	SuccessSymbols   int       `json:"success_symbols"`
	StaleSymbols     int       `json:"stale_symbols"`
	FailedSymbols    int       `json:"failed_symbols"`
	UnmappedSymbols  []string  `json:"unmapped_symbols,omitempty"`
	Errors           []string  `json:"errors,omitempty"`
	UsedProvider     string    `json:"used_provider,omitempty"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at"`
}

// MarketDataStatusResponse is the body returned by
// GET /investment/funds/{id}/market-data/status.
type MarketDataStatusResponse struct {
	FundID          uuid.UUID `json:"fund_id"`
	PrimaryProvider string    `json:"primary_provider,omitempty"`
	Healthy         bool      `json:"healthy"`
	StaleAfterSec   int64     `json:"stale_after_seconds"`
	StalePositions  int       `json:"stale_positions"`
	TotalPositions  int       `json:"total_positions"`
	LastQuoteAt     string    `json:"last_quote_at,omitempty"`
	UnmappedSymbols []string  `json:"unmapped_symbols,omitempty"`
	Note            string    `json:"note,omitempty"`
}

// FromIntradayValuation maps the service result onto the wire DTO. Decimals
// are emitted as strings to preserve precision.
func FromIntradayValuation(in *service.IntradayValuationResult) IntradayValuationResponse {
	if in == nil {
		return IntradayValuationResponse{}
	}
	out := IntradayValuationResponse{
		FundID:              in.FundID,
		FundCode:            in.FundCode,
		ValuationCcy:        in.ValuationCcy,
		AsOf:                in.AsOf,
		PortfolioCount:      in.PortfolioCount,
		OfficialAUM:         in.OfficialAUM.String(),
		EstimatedAUM:        in.EstimatedAUM.String(),
		DeltaPctVsLastClose: in.DeltaPctVsLastClose.String(),
		UnrealisedPnL:       in.UnrealisedPnL.String(),
		CashBalance:         in.CashBalance.String(),
		HasUnits:            in.HasUnits,
		HasOfficial:         in.HasOfficial,
		HasLivePrices:       in.HasLivePrices,
		UnitsIndicative:     in.UnitsIndicative,
		IsStale:             in.IsStale,
		StaleReason:         in.StaleReason,
		PrimaryProvider:     in.PrimaryProvider,
		ProvidersUsed:       in.ProvidersUsed,
		UnmappedSymbols:     in.UnmappedSymbols,
	}
	if !in.BusinessDate.IsZero() {
		out.BusinessDate = in.BusinessDate.Format("2006-01-02")
	}
	if !in.OfficialAsOf.IsZero() {
		out.OfficialAsOf = in.OfficialAsOf.Format("2006-01-02")
	}
	if in.OfficialNAVPerUnit != nil {
		out.OfficialNAVPerUnit = in.OfficialNAVPerUnit.String()
	}
	if in.EstimatedNAVPerUnit != nil {
		out.EstimatedNAVPerUnit = in.EstimatedNAVPerUnit.String()
	}
	if in.UnitsOutstanding != nil {
		out.UnitsOutstanding = in.UnitsOutstanding.String()
	}

	out.Positions = make([]IntradayPositionResponse, 0, len(in.Positions))
	for _, p := range in.Positions {
		row := IntradayPositionResponse{
			InstrumentID:    p.InstrumentID,
			Ticker:          p.Ticker,
			Name:            p.Name,
			AssetClassCode:  p.AssetClassCode,
			AssetClassLabel: p.AssetClassName,
			Currency:        p.Currency,
			Quantity:        p.Quantity.String(),
			AverageCost:     p.AverageCost.String(),
			CostBasis:       p.CostBasis.String(),
			LatestPrice:     p.LatestPrice.String(),
			MarketValue:     p.MarketValue.String(),
			UnrealisedPnL:   p.UnrealisedPnL.String(),
			Provider:        p.Provider,
			IsStale:         p.IsStale,
			StaleReason:     p.StaleReason,
		}
		if p.UnrealisedPnLPct != nil {
			row.UnrealisedPnLPct = p.UnrealisedPnLPct.String()
		}
		if !p.PriceAt.IsZero() {
			row.PriceAt = p.PriceAt.Format(time.RFC3339)
		}
		out.Positions = append(out.Positions, row)
	}

	out.Cash = make([]IntradayCashRowResponse, 0, len(in.CashRows))
	for _, c := range in.CashRows {
		out.Cash = append(out.Cash, IntradayCashRowResponse{Currency: c.Currency, Balance: c.Balance.String()})
	}

	out.Allocation = make([]IntradayAllocationBucketResponse, 0, len(in.Allocation))
	for _, a := range in.Allocation {
		out.Allocation = append(out.Allocation, IntradayAllocationBucketResponse{
			Key:         a.Key,
			Label:       a.Label,
			MarketValue: a.MarketValue.String(),
			PctOfTotal:  a.PctOfTotal.String(),
		})
	}
	return out
}

// FromMarketDataRefresh maps the service result onto the wire DTO.
func FromMarketDataRefresh(fundID uuid.UUID, in *service.RefreshResult) MarketDataRefreshResponse {
	if in == nil {
		return MarketDataRefreshResponse{FundID: fundID}
	}
	return MarketDataRefreshResponse{
		FundID:           fundID,
		RequestedSymbols: in.RequestedSymbols,
		SuccessSymbols:   in.SuccessSymbols,
		StaleSymbols:     in.StaleSymbols,
		FailedSymbols:    in.FailedSymbols,
		UnmappedSymbols:  in.UnmappedSymbols,
		Errors:           in.Errors,
		UsedProvider:     in.UsedProvider,
		StartedAt:        in.StartedAt,
		CompletedAt:      in.CompletedAt,
	}
}

// FromMarketDataStatus maps the service result onto the wire DTO.
func FromMarketDataStatus(in *service.ProviderStatus) MarketDataStatusResponse {
	if in == nil {
		return MarketDataStatusResponse{}
	}
	out := MarketDataStatusResponse{
		FundID:          in.FundID,
		PrimaryProvider: in.PrimaryProvider,
		Healthy:         in.Healthy,
		StaleAfterSec:   int64(in.StaleAfter.Seconds()),
		StalePositions:  in.StalePositions,
		TotalPositions:  in.TotalPositions,
		UnmappedSymbols: in.UnmappedSymbols,
		Note:            in.Note,
	}
	if !in.LastQuoteAt.IsZero() {
		out.LastQuoteAt = in.LastQuoteAt.Format(time.RFC3339)
	}
	return out
}
