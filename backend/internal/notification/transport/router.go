package transport

import (
	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/transport/handler"
	notifperm "github.com/neo-kanta/ims-th-solution/backend/internal/notification/permission"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// RegisterRoutes mounts all notification endpoints on the authenticated router.
// User-facing notification center endpoints are open to any authenticated user.
// Admin/demo email endpoints are gated with permission middleware.
func RegisterRoutes(
	r chi.Router,
	h *handler.Handler,
	emailH *handler.EmailOutboxHandler,
	checker middleware.PermissionChecker,
) {
	r.Route("/notifications", func(r chi.Router) {
		// User-facing notification center (any authenticated user).
		r.Get("/", h.List)
		r.Post("/{id}/read", h.MarkRead)
		r.Post("/read-all", h.MarkAllRead)

		// Email admin/demo endpoints (permission-gated).
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(checker, notifperm.CodeView))
			r.Get("/email-outbox", emailH.ListOutbox)
			r.Get("/email-outbox/{outbox_id}", emailH.GetOutboxDetail)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(checker, notifperm.CodeRetry))
			r.Post("/email-outbox/{outbox_id}/retry", emailH.RetryOutbox)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(checker, notifperm.CodeTest))
			r.Post("/email/test", emailH.SendTestEmail)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission(checker, notifperm.CodeHealth))
			r.Get("/email/health", emailH.EmailHealth)
		})
	})
}
