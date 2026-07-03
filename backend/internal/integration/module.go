package integration

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/infrastructure/persistence"
	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/transport"
	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/transport/handler"
)

// Module wires the integration dashboard aggregator.
type Module struct {
	dashboardHandler *handler.DashboardHandler
}

// NewModule creates a new integration module with its dependencies.
func NewModule(pool *pgxpool.Pool, iam query.IAMPort) *Module {
	repo := persistence.NewTaskRepository(pool)

	snapshotHandler := query.NewGetDashboardSnapshotHandler(iam, repo)
	tasksHandler := query.NewGetMyTasksHandler(iam, repo)

	dashboardHandler := handler.NewDashboardHandler(snapshotHandler, tasksHandler)

	return &Module{
		dashboardHandler: dashboardHandler,
	}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(r chi.Router) {
	transport.RegisterRoutes(r, m.dashboardHandler)
}
