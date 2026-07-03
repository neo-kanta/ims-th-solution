// Package query holds investment module read use cases.
package query

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// GetLatestFundNAVRequest is the application boundary for fetching a
// fund-level latest valuation snapshot.
//
// The query is read-only: it never writes an AUM/NAV row. Callers that need
// an authoritative dated snapshot must invoke POST .../aum/compute instead.
type GetLatestFundNAVRequest struct {
	FundID uuid.UUID
}

// GetLatestFundNAVResult is the aggregated fund-level snapshot. Decimals
// are returned as values; the transport layer is responsible for emitting
// them as strings to preserve precision on the wire.
type GetLatestFundNAVResult struct {
	Fund           *entity.Fund
	BusinessDate   time.Time
	ValuationCcy   string
	MarketValue    decimal.Decimal
	AUM            decimal.Decimal
	CashBalance    decimal.Decimal
	UnrealisedPnL  decimal.Decimal
	RealisedPnL    decimal.Decimal
	ROI            *decimal.Decimal
	TotalUnits     *decimal.Decimal
	NAVPerUnit     *decimal.Decimal
	HasStaleInputs bool
	IsIndicative   bool
	PortfolioCount int
	// HasAnySnapshot is false when no portfolio under this fund has produced
	// a valuation yet. The handler maps this to a 404.
	HasAnySnapshot bool
}

// GetLatestFundNAVHandler aggregates per-portfolio latest valuation snapshots
// and live cash balances into a fund-level NAV view.
//
// IsIndicative is true whenever the fund holds portfolios in mixed valuation
// currencies (the aggregate then uses the first contributing portfolio's
// currency). NAVPerUnit is populated only when the fund.HasUnits is true and
// the most recent NAV snapshot of the highest-AUM portfolio is available.
type GetLatestFundNAVHandler struct {
	funds      domain.FundRepository
	portfolios domain.PortfolioRepository
	cash       domain.CashLedgerRepository
	valuation  domain.ValuationRepository
}

func NewGetLatestFundNAVHandler(
	funds domain.FundRepository,
	portfolios domain.PortfolioRepository,
	cash domain.CashLedgerRepository,
	valuation domain.ValuationRepository,
) *GetLatestFundNAVHandler {
	return &GetLatestFundNAVHandler{
		funds:      funds,
		portfolios: portfolios,
		cash:       cash,
		valuation:  valuation,
	}
}

func (h *GetLatestFundNAVHandler) Handle(
	ctx context.Context,
	req GetLatestFundNAVRequest,
) (*GetLatestFundNAVResult, error) {
	if h == nil || h.funds == nil || h.portfolios == nil || h.valuation == nil {
		return nil, fmt.Errorf("fund nav query handler not initialised")
	}

	fund, err := h.funds.GetByID(ctx, req.FundID)
	if err != nil {
		return nil, fmt.Errorf("loading fund: %w", err)
	}
	if fund == nil {
		return nil, &domain.ErrFundNotFound{FundID: req.FundID.String()}
	}

	rows, _, err := h.portfolios.List(ctx, domain.PortfolioListFilter{
		FundID: &req.FundID,
		Page:   1,
		Limit:  200,
	})
	if err != nil {
		return nil, fmt.Errorf("loading portfolios: %w", err)
	}

	result := &GetLatestFundNAVResult{
		Fund:         fund,
		BusinessDate: time.Time{},
	}

	var topPortfolioID uuid.UUID
	topAUM := decimal.Zero
	mixedCcy := false

	for _, p := range rows {
		val, err := h.valuation.GetLatest(ctx, p.ID, vo.ValuationSourceInternal)
		if err != nil {
			return nil, fmt.Errorf("loading latest valuation for portfolio %s: %w", p.ID, err)
		}
		if val == nil {
			continue
		}
		result.HasAnySnapshot = true
		result.PortfolioCount++

		if result.ValuationCcy == "" {
			result.ValuationCcy = val.ValuationCcy
		} else if val.ValuationCcy != result.ValuationCcy {
			mixedCcy = true
		}

		result.MarketValue = result.MarketValue.Add(val.MarketValue)
		result.AUM = result.AUM.Add(val.AUM)
		result.UnrealisedPnL = result.UnrealisedPnL.Add(val.UnrealisedPnL)
		result.RealisedPnL = result.RealisedPnL.Add(val.RealisedPnL)
		result.CashBalance = result.CashBalance.Add(val.CashBalance)

		if val.HasStaleInputs {
			result.HasStaleInputs = true
		}
		if val.IsIndicative {
			result.IsIndicative = true
		}
		if val.BusinessDate.After(result.BusinessDate) {
			result.BusinessDate = val.BusinessDate
		}
		if val.AUM.GreaterThan(topAUM) {
			topAUM = val.AUM
			topPortfolioID = p.ID
		}
	}

	if mixedCcy {
		result.IsIndicative = true
	}

	// Live cash buffer: prefer the cash ledger over the valuation snapshot,
	// because the snapshot's cash is point-in-time and the ledger reflects
	// post-valuation movements.
	if h.cash != nil {
		var liveCash decimal.Decimal
		haveLive := false
		for _, p := range rows {
			balances, err := h.cash.ListBalances(ctx, p.ID)
			if err != nil {
				return nil, fmt.Errorf("loading cash balances for portfolio %s: %w", p.ID, err)
			}
			for _, b := range balances {
				liveCash = liveCash.Add(b.Balance)
				haveLive = true
			}
		}
		if haveLive {
			result.CashBalance = liveCash
		}
	}

	// AUM-weighted ROI across contributing portfolios.
	if result.AUM.GreaterThan(decimal.Zero) {
		var weighted decimal.Decimal
		var weightDenom decimal.Decimal
		for _, p := range rows {
			val, err := h.valuation.GetLatest(ctx, p.ID, vo.ValuationSourceInternal)
			if err != nil || val == nil || val.ROI == nil {
				continue
			}
			weighted = weighted.Add(val.ROI.Mul(val.AUM))
			weightDenom = weightDenom.Add(val.AUM)
		}
		if weightDenom.GreaterThan(decimal.Zero) {
			roi := weighted.Div(weightDenom)
			result.ROI = &roi
		}
	}

	// NAV-per-unit is populated only when the fund declares unitised
	// accounting AND the highest-AUM portfolio has produced a NAV snapshot.
	// For multi-portfolio unitised funds (rare in this PoC), the value is
	// derived from the dominant portfolio so it remains comparable to the
	// existing NAVSnapshot rows.
	if fund.HasUnits && topPortfolioID != uuid.Nil {
		if navList, _, navErr := h.valuation.ListNAV(
			ctx, topPortfolioID, time.Time{}, time.Time{}, 1, 1,
		); navErr == nil && len(navList) > 0 {
			top := navList[0]
			tu := top.TotalUnits
			np := top.NAVPerUnit
			result.TotalUnits = &tu
			result.NAVPerUnit = &np
			if result.BusinessDate.IsZero() || top.BusinessDate.After(result.BusinessDate) {
				result.BusinessDate = top.BusinessDate
			}
		}
	}

	return result, nil
}
