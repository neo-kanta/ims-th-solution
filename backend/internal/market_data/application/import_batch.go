package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
	refdomain "github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
)

const (
	defaultBatchChunkSize = 25
	maxBatchChunkSize     = 200
	maxBatchSymbols       = 10000
)

var ErrImportBatchNotFound = errors.New("import batch not found")

type CreateImportBatchRequest struct {
	Provider       string
	ImportType     string
	Symbols        []string
	ChunkSize      int
	IncludeQuote   bool
	IncludeHistory bool
	HistoryLimit   int
	IdempotencyKey string
	CreatedBy      string
}

type CreateImportBatchResult struct {
	BatchID      string `json:"batch_id"`
	Status       string `json:"status"`
	TotalSymbols int    `json:"total_symbols"`
	TotalChunks  int    `json:"total_chunks"`
	Reused       bool   `json:"reused,omitempty"`
}

type RunImportBatchResult struct {
	BatchID         string                   `json:"batch_id"`
	Status          string                   `json:"status"`
	TotalSymbols    int                      `json:"total_symbols"`
	TotalChunks     int                      `json:"total_chunks"`
	AcceptedRecords int                      `json:"accepted_records"`
	RejectedRecords int                      `json:"rejected_records"`
	WarningRecords  int                      `json:"warning_records"`
	Chunks          []ImportChunkSummary     `json:"chunks"`
	Errors          []domain.ImportChunkItem `json:"errors,omitempty"`
}

type ImportChunkSummary struct {
	ChunkID         string `json:"chunk_id"`
	ChunkIndex      int    `json:"chunk_index"`
	Status          string `json:"status"`
	TotalRecords    int    `json:"total_records"`
	AcceptedRecords int    `json:"accepted_records"`
	RejectedRecords int    `json:"rejected_records"`
	WarningRecords  int    `json:"warning_records"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

type ImportBatchStatusResponse struct {
	Batch  domain.ImportBatch       `json:"batch"`
	Chunks []ImportChunkSummary     `json:"chunks"`
	Errors []domain.ImportChunkItem `json:"errors,omitempty"`
}

func (s *Service) CreateImportBatch(ctx context.Context, req CreateImportBatchRequest) (*CreateImportBatchResult, error) {
	if s.batchRepo == nil {
		return nil, fmt.Errorf("%w: import batch repository not configured", ErrInvalidMarketDataRequest)
	}

	importType := strings.ToUpper(strings.TrimSpace(req.ImportType))
	if importType == "" {
		switch {
		case req.IncludeQuote && req.IncludeHistory:
			importType = domain.ImportTypeQuoteAndHistorySync
		case req.IncludeHistory:
			importType = domain.ImportTypeHistorySync
		default:
			importType = domain.ImportTypeQuoteSync
		}
	}
	if !validImportType(importType) {
		return nil, fmt.Errorf("%w: unsupported import_type %q", ErrInvalidMarketDataRequest, importType)
	}

	includeQuote := req.IncludeQuote
	includeHistory := req.IncludeHistory
	switch importType {
	case domain.ImportTypeQuoteSync:
		includeQuote = true
	case domain.ImportTypeHistorySync:
		includeHistory = true
	case domain.ImportTypeQuoteAndHistorySync:
		includeQuote = true
		includeHistory = true
	}
	if !includeQuote && !includeHistory {
		includeQuote = true
	}

	if len(req.Symbols) == 0 {
		return nil, fmt.Errorf("%w: at least one symbol is required", ErrInvalidMarketDataRequest)
	}
	if len(req.Symbols) > maxBatchSymbols {
		return nil, fmt.Errorf("%w: too many symbols (max %d)", ErrInvalidMarketDataRequest, maxBatchSymbols)
	}

	symbols, err := normalizeSymbols(req.Symbols)
	if err != nil {
		return nil, err
	}

	chunkSize := req.ChunkSize
	if chunkSize <= 0 {
		chunkSize = defaultBatchChunkSize
	}
	if chunkSize > maxBatchChunkSize {
		chunkSize = maxBatchChunkSize
	}

	providerCode := strings.TrimSpace(req.Provider)
	if providerCode == "" {
		providerCode = "default"
	}
	if providerCode != "default" {
		if _, _, err := s.providerOrderForSelection(providerCode); err != nil {
			return nil, err
		}
	}

	if req.IdempotencyKey != "" {
		existing, err := s.batchRepo.FindBatchByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return &CreateImportBatchResult{
				BatchID:      existing.ID,
				Status:       existing.Status,
				TotalSymbols: existing.TotalRecords,
				TotalChunks:  existing.TotalChunks,
				Reused:       true,
			}, nil
		}
	}

	chunks := chunkSymbols(symbols, chunkSize)
	plan := domain.ImportBatchPlan{
		ProviderCode:   providerCode,
		ImportType:     importType,
		IncludeQuote:   includeQuote,
		IncludeHistory: includeHistory,
		HistoryLimit:   req.HistoryLimit,
		IdempotencyKey: req.IdempotencyKey,
		CreatedBy:      req.CreatedBy,
		Chunks:         chunks,
	}
	batch, createdChunks, err := s.batchRepo.CreateBatch(ctx, plan)
	if err != nil {
		return nil, err
	}
	return &CreateImportBatchResult{
		BatchID:      batch.ID,
		Status:       batch.Status,
		TotalSymbols: batch.TotalRecords,
		TotalChunks:  len(createdChunks),
	}, nil
}

func (s *Service) RunImportBatch(ctx context.Context, batchID string) (*RunImportBatchResult, error) {
	if s.batchRepo == nil {
		return nil, fmt.Errorf("%w: import batch repository not configured", ErrInvalidMarketDataRequest)
	}

	batch, err := s.batchRepo.GetBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, ErrImportBatchNotFound
	}

	chunks, err := s.batchRepo.ListChunks(ctx, batchID)
	if err != nil {
		return nil, err
	}

	startedAt := s.now()
	_ = s.batchRepo.UpdateBatchStatus(ctx, batchID, domain.ImportBatchStatusUpdate{
		Status:    domain.ImportBatchStatusRunning,
		StartedAt: &startedAt,
	})

	order, _, err := s.providerOrderForSelection(batch.ProviderCode)
	if err != nil {
		errMsg := err.Error()
		_ = s.batchRepo.UpdateBatchStatus(ctx, batchID, domain.ImportBatchStatusUpdate{
			Status:       domain.ImportBatchStatusFailed,
			ErrorMessage: &errMsg,
		})
		return nil, err
	}

	includeQuote, includeHistory, historyLimit := importFlagsFor(batch.ImportType)

	totalAccepted := 0
	totalRejected := 0
	totalWarnings := 0
	chunkSummaries := make([]ImportChunkSummary, 0, len(chunks))
	failedChunks := 0
	succeededChunks := 0

	for _, chunk := range chunks {
		summary := s.runChunk(ctx, chunk, batch, order, includeQuote, includeHistory, historyLimit)
		chunkSummaries = append(chunkSummaries, summary)
		totalAccepted += summary.AcceptedRecords
		totalRejected += summary.RejectedRecords
		totalWarnings += summary.WarningRecords
		switch summary.Status {
		case domain.ImportChunkStatusCompleted, domain.ImportChunkStatusCompletedWithWarnings:
			succeededChunks++
		case domain.ImportChunkStatusFailed, domain.ImportChunkStatusRateLimited:
			failedChunks++
		}
	}

	finalStatus := finalBatchStatus(len(chunks), succeededChunks, failedChunks, totalRejected, totalWarnings)
	completedAt := s.now()
	_ = s.batchRepo.UpdateBatchStatus(ctx, batchID, domain.ImportBatchStatusUpdate{
		Status:          finalStatus,
		AcceptedRecords: &totalAccepted,
		RejectedRecords: &totalRejected,
		WarningRecords:  &totalWarnings,
		CompletedAt:     &completedAt,
	})

	errs, _ := s.batchRepo.ListBatchErrors(ctx, batchID)

	return &RunImportBatchResult{
		BatchID:         batchID,
		Status:          finalStatus,
		TotalSymbols:    batch.TotalRecords,
		TotalChunks:     len(chunks),
		AcceptedRecords: totalAccepted,
		RejectedRecords: totalRejected,
		WarningRecords:  totalWarnings,
		Chunks:          chunkSummaries,
		Errors:          errs,
	}, nil
}

func (s *Service) GetImportBatch(ctx context.Context, batchID string) (*ImportBatchStatusResponse, error) {
	if s.batchRepo == nil {
		return nil, fmt.Errorf("%w: import batch repository not configured", ErrInvalidMarketDataRequest)
	}
	batch, err := s.batchRepo.GetBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, ErrImportBatchNotFound
	}
	chunks, err := s.batchRepo.ListChunks(ctx, batchID)
	if err != nil {
		return nil, err
	}
	summaries := make([]ImportChunkSummary, 0, len(chunks))
	for _, c := range chunks {
		summaries = append(summaries, ImportChunkSummary{
			ChunkID:         c.ID,
			ChunkIndex:      c.ChunkIndex,
			Status:          c.Status,
			TotalRecords:    c.TotalRecords,
			AcceptedRecords: c.AcceptedRecords,
			RejectedRecords: c.RejectedRecords,
			WarningRecords:  c.WarningRecords,
			ErrorMessage:    c.ErrorMessage,
		})
	}
	errs, _ := s.batchRepo.ListBatchErrors(ctx, batchID)
	resp := &ImportBatchStatusResponse{
		Batch:  *batch,
		Chunks: summaries,
		Errors: errs,
	}
	resp.Batch.TotalChunks = len(chunks)
	return resp, nil
}

func (s *Service) ListImportBatchErrors(ctx context.Context, batchID string) ([]domain.ImportChunkItem, error) {
	if s.batchRepo == nil {
		return nil, fmt.Errorf("%w: import batch repository not configured", ErrInvalidMarketDataRequest)
	}
	batch, err := s.batchRepo.GetBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, ErrImportBatchNotFound
	}
	return s.batchRepo.ListBatchErrors(ctx, batchID)
}

func (s *Service) runChunk(ctx context.Context, chunk domain.ImportChunk, batch *domain.ImportBatch, order []string, includeQuote, includeHistory bool, historyLimit int) ImportChunkSummary {
	startedAt := s.now()
	attempt := chunk.AttemptCount + 1
	_ = s.batchRepo.UpdateChunkStatus(ctx, chunk.ID, domain.ImportChunkStatusUpdate{
		Status:       domain.ImportChunkStatusRunning,
		AttemptCount: &attempt,
		StartedAt:    &startedAt,
		LockedAt:     &startedAt,
	})
	_ = s.batchRepo.ResetChunkItems(ctx, chunk.ID)

	items, err := s.batchRepo.ListChunkItems(ctx, chunk.ID)
	if err != nil || len(items) == 0 {
		errMsg := "no items in chunk"
		if err != nil {
			errMsg = err.Error()
		}
		completedAt := s.now()
		_ = s.batchRepo.UpdateChunkStatus(ctx, chunk.ID, domain.ImportChunkStatusUpdate{
			Status:       domain.ImportChunkStatusFailed,
			CompletedAt:  &completedAt,
			ErrorMessage: &errMsg,
		})
		return ImportChunkSummary{
			ChunkID:      chunk.ID,
			ChunkIndex:   chunk.ChunkIndex,
			Status:       domain.ImportChunkStatusFailed,
			TotalRecords: chunk.TotalRecords,
			ErrorMessage: errMsg,
		}
	}

	accepted := 0
	rejected := 0
	warnings := 0
	rateLimitedChunk := false
	providerSuccessAny := false

	for _, item := range items {
		outcome := s.processItem(ctx, item, batch, order, includeQuote, includeHistory, historyLimit)
		_ = s.batchRepo.RecordChunkItem(ctx, outcome.Item)
		switch outcome.Item.Status {
		case domain.ImportItemStatusAccepted:
			accepted++
			providerSuccessAny = true
		case domain.ImportItemStatusWarning:
			warnings++
			providerSuccessAny = true
		case domain.ImportItemStatusRateLimited:
			rejected++
			rateLimitedChunk = true
		default:
			rejected++
		}
	}

	completedAt := s.now()
	finalStatus := finalChunkStatus(len(items), accepted, rejected, warnings, rateLimitedChunk, providerSuccessAny)
	var errorMessage *string
	if finalStatus == domain.ImportChunkStatusFailed || finalStatus == domain.ImportChunkStatusRateLimited {
		msg := "chunk had no accepted items"
		if rateLimitedChunk {
			msg = "provider rate limited"
		}
		errorMessage = &msg
	}
	_ = s.batchRepo.UpdateChunkStatus(ctx, chunk.ID, domain.ImportChunkStatusUpdate{
		Status:          finalStatus,
		AcceptedRecords: &accepted,
		RejectedRecords: &rejected,
		WarningRecords:  &warnings,
		CompletedAt:     &completedAt,
		ErrorMessage:    errorMessage,
	})

	summary := ImportChunkSummary{
		ChunkID:         chunk.ID,
		ChunkIndex:      chunk.ChunkIndex,
		Status:          finalStatus,
		TotalRecords:    len(items),
		AcceptedRecords: accepted,
		RejectedRecords: rejected,
		WarningRecords:  warnings,
	}
	if errorMessage != nil {
		summary.ErrorMessage = *errorMessage
	}
	return summary
}

type itemOutcome struct {
	Item domain.ImportChunkItem
}

func (s *Service) processItem(ctx context.Context, item domain.ImportChunkItem, batch *domain.ImportBatch, order []string, includeQuote, includeHistory bool, historyLimit int) itemOutcome {
	symbol, err := normalizeSymbol(item.Symbol)
	if err != nil {
		item.Status = domain.ImportItemStatusRejected
		item.ErrorCode = "INVALID_SYMBOL"
		item.ErrorMessage = err.Error()
		return itemOutcome{Item: item}
	}

	// Phase 4: try reference_data resolution before touching the provider.
	resolved, hadResolver := s.resolveSymbolForBatch(ctx, symbol, batch, order)
	if hadResolver {
		if resolved.SecurityID != "" {
			item.SecurityID = resolved.SecurityID
		}
		if resolved.ProviderSymbol != "" {
			item.ProviderSymbol = resolved.ProviderSymbol
		}
		if !resolved.HasMapping {
			// No active provider mapping → record candidate, do NOT fetch.
			s.recordUnmappedCandidate(ctx, batch, symbol, resolved.SecurityID)
			item.Status = domain.ImportItemStatusUnmapped
			item.ErrorCode = domain.ImportErrorCodeUnmapped
			item.ErrorMessage = "no active provider mapping for " + symbol
			return itemOutcome{Item: item}
		}
	}

	// Legacy mapping path — only consulted when reference_data didn't resolve.
	mapping := s.symbolMapping(ctx, symbol)
	if !hadResolver && mapping == nil && len(order) > 0 && order[0] != "default" {
		item.Status = domain.ImportItemStatusUnmapped
		item.ErrorCode = domain.ImportErrorCodeUnmapped
		item.ErrorMessage = "no provider symbol mapping for " + symbol
	}

	// Prefer the canonical provider_symbol the resolver returned when
	// available; otherwise fall back to legacy market_symbols mapping.
	if hadResolver && resolved.ProviderSymbol != "" {
		mapping = providerSymbolMappingForOrder(resolved.ProviderSymbol, order)
	}

	var quoteFailures []string
	var historyFailures []string
	quoteOk := false
	historyOk := false
	rateLimited := false

	if includeQuote {
		q, failures, err := s.fetchQuoteWithOrder(ctx, symbol, order, mapping)
		if err == nil && q != nil {
			quoteOk = true
			if mapping == nil {
				item.ProviderSymbol = q.Symbol
			}
		} else {
			quoteFailures = failures
			if containsRateLimited(failures) {
				rateLimited = true
			}
		}
	}

	if includeHistory {
		limit := historyLimit
		if limit <= 0 {
			limit = 250
		}
		bars, failures, err := s.fetchHistoryWithOrder(ctx, symbol, limit, order, mapping)
		if err == nil && len(bars) > 0 {
			historyOk = true
		} else {
			historyFailures = failures
			if containsRateLimited(failures) {
				rateLimited = true
			}
		}
	}

	if rateLimited && !quoteOk && !historyOk {
		item.Status = domain.ImportItemStatusRateLimited
		item.ErrorCode = domain.ImportErrorCodeRateLimited
		item.ErrorMessage = joinFailures(quoteFailures, historyFailures)
		return itemOutcome{Item: item}
	}

	needsQuote := includeQuote && !quoteOk
	needsHistory := includeHistory && !historyOk
	if !needsQuote && !needsHistory {
		if item.Status == domain.ImportItemStatusUnmapped {
			item.ErrorCode = ""
			item.ErrorMessage = ""
		}
		item.Status = domain.ImportItemStatusAccepted
		return itemOutcome{Item: item}
	}

	if (!needsQuote && quoteOk) || (!needsHistory && historyOk) {
		item.Status = domain.ImportItemStatusWarning
		item.ErrorCode = domain.ImportErrorCodeProviderFailure
		item.ErrorMessage = joinFailures(quoteFailures, historyFailures)
		return itemOutcome{Item: item}
	}

	if item.Status == domain.ImportItemStatusUnmapped {
		return itemOutcome{Item: item}
	}
	item.Status = domain.ImportItemStatusRejected
	item.ErrorCode = domain.ImportErrorCodeProviderFailure
	item.ErrorMessage = joinFailures(quoteFailures, historyFailures)
	return itemOutcome{Item: item}
}

// resolveSymbolForBatch attempts to resolve an input symbol against the
// reference_data resolver. Returns (resolved, false) when no resolver is
// wired so callers can fall back to legacy behaviour.
//
// Resolution order: IMS symbol → display symbol → provider symbol against
// each provider in `order`. Once a canonical security is found, the active
// provider mapping for the FIRST provider in `order` (if any) is consulted
// to pick the provider_symbol to send.
func (s *Service) resolveSymbolForBatch(ctx context.Context, symbol string, batch *domain.ImportBatch, order []string) (resolvedSymbol, bool) {
	if s == nil || s.securityResolver == nil {
		return resolvedSymbol{}, false
	}
	var security *refdomain.Security

	if sec, err := s.securityResolver.GetSecurityByIMSSymbol(ctx, symbol); err == nil && sec != nil {
		security = sec
	}
	if security == nil {
		if sec, err := s.securityResolver.GetSecurityByDisplaySymbol(ctx, symbol); err == nil && sec != nil {
			security = sec
		}
	}
	if security == nil {
		// Try each provider in the configured order with the input symbol
		// treated as a provider symbol.
		for _, providerCode := range order {
			if providerCode == "" || providerCode == "default" {
				continue
			}
			if sec, err := s.securityResolver.ResolveSecurityByProviderSymbol(ctx, providerCode, symbol); err == nil && sec != nil {
				security = sec
				break
			}
		}
	}

	out := resolvedSymbol{}
	if security == nil {
		return out, true
	}
	out.SecurityID = security.ID
	out.IMSSymbol = security.IMSSymbol
	out.DisplaySymbol = security.DisplaySymbol

	// Pick the provider symbol to use for the first usable provider in order.
	for _, providerCode := range order {
		if providerCode == "" || providerCode == "default" {
			continue
		}
		mapping, err := s.securityResolver.ResolveProviderSymbol(ctx, security.ID, providerCode)
		if err != nil || mapping == nil {
			continue
		}
		out.ProviderSymbol = mapping.ProviderSymbol
		out.HasMapping = true
		return out, true
	}
	// If "default" provider was requested, accept any active mapping the
	// security carries — adapters will iterate and the symbol on the security
	// is enough to identify the canonical row.
	if isDefaultOrder(order) {
		for _, m := range security.ProviderMappings {
			if m.MappingStatus == refdomain.MappingStatusActive {
				out.ProviderSymbol = m.ProviderSymbol
				out.HasMapping = true
				return out, true
			}
		}
	}
	return out, true
}

func (s *Service) recordUnmappedCandidate(ctx context.Context, batch *domain.ImportBatch, symbol, suggestedSecurityID string) {
	if s == nil || s.securityResolver == nil || batch == nil {
		return
	}
	providerCode := strings.ToLower(strings.TrimSpace(batch.ProviderCode))
	if providerCode == "" || providerCode == "default" {
		// No specific provider context — pick first non-default provider, if any.
		// (Avoids logging "default" as a code.)
		providerCode = "unknown"
	}
	batchID := batch.ID
	candidate := refdomain.UnmappedSecurityCandidate{
		BatchID:        &batchID,
		ProviderCode:   providerCode,
		ProviderSymbol: symbol,
	}
	if suggestedSecurityID != "" {
		id := suggestedSecurityID
		candidate.SuggestedSecurityID = &id
	}
	_, _ = s.securityResolver.CreateUnmappedCandidate(ctx, candidate)
}

func providerSymbolMappingForOrder(providerSymbol string, order []string) *domain.SymbolMapping {
	// Build a minimal SymbolMapping so the existing fetch helpers route the
	// resolved provider_symbol to whichever adapter is invoked first in
	// `order`. The legacy mappedProviderSymbol() helper inspects only the
	// known provider fields (alpha_vantage / yahoo today), so this stays
	// compatible without hardcoding additional providers in DB columns:
	// for unknown providers, fetchQuoteWithOrder falls back to the input
	// symbol passed by the caller.
	if len(order) == 0 {
		return &domain.SymbolMapping{}
	}
	m := &domain.SymbolMapping{}
	for _, providerCode := range order {
		switch strings.ToLower(strings.TrimSpace(providerCode)) {
		case domain.ProviderAlphaVantage:
			if m.AlphaVantageSymbol == "" {
				m.AlphaVantageSymbol = providerSymbol
			}
		case domain.ProviderYahoo:
			if m.YahooFinanceSymbol == "" {
				m.YahooFinanceSymbol = providerSymbol
			}
		}
	}
	return m
}

func isDefaultOrder(order []string) bool {
	if len(order) == 0 {
		return true
	}
	for _, p := range order {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" || p == "default" {
			return true
		}
	}
	return false
}

func chunkSymbols(symbols []string, size int) []domain.ImportChunkPlan {
	if size <= 0 {
		size = defaultBatchChunkSize
	}
	out := []domain.ImportChunkPlan{}
	for i := 0; i < len(symbols); i += size {
		end := i + size
		if end > len(symbols) {
			end = len(symbols)
		}
		out = append(out, domain.ImportChunkPlan{
			ChunkIndex: len(out),
			Symbols:    append([]string(nil), symbols[i:end]...),
		})
	}
	return out
}

func normalizeSymbols(in []string) ([]string, error) {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		symbol, err := normalizeSymbol(raw)
		if err != nil {
			return nil, err
		}
		if seen[symbol] {
			continue
		}
		seen[symbol] = true
		out = append(out, symbol)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: no valid symbols", ErrInvalidMarketDataRequest)
	}
	return out, nil
}

func importFlagsFor(importType string) (bool, bool, int) {
	switch strings.ToUpper(importType) {
	case domain.ImportTypeQuoteSync:
		return true, false, 0
	case domain.ImportTypeHistorySync:
		return false, true, 250
	case domain.ImportTypeQuoteAndHistorySync:
		return true, true, 250
	case domain.ImportTypeManualFile:
		return true, false, 0
	default:
		return true, false, 0
	}
}

func validImportType(value string) bool {
	switch value {
	case domain.ImportTypeQuoteSync,
		domain.ImportTypeHistorySync,
		domain.ImportTypeQuoteAndHistorySync,
		domain.ImportTypeManualFile:
		return true
	}
	return false
}

func finalChunkStatus(total, accepted, rejected, warnings int, rateLimited, anySuccess bool) string {
	if total == 0 {
		return domain.ImportChunkStatusFailed
	}
	if rateLimited && !anySuccess {
		return domain.ImportChunkStatusRateLimited
	}
	if accepted+warnings == 0 {
		return domain.ImportChunkStatusFailed
	}
	if rejected > 0 || warnings > 0 {
		return domain.ImportChunkStatusCompletedWithWarnings
	}
	return domain.ImportChunkStatusCompleted
}

func finalBatchStatus(totalChunks, succeeded, failed, rejected, warnings int) string {
	if totalChunks == 0 {
		return domain.ImportBatchStatusFailed
	}
	if failed == totalChunks {
		return domain.ImportBatchStatusFailed
	}
	if failed > 0 {
		return domain.ImportBatchStatusPartialFailed
	}
	if rejected > 0 || warnings > 0 {
		return domain.ImportBatchStatusCompletedWithWarnings
	}
	if succeeded == totalChunks {
		return domain.ImportBatchStatusCompleted
	}
	return domain.ImportBatchStatusCompletedWithWarnings
}

func containsRateLimited(failures []string) bool {
	for _, f := range failures {
		lower := strings.ToLower(f)
		if strings.Contains(lower, "rate") && strings.Contains(lower, "limit") {
			return true
		}
		if strings.Contains(lower, "429") {
			return true
		}
		if strings.Contains(lower, "rate_limited") {
			return true
		}
	}
	return false
}

func joinFailures(quote, history []string) string {
	parts := []string{}
	if len(quote) > 0 {
		parts = append(parts, "quote: "+strings.Join(quote, ", "))
	}
	if len(history) > 0 {
		parts = append(parts, "history: "+strings.Join(history, ", "))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}
