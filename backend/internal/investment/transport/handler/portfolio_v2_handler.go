package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// Portfolio V2 read endpoints (docs/api/portfolio-v2-api-ddd.md, Milestone 2
// of docs/handoff/portfolio-v2-claude-implementation-prompt.md).
//
// Every handler in this file resolves {portfolioCode} to the internal
// portfolio_id exactly once, at the transport boundary, then delegates to
// the same domain repositories the V1 UUID-keyed routes already use.
// Application services and repositories never see portfolioCode — they
// keep working in portfolio_id, matching the V1 handlers in
// investment_handler.go.

// resolvePortfolioCode resolves the {portfolioCode} path param to a
// portfolio, enforcing the same fund-scoped data permission V1 routes use.
// Writes the appropriate error response and returns (nil, false) on any
// failure — callers should return immediately when ok is false.
func (h *InvestmentHandler) resolvePortfolioCode(w http.ResponseWriter, r *http.Request) (*entity.Portfolio, bool) {
	code := chi.URLParam(r, "portfolioCode")
	if code == "" {
		httputil.BadRequest(w, "invalid portfolio code")
		return nil, false
	}
	p, err := h.portfolios.GetByCode(r.Context(), code)
	if err != nil {
		// Includes *domain.ErrAmbiguousPortfolioCode (409).
		writeDomainError(w, err)
		return nil, false
	}
	if p == nil {
		httputil.NotFound(w, "portfolio not found")
		return nil, false
	}
	if !hasFundAccess(r.Context(), h.pc, p.FundID) {
		httputil.Forbidden(w, "no access to this portfolio")
		return nil, false
	}
	return p, true
}

// GetHoldingsByCode handles GET /api/v2/portfolios/{portfolioCode}/holdings.
// @Summary Get Portfolio Holdings By Code
// @Description List current holdings for a portfolio, resolved by business code (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Success 200 {array} response.HoldingResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/holdings [get]
func (h *InvestmentHandler) GetHoldingsByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	rows, err := h.positions.ListByPortfolio(r.Context(), p.ID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.HoldingResponse, 0, len(rows))
	for _, pos := range rows {
		out = append(out, response.FromPosition(pos))
	}
	httputil.OK(w, out)
}

// GetCashByCode handles GET /api/v2/portfolios/{portfolioCode}/cash.
// @Summary Get Portfolio Cash Balances By Code
// @Description List cash balances by currency for a portfolio, resolved by business code (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Success 200 {array} response.CashBalanceResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/cash [get]
func (h *InvestmentHandler) GetCashByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	bals, err := h.cash.ListBalances(r.Context(), p.ID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.CashBalanceResponse, 0, len(bals))
	for _, b := range bals {
		out = append(out, response.FromCashBalance(b))
	}
	httputil.OK(w, out)
}

// ListTransactionsByCode handles GET /api/v2/portfolios/{portfolioCode}/transactions.
// @Summary List Portfolio Transactions By Code
// @Description List transactions for a portfolio with optional instrument and date filters, resolved by business code (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param instrument_id query string false "Instrument UUID"
// @Param from query string false "Start business date (YYYY-MM-DD)"
// @Param to query string false "End business date (YYYY-MM-DD)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.TransactionListResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/transactions [get]
func (h *InvestmentHandler) ListTransactionsByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	page, limit := paginationParams(r)
	filter := domain.TransactionListFilter{
		PortfolioID: &p.ID,
		Page:        page,
		Limit:       limit,
	}
	if v := r.URL.Query().Get("instrument_id"); v != "" {
		if iid, err := parseUUID(v); err == nil {
			filter.InstrumentID = &iid
		}
	}
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := parseDate(v); err == nil {
			filter.BusinessFrom = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := parseDate(v); err == nil {
			filter.BusinessTo = &t
		}
	}
	rows, total, err := h.txns.List(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.TransactionResponse, 0, len(rows))
	for _, t := range rows {
		out = append(out, response.FromTransaction(t))
	}
	httputil.OK(w, response.PaginatedResponse[response.TransactionResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

// ListValuationsByCode handles GET /api/v2/portfolios/{portfolioCode}/valuations.
// @Summary List Portfolio Valuations By Code
// @Description List valuation snapshots for a portfolio, resolved by business code (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.ValuationListResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/valuations [get]
func (h *InvestmentHandler) ListValuationsByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	page, limit := paginationParams(r)
	from, to := dateRangeParams(r)

	rows, total, err := h.valuation.List(r.Context(), p.ID, from, to, page, limit)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.ValuationResponse, 0, len(rows))
	for _, v := range rows {
		out = append(out, response.FromValuation(v))
	}
	httputil.OK(w, response.PaginatedResponse[response.ValuationResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

// GetLatestValuationByCode handles GET /api/v2/portfolios/{portfolioCode}/valuations/latest.
// @Summary Get Latest Portfolio Valuation By Code
// @Description Retrieve the latest internal valuation snapshot for a portfolio, resolved by business code (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Success 200 {object} response.ValuationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/valuations/latest [get]
func (h *InvestmentHandler) GetLatestValuationByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	v, err := h.valuation.GetLatest(r.Context(), p.ID, vo.ValuationSourceInternal)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if v == nil {
		httputil.NotFound(w, "no valuation snapshots available")
		return
	}
	httputil.OK(w, response.FromValuation(v))
}
