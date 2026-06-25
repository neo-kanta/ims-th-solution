package adapter

import (
	"context"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// QuoteAdapter adapts contract.MarketQuoteProvider to the watchlist QuotePort.
type QuoteAdapter struct {
	provider contract.MarketQuoteProvider
}

func NewQuoteAdapter(provider contract.MarketQuoteProvider) *QuoteAdapter {
	return &QuoteAdapter{provider: provider}
}

func (a *QuoteAdapter) GetLatestQuote(ctx context.Context, providerSymbol string) (*domain.QuoteInfo, error) {
	q, err := a.provider.GetLatestQuote(ctx, providerSymbol)
	if err != nil {
		return nil, domain.ErrProviderUnavailable
	}
	if q == nil {
		return nil, nil
	}
	return &domain.QuoteInfo{
		Symbol:        q.Symbol,
		Provider:      q.Provider,
		Currency:      q.Currency,
		MarketStatus:  q.MarketStatus,
		Price:         q.Price,
		PreviousClose: q.PreviousClose,
		ChangePercent: q.ChangePercent,
		EffectiveAt:   q.EffectiveAt,
		FetchedAt:     q.FetchedAt,
		Stale:         q.Stale,
		StaleReason:   q.StaleReason,
	}, nil
}

func (a *QuoteAdapter) PrimaryProviderName() string {
	return a.provider.PrimaryProviderName()
}
