package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetQuote(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		httputil.BadRequest(w, "symbol query parameter is required")
		return
	}
	provider := r.URL.Query().Get("provider")
	var (
		q   any
		err error
	)
	if provider != "" {
		q, err = h.service.GetQuoteFromProvider(r.Context(), symbol, provider)
	} else {
		q, err = h.service.GetQuote(r.Context(), symbol)
	}
	if err != nil {
		writeProviderUnavailable(w, err)
		return
	}
	httputil.OK(w, q)
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		httputil.BadRequest(w, "symbol query parameter is required")
		return
	}
	limit := 250
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 1000 {
		limit = 1000
	}
	provider := r.URL.Query().Get("provider")
	var (
		bars any
		err  error
	)
	if provider != "" {
		bars, err = h.service.GetDailyPricesFromProvider(r.Context(), symbol, limit, provider)
	} else {
		bars, err = h.service.GetDailyPrices(r.Context(), symbol, limit)
	}
	if err != nil {
		writeProviderUnavailable(w, err)
		return
	}
	httputil.OK(w, bars)
}

type importMarketDataRequest struct {
	Symbol         string `json:"symbol"`
	Provider       string `json:"provider"`
	IncludeQuote   *bool  `json:"include_quote"`
	IncludeHistory *bool  `json:"include_history"`
	HistoryLimit   int    `json:"history_limit"`
}

func (h *Handler) ImportMarketData(w http.ResponseWriter, r *http.Request) {
	var req importMarketDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	includeQuote := true
	includeHistory := true
	if req.IncludeQuote != nil {
		includeQuote = *req.IncludeQuote
	}
	if req.IncludeHistory != nil {
		includeHistory = *req.IncludeHistory
	}
	result, err := h.service.ImportMarketData(r.Context(), application.ImportMarketDataRequest{
		Symbol:         req.Symbol,
		Provider:       req.Provider,
		IncludeQuote:   includeQuote,
		IncludeHistory: includeHistory,
		HistoryLimit:   req.HistoryLimit,
	})
	if err != nil {
		writeProviderUnavailable(w, err)
		return
	}
	httputil.Created(w, result)
}

func (h *Handler) ProviderHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.service.ProviderHealth(r.Context())
	if err != nil {
		httputil.InternalError(w, "failed to load market data provider health")
		return
	}
	httputil.OK(w, health)
}

func writeProviderUnavailable(w http.ResponseWriter, err error) {
	if errors.Is(err, application.ErrInvalidMarketDataRequest) {
		httputil.BadRequest(w, err.Error())
		return
	}
	if errors.Is(err, application.ErrMarketDataProviderNotAvailable) {
		httputil.BadRequest(w, err.Error())
		return
	}
	httputil.JSON(w, http.StatusBadGateway, httputil.ErrorResponse{
		Error: "market data provider unavailable",
		Code:  "MARKET_DATA_PROVIDER_UNAVAILABLE",
	})
}
