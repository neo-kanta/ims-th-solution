package transport

import (
	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/transport/handler"
)

// RegisterRoutes mounts the integration module's HTTP routes onto r.
// All routes require an authenticated caller (auth middleware is applied by
// the parent router group in main.go).
func RegisterRoutes(r chi.Router, h *handler.DashboardHandler) {
	r.Route("/integration", func(r chi.Router) {
		r.Get("/dashboard/me", h.GetDashboardSnapshot)
		r.Get("/tasks/my", h.GetMyTasks)
		r.Get("/tasks/my/summary", h.GetMyTasksSummary)
	})
}
