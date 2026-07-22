package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// Portfolio V2 execution/confirmation read endpoints for a resolved
// portfolio. Unlike the write endpoints in portfolio_v2_execution_handler.go
// (which reuse loadOwnedExecution for a {executionId} nested under an
// already-known decision/execution), these are the top-level
// GET .../executions and GET .../confirmations list+detail routes: every
// request resolves {portfolioCode} first (enforcing fund-scoped data
// permission via resolvePortfolioByCode), then either lists rows already
// scoped to that single portfolio_id, or loads a single row by ID and
// verifies it belongs to the resolved portfolio before returning it —
// never trusting a raw ID on the detail route in isolation.

// ListExecutionsByCode handles GET /api/v2/portfolios/{portfolioCode}/executions.
// @Summary List Portfolio Executions By Code
// @Description List trade executions for a portfolio, resolved by business code (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param status query string false "Execution status filter"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.ExecutionListResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/executions [get]
func (h *ExecutionHandler) ListExecutionsByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	page, limit := paginationParams(r)
	filter := domain.ExecutionListFilter{Page: page, Limit: limit}
	if v := r.URL.Query().Get("status"); v != "" {
		s := vo.ExecutionStatus(v)
		filter.Status = &s
	}
	items, total, err := h.executions.ListByPortfolio(r.Context(), p.ID, filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.ExecutionResponse, 0, len(items))
	for _, e := range items {
		out = append(out, response.FromExecution(e))
	}
	httputil.OK(w, response.PaginatedResponse[response.ExecutionResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

// GetExecutionByCode handles
// GET /api/v2/portfolios/{portfolioCode}/executions/{executionId}.
// @Summary Get Portfolio Execution By Code
// @Description Retrieve one execution that belongs to the resolved portfolio (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param executionId path string true "Execution UUID"
// @Success 200 {object} response.ExecutionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/executions/{executionId} [get]
func (h *ExecutionHandler) GetExecutionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	e, ok := h.loadOwnedExecution(w, r, p.ID)
	if !ok {
		return
	}
	httputil.OK(w, response.FromExecution(e))
}

// ListConfirmationsByCode handles
// GET /api/v2/portfolios/{portfolioCode}/confirmations.
// @Summary List Portfolio Trade Confirmations By Code
// @Description List trade confirmations for a portfolio, resolved by business code (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param status query string false "Confirmation status filter"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.TradeConfirmationListResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/confirmations [get]
func (h *TradeConfirmationHandler) ListConfirmationsByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	page, limit := paginationParams(r)
	filter := domain.TradeConfirmationListFilter{Page: page, Limit: limit}
	if v := r.URL.Query().Get("status"); v != "" {
		s := vo.TradeConfirmationStatus(v)
		filter.Status = &s
	}
	items, total, err := h.confirmations.ListByPortfolio(r.Context(), p.ID, filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.TradeConfirmationResponse, 0, len(items))
	for _, c := range items {
		out = append(out, response.FromTradeConfirmation(c))
	}
	httputil.OK(w, response.PaginatedResponse[response.TradeConfirmationResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

// GetConfirmationByCode handles
// GET /api/v2/portfolios/{portfolioCode}/confirmations/{confirmationId}.
// @Summary Get Portfolio Trade Confirmation By Code
// @Description Retrieve one trade confirmation that belongs to the resolved portfolio (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param confirmationId path string true "Confirmation UUID"
// @Success 200 {object} response.TradeConfirmationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/confirmations/{confirmationId} [get]
func (h *TradeConfirmationHandler) GetConfirmationByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	confirmationID, err := parseUUID(chi.URLParam(r, "confirmationId"))
	if err != nil {
		httputil.BadRequest(w, "invalid confirmation id")
		return
	}
	c, err := h.confirmations.GetByID(r.Context(), confirmationID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if c == nil || c.PortfolioID != p.ID {
		httputil.NotFound(w, "trade confirmation not found")
		return
	}
	httputil.OK(w, response.FromTradeConfirmation(c))
}
