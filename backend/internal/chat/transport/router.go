package transport

import (
	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/transport/handler"
)

// RegisterRoutes mounts chat routes onto the given authenticated router.
// Expected to be called inside an /api/v1 group that already has auth
// middleware applied.
//
//   - POST /chat                              streams one assistant turn (SSE)
//   - GET  /chat/sessions                     list the caller's sessions
//   - GET  /chat/sessions/{session_id}        one session's metadata
//   - GET  /chat/sessions/{session_id}/messages  one session's transcript
//
// Session reads are owner-scoped; IAM_AUDIT_VIEW holders may read any session
// (strictly audited). sessionHandler may be nil (history endpoints omitted).
func RegisterRoutes(r chi.Router, h *handler.ChatHandler, sessionHandler *handler.SessionHandler) {
	if h == nil {
		return
	}
	r.Route("/chat", func(r chi.Router) {
		r.Post("/", h.SendMessage)
		if sessionHandler != nil {
			r.Get("/sessions", sessionHandler.ListSessions)
			r.Get("/sessions/{session_id}", sessionHandler.GetSession)
			r.Get("/sessions/{session_id}/messages", sessionHandler.ListMessages)
		}
	})
}
