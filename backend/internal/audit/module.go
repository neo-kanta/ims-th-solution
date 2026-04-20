package audit

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/application/query"
	auditservice "github.com/neo-kanta/ims-th-solution/backend/internal/audit/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/infrastructure/persistence"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/transport/handler"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Module owns audit trail recording, querying, and export concerns.
type Module struct {
	recorder     domain.Recorder
	auditHandler *handler.AuditHandler
}

// NewModule creates the audit module with Postgres-backed persistence.
func NewModule(pool *pgxpool.Pool) *Module {
	repo := persistence.NewPostgresAuditRepository(pool)
	recorder := auditservice.NewRecorder(repo)
	listAuditQry := query.NewListAuditEventsQuery(repo)
	auditHandler := handler.NewAuditHandler(listAuditQry, recorder)

	return &Module{
		recorder:     recorder,
		auditHandler: auditHandler,
	}
}

// Recorder returns the audit writer port used by other modules.
func (m *Module) Recorder() domain.Recorder {
	if m == nil || m.recorder == nil {
		return domain.NopRecorder{}
	}
	return m.recorder
}

// RegisterAdminRoutes mounts audit query/export endpoints under an existing admin router.
func (m *Module) RegisterAdminRoutes(
	r chi.Router,
	permissionChecker middleware.PermissionChecker,
	exportLimiter middleware.RateLimiter,
	exportPolicy middleware.RateLimitPolicy,
) {
	if m == nil || m.auditHandler == nil || permissionChecker == nil {
		return
	}

	r.With(middleware.RequirePermission(permissionChecker, "IAM_AUDIT_VIEW")).Get("/audit", m.auditHandler.ListAuditEvents)

	if exportLimiter == nil {
		r.With(middleware.RequirePermission(permissionChecker, "IAM_AUDIT_VIEW")).Get("/audit/export", m.auditHandler.ExportAuditCSV)
		return
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitByUser(exportLimiter, exportPolicy))
		r.With(middleware.RequirePermission(permissionChecker, "IAM_AUDIT_VIEW")).Get("/audit/export", m.auditHandler.ExportAuditCSV)
	})
}
