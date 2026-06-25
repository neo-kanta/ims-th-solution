// Package watchlist implements the Watchlist module: personal and portfolio
// price-alert watchlists backed by market-data threshold rules.
//
// Cross-module dependencies (all consumed through pkg/contract interfaces or
// narrow domain ports — this module never imports another module's internal/):
//   - reference_data: SecurityResolver → SecurityPort
//   - market_data: MarketQuoteProvider → QuotePort
//   - investment: PortfolioScopeResolver → PortfolioScopePort
//   - notification: WatchlistAlertNotifier (implemented by notification module)
//   - audit: Recorder → WatchlistAuditRecorder
//   - iam: PermissionChecker (via PermissionCheckerAdapter)
package watchlist

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/application/query"
	appsvc "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/application/service"
	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	wladapter "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/infrastructure/persistence"
	wlhttp "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/transport/http"
	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	refdatadomain "github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Module owns the watchlist feature: persistence, application services and HTTP transport.
type Module struct {
	handler *wlhttp.Handler
}

// Dependencies bundles everything the watchlist module needs from the outside world.
type Dependencies struct {
	Pool             *pgxpool.Pool
	SecurityResolver refdatadomain.SecurityResolver
	QuoteProvider    contract.MarketQuoteProvider
	PortfolioScope   contract.PortfolioScopeResolver
	Notifier         watchlistdomain.WatchlistAlertNotifier
	AuditRecorder    auditdomain.Recorder
	IAM              wladapter.IAMPermissionPort
}

// NewModule wires the watchlist module.
func NewModule(deps Dependencies) *Module {
	// Repositories.
	items := persistence.NewPostgresWatchlistItemRepository(deps.Pool)
	rules := persistence.NewPostgresThresholdRuleRepository(deps.Pool)
	alerts := persistence.NewPostgresAlertEventRepository(deps.Pool)

	// Port adapters.
	securityPort := wladapter.NewSecurityAdapter(deps.SecurityResolver)
	quotePort := wladapter.NewQuoteAdapter(deps.QuoteProvider)
	portfolioPort := wladapter.NewPortfolioScopeAdapter(deps.PortfolioScope)
	auditRecorder := wladapter.NewAuditAdapter(deps.AuditRecorder)
	iamChecker := wladapter.NewPermissionCheckerAdapter(deps.IAM)

	// Application layer.
	createItemH := command.NewCreateItemHandler(deps.Pool, items, rules, securityPort, portfolioPort, iamChecker, auditRecorder)
	updateItemH := command.NewUpdateItemHandler(deps.Pool, items, rules, portfolioPort, iamChecker, auditRecorder)
	deleteItemH := command.NewDeleteItemHandler(deps.Pool, items, portfolioPort, iamChecker, auditRecorder)
	ackAlertH := command.NewAcknowledgeAlertHandler(deps.Pool, alerts, portfolioPort, iamChecker, auditRecorder)
	listItemsH := query.NewListItemsHandler(items, portfolioPort, iamChecker)
	listAlertsH := query.NewListAlertsHandler(alerts, portfolioPort, iamChecker)
	evaluatorSvc := appsvc.NewEvaluatorService(
		deps.Pool, rules, items, alerts,
		securityPort, quotePort, portfolioPort,
		deps.Notifier, auditRecorder, iamChecker,
	)

	// User lookup (read-only projection from iam_users; no module import).
	userLookup := wladapter.NewPostgresUserLookupAdapter(deps.Pool)

	// Transport.
	h := wlhttp.NewHandler(
		createItemH, updateItemH, deleteItemH, ackAlertH,
		listItemsH, listAlertsH,
		evaluatorSvc,
		rules,
		securityPort, quotePort, portfolioPort,
		userLookup,
	)

	return &Module{handler: h}
}

// RegisterRoutes mounts the watchlist routes onto the authenticated router.
func (m *Module) RegisterRoutes(r chi.Router, pc middleware.PermissionChecker) {
	wlhttp.RegisterRoutes(r, m.handler, pc)
}
