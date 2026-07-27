package handler

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// RuntimeHandler exposes the approval runtime endpoints.
type RuntimeHandler struct {
	svc *service.ApprovalRuntimeService
}

// NewRuntimeHandler wires the runtime handler.
func NewRuntimeHandler(svc *service.ApprovalRuntimeService) *RuntimeHandler {
	return &RuntimeHandler{svc: svc}
}

// GetInbox handles GET /approvals/inbox.
// @Summary List my approval inbox
// @Description Returns the authenticated user's pending approval tasks joined with their requests.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Produce json
// @Param status query string false "Task status filter (default PENDING)"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} response.InboxListResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /approvals/inbox [get]
func (h *RuntimeHandler) GetInbox(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	page, limit := pagination(r)
	f := domain.InboxFilter{
		UserID: actor,
		Status: vo.TaskStatus(r.URL.Query().Get("status")),
		Page:   page,
		Limit:  limit,
	}
	items, total, err := h.svc.GetMyInbox(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.InboxListResponse{
		Items: response.FromInboxItems(items), Total: total, Page: page, Limit: limit,
	})
}

// ListRequests handles GET /approvals/requests.
// @Summary List approval requests
// @Description Lists approval requests with optional filters.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Produce json
// @Param status query string false "Request status"
// @Param process_type query string false "Process type"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} response.RequestListResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /approvals/requests [get]
func (h *RuntimeHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	page, limit := pagination(r)
	f := domain.RequestListFilter{
		Status:      vo.RequestStatus(r.URL.Query().Get("status")),
		ProcessType: vo.ProcessType(r.URL.Query().Get("process_type")),
		ViewerID:    actor,
		Page:        page,
		Limit:       limit,
	}
	if cid := r.URL.Query().Get("contract_id"); cid != "" {
		id, err := uuid.Parse(cid)
		if err == nil {
			f.ContractID = &id
		}
	}
	items, total, err := h.svc.ListRequests(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.RequestListResponse{
		Items: response.FromRequests(items), Total: total, Page: page, Limit: limit,
	})
}

// GetRequest handles GET /approvals/requests/{requestId}.
// @Summary Get approval request detail
// @Description Returns a request with its tasks, immutable timeline and signature records.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Produce json
// @Param requestId path string true "Approval request UUID"
// @Success 200 {object} response.RequestDetailResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approvals/requests/{requestId} [get]
func (h *RuntimeHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "requestId")
	if err != nil {
		httputil.BadRequest(w, "invalid request id")
		return
	}
	detail, err := h.svc.GetApprovalRequest(r.Context(), id, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromRequestDetail(detail))
}

// GetTimeline handles GET /approvals/requests/{requestId}/timeline.
// @Summary Get approval timeline
// @Description Returns the immutable, ordered approval timeline for a request.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Produce json
// @Param requestId path string true "Approval request UUID"
// @Success 200 {array} response.EventResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /approvals/requests/{requestId}/timeline [get]
func (h *RuntimeHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "requestId")
	if err != nil {
		httputil.BadRequest(w, "invalid request id")
		return
	}
	events, err := h.svc.GetApprovalTimeline(r.Context(), id, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromEvents(events))
}

// GetSubjectStatus handles GET /approvals/subjects/{subjectType}/{subjectId}/status.
// @Summary Get subject approval status
// @Description Returns the latest approval request for a business object (subject), if any.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Produce json
// @Param subjectType path string true "Subject type"
// @Param subjectId path string true "Subject UUID"
// @Success 200 {object} response.SubjectStatusResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Router /approvals/subjects/{subjectType}/{subjectId}/status [get]
func (h *RuntimeHandler) GetSubjectStatus(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	subjectType := vo.SubjectType(pathParam(r, "subjectType"))
	subjectID, err := uuid.Parse(pathParam(r, "subjectId"))
	if err != nil {
		httputil.BadRequest(w, "invalid subject id")
		return
	}
	req, err := h.svc.GetSubjectApprovalStatus(r.Context(), subjectType, subjectID, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	resp := response.SubjectStatusResponse{HasRequest: req != nil, AllowedActions: []string{}}
	if req != nil {
		rr := response.FromRequest(req)
		resp.Request = &rr
		resp.AllowedActions = subjectAllowedActions(req, actor)
	}
	httputil.OK(w, resp)
}

// subjectAllowedActions returns viewer-specific top-level actions for a subject's
// latest request. Task-level actions (approve/reject) belong in request detail.
func subjectAllowedActions(req *entity.ApprovalRequest, viewerID uuid.UUID) []string {
	var actions []string
	if req.Status.IsTerminal() {
		if req.Status == vo.RequestStatusApproved {
			actions = append(actions, "revoke")
		}
		return actions
	}
	if req.SubmitterID == viewerID {
		actions = append(actions, "withdraw")
	}
	actions = append(actions, "cancel")
	return actions
}

// Submit handles POST /approvals/submit.
// @Summary Submit a subject for approval
// @Description Creates an approval request and the first stage tasks for a business object.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body request.SubmitRequest true "Submit payload"
// @Success 201 {object} response.RequestResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Router /approvals/submit [post]
func (h *RuntimeHandler) Submit(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	var body request.SubmitRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	subjectID, err := uuid.Parse(body.SubjectID)
	if err != nil {
		httputil.BadRequest(w, "invalid subject_id")
		return
	}
	contractID, err := optUUID(body.ContractID)
	if err != nil {
		httputil.BadRequest(w, "invalid contract_id")
		return
	}
	portfolioID, err := optUUID(body.PortfolioID)
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio_id")
		return
	}
	req, err := h.svc.SubmitApproval(r.Context(), service.SubmitInput{
		ProcessType:      vo.ProcessType(body.ProcessType),
		SubjectType:      vo.SubjectType(body.SubjectType),
		SubjectID:        subjectID,
		SubjectTitle:     body.SubjectTitle,
		SubjectReference: body.SubjectReference,
		ContractType:     vo.ContractType(body.ContractType),
		ContractID:       contractID,
		PortfolioID:      portfolioID,
		SubmitterID:      actor,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, response.FromRequest(req))
}

// Approve handles POST /approvals/tasks/{taskId}/approve.
// @Summary Approve an approval task
// @Description Records an approval on an assigned task and advances the request.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param taskId path string true "Approval task UUID"
// @Param payload body request.ActionRequest false "Optional comment"
// @Success 200 {object} response.RequestResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approvals/tasks/{taskId}/approve [post]
func (h *RuntimeHandler) Approve(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	taskID, err := parseUUIDParam(r, "taskId")
	if err != nil {
		httputil.BadRequest(w, "invalid task id")
		return
	}
	var body request.ActionRequest
	_ = decodeJSON(r, &body) // body is optional for approve
	req, err := h.svc.ApproveTask(r.Context(), taskID, actor, body.Comment)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromRequest(req))
}

// Reject handles POST /approvals/tasks/{taskId}/reject.
// @Summary Reject an approval task
// @Description Records a rejection (reason required); one rejection stops the request.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param taskId path string true "Approval task UUID"
// @Param payload body request.ActionRequest true "Rejection reason"
// @Success 200 {object} response.RequestResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approvals/tasks/{taskId}/reject [post]
func (h *RuntimeHandler) Reject(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	taskID, err := parseUUIDParam(r, "taskId")
	if err != nil {
		httputil.BadRequest(w, "invalid task id")
		return
	}
	var body request.ActionRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	req, err := h.svc.RejectTask(r.Context(), taskID, actor, body.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromRequest(req))
}

// Withdraw handles POST /approvals/requests/{requestId}/withdraw.
// @Summary Withdraw an approval request
// @Description The submitter withdraws their own active approval request.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Produce json
// @Param requestId path string true "Approval request UUID"
// @Success 200 {object} response.RequestResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approvals/requests/{requestId}/withdraw [post]
func (h *RuntimeHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "requestId")
	if err != nil {
		httputil.BadRequest(w, "invalid request id")
		return
	}
	req, err := h.svc.WithdrawRequest(r.Context(), id, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromRequest(req))
}

// Revoke handles POST /approvals/requests/{requestId}/revoke.
// @Summary Revoke an approved request
// @Description Revokes a previously approved request, returning the subject to an un-approved state. Requires a reason.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param requestId path string true "Approval request UUID"
// @Param payload body request.ActionRequest true "Revocation reason"
// @Success 200 {object} response.RequestResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approvals/requests/{requestId}/revoke [post]
func (h *RuntimeHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "requestId")
	if err != nil {
		httputil.BadRequest(w, "invalid request id")
		return
	}
	var body request.ActionRequest
	if err := decodeJSON(r, &body); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	req, err := h.svc.RevokeRequest(r.Context(), id, actor, body.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromRequest(req))
}

// Cancel handles POST /approvals/requests/{requestId}/cancel.
// @Summary Cancel an approval request
// @Description Cancels an in-flight approval request (privileged).
// @Tags Approval - Runtime
// @Security BearerAuth
// @Produce json
// @Param requestId path string true "Approval request UUID"
// @Success 200 {object} response.RequestResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approvals/requests/{requestId}/cancel [post]
func (h *RuntimeHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "requestId")
	if err != nil {
		httputil.BadRequest(w, "invalid request id")
		return
	}
	req, err := h.svc.CancelRequest(r.Context(), id, actor)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.FromRequest(req))
}

// ListSyncFailures handles GET /approvals/sync-failures.
// @Summary List approval subject-sync failures
// @Description Operator-facing inbox of approval decisions whose post-commit business-module callback (subject sync) failed. Never blocks or reverses the approval decision itself.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Produce json
// @Param status query string false "Status filter: PENDING, RESOLVED, or EXHAUSTED (default: all)"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} response.SyncFailureListResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /approvals/sync-failures [get]
func (h *RuntimeHandler) ListSyncFailures(w http.ResponseWriter, r *http.Request) {
	if _, ok := actorID(r); !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	page, limit := pagination(r)
	f := domain.SyncFailureFilter{
		Status: vo.SyncFailureStatus(r.URL.Query().Get("status")),
		Page:   page,
		Limit:  limit,
	}
	items, total, err := h.svc.ListSyncFailures(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, response.SyncFailureListResponse{
		Items: response.FromSyncFailures(items), Total: total, Page: page, Limit: limit,
	})
}

// RetrySyncFailure handles POST /approvals/sync-failures/{id}/retry.
// @Summary Retry an approval subject-sync failure
// @Description Re-invokes the failed business-module callback for a persisted sync-failure record. No-op (409) once the record is RESOLVED/EXHAUSTED or has no registered callback for its subject type. Bounded: each record has a fixed max_attempts.
// @Tags Approval - Runtime
// @Security BearerAuth
// @Produce json
// @Param id path string true "Sync failure record UUID"
// @Success 200 {object} response.SyncFailureResponse "Retry attempted; check status/last_error — a 200 does not by itself mean the retry succeeded, only that the attempt was recorded"
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /approvals/sync-failures/{id}/retry [post]
func (h *RuntimeHandler) RetrySyncFailure(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid sync failure id")
		return
	}
	rec, err := h.svc.RetrySyncFailure(r.Context(), id, actor)
	if err != nil && rec == nil {
		writeError(w, err)
		return
	}
	// rec != nil here even when err != nil (the retry attempt itself failed but
	// was durably recorded) — render the record either way so the operator sees
	// the actual outcome (status/last_error), never a raw callback error.
	httputil.OK(w, response.FromSyncFailure(rec))
}
