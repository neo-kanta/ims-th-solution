// Package transport wires the workflow module's HTTP routes.
package transport

import (
	"github.com/go-chi/chi/v5"

	workflowperm "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/permission"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/transport/handler"
	platformmw "github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// RegisterRoutes mounts all workflow routes under the given router prefix.
// Caller is responsible for applying auth middleware before calling this.
//
// Legacy UUID-based routes (unchanged):
//
//	GET  /workflow/day-states/{contractId}               → current state
//	GET  /workflow/day-states/{contractId}/history       → transition history
//	POST /workflow/day-states/{contractId}/transitions   → execute transition
//
// New business-readable routes:
//
//	GET  /workflow/daily                                 → aggregated daily state
//	POST /workflow/daily/execute                         → execute via operationType
//	GET  /workflow/daily/transitions                     → paginated history
//	GET  /workflow/transition-rules                      → static state machine topology
//	GET  /workflow/settings                              → approver configuration
//	PUT  /workflow/settings                              → update approver configuration (Admin only)
func RegisterRoutes(r chi.Router, h *handler.WorkflowHandler, daily *handler.DailyHandler, permChecker platformmw.PermissionChecker) {
	// ── Legacy UUID-based endpoints ───────────────────────────────────────────
	r.Route("/workflow/day-states", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			if permChecker != nil {
				r.Use(platformmw.RequirePermission(permChecker, workflowperm.CodeView))
			}
			r.Get("/{contractId}", h.GetCurrentState)
			r.Get("/{contractId}/history", h.GetHistory)
		})

		r.Post("/{contractId}/transitions", h.ExecuteTransition)
	})

	r.Route("/workflow/scheduler", func(r chi.Router) {
		if permChecker != nil {
			r.Use(platformmw.RequirePermission(permChecker, workflowperm.CodeRunScheduler))
		}
		r.Post("/run-once", h.RunSchedulerOnce)
	})

	// ── New business-readable daily endpoints ─────────────────────────────────
	// Auth is enforced by the underlying handlers (JWT claims + admin/approver check).
	// No RequirePermission middleware here — the new API uses settings-based approval.
	r.Route("/workflow", func(r chi.Router) {
		r.Get("/daily", daily.GetDailyWorkflow)
		r.Post("/daily/execute", daily.ExecuteDailyTransition)
		r.Get("/daily/transitions", daily.GetDailyTransitions)
		r.Get("/transition-rules", daily.GetTransitionRules)
		r.Get("/settings", daily.GetSettings)
		r.Put("/settings", daily.UpdateSettings)
	})
}
