package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// DailyHandler serves the business-readable workflow endpoints.
// These endpoints accept businessDate only; workflow is global per date.
type DailyHandler struct {
	getDailyWorkflow *query.GetDailyWorkflowHandler
	settingRepo      domain.WorkflowApprovalSettingRepository
	logRepo          domain.TransitionLogRepository
	pool             *pgxpool.Pool

	openDay                 *command.OpenDayHandler
	approve                 *command.ManagerApprovalHandler
	cancelDayStart          *command.CancelDayStartHandler
	cancelApproval          *command.CancelApprovalHandler
	closeTransactions       *command.CloseTransactionsHandler
	cancelTransactionClose  *command.CancelTransactionCloseHandler
	closeAccounting         *command.CloseAccountingHandler
	rollbackAccountingClose *command.RollbackAccountingCloseHandler
}

// NewDailyHandler wires the handler.
func NewDailyHandler(
	getDailyWorkflow *query.GetDailyWorkflowHandler,
	settingRepo domain.WorkflowApprovalSettingRepository,
	logRepo domain.TransitionLogRepository,
	_ any,
	pool *pgxpool.Pool,
	openDay *command.OpenDayHandler,
	approve *command.ManagerApprovalHandler,
	cancelDayStart *command.CancelDayStartHandler,
	cancelApproval *command.CancelApprovalHandler,
	closeTransactions *command.CloseTransactionsHandler,
	cancelTransactionClose *command.CancelTransactionCloseHandler,
	closeAccounting *command.CloseAccountingHandler,
	rollbackAccountingClose *command.RollbackAccountingCloseHandler,
) *DailyHandler {
	return &DailyHandler{
		getDailyWorkflow:        getDailyWorkflow,
		settingRepo:             settingRepo,
		logRepo:                 logRepo,
		pool:                    pool,
		openDay:                 openDay,
		approve:                 approve,
		cancelDayStart:          cancelDayStart,
		cancelApproval:          cancelApproval,
		closeTransactions:       closeTransactions,
		cancelTransactionClose:  cancelTransactionClose,
		closeAccounting:         closeAccounting,
		rollbackAccountingClose: rollbackAccountingClose,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /workflow/daily?businessDate=YYYY-MM-DD
// ─────────────────────────────────────────────────────────────────────────────

// SetContractCatalog is retained as a no-op for legacy module wiring.
// Daily workflow APIs are date-only.
func (h *DailyHandler) SetContractCatalog(_ any) {}

// GetDailyWorkflow returns the aggregated daily workflow state.
// @Summary Get Daily Workflow State
// @Description Returns aggregated global workflow state for a business date.
// @Tags Workflow
// @Security BearerAuth
// @Produce json
// @Param businessDate query string true "Business date (YYYY-MM-DD)"
// @Success 200 {object} response.DailyWorkflowResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /workflow/daily [get]
func (h *DailyHandler) GetDailyWorkflow(w http.ResponseWriter, r *http.Request) {
	businessDate, ok := parseBusinessDateQuery(w, r)
	if !ok {
		return
	}

	result, err := h.getDailyWorkflow.Handle(r.Context(), query.DailyWorkflowRequest{
		BusinessDate: businessDate,
	})
	if err != nil {
		httputil.InternalError(w, "failed to read daily workflow state")
		return
	}

	httputil.OK(w, mapDailyWorkflowResponse(result))
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /workflow/daily/execute
// ─────────────────────────────────────────────────────────────────────────────

// ExecuteDailyTransition executes a workflow state transition via the daily API.
// @Summary Execute Daily Workflow Transition
// @Description Execute a workflow transition using business-readable identifiers and canonical operation types.
// @Tags Workflow
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.DailyExecuteRequest true "Execute payload"
// @Success 200 {object} response.DailyWorkflowResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /workflow/daily/execute [post]
func (h *DailyHandler) ExecuteDailyTransition(w http.ResponseWriter, r *http.Request) {
	var req request.DailyExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}

	businessDate, err := parseBusinessDate(req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, err.Error())
		return
	}

	internalAction, ok := vo.ParseAPIOperationType(req.OperationType)
	if !ok {
		httputil.BadRequest(w, fmt.Sprintf("unknown operationType: %q", req.OperationType))
		return
	}

	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		httputil.Unauthorized(w, "invalid actor id in token")
		return
	}

	isAdmin := slices.Contains(claims.Roles, "Admin")

	// Authorization: Admin bypasses configured-approver check.
	// Non-admin must be a configured approver for this operation.
	if !isAdmin {
		if denied := h.checkDailyPermission(r.Context(), w, req.OperationType, claims.Username); denied {
			return
		}
	}

	actor := vo.ActorContext{
		UserID:          userID,
		Username:        claims.Username,
		AccountCode:     claims.Username,
		ActorType:       vo.ActorTypeHuman,
		RequestID:       chimw.GetReqID(r.Context()),
		Roles:           claims.Roles,
		IsAdminOverride: isAdmin,
	}

	if err := h.runDailyTransition(r.Context(), uuid.Nil, businessDate, actor, internalAction, &req); err != nil {
		writeDailyCommandError(w, err)
		return
	}

	result, err := h.getDailyWorkflow.Handle(r.Context(), query.DailyWorkflowRequest{
		BusinessDate: businessDate,
	})
	if err != nil {
		httputil.InternalError(w, "failed to read daily workflow state")
		return
	}

	httputil.OK(w, mapDailyWorkflowResponse(result))
}

// checkDailyPermission verifies the caller is a configured approver for the given operationType.
// Returns true (denied) and writes the HTTP error when the check fails.
func (h *DailyHandler) checkDailyPermission(
	ctx context.Context,
	w http.ResponseWriter,
	operationType string,
	callerAccountCode string,
) bool {
	settings, err := h.settingRepo.ListByOperationType(ctx, operationType)
	if err != nil {
		httputil.InternalError(w, "failed to verify approver configuration")
		return true
	}
	for _, s := range settings {
		if s.ApproverAccountCode == callerAccountCode {
			return false
		}
	}
	httputil.Forbidden(w, fmt.Sprintf(
		"account %q is not a configured approver for operation %s",
		callerAccountCode, operationType,
	))
	return true
}

func (h *DailyHandler) runDailyTransition(
	ctx context.Context,
	contractID uuid.UUID,
	businessDate time.Time,
	actor vo.ActorContext,
	action vo.WorkflowAction,
	req *request.DailyExecuteRequest,
) error {
	reason := dailyReason(req)
	notes := req.Notes
	if notes == nil && req.Remark != "" {
		remark := req.Remark
		notes = &remark
	}

	switch action {
	case vo.ActionOpenDay:
		_, err := h.openDay.Handle(ctx, command.OpenDayRequest{
			ContractID:   contractID,
			BusinessDate: businessDate,
			Actor:        actor,
		})
		return err
	case vo.ActionApprove:
		_, err := h.approve.Handle(ctx, command.ManagerApprovalRequest{
			ContractID:                 contractID,
			BusinessDate:               businessDate,
			Actor:                      actor,
			ZeroTransactionAttestation: req.ZeroTransactionAttestation,
			AttestationReason:          req.AttestationReason,
			Notes:                      notes,
		})
		return err
	case vo.ActionCancelDayStart:
		_, err := h.cancelDayStart.Handle(ctx, command.CancelDayStartRequest{
			ContractID:   contractID,
			BusinessDate: businessDate,
			Actor:        actor,
			Reason:       reason,
		})
		return err
	case vo.ActionCancelApproval:
		_, err := h.cancelApproval.Handle(ctx, command.CancelApprovalRequest{
			ContractID:   contractID,
			BusinessDate: businessDate,
			Actor:        actor,
			Reason:       reason,
		})
		return err
	case vo.ActionCloseTransactions:
		_, err := h.closeTransactions.Handle(ctx, command.CloseTransactionsRequest{
			ContractID:   contractID,
			BusinessDate: businessDate,
			Actor:        actor,
		})
		return err
	case vo.ActionCancelTransactionClose:
		_, err := h.cancelTransactionClose.Handle(ctx, command.CancelTransactionCloseRequest{
			ContractID:   contractID,
			BusinessDate: businessDate,
			Actor:        actor,
			Reason:       reason,
		})
		return err
	case vo.ActionCloseAccounting:
		cmdReq := command.CloseAccountingRequest{
			ContractID:   contractID,
			BusinessDate: businessDate,
			Actor:        actor,
		}
		if req.AccountingDate != nil {
			ad, err := time.Parse("2006-01-02", *req.AccountingDate)
			if err != nil {
				return &domain.ErrInvalidTransition{
					Code:            "WORKFLOW_INVALID_ACCOUNTING_DATE",
					AttemptedAction: string(vo.ActionCloseAccounting),
					CurrentState:    "",
					Reason:          "invalid accountingDate (expected YYYY-MM-DD)",
				}
			}
			cmdReq.AccountingDate = ad
		}
		_, err := h.closeAccounting.Handle(ctx, cmdReq)
		return err
	case vo.ActionRollbackAccountingClose:
		_, err := h.rollbackAccountingClose.Handle(ctx, command.RollbackAccountingCloseRequest{
			ContractID:   contractID,
			BusinessDate: businessDate,
			Actor:        actor,
			Reason:       reason,
		})
		return err
	default:
		return &domain.ErrInvalidTransition{
			Code:            "WORKFLOW_UNSUPPORTED_OPERATION",
			AttemptedAction: string(action),
			CurrentState:    "",
			Reason:          fmt.Sprintf("unsupported operationType: %q", action),
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /workflow/daily/transitions?businessDate=&page=&pageSize=
// ─────────────────────────────────────────────────────────────────────────────

// GetDailyTransitions returns paginated transition history for a business date.
// @Summary Get Daily Transition History
// @Description Returns paginated global workflow transition history.
// @Tags Workflow
// @Security BearerAuth
// @Produce json
// @Param businessDate query string true "Business date (YYYY-MM-DD)"
// @Param page query int false "Page number (1-based, default 1)"
// @Param pageSize query int false "Page size (default 20, max 100)"
// @Success 200 {object} response.DailyTransitionsResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /workflow/daily/transitions [get]
func (h *DailyHandler) GetDailyTransitions(w http.ResponseWriter, r *http.Request) {
	businessDate, ok := parseBusinessDateQuery(w, r)
	if !ok {
		return
	}

	page := parseIntQueryDefault(r, "page", 1)
	pageSize := parseIntQueryDefault(r, "pageSize", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := h.logRepo.ListByBusinessDatePaginated(r.Context(), domain.TransitionLogPageRequest{
		BusinessDate: businessDate,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		httputil.InternalError(w, "failed to read transition history")
		return
	}

	entries := make([]response.DailyTimelineEntry, 0, len(result.Transitions))
	for _, t := range result.Transitions {
		entries = append(entries, response.DailyTimelineEntry{
			TransitionID:          t.ID.String(),
			OperationType:         t.Action.ToAPIName(),
			FromState:             t.FromState.ToAPIName(),
			ToState:               t.ToState.ToAPIName(),
			ExecutedByUsername:    t.ActorUsername,
			ExecutedByAccountCode: t.ActorAccountCode,
			IsAdminOverride:       t.IsAdminOverride,
			OccurredAt:            t.OccurredAt,
			Reason:                t.Reason,
		})
	}

	httputil.OK(w, response.DailyTransitionsResponse{
		BusinessDate: businessDate.Format("2006-01-02"),
		Total:        result.Total,
		Page:         result.Page,
		PageSize:     result.PageSize,
		Transitions:  entries,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /workflow/transition-rules
// ─────────────────────────────────────────────────────────────────────────────

// GetTransitionRules returns the static workflow state machine topology.
// @Summary Get Workflow Transition Rules
// @Description Returns the full state machine topology as a list of valid from→operation→to transitions.
// @Tags Workflow
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.TransitionRulesResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Router /workflow/transition-rules [get]
func (h *DailyHandler) GetTransitionRules(w http.ResponseWriter, r *http.Request) {
	httputil.OK(w, response.TransitionRulesResponse{Rules: workflowTransitionRules()})
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /workflow/settings
// ─────────────────────────────────────────────────────────────────────────────

// GetSettings returns all approver configuration for all operation types.
// @Summary Get Workflow Approval Settings
// @Description Returns the configured approvers for each workflow operation type.
// @Tags Workflow
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.WorkflowSettingsResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /workflow/settings [get]
func (h *DailyHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	all, err := h.settingRepo.ListAll(r.Context())
	if err != nil {
		httputil.InternalError(w, "failed to read workflow settings")
		return
	}
	httputil.OK(w, mapSettingsResponse(all))
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /workflow/settings
// ─────────────────────────────────────────────────────────────────────────────

// UpdateSettings replaces the approver list for one operation type.
// @Summary Update Workflow Approval Settings
// @Description Replace the configured approvers for a workflow operation type. Admin only. Note: approver account codes are stored without IAM validation (no cross-module IAM lookup interface exists).
// @Tags Workflow
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.DailySettingsUpdateRequest true "Settings payload"
// @Success 200 {object} response.WorkflowSettingsResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /workflow/settings [put]
func (h *DailyHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	if !slices.Contains(claims.Roles, "Admin") {
		httputil.Forbidden(w, "only Admin group members may update workflow settings")
		return
	}

	var req request.DailySettingsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body")
		return
	}
	if req.OperationType == "" {
		httputil.BadRequest(w, "operationType is required")
		return
	}
	if _, ok := vo.ParseAPIOperationType(req.OperationType); !ok {
		httputil.BadRequest(w, fmt.Sprintf("unknown operationType: %q", req.OperationType))
		return
	}

	newSettings := make([]*entity.WorkflowApprovalSetting, 0, len(req.Approvers))
	for _, a := range req.Approvers {
		if a.AccountCode == "" {
			httputil.BadRequest(w, "each approver must have a non-empty accountCode")
			return
		}
		newSettings = append(newSettings, &entity.WorkflowApprovalSetting{
			OperationType:       req.OperationType,
			ApprovalMode:        entity.ApprovalModeAnyOf,
			ApproverAccountCode: a.AccountCode,
			ApproverUsername:    a.Username,
			ApproverRole:        a.Role,
			IsActive:            true,
			UpdatedBy:           claims.Username,
		})
	}

	var txErr error
	if h.pool != nil {
		txErr = database.WithTransaction(r.Context(), h.pool, func(tx pgx.Tx) error {
			return h.settingRepo.Upsert(r.Context(), tx, req.OperationType, newSettings)
		})
	} else {
		txErr = h.settingRepo.Upsert(r.Context(), nil, req.OperationType, newSettings)
	}
	if txErr != nil {
		httputil.InternalError(w, "failed to update workflow settings")
		return
	}

	all, err := h.settingRepo.ListAll(r.Context())
	if err != nil {
		httputil.InternalError(w, "failed to read updated settings")
		return
	}
	httputil.OK(w, mapSettingsResponse(all))
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func parseIntQueryDefault(r *http.Request, key string, def int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 1 {
		return def
	}
	return v
}

func dailyReason(req *request.DailyExecuteRequest) string {
	if req == nil {
		return ""
	}
	if req.Reason != nil {
		return *req.Reason
	}
	return req.Remark
}

func writeDailyCommandError(w http.ResponseWriter, err error) {
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
			"businessDate":  postTradeBlocked.BusinessDate,
			"checkGroupId":  postTradeBlocked.CheckGroupID,
			"blockingCount": postTradeBlocked.BreachCount,
		})
	case errors.As(err, &cancelBlocked):
		httputil.JSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":            cancelBlocked.Error(),
			"code":             "WORKFLOW_CANCEL_BLOCKED_BY_TRANSACTIONS",
			"businessDate":     cancelBlocked.BusinessDate,
			"transactionCount": cancelBlocked.TransactionCount,
		})
	case errors.As(err, &notFound):
		httputil.JSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":        notFound.Error(),
			"code":         "WORKFLOW_DAY_NOT_STARTED",
			"businessDate": notFound.BusinessDate,
		})
	default:
		httputil.InternalError(w, "workflow command failed")
	}
}

func mapDailyWorkflowResponse(result *query.DailyWorkflowResult) response.DailyWorkflowResponse {
	timeline := make([]response.DailyTimelineEntry, 0, len(result.Timeline))
	for _, t := range result.Timeline {
		timeline = append(timeline, response.DailyTimelineEntry{
			TransitionID:          t.TransitionID,
			OperationType:         t.OperationType,
			FromState:             t.FromState,
			ToState:               t.ToState,
			ExecutedByUsername:    t.ExecutedByUsername,
			ExecutedByAccountCode: t.ExecutedByAccountCode,
			IsAdminOverride:       t.IsAdminOverride,
			OccurredAt:            t.OccurredAt,
			Reason:                t.Reason,
		})
	}

	blockedReasons := make([]response.DailyBlockingReason, 0, len(result.BlockedReasons))
	for _, br := range result.BlockedReasons {
		blockedReasons = append(blockedReasons, response.DailyBlockingReason{
			Code:    br.Code,
			Message: br.Message,
		})
	}

	approvers := make(map[string][]response.DailyApproverEntry)
	for op, entries := range result.Approvers {
		respEntries := make([]response.DailyApproverEntry, 0, len(entries))
		for _, e := range entries {
			respEntries = append(respEntries, response.DailyApproverEntry{
				AccountCode: e.AccountCode,
				Username:    e.Username,
				Role:        e.Role,
			})
		}
		approvers[op] = respEntries
	}

	settings := make(map[string]response.DailyOperationSetting)
	for op, s := range result.Settings {
		respApprovers := make([]response.DailyApproverEntry, 0, len(s.Approvers))
		for _, a := range s.Approvers {
			respApprovers = append(respApprovers, response.DailyApproverEntry{
				AccountCode: a.AccountCode,
				Username:    a.Username,
				Role:        a.Role,
			})
		}
		settings[op] = response.DailyOperationSetting{
			ApprovalMode: s.ApprovalMode,
			Approvers:    respApprovers,
		}
	}

	moduleReadiness := make(map[string]response.DailyModuleStatus)
	for k, v := range result.ModuleReadiness {
		moduleReadiness[k] = response.DailyModuleStatus{
			Ready:  v.Ready,
			Reason: v.Reason,
		}
	}

	var auditLastAt *time.Time
	if result.AuditSummary.LastTransitionAt != nil {
		t := *result.AuditSummary.LastTransitionAt
		auditLastAt = &t
	}

	return response.DailyWorkflowResponse{
		BusinessDate:      result.BusinessDate,
		CurrentState:      result.CurrentState,
		StateLabel:        workflowStateLabel(result.CurrentState),
		IsToday:           result.IsToday,
		Persisted:         result.Persisted,
		AllowedOperations: result.AllowedOperations,
		BlockedOperations: blockedOperations(result.AllowedOperations),
		BlockedReasons:    blockedReasons,
		TransitionRules:   workflowTransitionRules(),
		Timeline:          timeline,
		Approvers:         approvers,
		Settings:          settings,
		SettingsSummary:   settings,
		AuditSummary: response.DailyAuditSummary{
			TotalTransitions: result.AuditSummary.TotalTransitions,
			LastTransitionAt: auditLastAt,
			LastTransitionBy: result.AuditSummary.LastTransitionBy,
		},
		ModuleReadiness: moduleReadiness,
	}
}

func mapSettingsResponse(settings []*entity.WorkflowApprovalSetting) response.WorkflowSettingsResponse {
	byOp := make(map[string]*response.OperationSettingEntry)
	for _, s := range settings {
		if !s.IsActive {
			continue
		}
		e, ok := byOp[s.OperationType]
		if !ok {
			e = &response.OperationSettingEntry{
				OperationType: s.OperationType,
				ApprovalMode:  string(s.ApprovalMode),
				Approvers:     []response.DailyApproverEntry{},
			}
			byOp[s.OperationType] = e
		}
		e.Approvers = append(e.Approvers, response.DailyApproverEntry{
			AccountCode: s.ApproverAccountCode,
			Username:    s.ApproverUsername,
			Role:        s.ApproverRole,
		})
	}
	entries := make([]response.OperationSettingEntry, 0, len(byOp))
	for _, e := range byOp {
		entries = append(entries, *e)
	}
	return response.WorkflowSettingsResponse{Settings: entries}
}

func workflowStateLabel(state string) string {
	switch state {
	case string(vo.StateNotStarted):
		return "Not Started"
	case string(vo.StateInvestmentDayStarted):
		return "Investment Day Started"
	case string(vo.StateManagerApproved):
		return "Manager Approved"
	case string(vo.StateTransactionClosed):
		return "Transaction Closed"
	case string(vo.StateAccountingClosed):
		return "Accounting Closed"
	default:
		return state
	}
}

func allAPIOperations() []string {
	return []string{
		vo.APIOpStartInvestmentDay,
		vo.APIOpCancelInvestmentDay,
		vo.APIOpManagerApprove,
		vo.APIOpCancelManagerApproval,
		vo.APIOpCloseTransaction,
		vo.APIOpCancelTransactionClose,
		vo.APIOpCloseAccounting,
		vo.APIOpCancelAccountingClose,
	}
}

func blockedOperations(allowed []string) []string {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, op := range allowed {
		allowedSet[op] = struct{}{}
	}
	blocked := make([]string, 0, len(allAPIOperations()))
	for _, op := range allAPIOperations() {
		if _, ok := allowedSet[op]; !ok {
			blocked = append(blocked, op)
		}
	}
	return blocked
}

func workflowTransitionRules() []response.TransitionRule {
	return []response.TransitionRule{
		{FromState: "NOT_STARTED", OperationType: vo.APIOpStartInvestmentDay, ToState: "INVESTMENT_DAY_STARTED", RequiresReason: false},
		{FromState: "INVESTMENT_DAY_STARTED", OperationType: vo.APIOpCancelInvestmentDay, ToState: "NOT_STARTED", RequiresReason: true},
		{FromState: "INVESTMENT_DAY_STARTED", OperationType: vo.APIOpManagerApprove, ToState: "MANAGER_APPROVED", RequiresReason: false},
		{FromState: "MANAGER_APPROVED", OperationType: vo.APIOpCancelManagerApproval, ToState: "INVESTMENT_DAY_STARTED", RequiresReason: true},
		{FromState: "MANAGER_APPROVED", OperationType: vo.APIOpCloseTransaction, ToState: "TRANSACTION_CLOSED", RequiresReason: false},
		{FromState: "TRANSACTION_CLOSED", OperationType: vo.APIOpCancelTransactionClose, ToState: "MANAGER_APPROVED", RequiresReason: true},
		{FromState: "TRANSACTION_CLOSED", OperationType: vo.APIOpCloseAccounting, ToState: "ACCOUNTING_CLOSED", RequiresReason: false},
		{FromState: "ACCOUNTING_CLOSED", OperationType: vo.APIOpCancelAccountingClose, ToState: "TRANSACTION_CLOSED", RequiresReason: true},
	}
}
