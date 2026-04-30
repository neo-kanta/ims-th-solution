package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// ValuationStaleThresholdDays is the default age beyond which a price
// snapshot is considered stale relative to a valuation date. Configurable
// via ValuationRunner constructor argument.
const ValuationStaleThresholdDays = 1

// ValuationRunner orchestrates a portfolio valuation snapshot:
//  1. Read positions + cash for the portfolio
//  2. Fetch the most recent price per held instrument as of business_date
//  3. Compute holding-level and portfolio-level totals (policy.ComputePortfolioValuation)
//  4. Persist a ValuationSnapshot + holding lines + AUM snapshot, optionally
//     a NAV snapshot when the portfolio is unitised
//
// All snapshots are written within a single DB transaction.
type ValuationRunner struct {
	pool        *pgxpool.Pool
	portfolios  domain.PortfolioRepository
	positions   domain.PortfolioPositionRepository
	cash        domain.CashLedgerRepository
	prices      domain.PriceSnapshotRepository
	instruments domain.InstrumentRepository
	valuation   domain.ValuationRepository
	txns        domain.PortfolioTransactionRepository

	staleThresholdDays int
	now                func() time.Time
}

// NewValuationRunner wires the runner.
func NewValuationRunner(
	pool *pgxpool.Pool,
	portfolios domain.PortfolioRepository,
	positions domain.PortfolioPositionRepository,
	cash domain.CashLedgerRepository,
	prices domain.PriceSnapshotRepository,
	instruments domain.InstrumentRepository,
	valuation domain.ValuationRepository,
	txns domain.PortfolioTransactionRepository,
	now func() time.Time,
) *ValuationRunner {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &ValuationRunner{
		pool:               pool,
		portfolios:         portfolios,
		positions:          positions,
		cash:               cash,
		prices:             prices,
		instruments:        instruments,
		valuation:          valuation,
		txns:               txns,
		staleThresholdDays: ValuationStaleThresholdDays,
		now:                now,
	}
}

// RunRequest are the inputs to a valuation run.
//
// FxRates maps "<currency>->ValuationCurrency" to a decimal rate.
// 1.0 is implied when a holding's quote currency equals the valuation ccy.
// FxRates is also used to translate cash balances. A missing rate when
// needed is rejected.
//
// HoldingsValuationCurrency, when set, overrides Portfolio.ValuationCurrency.
type RunRequest struct {
	PortfolioID  uuid.UUID
	BusinessDate time.Time
	FxRates      map[string]decimal.Decimal
	TotalUnits   *decimal.Decimal // required only when Portfolio.HasUnits
	ActorID      uuid.UUID
}

// RunResult is the persisted valuation snapshot. AUM and NAV snapshots are
// embedded for caller convenience.
type RunResult struct {
	Valuation *entity.ValuationSnapshot
	AUM       *entity.AUMSnapshot
	NAV       *entity.NAVSnapshot // nil when portfolio.HasUnits == false
}

// Run performs the valuation pass and persists the result.
func (r *ValuationRunner) Run(ctx context.Context, req RunRequest) (*RunResult, error) {
	if r == nil {
		return nil, fmt.Errorf("valuation runner not initialised")
	}
	if req.PortfolioID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "portfolio_id", Detail: "is required"}
	}
	if req.BusinessDate.IsZero() {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "business_date", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}

	portfolio, err := r.portfolios.GetByID(ctx, req.PortfolioID)
	if err != nil {
		return nil, fmt.Errorf("loading portfolio: %w", err)
	}
	if portfolio == nil {
		return nil, &domain.ErrPortfolioNotFound{PortfolioID: req.PortfolioID.String()}
	}

	positions, err := r.positions.ListByPortfolio(ctx, portfolio.ID)
	if err != nil {
		return nil, fmt.Errorf("loading positions: %w", err)
	}
	balances, err := r.cash.ListBalances(ctx, portfolio.ID)
	if err != nil {
		return nil, fmt.Errorf("loading cash balances: %w", err)
	}

	holdingInputs := make([]policy.HoldingValuationInput, 0, len(positions))
	for _, pos := range positions {
		if pos.Quantity.Sign() == 0 {
			continue
		}
		inst, err := r.instruments.GetByID(ctx, pos.InstrumentID)
		if err != nil {
			return nil, fmt.Errorf("loading instrument %s: %w", pos.InstrumentID, err)
		}
		if inst == nil {
			return nil, &domain.ErrInstrumentNotFound{InstrumentID: pos.InstrumentID.String()}
		}

		latest, err := r.prices.GetLatest(ctx, inst.ID, req.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("loading latest price for %s: %w", inst.ID, err)
		}

		var (
			price       decimal.Decimal
			priceCcy    = inst.Currency
			priceSnapID *uuid.UUID
			isStale     = false
		)
		if latest != nil {
			price = latest.Price
			priceCcy = latest.Currency
			id := latest.ID
			priceSnapID = &id
			isStale = latest.IsStale || isPriceStale(latest.BusinessDate, req.BusinessDate, r.staleThresholdDays)
		} else {
			return nil, &domain.ErrInvalidDecisionRequest{
				Field: "price_snapshot",
				Detail: fmt.Sprintf(
					"missing price snapshot for instrument %s on or before %s",
					inst.ID,
					req.BusinessDate.Format("2006-01-02"),
				),
			}
		}

		fx, err := lookupFx(req.FxRates, priceCcy, portfolio.ValuationCurrency)
		if err != nil {
			return nil, err
		}
		costFx, err := lookupFx(req.FxRates, portfolio.BaseCurrency, portfolio.ValuationCurrency)
		if err != nil {
			return nil, err
		}
		costBasisValuationCcy := pos.CostBasis.Mul(costFx)

		holdingInputs = append(holdingInputs, policy.HoldingValuationInput{
			InstrumentID:         inst.ID,
			PriceSnapshotID:      priceSnapID,
			Quantity:             pos.Quantity,
			CostBasisBase:        costBasisValuationCcy,
			PriceInQuoteCcy:      price,
			QuoteCurrency:        priceCcy,
			FxRateToValuationCcy: fx,
			IsStale:              isStale,
		})
	}

	output := policy.ComputePortfolioValuation(holdingInputs)

	cashInValCcy := decimal.Zero
	for _, bal := range balances {
		fx, err := lookupFx(req.FxRates, bal.Currency, portfolio.ValuationCurrency)
		if err != nil {
			return nil, err
		}
		cashInValCcy = cashInValCcy.Add(bal.Balance.Mul(fx))
	}
	aum := policy.ComputeAUM(output.MarketValue, cashInValCcy)
	realisedPnL := decimal.Zero
	if r.txns != nil {
		realisedPnL, err = r.txns.SumRealisedPnLBase(ctx, portfolio.ID, req.BusinessDate)
		if err != nil {
			return nil, err
		}
		if portfolio.BaseCurrency != portfolio.ValuationCurrency {
			fx, fxErr := lookupFx(req.FxRates, portfolio.BaseCurrency, portfolio.ValuationCurrency)
			if fxErr != nil {
				return nil, fxErr
			}
			realisedPnL = realisedPnL.Mul(fx)
		}
	}

	now := r.now()
	priceSetHash := policy.ComputePriceSetHash(holdingInputs)

	snap := &entity.ValuationSnapshot{
		ID:             uuid.New(),
		PortfolioID:    portfolio.ID,
		BusinessDate:   req.BusinessDate,
		ValuationCcy:   portfolio.ValuationCurrency,
		MarketValue:    output.MarketValue,
		CostBasis:      output.CostBasis,
		UnrealisedPnL:  output.UnrealisedPnL,
		RealisedPnL:    realisedPnL,
		ROI:            output.ROI,
		AUM:            aum,
		CashBalance:    cashInValCcy,
		PriceSetHash:   priceSetHash,
		HasStaleInputs: output.HasStaleInputs,
		IsIndicative:   true,
		Source:         vo.ValuationSourceInternal,
		CreatedAt:      now,
		CreatedBy:      req.ActorID,
	}

	holdingLines := make([]entity.ValuationHoldingLine, 0, len(output.Holdings))
	for _, h := range output.Holdings {
		holdingLines = append(holdingLines, entity.ValuationHoldingLine{
			ID:                   uuid.New(),
			ValuationSnapshotID:  snap.ID,
			InstrumentID:         h.InstrumentID,
			PriceSnapshotID:      h.PriceSnapshotID,
			Quantity:             h.Quantity,
			PriceInQuoteCcy:      h.PriceInQuoteCcy,
			QuoteCurrency:        h.QuoteCurrency,
			FxRateToValuationCcy: h.FxRateToValuationCcy,
			MarketValue:          h.MarketValue,
			CostBasis:            h.CostBasisBase,
			UnrealisedPnL:        h.UnrealisedPnL,
			IsStale:              h.IsStale,
			CreatedAt:            now,
		})
	}
	snap.HoldingLines = holdingLines

	aumSnap := &entity.AUMSnapshot{
		ID:           uuid.New(),
		ScopeType:    vo.AumScopePortfolio,
		ScopeID:      portfolio.ID,
		BusinessDate: req.BusinessDate,
		AUM:          aum,
		ValuationCcy: portfolio.ValuationCurrency,
		Source:       vo.ValuationSourceInternal,
		CreatedAt:    now,
		CreatedBy:    req.ActorID,
	}

	var navSnap *entity.NAVSnapshot
	if portfolio.HasUnits {
		if req.TotalUnits == nil || req.TotalUnits.Sign() <= 0 {
			return nil, &domain.ErrInvalidDecisionRequest{
				Field: "total_units", Detail: "required and must be positive for unitised portfolios",
			}
		}
		navPerUnit := policy.ComputeNAVPerUnit(aum, *req.TotalUnits)
		navSnap = &entity.NAVSnapshot{
			ID:                  uuid.New(),
			PortfolioID:         portfolio.ID,
			BusinessDate:        req.BusinessDate,
			TotalUnits:          *req.TotalUnits,
			NAVPerUnit:          navPerUnit,
			ValuationSnapshotID: snap.ID,
			IsIndicative:        true,
			CreatedAt:           now,
			CreatedBy:           req.ActorID,
		}
	}

	if err := withTx(ctx, r.pool, func(dbtx pgx.Tx) error {
		if err := r.valuation.Insert(ctx, dbtx, snap); err != nil {
			return fmt.Errorf("inserting valuation snapshot: %w", err)
		}
		if err := r.valuation.InsertAUM(ctx, dbtx, aumSnap); err != nil {
			return fmt.Errorf("inserting aum snapshot: %w", err)
		}
		if navSnap != nil {
			if err := r.valuation.InsertNAV(ctx, dbtx, navSnap); err != nil {
				return fmt.Errorf("inserting nav snapshot: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &RunResult{Valuation: snap, AUM: aumSnap, NAV: navSnap}, nil
}

// lookupFx returns 1 when from == to, otherwise looks up the rate or errors.
func lookupFx(rates map[string]decimal.Decimal, from, to string) (decimal.Decimal, error) {
	if from == to {
		return decimal.NewFromInt(1), nil
	}
	if rate, ok := rates[from+"->"+to]; ok && rate.Sign() > 0 {
		return rate, nil
	}
	if rate, ok := rates[from]; ok && rate.Sign() > 0 {
		return rate, nil
	}
	return decimal.Zero, &domain.ErrInvalidDecisionRequest{
		Field:  "fx_rates",
		Detail: fmt.Sprintf("missing FX rate from %s to %s", from, to),
	}
}

func isPriceStale(priceDate, valuationDate time.Time, thresholdDays int) bool {
	if priceDate.IsZero() {
		return true
	}
	cutoff := valuationDate.AddDate(0, 0, -thresholdDays)
	return priceDate.Before(cutoff)
}

// withTx is the runner-local transaction helper. It mirrors the same shape as
// command.withTransaction but is duplicated so the service package does not
// import command (which would create an import cycle).
func withTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
