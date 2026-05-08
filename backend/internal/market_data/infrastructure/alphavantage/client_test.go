package alphavantage

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

func TestGetQuoteParsesGlobalQuote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GLOBAL_QUOTE", r.URL.Query().Get("function"))
		require.Equal(t, "AAPL", r.URL.Query().Get("symbol"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Global Quote": {
				"01. symbol": "AAPL",
				"02. open": "170.1000",
				"03. high": "172.5000",
				"04. low": "169.0000",
				"05. price": "171.2500",
				"06. volume": "123456",
				"07. latest trading day": "2026-04-28",
				"08. previous close": "170.0000",
				"09. change": "1.2500",
				"10. change percent": "0.7353%"
			}
		}`))
	}))
	defer server.Close()

	provider := NewProvider(
		"test-key",
		time.Second,
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithRateLimitInterval(0),
		WithRetryPolicy(1, 0),
	)

	quote, err := provider.GetQuote(context.Background(), "AAPL")
	require.NoError(t, err)
	require.Equal(t, domain.ProviderAlphaVantage, quote.Provider)
	require.Equal(t, "AAPL", quote.Symbol)
	require.True(t, quote.Price.Equal(decimal.RequireFromString("171.2500")))
	require.True(t, quote.ChangePercent.Equal(decimal.RequireFromString("0.7353")))
	require.Equal(t, int64(123456), quote.Volume)
	require.Equal(t, "2026-04-28", quote.AsOf.Format("2006-01-02"))
	require.False(t, quote.CapturedAt.IsZero())
}

func TestGetDailyPricesParsesTimeSeries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "TIME_SERIES_DAILY", r.URL.Query().Get("function"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Meta Data": {"2. Symbol": "AAPL"},
			"Time Series (Daily)": {
				"2026-04-28": {
					"1. open": "170.0000",
					"2. high": "173.0000",
					"3. low": "169.5000",
					"4. close": "172.7500",
					"5. volume": "2000"
				},
				"2026-04-27": {
					"1. open": "168.0000",
					"2. high": "171.0000",
					"3. low": "167.5000",
					"4. close": "170.0000",
					"5. volume": "1000"
				}
			}
		}`))
	}))
	defer server.Close()

	provider := NewProvider(
		"test-key",
		time.Second,
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithRateLimitInterval(0),
		WithRetryPolicy(1, 0),
	)

	bars, err := provider.GetDailyPrices(context.Background(), "AAPL")
	require.NoError(t, err)
	require.Len(t, bars, 2)
	require.Equal(t, "2026-04-28", bars[0].Date.Format("2006-01-02"))
	require.True(t, bars[0].Close.Equal(decimal.RequireFromString("172.7500")))
	require.Equal(t, int64(2000), bars[0].Volume)
}

func TestAlphaVantageProviderMessagesBecomeProviderErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Note": "Thank you for using Alpha Vantage! Our standard API rate limit is 5 calls per minute."
		}`))
	}))
	defer server.Close()

	provider := NewProvider(
		"test-key",
		time.Second,
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithRateLimitInterval(0),
		WithRetryPolicy(1, 0),
	)

	_, err := provider.GetQuote(context.Background(), "AAPL")
	require.Error(t, err)

	var providerErr *domain.ProviderError
	require.True(t, errors.As(err, &providerErr))
	require.Equal(t, domain.ProviderAlphaVantage, providerErr.Provider)
	require.Equal(t, "quote", providerErr.Operation)
	require.Equal(t, "rate_limited", providerErr.Code)
	require.False(t, providerErr.Retriable)
}
