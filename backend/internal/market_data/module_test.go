package marketdata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/infrastructure/alphavantage"
)

func TestMarketDataOperationTimeoutIncludesAlphaVantageLimiterWait(t *testing.T) {
	httpTimeout := 15 * time.Second

	got := marketDataOperationTimeout(httpTimeout, domain.ProviderAlphaVantage, domain.ProviderYahoo)

	require.Equal(t, httpTimeout+alphavantage.DefaultRateLimitInterval(), got)
	require.Greater(t, got, alphavantage.DefaultRateLimitInterval())
}

func TestMarketDataOperationTimeoutDoesNotPadWhenAlphaVantageIsNotConfigured(t *testing.T) {
	httpTimeout := 15 * time.Second

	got := marketDataOperationTimeout(httpTimeout, domain.ProviderYahoo)

	require.Equal(t, httpTimeout, got)
}
