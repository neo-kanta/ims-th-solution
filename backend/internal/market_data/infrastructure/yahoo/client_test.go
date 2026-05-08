package yahoo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

func TestGetQuoteParsesYahooChartLatestClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v8/finance/chart/AAPL", r.URL.Path)
		require.Equal(t, "5d", r.URL.Query().Get("range"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(chartJSON()))
	}))
	defer server.Close()

	provider := NewProvider(
		time.Second,
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithRateLimitInterval(0),
		WithRetryPolicy(1, 0),
	)

	quote, err := provider.GetQuote(context.Background(), "AAPL")
	require.NoError(t, err)
	require.Equal(t, domain.ProviderYahoo, quote.Provider)
	require.Equal(t, "AAPL", quote.Symbol)
	require.Equal(t, "USD", quote.Currency)
	require.True(t, quote.Price.Equal(decimal.RequireFromString("171.25")))
	require.True(t, quote.PreviousClose.Equal(decimal.RequireFromString("170.00")))
	require.False(t, quote.CapturedAt.IsZero())
}

func TestGetDailyPricesParsesYahooChartBars(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "1y", r.URL.Query().Get("range"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(chartJSON()))
	}))
	defer server.Close()

	provider := NewProvider(
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
	require.True(t, bars[0].Close.Equal(decimal.RequireFromString("171.25")))
	require.True(t, bars[0].AdjustedClose.Equal(decimal.RequireFromString("171.00")))
	require.Equal(t, int64(123456), bars[0].Volume)
}

func chartJSON() string {
	return `{
		"chart": {
			"result": [{
				"meta": {
					"currency": "USD",
					"symbol": "AAPL",
					"regularMarketTime": 1777334400,
					"regularMarketPrice": 171.25,
					"chartPreviousClose": 170.00
				},
				"timestamp": [1777248000, 1777334400],
				"indicators": {
					"quote": [{
						"open": [169.10, 170.10],
						"high": [171.50, 172.50],
						"low": [168.90, 169.00],
						"close": [170.00, 171.25],
						"volume": [100000, 123456]
					}],
					"adjclose": [{
						"adjclose": [169.75, 171.00]
					}]
				}
			}],
			"error": null
		}
	}`
}
