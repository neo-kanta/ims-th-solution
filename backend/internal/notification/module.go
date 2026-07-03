// Package notification owns the in-app notification center, email outbox,
// SMTP delivery, and background email worker.
package notification

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	approvaldomain "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/infrastructure/adapter"
	emailinfra "github.com/neo-kanta/ims-th-solution/backend/internal/notification/infrastructure/email"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/infrastructure/persistence"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/jobs"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/transport"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/transport/handler"
	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	workflowports "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Module wires the notification module dependencies.
type Module struct {
	repo                   *persistence.PostgresNotificationRepository
	svc                    *service.Service
	emailSvc               *service.EmailOutboxService
	handler                *handler.Handler
	emailHandler           *handler.EmailOutboxHandler
	notifier               *adapter.ApprovalNotifier
	stuckDayNotifier       *adapter.WorkflowStuckDayNotifier
	watchlistAlertNotifier *adapter.WatchlistAlertNotifier
	worker                 *jobs.EmailOutboxWorker
	checker                middleware.PermissionChecker
}

// NewModule creates the notification module. cfg and checker may both be nil for
// the in-app-only baseline (no email, no admin routes). When cfg has email
// enabled, the worker is wired but not started; call StartWorker(ctx) after
// the server routes are mounted.
func NewModule(pool *pgxpool.Pool, cfg *config.AppConfig, checker middleware.PermissionChecker) *Module {
	repo := persistence.NewPostgresNotificationRepository(pool)
	svc := service.NewService(repo)
	m := &Module{
		repo:                   repo,
		svc:                    svc,
		handler:                handler.NewHandler(svc),
		notifier:               adapter.NewApprovalNotifier(svc),
		stuckDayNotifier:       adapter.NewWorkflowStuckDayNotifier(svc),
		watchlistAlertNotifier: adapter.NewWatchlistAlertNotifier(svc),
		checker:                checker,
	}

	if cfg == nil || !cfg.EmailEnabled {
		return m
	}

	outboxRepo := persistence.NewPostgresEmailOutboxRepository(pool)
	templateSvc := service.NewEmailTemplateService(cfg.AppPublicBaseURL)

	svc.WithEmail(outboxRepo, templateSvc, service.EmailConfig{
		Enabled:        cfg.EmailEnabled,
		MaxAttempts:    cfg.EmailMaxAttempts,
		AllowedDomains: cfg.EmailAllowedDomains,
	})

	emailSvc := service.NewEmailOutboxService(outboxRepo, repo, templateSvc)
	emailHandler := handler.NewEmailOutboxHandler(emailSvc, cfg)

	m.emailSvc = emailSvc
	m.emailHandler = emailHandler

	if cfg.EmailWorkerEnabled {
		sender := emailinfra.NewSMTPSender(cfg)
		m.worker = jobs.NewEmailOutboxWorker(
			outboxRepo,
			sender,
			cfg.EmailWorkerInterval,
			cfg.EmailWorkerBatchSize,
			cfg.EmailStaleSendingTimeout,
			cfg.EmailRetryPolicy,
		)
	}

	return m
}

// StartWorker starts the email outbox background worker. Call after routing is
// set up. No-op when email or the worker is not configured.
func (m *Module) StartWorker(ctx context.Context) {
	if m == nil || m.worker == nil {
		return
	}
	go m.worker.Start(ctx)
}

// RegisterRoutes mounts the recipient-side notification endpoints and, when
// the email handler is configured, the admin/demo email endpoints.
func (m *Module) RegisterRoutes(r chi.Router) {
	if m == nil || m.handler == nil {
		return
	}

	emailH := m.emailHandler
	checker := m.checker
	if checker == nil || emailH == nil {
		// No permission checker or no email handler — mount user-facing routes only.
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", m.handler.List)
			r.Post("/{id}/read", m.handler.MarkRead)
			r.Post("/read-all", m.handler.MarkAllRead)
		})
		return
	}

	transport.RegisterRoutes(r, m.handler, emailH, checker)
}

// ApprovalNotifier returns the adapter the approval module wires as its
// domain.Notifier.
func (m *Module) ApprovalNotifier() approvaldomain.Notifier {
	if m == nil {
		return nil
	}
	return m.notifier
}

// WorkflowStuckDayNotifier returns the adapter the workflow module wires as
// its ports.OperatorNotifier.
func (m *Module) WorkflowStuckDayNotifier() workflowports.OperatorNotifier {
	if m == nil {
		return nil
	}
	return m.stuckDayNotifier
}

// WatchlistAlertNotifier returns the adapter the watchlist module uses to
// deliver threshold-breach notifications.
func (m *Module) WatchlistAlertNotifier() watchlistdomain.WatchlistAlertNotifier {
	if m == nil {
		return nil
	}
	return m.watchlistAlertNotifier
}
