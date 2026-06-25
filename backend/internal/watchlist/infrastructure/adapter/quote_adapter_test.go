package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// fakeProvider implements contract.MarketQuoteProvider.
type fakeProvider struct {
	quote *contract.MarketQuote
	err   error
}

func (f *fakeProvider) GetLatestQuote(_ context.Context, _ string) (*contract.MarketQuote, error) {
	return f.quote, f.err
}
func (f *fakeProvider) PrimaryProviderName() string      { return "fake" }
func (f *fakeProvider) ProviderConfigured(_ string) bool { return true }

// Test 8 (P1-6): provider error maps to domain.ErrProviderUnavailable.
func TestQuoteAdapter_ProviderError_ReturnsErrProviderUnavailable(t *testing.T) {
	p := &fakeProvider{err: errors.New("upstream 500")}
	a := NewQuoteAdapter(p)

	_, err := a.GetLatestQuote(context.Background(), "MOCK.TH")
	if !errors.Is(err, domain.ErrProviderUnavailable) {
		t.Errorf("expected ErrProviderUnavailable, got %v", err)
	}
}

// Bonus: nil quote from provider returns nil without error.
func TestQuoteAdapter_NilQuote_ReturnsNil(t *testing.T) {
	p := &fakeProvider{quote: nil, err: nil}
	a := NewQuoteAdapter(p)

	q, err := a.GetLatestQuote(context.Background(), "MOCK.TH")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if q != nil {
		t.Errorf("expected nil quote, got %+v", q)
	}
}

// Bonus: successful quote maps fields correctly.
func TestQuoteAdapter_SuccessfulQuote_MapsFields(t *testing.T) {
	price := decimal.NewFromFloat(123.45)
	p := &fakeProvider{quote: &contract.MarketQuote{
		Symbol:   "MOCK.TH",
		Provider: "fake",
		Price:    price,
		Stale:    false,
	}}
	a := NewQuoteAdapter(p)

	q, err := a.GetLatestQuote(context.Background(), "MOCK.TH")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q == nil {
		t.Fatal("expected non-nil quote")
	}
	if !q.Price.Equal(price) {
		t.Errorf("price = %v, want %v", q.Price, price)
	}
	if q.Stale {
		t.Error("expected Stale=false")
	}
}
