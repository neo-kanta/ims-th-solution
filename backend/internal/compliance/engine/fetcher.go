package engine

import (
	"context"
	"fmt"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/ports"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// Fetcher batch-fetches data through ports based on declared dependencies.
type Fetcher struct {
	positions       ports.PositionSnapshotPort
	marketData      ports.MarketDataPort
	classifications ports.InstrumentClassificationPort
	ratings         ports.CreditRatingPort
	restrictions    ports.RestrictionListPort
	tradeHistory    ports.TradeHistoryPort
	calendar        ports.CalendarPort
	portfolioMeta   ports.PortfolioMetadataPort
}

// NewFetcher creates a Fetcher with all port dependencies.
func NewFetcher(
	positions ports.PositionSnapshotPort,
	marketData ports.MarketDataPort,
	classifications ports.InstrumentClassificationPort,
	ratings ports.CreditRatingPort,
	restrictions ports.RestrictionListPort,
	tradeHistory ports.TradeHistoryPort,
	calendar ports.CalendarPort,
	portfolioMeta ports.PortfolioMetadataPort,
) *Fetcher {
	return &Fetcher{
		positions:       positions,
		marketData:      marketData,
		classifications: classifications,
		ratings:         ratings,
		restrictions:    restrictions,
		tradeHistory:    tradeHistory,
		calendar:        calendar,
		portfolioMeta:   portfolioMeta,
	}
}

// Fetch fetches only the data declared in deps. Returns error on any failure (fail-closed).
func (f *Fetcher) Fetch(ctx context.Context, deps spi.DataDependencies, input spi.CheckInput) (*spi.DataBundle, error) {
	bundle := &spi.DataBundle{}

	// Collect tickers for classification and restriction lookups
	var tickers []string
	if input.ProposedOrder != nil && input.ProposedOrder.Ticker != "" {
		tickers = append(tickers, input.ProposedOrder.Ticker)
	}

	if deps.Positions && f.positions != nil {
		snap, err := f.positions.GetSnapshot(ctx, input.PortfolioID, input.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("fetching positions: %w", err)
		}
		bundle.Positions = snap
		// Add holding tickers for classification lookup
		if snap != nil {
			for _, h := range snap.Holdings {
				tickers = append(tickers, h.Ticker)
			}
		}
	}

	if deps.NAV && f.marketData != nil {
		nav, err := f.marketData.GetNAV(ctx, input.PortfolioID, input.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("fetching NAV: %w", err)
		}
		bundle.NAV = nav
	}

	if deps.MarketPrices && f.marketData != nil {
		prices, err := f.marketData.GetPrices(ctx, tickers, input.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("fetching market prices: %w", err)
		}
		bundle.MarketPrices = prices
	}

	if deps.FXRates && f.marketData != nil {
		rates, err := f.marketData.GetFXRates(ctx, "THB", input.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("fetching FX rates: %w", err)
		}
		bundle.FXRates = rates
	}

	if deps.Classifications && f.classifications != nil {
		cls, err := f.classifications.GetClassifications(ctx, tickers)
		if err != nil {
			return nil, fmt.Errorf("fetching classifications: %w", err)
		}
		bundle.Classifications = cls
	}

	if deps.CreditRatings && f.ratings != nil {
		ratings, err := f.ratings.GetRatings(ctx, tickers)
		if err != nil {
			return nil, fmt.Errorf("fetching credit ratings: %w", err)
		}
		bundle.CreditRatings = ratings
	}

	if deps.Restrictions && f.restrictions != nil {
		restr, err := f.restrictions.GetRestrictions(ctx, input.PortfolioID, tickers)
		if err != nil {
			return nil, fmt.Errorf("fetching restrictions: %w", err)
		}
		bundle.Restrictions = restr
	}

	if deps.TradeHistory && f.tradeHistory != nil {
		hist, err := f.tradeHistory.GetHistory(ctx, input.PortfolioID, deps.TradeHistoryLookbackDays, input.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("fetching trade history: %w", err)
		}
		bundle.TradeHistory = hist
	}

	if deps.Calendar && f.calendar != nil {
		cal, err := f.calendar.GetCalendar(ctx, "TH", input.BusinessDate, input.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("fetching calendar: %w", err)
		}
		bundle.Calendar = cal
	}

	if deps.PortfolioMeta && f.portfolioMeta != nil {
		meta, err := f.portfolioMeta.GetMetadata(ctx, input.PortfolioID)
		if err != nil {
			return nil, fmt.Errorf("fetching portfolio metadata: %w", err)
		}
		bundle.PortfolioMeta = meta
	}

	return bundle, nil
}
