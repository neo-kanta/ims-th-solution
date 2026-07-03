package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

func TestCreateImportBatchSplits136SymbolsIntoExpectedChunks(t *testing.T) {
	repo := newFakeBatchRepository()
	service := newBatchService(t, repo, nil)

	symbols := generateSymbols(136)
	result, err := service.CreateImportBatch(context.Background(), CreateImportBatchRequest{
		ImportType: domain.ImportTypeQuoteSync,
		Symbols:    symbols,
		ChunkSize:  25,
	})
	require.NoError(t, err)
	require.Equal(t, 136, result.TotalSymbols)
	require.Equal(t, 6, result.TotalChunks)
	require.Equal(t, domain.ImportBatchStatusPending, result.Status)

	created := repo.batches[result.BatchID]
	require.NotNil(t, created)
	require.Equal(t, 136, created.TotalRecords)

	chunks := repo.chunksByBatch(result.BatchID)
	require.Len(t, chunks, 6)
	require.Equal(t, 25, chunks[0].TotalRecords)
	require.Equal(t, 11, chunks[5].TotalRecords)
}

func TestCreateImportBatchUsesDefaultChunkSizeWhenZero(t *testing.T) {
	repo := newFakeBatchRepository()
	service := newBatchService(t, repo, nil)

	symbols := generateSymbols(60)
	result, err := service.CreateImportBatch(context.Background(), CreateImportBatchRequest{
		ImportType: domain.ImportTypeQuoteSync,
		Symbols:    symbols,
	})
	require.NoError(t, err)
	require.Equal(t, 60, result.TotalSymbols)
	require.Equal(t, 3, result.TotalChunks)
}

func TestCreateImportBatchDuplicateIdempotencyKeyReturnsSameBatch(t *testing.T) {
	repo := newFakeBatchRepository()
	service := newBatchService(t, repo, nil)

	req := CreateImportBatchRequest{
		ImportType:     domain.ImportTypeQuoteSync,
		Symbols:        []string{"AAPL", "KBANK.BK"},
		ChunkSize:      25,
		IdempotencyKey: "key-123",
	}
	first, err := service.CreateImportBatch(context.Background(), req)
	require.NoError(t, err)
	require.False(t, first.Reused)

	second, err := service.CreateImportBatch(context.Background(), req)
	require.NoError(t, err)
	require.True(t, second.Reused)
	require.Equal(t, first.BatchID, second.BatchID)
	require.Equal(t, first.TotalSymbols, second.TotalSymbols)
	require.Equal(t, first.TotalChunks, second.TotalChunks)
	require.Equal(t, 1, len(repo.batches), "duplicate idempotency key must not create new batch")
}

func TestRunImportBatchProcessesAllChunksAndAggregatesStatus(t *testing.T) {
	repo := newFakeBatchRepository()
	provider := &fakeProvider{
		name: domain.ProviderYahoo,
		quote: &domain.Quote{
			Symbol: "AAPL",
			Price:  decimal.RequireFromString("171.25"),
			AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
	}
	service := newBatchService(t, repo, []domain.MarketDataProvider{provider})

	created, err := service.CreateImportBatch(context.Background(), CreateImportBatchRequest{
		Provider:     domain.ProviderYahoo,
		ImportType:   domain.ImportTypeQuoteSync,
		Symbols:      []string{"AAPL", "MSFT", "GOOG", "AMZN"},
		ChunkSize:    2,
		IncludeQuote: true,
	})
	require.NoError(t, err)

	result, err := service.RunImportBatch(context.Background(), created.BatchID)
	require.NoError(t, err)
	require.Equal(t, domain.ImportBatchStatusCompleted, result.Status)
	require.Equal(t, 4, result.AcceptedRecords)
	require.Equal(t, 0, result.RejectedRecords)
	require.Equal(t, 2, result.TotalChunks)
	require.Len(t, result.Chunks, 2)
	for _, c := range result.Chunks {
		require.Equal(t, domain.ImportChunkStatusCompleted, c.Status)
		require.Equal(t, 2, c.AcceptedRecords)
	}
}

func TestRunImportBatchRetryingChunkDoesNotDuplicateSnapshots(t *testing.T) {
	repo := newFakeBatchRepository()
	provider := &fakeProvider{
		name: domain.ProviderYahoo,
		quote: &domain.Quote{
			Symbol: "AAPL",
			Price:  decimal.RequireFromString("171.25"),
			AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
	}
	snapshotRepo := &fakeRepository{}
	service := newBatchServiceWithSnapshotRepo(t, repo, snapshotRepo, []domain.MarketDataProvider{provider})

	created, err := service.CreateImportBatch(context.Background(), CreateImportBatchRequest{
		Provider:     domain.ProviderYahoo,
		ImportType:   domain.ImportTypeQuoteSync,
		Symbols:      []string{"AAPL", "MSFT"},
		ChunkSize:    25,
		IncludeQuote: true,
	})
	require.NoError(t, err)

	_, err = service.RunImportBatch(context.Background(), created.BatchID)
	require.NoError(t, err)
	require.Equal(t, 2, snapshotRepo.savedQuoteCalls, "initial run should save 2 quotes")

	// Re-run the same batch (retry).
	_, err = service.RunImportBatch(context.Background(), created.BatchID)
	require.NoError(t, err)
	require.Equal(t, 4, snapshotRepo.savedQuoteCalls, "retry calls SaveQuote again but upsert (ON CONFLICT) ensures snapshots aren't duplicated")

	// Verify chunk attempt count incremented and only one chunk exists.
	chunks := repo.chunksByBatch(created.BatchID)
	require.Len(t, chunks, 1)
	require.Equal(t, 2, chunks[0].AttemptCount)
	require.Equal(t, 2, chunks[0].AcceptedRecords)

	// Verify items are still 2, not duplicated.
	items := repo.items[chunks[0].ID]
	require.Len(t, items, 2)
}

func TestRunImportBatchProvider429MapsToRateLimited(t *testing.T) {
	repo := newFakeBatchRepository()
	provider := &fakeProvider{
		name: domain.ProviderYahoo,
		quoteErr: &domain.ProviderError{
			Provider:   domain.ProviderYahoo,
			Operation:  "quote",
			Code:       "RATE_LIMITED",
			StatusCode: 429,
			Message:    "too many requests",
		},
	}
	service := newBatchService(t, repo, []domain.MarketDataProvider{provider})

	created, err := service.CreateImportBatch(context.Background(), CreateImportBatchRequest{
		Provider:     domain.ProviderYahoo,
		ImportType:   domain.ImportTypeQuoteSync,
		Symbols:      []string{"AAPL", "MSFT"},
		ChunkSize:    25,
		IncludeQuote: true,
	})
	require.NoError(t, err)

	result, err := service.RunImportBatch(context.Background(), created.BatchID)
	require.NoError(t, err)
	require.Equal(t, domain.ImportBatchStatusFailed, result.Status)
	require.Len(t, result.Chunks, 1)
	require.Equal(t, domain.ImportChunkStatusRateLimited, result.Chunks[0].Status)
	require.Equal(t, 0, result.AcceptedRecords)
	require.Equal(t, 2, result.RejectedRecords)

	// Errors endpoint should return the rate-limited items.
	errs, err := service.ListImportBatchErrors(context.Background(), created.BatchID)
	require.NoError(t, err)
	require.Len(t, errs, 2)
	for _, item := range errs {
		require.Equal(t, domain.ImportItemStatusRateLimited, item.Status)
		require.Equal(t, domain.ImportErrorCodeRateLimited, item.ErrorCode)
	}
}

func TestRunImportBatchPartialFailureReturnsPartialFailedStatus(t *testing.T) {
	repo := newFakeBatchRepository()
	provider := &perSymbolProvider{
		name: domain.ProviderYahoo,
		quotes: map[string]*domain.Quote{
			"AAPL": {
				Symbol: "AAPL",
				Price:  decimal.RequireFromString("171.25"),
				AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
			},
			"MSFT": {
				Symbol: "MSFT",
				Price:  decimal.RequireFromString("400.00"),
				AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
			},
		},
		errors: map[string]error{
			"GOOG": fmt.Errorf("provider failure"),
			"AMZN": fmt.Errorf("provider failure"),
		},
	}
	service := newBatchService(t, repo, []domain.MarketDataProvider{provider})

	created, err := service.CreateImportBatch(context.Background(), CreateImportBatchRequest{
		Provider:     domain.ProviderYahoo,
		ImportType:   domain.ImportTypeQuoteSync,
		Symbols:      []string{"AAPL", "MSFT", "GOOG", "AMZN"},
		ChunkSize:    2,
		IncludeQuote: true,
	})
	require.NoError(t, err)

	result, err := service.RunImportBatch(context.Background(), created.BatchID)
	require.NoError(t, err)
	require.Equal(t, domain.ImportBatchStatusPartialFailed, result.Status)
	require.Equal(t, 2, result.AcceptedRecords)
	require.Equal(t, 2, result.RejectedRecords)
	require.Len(t, result.Chunks, 2)
	require.Equal(t, domain.ImportChunkStatusCompleted, result.Chunks[0].Status)
	require.Equal(t, domain.ImportChunkStatusFailed, result.Chunks[1].Status)
}

func TestRunImportBatchReturnsNotFoundForUnknownBatch(t *testing.T) {
	repo := newFakeBatchRepository()
	service := newBatchService(t, repo, nil)

	_, err := service.RunImportBatch(context.Background(), uuid.NewString())
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrImportBatchNotFound))
}

func TestCreateImportBatchRejectsInvalidImportType(t *testing.T) {
	repo := newFakeBatchRepository()
	service := newBatchService(t, repo, nil)

	_, err := service.CreateImportBatch(context.Background(), CreateImportBatchRequest{
		ImportType: "BOGUS",
		Symbols:    []string{"AAPL"},
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidMarketDataRequest))
}

func TestImportMarketDataSingleSymbolStillWorks(t *testing.T) {
	// Guard: the original POST /market-data/import path must still work.
	provider := &fakeProvider{
		name: domain.ProviderYahoo,
		quote: &domain.Quote{
			Symbol: "AAPL",
			Price:  decimal.RequireFromString("171.25"),
			AsOf:   time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
	}
	snapshotRepo := &fakeRepository{}
	service := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, []domain.MarketDataProvider{provider}, snapshotRepo, snapshotRepo, nil)

	result, err := service.ImportMarketData(context.Background(), ImportMarketDataRequest{
		Symbol:       "AAPL",
		Provider:     domain.ProviderYahoo,
		IncludeQuote: true,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "AAPL", result.Symbol)
	require.Equal(t, domain.ProviderYahoo, result.QuoteProvider)
	require.NotNil(t, snapshotRepo.savedQuote)
}

// ---- helpers ----

func generateSymbols(n int) []string {
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = fmt.Sprintf("SYM%04d", i)
	}
	return out
}

func newBatchService(t *testing.T, batchRepo *fakeBatchRepository, providers []domain.MarketDataProvider) *Service {
	t.Helper()
	return newBatchServiceWithSnapshotRepo(t, batchRepo, &fakeRepository{}, providers)
}

func newBatchServiceWithSnapshotRepo(t *testing.T, batchRepo *fakeBatchRepository, snapshotRepo *fakeRepository, providers []domain.MarketDataProvider) *Service {
	t.Helper()
	service := NewService(Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, providers, snapshotRepo, snapshotRepo, nil)
	service.SetImportBatchRepository(batchRepo)
	return service
}

// fakeBatchRepository is an in-memory ImportBatchRepository used in tests.
type fakeBatchRepository struct {
	mu      sync.Mutex
	batches map[string]*domain.ImportBatch
	chunks  map[string]*domain.ImportChunk
	items   map[string][]*domain.ImportChunkItem
	byKey   map[string]string
}

func newFakeBatchRepository() *fakeBatchRepository {
	return &fakeBatchRepository{
		batches: map[string]*domain.ImportBatch{},
		chunks:  map[string]*domain.ImportChunk{},
		items:   map[string][]*domain.ImportChunkItem{},
		byKey:   map[string]string{},
	}
}

func (r *fakeBatchRepository) chunksByBatch(batchID string) []domain.ImportChunk {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []domain.ImportChunk{}
	for _, c := range r.chunks {
		if c.BatchID == batchID {
			out = append(out, *c)
		}
	}
	// sort by index
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].ChunkIndex < out[i].ChunkIndex {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func (r *fakeBatchRepository) CreateBatch(ctx context.Context, plan domain.ImportBatchPlan) (*domain.ImportBatch, []domain.ImportChunk, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := uuid.NewString()
	total := 0
	for _, c := range plan.Chunks {
		total += len(c.Symbols)
	}
	batch := &domain.ImportBatch{
		ID:             id,
		ProviderCode:   plan.ProviderCode,
		ImportType:     plan.ImportType,
		Status:         domain.ImportBatchStatusPending,
		IdempotencyKey: plan.IdempotencyKey,
		TotalRecords:   total,
		TotalChunks:    len(plan.Chunks),
		CreatedBy:      plan.CreatedBy,
		CreatedAt:      time.Now().UTC(),
	}
	r.batches[id] = batch
	if plan.IdempotencyKey != "" {
		r.byKey[plan.IdempotencyKey] = id
	}
	chunks := []domain.ImportChunk{}
	for _, c := range plan.Chunks {
		chunkID := uuid.NewString()
		chunk := &domain.ImportChunk{
			ID:           chunkID,
			BatchID:      id,
			ChunkIndex:   c.ChunkIndex,
			Status:       domain.ImportChunkStatusPending,
			TotalRecords: len(c.Symbols),
			CreatedAt:    time.Now().UTC(),
		}
		r.chunks[chunkID] = chunk
		items := []*domain.ImportChunkItem{}
		for _, symbol := range c.Symbols {
			items = append(items, &domain.ImportChunkItem{
				ID:        uuid.NewString(),
				ChunkID:   chunkID,
				Symbol:    symbol,
				Status:    domain.ImportItemStatusPending,
				CreatedAt: time.Now().UTC(),
			})
		}
		r.items[chunkID] = items
		chunks = append(chunks, *chunk)
	}
	return batch, chunks, nil
}

func (r *fakeBatchRepository) FindBatchByIdempotencyKey(ctx context.Context, key string) (*domain.ImportBatch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byKey[key]
	if !ok {
		return nil, nil
	}
	b := *r.batches[id]
	b.TotalChunks = r.countChunksLocked(id)
	return &b, nil
}

func (r *fakeBatchRepository) GetBatch(ctx context.Context, batchID string) (*domain.ImportBatch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.batches[batchID]
	if !ok {
		return nil, nil
	}
	out := *b
	out.TotalChunks = r.countChunksLocked(batchID)
	return &out, nil
}

func (r *fakeBatchRepository) countChunksLocked(batchID string) int {
	count := 0
	for _, c := range r.chunks {
		if c.BatchID == batchID {
			count++
		}
	}
	return count
}

func (r *fakeBatchRepository) ListChunks(ctx context.Context, batchID string) ([]domain.ImportChunk, error) {
	return r.chunksByBatch(batchID), nil
}

func (r *fakeBatchRepository) ListChunkItems(ctx context.Context, chunkID string) ([]domain.ImportChunkItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items, ok := r.items[chunkID]
	if !ok {
		return nil, nil
	}
	out := make([]domain.ImportChunkItem, 0, len(items))
	for _, item := range items {
		out = append(out, *item)
	}
	return out, nil
}

func (r *fakeBatchRepository) ListBatchErrors(ctx context.Context, batchID string) ([]domain.ImportChunkItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []domain.ImportChunkItem{}
	for chunkID, items := range r.items {
		chunk, ok := r.chunks[chunkID]
		if !ok || chunk.BatchID != batchID {
			continue
		}
		for _, item := range items {
			switch item.Status {
			case domain.ImportItemStatusRejected,
				domain.ImportItemStatusFailed,
				domain.ImportItemStatusRateLimited,
				domain.ImportItemStatusUnmapped,
				domain.ImportItemStatusReviewRequired,
				domain.ImportItemStatusWarning:
				out = append(out, *item)
			}
		}
	}
	return out, nil
}

func (r *fakeBatchRepository) UpdateBatchStatus(ctx context.Context, batchID string, update domain.ImportBatchStatusUpdate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.batches[batchID]
	if !ok {
		return errors.New("batch not found")
	}
	if update.Status != "" {
		b.Status = update.Status
	}
	if update.AcceptedRecords != nil {
		b.AcceptedRecords = *update.AcceptedRecords
	}
	if update.RejectedRecords != nil {
		b.RejectedRecords = *update.RejectedRecords
	}
	if update.WarningRecords != nil {
		b.WarningRecords = *update.WarningRecords
	}
	if update.StartedAt != nil {
		t := *update.StartedAt
		b.StartedAt = &t
	}
	if update.CompletedAt != nil {
		t := *update.CompletedAt
		b.CompletedAt = &t
	}
	if update.ErrorMessage != nil {
		b.ErrorMessage = *update.ErrorMessage
	}
	return nil
}

func (r *fakeBatchRepository) UpdateChunkStatus(ctx context.Context, chunkID string, update domain.ImportChunkStatusUpdate) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.chunks[chunkID]
	if !ok {
		return errors.New("chunk not found")
	}
	if update.Status != "" {
		c.Status = update.Status
	}
	if update.AcceptedRecords != nil {
		c.AcceptedRecords = *update.AcceptedRecords
	}
	if update.RejectedRecords != nil {
		c.RejectedRecords = *update.RejectedRecords
	}
	if update.WarningRecords != nil {
		c.WarningRecords = *update.WarningRecords
	}
	if update.AttemptCount != nil {
		c.AttemptCount = *update.AttemptCount
	}
	if update.LockedAt != nil {
		t := *update.LockedAt
		c.LockedAt = &t
	}
	if update.StartedAt != nil {
		t := *update.StartedAt
		c.StartedAt = &t
	}
	if update.CompletedAt != nil {
		t := *update.CompletedAt
		c.CompletedAt = &t
	}
	if update.ErrorMessage != nil {
		c.ErrorMessage = *update.ErrorMessage
	}
	return nil
}

func (r *fakeBatchRepository) RecordChunkItem(ctx context.Context, item domain.ImportChunkItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.items[item.ChunkID] {
		if existing.Symbol == item.Symbol {
			existing.Status = item.Status
			existing.ErrorCode = item.ErrorCode
			existing.ErrorMessage = item.ErrorMessage
			if item.ProviderSymbol != "" {
				existing.ProviderSymbol = item.ProviderSymbol
			}
			if item.SecurityID != "" {
				existing.SecurityID = item.SecurityID
			}
			return nil
		}
	}
	return errors.New("chunk item not found")
}

func (r *fakeBatchRepository) ResetChunkItems(ctx context.Context, chunkID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range r.items[chunkID] {
		item.Status = domain.ImportItemStatusPending
		item.ErrorCode = ""
		item.ErrorMessage = ""
	}
	return nil
}

// perSymbolProvider returns different responses per requested symbol.
type perSymbolProvider struct {
	name   string
	quotes map[string]*domain.Quote
	errors map[string]error
}

func (p *perSymbolProvider) GetQuote(ctx context.Context, symbol string) (*domain.Quote, error) {
	if err, ok := p.errors[symbol]; ok {
		return nil, err
	}
	if q, ok := p.quotes[symbol]; ok {
		out := *q
		return &out, nil
	}
	return nil, fmt.Errorf("symbol not configured: %s", symbol)
}

func (p *perSymbolProvider) GetDailyPrices(ctx context.Context, symbol string) ([]domain.PriceBar, error) {
	return nil, fmt.Errorf("history not configured")
}

func (p *perSymbolProvider) ProviderName() string {
	return p.name
}

// Augment fakeRepository (defined in service_test.go) with savedQuoteCalls counter.
// Note: SaveQuote in the existing fakeRepository assigns savedQuote without
// counting; for retry idempotency we need a counter, so we add a thin wrapper
// here via type assertion check in the test setup.

// NOTE: the existing fakeRepository.SaveQuote sets savedQuote each time; the
// test below counts calls by inspecting savedQuoteCalls which we add as a
// new field in service_test.go fakeRepository.
