// Package adapter contains infrastructure adapters for compliance ports.
// The Nop* types are PoC stubs that return sensible empty/zero data,
// allowing the compliance engine to run without live market data feeds.
// Replace them with real adapters backed by the market_data / integration modules.
package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// ==============================================================================
// NopPositionSnapshotAdapter — returns an empty position snapshot
// ==============================================================================

type NopPositionSnapshotAdapter struct{}

func (a *NopPositionSnapshotAdapter) GetSnapshot(
	ctx context.Context,
	portfolioID uuid.UUID,
	asOf time.Time,
) (*spi.PositionSnapshot, error) {
	return &spi.PositionSnapshot{
		AsOf:        asOf,
		PortfolioID: portfolioID,
		Holdings:    []spi.Holding{},
	}, nil
}

// ==============================================================================
// NopMarketDataAdapter — returns dummy NAV, prices, and FX rates
// ==============================================================================

type NopMarketDataAdapter struct{}

func (a *NopMarketDataAdapter) GetNAV(
	ctx context.Context,
	portfolioID uuid.UUID,
	asOf time.Time,
) (*spi.NAVSnapshot, error) {
	return &spi.NAVSnapshot{
		PortfolioID:  portfolioID,
		AsOf:         asOf,
		NAV:          decimal.NewFromInt(100_000_000), // 100M THB default
		CashBalance:  decimal.NewFromInt(10_000_000),  // 10M cash
		ReservedCash: decimal.Zero,
	}, nil
}

func (a *NopMarketDataAdapter) GetPrices(
	ctx context.Context,
	tickers []string,
	asOf time.Time,
) (*spi.MarketPriceSnapshot, error) {
	prices := make(map[string]decimal.Decimal, len(tickers))
	for _, t := range tickers {
		prices[t] = decimal.NewFromInt(10) // 10 THB placeholder
	}
	return &spi.MarketPriceSnapshot{Prices: prices, AsOf: asOf}, nil
}

func (a *NopMarketDataAdapter) GetFXRates(
	ctx context.Context,
	baseCurrency string,
	asOf time.Time,
) (*spi.FXRateSnapshot, error) {
	return &spi.FXRateSnapshot{
		BaseCurrency: baseCurrency,
		Rates:        map[string]decimal.Decimal{"USD": decimal.NewFromFloat(35.5)},
		AsOf:         asOf,
	}, nil
}

// ==============================================================================
// NopInstrumentClassificationAdapter — returns unknown classification
// ==============================================================================

type NopInstrumentClassificationAdapter struct{}

func (a *NopInstrumentClassificationAdapter) GetClassifications(
	ctx context.Context,
	tickers []string,
) (*spi.ClassificationSnapshot, error) {
	items := make([]spi.InstrumentClassification, 0, len(tickers))
	for _, t := range tickers {
		items = append(items, spi.InstrumentClassification{
			Ticker:       t,
			Issuer:       t,
			ParentEntity: t,
			Sector:       "UNKNOWN",
			AssetClass:   "EQUITY",
			Country:      "TH",
		})
	}
	return spi.NewClassificationSnapshot(items), nil
}

// ==============================================================================
// NopCreditRatingAdapter — returns no ratings
// ==============================================================================

type NopCreditRatingAdapter struct{}

func (a *NopCreditRatingAdapter) GetRatings(
	ctx context.Context,
	tickers []string,
) (*spi.CreditRatingSnapshot, error) {
	return &spi.CreditRatingSnapshot{
		Ratings: make(map[string][]spi.CreditRating),
	}, nil
}

// ==============================================================================
// NopRestrictionListAdapter — no restrictions configured
// ==============================================================================

type NopRestrictionListAdapter struct{}

func (a *NopRestrictionListAdapter) GetRestrictions(
	ctx context.Context,
	portfolioID uuid.UUID,
	tickers []string,
) (*spi.RestrictionSnapshot, error) {
	return &spi.RestrictionSnapshot{
		Blacklisted:  make(map[string]spi.RestrictionEntry),
		Whitelisted:  make(map[string]bool),
		HasWhitelist: false,
		GrayListed:   make(map[string]spi.RestrictionEntry),
	}, nil
}

// ==============================================================================
// NopTradeHistoryAdapter — no trade history
// ==============================================================================

type NopTradeHistoryAdapter struct{}

func (a *NopTradeHistoryAdapter) GetHistory(
	ctx context.Context,
	portfolioID uuid.UUID,
	lookbackDays int,
	asOf time.Time,
) (*spi.TradeHistorySnapshot, error) {
	return &spi.TradeHistorySnapshot{
		PortfolioID:  portfolioID,
		LookbackDays: lookbackDays,
		Trades:       []spi.HistoricalTrade{},
	}, nil
}

// ==============================================================================
// NopCalendarAdapter — all days are business days
// ==============================================================================

type NopCalendarAdapter struct{}

func (a *NopCalendarAdapter) GetCalendar(
	ctx context.Context,
	jurisdiction string,
	from, to time.Time,
) (*spi.CalendarSnapshot, error) {
	// Mark every day in range as a business day.
	days := make(map[string]bool)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		days[d.Format("2006-01-02")] = true
	}
	return &spi.CalendarSnapshot{
		BusinessDays: days,
		Holidays:     nil,
	}, nil
}

// ==============================================================================
// NopPortfolioMetadataAdapter — returns minimal metadata
// ==============================================================================

type NopPortfolioMetadataAdapter struct{}

func (a *NopPortfolioMetadataAdapter) GetMetadata(
	ctx context.Context,
	portfolioID uuid.UUID,
) (*spi.PortfolioMetadata, error) {
	return &spi.PortfolioMetadata{
		PortfolioID:   portfolioID,
		PortfolioType: "LIVE",
		BaseCurrency:  "THB",
		Jurisdiction:  "TH",
		MandateType:   "EQUITY",
	}, nil
}
