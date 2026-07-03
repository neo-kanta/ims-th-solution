package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Handler serves the user-facing in-app notification endpoints.
type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// --- response types ---

type userSummaryResponse struct {
	ID          *uuid.UUID `json:"id"`
	Username    string     `json:"username"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
}

type eventResponse struct {
	Type     string `json:"type"`
	Label    string `json:"label"`
	Category string `json:"category"`
	Severity string `json:"severity"`
}

type contextResponse struct {
	BusinessType      string     `json:"business_type"`
	BusinessLabel     string     `json:"business_label"`
	BusinessID        *uuid.UUID `json:"business_id,omitempty"`
	BusinessReference string     `json:"business_reference"`
	BusinessTitle     string     `json:"business_title"`
}

type actionResponse struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type notificationResponse struct {
	NotificationID uuid.UUID           `json:"notification_id"`
	ID             uuid.UUID           `json:"id"` // backward-compat alias
	Recipient      userSummaryResponse `json:"recipient"`
	Event          eventResponse       `json:"event"`
	Title          string              `json:"title"`
	Body           string              `json:"body,omitempty"`
	Context        contextResponse     `json:"context"`
	Action         actionResponse      `json:"action,omitempty"`
	IsRead         bool                `json:"is_read"`
	ReadAt         *string             `json:"read_at,omitempty"`
	CreatedAt      string              `json:"created_at"`
}

type listResponse struct {
	Items  []notificationResponse `json:"items"`
	Total  int                    `json:"total"`
	Unread int                    `json:"unread"`
}

type markReadResponse struct {
	NotificationID uuid.UUID           `json:"notification_id"`
	Recipient      userSummaryResponse `json:"recipient"`
	Status         string              `json:"status"`
	ReadAt         *string             `json:"read_at,omitempty"`
}

type markAllReadResponse struct {
	Recipient userSummaryResponse `json:"recipient"`
	Updated   int                 `json:"updated"`
}

// List handles GET /notifications
// @Summary List my in-app notifications
// @Description Returns the authenticated user's in-app notification center list joined with IAM user data.
// @Tags Notification
// @Security BearerAuth
// @Produce json
// @Param unread_only query bool false "Return only unread notifications"
// @Param limit query int false "Page size (default 50, max 200)"
// @Param offset query int false "Offset for paging"
// @Success 200 {object} listResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /notifications [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor := currentUser(r)
	if actor == uuid.Nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	unread := strings.EqualFold(r.URL.Query().Get("unread_only"), "true")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, unreadCount, err := h.svc.List(r.Context(), actor, unread, limit, offset)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]notificationResponse, 0, len(items))
	for _, n := range items {
		out = append(out, toNotificationResponse(n))
	}
	httputil.OK(w, listResponse{Items: out, Total: total, Unread: unreadCount})
}

// MarkRead handles POST /notifications/{id}/read
// @Summary Mark one notification as read
// @Description Marks a single in-app notification as read for the authenticated user.
// @Tags Notification
// @Security BearerAuth
// @Produce json
// @Param id path string true "Notification ID (UUID)"
// @Success 200 {object} markReadResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /notifications/{id}/read [post]
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	actor := currentUser(r)
	if actor == uuid.Nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.BadRequest(w, "invalid notification id")
		return
	}
	if err := h.svc.MarkRead(r.Context(), id, actor); err != nil {
		if err.Error() != "notification: no row updated" {
			httputil.InternalError(w, err.Error())
			return
		}
	}
	user, _ := h.svc.GetUserByID(r.Context(), actor)
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
	httputil.OK(w, markReadResponse{
		NotificationID: id,
		Recipient:      toUserSummaryResponse(actor, user),
		Status:         "READ",
		ReadAt:         &now,
	})
}

// MarkAllRead handles POST /notifications/read-all
// @Summary Mark all notifications as read
// @Description Marks every unread in-app notification for the authenticated user as read.
// @Tags Notification
// @Security BearerAuth
// @Produce json
// @Success 200 {object} markAllReadResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /notifications/read-all [post]
func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	actor := currentUser(r)
	if actor == uuid.Nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	n, user, err := h.svc.MarkAllRead(r.Context(), actor)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.OK(w, markAllReadResponse{
		Recipient: toUserSummaryResponse(actor, user),
		Updated:   n,
	})
}

// --- helpers ---

func toNotificationResponse(n *entity.Notification) notificationResponse {
	if n == nil {
		return notificationResponse{}
	}
	var readAt *string
	if n.ReadAt != nil {
		s := n.ReadAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		readAt = &s
	}

	def := valueobject.LookupEventByString(n.EventType)

	actionLabel := n.ActionLabel
	if actionLabel == "" {
		actionLabel = def.DefaultActionLabel
	}

	recipientID := n.RecipientUserID
	return notificationResponse{
		NotificationID: n.ID,
		ID:             n.ID,
		Recipient: userSummaryResponse{
			ID:          &recipientID,
			Username:    n.RecipientUsername,
			DisplayName: n.RecipientDisplayName,
			Email:       n.RecipientEmail,
		},
		Event: eventResponse{
			Type:     string(def.Type),
			Label:    def.Label,
			Category: def.Category,
			Severity: def.Severity,
		},
		Title: n.Title,
		Body:  n.Body,
		Context: contextResponse{
			BusinessType:      n.BusinessType,
			BusinessLabel:     n.BusinessLabel,
			BusinessID:        n.BusinessID,
			BusinessReference: n.BusinessReference,
			BusinessTitle:     n.BusinessTitle,
		},
		Action: actionResponse{
			Label: actionLabel,
			URL:   n.ActionURL,
		},
		IsRead:    n.IsRead,
		ReadAt:    readAt,
		CreatedAt: n.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toUserSummaryResponse(id uuid.UUID, u domain.UserSummary) userSummaryResponse {
	uid := id
	if u.ID != uuid.Nil {
		uid = u.ID
	}
	return userSummaryResponse{
		ID:          &uid,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Email:       u.Email,
	}
}

func currentUser(r *http.Request) uuid.UUID {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		return uuid.Nil
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil
	}
	return id
}
