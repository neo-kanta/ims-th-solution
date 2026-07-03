package query

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// NAVHistoryRange names a supported time bucket. Values map directly to the
// frontend range selector.
type NAVHistoryRange string

const (
	NAVHistoryRange1M  NAVHistoryRange = "1M"
	NAVHistoryRange3M  NAVHistoryRange = "3M"
	NAVHistoryRange6M  NAVHistoryRange = "6M"
	NAVHistoryRange1Y  NAVHistoryRange = "1Y"
	NAVHistoryRange5Y  NAVHistoryRange = "5Y"
	NAVHistoryRangeYTD NAVHistoryRange = "YTD"
)

// GetFundNAVHistoryRequest is the application boundary for the time-series
// read. Range defaults to 3M when empty.
type GetFundNAVHistoryRequest struct {
	FundID uuid.UUID
	Range  NAVHistoryRange
}

// NAVHistoryPoint is one daily observation.
type NAVHistoryPoint struct {
	BusinessDate time.Time
	NAVPerUnit   *decimal.Decimal // nil when the fund is not unitised
	AUM          decimal.Decimal
}

// GetFundNAVHistoryResult holds the series plus convenience aggregates the
// UI needs (high/low/latest/delta) so the frontend doesn't redo arithmetic.
type GetFundNAVHistoryResult struct {
	FundID   uuid.UUID
	Range    NAVHistoryRange
	HasUnits bool
	Series   []NAVHistoryPoint
	High     decimal.Decimal // max value across the series (NAV per unit or AUM)
	Low      decimal.Decimal // min value
	Latest   decimal.Decimal // last point
	DeltaPct decimal.Decimal // signed (latest - first) / first * 100
	From     time.Time
	To       time.Time
	IsEmpty  bool // true when the fund has no history at all
}

// GetFundNAVHistoryHandler reads the per-day NAV (unitised funds) or AUM
// (non-unitised) trail for a single fund across a configurable range.
//
// Unitised funds: read investment__nav_snapshots × portfolio with the highest
// recent AUM (matches the GetLatestFundNAV convention).
// Non-unitised funds: read investment__valuation_snapshots for the same
// dominant portfolio and use AUM as the time-series value.
type GetFundNAVHistoryHandler struct {
	funds      domain.FundRepository
	portfolios domain.PortfolioRepository
	valuation  domain.ValuationRepository
}

func NewGetFundNAVHistoryHandler(
	funds domain.FundRepository,
	portfolios domain.PortfolioRepository,
	valuation domain.ValuationRepository,
) *GetFundNAVHistoryHandler {
	return &GetFundNAVHistoryHandler{
		funds:      funds,
		portfolios: portfolios,
		valuation:  valuation,
	}
}

// rangeToWindow converts a logical range into an (inclusive) from/to date pair.
// All ranges anchor on the current Thai business day for stable demo output.
func rangeToWindow(r NAVHistoryRange, now time.Time) (time.Time, time.Time) {
	to := now
	switch r {
	case NAVHistoryRange1M:
		return to.AddDate(0, -1, 0), to
	case NAVHistoryRange3M:
		return to.AddDate(0, -3, 0), to
	case NAVHistoryRange6M:
		return to.AddDate(0, -6, 0), to
	case NAVHistoryRange1Y:
		return to.AddDate(-1, 0, 0), to
	case NAVHistoryRange5Y:
		return to.AddDate(-5, 0, 0), to
	case NAVHistoryRangeYTD:
		return time.Date(to.Year(), 1, 1, 0, 0, 0, 0, to.Location()), to
	default:
		return to.AddDate(0, -3, 0), to
	}
}

func (h *GetFundNAVHistoryHandler) Handle(
	ctx context.Context,
	req GetFundNAVHistoryRequest,
) (*GetFundNAVHistoryResult, error) {
	if h == nil || h.funds == nil || h.portfolios == nil || h.valuation == nil {
		return nil, errors.New("fund nav history handler not initialised")
	}

	fund, err := h.funds.GetByID(ctx, req.FundID)
	if err != nil {
		return nil, fmt.Errorf("loading fund: %w", err)
	}
	if fund == nil {
		return nil, &domain.ErrFundNotFound{FundID: req.FundID.String()}
	}

	rng := req.Range
	if rng == "" {
		rng = NAVHistoryRange3M
	}
	from, to := rangeToWindow(rng, time.Now().UTC())

	// Find the dominant portfolio by latest AUM. The history is reported off
	// that one so unitised funds get a consistent NAV-per-unit trend.
	pf, _, err := h.portfolios.List(ctx, domain.PortfolioListFilter{
		FundID: &req.FundID,
		Page:   1,
		Limit:  200,
	})
	if err != nil {
		return nil, fmt.Errorf("loading portfolios: %w", err)
	}

	var dominantID uuid.UUID
	topAUM := decimal.Zero
	for _, p := range pf {
		val, vErr := h.valuation.GetLatest(ctx, p.ID, vo.ValuationSourceInternal)
		if vErr != nil || val == nil {
			continue
		}
		if val.AUM.GreaterThan(topAUM) {
			topAUM = val.AUM
			dominantID = p.ID
		}
	}

	res := &GetFundNAVHistoryResult{
		FundID:   req.FundID,
		Range:    rng,
		HasUnits: fund.HasUnits,
		From:     from,
		To:       to,
	}

	if dominantID == uuid.Nil {
		res.IsEmpty = true
		return res, nil
	}

	if fund.HasUnits {
		navs, _, navErr := h.valuation.ListNAV(ctx, dominantID, from, to, 1, 1000)
		if navErr != nil {
			return nil, fmt.Errorf("loading nav history: %w", navErr)
		}
		// Repo returns descending by business_date — flip to ascending for plotting.
		for i := len(navs) - 1; i >= 0; i-- {
			n := navs[i]
			np := n.NAVPerUnit
			res.Series = append(res.Series, NAVHistoryPoint{
				BusinessDate: n.BusinessDate,
				NAVPerUnit:   &np,
				AUM:          decimal.Zero, // populated below from valuation row if available
			})
		}
	} else {
		vals, _, vErr := h.valuation.List(ctx, dominantID, from, to, 1, 1000)
		if vErr != nil {
			return nil, fmt.Errorf("loading valuation history: %w", vErr)
		}
		for i := len(vals) - 1; i >= 0; i-- {
			v := vals[i]
			res.Series = append(res.Series, NAVHistoryPoint{
				BusinessDate: v.BusinessDate,
				NAVPerUnit:   nil,
				AUM:          v.AUM,
			})
		}
	}

	if len(res.Series) == 0 {
		res.IsEmpty = true
		return res, nil
	}

	// High / low / latest are computed off the series' "interesting" value:
	// NAV-per-unit when present, otherwise AUM.
	valueAt := func(p NAVHistoryPoint) decimal.Decimal {
		if p.NAVPerUnit != nil {
			return *p.NAVPerUnit
		}
		return p.AUM
	}

	first := valueAt(res.Series[0])
	last := valueAt(res.Series[len(res.Series)-1])
	res.High = first
	res.Low = first
	res.Latest = last

	for _, p := range res.Series {
		v := valueAt(p)
		if v.GreaterThan(res.High) {
			res.High = v
		}
		if v.LessThan(res.Low) {
			res.Low = v
		}
	}

	if first.GreaterThan(decimal.Zero) {
		res.DeltaPct = last.Sub(first).Mul(decimal.NewFromInt(100)).Div(first).Round(4)
	}

	return res, nil
}
