package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// DecisionHandler exposes CRUD + lifecycle endpoints for investment decisions.
type DecisionHandler struct {
	decisions     domain.DecisionRepository
	lines         domain.DecisionLineRepository
	cmd           *command.DecisionCommandHandler
	batchApproval *command.DecisionBatchApprovalHandler
	approvalStage contract.ApprovalStatusProvider
	portfolios    domain.PortfolioRepository
	pc            contract.PermissionChecker
}

func NewDecisionHandler(repo domain.DecisionRepository, cmd *command.DecisionCommandHandler) *DecisionHandler {
	return &DecisionHandler{decisions: repo, cmd: cmd}
}

// SetPortfolioRepository wires the portfolio repository post-construction so
// the Portfolio V2 (portfolioCode) routes in portfolio_v2_decision_handler.go
// can resolve portfolioCode -> portfolio_id.
func (h *DecisionHandler) SetPortfolioRepository(r domain.PortfolioRepository) {
	if h != nil {
		h.portfolios = r
	}
}

// SetPermissionChecker wires the fund-scoped data-permission checker
// post-construction. Used both by the Portfolio V2 (portfolioCode) routes in
// portfolio_v2_decision_handler.go (via resolvePortfolioByCode's
// hasFundAccess check) and by the V1 {id}-keyed decision routes in this file
// (ListDecisions/GetDecision/GetDecisionWithLines/ListApprovalItems), which
// previously had no fund/portfolio data-scope enforcement at all beyond the
// route-level function permission. Production wiring in module.go always
// passes a real, fail-closed checker here (never nil) — a nil pc makes every
// fund-scope check in this file deny access (hasFundAccess/accessibleFundIDs
// both treat nil pc as "no access"), so an unwired handler fails closed
// rather than silently allowing cross-fund reads.
func (h *DecisionHandler) SetPermissionChecker(pc contract.PermissionChecker) {
	if h != nil {
		h.pc = pc
	}
}

// SetDecisionLineRepository wires the line repository post-construction.
func (h *DecisionHandler) SetDecisionLineRepository(r domain.DecisionLineRepository) {
	if h != nil {
		h.lines = r
	}
}

// SetBatchApprovalHandler wires the batch approval command handler post-construction.
func (h *DecisionHandler) SetBatchApprovalHandler(b *command.DecisionBatchApprovalHandler) {
	if h != nil {
		h.batchApproval = b
	}
}

// SetApprovalStatusProvider wires the approval stage provider post-construction.
func (h *DecisionHandler) SetApprovalStatusProvider(p contract.ApprovalStatusProvider) {
	if h != nil {
		h.approvalStage = p
	}
}

// @Summary List Investment Decisions
// @Description Returns a paginated list of investment decisions with optional filters.
// @Tags Investment - Decisions
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Param fund_id query string false "Filter by fund UUID"
// @Param portfolio_id query string false "Filter by portfolio UUID"
// @Param business_date query string false "Filter by business date YYYY-MM-DD"
// @Param status query string false "Filter by lifecycle status"
// @Param instrument_code query string false "Filter by instrument code"
// @Param search query string false "Substring search on decision_no / instrument_code"
// @Success 200 {object} response.DecisionListResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/decisions [get]
func (h *DecisionHandler) ListDecisions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, limit := paginationParams(r)
	filter := domain.DecisionListFilter{Page: page, Limit: limit}
	if v := strings.TrimSpace(q.Get("fund_id")); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httputil.BadRequest(w, "invalid fund_id")
			return
		}
		filter.FundID = &id
	}
	if v := strings.TrimSpace(q.Get("portfolio_id")); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httputil.BadRequest(w, "invalid portfolio_id")
			return
		}
		filter.PortfolioID = &id
	}
	if v := strings.TrimSpace(q.Get("business_date")); v != "" {
		t, err := parseDate(v)
		if err != nil {
			httputil.BadRequest(w, "invalid business_date (expected YYYY-MM-DD)")
			return
		}
		filter.BusinessDate = &t
	}
	if v := strings.TrimSpace(q.Get("status")); v != "" {
		s := vo.DecisionLifecycleStatus(v)
		if !s.IsValid() {
			httputil.BadRequest(w, "invalid status")
			return
		}
		filter.Status = &s
	}
	filter.InstrumentCode = strings.TrimSpace(q.Get("instrument_code"))
	filter.Search = strings.TrimSpace(q.Get("search"))

	// Data permission: scope to the funds the user can access, matching
	// InvestmentHandler.ListPortfolios's accessibleFundIDs pattern. nil means
	// unrestricted (global/company-wide data scope); a non-nil empty slice
	// means zero fund access, and the repository returns no rows.
	filter.AccessibleFundIDs = accessibleFundIDs(r.Context(), h.pc)

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

// GetDecision handles GET /investment/decisions/{id}.
// @Summary Get Investment Decision
// @Description Retrieve one investment decision by UUID. Requires data-permission on the decision's fund.
// @Tags Investment - Decisions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Decision UUID"
// @Success 200 {object} response.DecisionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/decisions/{id} [get]
func (h *DecisionHandler) GetDecision(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid decision id")
		return
	}
	d, err := h.decisions.GetByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if d == nil {
		httputil.NotFound(w, "decision not found")
		return
	}
	// Data permission: verify the decision's fund is within the caller's
	// accessible scope. hasFundAccess denies when h.pc is nil (fail closed).
	if !hasFundAccess(r.Context(), h.pc, decisionScopeID(d)) {
		httputil.Forbidden(w, "no access to this decision")
		return
	}
	httputil.OK(w, response.FromDecision(d))
}

// @Summary Create Investment Decision
// @Description Creates a new investment decision in DRAFT status.
// @Tags Investment - Decisions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.CreateDecisionRequest true "Decision fields"
// @Success 201 {object} response.DecisionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/decisions [post]
func (h *DecisionHandler) CreateDecision(w http.ResponseWriter, r *http.Request) {
	var req request.CreateDecisionRequest
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
		// Legacy V1 route: fund_id is always required in the request body here.
		FundID:           &req.FundID,
		PortfolioID:      req.PortfolioID,
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

// UpdateDecision handles PUT /investment/decisions/{id}.
func (h *DecisionHandler) UpdateDecision(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid decision id")
		return
	}
	var req request.UpdateDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	cmdReq := command.UpdateDecisionRequest{DecisionID: id, ActorID: actor}
	if req.BusinessDate != nil {
		t, err := parseDate(*req.BusinessDate)
		if err != nil {
			httputil.BadRequest(w, "invalid business_date (expected YYYY-MM-DD)")
			return
		}
		cmdReq.BusinessDate = &t
	}
	cmdReq.InstrumentID = req.InstrumentID
	cmdReq.InstrumentCode = req.InstrumentCode
	if req.Side != nil {
		s := vo.OrderSide(*req.Side)
		cmdReq.Side = &s
	}
	if req.Quantity != nil {
		q, err := parseDecimalOpt(*req.Quantity)
		if err != nil {
			httputil.BadRequest(w, "invalid quantity")
			return
		}
		cmdReq.Quantity = q
	}
	if req.Amount != nil {
		a, err := parseDecimalOpt(*req.Amount)
		if err != nil {
			httputil.BadRequest(w, "invalid amount")
			return
		}
		cmdReq.Amount = a
	}
	if req.LimitPrice != nil {
		p, err := parseDecimalOpt(*req.LimitPrice)
		if err != nil {
			httputil.BadRequest(w, "invalid limit_price")
			return
		}
		cmdReq.LimitPrice = p
	}
	cmdReq.Currency = req.Currency
	cmdReq.Exchange = req.Exchange
	cmdReq.ResearchReportID = req.ResearchReportID
	cmdReq.Rationale = req.Rationale

	d, err := h.cmd.Update(r.Context(), cmdReq)
	if err != nil {
		writeDecisionError(w, err)
		return
	}
	httputil.OK(w, response.FromDecision(d))
}

// @Summary Submit Investment Decision
// @Description Submits a DRAFT decision for approval. Transitions status to SUBMITTED.
// @Tags Investment - Decisions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Decision UUID"
// @Success 200 {object} response.DecisionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse "COMPLIANCE_NOT_CONFIGURED, COMPLIANCE_UNAVAILABLE, or evaluated rule rejection"
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/decisions/{id}/submit [post]
func (h *DecisionHandler) SubmitDecision(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid decision id")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	d, err := h.cmd.Submit(r.Context(), id, actor)
	if err != nil {
		writeDecisionError(w, err)
		return
	}
	httputil.OK(w, response.FromDecision(d))
}

// @Summary Cancel Investment Decision
// @Description Cancels a DRAFT or SUBMITTED decision. A cancellation reason is required.
// @Tags Investment - Decisions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Decision UUID"
// @Param body body request.CancelDecisionRequest true "Cancellation reason"
// @Success 200 {object} response.DecisionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/decisions/{id}/cancel [post]
func (h *DecisionHandler) CancelDecision(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid decision id")
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
	d, err := h.cmd.Cancel(r.Context(), command.CancelDecisionRequest{
		DecisionID: id, Reason: req.Reason, ActorID: actor,
	})
	if err != nil {
		writeDecisionError(w, err)
		return
	}
	httputil.OK(w, response.FromDecision(d))
}

// @Summary List Investment Decision Approval Items
// @Description Returns decisions pending approval (default status=PENDING_APPROVAL), enriched with current and previous approver names and stage numbers.
// @Tags Investment - Decisions
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Param portfolio_id query string false "Filter by portfolio UUID"
// @Param fund_id query string false "Filter by fund UUID"
// @Param business_date_from query string false "Inclusive lower bound YYYY-MM-DD"
// @Param business_date_to query string false "Inclusive upper bound YYYY-MM-DD"
// @Param decision_no query string false "Filter by exact decision number"
// @Param process_type query string false "INVESTMENT_DECISION | ORDER_CANCEL | ORDER_AMEND"
// @Param product_type query string false "MUTUAL_FUND | ETF | STOCK | BOND | CASH"
// @Param research_no query string false "Filter by research report number"
// @Param status query string false "Decision lifecycle status (default PENDING_APPROVAL)"
// @Param search query string false "Substring search on decision_no / instrument_code / research_report_no"
// @Success 200 {object} response.DecisionListResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/decisions/approval-items [get]
func (h *DecisionHandler) ListApprovalItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, limit := paginationParams(r)
	status := vo.DecisionLifecyclePendingApproval
	filter := domain.DecisionListFilter{
		Page:   page,
		Limit:  limit,
		Status: &status,
	}
	if v := strings.TrimSpace(q.Get("portfolio_id")); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httputil.BadRequest(w, "invalid portfolio_id")
			return
		}
		filter.PortfolioID = &id
	}
	if v := strings.TrimSpace(q.Get("fund_id")); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httputil.BadRequest(w, "invalid fund_id")
			return
		}
		filter.FundID = &id
	}
	if v := strings.TrimSpace(q.Get("business_date_from")); v != "" {
		t, err := parseDate(v)
		if err != nil {
			httputil.BadRequest(w, "invalid business_date_from (expected YYYY-MM-DD)")
			return
		}
		filter.BusinessDateFrom = &t
	}
	if v := strings.TrimSpace(q.Get("business_date_to")); v != "" {
		t, err := parseDate(v)
		if err != nil {
			httputil.BadRequest(w, "invalid business_date_to (expected YYYY-MM-DD)")
			return
		}
		filter.BusinessDateTo = &t
	}
	filter.DecisionNumber = strings.TrimSpace(q.Get("decision_no"))
	filter.ProcessType = strings.TrimSpace(q.Get("process_type"))
	filter.ProductType = strings.TrimSpace(q.Get("product_type"))
	filter.ResearchReportNo = strings.TrimSpace(q.Get("research_no"))
	filter.Search = strings.TrimSpace(q.Get("search"))

	// Allow overriding status for the approval screen (e.g. showing all or approved)
	if v := strings.TrimSpace(q.Get("status")); v != "" {
		s := vo.DecisionLifecycleStatus(v)
		filter.Status = &s
	}

	// Data permission: scope to the funds the user can access (see ListDecisions).
	filter.AccessibleFundIDs = accessibleFundIDs(r.Context(), h.pc)

	items, total, err := h.decisions.List(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	out := make([]response.DecisionResponse, 0, len(items))
	for _, d := range items {
		dr := response.FromDecision(d)
		// Enrich with approval stage info if the provider is wired.
		if h.approvalStage != nil && d.ApprovalRequestID != nil {
			stage, err := h.approvalStage.GetApprovalStage(r.Context(), "INVESTMENT_DECISION", d.ID)
			if err == nil && stage != nil {
				dr.ApprovalStage = stage.CurrentStageNumber
				dr.ApprovalTotalStages = stage.TotalStages
				for _, a := range stage.CurrentApprovers {
					dr.CurrentApprovers = append(dr.CurrentApprovers, a.DisplayName)
				}
				for _, a := range stage.PreviousApprovers {
					dr.PreviousApprovers = append(dr.PreviousApprovers, a.DisplayName)
				}
			}
		}
		out = append(out, dr)
	}
	httputil.OK(w, response.DecisionListResponse{Items: out, Total: total, Page: page, Limit: limit})
}

// @Summary Get Investment Decision With Lines
// @Description Returns a decision header with all child decision lines (for basket/rebalance/switch decisions) and approval stage info.
// @Tags Investment - Decisions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Decision UUID"
// @Success 200 {object} response.DecisionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/decisions/{id}/details [get]
func (h *DecisionHandler) GetDecisionWithLines(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid decision id")
		return
	}
	d, err := h.decisions.GetByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if d == nil {
		httputil.NotFound(w, "decision not found")
		return
	}
	// Data permission: verify the decision's fund is within the caller's
	// accessible scope. hasFundAccess denies when h.pc is nil (fail closed).
	if !hasFundAccess(r.Context(), h.pc, decisionScopeID(d)) {
		httputil.Forbidden(w, "no access to this decision")
		return
	}
	if h.lines != nil {
		d.Lines, _ = h.lines.ListByDecision(r.Context(), d.ID)
	}
	dr := response.FromDecision(d)
	if h.approvalStage != nil && d.ApprovalRequestID != nil {
		stage, err := h.approvalStage.GetApprovalStage(r.Context(), "INVESTMENT_DECISION", d.ID)
		if err == nil && stage != nil {
			dr.ApprovalStage = stage.CurrentStageNumber
			dr.ApprovalTotalStages = stage.TotalStages
			for _, a := range stage.CurrentApprovers {
				dr.CurrentApprovers = append(dr.CurrentApprovers, a.DisplayName)
			}
			for _, a := range stage.PreviousApprovers {
				dr.PreviousApprovers = append(dr.PreviousApprovers, a.DisplayName)
			}
		}
	}
	httputil.OK(w, dr)
}

// @Summary Batch Approve Investment Decisions
// @Description Approves multiple investment decision headers in a single call. Each decision must be PENDING_APPROVAL and have a pending approval task assigned to the authenticated user. Returns per-decision results — partial failures are reported without aborting the rest.
// @Tags Investment - Decisions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.BatchApprovalRequest true "Decision numbers and optional comment"
// @Success 200 {object} response.BatchApprovalResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/decisions/batch-approve [post]
func (h *DecisionHandler) BatchApproveDecisions(w http.ResponseWriter, r *http.Request) {
	if h.batchApproval == nil {
		httputil.InternalError(w, "batch approval not configured")
		return
	}
	var req request.BatchApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	results, err := h.batchApproval.BatchApprove(r.Context(), command.BatchApproveRequest{
		DecisionNos: req.DecisionNos,
		Comment:     req.Comment,
		ActorID:     actor,
	})
	if err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}
	out := buildBatchResponse(results)
	httputil.OK(w, out)
}

// @Summary Batch Reject Investment Decisions
// @Description Rejects multiple investment decision headers in a single call. A rejection reason is required and applied to all selected decisions. Returns per-decision results.
// @Tags Investment - Decisions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.BatchRejectionRequest true "Decision numbers and required reason"
// @Success 200 {object} response.BatchApprovalResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/decisions/batch-reject [post]
func (h *DecisionHandler) BatchRejectDecisions(w http.ResponseWriter, r *http.Request) {
	if h.batchApproval == nil {
		httputil.InternalError(w, "batch approval not configured")
		return
	}
	var req request.BatchRejectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	results, err := h.batchApproval.BatchReject(r.Context(), command.BatchRejectRequest{
		DecisionNos: req.DecisionNos,
		Reason:      req.Reason,
		ActorID:     actor,
	})
	if err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}
	out := buildBatchResponse(results)
	httputil.OK(w, out)
}

func buildBatchResponse(results []command.BatchApprovalResult) response.BatchApprovalResponse {
	out := response.BatchApprovalResponse{
		Results: make([]response.BatchApprovalResultResponse, 0, len(results)),
	}
	for _, r := range results {
		out.Results = append(out.Results, response.BatchApprovalResultResponse{
			DecisionNo: r.DecisionNo,
			OK:         r.OK,
			Error:      r.Error,
		})
		if r.OK {
			out.Succeeded++
		} else {
			out.Failed++
		}
	}
	return out
}

// writeDecisionError maps decision/lifecycle errors to HTTP codes.
func writeDecisionError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var (
		invalid     *domain.ErrInvalidDecisionRequest
		notFound    *domain.ErrDecisionNotFound
		lifecycle   *domain.ErrDecisionLifecycle
		referenceB  *domain.ErrDecisionReferenceInvalid
		number      *domain.ErrDecisionNumberConflict
		compBlocked *domain.ErrComplianceRejected
		compMissing *command.ErrComplianceNotConfigured
		compDown    *command.ErrComplianceUnavailable
	)
	switch {
	case errors.As(err, &invalid):
		httputil.BadRequest(w, err.Error())
	case errors.As(err, &notFound):
		httputil.NotFound(w, err.Error())
	case errors.As(err, &lifecycle):
		httputil.Conflict(w, err.Error())
	case errors.As(err, &referenceB):
		httputil.UnprocessableEntity(w, err.Error())
	case errors.As(err, &number):
		httputil.Conflict(w, err.Error())
	case errors.As(err, &compBlocked):
		httputil.UnprocessableEntity(w, err.Error())
	case errors.As(err, &compMissing):
		writeComplianceStatusError(w, command.ComplianceErrorCodeNotConfigured, compMissing.Error(), compMissing.CheckGroupID)
	case errors.As(err, &compDown):
		writeComplianceStatusError(w, command.ComplianceErrorCodeUnavailable, compDown.Error(), compDown.CheckGroupID)
	default:
		httputil.InternalError(w, err.Error())
	}
}
