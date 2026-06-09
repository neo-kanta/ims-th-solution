package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/types"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// SessionHandler serves chat session-history reads.
type SessionHandler struct {
	queries *query.SessionQueries
}

// NewSessionHandler wires the handler.
func NewSessionHandler(queries *query.SessionQueries) *SessionHandler {
	return &SessionHandler{queries: queries}
}

// ListSessions returns the caller's chat sessions (most-recent first).
// @Summary      List chat sessions
// @Description  Lists the caller's chat sessions. Auditors (IAM_AUDIT_VIEW) may pass user_id to list another user's sessions; such access is strictly audited.
// @Tags         Chat
// @Produce      json
// @Param        page query int false "Page (default 1)"
// @Param        limit query int false "Page size (default 20, max 100)"
// @Param        user_id query string false "Auditor-only: list this user's sessions (requires IAM_AUDIT_VIEW)"
// @Success      200 {object} response.SessionListResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /chat/sessions [get]
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	caller, ok := h.accessContext(w, r)
	if !ok {
		return
	}
	pag := parsePagination(r)

	var target uuid.UUID
	if raw := r.URL.Query().Get("user_id"); raw != "" {
		tid, err := uuid.Parse(raw)
		if err != nil {
			httputil.BadRequest(w, "invalid user_id")
			return
		}
		target = tid
	}

	page, err := h.queries.ListSessions(r.Context(), caller, target, pag)
	if err != nil {
		h.writeQueryError(w, err)
		return
	}

	items := make([]response.SessionSummaryResponse, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, toSessionSummary(page.Items[i]))
	}
	httputil.OK(w, response.SessionListResponse{
		Items: items, Total: page.Total, Page: page.Page, Limit: page.Limit, TotalPages: page.TotalPages,
	})
}

// GetSession returns one chat session's metadata.
// @Summary      Get a chat session
// @Description  Returns one session's metadata. Owner or auditor (IAM_AUDIT_VIEW). Auditor access is strictly audited.
// @Tags         Chat
// @Produce      json
// @Param        session_id path string true "Session UUID"
// @Success      200 {object} response.SessionSummaryResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /chat/sessions/{session_id} [get]
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	caller, ok := h.accessContext(w, r)
	if !ok {
		return
	}
	sessionID, err := uuid.Parse(chi.URLParam(r, "session_id"))
	if err != nil {
		httputil.BadRequest(w, "invalid session_id")
		return
	}
	s, err := h.queries.GetSession(r.Context(), caller, sessionID)
	if err != nil {
		h.writeQueryError(w, err)
		return
	}
	httputil.OK(w, toSessionSummary(*s))
}

// ListMessages returns one session's transcript (paginated).
// @Summary      List chat session messages
// @Description  Returns a session's messages with safe provenance (no raw provider payload). Owner or auditor (IAM_AUDIT_VIEW). Auditor access is strictly audited.
// @Tags         Chat
// @Produce      json
// @Param        session_id path string true "Session UUID"
// @Param        page query int false "Page (default 1)"
// @Param        limit query int false "Page size (default 20, max 100)"
// @Success      200 {object} response.SessionMessagesResponse
// @Failure      401 {object} httputil.ErrorResponse
// @Failure      403 {object} httputil.ErrorResponse
// @Failure      404 {object} httputil.ErrorResponse
// @Security     BearerAuth
// @Router       /chat/sessions/{session_id}/messages [get]
func (h *SessionHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	caller, ok := h.accessContext(w, r)
	if !ok {
		return
	}
	sessionID, err := uuid.Parse(chi.URLParam(r, "session_id"))
	if err != nil {
		httputil.BadRequest(w, "invalid session_id")
		return
	}
	pag := parsePagination(r)

	_, page, err := h.queries.ListMessages(r.Context(), caller, sessionID, pag)
	if err != nil {
		h.writeQueryError(w, err)
		return
	}
	items := make([]response.ChatMessageResponse, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, toMessageResponse(page.Items[i]))
	}
	httputil.OK(w, response.SessionMessagesResponse{
		SessionID: sessionID.String(), Items: items, Total: page.Total, Page: page.Page, Limit: page.Limit, TotalPages: page.TotalPages,
	})
}

// accessContext builds the authenticated AccessContext or writes 401 and
// returns ok=false.
func (h *SessionHandler) accessContext(w http.ResponseWriter, r *http.Request) (query.AccessContext, bool) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil || claims.Subject == "" {
		httputil.Unauthorized(w, "missing user context")
		return query.AccessContext{}, false
	}
	uid, err := uuid.Parse(claims.Subject)
	if err != nil {
		httputil.Unauthorized(w, "invalid user identity")
		return query.AccessContext{}, false
	}
	return query.AccessContext{
		CallerID:      uid,
		IPAddress:     middleware.GetClientIP(r),
		UserAgent:     r.UserAgent(),
		CorrelationID: chimw.GetReqID(r.Context()),
	}, true
}

func (h *SessionHandler) writeQueryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, query.ErrForbidden):
		httputil.Forbidden(w, "you do not have access to this session")
	case errors.Is(err, query.ErrNotFound):
		httputil.NotFound(w, "session not found")
	default:
		httputil.InternalError(w, "failed to read session history")
	}
}

func parsePagination(r *http.Request) types.Pagination {
	page := atoiDefault(r.URL.Query().Get("page"), 1)
	if page < 1 {
		page = 1
	}
	limit := atoiDefault(r.URL.Query().Get("limit"), 20)
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return types.Pagination{Page: page, Limit: limit, Offset: (page - 1) * limit}
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

func toSessionSummary(s entity.Session) response.SessionSummaryResponse {
	return response.SessionSummaryResponse{
		ID:        s.ID.String(),
		Provider:  s.Provider.String(),
		Model:     s.Model,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func toMessageResponse(m entity.Message) response.ChatMessageResponse {
	var prov json.RawMessage
	if len(m.ProvenanceMap) > 0 && string(m.ProvenanceMap) != "{}" {
		prov = m.ProvenanceMap
	}
	return response.ChatMessageResponse{
		ID:         m.ID.String(),
		Role:       m.Role.String(),
		Content:    m.Content,
		Provenance: prov,
		CreatedAt:  m.CreatedAt,
	}
}
