package http

import (
	"github.com/go-chi/chi/v5"

	perm "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/permission"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// RegisterRoutes mounts the watchlist API routes under the provided router.
// The router is assumed to be inside an already-authenticated group.
func RegisterRoutes(r chi.Router, h *Handler, pc middleware.PermissionChecker) {
	r.Route("/watchlists", func(r chi.Router) {
		// A. List items
		r.With(middleware.RequirePermission(pc, perm.CodeView)).Get("/", h.ListItems)

		// B. Create item
		r.With(middleware.RequirePermission(pc, perm.CodeManage)).Post("/items", h.CreateItem)

		// C. Update item
		r.With(middleware.RequirePermission(pc, perm.CodeManage)).Patch("/items/{id}", h.UpdateItem)

		// D. Delete item
		r.With(middleware.RequirePermission(pc, perm.CodeManage)).Delete("/items/{id}", h.DeleteItem)

		// E. List alerts
		r.With(middleware.RequirePermission(pc, perm.CodeView)).Get("/alerts", h.ListAlerts)

		// F. Acknowledge alert
		r.With(middleware.RequirePermission(pc, perm.CodeAlertAck)).Post("/alerts/{id}/acknowledge", h.AcknowledgeAlert)

		// G. Manual evaluation
		r.With(middleware.RequirePermission(pc, perm.CodeEvaluate)).Post("/evaluate", h.ManualEvaluate)
	})
}
