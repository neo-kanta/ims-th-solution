package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// loadOwnedExecution loads an execution by the {executionId} path param and
// verifies it belongs to portfolioID. Writes 404 and returns (nil, false)
// both when the execution doesn't exist and when it belongs to a different
// portfolio — deliberately indistinguishable, mirroring loadOwnedDecision.
func (h *ExecutionHandler) loadOwnedExecution(w http.ResponseWriter, r *http.Request, portfolioID uuid.UUID) (*entity.Execution, bool) {
	executionID, err := parseUUID(chi.URLParam(r, "executionId"))
	if err != nil {
		httputil.BadRequest(w, "invalid execution id")
		return nil, false
	}
	e, err := h.executions.GetByID(r.Context(), executionID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return nil, false
	}
	if e == nil || e.PortfolioID != portfolioID {
		httputil.NotFound(w, "execution not found")
		return nil, false
	}
	return e, true
}

// Portfolio V2 execution and trade-confirmation endpoints
// (docs/api/portfolio-v2-api-ddd.md, Milestone 5 of
// docs/handoff/portfolio-v2-claude-implementation-prompt.md).
//
// Unlike decisions, V1's CreateExecutionRequest/RecordConfirmationRequest
// already derive fund_id/portfolio_id/contract_id transitively — an
// execution is created from its parent decision
// (application/command/execution.go's Create loads the decision and copies
// its fund/portfolio/contract fields) and a confirmation is created from
// its parent execution the same way. So V2 does not need to derive
// anything here; it only has to prove ownership of the *parent* entity
// against the resolved portfolio before delegating to the same V1 command
// handlers, exactly like ReverseTransactionByCode does for ledger rows
// (portfolio_v2_ledger_handler.go).

// CreateExecutionByCode handles
// POST /api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/executions.
// @Summary Create Portfolio Execution By Code
// @Description Opens an execution against an APPROVED decision that belongs to the resolved portfolio (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param decisionId path string true "Decision UUID"
// @Param request body request.CreateExecutionV2Request true "Execution fields"
// @Success 201 {object} response.ExecutionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse "COMPLIANCE_NOT_CONFIGURED, COMPLIANCE_UNAVAILABLE, or evaluated rule rejection"
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/decisions/{decisionId}/executions [post]
func (h *ExecutionHandler) CreateExecutionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, nil)
	if !ok {
		return
	}
	decisionID, err := parseUUID(chi.URLParam(r, "decisionId"))
	if err != nil {
		httputil.BadRequest(w, "invalid decision id")
		return
	}
	decision, err := h.decisions.GetByID(r.Context(), decisionID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if decision == nil || decision.PortfolioID != p.ID {
		httputil.NotFound(w, "decision not found")
		return
	}
	var req request.CreateExecutionV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	q, err := parseDecimalOpt(req.OrderedQuantity)
	if err != nil {
		httputil.BadRequest(w, "invalid ordered_quantity")
		return
	}
	a, err := parseDecimalOpt(req.OrderedAmount)
	if err != nil {
		httputil.BadRequest(w, "invalid ordered_amount")
		return
	}
	e, err := h.cmd.Create(r.Context(), command.CreateExecutionRequest{
		DecisionID:      decisionID,
		OrderedQuantity: q,
		OrderedAmount:   a,
		BrokerReference: req.BrokerReference,
		ActorID:         actor,
	})
	if err != nil {
		writeExecutionError(w, err)
		return
	}
	httputil.Created(w, response.FromExecution(e))
}

// FillExecutionByCode handles
// POST /api/v2/portfolios/{portfolioCode}/executions/{executionId}/fill.
// @Summary Fill Portfolio Execution By Code
// @Description Records a fill against an execution that belongs to the resolved portfolio (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param executionId path string true "Execution UUID"
// @Param request body request.FillExecutionRequest true "Fill fields"
// @Success 200 {object} response.ExecutionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/executions/{executionId}/fill [post]
func (h *ExecutionHandler) FillExecutionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, nil)
	if !ok {
		return
	}
	execution, ok := h.loadOwnedExecution(w, r, p.ID)
	if !ok {
		return
	}
	var req request.FillExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	q, err := parseDecimalOpt(req.ExecutedQuantity)
	if err != nil {
		httputil.BadRequest(w, "invalid executed_quantity")
		return
	}
	a, err := parseDecimalOpt(req.ExecutedAmount)
	if err != nil {
		httputil.BadRequest(w, "invalid executed_amount")
		return
	}
	price, err := parseDecimalOpt(req.ExecutionPrice)
	if err != nil {
		httputil.BadRequest(w, "invalid execution_price")
		return
	}
	updated, err := h.cmd.Fill(r.Context(), command.FillExecutionRequest{
		ExecutionID:      execution.ID,
		ExecutedQuantity: q,
		ExecutedAmount:   a,
		ExecutionPrice:   price,
		Status:           vo.ExecutionStatus(req.Status),
		BrokerReference:  req.BrokerReference,
		ActorID:          actor,
	})
	if err != nil {
		writeExecutionError(w, err)
		return
	}
	httputil.OK(w, response.FromExecution(updated))
}

// CancelExecutionByCode handles
// POST /api/v2/portfolios/{portfolioCode}/executions/{executionId}/cancel.
// @Summary Cancel Portfolio Execution By Code
// @Description Cancels an execution that belongs to the resolved portfolio (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param executionId path string true "Execution UUID"
// @Param request body request.CancelExecutionRequest true "Cancellation reason"
// @Success 200 {object} response.ExecutionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/executions/{executionId}/cancel [post]
func (h *ExecutionHandler) CancelExecutionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, nil)
	if !ok {
		return
	}
	execution, ok := h.loadOwnedExecution(w, r, p.ID)
	if !ok {
		return
	}
	var req request.CancelExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	updated, err := h.cmd.Cancel(r.Context(), command.CancelExecutionRequest{
		ExecutionID: execution.ID, Reason: req.Reason, ActorID: actor,
	})
	if err != nil {
		writeExecutionError(w, err)
		return
	}
	httputil.OK(w, response.FromExecution(updated))
}

// RecordConfirmationByCode handles
// POST /api/v2/portfolios/{portfolioCode}/executions/{executionId}/confirmations.
// @Summary Record Portfolio Trade Confirmation By Code
// @Description Records a broker confirmation against an execution that belongs to the resolved portfolio (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param executionId path string true "Execution UUID"
// @Param request body request.RecordConfirmationV2Request true "Confirmation fields"
// @Success 201 {object} response.TradeConfirmationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/executions/{executionId}/confirmations [post]
func (h *TradeConfirmationHandler) RecordConfirmationByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, nil)
	if !ok {
		return
	}
	executionID, err := parseUUID(chi.URLParam(r, "executionId"))
	if err != nil {
		httputil.BadRequest(w, "invalid execution id")
		return
	}
	execution, err := h.executions.GetByID(r.Context(), executionID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if execution == nil || execution.PortfolioID != p.ID {
		httputil.NotFound(w, "execution not found")
		return
	}
	var req request.RecordConfirmationV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	q, err := parseDecimalOpt(req.ConfirmedQuantity)
	if err != nil {
		httputil.BadRequest(w, "invalid confirmed_quantity")
		return
	}
	a, err := parseDecimalOpt(req.ConfirmedAmount)
	if err != nil {
		httputil.BadRequest(w, "invalid confirmed_amount")
		return
	}
	price, err := parseDecimalOpt(req.ConfirmedPrice)
	if err != nil {
		httputil.BadRequest(w, "invalid confirmed_price")
		return
	}
	c, err := h.cmd.Record(r.Context(), command.RecordConfirmationRequest{
		ExecutionID:       executionID,
		ConfirmedQuantity: q,
		ConfirmedAmount:   a,
		ConfirmedPrice:    price,
		BrokerReference:   req.BrokerReference,
		ImportBatchID:     req.ImportBatchID,
		ActorID:           actor,
	})
	if err != nil {
		writeConfirmationError(w, err)
		return
	}
	httputil.Created(w, response.FromTradeConfirmation(c))
}

// ResolveConfirmationByCode handles
// POST /api/v2/portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve.
// @Summary Resolve Portfolio Trade Confirmation By Code
// @Description Matches or reviews a trade confirmation that belongs to the resolved portfolio (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param confirmationId path string true "Confirmation UUID"
// @Param request body request.ResolveConfirmationRequest true "Resolution fields"
// @Success 200 {object} response.TradeConfirmationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve [post]
func (h *TradeConfirmationHandler) ResolveConfirmationByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, nil)
	if !ok {
		return
	}
	confirmationID, err := parseUUID(chi.URLParam(r, "confirmationId"))
	if err != nil {
		httputil.BadRequest(w, "invalid confirmation id")
		return
	}
	confirmation, err := h.confirmations.GetByID(r.Context(), confirmationID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if confirmation == nil || confirmation.PortfolioID != p.ID {
		httputil.NotFound(w, "trade confirmation not found")
		return
	}
	var req request.ResolveConfirmationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	updated, err := h.cmd.Resolve(r.Context(), command.ResolveConfirmationRequest{
		ConfirmationID:    confirmationID,
		TargetStatus:      vo.TradeConfirmationStatus(req.TargetStatus),
		DiscrepancyReason: req.DiscrepancyReason,
		ActorID:           actor,
	})
	if err != nil {
		writeConfirmationError(w, err)
		return
	}
	httputil.OK(w, response.FromTradeConfirmation(updated))
}
