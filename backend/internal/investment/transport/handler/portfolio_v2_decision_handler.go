package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// Portfolio V2 decision endpoints (docs/api/portfolio-v2-api-ddd.md,
// Milestone 5 of
// docs/handoff/portfolio-v2-claude-implementation-prompt.md).
//
// Unlike Milestone 2/3's transactions, V1's request.CreateDecisionRequest
// requires the client to supply fund_id and portfolio_id directly
// (application/command/decision_lifecycle.go's CreateDecisionRequest has no
// server-side derivation). So V2 create must do real work here: resolve
// portfolioCode -> portfolio, then derive FundID from the portfolio before
// calling the same DecisionCommandHandler.Create used by V1.

// loadOwnedDecision loads a decision by the {decisionId} path param and
// verifies it belongs to portfolioID. Writes 404 and returns (nil, false)
// both when the decision doesn't exist and when it belongs to a different
// portfolio — deliberately indistinguishable, so a caller cannot probe for
// the existence of another portfolio's decisions via this route.
func (h *DecisionHandler) loadOwnedDecision(w http.ResponseWriter, r *http.Request, portfolioID uuid.UUID) (*entity.Decision, bool) {
	decisionID, err := parseUUID(chi.URLParam(r, "decisionId"))
	if err != nil {
		httputil.BadRequest(w, "invalid decision id")
		return nil, false
	}
	d, err := h.decisions.GetByID(r.Context(), decisionID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return nil, false
	}
	if d == nil || d.PortfolioID != portfolioID {
		httputil.NotFound(w, "decision not found")
		return nil, false
	}
	return d, true
}

// GetDecisionsByCode handles GET /api/v2/portfolios/{portfolioCode}/decisions.
// @Summary List Portfolio Decisions By Code
// @Description List investment decisions for a portfolio, resolved by business code (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.DecisionListResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/decisions [get]
func (h *DecisionHandler) ListDecisionsByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	page, limit := paginationParams(r)
	filter := domain.DecisionListFilter{PortfolioID: &p.ID, Page: page, Limit: limit}
	items, total, err := h.decisions.List(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.DecisionResponse, 0, len(items))
	for _, d := range items {
		out = append(out, response.FromDecision(d))
	}
	httputil.OK(w, response.DecisionListResponse{Items: out, Total: total, Page: page, Limit: limit})
}

// CreateDecisionByCode handles POST /api/v2/portfolios/{portfolioCode}/decisions.
// @Summary Create Portfolio Decision By Code
// @Description Creates a new investment decision in DRAFT status for a portfolio, resolved by business code (Portfolio V2). Request body must not include fund_id, portfolio_id, or contract_id.
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param request body request.CreateDecisionV2Request true "Decision fields"
// @Success 201 {object} response.DecisionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/decisions [post]
func (h *DecisionHandler) CreateDecisionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	var req request.CreateDecisionV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	bDate, err := parseDate(req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, "invalid business_date (expected YYYY-MM-DD)")
		return
	}
	qty, err := parseDecimalOpt(req.Quantity)
	if err != nil {
		httputil.BadRequest(w, "invalid quantity")
		return
	}
	amt, err := parseDecimalOpt(req.Amount)
	if err != nil {
		httputil.BadRequest(w, "invalid amount")
		return
	}
	limit, err := parseDecimalOpt(req.LimitPrice)
	if err != nil {
		httputil.BadRequest(w, "invalid limit_price")
		return
	}

	d, err := h.cmd.Create(r.Context(), command.CreateDecisionRequest{
		// Derived from the resolved portfolio — never trusted from the client.
		FundID:           p.FundID,
		PortfolioID:      p.ID,
		InstrumentID:     req.InstrumentID,
		InstrumentCode:   req.InstrumentCode,
		BusinessDate:     bDate,
		Side:             vo.OrderSide(req.Side),
		Quantity:         qty,
		Amount:           amt,
		LimitPrice:       limit,
		Currency:         req.Currency,
		Exchange:         req.Exchange,
		ResearchReportID: req.ResearchReportID,
		Rationale:        req.Rationale,
		ActorID:          actor,
	})
	if err != nil {
		writeDecisionError(w, err)
		return
	}
	httputil.Created(w, response.FromDecision(d))
}

// GetDecisionByCode handles GET /api/v2/portfolios/{portfolioCode}/decisions/{decisionId}.
// @Summary Get Portfolio Decision By Code
// @Description Retrieve one investment decision that belongs to the resolved portfolio (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param decisionId path string true "Decision UUID"
// @Success 200 {object} response.DecisionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/decisions/{decisionId} [get]
func (h *DecisionHandler) GetDecisionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	d, ok := h.loadOwnedDecision(w, r, p.ID)
	if !ok {
		return
	}
	httputil.OK(w, response.FromDecision(d))
}

// SubmitDecisionByCode handles POST /api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/submit.
// @Summary Submit Portfolio Decision By Code
// @Description Submits a DRAFT decision that belongs to the resolved portfolio for approval (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param decisionId path string true "Decision UUID"
// @Success 200 {object} response.DecisionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse "COMPLIANCE_NOT_CONFIGURED, COMPLIANCE_UNAVAILABLE, or evaluated rule rejection"
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/decisions/{decisionId}/submit [post]
func (h *DecisionHandler) SubmitDecisionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	d, ok := h.loadOwnedDecision(w, r, p.ID)
	if !ok {
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	updated, err := h.cmd.Submit(r.Context(), d.ID, actor)
	if err != nil {
		writeDecisionError(w, err)
		return
	}
	httputil.OK(w, response.FromDecision(updated))
}

// CancelDecisionByCode handles POST /api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/cancel.
// @Summary Cancel Portfolio Decision By Code
// @Description Cancels a DRAFT or SUBMITTED decision that belongs to the resolved portfolio (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param decisionId path string true "Decision UUID"
// @Param request body request.CancelDecisionRequest true "Cancellation reason"
// @Success 200 {object} response.DecisionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/decisions/{decisionId}/cancel [post]
func (h *DecisionHandler) CancelDecisionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	d, ok := h.loadOwnedDecision(w, r, p.ID)
	if !ok {
		return
	}
	var req request.CancelDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	updated, err := h.cmd.Cancel(r.Context(), command.CancelDecisionRequest{
		DecisionID: d.ID, Reason: req.Reason, ActorID: actor,
	})
	if err != nil {
		writeDecisionError(w, err)
		return
	}
	httputil.OK(w, response.FromDecision(updated))
}
