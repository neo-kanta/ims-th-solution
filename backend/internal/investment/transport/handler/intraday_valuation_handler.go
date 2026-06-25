package handler

import (
	"net/http"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// IntradayValuationHandler exposes the intraday holdings valuation, manual
// refresh, and feed-status endpoints.
//
// Route gating is configured by the module: read endpoints require
// INVESTMENT_VALUATION_VIEW; the refresh requires INVESTMENT_VALUATION_RUN.
//
// All endpoints degrade gracefully on provider outage: positions are
// returned with IsStale=true rather than returning a 5xx.
type IntradayValuationHandler struct {
	svc   *service.IntradayValuationService
	audit contract.AuditLogger
}

// NewIntradayValuationHandler wires the handler.
func NewIntradayValuationHandler(svc *service.IntradayValuationService, audit contract.AuditLogger) *IntradayValuationHandler {
	return &IntradayValuationHandler{svc: svc, audit: audit}
}

// SetService swaps in the intraday valuation service after construction. Used
// by the module to wire the service once the market_data quote provider port
// is available — avoids a circular construction dependency between investment
// and market_data.
func (h *IntradayValuationHandler) SetService(svc *service.IntradayValuationService) {
	if h == nil {
		return
	}
	h.svc = svc
}

// GetFundHoldingsValuation handles GET /investment/funds/{id}/holdings/valuation.
// @Summary Get Intraday Holdings Valuation
// @Description Compute estimated NAV / AUM and per-position unrealised P&L from live market data, alongside the official accounting NAV.
// @Tags Investment - Intraday
// @Security BearerAuth
// @Produce json
// @Param id path string true "Fund UUID"
// @Success 200 {object} response.IntradayValuationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds/{id}/holdings/valuation [get]
func (h *IntradayValuationHandler) GetFundHoldingsValuation(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.svc == nil {
		httputil.InternalError(w, "intraday valuation service not wired")
		return
	}
	fundID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	result, err := h.svc.ComputeFundValuation(r.Context(), fundID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.OK(w, response.FromIntradayValuation(result))
}

// RefreshFundMarketData handles POST /investment/funds/{id}/market-data/refresh.
// @Summary Refresh Fund Market Data
// @Description Trigger a provider fetch for every instrument held by the fund. Records an audit event.
// @Tags Investment - Intraday
// @Security BearerAuth
// @Produce json
// @Param id path string true "Fund UUID"
// @Success 200 {object} response.MarketDataRefreshResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds/{id}/market-data/refresh [post]
func (h *IntradayValuationHandler) RefreshFundMarketData(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.svc == nil {
		httputil.InternalError(w, "intraday valuation service not wired")
		return
	}
	fundID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	result, err := h.svc.RefreshFundQuotes(r.Context(), fundID)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	if h.audit != nil {
		_ = h.audit.LogAction(contract.AuditEntry{
			ActorID:      actor.String(),
			Action:       "INVESTMENT_MARKET_DATA_REFRESHED",
			Module:       "investment",
			ResourceType: "INVESTMENT_FUND",
			ResourceID:   fundID.String(),
			Details: map[string]any{
				"requested_symbols": result.RequestedSymbols,
				"success_symbols":   result.SuccessSymbols,
				"stale_symbols":     result.StaleSymbols,
				"failed_symbols":    result.FailedSymbols,
				"used_provider":     result.UsedProvider,
			},
		})
	}

	httputil.OK(w, response.FromMarketDataRefresh(fundID, result))
}

// GetFundMarketDataStatus handles GET /investment/funds/{id}/market-data/status.
// @Summary Get Fund Market Data Status
// @Description Report provider health and the number of stale positions for a fund.
// @Tags Investment - Intraday
// @Security BearerAuth
// @Produce json
// @Param id path string true "Fund UUID"
// @Success 200 {object} response.MarketDataStatusResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds/{id}/market-data/status [get]
func (h *IntradayValuationHandler) GetFundMarketDataStatus(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.svc == nil {
		httputil.InternalError(w, "intraday valuation service not wired")
		return
	}
	fundID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	status, err := h.svc.GetFeedStatus(r.Context(), fundID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.OK(w, response.FromMarketDataStatus(status))
}
