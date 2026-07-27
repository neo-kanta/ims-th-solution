package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// iso4217Shape matches a 3-letter uppercase currency code. This mirrors the
// DB-level CHECK constraints (chk_inv_portfolios_base_currency /
// chk_inv_portfolios_valuation_currency) so a malformed currency is rejected
// with a clean 400 at the transport boundary instead of surfacing as a raw
// Postgres constraint-violation 500.
var iso4217Shape = regexp.MustCompile(`^[A-Z]{3}$`)

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
	return resolvePortfolioByCode(w, r, h.portfolios, h.pc)
}

// resolvePortfolioByCode is the shared {portfolioCode} -> portfolio resolver
// for every Portfolio V2 handler (portfolio_v2_handler.go,
// portfolio_v2_ledger_handler.go, portfolio_v2_decision_handler.go,
// portfolio_v2_execution_handler.go). Writes the appropriate error response
// and returns (nil, false) on any failure — callers should return
// immediately when ok is false.
//
// pc is expected to be non-nil in production: every Portfolio V2 handler
// (InvestmentHandler, DecisionHandler, ExecutionHandler,
// TradeConfirmationHandler) is wired in module.go with a real
// contract.PermissionChecker (module.go's m.permissionAdapter), which itself
// fails closed — returning "no access" — if the underlying IAM port was
// never configured. pc == nil is tolerated here only so unit tests that
// construct a handler struct literal without setting pc (and are not
// exercising fund-scope behavior) keep compiling and passing; it must never
// happen via the production constructor/setter path. Do not pass a literal
// nil for pc from a handler method — always pass h.pc.
func resolvePortfolioByCode(
	w http.ResponseWriter, r *http.Request,
	portfolios domain.PortfolioRepository,
	pc contract.PermissionChecker,
) (*entity.Portfolio, bool) {
	code := chi.URLParam(r, "portfolioCode")
	if code == "" {
		httputil.BadRequest(w, "invalid portfolio code")
		return nil, false
	}
	p, err := portfolios.GetByCode(r.Context(), code)
	if err != nil {
		// Includes *domain.ErrAmbiguousPortfolioCode (409).
		writeDomainError(w, err)
		return nil, false
	}
	if p == nil {
		httputil.NotFound(w, "portfolio not found")
		return nil, false
	}
	// hasFundAccess itself fails closed when pc is nil (denies rather than
	// skipping the check) — do not special-case pc==nil here, or a handler
	// that's accidentally left unwired in production would silently allow
	// cross-fund access instead of denying it. portfolioScopeID falls back to
	// the portfolio's own id for a fund-less portfolio.
	if !hasFundAccess(r.Context(), pc, portfolioScopeID(p)) {
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

// ListPortfoliosV2 handles GET /api/v2/portfolios.
// @Summary List Portfolios (V2)
// @Description List portfolios visible to the authenticated user, scoped to their accessible funds (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param fund_code query string false "Fund code"
// @Param status query string false "Portfolio status"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.PortfolioListResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios [get]
func (h *InvestmentHandler) ListPortfoliosV2(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationParams(r)
	filter := domain.PortfolioListFilter{Page: page, Limit: limit}

	if v := strings.TrimSpace(r.URL.Query().Get("fund_code")); v != "" {
		fund, err := h.funds.GetByCode(r.Context(), v)
		if err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		if fund == nil {
			httputil.NotFound(w, "fund not found")
			return
		}
		filter.FundID = &fund.ID
	}
	if v := r.URL.Query().Get("status"); v != "" {
		s := vo.PortfolioStatus(v)
		filter.Status = &s
	}

	// Data permission: scope to the funds the user can access, matching
	// InvestmentHandler.ListPortfolios's existing accessibleFundIDs pattern.
	filter.AccessibleFundIDs = accessibleFundIDs(r.Context(), h.pc)

	ps, total, err := h.portfolios.List(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.PortfolioResponse, 0, len(ps))
	for _, p := range ps {
		out = append(out, response.FromPortfolio(p))
	}
	httputil.OK(w, response.PaginatedResponse[response.PortfolioResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

// CreatePortfolioV2 handles POST /api/v2/portfolios.
// @Summary Create Portfolio (V2)
// @Description Create a portfolio under a fund, resolved by business fund_code (Portfolio V2). Request body must not include fund_id.
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.CreatePortfolioV2Request true "Portfolio create payload"
// @Success 201 {object} response.PortfolioResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios [post]
func (h *InvestmentHandler) CreatePortfolioV2(w http.ResponseWriter, r *http.Request) {
	var req request.CreatePortfolioV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	fundCode := strings.TrimSpace(req.FundCode)
	code := strings.TrimSpace(req.Code)
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		httputil.BadRequest(w, "code and name are required")
		return
	}
	baseCcy := strings.ToUpper(strings.TrimSpace(req.BaseCurrency))
	valCcy := strings.ToUpper(strings.TrimSpace(req.ValuationCurrency))
	if !iso4217Shape.MatchString(baseCcy) {
		httputil.BadRequest(w, "invalid base_currency (expected a 3-letter ISO 4217 code)")
		return
	}
	if !iso4217Shape.MatchString(valCcy) {
		httputil.BadRequest(w, "invalid valuation_currency (expected a 3-letter ISO 4217 code)")
		return
	}
	portfolioType := vo.PortfolioType(strings.ToUpper(strings.TrimSpace(req.PortfolioType)))
	if !portfolioType.IsValid() {
		httputil.BadRequest(w, "invalid portfolio_type (expected LIVE, SIMULATION, or MODEL)")
		return
	}
	// risk_profile is optional (empty means "not set"), but when supplied
	// must be one of the values chk_inv_portfolios_risk_profile enforces at
	// the DB layer — validated here so a bad value returns a clean 400
	// instead of a raw Postgres constraint-violation error.
	riskProfile := vo.RiskProfile(strings.ToUpper(strings.TrimSpace(req.RiskProfile)))
	if riskProfile != "" && !riskProfile.IsValid() {
		httputil.BadRequest(w, "invalid risk_profile (expected LOW, MEDIUM, HIGH, or SPECULATIVE)")
		return
	}
	inception, err := parseDate(req.InceptionDate)
	if err != nil {
		httputil.BadRequest(w, "invalid inception_date")
		return
	}

	// Resolve fund_code -> fund_id server-side. Never trust a client-supplied
	// fund_id per docs/MANAGER/MEMORY.md's V2 request-body rule. fund_code is
	// optional — an empty value creates a fund-less portfolio ("Bind with
	// Fund: N"); access to it is governed by a portfolio_id-scoped
	// permission_data_rights grant instead of a fund-scoped one.
	var fundID *uuid.UUID
	if fundCode != "" {
		fund, err := h.funds.GetByCode(r.Context(), fundCode)
		if err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		if fund == nil {
			httputil.NotFound(w, "fund not found")
			return
		}
		// Data permission: verify the caller can access the resolved fund
		// before creating anything under it.
		if !hasFundAccess(r.Context(), h.pc, fund.ID) {
			httputil.Forbidden(w, "no access to target fund")
			return
		}
		fundID = &fund.ID
	}

	taxMethod := vo.TaxLotMethod(req.TaxLotMethod)
	if taxMethod == "" {
		taxMethod = vo.TaxLotMethodAverage
	}

	p, err := h.portfolioCmd.Create(r.Context(), command.CreatePortfolioRequest{
		FundID:            fundID,
		PortfolioType:     portfolioType,
		Code:              code,
		Name:              name,
		Description:       strings.TrimSpace(req.Description),
		BaseCurrency:      baseCcy,
		ValuationCurrency: valCcy,
		StrategyCode:      req.StrategyCode,
		StyleID:           req.StyleID,
		ManagerUserID:     req.ManagerUserID,
		Benchmark:         req.Benchmark,
		RiskProfile:       riskProfile,
		InceptionDate:     inception,
		HasUnits:          req.HasUnits,
		TaxLotMethod:      taxMethod,
		ActorID:           actor,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.Created(w, response.FromPortfolio(p))
}

// PatchPortfolioByCode handles PATCH /api/v2/portfolios/{portfolioCode}.
// @Summary Update Portfolio By Code
// @Description Update mutable descriptive metadata on a portfolio, resolved by business code (Portfolio V2), using optimistic version control. Never changes fund association, code, or lifecycle status.
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param request body request.PatchPortfolioV2Request true "Portfolio patch payload"
// @Success 200 {object} response.PortfolioResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode} [patch]
func (h *InvestmentHandler) PatchPortfolioByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	var req request.PatchPortfolioV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	if req.ExpectedVersion < 1 {
		httputil.BadRequest(w, "expected_version is required")
		return
	}

	cmdReq := command.UpdatePortfolioRequest{
		PortfolioID:     p.ID,
		ExpectedVersion: req.ExpectedVersion,
		Name:            req.Name,
		Description:     req.Description,
		StrategyCode:    req.StrategyCode,
		StyleID:         req.StyleID,
		ManagerUserID:   req.ManagerUserID,
		Benchmark:       req.Benchmark,
		// Status is intentionally never populated — this endpoint never
		// mutates portfolio lifecycle/status. Fund association and code are
		// not part of request.PatchPortfolioV2Request at all.
		ActorID: actor,
	}
	if req.RiskProfile != nil {
		rp := vo.RiskProfile(strings.ToUpper(strings.TrimSpace(*req.RiskProfile)))
		if rp != "" && !rp.IsValid() {
			httputil.BadRequest(w, "invalid risk_profile (expected LOW, MEDIUM, HIGH, or SPECULATIVE)")
			return
		}
		cmdReq.RiskProfile = &rp
	}
	updated, err := h.portfolioCmd.Update(r.Context(), cmdReq)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.OK(w, response.FromPortfolio(updated))
}

// RunValuationByCode handles POST /api/v2/portfolios/{portfolioCode}/valuations/run.
// @Summary Run Portfolio Valuation By Code
// @Description Manually triggers the valuation runner for a portfolio, resolved by business code (Portfolio V2). Rejected for MODEL portfolios, which have no official valuation.
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param request body request.RunValuationRequest true "Valuation run payload"
// @Success 201 {object} response.ValuationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/valuations/run [post]
func (h *InvestmentHandler) RunValuationByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	// A MODEL portfolio is a target-allocation template with no real ledger
	// or execution (docs/MANAGER/MEMORY.md's portfolio-type rules) — an
	// official valuation is not a meaningful concept for it.
	if p.PortfolioType == vo.PortfolioTypeModel {
		httputil.UnprocessableEntity(w, "MODEL portfolios do not support official valuation runs")
		return
	}
	var req request.RunValuationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	bizDate, err := parseDate(req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, "invalid business_date")
		return
	}
	rates := map[string]decimal.Decimal{}
	for k, v := range req.FxRates {
		d, err := decimal.NewFromString(v)
		if err != nil {
			httputil.BadRequest(w, "invalid fx_rate for "+k)
			return
		}
		rates[k] = d
	}
	var totalUnits *decimal.Decimal
	if req.TotalUnits != "" {
		d, err := decimal.NewFromString(req.TotalUnits)
		if err != nil {
			httputil.BadRequest(w, "invalid total_units")
			return
		}
		totalUnits = &d
	}

	res, err := h.valuationRun.Run(r.Context(), service.RunRequest{
		PortfolioID:  p.ID,
		BusinessDate: bizDate,
		FxRates:      rates,
		TotalUnits:   totalUnits,
		ActorID:      actor,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.Created(w, response.FromValuation(res.Valuation))
}
