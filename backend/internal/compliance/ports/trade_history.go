package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// TradeHistoryPort provides historical trades for look-back rules.
type TradeHistoryPort interface {
	GetHistory(ctx context.Context, portfolioID uuid.UUID, lookbackDays int, asOf time.Time) (*spi.TradeHistorySnapshot, error)
}
