package application

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
	refdomain "github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
)

func TestBatchImport_MappedProviderSymbolResolvesAndWritesSnapshot(t *testing.T) {
	batchRepo := newFakeBatchRepository()
	snapshotRepo := &fakeRepository{}
	resolver := newFakeResolver()
	provider := &perSymbolProvider{
		name: domain.ProviderYahoo,
		quotes: map[string]*domain.Quote{
			"KBANK.BK": {
				Symbol: "KBANK.BK",
				Price:  decimal.RequireFromString("142.50"),
				AsOf:   time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	// Reference: canonical KBANK with active Yahoo mapping.
	resolver.addSecurity(refdomain.Security{
		ID:            "sec-kbank",
		IMSSymbol:     "TH_EQ_XBKK_KBANK",
		DisplaySymbol: "KBANK.BK",
		Name:          "Kasikornbank",
		AssetType:     refdomain.AssetTypeEquity,
		Status:        refdomain.SecurityStatusActive,
		ProviderMappings: []refdomain.ProviderMapping{
			{
				ID:             "map-1",
				SecurityID:     "sec-kbank",
				ProviderCode:   "yahoo",
				ProviderSymbol: "KBANK.BK",
				MappingStatus:  refdomain.MappingStatusActive,
			},
		},
	})

	svc := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{provider}, snapshotRepo, snapshotRepo, nil)
	svc.SetImportBatchRepository(batchRepo)
	svc.SetSecurityResolver(resolver)

	created, err := svc.CreateImportBatch(context.Background(), CreateImportBatchRequest{
		Provider:     domain.ProviderYahoo,
		ImportType:   domain.ImportTypeQuoteSync,
		Symbols:      []string{"KBANK.BK"},
		ChunkSize:    25,
		IncludeQuote: true,
	})
	require.NoError(t, err)

	result, err := svc.RunImportBatch(context.Background(), created.BatchID)
	require.NoError(t, err)
	require.Equal(t, domain.ImportBatchStatusCompleted, result.Status)
	require.Equal(t, 1, result.AcceptedRecords)
	require.Equal(t, 0, result.RejectedRecords)

	// Snapshot was written.
	require.NotNil(t, snapshotRepo.savedQuote)
	require.Equal(t, "KBANK.BK", snapshotRepo.savedQuote.Symbol)

	// Item carries the canonical security_id.
	chunks := batchRepo.chunksByBatch(created.BatchID)
	require.Len(t, chunks, 1)
	items := batchRepo.items[chunks[0].ID]
	require.Len(t, items, 1)
	require.Equal(t, "sec-kbank", items[0].SecurityID)
	require.Equal(t, domain.ImportItemStatusAccepted, items[0].Status)
}

func TestBatchImport_UnmappedSymbolBecomesUnmappedAndCreatesCandidate(t *testing.T) {
	batchRepo := newFakeBatchRepository()
	snapshotRepo := &fakeRepository{}
	resolver := newFakeResolver()
	provider := &perSymbolProvider{
		name: domain.ProviderYahoo,
		quotes: map[string]*domain.Quote{
			"UNKNOWN.BK": {
				Symbol: "UNKNOWN.BK",
				Price:  decimal.RequireFromString("1.00"),
				AsOf:   time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	svc := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{provider}, snapshotRepo, snapshotRepo, nil)
	svc.SetImportBatchRepository(batchRepo)
	svc.SetSecurityResolver(resolver)

	created, err := svc.CreateImportBatch(context.Background(), CreateImportBatchRequest{
		Provider:     domain.ProviderYahoo,
		ImportType:   domain.ImportTypeQuoteSync,
		Symbols:      []string{"UNKNOWN.BK"},
		ChunkSize:    25,
		IncludeQuote: true,
	})
	require.NoError(t, err)

	result, err := svc.RunImportBatch(context.Background(), created.BatchID)
	require.NoError(t, err)
	// Whole chunk failed because the only item is unmapped.
	require.Equal(t, domain.ImportBatchStatusFailed, result.Status)
	require.Equal(t, 0, result.AcceptedRecords)
	require.Equal(t, 1, result.RejectedRecords)

	chunks := batchRepo.chunksByBatch(created.BatchID)
	require.Len(t, chunks, 1)
	items := batchRepo.items[chunks[0].ID]
	require.Len(t, items, 1)
	require.Equal(t, domain.ImportItemStatusUnmapped, items[0].Status)
	require.Equal(t, domain.ImportErrorCodeUnmapped, items[0].ErrorCode)

	// Snapshot was NOT written.
	require.Nil(t, snapshotRepo.savedQuote, "unmapped item must not trigger a snapshot write")

	// Candidate was created in reference_data.
	require.Len(t, resolver.candidates, 1)
	require.Equal(t, "UNKNOWN.BK", resolver.candidates[0].ProviderSymbol)
	require.Equal(t, "yahoo", resolver.candidates[0].ProviderCode)
}

func TestBatchImport_ResolverIdleWhenUnset(t *testing.T) {
	// Confirms backward compatibility: with no resolver wired the legacy
	// path still works (provider symbol passed straight through to adapter).
	batchRepo := newFakeBatchRepository()
	snapshotRepo := &fakeRepository{}
	provider := &perSymbolProvider{
		name: domain.ProviderYahoo,
		quotes: map[string]*domain.Quote{
			"AAPL": {Symbol: "AAPL", Price: decimal.RequireFromString("171.25"),
				AsOf: time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC)},
		},
	}
	svc := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{provider}, snapshotRepo, snapshotRepo, nil)
	svc.SetImportBatchRepository(batchRepo)

	created, err := svc.CreateImportBatch(context.Background(), CreateImportBatchRequest{
		Provider:     domain.ProviderYahoo,
		ImportType:   domain.ImportTypeQuoteSync,
		Symbols:      []string{"AAPL"},
		ChunkSize:    25,
		IncludeQuote: true,
	})
	require.NoError(t, err)

	result, err := svc.RunImportBatch(context.Background(), created.BatchID)
	require.NoError(t, err)
	require.Equal(t, domain.ImportBatchStatusCompleted, result.Status)
	require.Equal(t, 1, result.AcceptedRecords)
	require.NotNil(t, snapshotRepo.savedQuote)
}

// ---- fakes ----

type fakeResolver struct {
	mu         sync.Mutex
	securities []refdomain.Security
	candidates []refdomain.UnmappedSecurityCandidate
}

func newFakeResolver() *fakeResolver {
	return &fakeResolver{}
}

func (r *fakeResolver) addSecurity(s refdomain.Security) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.securities = append(r.securities, s)
}

func (r *fakeResolver) SearchSecurities(ctx context.Context, filter refdomain.SecuritySearchFilter) ([]refdomain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := append([]refdomain.Security(nil), r.securities...)
	return out, nil
}

func (r *fakeResolver) GetSecurityByID(ctx context.Context, id string) (*refdomain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.securities {
		if s.ID == id {
			clone := s
			return &clone, nil
		}
	}
	return nil, nil
}

func (r *fakeResolver) GetSecurityByIMSSymbol(ctx context.Context, ims string) (*refdomain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.securities {
		if s.IMSSymbol == ims {
			clone := s
			return &clone, nil
		}
	}
	return nil, nil
}

func (r *fakeResolver) GetSecurityByDisplaySymbol(ctx context.Context, ds string) (*refdomain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.securities {
		if s.DisplaySymbol == ds {
			clone := s
			return &clone, nil
		}
	}
	return nil, nil
}

func (r *fakeResolver) ResolveSecurityByProviderSymbol(ctx context.Context, providerCode, providerSymbol string) (*refdomain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.securities {
		for _, m := range s.ProviderMappings {
			if m.ProviderCode == providerCode && m.ProviderSymbol == providerSymbol && m.MappingStatus == refdomain.MappingStatusActive {
				clone := s
				return &clone, nil
			}
		}
	}
	return nil, nil
}

func (r *fakeResolver) ResolveProviderSymbol(ctx context.Context, securityID, providerCode string) (*refdomain.ProviderMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.securities {
		if s.ID != securityID {
			continue
		}
		for _, m := range s.ProviderMappings {
			if m.ProviderCode == providerCode && m.MappingStatus == refdomain.MappingStatusActive {
				clone := m
				return &clone, nil
			}
		}
	}
	return nil, nil
}

func (r *fakeResolver) CreateUnmappedCandidate(ctx context.Context, c refdomain.UnmappedSecurityCandidate) (*refdomain.UnmappedSecurityCandidate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c.ID = uuid.NewString()
	c.CandidateStatus = refdomain.CandidateStatusReviewRequired
	r.candidates = append(r.candidates, c)
	return &c, nil
}

func (r *fakeResolver) ListUnmappedCandidates(ctx context.Context, filter refdomain.UnmappedCandidateFilter) ([]refdomain.UnmappedSecurityCandidate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]refdomain.UnmappedSecurityCandidate(nil), r.candidates...), nil
}

func (r *fakeResolver) MapUnmappedCandidate(ctx context.Context, candidateID, securityID string) error {
	return nil
}

func (r *fakeResolver) RejectUnmappedCandidate(ctx context.Context, candidateID, reason string) error {
	return nil
}

// Compile-time assertion that fakeResolver satisfies the local interface and
// the reference_data resolver contract.
var _ SecurityResolver = (*fakeResolver)(nil)
var _ refdomain.SecurityResolver = (*fakeResolver)(nil)
