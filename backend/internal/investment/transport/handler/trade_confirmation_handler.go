package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

type TradeConfirmationHandler struct {
	confirmations domain.TradeConfirmationRepository
	executions    domain.ExecutionRepository
	portfolios    domain.PortfolioRepository
	cmd           *command.TradeConfirmationCommandHandler
	batchImport   *command.ConfirmationBatchImportHandler
	pc            contract.PermissionChecker
}

func NewTradeConfirmationHandler(repo domain.TradeConfirmationRepository, cmd *command.TradeConfirmationCommandHandler) *TradeConfirmationHandler {
	return &TradeConfirmationHandler{confirmations: repo, cmd: cmd}
}

// SetExecutionRepository wires the execution repository post-construction so
// the Portfolio V2 (portfolioCode) route
// POST /portfolios/{portfolioCode}/executions/{executionId}/confirmations
// can verify the execution belongs to the resolved portfolio before
// recording a confirmation.
func (h *TradeConfirmationHandler) SetExecutionRepository(r domain.ExecutionRepository) {
	if h != nil {
		h.executions = r
	}
}

// SetPortfolioRepository wires the portfolio repository post-construction so
// the Portfolio V2 (portfolioCode) routes can resolve portfolioCode ->
// portfolio_id.
func (h *TradeConfirmationHandler) SetPortfolioRepository(r domain.PortfolioRepository) {
	if h != nil {
		h.portfolios = r
	}
}

// SetPermissionChecker wires the fund-scoped data-permission checker
// post-construction. Used both by the Portfolio V2 (portfolioCode) routes
// (via resolvePortfolioByCode's hasFundAccess check) and by the V1
// {id}-keyed confirmation routes in this file
// (ListConfirmations/GetConfirmation), which previously had no fund/
// portfolio data-scope enforcement at all beyond the route-level function
// permission. Production wiring in module.go always passes a real,
// fail-closed checker here (never nil).
func (h *TradeConfirmationHandler) SetPermissionChecker(pc contract.PermissionChecker) {
	if h != nil {
		h.pc = pc
	}
}

// SetBatchImportHandler wires the batch importer after construction. Kept as a
// setter so the existing constructor signature (and its test call-sites)
// remains stable.
func (h *TradeConfirmationHandler) SetBatchImportHandler(b *command.ConfirmationBatchImportHandler) {
	if h != nil {
		h.batchImport = b
	}
}

// ImportBatch handles POST /investment/trade-confirmations/batch.
// The request body is a structured JSON payload describing N rows; each row
// becomes a PENDING_REVIEW trade confirmation when validation passes. Failed
// rows are reported with per-row error messages and do not roll back the
// successful rows.
func (h *TradeConfirmationHandler) ImportBatch(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.batchImport == nil {
		httputil.InternalError(w, "batch import not wired")
		return
	}
	var req request.ImportConfirmationBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	rows := make([]command.ConfirmationBatchImportRow, 0, len(req.Rows))
	for _, in := range req.Rows {
		rows = append(rows, command.ConfirmationBatchImportRow{
			ExecutionID:       in.ExecutionID,
			ConfirmedQuantity: in.ConfirmedQuantity,
			ConfirmedAmount:   in.ConfirmedAmount,
			ConfirmedPrice:    in.ConfirmedPrice,
			BrokerReference:   in.BrokerReference,
		})
	}

	result, err := h.batchImport.Handle(r.Context(), command.ImportConfirmationBatchRequest{
		SourceFilename: req.SourceFilename,
		Rows:           rows,
		ActorID:        actor,
	})
	if err != nil {
		writeConfirmationError(w, err)
		return
	}

	resp := response.ConfirmationBatchImportResponse{
		BatchID:         result.BatchID,
		Status:          string(result.Status),
		TotalRecords:    result.TotalRecords,
		AcceptedRecords: result.AcceptedRecords,
		RejectedRecords: result.RejectedRecords,
		Rows:            make([]response.ConfirmationBatchImportRowResponse, 0, len(result.Rows)),
	}
	for _, row := range result.Rows {
		resp.Rows = append(resp.Rows, response.ConfirmationBatchImportRowResponse{
			RowIndex:       row.RowIndex,
			Accepted:       row.Accepted,
			ConfirmationID: row.ConfirmationID,
			Error:          row.Error,
		})
	}
	httputil.Created(w, resp)
}

// ListConfirmations handles GET /investment/trade-confirmations.
// @Summary List Trade Confirmations
// @Description List trade confirmations for an execution, or for a fund+business_date. Requires data-permission on the resolved fund.
// @Tags Investment - Trade Confirmations
// @Security BearerAuth
// @Produce json
// @Param execution_id query string false "Filter by execution UUID"
// @Param fund_id query string false "Filter by fund UUID (requires business_date)"
// @Param business_date query string false "Business date YYYY-MM-DD (required with fund_id)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/trade-confirmations [get]
func (h *TradeConfirmationHandler) ListConfirmations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if v := q.Get("execution_id"); v != "" {
		id, err := parseUUID(v)
		if err != nil {
			httputil.BadRequest(w, "invalid execution_id")
			return
		}
		// Data permission: resolve the execution's fund and verify access
		// before revealing any confirmation under it.
		execution, err := h.executions.GetByID(r.Context(), id)
		if err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		if execution == nil {
			httputil.NotFound(w, "execution not found")
			return
		}
		if !hasFundAccess(r.Context(), h.pc, execution.FundID) {
			httputil.Forbidden(w, "no access to this execution")
			return
		}
		items, err := h.confirmations.ListByExecution(r.Context(), id)
		if err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		out := make([]response.TradeConfirmationResponse, 0, len(items))
		for _, c := range items {
			out = append(out, response.FromTradeConfirmation(c))
		}
		httputil.OK(w, map[string]any{"items": out})
		return
	}
	if v := q.Get("fund_id"); v != "" {
		id, err := parseUUID(v)
		if err != nil {
			httputil.BadRequest(w, "invalid fund_id")
			return
		}
		bd, err := parseDate(q.Get("business_date"))
		if err != nil {
			httputil.BadRequest(w, "business_date required (YYYY-MM-DD)")
			return
		}
		// Data permission: verify access to the requested fund directly.
		if !hasFundAccess(r.Context(), h.pc, id) {
			httputil.Forbidden(w, "no access to this fund")
			return
		}
		items, err := h.confirmations.ListByFundDate(r.Context(), id, bd)
		if err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		out := make([]response.TradeConfirmationResponse, 0, len(items))
		for _, c := range items {
			out = append(out, response.FromTradeConfirmation(c))
		}
		httputil.OK(w, map[string]any{"items": out})
		return
	}
	httputil.BadRequest(w, "execution_id or (fund_id+business_date) is required")
}

// GetConfirmation handles GET /investment/trade-confirmations/{id}.
// @Summary Get Trade Confirmation
// @Description Retrieve one trade confirmation by UUID. Requires data-permission on the confirmation's fund.
// @Tags Investment - Trade Confirmations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Confirmation UUID"
// @Success 200 {object} response.TradeConfirmationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/trade-confirmations/{id} [get]
func (h *TradeConfirmationHandler) GetConfirmation(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid confirmation id")
		return
	}
	c, err := h.confirmations.GetByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if c == nil {
		httputil.NotFound(w, "trade confirmation not found")
		return
	}
	// Data permission: verify the confirmation's fund is within the caller's
	// accessible scope. hasFundAccess denies when h.pc is nil (fail closed).
	if !hasFundAccess(r.Context(), h.pc, c.FundID) {
		httputil.Forbidden(w, "no access to this trade confirmation")
		return
	}
	httputil.OK(w, response.FromTradeConfirmation(c))
}

func (h *TradeConfirmationHandler) RecordConfirmation(w http.ResponseWriter, r *http.Request) {
	var req request.RecordConfirmationRequest
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
	p, err := parseDecimalOpt(req.ConfirmedPrice)
	if err != nil {
		httputil.BadRequest(w, "invalid confirmed_price")
		return
	}
	c, err := h.cmd.Record(r.Context(), command.RecordConfirmationRequest{
		ExecutionID:       req.ExecutionID,
		ConfirmedQuantity: q,
		ConfirmedAmount:   a,
		ConfirmedPrice:    p,
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

func (h *TradeConfirmationHandler) ResolveConfirmation(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid confirmation id")
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
	target := vo.TradeConfirmationStatus(req.TargetStatus)
	c, err := h.cmd.Resolve(r.Context(), command.ResolveConfirmationRequest{
		ConfirmationID:    id,
		TargetStatus:      target,
		DiscrepancyReason: req.DiscrepancyReason,
		ActorID:           actor,
	})
	if err != nil {
		writeConfirmationError(w, err)
		return
	}
	httputil.OK(w, response.FromTradeConfirmation(c))
}

func writeConfirmationError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var (
		invalid    *domain.ErrInvalidDecisionRequest
		execNot    *domain.ErrExecutionNotFound
		execLife   *domain.ErrExecutionLifecycle
		confNot    *domain.ErrConfirmationNotFound
		confLife   *domain.ErrConfirmationLifecycle
		confReason *domain.ErrConfirmationMismatchReasonRequired
	)
	switch {
	case errors.As(err, &invalid):
		httputil.BadRequest(w, err.Error())
	case errors.As(err, &execNot), errors.As(err, &confNot):
		httputil.NotFound(w, err.Error())
	case errors.As(err, &execLife), errors.As(err, &confLife):
		httputil.Conflict(w, err.Error())
	case errors.As(err, &confReason):
		httputil.UnprocessableEntity(w, err.Error())
	default:
		httputil.InternalError(w, err.Error())
	}
}
