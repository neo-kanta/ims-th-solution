package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// ResearchReportHandler exposes the basic CRUD + submit endpoints for
// investment research reports. Kept in its own file so it does not enlarge
// the already-busy InvestmentHandler.
type ResearchReportHandler struct {
	reports        domain.ResearchReportRepository
	cmd            *command.ResearchReportCommandHandler
	approvalStatus contract.ApprovalStatusProvider
}

// NewResearchReportHandler wires the handler. approvalStatus may be nil — when
// unset, the handler falls back to the report's own review_status for the
// derived stage label.
func NewResearchReportHandler(
	reports domain.ResearchReportRepository,
	cmd *command.ResearchReportCommandHandler,
) *ResearchReportHandler {
	return &ResearchReportHandler{reports: reports, cmd: cmd}
}

// SetApprovalStatusProvider injects the cross-module approval lookup used to
// enrich response DTOs with the precise multi-level review stage. Wired
// post-construction from cmd/server/main.go to avoid a circular dependency.
func (h *ResearchReportHandler) SetApprovalStatusProvider(p contract.ApprovalStatusProvider) {
	if h != nil {
		h.approvalStatus = p
	}
}

// enrichReviewStage overlays the live approval stage onto a single response
// when the approval status provider is wired. Falls back to the response's
// existing (status-derived) value on any lookup error so the read path never
// fails because of a cross-module hiccup.
func (h *ResearchReportHandler) enrichReviewStage(ctx context.Context, resp *response.ResearchReportResponse) {
	if h == nil || h.approvalStatus == nil || resp == nil {
		return
	}
	info, err := h.approvalStatus.GetApprovalStage(ctx, "RESEARCH_REPORT", resp.ID)
	if err != nil || info == nil {
		return
	}
	switch info.Status {
	case "PENDING_APPROVAL":
		if info.CurrentStageNumber > 0 {
			resp.DerivedReviewStage = fmt.Sprintf("PENDING_LEVEL_%d", info.CurrentStageNumber)
		}
	case "APPROVED":
		// Leave whatever the response already says — if the callback ran the
		// report status is already REVIEW_COMPLETED.
	case "REJECTED":
		resp.DerivedReviewStage = "REJECTED"
	case "CANCELLED", "WITHDRAWN":
		resp.DerivedReviewStage = "NOT_SUBMITTED"
	}
}

// ListResearchReports handles GET /investment/research-reports.
// @Summary List Investment Research Reports
// @Description Paginated list of research reports with optional filters.
// @Tags Investment - Research
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Param report_status query string false "DRAFT | ACTIVE | EXPIRED | REJECTED"
// @Param review_status query string false "NOT_SUBMITTED | SUBMITTED | REVIEW_COMPLETED"
// @Param recommendation query string false "BUY | SELL | HOLD"
// @Param instrument_code query string false "Exact instrument code filter"
// @Param owner_user_id query string false "Owner user UUID"
// @Param report_date_from query string false "Inclusive lower bound YYYY-MM-DD"
// @Param report_date_to query string false "Inclusive upper bound YYYY-MM-DD"
// @Param search query string false "Substring search on report_no/instrument_code/report_title"
// @Success 200 {object} response.ResearchReportListResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/research-reports [get]
func (h *ResearchReportHandler) ListResearchReports(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, limit := paginationParams(r)
	filter := domain.ResearchReportListFilter{Page: page, Limit: limit}

	if v := strings.TrimSpace(q.Get("report_status")); v != "" {
		s := vo.ReportStatus(v)
		if !s.IsValid() {
			httputil.BadRequest(w, "invalid report_status")
			return
		}
		filter.ReportStatus = &s
	}
	if v := strings.TrimSpace(q.Get("review_status")); v != "" {
		s := vo.ReviewStatus(v)
		if !s.IsValid() {
			httputil.BadRequest(w, "invalid review_status")
			return
		}
		filter.ReviewStatus = &s
	}
	if v := strings.TrimSpace(q.Get("recommendation")); v != "" {
		rec := vo.Recommendation(v)
		if !rec.IsValid() {
			httputil.BadRequest(w, "invalid recommendation")
			return
		}
		filter.Recommendation = &rec
	}
	if v := strings.TrimSpace(q.Get("instrument_code")); v != "" {
		filter.InstrumentCode = v
	}
	if v := strings.TrimSpace(q.Get("owner_user_id")); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httputil.BadRequest(w, "invalid owner_user_id")
			return
		}
		filter.OwnerUserID = &id
	}
	if v := strings.TrimSpace(q.Get("report_date_from")); v != "" {
		t, err := parseDate(v)
		if err != nil {
			httputil.BadRequest(w, "invalid report_date_from (expected YYYY-MM-DD)")
			return
		}
		filter.ReportDateFrom = &t
	}
	if v := strings.TrimSpace(q.Get("report_date_to")); v != "" {
		t, err := parseDate(v)
		if err != nil {
			httputil.BadRequest(w, "invalid report_date_to (expected YYYY-MM-DD)")
			return
		}
		filter.ReportDateTo = &t
	}
	if v := strings.TrimSpace(q.Get("search")); v != "" {
		filter.Search = v
	}

	items, total, err := h.reports.List(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	out := make([]response.ResearchReportResponse, 0, len(items))
	for _, item := range items {
		row := response.FromResearchReport(item)
		h.enrichReviewStage(r.Context(), &row)
		out = append(out, row)
	}
	httputil.OK(w, response.ResearchReportListResponse{
		Items: out,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetResearchReport handles GET /investment/research-reports/{id}.
// @Summary Get Investment Research Report
// @Description Retrieve one research report by ID.
// @Tags Investment - Research
// @Security BearerAuth
// @Produce json
// @Param id path string true "Research report UUID"
// @Success 200 {object} response.ResearchReportResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /investment/research-reports/{id} [get]
func (h *ResearchReportHandler) GetResearchReport(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid report id")
		return
	}
	rep, err := h.reports.GetByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if rep == nil {
		httputil.NotFound(w, "research report not found")
		return
	}
	row := response.FromResearchReport(rep)
	h.enrichReviewStage(r.Context(), &row)
	httputil.OK(w, row)
}

// CreateResearchReport handles POST /investment/research-reports.
// @Summary Create Investment Research Report
// @Description Create a new DRAFT research report.
// @Tags Investment - Research
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.CreateResearchReportRequest true "Research report create payload"
// @Success 201 {object} response.ResearchReportResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Router /investment/research-reports [post]
func (h *ResearchReportHandler) CreateResearchReport(w http.ResponseWriter, r *http.Request) {
	var req request.CreateResearchReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	reportDate, err := parseDate(req.ReportDate)
	if err != nil {
		httputil.BadRequest(w, "invalid report_date (expected YYYY-MM-DD)")
		return
	}
	effectiveDate, err := parseDateOpt(req.EffectiveDate)
	if err != nil {
		httputil.BadRequest(w, "invalid effective_date (expected YYYY-MM-DD)")
		return
	}

	report, err := h.cmd.Create(r.Context(), command.CreateResearchReportRequest{
		ReportNo:             req.ReportNo,
		ReportDate:           reportDate,
		EffectiveDate:        effectiveDate,
		OwnerUserID:          req.OwnerUserID,
		AuthorUserID:         req.AuthorUserID,
		ApplicableContractID: req.ApplicableContractID,
		InstrumentType:       req.InstrumentType,
		InstrumentCode:       req.InstrumentCode,
		InstrumentName:       req.InstrumentName,
		Market:               req.Market,
		Currency:             req.Currency,
		Recommendation:       vo.Recommendation(req.Recommendation),
		ReportTitle:          req.ReportTitle,
		CompanyOverview:      req.CompanyOverview,
		CompanyOutlook:       req.CompanyOutlook,
		ESGComment:           req.ESGComment,
		FinancialStatus:      req.FinancialStatus,
		InvestmentAnalysis:   req.InvestmentAnalysis,
		ActorID:              actor,
	})
	if err != nil {
		writeResearchReportError(w, err)
		return
	}
	httputil.Created(w, response.FromResearchReport(report))
}

// UpdateResearchReport handles PUT /investment/research-reports/{id}.
// @Summary Update Investment Research Report
// @Description Apply partial updates to a research report. Refused when the report has been deleted or its review is completed.
// @Tags Investment - Research
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Research report UUID"
// @Param request body request.UpdateResearchReportRequest true "Research report update payload"
// @Success 200 {object} response.ResearchReportResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Router /investment/research-reports/{id} [put]
func (h *ResearchReportHandler) UpdateResearchReport(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid report id")
		return
	}
	var req request.UpdateResearchReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	cmdReq := command.UpdateResearchReportRequest{ReportID: id, ActorID: actor}

	if req.ReportDate != nil {
		t, err := parseDate(*req.ReportDate)
		if err != nil {
			httputil.BadRequest(w, "invalid report_date (expected YYYY-MM-DD)")
			return
		}
		cmdReq.ReportDate = &t
	}
	if req.EffectiveDate != nil && strings.TrimSpace(*req.EffectiveDate) != "" {
		t, err := parseDate(*req.EffectiveDate)
		if err != nil {
			httputil.BadRequest(w, "invalid effective_date (expected YYYY-MM-DD)")
			return
		}
		cmdReq.EffectiveDate = &t
	}
	cmdReq.OwnerUserID = req.OwnerUserID
	cmdReq.AuthorUserID = req.AuthorUserID
	cmdReq.ApplicableContractID = req.ApplicableContractID
	cmdReq.InstrumentType = req.InstrumentType
	cmdReq.InstrumentCode = req.InstrumentCode
	cmdReq.InstrumentName = req.InstrumentName
	cmdReq.Market = req.Market
	cmdReq.Currency = req.Currency
	if req.Recommendation != nil {
		rec := vo.Recommendation(*req.Recommendation)
		if !rec.IsValid() {
			httputil.BadRequest(w, "invalid recommendation")
			return
		}
		cmdReq.Recommendation = &rec
	}
	cmdReq.ReportTitle = req.ReportTitle
	cmdReq.CompanyOverview = req.CompanyOverview
	cmdReq.CompanyOutlook = req.CompanyOutlook
	cmdReq.ESGComment = req.ESGComment
	cmdReq.FinancialStatus = req.FinancialStatus
	cmdReq.InvestmentAnalysis = req.InvestmentAnalysis
	cmdReq.PostSubmissionNote = req.PostSubmissionNote

	report, err := h.cmd.Update(r.Context(), cmdReq)
	if err != nil {
		writeResearchReportError(w, err)
		return
	}
	httputil.OK(w, response.FromResearchReport(report))
}

// DeleteResearchReport handles DELETE /investment/research-reports/{id}.
// @Summary Delete Investment Research Report
// @Description Soft-delete a research report. Only allowed while review_status = NOT_SUBMITTED.
// @Tags Investment - Research
// @Security BearerAuth
// @Param id path string true "Research report UUID"
// @Success 204 "No Content"
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Router /investment/research-reports/{id} [delete]
func (h *ResearchReportHandler) DeleteResearchReport(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid report id")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	if err := h.cmd.SoftDelete(r.Context(), id, actor); err != nil {
		writeResearchReportError(w, err)
		return
	}
	httputil.NoContent(w)
}

// SubmitResearchReport handles POST /investment/research-reports/{id}/submit.
// @Summary Submit Investment Research Report
// @Description Move review_status from NOT_SUBMITTED to SUBMITTED. No real approval workflow is invoked.
// @Tags Investment - Research
// @Security BearerAuth
// @Produce json
// @Param id path string true "Research report UUID"
// @Success 200 {object} response.ResearchReportResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Router /investment/research-reports/{id}/submit [post]
func (h *ResearchReportHandler) SubmitResearchReport(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid report id")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	report, err := h.cmd.Submit(r.Context(), id, actor)
	if err != nil {
		writeResearchReportError(w, err)
		return
	}
	httputil.OK(w, response.FromResearchReport(report))
}

// CancelSubmitResearchReport handles POST /investment/research-reports/{id}/cancel-submit.
// @Summary Cancel Submission Of Investment Research Report
// @Description Move review_status from SUBMITTED back to NOT_SUBMITTED. Refused once review has been completed.
// @Tags Investment - Research
// @Security BearerAuth
// @Produce json
// @Param id path string true "Research report UUID"
// @Success 200 {object} response.ResearchReportResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Router /investment/research-reports/{id}/cancel-submit [post]
func (h *ResearchReportHandler) CancelSubmitResearchReport(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid report id")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	report, err := h.cmd.CancelSubmit(r.Context(), id, actor)
	if err != nil {
		writeResearchReportError(w, err)
		return
	}
	httputil.OK(w, response.FromResearchReport(report))
}

// InvalidateResearchReport handles POST /investment/research-reports/{id}/invalidate.
// @Summary Invalidate Investment Research Report
// @Description One-way transition to INVALIDATED. An invalidated report cannot be referenced by a decision, edited, submitted, cancelled, or soft-deleted. The reason (≥20 characters) is stored and audited.
// @Tags Investment - Research
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Research report UUID"
// @Param request body request.InvalidateResearchReportRequest true "Invalidation reason"
// @Success 200 {object} response.ResearchReportResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /investment/research-reports/{id}/invalidate [post]
func (h *ResearchReportHandler) InvalidateResearchReport(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid report id")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	var req request.InvalidateResearchReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	report, err := h.cmd.Invalidate(r.Context(), command.InvalidateResearchReportRequest{
		ReportID: id,
		ActorID:  actor,
		Reason:   req.Reason,
	})
	if err != nil {
		writeResearchReportError(w, err)
		return
	}
	httputil.OK(w, response.FromResearchReport(report))
}

// writeResearchReportError maps research-report domain errors to HTTP statuses.
func writeResearchReportError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var (
		invalidResearch *domain.ErrInvalidResearchReportRequest
		invalid         *domain.ErrInvalidDecisionRequest
		notFound        *domain.ErrResearchReportNotFound
		duplicate       *domain.ErrResearchReportNoAlreadyExists
		cannotUpdate    *domain.ErrResearchReportCannotUpdate
		cannotDelete    *domain.ErrResearchReportCannotDelete
		cannotSubmit    *domain.ErrResearchReportCannotSubmit
		cannotCancelSub *domain.ErrResearchReportCannotCancelSubmit
		cannotInvalid   *domain.ErrResearchReportCannotInvalidate
	)
	switch {
	case errors.As(err, &invalidResearch):
		httputil.BadRequest(w, err.Error())
	case errors.As(err, &invalid):
		// Retained for compatibility; new research validation paths emit
		// ErrInvalidResearchReportRequest. Should be unreachable once the
		// migration to the typed error is complete.
		httputil.BadRequest(w, err.Error())
	case errors.As(err, &notFound):
		httputil.NotFound(w, err.Error())
	case errors.As(err, &duplicate),
		errors.As(err, &cannotInvalid):
		httputil.Conflict(w, err.Error())
	case errors.As(err, &cannotUpdate),
		errors.As(err, &cannotDelete),
		errors.As(err, &cannotSubmit),
		errors.As(err, &cannotCancelSub):
		httputil.UnprocessableEntity(w, err.Error())
	default:
		httputil.InternalError(w, err.Error())
	}
}
