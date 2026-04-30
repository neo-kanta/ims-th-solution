package application

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

func TestGetQuoteFallsBackFromPrimaryToSecondaryProvider(t *testing.T) {
	primary := &fakeProvider{
		name:     domain.ProviderAlphaVantage,
		quoteErr: fmt.Errorf("primary unavailable"),
	}
	fallback := &fakeProvider{
		name: domain.ProviderYahoo,
		quote: &domain.Quote{
			Symbol: "AAPL",
			Price:  decimal.RequireFromString("171.25"),
			AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
	}
	repo := &fakeRepository{}
	service := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{primary, fallback}, repo, repo, nil)
	capturedAt := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return capturedAt }

	quote, err := service.GetQuote(context.Background(), "aapl")
	require.NoError(t, err)
	require.Equal(t, domain.ProviderYahoo, quote.Provider)
	require.False(t, quote.Stale)
	require.Equal(t, 1, primary.quoteCalls)
	require.Equal(t, 1, fallback.quoteCalls)
	require.NotNil(t, repo.savedQuote)
	require.Equal(t, domain.ProviderYahoo, repo.savedQuote.Provider)
	require.Equal(t, capturedAt, quote.CapturedAt)
}

func TestGetQuoteReturnsStaleCachedSnapshotWhenProvidersFail(t *testing.T) {
	primary := &fakeProvider{name: domain.ProviderAlphaVantage, quoteErr: fmt.Errorf("rate limited")}
	fallback := &fakeProvider{name: domain.ProviderYahoo, quoteErr: fmt.Errorf("fallback unavailable")}
	repo := &fakeRepository{
		latestQuote: &domain.Quote{
			Symbol:   "AAPL",
			Provider: domain.ProviderAlphaVantage,
			Price:    decimal.RequireFromString("170.00"),
			AsOf:     time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		},
	}
	service := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{primary, fallback}, repo, repo, nil)

	quote, err := service.GetQuote(context.Background(), "AAPL")
	require.NoError(t, err)
	require.True(t, quote.Stale)
	require.True(t, quote.Cached)
	require.Contains(t, quote.StaleReason, "returning latest cached snapshot")
	require.True(t, quote.Price.Equal(decimal.RequireFromString("170.00")))
}

func TestGetQuoteUsesProviderSpecificSymbolMapping(t *testing.T) {
	primary := &fakeProvider{
		name: domain.ProviderAlphaVantage,
		quote: &domain.Quote{
			Symbol: "MUTFUND.BK",
			Price:  decimal.RequireFromString("10.50"),
			AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
	}
	repo := &fakeRepository{
		mapping: &domain.SymbolMapping{
			Symbol:             "FUND1",
			AssetType:          domain.AssetTypeMutualFund,
			Currency:           "THB",
			AlphaVantageSymbol: "MUTFUND.BK",
		},
	}
	service := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{primary}, repo, repo, nil)

	quote, err := service.GetQuote(context.Background(), "fund1")
	require.NoError(t, err)
	require.Equal(t, "MUTFUND.BK", primary.lastSymbol)
	require.Equal(t, "FUND1", quote.Symbol)
	require.Equal(t, domain.AssetTypeMutualFund, quote.AssetType)
	require.Equal(t, "THB", quote.Currency)
}

func TestGetQuoteFromProviderUsesOnlySelectedProvider(t *testing.T) {
	alpha := &fakeProvider{
		name: domain.ProviderAlphaVantage,
		quote: &domain.Quote{
			Symbol: "AAPL",
			Price:  decimal.RequireFromString("170.00"),
			AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
	}
	yahoo := &fakeProvider{
		name: domain.ProviderYahoo,
		quote: &domain.Quote{
			Symbol: "AAPL",
			Price:  decimal.RequireFromString("171.25"),
			AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
	}
	service := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{alpha, yahoo}, &fakeRepository{}, nil, nil)

	quote, err := service.GetQuoteFromProvider(context.Background(), "AAPL", domain.ProviderYahoo)
	require.NoError(t, err)
	require.Equal(t, domain.ProviderYahoo, quote.Provider)
	require.True(t, quote.Price.Equal(decimal.RequireFromString("171.25")))
	require.Equal(t, 0, alpha.quoteCalls)
	require.Equal(t, 1, yahoo.quoteCalls)
}

func TestGetQuoteRejectsInvalidSymbol(t *testing.T) {
	service := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, nil, nil, nil, nil)

	_, err := service.GetQuote(context.Background(), strings.Repeat("A", 65))
	require.Error(t, err)
	require.Contains(t, err.Error(), "at most 64")

	_, err = service.GetQuote(context.Background(), "AA PL")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid characters")
}

func TestImportMarketDataImportsQuoteAndHistoryFromSelectedProvider(t *testing.T) {
	yahoo := &fakeProvider{
		name: domain.ProviderYahoo,
		quote: &domain.Quote{
			Symbol: "AAPL",
			Price:  decimal.RequireFromString("171.25"),
			AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
		history: []domain.PriceBar{
			{
				Symbol: "AAPL",
				Close:  decimal.RequireFromString("171.25"),
				Date:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
			},
			{
				Symbol: "AAPL",
				Close:  decimal.RequireFromString("170.00"),
				Date:   time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	repo := &fakeRepository{}
	service := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{yahoo}, repo, repo, nil)
	importedAt := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return importedAt }

	result, err := service.ImportMarketData(context.Background(), ImportMarketDataRequest{
		Symbol:         "AAPL",
		Provider:       domain.ProviderYahoo,
		IncludeQuote:   true,
		IncludeHistory: true,
		HistoryLimit:   1,
	})
	require.NoError(t, err)
	require.Equal(t, "AAPL", result.Symbol)
	require.Equal(t, domain.ProviderYahoo, result.RequestedProvider)
	require.Equal(t, domain.ProviderYahoo, result.QuoteProvider)
	require.Equal(t, domain.ProviderYahoo, result.HistoryProvider)
	require.Equal(t, importedAt, result.ImportedAt)
	require.NotNil(t, result.Quote)
	require.Len(t, result.DailyPrices, 1)
	require.Equal(t, 1, result.DailyPricesImported)
	require.NotNil(t, repo.savedQuote)
	require.Len(t, repo.dailyBars, 1)
}

func TestGetDailyPricesReturnsStaleCachedBarsWhenProvidersFail(t *testing.T) {
	primary := &fakeProvider{name: domain.ProviderAlphaVantage, historyErr: fmt.Errorf("rate limited")}
	fallback := &fakeProvider{name: domain.ProviderYahoo, historyErr: fmt.Errorf("fallback unavailable")}
	repo := &fakeRepository{
		dailyBars: []domain.PriceBar{{
			Symbol: "AAPL",
			Close:  decimal.RequireFromString("170.00"),
			Date:   time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
		}},
	}
	service := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{primary, fallback}, repo, repo, nil)

	bars, err := service.GetDailyPrices(context.Background(), "AAPL", 10)
	require.NoError(t, err)
	require.Len(t, bars, 1)
	require.True(t, bars[0].Stale)
	require.Contains(t, bars[0].StaleReason, "returning latest cached daily prices")
}

func TestGetDailyPricesDedupeCachedBarsByDateUsingProviderPriority(t *testing.T) {
	primary := &fakeProvider{name: domain.ProviderAlphaVantage, historyErr: fmt.Errorf("rate limited")}
	fallback := &fakeProvider{name: domain.ProviderYahoo, historyErr: fmt.Errorf("fallback unavailable")}
	repo := &fakeRepository{
		dailyBars: []domain.PriceBar{
			{
				Symbol:   "AAPL",
				Provider: domain.ProviderYahoo,
				Close:    decimal.RequireFromString("99.00"),
				Date:     time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
			},
			{
				Symbol:   "AAPL",
				Provider: domain.ProviderAlphaVantage,
				Close:    decimal.RequireFromString("100.00"),
				Date:     time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
			},
			{
				Symbol:   "AAPL",
				Provider: domain.ProviderYahoo,
				Close:    decimal.RequireFromString("98.00"),
				Date:     time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	service := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{primary, fallback}, repo, repo, nil)

	bars, err := service.GetDailyPrices(context.Background(), "AAPL", 10)
	require.NoError(t, err)
	require.Len(t, bars, 2)
	require.Equal(t, "2026-04-28", bars[0].Date.Format("2006-01-02"))
	require.Equal(t, domain.ProviderAlphaVantage, bars[0].Provider)
	require.True(t, bars[0].Close.Equal(decimal.RequireFromString("100.00")))
	require.Equal(t, "2026-04-27", bars[1].Date.Format("2006-01-02"))
}

type fakeProvider struct {
	name       string
	quote      *domain.Quote
	quoteErr   error
	history    []domain.PriceBar
	historyErr error
	quoteCalls int
	lastSymbol string
}

func (p *fakeProvider) GetQuote(ctx context.Context, symbol string) (*domain.Quote, error) {
	p.quoteCalls++
	p.lastSymbol = symbol
	if p.quoteErr != nil {
		return nil, p.quoteErr
	}
	if p.quote == nil {
		return nil, fmt.Errorf("no quote")
	}
	q := *p.quote
	return &q, nil
}

func (p *fakeProvider) GetDailyPrices(ctx context.Context, symbol string) ([]domain.PriceBar, error) {
	p.lastSymbol = symbol
	if p.historyErr != nil {
		return nil, p.historyErr
	}
	return append([]domain.PriceBar(nil), p.history...), nil
}

func (p *fakeProvider) ProviderName() string {
	return p.name
}

type fakeRepository struct {
	latestQuote *domain.Quote
	savedQuote  *domain.Quote
	dailyBars   []domain.PriceBar
	logs        []domain.ProviderRequestLog
	mapping     *domain.SymbolMapping
}

func (r *fakeRepository) UpsertSymbol(ctx context.Context, mapping domain.SymbolMapping) error {
	return nil
}

func (r *fakeRepository) GetSymbolMapping(ctx context.Context, symbol string) (*domain.SymbolMapping, error) {
	if r.mapping == nil {
		return nil, nil
	}
	m := *r.mapping
	return &m, nil
}

func (r *fakeRepository) SaveQuote(ctx context.Context, quote domain.Quote) error {
	q := quote
	r.savedQuote = &q
	return nil
}

func (r *fakeRepository) SaveDailyPrices(ctx context.Context, symbol string, provider string, bars []domain.PriceBar) error {
	r.dailyBars = append([]domain.PriceBar(nil), bars...)
	return nil
}

func (r *fakeRepository) GetLatestQuote(ctx context.Context, symbol string) (*domain.Quote, error) {
	if r.latestQuote == nil {
		return nil, nil
	}
	q := *r.latestQuote
	return &q, nil
}

func (r *fakeRepository) ListDailyPrices(ctx context.Context, symbol string, limit int) ([]domain.PriceBar, error) {
	out := append([]domain.PriceBar(nil), r.dailyBars...)
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *fakeRepository) LogProviderRequest(ctx context.Context, log domain.ProviderRequestLog) error {
	r.logs = append(r.logs, log)
	return nil
}

func (r *fakeRepository) ListLatestProviderRequests(ctx context.Context, limit int) ([]domain.ProviderRequestLog, error) {
	out := append([]domain.ProviderRequestLog(nil), r.logs...)
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
