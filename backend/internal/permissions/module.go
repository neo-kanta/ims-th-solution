package permissions

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	appservice "github.com/neo-kanta/ims-th-solution/backend/internal/permissions/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/permissions/infrastructure/persistence"
	permcode "github.com/neo-kanta/ims-th-solution/backend/internal/permissions/permission"
	"github.com/neo-kanta/ims-th-solution/backend/internal/permissions/transport/handler"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Module Permissions Management - accounts, groups, function permissions, data permissions.
type Module struct {
	handler *handler.Handler
	checker middleware.PermissionChecker
}

// NewModule creates a new permissions module with its dependencies.
func NewModule(pool *pgxpool.Pool, checker middleware.PermissionChecker) *Module {
	repo := persistence.NewPostgresRepository(pool)
	svc := appservice.NewService(pool, repo, checker)
	return &Module{handler: handler.NewHandler(svc), checker: checker}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(r chi.Router) {
	if m == nil || m.handler == nil || m.checker == nil {
		return
	}
	h := m.handler
	pc := m.checker

	r.Route("/permissions", func(r chi.Router) {
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Get("/change-requests", h.ListChangeRequests)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Get("/change-requests/{id}", h.GetChangeRequest)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCreate)).Post("/change-requests", h.CreateChangeRequest)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCreate)).Put("/change-requests/{id}", h.UpdateChangeRequest)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestSubmit)).Post("/change-requests/{id}/submit", h.Submit)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestApprove)).Post("/change-requests/{id}/approve", h.Approve)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestApprove)).Post("/change-requests/{id}/request-changes", h.RequestChanges)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReject)).Post("/change-requests/{id}/reject", h.Reject)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestMerge)).Post("/change-requests/{id}/merge", h.Merge)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestClose)).Post("/change-requests/{id}/close", h.Close)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCancel)).Post("/change-requests/{id}/cancel", h.Cancel)

		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Get("/change-requests/{id}/items", h.ListItems)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCreate)).Post("/change-requests/{id}/items", h.AddItem)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCreate)).Put("/change-requests/{id}/items/{itemId}", h.UpdateItem)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCreate)).Delete("/change-requests/{id}/items/{itemId}", h.DeleteItem)

		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Get("/change-requests/{id}/approval-steps", h.ListApprovalSteps)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestApprove)).Post("/change-requests/{id}/approval-steps/{stepId}/approve", h.StepApprove)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestApprove)).Post("/change-requests/{id}/approval-steps/{stepId}/request-changes", h.StepRequestChanges)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReject)).Post("/change-requests/{id}/approval-steps/{stepId}/reject", h.StepReject)

		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Get("/change-requests/{id}/comments", h.ListComments)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Post("/change-requests/{id}/comments", h.AddComment)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Put("/change-requests/{id}/comments/{commentId}", h.UpdateComment)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Delete("/change-requests/{id}/comments/{commentId}", h.DeleteComment)

		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Get("/change-requests/{id}/checks", h.ListChecks)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Post("/change-requests/{id}/rerun-checks", h.RerunChecks)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Get("/change-requests/{id}/diff", h.Diff)

		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestReview)).Get("/labels", h.ListLabels)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCreate)).Post("/labels", h.UpsertLabel)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCreate)).Put("/labels/{id}", h.UpsertLabel)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCreate)).Post("/change-requests/{id}/labels", h.AddLabel)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCreate)).Delete("/change-requests/{id}/labels/{labelId}", h.RemoveLabel)

		r.With(middleware.RequirePermission(pc, permcode.CodeRolesView)).Get("/roles", h.ListRoles)
		r.With(middleware.RequirePermission(pc, permcode.CodeRolesView)).Get("/roles/{id}", h.GetRole)
		r.With(middleware.RequirePermission(pc, permcode.CodeRolesView)).Get("/role-assignment-policies", h.RoleAssignmentPolicies)
		r.With(middleware.RequirePermission(pc, permcode.CodeChangeRequestCreate)).Post("/users/{userId}/role-assignment-request", h.RoleAssignmentRequest)

		r.With(middleware.RequirePermission(pc, permcode.CodeUsersView)).Get("/users", h.ListUsers)
		r.With(middleware.RequirePermission(pc, permcode.CodeUsersView)).Get("/users/{id}", h.GetUser)
		r.With(middleware.RequirePermission(pc, permcode.CodeGroupsView)).Get("/groups", h.ListGroups)
		r.With(middleware.RequirePermission(pc, permcode.CodeGroupsView)).Get("/groups/{id}", h.GetGroup)

		r.With(middleware.RequirePermission(pc, permcode.CodeFunctionRightsView)).Get("/function-definitions", h.FunctionDefinitions)
		r.With(middleware.RequirePermission(pc, permcode.CodeFunctionRightsView)).Get("/function-rights", h.FunctionRights)
		r.With(middleware.RequirePermission(pc, permcode.CodeUsersView)).Get("/effective/users/{userId}", h.EffectivePermissions)
		r.With(middleware.RequirePermission(pc, permcode.CodeDataRightsView)).Get("/data-rights", h.DataRights)

		r.With(middleware.RequirePermission(pc, permcode.CodeApprovalSettingsView)).Get("/approval-settings", h.ApprovalSettings)
		r.With(middleware.RequirePermission(pc, permcode.CodeApprovalSettingsView)).Get("/approval-settings/{id}", h.ApprovalSetting)

		r.With(middleware.RequirePermission(pc, permcode.CodeNotificationView)).Get("/notification-settings", h.NotificationSettings)
		r.With(middleware.RequirePermission(pc, permcode.CodeNotificationEdit)).Put("/notification-settings", h.UpdateNotificationSettings)
	})

	r.With(middleware.RequirePermission(pc, permcode.CodeAuditView)).Get("/audit/logs", h.AuditLogs)
	r.With(middleware.RequirePermission(pc, permcode.CodeAuditExport)).Get("/audit/logs/export", h.AuditExport)
}
