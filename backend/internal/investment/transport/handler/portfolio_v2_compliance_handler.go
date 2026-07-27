package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// Portfolio Compliance V2 — {portfolioCode} compliance endpoints
// (docs/compliance-module.md §"Portfolio Compliance V2").
//
// Every handler here resolves {portfolioCode} to a portfolio exactly once,
// then delegates to contract.PortfolioComplianceContract (implemented by the
// compliance module — see internal/compliance/transport/portfolio_contract_adapter.go).
// Clients never send portfolio_id, fund_id, or contract_id: fund_id is read
// from the resolved portfolio and passed through only when non-nil, so a
// portfolio with no fund still gets a full pre/post-trade check at
// GLOBAL + PORTFOLIO scope.
//
// NOTE: investment__portfolios.fund_id is currently NOT NULL in the schema —
// every portfolio has a fund today. This handler is written to the eventual
// fund_id-optional contract described in the compliance V2 business rules
// so no further change is needed once that column is relaxed; until then the
// "no fund_id" branch is unreachable via production data but is exercised by
// unit tests against fakes.

// portfolioFundID returns p.FundID, or uuid.Nil if the portfolio has no fund
// (either p itself is nil, or it's a fund-less portfolio created with
// "Bind with Fund: N").
func portfolioFundID(p *entity.Portfolio) uuid.UUID {
	if p == nil || p.FundID == nil {
		return uuid.Nil
	}
	return *p.FundID
}

// writePortfolioComplianceError classifies pkg/contract sentinel errors
// returned by PortfolioComplianceContract into HTTP responses. Falls back to
// 500 for anything unrecognised (infrastructure failures) rather than
// leaking internal details.
func writePortfolioComplianceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, contract.ErrPortfolioRuleNotFound):
		httputil.UnprocessableEntity(w, err.Error())
	case errors.Is(err, contract.ErrPortfolioRuleInactive):
		httputil.Conflict(w, err.Error())
	case errors.Is(err, contract.ErrPortfolioBindingDuplicate):
		httputil.Conflict(w, err.Error())
	case errors.Is(err, contract.ErrPortfolioBindingInvalid):
		httputil.BadRequest(w, err.Error())
	case errors.Is(err, contract.ErrPortfolioBindingNotFound):
		httputil.NotFound(w, err.Error())
	default:
		httputil.InternalError(w, "compliance operation failed")
	}
}

// ============================================================
// POST /api/v2/portfolios/{portfolioCode}/compliance/checks/pre-trade
// ============================================================

// PortfolioPreTradeRequest is the V2 pre-trade check payload. No
// portfolio_id, fund_id, or contract_id field — those are resolved from the
// path and the portfolio record.
type PortfolioPreTradeRequest struct {
	BusinessDate string `json:"business_date"` // "2006-01-02"
	OrderID      string `json:"order_id,omitempty"`
	Ticker       string `json:"ticker"`
	Side         string `json:"side"` // BUY | SELL
	Quantity     string `json:"quantity"`
	Price        string `json:"price"`
	Fees         string `json:"fees,omitempty"`
	Currency     string `json:"currency"`
	Exchange     string `json:"exchange"`
}

// RunPreTradeCheckByCode handles POST /api/v2/portfolios/{portfolioCode}/compliance/checks/pre-trade.
// @Summary Run Portfolio Pre-Trade Compliance Check (V2)
// @Description Evaluate compliance rules for a proposed order against a portfolio resolved by business code. fund_id is used only if the portfolio has one.
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param request body PortfolioPreTradeRequest true "Pre-trade check payload"
// @Success 201 {object} contract.ProposedOrderResult
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/compliance/checks/pre-trade [post]
func (h *InvestmentHandler) RunPreTradeCheckByCode(w http.ResponseWriter, r *http.Request) {
	if h.compliance == nil {
		httputil.InternalError(w, "compliance service not available")
		return
	}
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}

	var req PortfolioPreTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body: "+err.Error())
		return
	}
	bizDate, err := parseDate(req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, "invalid business_date format, expected YYYY-MM-DD")
		return
	}
	qty, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		httputil.BadRequest(w, "invalid quantity")
		return
	}
	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		httputil.BadRequest(w, "invalid price")
		return
	}
	fees := decimal.Zero
	if req.Fees != "" {
		fees, err = decimal.NewFromString(req.Fees)
		if err != nil {
			httputil.BadRequest(w, "invalid fees")
			return
		}
	}
	orderID := uuid.New()
	if req.OrderID != "" {
		parsed, err := uuid.Parse(req.OrderID)
		if err != nil {
			httputil.BadRequest(w, "invalid order_id")
			return
		}
		orderID = parsed
	}

	actor, _ := actorID(r)
	result, err := h.compliance.CheckProposedOrder(r.Context(), contract.ProposedOrderCheck{
		PortfolioID:  p.ID,
		ContractID:   portfolioFundID(p),
		BusinessDate: bizDate,
		Actor:        actor.String(),
		OrderID:      orderID,
		Ticker:       req.Ticker,
		Side:         contract.ComplianceOrderSide(req.Side),
		Quantity:     qty,
		Price:        price,
		Fees:         fees,
		Currency:     req.Currency,
		Exchange:     req.Exchange,
	})
	if err != nil {
		if errors.Is(err, contract.ErrInvalidProposedOrder) {
			httputil.BadRequest(w, "pre-trade check failed: "+err.Error())
			return
		}
		httputil.InternalError(w, "pre-trade check failed: "+err.Error())
		return
	}
	// A BLOCK verdict is communicated via the response body, not HTTP 4xx —
	// same design as the V1 endpoint (compliance_handler.go).
	httputil.Created(w, result)
}

// ============================================================
// POST /api/v2/portfolios/{portfolioCode}/compliance/checks/post-trade
// ============================================================

// PortfolioPostTradeRequest is the V2 post-trade check payload.
type PortfolioPostTradeRequest struct {
	BusinessDate string `json:"business_date"`
}

// RunPostTradeCheckByCode handles POST /api/v2/portfolios/{portfolioCode}/compliance/checks/post-trade.
// @Summary Run Portfolio Post-Trade Compliance Check (V2)
// @Description Evaluate compliance rules for a portfolio resolved by business code. fund_id is used only if the portfolio has one.
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param request body PortfolioPostTradeRequest true "Post-trade check payload"
// @Success 201 {object} contract.ProposedOrderResult
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/compliance/checks/post-trade [post]
func (h *InvestmentHandler) RunPostTradeCheckByCode(w http.ResponseWriter, r *http.Request) {
	if h.compliance == nil {
		httputil.InternalError(w, "compliance service not available")
		return
	}
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}

	var req PortfolioPostTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body: "+err.Error())
		return
	}
	bizDate, err := parseDate(req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, "invalid business_date format, expected YYYY-MM-DD")
		return
	}

	actor, _ := actorID(r)
	result, err := h.compliance.RunPortfolioPostTradeCheck(r.Context(), contract.PortfolioPostTradeRequest{
		PortfolioID:  p.ID,
		FundID:       portfolioFundID(p),
		BusinessDate: bizDate,
		Actor:        actor.String(),
	})
	if err != nil {
		httputil.InternalError(w, "post-trade check failed: "+err.Error())
		return
	}
	httputil.Created(w, result)
}

// ============================================================
// GET /api/v2/portfolios/{portfolioCode}/compliance/rules
// ============================================================

// ListRulesByCode handles GET /api/v2/portfolios/{portfolioCode}/compliance/rules.
// @Summary List Compliance Rules For Portfolio (V2)
// @Description List the active rule catalog, annotated with each rule's binding to this portfolio when one exists.
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Success 200 {array} contract.PortfolioRuleCatalogEntry
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/compliance/rules [get]
func (h *InvestmentHandler) ListRulesByCode(w http.ResponseWriter, r *http.Request) {
	if h.compliance == nil {
		httputil.InternalError(w, "compliance service not available")
		return
	}
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	rules, err := h.compliance.ListPortfolioRules(r.Context(), p.ID)
	if err != nil {
		httputil.InternalError(w, "failed to list portfolio rules")
		return
	}
	httputil.OK(w, rules)
}

// ============================================================
// POST /api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings
// ============================================================

// BindRuleRequest is the V2 rule-binding payload. No scope_type/scope_id
// fields — binding is always scope_type=PORTFOLIO, scope_id=this portfolio.
type BindRuleRequest struct {
	Severity      string  `json:"severity"` // BLOCK | WARN | REQUIRE_APPROVAL | MONITOR
	Priority      int     `json:"priority,omitempty"`
	EffectiveFrom string  `json:"effective_from"`         // "2006-01-02"
	EffectiveTo   *string `json:"effective_to,omitempty"` // "2006-01-02" or null
}

// BindRuleByCode handles POST /api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings.
// @Summary Bind Compliance Rule To Portfolio (V2)
// @Description Bind an existing rule instance to this portfolio (scope_type=PORTFOLIO).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param ruleInstanceID path string true "Rule instance UUID"
// @Param request body BindRuleRequest true "Binding payload"
// @Success 201 {object} contract.PortfolioRuleBindingView
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings [post]
func (h *InvestmentHandler) BindRuleByCode(w http.ResponseWriter, r *http.Request) {
	if h.compliance == nil {
		httputil.InternalError(w, "compliance service not available")
		return
	}
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	ruleInstanceID, err := parseUUIDParam(r, "ruleInstanceID")
	if err != nil {
		httputil.BadRequest(w, "invalid rule instance id")
		return
	}

	var req BindRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body: "+err.Error())
		return
	}
	from, err := parseDate(req.EffectiveFrom)
	if err != nil {
		httputil.BadRequest(w, "invalid effective_from, expected YYYY-MM-DD")
		return
	}
	var to *time.Time
	if req.EffectiveTo != nil && *req.EffectiveTo != "" {
		parsed, err := parseDate(*req.EffectiveTo)
		if err != nil {
			httputil.BadRequest(w, "invalid effective_to, expected YYYY-MM-DD")
			return
		}
		to = &parsed
	}

	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "authentication required")
		return
	}

	binding, err := h.compliance.BindPortfolioRule(r.Context(), contract.PortfolioRuleBindingRequest{
		PortfolioID:    p.ID,
		RuleInstanceID: ruleInstanceID,
		Severity:       req.Severity,
		Priority:       req.Priority,
		EffectiveFrom:  from,
		EffectiveTo:    to,
		ActorID:        actor,
	})
	if err != nil {
		writePortfolioComplianceError(w, err)
		return
	}
	httputil.Created(w, binding)
}

// ============================================================
// DELETE /api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}
// ============================================================

// DeactivateRuleBindingByCode handles
// DELETE /api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}.
// @Summary Deactivate Portfolio Compliance Rule Binding (V2)
// @Description Deactivate a rule binding on this portfolio. Bindings owned by another portfolio are reported as 404.
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Param portfolioCode path string true "Portfolio code"
// @Param ruleInstanceID path string true "Rule instance UUID (unused for lookup; kept for a stable REST shape)"
// @Param bindingID path string true "Binding UUID"
// @Success 204 "No Content"
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID} [delete]
func (h *InvestmentHandler) DeactivateRuleBindingByCode(w http.ResponseWriter, r *http.Request) {
	if h.compliance == nil {
		httputil.InternalError(w, "compliance service not available")
		return
	}
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	bindingID, err := parseUUIDParam(r, "bindingID")
	if err != nil {
		httputil.BadRequest(w, "invalid binding id")
		return
	}

	if err := h.compliance.DeactivatePortfolioRuleBinding(r.Context(), p.ID, bindingID); err != nil {
		writePortfolioComplianceError(w, err)
		return
	}
	httputil.NoContent(w)
}

// ============================================================
// GET /api/v2/portfolios/{portfolioCode}/compliance/breaches
// ============================================================

// ListBreachesByCode handles GET /api/v2/portfolios/{portfolioCode}/compliance/breaches.
// @Summary List Compliance Breaches For Portfolio (V2)
// @Description List compliance breaches for a portfolio resolved by business code.
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param status query string false "Breach status"
// @Param rule_type_id query string false "Rule type ID"
// @Param date_from query string false "Start business date (YYYY-MM-DD)"
// @Param date_to query string false "End business date (YYYY-MM-DD)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {array} contract.PortfolioBreachView
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/compliance/breaches [get]
func (h *InvestmentHandler) ListBreachesByCode(w http.ResponseWriter, r *http.Request) {
	if h.compliance == nil {
		httputil.InternalError(w, "compliance service not available")
		return
	}
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}

	page, limit := paginationParams(r)
	filter := contract.PortfolioBreachFilter{
		RuleTypeID: r.URL.Query().Get("rule_type_id"),
		Offset:     (page - 1) * limit,
		Limit:      limit,
	}
	if v := r.URL.Query().Get("status"); v != "" {
		filter.Status = &v
	}
	if v := r.URL.Query().Get("date_from"); v != "" {
		if t, err := parseDate(v); err == nil {
			filter.DateFrom = &t
		}
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		if t, err := parseDate(v); err == nil {
			filter.DateTo = &t
		}
	}

	breaches, err := h.compliance.ListPortfolioBreaches(r.Context(), p.ID, filter)
	if err != nil {
		httputil.InternalError(w, "failed to list portfolio breaches")
		return
	}
	httputil.OK(w, breaches)
}
