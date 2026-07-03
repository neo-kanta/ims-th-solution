package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

type CreateImportBatchRequest struct {
	Provider       string   `json:"provider"`
	ImportType     string   `json:"import_type"`
	Symbols        []string `json:"symbols"`
	ChunkSize      int      `json:"chunk_size"`
	IncludeQuote   *bool    `json:"include_quote"`
	IncludeHistory *bool    `json:"include_history"`
	HistoryLimit   int      `json:"history_limit"`
	IdempotencyKey string   `json:"idempotency_key"`
}

type CreateImportBatchResponse = application.CreateImportBatchResult
type RunImportBatchResponse = application.RunImportBatchResult
type ImportBatchStatusResponseDTO = application.ImportBatchStatusResponse

// CreateImportBatch handles POST /market-data/import-batches.
// @Summary Create Market Data Import Batch
// @Description Create a chunked market data import batch for the given symbols.
// @Tags MarketData
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateImportBatchRequest true "Import batch payload"
// @Success 201 {object} CreateImportBatchResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /market-data/import-batches [post]
func (h *Handler) CreateImportBatch(w http.ResponseWriter, r *http.Request) {
	var req CreateImportBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	includeQuote := true
	includeHistory := false
	if req.IncludeQuote != nil {
		includeQuote = *req.IncludeQuote
	}
	if req.IncludeHistory != nil {
		includeHistory = *req.IncludeHistory
	}
	result, err := h.service.CreateImportBatch(r.Context(), application.CreateImportBatchRequest{
		Provider:       req.Provider,
		ImportType:     req.ImportType,
		Symbols:        req.Symbols,
		ChunkSize:      req.ChunkSize,
		IncludeQuote:   includeQuote,
		IncludeHistory: includeHistory,
		HistoryLimit:   req.HistoryLimit,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		writeBatchError(w, err)
		return
	}
	httputil.Created(w, result)
}

// RunImportBatch handles POST /market-data/import-batches/{batch_id}/run.
// @Summary Run Market Data Import Batch
// @Description Runs all chunks for the batch synchronously.
// @Tags MarketData
// @Security BearerAuth
// @Produce json
// @Param batch_id path string true "Import batch id"
// @Success 200 {object} RunImportBatchResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 502 {object} httputil.ErrorResponse
// @Router /market-data/import-batches/{batch_id}/run [post]
func (h *Handler) RunImportBatch(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batch_id")
	if batchID == "" {
		httputil.BadRequest(w, "batch_id is required")
		return
	}
	result, err := h.service.RunImportBatch(r.Context(), batchID)
	if err != nil {
		writeBatchError(w, err)
		return
	}
	httputil.OK(w, result)
}

// GetImportBatch handles GET /market-data/import-batches/{batch_id}.
// @Summary Get Market Data Import Batch
// @Description Returns batch status with chunk summary and error rows.
// @Tags MarketData
// @Security BearerAuth
// @Produce json
// @Param batch_id path string true "Import batch id"
// @Success 200 {object} ImportBatchStatusResponseDTO
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /market-data/import-batches/{batch_id} [get]
func (h *Handler) GetImportBatch(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batch_id")
	if batchID == "" {
		httputil.BadRequest(w, "batch_id is required")
		return
	}
	resp, err := h.service.GetImportBatch(r.Context(), batchID)
	if err != nil {
		writeBatchError(w, err)
		return
	}
	httputil.OK(w, resp)
}

// GetImportBatchErrors handles GET /market-data/import-batches/{batch_id}/errors.
// @Summary List Market Data Import Batch Errors
// @Description Returns rejected/failed/warning items for the batch.
// @Tags MarketData
// @Security BearerAuth
// @Produce json
// @Param batch_id path string true "Import batch id"
// @Success 200 {array} domain.ImportChunkItem
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /market-data/import-batches/{batch_id}/errors [get]
func (h *Handler) GetImportBatchErrors(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batch_id")
	if batchID == "" {
		httputil.BadRequest(w, "batch_id is required")
		return
	}
	items, err := h.service.ListImportBatchErrors(r.Context(), batchID)
	if err != nil {
		writeBatchError(w, err)
		return
	}
	httputil.OK(w, items)
}

func writeBatchError(w http.ResponseWriter, err error) {
	if errors.Is(err, application.ErrImportBatchNotFound) {
		httputil.NotFound(w, err.Error())
		return
	}
	if errors.Is(err, application.ErrInvalidMarketDataRequest) {
		httputil.BadRequest(w, err.Error())
		return
	}
	if errors.Is(err, application.ErrMarketDataProviderNotAvailable) {
		httputil.BadRequest(w, err.Error())
		return
	}
	httputil.InternalError(w, err.Error())
}
