package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

type QuoteResponse = domain.Quote
type PriceBarResponse = domain.PriceBar
type ProviderHealthResponse = domain.ProviderHealth
type ImportMarketDataResponse = application.ImportMarketDataResult

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

// GetQuote handles GET /market-data/quote.
// @Summary Get Market Quote
// @Description Get a latest quote for a symbol from the configured provider chain or a requested provider.
// @Tags MarketData
// @Security BearerAuth
// @Produce json
// @Param symbol query string true "Market symbol"
// @Param provider query string false "Provider name (alpha_vantage or yahoo)"
// @Success 200 {object} QuoteResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 502 {object} httputil.ErrorResponse
// @Router /market-data/quote [get]
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

// GetHistory handles GET /market-data/history.
// @Summary Get Market Price History
// @Description Get daily price bars for a symbol from the configured provider chain or a requested provider.
// @Tags MarketData
// @Security BearerAuth
// @Produce json
// @Param symbol query string true "Market symbol"
// @Param provider query string false "Provider name (alpha_vantage or yahoo)"
// @Param limit query int false "Maximum bars to return (default 250, max 1000)"
// @Success 200 {array} PriceBarResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 502 {object} httputil.ErrorResponse
// @Router /market-data/history [get]
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

type ImportMarketDataRequest struct {
	Symbol         string `json:"symbol"`
	Provider       string `json:"provider"`
	IncludeQuote   *bool  `json:"include_quote"`
	IncludeHistory *bool  `json:"include_history"`
	HistoryLimit   int    `json:"history_limit"`
}

// ImportMarketData handles POST /market-data/import.
// @Summary Import Market Data
// @Description Fetch and persist quote and/or daily price history for a symbol.
// @Tags MarketData
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ImportMarketDataRequest true "Market data import payload"
// @Success 201 {object} ImportMarketDataResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 502 {object} httputil.ErrorResponse
// @Router /market-data/import [post]
func (h *Handler) ImportMarketData(w http.ResponseWriter, r *http.Request) {
	var req ImportMarketDataRequest
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

// ProviderHealth handles GET /market-data/provider-health.
// @Summary Get Market Data Provider Health
// @Description Get configured market data provider health and recent status information.
// @Tags MarketData
// @Security BearerAuth
// @Produce json
// @Success 200 {array} ProviderHealthResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /market-data/provider-health [get]
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
