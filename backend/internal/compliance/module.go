// Package compliance wires the IRG (Investment Regulation Guard) module.
//
// Blank imports below trigger each rule package's init() → spi.Register(),
// loading all rule evaluators into the global registry before the engine starts.
package compliance

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/engine"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/infrastructure/persistence"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/transport"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/transport/handler"
	platformmw "github.com/neo-kanta/ims-th-solution/backend/platform/middleware"

	// Self-registering rule packages — must be blank-imported to run init().
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/cash"
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/concentration"
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/quantity"
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/ratio"
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/restriction"
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/stub"
)

// Module owns the IRG compliance pipeline, rule administration, and breach management.
type Module struct {
	handler          *handler.ComplianceHandler
	registry         *spi.RuleRegistry
	permissionChecker platformmw.PermissionChecker
}

// NewModule constructs the full compliance module with Postgres-backed repositories
// and Nop port adapters suitable for the PoC. Swap the Nop adapters for real
// implementations (backed by market_data / integration modules) in production.
func NewModule(pool *pgxpool.Pool, permChecker platformmw.PermissionChecker) *Module {
	// Repositories
	instanceRepo := persistence.NewPostgresRuleInstanceRepository(pool)
	bindingRepo  := persistence.NewPostgresRuleBindingRepository(pool)
	checkRepo    := persistence.NewPostgresCheckRecordRepository(pool)
	breachRepo   := persistence.NewPostgresBreachRepository(pool)
	overrideRepo := persistence.NewPostgresOverrideRepository(pool)

	// Port adapters (Nop stubs for PoC)
	fetcher := engine.NewFetcher(
		&adapter.NopPositionSnapshotAdapter{},
		&adapter.NopMarketDataAdapter{},
		&adapter.NopInstrumentClassificationAdapter{},
		&adapter.NopCreditRatingAdapter{},
		&adapter.NopRestrictionListAdapter{},
		&adapter.NopTradeHistoryAdapter{},
		&adapter.NopCalendarAdapter{},
		&adapter.NopPortfolioMetadataAdapter{},
	)

	// Engine
	registry := spi.GlobalRegistry()
	pipeline := engine.NewPipeline(registry, bindingRepo, checkRepo, breachRepo, fetcher)

	// Application layer — commands
	preTradeCmd  := command.NewRunPreTradeCheckHandler(pipeline, registry)
	postTradeCmd := command.NewRunPostTradeCheckHandler(pipeline, registry)
	overrideCmd  := command.NewOverrideBreachHandler(overrideRepo)

	// Application layer — queries
	checkGroupQry    := query.NewGetCheckGroupHandler(checkRepo, breachRepo)
	listBreachesQry  := query.NewListBreachesHandler(breachRepo)
	listInstancesQry := query.NewListRuleInstancesHandler(instanceRepo, registry)

	// Transport
	h := handler.NewComplianceHandler(
		preTradeCmd,
		postTradeCmd,
		overrideCmd,
		checkGroupQry,
		listBreachesQry,
		listInstancesQry,
	)

	return &Module{
		handler:           h,
		registry:          registry,
		permissionChecker: permChecker,
	}
}

// RegisterRoutes mounts compliance API routes onto the given authenticated router.
// Expected to be called inside an r.Route("/api/v1", ...) block that already
// has auth middleware applied.
func (m *Module) RegisterRoutes(r chi.Router) {
	if m == nil || m.handler == nil {
		return
	}

	// Permission-gated sub-routes.
	if m.permissionChecker != nil {
		r.Group(func(r chi.Router) {
			r.Use(platformmw.RequirePermission(m.permissionChecker, "IRG_VIEW_RULES"))
			r.Get("/compliance/rules", m.handler.ListRuleInstances)
			r.Get("/compliance/checks/{groupID}", m.handler.GetCheckGroup)
			r.Get("/compliance/breaches", m.handler.ListBreaches)
		})
		r.Group(func(r chi.Router) {
			r.Use(platformmw.RequirePermission(m.permissionChecker, "IRG_OVERRIDE_BREACH"))
			r.Post("/compliance/breaches/{breachID}/override", m.handler.OverrideBreach)
		})
		// Pre/post-trade checks are invoked by the OMS layer (WORKFLOW_EXECUTE permission).
		r.Group(func(r chi.Router) {
			r.Post("/compliance/checks/pre-trade", m.handler.RunPreTradeCheck)
			r.Post("/compliance/checks/post-trade", m.handler.RunPostTradeCheck)
		})
	} else {
		// Fallback: no permission enforcement (dev/test mode).
		transport.RegisterRoutes(r, m.handler)
	}
}

// Registry returns the live SPI rule registry (used by the OMS module for internal calls).
func (m *Module) Registry() *spi.RuleRegistry {
	return m.registry
}
