package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

func TestGetQuoteUsesSelectedProviderQueryParam(t *testing.T) {
	alpha := &stubProvider{name: domain.ProviderAlphaVantage}
	yahoo := &stubProvider{
		name: domain.ProviderYahoo,
		quote: &domain.Quote{
			Symbol: "AAPL",
			Price:  decimal.RequireFromString("171.25"),
			AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
	}
	handler := NewHandler(application.NewService(application.Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{alpha, yahoo}, nil, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/market-data/quote?symbol=AAPL&provider=yahoo", nil)
	rec := httptest.NewRecorder()

	handler.GetQuote(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 0, alpha.quoteCalls)
	require.Equal(t, 1, yahoo.quoteCalls)

	var resp struct {
		Data domain.Quote `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, domain.ProviderYahoo, resp.Data.Provider)
	require.Equal(t, "AAPL", resp.Data.Symbol)
}

func TestImportMarketDataUsesSelectedProviderAndHistoryLimit(t *testing.T) {
	yahoo := &stubProvider{
		name: domain.ProviderYahoo,
		quote: &domain.Quote{
			Symbol: "AAPL",
			Price:  decimal.RequireFromString("171.25"),
			AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
		history: []domain.PriceBar{
			{
				Symbol: "AAPL",
				Date:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
				Close:  decimal.RequireFromString("171.25"),
			},
			{
				Symbol: "AAPL",
				Date:   time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
				Close:  decimal.RequireFromString("170.00"),
			},
		},
	}
	repo := &stubRepository{}
	handler := NewHandler(application.NewService(application.Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{yahoo}, repo, repo, nil))

	body := bytes.NewBufferString(`{
		"symbol": "AAPL",
		"provider": "yahoo",
		"include_quote": true,
		"include_history": true,
		"history_limit": 1
	}`)
	req := httptest.NewRequest(http.MethodPost, "/market-data/import", body)
	rec := httptest.NewRecorder()

	handler.ImportMarketData(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.NotNil(t, repo.savedQuote)
	require.Equal(t, domain.ProviderYahoo, repo.savedQuote.Provider)
	require.Len(t, repo.dailyBars, 1)

	var resp struct {
		Data application.ImportMarketDataResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, domain.ProviderYahoo, resp.Data.RequestedProvider)
	require.Equal(t, 1, resp.Data.DailyPricesImported)
}

func TestImportMarketDataRejectsUnknownProvider(t *testing.T) {
	handler := NewHandler(application.NewService(application.Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, nil, nil, nil, nil))

	body := bytes.NewBufferString(`{
		"symbol": "AAPL",
		"provider": "not-a-provider",
		"include_quote": true
	}`)
	req := httptest.NewRequest(http.MethodPost, "/market-data/import", body)
	rec := httptest.NewRecorder()

	handler.ImportMarketData(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

type stubProvider struct {
	name         string
	quote        *domain.Quote
	quoteErr     error
	history      []domain.PriceBar
	historyErr   error
	quoteCalls   int
	historyCalls int
}

func (p *stubProvider) GetQuote(ctx context.Context, symbol string) (*domain.Quote, error) {
	p.quoteCalls++
	if p.quoteErr != nil {
		return nil, p.quoteErr
	}
	if p.quote == nil {
		return nil, fmt.Errorf("no quote")
	}
	q := *p.quote
	return &q, nil
}

func (p *stubProvider) GetDailyPrices(ctx context.Context, symbol string) ([]domain.PriceBar, error) {
	p.historyCalls++
	if p.historyErr != nil {
		return nil, p.historyErr
	}
	return append([]domain.PriceBar(nil), p.history...), nil
}

func (p *stubProvider) ProviderName() string {
	return p.name
}

type stubRepository struct {
	savedQuote *domain.Quote
	dailyBars  []domain.PriceBar
}

func (r *stubRepository) UpsertSymbol(ctx context.Context, mapping domain.SymbolMapping) error {
	return nil
}

func (r *stubRepository) GetSymbolMapping(ctx context.Context, symbol string) (*domain.SymbolMapping, error) {
	return nil, nil
}

func (r *stubRepository) SaveQuote(ctx context.Context, quote domain.Quote) error {
	q := quote
	r.savedQuote = &q
	return nil
}

func (r *stubRepository) SaveDailyPrices(ctx context.Context, symbol string, provider string, bars []domain.PriceBar) error {
	r.dailyBars = append([]domain.PriceBar(nil), bars...)
	return nil
}

func (r *stubRepository) GetLatestQuote(ctx context.Context, symbol string) (*domain.Quote, error) {
	return nil, nil
}

func (r *stubRepository) ListDailyPrices(ctx context.Context, symbol string, limit int) ([]domain.PriceBar, error) {
	return nil, nil
}

func (r *stubRepository) LogProviderRequest(ctx context.Context, log domain.ProviderRequestLog) error {
	return nil
}

func (r *stubRepository) ListLatestProviderRequests(ctx context.Context, limit int) ([]domain.ProviderRequestLog, error) {
	return nil, nil
}
