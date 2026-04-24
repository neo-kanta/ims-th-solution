package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// MarketDataPort provides NAV, prices, and FX rate data.
type MarketDataPort interface {
	GetNAV(ctx context.Context, portfolioID uuid.UUID, asOf time.Time) (*spi.NAVSnapshot, error)
	GetPrices(ctx context.Context, tickers []string, asOf time.Time) (*spi.MarketPriceSnapshot, error)
	GetFXRates(ctx context.Context, baseCurrency string, asOf time.Time) (*spi.FXRateSnapshot, error)
}
