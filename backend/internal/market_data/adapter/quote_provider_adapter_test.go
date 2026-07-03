package adapter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	mdapp "github.com/neo-kanta/ims-th-solution/backend/internal/market_data/application"
	mddomain "github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// snapshotOnlyRepo is a minimal domain.SnapshotRepository fake — only
// GetSnapshotAsOf is exercised by these tests.
type snapshotOnlyRepo struct {
	snapshot *mddomain.Quote
}

func (r *snapshotOnlyRepo) UpsertSymbol(context.Context, mddomain.SymbolMapping) error { return nil }
func (r *snapshotOnlyRepo) GetSymbolMapping(context.Context, string) (*mddomain.SymbolMapping, error) {
	return nil, nil
}
func (r *snapshotOnlyRepo) SaveQuote(context.Context, mddomain.Quote) error { return nil }
func (r *snapshotOnlyRepo) SaveDailyPrices(context.Context, string, string, []mddomain.PriceBar) error {
	return nil
}
func (r *snapshotOnlyRepo) GetLatestQuote(context.Context, string) (*mddomain.Quote, error) {
	return nil, nil
}
func (r *snapshotOnlyRepo) ListDailyPrices(context.Context, string, int) ([]mddomain.PriceBar, error) {
	return nil, nil
}
func (r *snapshotOnlyRepo) GetSnapshotAsOf(_ context.Context, _ string, _ time.Time) (*mddomain.Quote, error) {
	if r.snapshot == nil {
		return nil, nil
	}
	q := *r.snapshot
	return &q, nil
}
func (r *snapshotOnlyRepo) LogProviderRequest(context.Context, mddomain.ProviderRequestLog) error {
	return nil
}
func (r *snapshotOnlyRepo) ListLatestProviderRequests(context.Context, int) ([]mddomain.ProviderRequestLog, error) {
	return nil, nil
}

// cachedQuoteOnlyRepo simulates total provider failure: GetQuote's live fetch
// always fails (no providers registered with the service), so the service
// falls back to repo.GetLatestQuote — the last cached snapshot row.
type cachedQuoteOnlyRepo struct {
	snapshotOnlyRepo
	cached *mddomain.Quote
}

func (r *cachedQuoteOnlyRepo) GetLatestQuote(context.Context, string) (*mddomain.Quote, error) {
	if r.cached == nil {
		return nil, nil
	}
	q := *r.cached
	return &q, nil
}

// Regression: when every provider fails, the service returns the cached
// snapshot with Stale=true — that must be labeled as a snapshot source, not
// "live_quote". Mislabeling it live would make a provider outage look like a
// fresh mark-to-market read on the frontend's stale badge.
func TestGetLatestQuoteLabelsCachedFallbackAsSnapshotSource(t *testing.T) {
	repo := &cachedQuoteOnlyRepo{cached: &mddomain.Quote{
		Symbol: "AOT.BK", Provider: "internal_demo",
		Price: decimal.RequireFromString("65.00"), Currency: "THB",
		AsOf: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), SnapshotID: "snap-999",
	}}
	svc := mdapp.NewService(mdapp.Config{PrimaryProvider: mddomain.ProviderAlphaVantage}, nil, repo, repo, nil)
	a := NewQuoteProviderAdapter(svc)

	q, err := a.GetLatestQuote(context.Background(), "AOT.BK")
	require.NoError(t, err)
	require.NotNil(t, q)
	require.True(t, q.Stale)
	require.Equal(t, contract.QuoteSourceMarketDataSnapshot, q.Source,
		"a stale cached fallback must not be labeled as a live quote")
}

func TestTranslateErrorMapsTypedProviderErrorsToContract(t *testing.T) {
	cases := []struct {
		name    string
		input   error
		wantErr error
	}{
		{"rate_limited", &mddomain.ProviderError{Provider: "alpha_vantage", Code: "rate_limited", Message: "throttled"}, contract.ErrRateLimited},
		{"unauthorized", &mddomain.ProviderError{Provider: "alpha_vantage", Code: "unauthorized", Message: "bad key"}, contract.ErrProviderUnauthorized},
		{"missing_api_key", &mddomain.ProviderError{Provider: "alpha_vantage", Code: "missing_api_key", Message: "no key"}, contract.ErrProviderUnauthorized},
		{"timeout", &mddomain.ProviderError{Provider: "yahoo", Code: "timeout", Message: "slow"}, contract.ErrProviderTimeout},
		{"symbol_not_found", &mddomain.ProviderError{Provider: "yahoo", Code: "symbol_not_found", Message: "no such symbol"}, contract.ErrSymbolNotMapped},
		{"malformed_response", &mddomain.ProviderError{Provider: "yahoo", Code: "malformed_response", Message: "bad json"}, contract.ErrMalformedResponse},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := translateError(c.input)
			require.True(t, errors.Is(got, c.wantErr), "translated error = %v; want wrap of %v", got, c.wantErr)
		})
	}
}

func TestTranslateErrorMapsContextDeadlineExceededToTimeout(t *testing.T) {
	got := translateError(context.DeadlineExceeded)
	require.True(t, errors.Is(got, contract.ErrProviderTimeout))
}

func TestTranslateErrorWrapsUnknownAsUnavailable(t *testing.T) {
	got := translateError(errors.New("kapow"))
	require.True(t, errors.Is(got, contract.ErrMarketDataUnavailable))
}

// Latest snapshot dated exactly the requested business_date is a fresh,
// non-stale mark-to-market read — the snapshot table is the source of truth,
// no live call is involved.
func TestGetQuoteAsOfFreshWhenSnapshotMatchesBusinessDate(t *testing.T) {
	repo := &snapshotOnlyRepo{snapshot: &mddomain.Quote{
		Symbol: "TH-LB30DA", Provider: "internal_demo",
		Price: decimal.RequireFromString("1032.00"), Currency: "THB",
		AsOf: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}}
	svc := mdapp.NewService(mdapp.Config{PrimaryProvider: mddomain.ProviderAlphaVantage}, nil, repo, repo, nil)
	a := NewQuoteProviderAdapter(svc)

	q, err := a.GetQuoteAsOf(context.Background(), "TH-LB30DA", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NotNil(t, q)
	require.False(t, q.Stale, "snapshot dated exactly business_date must not be stale")
	require.Empty(t, q.StaleReason)
	require.Equal(t, contract.QuoteSourceMarketDataSnapshot, q.Source)
}

// A snapshot older than the requested business_date is still the "latest
// valid on or before" mark (carried forward across a weekend/holiday), but
// must be flagged stale so the caller can label it as carried-forward.
func TestGetQuoteAsOfStaleWhenSnapshotOlderThanBusinessDate(t *testing.T) {
	repo := &snapshotOnlyRepo{snapshot: &mddomain.Quote{
		Symbol: "TH-LB30DA", Provider: "internal_demo",
		Price: decimal.RequireFromString("1030.00"), Currency: "THB",
		AsOf: time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC),
	}}
	svc := mdapp.NewService(mdapp.Config{PrimaryProvider: mddomain.ProviderAlphaVantage}, nil, repo, repo, nil)
	a := NewQuoteProviderAdapter(svc)

	q, err := a.GetQuoteAsOf(context.Background(), "TH-LB30DA", time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NotNil(t, q)
	require.True(t, q.Stale)
	require.Contains(t, q.StaleReason, "2026-06-28")
}

// No snapshot at all on or before the date must degrade gracefully —
// ErrMarketDataUnavailable, never a panic/crash — so the caller can fall
// through to the next pricing tier instead of a 500.
func TestGetQuoteAsOfReturnsMarketDataUnavailableWhenNoSnapshot(t *testing.T) {
	repo := &snapshotOnlyRepo{}
	svc := mdapp.NewService(mdapp.Config{PrimaryProvider: mddomain.ProviderAlphaVantage}, nil, repo, repo, nil)
	a := NewQuoteProviderAdapter(svc)

	q, err := a.GetQuoteAsOf(context.Background(), "TH-LB30DA", time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))
	require.Nil(t, q)
	require.True(t, errors.Is(err, contract.ErrMarketDataUnavailable))
}
