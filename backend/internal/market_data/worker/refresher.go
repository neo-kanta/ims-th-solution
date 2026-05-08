package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/application"
)

type Refresher struct {
	service *application.Service
	symbols []string
	period  time.Duration
}

func NewRefresher(service *application.Service, symbols []string, period time.Duration) *Refresher {
	if period <= 0 {
		period = 15 * time.Minute
	}
	return &Refresher{
		service: service,
		symbols: symbols,
		period:  period,
	}
}

func (r *Refresher) Start(ctx context.Context) {
	if r == nil || r.service == nil || len(r.symbols) == 0 {
		return
	}
	ticker := time.NewTicker(r.period)
	defer ticker.Stop()

	r.refresh(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.refresh(ctx)
		}
	}
}

func (r *Refresher) refresh(ctx context.Context) {
	for _, symbol := range r.symbols {
		if _, err := r.service.GetQuote(ctx, symbol); err != nil {
			slog.Warn("market data refresh failed", "symbol", symbol, "error", err)
		}
	}
}
