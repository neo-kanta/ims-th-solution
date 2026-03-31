package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// SessionHandler handles session management endpoints.
type SessionHandler struct {
	listSessionsQry *query.ListSessionsQuery
	sessionRepo     domain.SessionRepository
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(listSessionsQry *query.ListSessionsQuery, sessionRepo domain.SessionRepository) *SessionHandler {
	return &SessionHandler{
		listSessionsQry: listSessionsQry,
		sessionRepo:     sessionRepo,
	}
}

// ListMySessions handles GET /auth/sessions.
// @Summary List My Sessions
// @Description List all active sessions for the current user
// @Tags Sessions
// @Security BearerAuth
// @Produce json
// @Success 200 {array} response.SessionResponse
// @Failure 401 {object} map[string]interface{}
// @Router /auth/sessions [get]
func (h *SessionHandler) ListMySessions(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	sessions, err := h.listSessionsQry.Execute(r.Context(), userID)
	if err != nil {
		httputil.InternalError(w, "failed to list sessions")
		return
	}

	result := make([]response.SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, response.SessionResponse{
			ID:             s.ID,
			IPAddress:      s.IPAddress,
			UserAgent:      s.UserAgent,
			LastActivityAt: s.LastActivityAt,
			CreatedAt:      s.CreatedAt,
			ExpiresAt:      s.ExpiresAt,
		})
	}

	httputil.OK(w, result)
}

// RevokeSession handles POST /auth/sessions/{id}/revoke.
// @Summary Revoke a Session
// @Description Revoke a specific active session
// @Tags Sessions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Session UUID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /auth/sessions/{id}/revoke [post]
func (h *SessionHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		httputil.BadRequest(w, "invalid session id")
		return
	}

	// Verify session belongs to user
	session, err := h.sessionRepo.FindByID(r.Context(), sessionID)
	if err != nil || session == nil {
		httputil.NotFound(w, "session not found")
		return
	}
	if session.UserID != userID {
		httputil.Forbidden(w, "cannot revoke another user's session")
		return
	}

	if err := h.sessionRepo.RevokeByIDWithReason(r.Context(), sessionID, entity.RevokeReasonLogout); err != nil {
		httputil.InternalError(w, "failed to revoke session")
		return
	}

	httputil.NoContent(w)
}

// AdminListUserSessions handles GET /admin/users/{id}/sessions.
// @Summary List User Sessions (Admin)
// @Description List all active sessions for a specific user (admin only)
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "User UUID"
// @Success 200 {array} response.SessionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/users/{id}/sessions [get]
func (h *SessionHandler) AdminListUserSessions(w http.ResponseWriter, r *http.Request) {
	targetIDStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		httputil.BadRequest(w, "invalid user id")
		return
	}

	sessions, err := h.listSessionsQry.Execute(r.Context(), targetID)
	if err != nil {
		httputil.InternalError(w, "failed to list sessions")
		return
	}

	result := make([]response.SessionResponse, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, response.SessionResponse{
			ID:             s.ID,
			IPAddress:      s.IPAddress,
			UserAgent:      s.UserAgent,
			LastActivityAt: s.LastActivityAt,
			CreatedAt:      s.CreatedAt,
			ExpiresAt:      s.ExpiresAt,
		})
	}

	httputil.OK(w, result)
}

// AdminRevokeSession handles POST /admin/sessions/{id}/revoke.
// @Summary Revoke Any Session (Admin)
// @Description Admin revoke any session by ID
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "Session UUID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Router /admin/sessions/{id}/revoke [post]
func (h *SessionHandler) AdminRevokeSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		httputil.BadRequest(w, "invalid session id")
		return
	}

	if err := h.sessionRepo.RevokeByIDWithReason(r.Context(), sessionID, entity.RevokeReasonAdminRevoke); err != nil {
		httputil.InternalError(w, "failed to revoke session")
		return
	}

	httputil.NoContent(w)
}
