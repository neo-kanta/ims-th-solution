package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appsvc "github.com/neo-kanta/ims-th-solution/backend/internal/permissions/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/permissions/domain"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

type Handler struct {
	service *appsvc.Service
}

func NewHandler(service *appsvc.Service) *Handler {
	return &Handler{service: service}
}

type changeRequestPayload struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	RequestType      string `json:"request_type"`
	RiskLevel        string `json:"risk_level"`
	TargetEntityType string `json:"target_entity_type"`
	TargetEntityID   string `json:"target_entity_id"`
}

type changeItemPayload struct {
	ItemType    string          `json:"item_type"`
	TargetTable string          `json:"target_table"`
	TargetID    string          `json:"target_id"`
	ActionType  string          `json:"action_type"`
	BeforeJSON  json.RawMessage `json:"before_json"`
	AfterJSON   json.RawMessage `json:"after_json"`
}

type decisionPayload struct {
	Comment string `json:"comment"`
}

type commentPayload struct {
	Comment string `json:"comment"`
}

type labelPayload struct {
	LabelCode string `json:"label_code"`
}

type roleAssignmentPayload struct {
	RoleID   string `json:"role_id"`
	RoleCode string `json:"role_code"`
	Reason   string `json:"reason"`
}

type notificationSettingsPayload struct {
	Settings []domain.NotificationSetting `json:"settings"`
}

func (h *Handler) ListChangeRequests(w http.ResponseWriter, r *http.Request) {
	filter := domain.ChangeRequestFilter{
		Status:    strings.TrimSpace(r.URL.Query().Get("status")),
		RiskLevel: strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("risk"))),
		Label:     strings.TrimSpace(r.URL.Query().Get("label")),
		Module:    strings.TrimSpace(r.URL.Query().Get("module")),
		Search:    strings.TrimSpace(r.URL.Query().Get("search")),
	}
	filter.Page, filter.Limit = pageLimit(r)
	if v := strings.TrimSpace(r.URL.Query().Get("requester")); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httputil.BadRequest(w, "invalid requester")
			return
		}
		filter.Requester = &id
	}
	if v := strings.TrimSpace(r.URL.Query().Get("reviewer")); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httputil.BadRequest(w, "invalid reviewer")
			return
		}
		filter.Reviewer = &id
	}
	items, total, err := h.service.ListChangeRequests(r.Context(), filter)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, map[string]any{"items": items, "total": total, "page": filter.Page, "limit": filter.Limit})
}

func (h *Handler) GetChangeRequest(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	cr, err := h.service.GetChangeRequest(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, cr)
}

func (h *Handler) CreateChangeRequest(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	var payload changeRequestPayload
	if !decode(w, r, &payload) {
		return
	}
	cr, err := h.service.CreateChangeRequest(r.Context(), appsvc.CreateRequestInput{
		Title:            payload.Title,
		Description:      payload.Description,
		RequestType:      payload.RequestType,
		RiskLevel:        payload.RiskLevel,
		TargetEntityType: payload.TargetEntityType,
		TargetEntityID:   payload.TargetEntityID,
		ActorID:          actor,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, cr)
}

func (h *Handler) UpdateChangeRequest(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	var payload changeRequestPayload
	if !decode(w, r, &payload) {
		return
	}
	cr, err := h.service.UpdateChangeRequest(r.Context(), appsvc.UpdateRequestInput{
		ID:               id,
		Title:            payload.Title,
		Description:      payload.Description,
		RequestType:      payload.RequestType,
		RiskLevel:        payload.RiskLevel,
		TargetEntityType: payload.TargetEntityType,
		TargetEntityID:   payload.TargetEntityID,
		ActorID:          actor,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, cr)
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	h.requestAction(w, r, func(ctxActor uuid.UUID, id uuid.UUID) (*domain.ChangeRequest, error) {
		return h.service.Submit(r.Context(), id, ctxActor)
	})
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	h.decision(w, r, func(input appsvc.DecisionInput) (*domain.ChangeRequest, error) {
		return h.service.Approve(r.Context(), input)
	})
}

func (h *Handler) RequestChanges(w http.ResponseWriter, r *http.Request) {
	h.decision(w, r, func(input appsvc.DecisionInput) (*domain.ChangeRequest, error) {
		return h.service.RequestChanges(r.Context(), input)
	})
}

func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	h.decision(w, r, func(input appsvc.DecisionInput) (*domain.ChangeRequest, error) {
		return h.service.Reject(r.Context(), input)
	})
}

func (h *Handler) Merge(w http.ResponseWriter, r *http.Request) {
	h.requestAction(w, r, func(ctxActor uuid.UUID, id uuid.UUID) (*domain.ChangeRequest, error) {
		return h.service.Merge(r.Context(), id, ctxActor)
	})
}

func (h *Handler) Close(w http.ResponseWriter, r *http.Request) {
	h.requestAction(w, r, func(ctxActor uuid.UUID, id uuid.UUID) (*domain.ChangeRequest, error) {
		return h.service.Close(r.Context(), id, ctxActor)
	})
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	h.requestAction(w, r, func(ctxActor uuid.UUID, id uuid.UUID) (*domain.ChangeRequest, error) {
		return h.service.Cancel(r.Context(), id, ctxActor)
	})
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	cr, err := h.service.GetChangeRequest(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, cr.Items)
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	var payload changeItemPayload
	if !decode(w, r, &payload) {
		return
	}
	item, err := h.service.AddItem(r.Context(), appsvc.ItemInput{
		RequestID:   requestID,
		ItemType:    payload.ItemType,
		TargetTable: payload.TargetTable,
		TargetID:    payload.TargetID,
		ActionType:  payload.ActionType,
		BeforeJSON:  payload.BeforeJSON,
		AfterJSON:   payload.AfterJSON,
		ActorID:     actor,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, item)
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	itemID, ok := parseID(w, r, "itemId")
	if !ok {
		return
	}
	var payload changeItemPayload
	if !decode(w, r, &payload) {
		return
	}
	err := h.service.UpdateItem(r.Context(), appsvc.ItemInput{
		ID:          itemID,
		RequestID:   requestID,
		ItemType:    payload.ItemType,
		TargetTable: payload.TargetTable,
		TargetID:    payload.TargetID,
		ActionType:  payload.ActionType,
		BeforeJSON:  payload.BeforeJSON,
		AfterJSON:   payload.AfterJSON,
		ActorID:     actor,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.NoContent(w)
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	itemID, ok := parseID(w, r, "itemId")
	if !ok {
		return
	}
	if err := h.service.DeleteItem(r.Context(), requestID, itemID, actor); err != nil {
		writeError(w, err)
		return
	}
	httputil.NoContent(w)
}

func (h *Handler) ListApprovalSteps(w http.ResponseWriter, r *http.Request) {
	h.detailSlice(w, r, func(cr *domain.ChangeRequest) any { return cr.Steps })
}

func (h *Handler) StepApprove(w http.ResponseWriter, r *http.Request) {
	h.stepDecision(w, r, h.service.Approve)
}

func (h *Handler) StepRequestChanges(w http.ResponseWriter, r *http.Request) {
	h.stepDecision(w, r, h.service.RequestChanges)
}

func (h *Handler) StepReject(w http.ResponseWriter, r *http.Request) {
	h.stepDecision(w, r, h.service.Reject)
}

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	h.detailSlice(w, r, func(cr *domain.ChangeRequest) any { return cr.Comments })
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	var payload commentPayload
	if !decode(w, r, &payload) {
		return
	}
	comment, err := h.service.AddComment(r.Context(), appsvc.CommentInput{RequestID: requestID, ActorID: actor, Comment: payload.Comment})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, comment)
}

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	h.commentAction(w, r, h.service.UpdateComment)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	commentID, ok := parseID(w, r, "commentId")
	if !ok {
		return
	}
	if err := h.service.DeleteComment(r.Context(), appsvc.CommentInput{RequestID: requestID, CommentID: commentID, ActorID: actor}); err != nil {
		writeError(w, err)
		return
	}
	httputil.NoContent(w)
}

func (h *Handler) ListChecks(w http.ResponseWriter, r *http.Request) {
	h.detailSlice(w, r, func(cr *domain.ChangeRequest) any { return cr.Checks })
}

func (h *Handler) RerunChecks(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	if err := h.service.RerunChecks(r.Context(), requestID, actor); err != nil {
		writeError(w, err)
		return
	}
	cr, err := h.service.GetChangeRequest(r.Context(), requestID)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, cr.Checks)
}

func (h *Handler) Diff(w http.ResponseWriter, r *http.Request) {
	h.detailSlice(w, r, func(cr *domain.ChangeRequest) any {
		return map[string]any{"request_id": cr.ID, "items": cr.Items}
	})
}

func (h *Handler) ListLabels(w http.ResponseWriter, r *http.Request) {
	labels, err := h.service.ListLabels(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, labels)
}

func (h *Handler) AddLabel(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	var payload labelPayload
	if !decode(w, r, &payload) {
		return
	}
	if err := h.service.AddLabel(r.Context(), appsvc.LabelInput{RequestID: requestID, LabelCode: payload.LabelCode, ActorID: actor}); err != nil {
		writeError(w, err)
		return
	}
	httputil.NoContent(w)
}

func (h *Handler) UpsertLabel(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	var payload domain.Label
	if !decode(w, r, &payload) {
		return
	}
	if rawID := chi.URLParam(r, "id"); rawID != "" {
		if id, err := uuid.Parse(rawID); err == nil {
			payload.ID = id
		}
	}
	if err := h.service.UpsertLabel(r.Context(), &payload, actor); err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, payload)
}

func (h *Handler) RemoveLabel(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	labelID, ok := parseID(w, r, "labelId")
	if !ok {
		return
	}
	if err := h.service.RemoveLabel(r.Context(), appsvc.LabelInput{RequestID: requestID, LabelID: labelID, ActorID: actor}); err != nil {
		writeError(w, err)
		return
	}
	httputil.NoContent(w)
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.ListRoles(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, roles)
}

func (h *Handler) GetRole(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	roles, err := h.service.ListRoles(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	for _, role := range roles {
		if role.ID == id {
			httputil.OK(w, role)
			return
		}
	}
	writeError(w, domain.NotFound("permission role", id))
}

func (h *Handler) RoleAssignmentPolicies(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.ListRoles(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	policies := []map[string]any{}
	for _, role := range roles {
		if !role.CanRequestRoleAssignment {
			continue
		}
		policies = append(policies, map[string]any{
			"grantor_role_id":               role.ID,
			"grantor_role_code":             role.RoleCode,
			"max_target_priority_exclusive": role.PriorityRank,
			"assignment_scope":              role.AssignmentScope,
			"min_approvals_required":        1,
			"can_request":                   role.CanRequestRoleAssignment,
			"can_direct_merge":              false,
		})
	}
	httputil.OK(w, policies)
}

func (h *Handler) RoleAssignmentRequest(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	userID, ok := parseID(w, r, "userId")
	if !ok {
		return
	}
	var payload roleAssignmentPayload
	if !decode(w, r, &payload) {
		return
	}
	cr, err := h.service.CreateChangeRequest(r.Context(), appsvc.CreateRequestInput{
		Title:            "Role assignment request",
		Description:      payload.Reason,
		RequestType:      domain.RequestTypeRoleAssignment,
		RiskLevel:        domain.RiskMedium,
		TargetEntityType: "USER",
		TargetEntityID:   userID.String(),
		ActorID:          actor,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	after := map[string]any{"user_id": userID.String(), "role_code": payload.RoleCode}
	if payload.RoleID != "" {
		after["role_id"] = payload.RoleID
	}
	raw, _ := json.Marshal(after)
	item, err := h.service.AddItem(r.Context(), appsvc.ItemInput{
		RequestID:   cr.ID,
		ItemType:    "ROLE_ASSIGNMENT",
		TargetTable: "permission_user_role_assignments",
		TargetID:    userID.String(),
		ActionType:  "ASSIGN_ROLE",
		AfterJSON:   raw,
		ActorID:     actor,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.Created(w, map[string]any{"request": cr, "item": item})
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	users, total, err := h.service.ListUsers(r.Context(), strings.TrimSpace(r.URL.Query().Get("search")), page, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, map[string]any{"items": users, "total": total, "page": page, "limit": limit})
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	user, err := h.service.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	if user == nil {
		writeError(w, domain.NotFound("user", id))
		return
	}
	httputil.OK(w, user)
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	groups, total, err := h.service.ListGroups(r.Context(), strings.TrimSpace(r.URL.Query().Get("search")), page, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, map[string]any{"items": groups, "total": total, "page": page, "limit": limit})
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	group, err := h.service.GetGroup(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	if group == nil {
		writeError(w, domain.NotFound("group", id))
		return
	}
	httputil.OK(w, group)
}

func (h *Handler) FunctionDefinitions(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListFunctionDefinitions(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, items)
}

func (h *Handler) FunctionRights(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListFunctionRights(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, items)
}

func (h *Handler) EffectivePermissions(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseID(w, r, "userId")
	if !ok {
		return
	}
	effective, err := h.service.EffectivePermissions(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, effective)
}

func (h *Handler) DataRights(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListDataRights(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, items)
}

func (h *Handler) ApprovalSettings(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListApprovalSettings(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, items)
}

func (h *Handler) ApprovalSetting(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	item, err := h.service.GetApprovalSetting(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	if item == nil {
		writeError(w, domain.NotFound("approval setting", id))
		return
	}
	httputil.OK(w, item)
}

func (h *Handler) AuditLogs(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	items, total, err := h.service.ListAuditLogs(r.Context(), page, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (h *Handler) AuditExport(w http.ResponseWriter, r *http.Request) {
	page, limit := pageLimit(r)
	items, _, err := h.service.ListAuditLogs(r.Context(), page, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="permission-audit.csv"`)
	_, _ = w.Write([]byte("id,actor,action,module,entity_type,entity_id,created_at\n"))
	for _, item := range items {
		_, _ = w.Write([]byte(item.ID.String() + "," + item.ActorName + "," + item.Action + "," + item.Module + "," + item.EntityType + "," + item.EntityID + "," + item.CreatedAt.Format("2006-01-02T15:04:05Z07:00") + "\n"))
	}
}

func (h *Handler) NotificationSettings(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListNotificationSettings(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, items)
}

func (h *Handler) UpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	var payload notificationSettingsPayload
	if !decode(w, r, &payload) {
		return
	}
	for _, setting := range payload.Settings {
		if err := h.service.UpdateNotificationSetting(r.Context(), setting, actor); err != nil {
			writeError(w, err)
			return
		}
	}
	httputil.NoContent(w)
}

func (h *Handler) requestAction(w http.ResponseWriter, r *http.Request, fn func(uuid.UUID, uuid.UUID) (*domain.ChangeRequest, error)) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	cr, err := fn(actor, id)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, cr)
}

func (h *Handler) decision(w http.ResponseWriter, r *http.Request, fn func(appsvc.DecisionInput) (*domain.ChangeRequest, error)) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	var payload decisionPayload
	_ = json.NewDecoder(r.Body).Decode(&payload)
	cr, err := fn(appsvc.DecisionInput{RequestID: requestID, ActorID: actor, Comment: payload.Comment})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, cr)
}

func (h *Handler) stepDecision(w http.ResponseWriter, r *http.Request, fn func(context.Context, appsvc.DecisionInput) (*domain.ChangeRequest, error)) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	stepID, ok := parseID(w, r, "stepId")
	if !ok {
		return
	}
	var payload decisionPayload
	_ = json.NewDecoder(r.Body).Decode(&payload)
	cr, err := fn(r.Context(), appsvc.DecisionInput{RequestID: requestID, StepID: stepID, ActorID: actor, Comment: payload.Comment})
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, cr)
}

func (h *Handler) detailSlice(w http.ResponseWriter, r *http.Request, fn func(*domain.ChangeRequest) any) {
	id, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	cr, err := h.service.GetChangeRequest(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.OK(w, fn(cr))
}

func (h *Handler) commentAction(w http.ResponseWriter, r *http.Request, fn func(context.Context, appsvc.CommentInput) error) {
	actor, ok := actorID(w, r)
	if !ok {
		return
	}
	requestID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	commentID, ok := parseID(w, r, "commentId")
	if !ok {
		return
	}
	var payload commentPayload
	if !decode(w, r, &payload) {
		return
	}
	if err := fn(r.Context(), appsvc.CommentInput{RequestID: requestID, CommentID: commentID, ActorID: actor, Comment: payload.Comment}); err != nil {
		writeError(w, err)
		return
	}
	httputil.NoContent(w)
}

func decode(w http.ResponseWriter, r *http.Request, dest any) bool {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return false
	}
	return true
}

func parseID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		httputil.BadRequest(w, "invalid "+key)
		return uuid.Nil, false
	}
	return id, true
}

func actorID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "not authenticated")
		return uuid.Nil, false
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		httputil.Unauthorized(w, "invalid user subject")
		return uuid.Nil, false
	}
	return id, true
}

func pageLimit(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func writeError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var pe *domain.PermissionError
	if errors.As(err, &pe) {
		httputil.JSON(w, pe.HTTPStatus(), httputil.ErrorResponse{
			Error:   pe.Message,
			Code:    pe.Code,
			Details: pe.Details,
		})
		return
	}
	httputil.InternalError(w, err.Error())
}
