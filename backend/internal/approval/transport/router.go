// Package transport wires the approval module's HTTP routes.
package transport

import (
	"github.com/go-chi/chi/v5"

	perm "github.com/neo-kanta/ims-th-solution/backend/internal/approval/permission"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/transport/handler"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// RegisterRoutes mounts the approval runtime and configuration routes onto an
// already-authenticated router. Each route group is gated by a function
// permission via middleware.RequirePermission — backend is the source of truth.
func RegisterRoutes(
	r chi.Router,
	runtime *handler.RuntimeHandler,
	config *handler.ConfigHandler,
	pc middleware.PermissionChecker,
) {
	// ── Runtime: /approvals ─────────────────────────────────────────────────
	r.Route("/approvals", func(r chi.Router) {
		r.With(middleware.RequirePermission(pc, perm.CodeViewInbox)).Get("/inbox", runtime.GetInbox)

		r.With(middleware.RequirePermission(pc, perm.CodeViewRequest)).Get("/requests", runtime.ListRequests)
		r.With(middleware.RequirePermission(pc, perm.CodeViewRequest)).Get("/requests/{requestId}", runtime.GetRequest)
		r.With(middleware.RequirePermission(pc, perm.CodeAuditView)).Get("/requests/{requestId}/timeline", runtime.GetTimeline)
		r.With(middleware.RequirePermission(pc, perm.CodeViewRequest)).Get("/subjects/{subjectType}/{subjectId}/status", runtime.GetSubjectStatus)

		r.With(middleware.RequirePermission(pc, perm.CodeSubmit)).Post("/submit", runtime.Submit)
		r.With(middleware.RequirePermission(pc, perm.CodeApprove)).Post("/tasks/{taskId}/approve", runtime.Approve)
		r.With(middleware.RequirePermission(pc, perm.CodeReject)).Post("/tasks/{taskId}/reject", runtime.Reject)
		r.With(middleware.RequirePermission(pc, perm.CodeWithdraw)).Post("/requests/{requestId}/withdraw", runtime.Withdraw)
		r.With(middleware.RequirePermission(pc, perm.CodeCancel)).Post("/requests/{requestId}/cancel", runtime.Cancel)
		r.With(middleware.RequirePermission(pc, perm.CodeRevoke)).Post("/requests/{requestId}/revoke", runtime.Revoke)
	})

	// ── Configuration: /approval-config ─────────────────────────────────────
	r.Route("/approval-config", func(r chi.Router) {
		// Groups
		r.With(middleware.RequirePermission(pc, perm.CodeConfigView)).Get("/groups", config.ListGroups)
		r.With(middleware.RequirePermission(pc, perm.CodeGroupManage)).Post("/groups", config.CreateGroup)
		r.With(middleware.RequirePermission(pc, perm.CodeGroupManage)).Put("/groups/{id}", config.UpdateGroup)
		r.With(middleware.RequirePermission(pc, perm.CodeConfigView)).Get("/groups/{id}/members", config.ListGroupMembers)
		r.With(middleware.RequirePermission(pc, perm.CodeGroupManage)).Post("/groups/{id}/members", config.AddGroupMember)
		r.With(middleware.RequirePermission(pc, perm.CodeGroupManage)).Put("/groups/{id}/members/{memberId}", config.UpdateGroupMember)
		r.With(middleware.RequirePermission(pc, perm.CodeGroupManage)).Post("/groups/{id}/members/{memberId}/approve", config.ApproveGroupMember)
		r.With(middleware.RequirePermission(pc, perm.CodeGroupManage)).Post("/groups/{id}/members/{memberId}/revoke", config.RevokeGroupMember)
		r.With(middleware.RequirePermission(pc, perm.CodeGroupManage)).Post("/groups/{id}/members/reorder", config.ReorderGroupMembers)

		// Teams
		r.With(middleware.RequirePermission(pc, perm.CodeConfigView)).Get("/teams", config.ListTeams)
		r.With(middleware.RequirePermission(pc, perm.CodeTeamManage)).Post("/teams", config.CreateTeam)
		r.With(middleware.RequirePermission(pc, perm.CodeTeamManage)).Put("/teams/{id}", config.UpdateTeam)
		r.With(middleware.RequirePermission(pc, perm.CodeConfigView)).Get("/teams/{id}/contracts", config.ListTeamContracts)
		r.With(middleware.RequirePermission(pc, perm.CodeTeamManage)).Post("/teams/{id}/contracts", config.AssignTeamContract)
		r.With(middleware.RequirePermission(pc, perm.CodeConfigView)).Get("/teams/{id}/members", config.ListTeamMembers)
		r.With(middleware.RequirePermission(pc, perm.CodeTeamManage)).Post("/teams/{id}/members", config.AddTeamMember)
		r.With(middleware.RequirePermission(pc, perm.CodeTeamManage)).Put("/teams/{id}/members/{memberId}", config.UpdateTeamMember)
		r.With(middleware.RequirePermission(pc, perm.CodeTeamManage)).Delete("/teams/{id}/members/{memberId}", config.RemoveTeamMember)

		// Processes
		r.With(middleware.RequirePermission(pc, perm.CodeConfigView)).Get("/processes", config.ListProcesses)
		r.With(middleware.RequirePermission(pc, perm.CodeConfigView)).Get("/processes/{id}", config.GetProcess)
		r.With(middleware.RequirePermission(pc, perm.CodeProcessManage)).Post("/processes", config.CreateProcess)
		r.With(middleware.RequirePermission(pc, perm.CodeProcessManage)).Put("/processes/{id}", config.UpdateProcess)
		r.With(middleware.RequirePermission(pc, perm.CodeProcessManage)).Post("/processes/{id}/activate", config.ActivateProcess)
		r.With(middleware.RequirePermission(pc, perm.CodeProcessManage)).Post("/processes/{id}/deactivate", config.DeactivateProcess)
	})
}
