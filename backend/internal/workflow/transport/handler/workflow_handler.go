package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	workflowperm "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/permission"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// WorkflowHandler bundles the HTTP endpoints for the workflow module.
type WorkflowHandler struct {
	getCurrentState *query.GetCurrentStateHandler
	getHistory      *query.GetHistoryHandler
	schedulerRunner SchedulerRunner

	openDay                 *command.OpenDayHandler
	approve                 *command.ManagerApprovalHandler
	cancelDayStart          *command.CancelDayStartHandler
	cancelApproval          *command.CancelApprovalHandler
	closeTransactions       *command.CloseTransactionsHandler
	cancelTransactionClose  *command.CancelTransactionCloseHandler
	closeAccounting         *command.CloseAccountingHandler
	rollbackAccountingClose *command.RollbackAccountingCloseHandler

	permissionChecker middleware.PermissionChecker
}

// SchedulerRunner is implemented by jobs.DayScheduler.
type SchedulerRunner interface {
	RunOnce(rctx context.Context) ([]*entity.SchedulerRun, error)
}

// NewWorkflowHandler wires every command handler and the permission checker.
func NewWorkflowHandler(
	getCurrentState *query.GetCurrentStateHandler,
	getHistory *query.GetHistoryHandler,
	schedulerRunner SchedulerRunner,
	openDay *command.OpenDayHandler,
	approve *command.ManagerApprovalHandler,
	cancelDayStart *command.CancelDayStartHandler,
	cancelApproval *command.CancelApprovalHandler,
	closeTransactions *command.CloseTransactionsHandler,
	cancelTransactionClose *command.CancelTransactionCloseHandler,
	closeAccounting *command.CloseAccountingHandler,
	rollbackAccountingClose *command.RollbackAccountingCloseHandler,
	permChecker middleware.PermissionChecker,
) *WorkflowHandler {
	return &WorkflowHandler{
		getCurrentState:         getCurrentState,
		getHistory:              getHistory,
		schedulerRunner:         schedulerRunner,
		openDay:                 openDay,
		approve:                 approve,
		cancelDayStart:          cancelDayStart,
		cancelApproval:          cancelApproval,
		closeTransactions:       closeTransactions,
		cancelTransactionClose:  cancelTransactionClose,
		closeAccounting:         closeAccounting,
		rollbackAccountingClose: rollbackAccountingClose,
		permissionChecker:       permChecker,
	}
}

// RunSchedulerOnce manually executes one workflow scheduler tick.
// @Summary Run Workflow Scheduler Once
// @Description Manually execute one workflow day scheduler tick.
// @Tags Workflow
// @Security BearerAuth
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /workflow/scheduler/run-once [post]
func (h *WorkflowHandler) RunSchedulerOnce(w http.ResponseWriter, r *http.Request) {
	if h.schedulerRunner == nil {
		httputil.InternalError(w, "workflow scheduler is not configured")
		return
	}

	runs, err := h.schedulerRunner.RunOnce(r.Context())
	if err != nil {
		httputil.JSON(w, http.StatusInternalServerError, map[string]any{
			"error": "workflow scheduler run failed",
			"runs":  mapSchedulerRunResponses(runs),
		})
		return
	}

	httputil.OK(w, mapSchedulerRunResponses(runs))
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /workflow/day-states/{contractId}?businessDate=YYYY-MM-DD
// ─────────────────────────────────────────────────────────────────────────────

// GetCurrentState handles GET /workflow/day-states/{contractId}.
// @Summary Get Workflow Day State
// @Description Get the current workflow state for a contract and business date.
// @Tags Workflow
// @Security BearerAuth
// @Produce json
// @Param contractId path string true "Contract UUID"
// @Param businessDate query string true "Business date (YYYY-MM-DD)"
// @Success 200 {object} response.WorkflowStateResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /workflow/day-states/{contractId} [get]
func (h *WorkflowHandler) GetCurrentState(w http.ResponseWriter, r *http.Request) {
	contractID, ok := parseContractID(w, r)
	if !ok {
		return
	}
	businessDate, ok := parseBusinessDateQuery(w, r)
	if !ok {
		return
	}

	result, err := h.getCurrentState.Handle(r.Context(), query.GetCurrentStateRequest{
		ContractID:   contractID,
		BusinessDate: businessDate,
	})
	if err != nil {
		httputil.InternalError(w, "failed to read workflow state")
		return
	}

	httputil.OK(w, mapStateResponse(contractID, businessDate, result))
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /workflow/day-states/{contractId}/history?businessDate=YYYY-MM-DD
// ─────────────────────────────────────────────────────────────────────────────

// GetHistory handles GET /workflow/day-states/{contractId}/history.
// @Summary Get Workflow Transition History
// @Description Get transition history for a contract and business date.
// @Tags Workflow
// @Security BearerAuth
// @Produce json
// @Param contractId path string true "Contract UUID"
// @Param businessDate query string true "Business date (YYYY-MM-DD)"
// @Success 200 {object} response.HistoryResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /workflow/day-states/{contractId}/history [get]
func (h *WorkflowHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	contractID, ok := parseContractID(w, r)
	if !ok {
		return
	}
	businessDate, ok := parseBusinessDateQuery(w, r)
	if !ok {
		return
	}

	result, err := h.getHistory.Handle(r.Context(), query.GetHistoryRequest{
		ContractID:   contractID,
		BusinessDate: businessDate,
	})
	if err != nil {
		httputil.InternalError(w, "failed to read workflow history")
		return
	}

	httputil.OK(w, mapHistoryResponse(result.ContractID, result.BusinessDate, result.Transitions))
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /workflow/day-states/{contractId}/transitions
// ─────────────────────────────────────────────────────────────────────────────

// ExecuteTransition validates the request, resolves the per-action permission,
// then dispatches to the matching command handler.
// @Summary Execute Workflow Transition
// @Description Execute a workflow transition such as OPEN_DAY, APPROVE, CLOSE_TRANSACTIONS, or ROLLBACK_ACCOUNTING_CLOSE.
// @Tags Workflow
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param contractId path string true "Contract UUID"
// @Param request body request.ExecuteTransitionRequest true "Transition payload"
// @Success 201 {object} response.TransitionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /workflow/day-states/{contractId}/transitions [post]
func (h *WorkflowHandler) ExecuteTransition(w http.ResponseWriter, r *http.Request) {
	contractID, ok := parseContractID(w, r)
	if !ok {
		return
	}

	var req request.ExecuteTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}

	businessDate, err := parseBusinessDate(req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

	action := vo.WorkflowAction(req.Action)
	if !action.IsValid() {
		httputil.BadRequest(w, fmt.Sprintf("unknown action: %q", req.Action))
		return
	}

	actor, ok := buildActorContext(w, r)
	if !ok {
		return
	}

	if !h.checkActionPermission(w, r, actor, action) {
		return
	}

	switch action {
	case vo.ActionOpenDay:
		h.handleOpenDay(w, r, contractID, businessDate, actor)
	case vo.ActionApprove:
		h.handleApprove(w, r, contractID, businessDate, actor, &req)
	case vo.ActionCancelDayStart:
		h.handleCancelDayStart(w, r, contractID, businessDate, actor, &req)
	case vo.ActionCancelApproval:
		h.handleCancelApproval(w, r, contractID, businessDate, actor, &req)
	case vo.ActionCloseTransactions:
		h.handleCloseTransactions(w, r, contractID, businessDate, actor)
	case vo.ActionCancelTransactionClose:
		h.handleCancelTransactionClose(w, r, contractID, businessDate, actor, &req)
	case vo.ActionCloseAccounting:
		h.handleCloseAccounting(w, r, contractID, businessDate, actor, &req)
	case vo.ActionRollbackAccountingClose:
		h.handleRollbackAccountingClose(w, r, contractID, businessDate, actor, &req)
	default:
		httputil.BadRequest(w, fmt.Sprintf("unsupported action: %q", req.Action))
	}
}

// actionPermissionCode returns the permission code required for the given action.
func actionPermissionCode(a vo.WorkflowAction) string {
	switch a {
	case vo.ActionOpenDay:
		return workflowperm.CodeOpenDay
	case vo.ActionApprove:
		return workflowperm.CodeApprove
	case vo.ActionCancelDayStart:
		return workflowperm.CodeCancelDayStart
	case vo.ActionCancelApproval:
		return workflowperm.CodeCancelApproval
	case vo.ActionCloseTransactions:
		return workflowperm.CodeCloseTransactions
	case vo.ActionCancelTransactionClose:
		return workflowperm.CodeCancelTransactionClose
	case vo.ActionCloseAccounting:
		return workflowperm.CodeCloseAccounting
	case vo.ActionRollbackAccountingClose:
		return workflowperm.CodeRollbackAccountingClose
	}
	return ""
}

// checkActionPermission enforces the per-action permission code before dispatch.
// Returns true when the call may proceed; writes the HTTP error and returns false
// otherwise.
func (h *WorkflowHandler) checkActionPermission(
	w http.ResponseWriter,
	r *http.Request,
	actor vo.ActorContext,
	action vo.WorkflowAction,
) bool {
	if h.permissionChecker == nil {
		return true
	}
	code := actionPermissionCode(action)
	if code == "" {
		httputil.BadRequest(w, fmt.Sprintf("no permission mapping for action %q", action))
		return false
	}
	ok, err := h.permissionChecker.HasFunctionPermission(r.Context(), actor.UserID.String(), code)
	if err != nil {
		httputil.InternalError(w, "failed to verify permissions")
		return false
	}
	if !ok {
		httputil.Forbidden(w, fmt.Sprintf("missing permission %s for action %s", code, action))
		return false
	}
	return true
}

// ─────────────────────────────────────────────────────────────────────────────
// Per-action handlers
// ─────────────────────────────────────────────────────────────────────────────

func (h *WorkflowHandler) handleOpenDay(
	w http.ResponseWriter, r *http.Request,
	contractID uuid.UUID, businessDate time.Time, actor vo.ActorContext,
) {
	result, err := h.openDay.Handle(r.Context(), command.OpenDayRequest{
		ContractID:   contractID,
		BusinessDate: businessDate,
		Actor:        actor,
	})
	if err != nil {
		writeCommandError(w, err)
		return
	}
	httputil.Created(w, response.TransitionResponse{
		TransitionID:  result.TransitionID.String(),
		WorkflowDayID: result.WorkflowDayID.String(),
		ContractID:    result.ContractID.String(),
		BusinessDate:  result.BusinessDate.Format("2006-01-02"),
		FromState:     string(result.FromState),
		ToState:       string(result.ToState),
		OccurredAt:    result.OccurredAt.Format(time.RFC3339),
	})
}

func (h *WorkflowHandler) handleApprove(
	w http.ResponseWriter, r *http.Request,
	contractID uuid.UUID, businessDate time.Time, actor vo.ActorContext,
	req *request.ExecuteTransitionRequest,
) {
	result, err := h.approve.Handle(r.Context(), command.ManagerApprovalRequest{
		ContractID:                 contractID,
		BusinessDate:               businessDate,
		Actor:                      actor,
		ZeroTransactionAttestation: req.ZeroTransactionAttestation,
		AttestationReason:          req.AttestationReason,
		Notes:                      req.Notes,
	})
	if err != nil {
		writeCommandError(w, err)
		return
	}
	approvalID := result.ApprovalID.String()
	isZero := result.IsZeroTransaction
	httputil.Created(w, response.TransitionResponse{
		TransitionID:      result.TransitionID.String(),
		WorkflowDayID:     result.WorkflowDayID.String(),
		ContractID:        result.ContractID.String(),
		BusinessDate:      result.BusinessDate.Format("2006-01-02"),
		FromState:         string(result.FromState),
		ToState:           string(result.ToState),
		OccurredAt:        result.OccurredAt.Format(time.RFC3339),
		ApprovalID:        &approvalID,
		IsZeroTransaction: &isZero,
	})
}

func (h *WorkflowHandler) handleCancelDayStart(
	w http.ResponseWriter, r *http.Request,
	contractID uuid.UUID, businessDate time.Time, actor vo.ActorContext,
	req *request.ExecuteTransitionRequest,
) {
	reason := stringOrEmpty(req.Reason)
	result, err := h.cancelDayStart.Handle(r.Context(), command.CancelDayStartRequest{
		ContractID:   contractID,
		BusinessDate: businessDate,
		Actor:        actor,
		Reason:       reason,
	})
	if err != nil {
		writeCommandError(w, err)
		return
	}
	httputil.Created(w, response.TransitionResponse{
		TransitionID:  result.TransitionID.String(),
		WorkflowDayID: result.WorkflowDayID.String(),
		ContractID:    result.ContractID.String(),
		BusinessDate:  result.BusinessDate.Format("2006-01-02"),
		FromState:     string(result.FromState),
		ToState:       string(result.ToState),
		OccurredAt:    result.OccurredAt.Format(time.RFC3339),
	})
}

func (h *WorkflowHandler) handleCancelApproval(
	w http.ResponseWriter, r *http.Request,
	contractID uuid.UUID, businessDate time.Time, actor vo.ActorContext,
	req *request.ExecuteTransitionRequest,
) {
	reason := stringOrEmpty(req.Reason)
	result, err := h.cancelApproval.Handle(r.Context(), command.CancelApprovalRequest{
		ContractID:   contractID,
		BusinessDate: businessDate,
		Actor:        actor,
		Reason:       reason,
	})
	if err != nil {
		writeCommandError(w, err)
		return
	}
	httputil.Created(w, response.TransitionResponse{
		TransitionID:  result.TransitionID.String(),
		WorkflowDayID: result.WorkflowDayID.String(),
		ContractID:    result.ContractID.String(),
		BusinessDate:  result.BusinessDate.Format("2006-01-02"),
		FromState:     string(result.FromState),
		ToState:       string(result.ToState),
		OccurredAt:    result.OccurredAt.Format(time.RFC3339),
	})
}

func (h *WorkflowHandler) handleCloseTransactions(
	w http.ResponseWriter, r *http.Request,
	contractID uuid.UUID, businessDate time.Time, actor vo.ActorContext,
) {
	result, err := h.closeTransactions.Handle(r.Context(), command.CloseTransactionsRequest{
		ContractID:   contractID,
		BusinessDate: businessDate,
		Actor:        actor,
	})
	if err != nil {
		writeCommandError(w, err)
		return
	}
	httputil.Created(w, response.TransitionResponse{
		TransitionID:  result.TransitionID.String(),
		WorkflowDayID: result.WorkflowDayID.String(),
		ContractID:    result.ContractID.String(),
		BusinessDate:  result.BusinessDate.Format("2006-01-02"),
		FromState:     string(result.FromState),
		ToState:       string(result.ToState),
		OccurredAt:    result.OccurredAt.Format(time.RFC3339),
	})
}

func (h *WorkflowHandler) handleCancelTransactionClose(
	w http.ResponseWriter, r *http.Request,
	contractID uuid.UUID, businessDate time.Time, actor vo.ActorContext,
	req *request.ExecuteTransitionRequest,
) {
	reason := stringOrEmpty(req.Reason)
	result, err := h.cancelTransactionClose.Handle(r.Context(), command.CancelTransactionCloseRequest{
		ContractID:   contractID,
		BusinessDate: businessDate,
		Actor:        actor,
		Reason:       reason,
	})
	if err != nil {
		writeCommandError(w, err)
		return
	}
	httputil.Created(w, response.TransitionResponse{
		TransitionID:  result.TransitionID.String(),
		WorkflowDayID: result.WorkflowDayID.String(),
		ContractID:    result.ContractID.String(),
		BusinessDate:  result.BusinessDate.Format("2006-01-02"),
		FromState:     string(result.FromState),
		ToState:       string(result.ToState),
		OccurredAt:    result.OccurredAt.Format(time.RFC3339),
	})
}

func (h *WorkflowHandler) handleCloseAccounting(
	w http.ResponseWriter, r *http.Request,
	contractID uuid.UUID, businessDate time.Time, actor vo.ActorContext,
	req *request.ExecuteTransitionRequest,
) {
	cmdReq := command.CloseAccountingRequest{
		ContractID:   contractID,
		BusinessDate: businessDate,
		Actor:        actor,
	}
	if req != nil && req.AccountingDate != nil {
		ad, err := time.Parse("2006-01-02", *req.AccountingDate)
		if err != nil {
			httputil.BadRequest(w, "invalid accountingDate (expected YYYY-MM-DD)")
			return
		}
		cmdReq.AccountingDate = ad
	}
	result, err := h.closeAccounting.Handle(r.Context(), cmdReq)
	if err != nil {
		writeCommandError(w, err)
		return
	}
	resp := response.TransitionResponse{
		TransitionID:  result.TransitionID.String(),
		WorkflowDayID: result.WorkflowDayID.String(),
		ContractID:    result.ContractID.String(),
		BusinessDate:  result.BusinessDate.Format("2006-01-02"),
		FromState:     string(result.FromState),
		ToState:       string(result.ToState),
		OccurredAt:    result.OccurredAt.Format(time.RFC3339),
	}
	resp.AccountingDate = result.AccountingDate.Format("2006-01-02")
	httputil.Created(w, resp)
}

func (h *WorkflowHandler) handleRollbackAccountingClose(
	w http.ResponseWriter, r *http.Request,
	contractID uuid.UUID, businessDate time.Time, actor vo.ActorContext,
	req *request.ExecuteTransitionRequest,
) {
	reason := stringOrEmpty(req.Reason)
	result, err := h.rollbackAccountingClose.Handle(r.Context(), command.RollbackAccountingCloseRequest{
		ContractID:   contractID,
		BusinessDate: businessDate,
		Actor:        actor,
		Reason:       reason,
	})
	if err != nil {
		writeCommandError(w, err)
		return
	}
	httputil.Created(w, response.TransitionResponse{
		TransitionID:  result.TransitionID.String(),
		WorkflowDayID: result.WorkflowDayID.String(),
		ContractID:    result.ContractID.String(),
		BusinessDate:  result.BusinessDate.Format("2006-01-02"),
		FromState:     string(result.FromState),
		ToState:       string(result.ToState),
		OccurredAt:    result.OccurredAt.Format(time.RFC3339),
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Parsing & error mapping helpers
// ─────────────────────────────────────────────────────────────────────────────

func parseContractID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	raw := chi.URLParam(r, "contractId")
	id, err := uuid.Parse(raw)
	if err != nil {
		httputil.BadRequest(w, "invalid contractId: must be a UUID")
		return uuid.Nil, false
	}
	return id, true
}

func parseBusinessDateQuery(w http.ResponseWriter, r *http.Request) (time.Time, bool) {
	raw := r.URL.Query().Get("businessDate")
	if raw == "" {
		httputil.BadRequest(w, "businessDate query parameter is required (format YYYY-MM-DD)")
		return time.Time{}, false
	}
	t, err := parseBusinessDate(raw)
	if err != nil {
		httputil.BadRequest(w, err.Error())
		return time.Time{}, false
	}
	return t, true
}

// parseBusinessDate parses a YYYY-MM-DD string in Asia/Bangkok so pgx stores
// the correct calendar date in a PostgreSQL DATE column regardless of server TZ.
func parseBusinessDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("businessDate is required")
	}
	t, err := time.ParseInLocation("2006-01-02", s, clock.BangkokLocation)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid businessDate format, expected YYYY-MM-DD: %w", err)
	}
	return t, nil
}

func buildActorContext(w http.ResponseWriter, r *http.Request) (vo.ActorContext, bool) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "not authenticated")
		return vo.ActorContext{}, false
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		httputil.Unauthorized(w, "invalid actor id in token")
		return vo.ActorContext{}, false
	}
	return vo.ActorContext{
		UserID:      userID,
		Username:    claims.Username,
		AccountCode: claims.Username,
		ActorType:   vo.ActorTypeHuman,
		RequestID:   chimw.GetReqID(r.Context()),
		Roles:       claims.Roles,
	}, true
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func writeCommandError(w http.ResponseWriter, err error) {
	var invalidTx *domain.ErrInvalidTransition
	var dayExists *domain.ErrWorkflowDayExists
	var versionConflict *domain.ErrVersionConflict
	var notFound *domain.ErrNotFound
	var cancelBlocked *domain.ErrCancelBlockedByTransactions
	var postTradeBlocked *domain.ErrPostTradeBreachesBlockClose
	var selfApproval *domain.ErrSelfApprovalForbidden

	switch {
	case errors.As(err, &selfApproval):
		httputil.JSON(w, http.StatusForbidden, map[string]any{
			"error":        selfApproval.Error(),
			"code":         "WORKFLOW_SELF_APPROVAL_FORBIDDEN",
			"contractId":   selfApproval.ContractID,
			"businessDate": selfApproval.BusinessDate,
			"actorId":      selfApproval.ActorID,
		})
	case errors.As(err, &invalidTx):
		httputil.JSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":           invalidTx.Reason,
			"code":            invalidTx.Code,
			"currentState":    invalidTx.CurrentState,
			"attemptedAction": invalidTx.AttemptedAction,
			"details":         invalidTx.Details,
		})
	case errors.As(err, &dayExists):
		httputil.JSON(w, http.StatusConflict, map[string]any{
			"error":        dayExists.Error(),
			"code":         "WORKFLOW_DAY_ALREADY_EXISTS",
			"contractId":   dayExists.ContractID,
			"businessDate": dayExists.BusinessDate,
			"currentState": dayExists.CurrentState,
		})
	case errors.As(err, &versionConflict):
		httputil.JSON(w, http.StatusConflict, httputil.ErrorResponse{
			Error: versionConflict.Error(),
			Code:  "WORKFLOW_VERSION_CONFLICT",
		})
	case errors.As(err, &postTradeBlocked):
		httputil.JSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":         postTradeBlocked.Error(),
			"code":          "WORKFLOW_CLOSE_BLOCKED_BY_IRG_BREACHES",
			"contractId":    postTradeBlocked.ContractID,
			"businessDate":  postTradeBlocked.BusinessDate,
			"checkGroupId":  postTradeBlocked.CheckGroupID,
			"blockingCount": postTradeBlocked.BreachCount,
		})
	case errors.As(err, &cancelBlocked):
		httputil.JSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":            cancelBlocked.Error(),
			"code":             "WORKFLOW_CANCEL_BLOCKED_BY_TRANSACTIONS",
			"contractId":       cancelBlocked.ContractID,
			"businessDate":     cancelBlocked.BusinessDate,
			"transactionCount": cancelBlocked.TransactionCount,
		})
	case errors.As(err, &notFound):
		httputil.JSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":        notFound.Error(),
			"code":         "WORKFLOW_DAY_NOT_STARTED",
			"contractId":   notFound.ContractID,
			"businessDate": notFound.BusinessDate,
		})
	default:
		httputil.InternalError(w, "workflow command failed")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Response mapping
// ─────────────────────────────────────────────────────────────────────────────

func mapStateResponse(
	contractID uuid.UUID,
	businessDate time.Time,
	result *query.GetCurrentStateResult,
) response.WorkflowStateResponse {
	actions := make([]string, 0, len(result.AllowedActions))
	for _, a := range result.AllowedActions {
		actions = append(actions, string(a))
	}

	resp := response.WorkflowStateResponse{
		ContractID:     contractID.String(),
		BusinessDate:   businessDate.Format("2006-01-02"),
		AllowedActions: actions,
	}

	if result.Persisted && result.Day != nil {
		d := result.Day
		ver := d.Version
		resp.CurrentState = string(d.CurrentState)
		resp.Persisted = true
		resp.TransactionsLocked = d.IsTransactionLocked()
		resp.PendingReclose = d.PendingReclose
		resp.RecloseCount = d.RecloseCount
		resp.OpenedAt = d.OpenedAt
		resp.OpenedBy = uuidPtrString(d.OpenedBy)
		resp.ManagerApprovedAt = d.ManagerApprovedAt
		resp.ManagerApprovedBy = uuidPtrString(d.ManagerApprovedBy)
		resp.TransactionClosedAt = d.TransactionClosedAt
		resp.AccountingClosedAt = d.AccountingClosedAt
		if d.AccountingDate != nil {
			s := d.AccountingDate.Format("2006-01-02")
			resp.AccountingDate = &s
		}
		if d.PrevAccountingDate != nil {
			s := d.PrevAccountingDate.Format("2006-01-02")
			resp.PrevAccountingDate = &s
		}
		resp.Version = &ver
		return resp
	}

	resp.CurrentState = string(vo.StateNotStarted)
	resp.Persisted = false

	prevDayStatus := response.PreviousDayStatus{
		BusinessDate: result.PreviousBusinessDate.Format("2006-01-02"),
		CurrentState: string(vo.StateNotStarted),
		Persisted:    false,
	}
	if result.PrevDay != nil {
		prevDayStatus.CurrentState = string(result.PrevDay.CurrentState)
		prevDayStatus.Persisted = true

		switch result.PrevDay.CurrentState {
		case vo.StateDayOpen:
			resp.BlockingReasons = append(resp.BlockingReasons, response.BlockingReason{
				Code: "WORKFLOW_PREVIOUS_DAY_NOT_APPROVED",
				Message: fmt.Sprintf(
					"previous business day %s is still in DAY_OPEN state",
					result.PreviousBusinessDate.Format("2006-01-02"),
				),
			})
			resp.AllowedActions = []string{}
		case vo.StateManagerApproved:
			resp.BlockingReasons = append(resp.BlockingReasons, response.BlockingReason{
				Code: "WORKFLOW_PREVIOUS_DAY_NOT_TRANSACTION_CLOSED",
				Message: fmt.Sprintf(
					"previous business day %s is in MANAGER_APPROVED state",
					result.PreviousBusinessDate.Format("2006-01-02"),
				),
			})
			resp.AllowedActions = []string{}
		}
	}
	resp.PreviousDay = &prevDayStatus

	return resp
}

func mapHistoryResponse(
	contractID uuid.UUID,
	businessDate time.Time,
	transitions []*entity.WorkflowTransition,
) response.HistoryResponse {
	entries := make([]response.TransitionEntry, 0, len(transitions))
	for _, t := range transitions {
		entries = append(entries, response.TransitionEntry{
			ID:            t.ID.String(),
			FromState:     string(t.FromState),
			ToState:       string(t.ToState),
			Action:        string(t.Action),
			ActorType:     string(t.ActorType),
			ActorID:       uuidPtrString(t.ActorID),
			ActorUsername: t.ActorUsername,
			Reason:        t.Reason,
			OccurredAt:    t.OccurredAt,
			RequestID:     t.RequestID,
		})
	}
	return response.HistoryResponse{
		ContractID:   contractID.String(),
		BusinessDate: businessDate.Format("2006-01-02"),
		Transitions:  entries,
	}
}

func mapSchedulerRunResponses(runs []*entity.SchedulerRun) []map[string]any {
	responses := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		if run == nil {
			continue
		}
		item := map[string]any{
			"id":            run.ID.String(),
			"schedulerName": run.SchedulerName,
			"businessDate":  run.BusinessDate.Format("2006-01-02"),
			"timezone":      run.Timezone,
			"startedAt":     run.StartedAt,
			"finishedAt":    run.FinishedAt,
			"status":        string(run.Status),
			"triggeredBy":   run.TriggeredBy,
			"lockedBy":      run.LockedBy,
			"summary":       run.Summary,
			"errorMessage":  run.ErrorMessage,
			"createdAt":     run.CreatedAt,
		}
		if run.RuleID != nil {
			item["ruleId"] = run.RuleID.String()
		}
		if run.Action != nil {
			item["action"] = string(*run.Action)
		}
		responses = append(responses, item)
	}
	return responses
}

func uuidPtrString(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}
