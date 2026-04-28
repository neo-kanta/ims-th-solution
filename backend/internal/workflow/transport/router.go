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
// Route layout (all prefixed with /workflow by the caller's router):
//
//	GET  /workflow/day-states/{contractId}               → current state (+ synthetic NOT_STARTED)
//	GET  /workflow/day-states/{contractId}/history       → transition history
//	POST /workflow/day-states/{contractId}/transitions   → execute transition (body.action)
//
// Per-action permission enforcement for the POST route is performed inside the
// handler (it reads req.Action before calling the matching WORKFLOW_* code).
// Gating the whole POST with a single code here would either be wrong (what
// Batch 1 did — APPROVE was reachable with WORKFLOW_OPEN_DAY) or over-restrictive.
func RegisterRoutes(r chi.Router, h *handler.WorkflowHandler, permChecker platformmw.PermissionChecker) {
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
}
