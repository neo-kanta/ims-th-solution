package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// AcknowledgeAlertHandler acknowledges an alert event.
type AcknowledgeAlertHandler struct {
	pool          *pgxpool.Pool
	alerts        repository.AlertEventRepository
	portfolioPort watchlistdomain.PortfolioScopePort
	iamChecker    contract.PermissionChecker
	auditRecorder watchlistdomain.WatchlistAuditRecorder
}

func NewAcknowledgeAlertHandler(
	pool *pgxpool.Pool,
	alerts repository.AlertEventRepository,
	portfolioPort watchlistdomain.PortfolioScopePort,
	iamChecker contract.PermissionChecker,
	auditRecorder watchlistdomain.WatchlistAuditRecorder,
) *AcknowledgeAlertHandler {
	return &AcknowledgeAlertHandler{
		pool:          pool,
		alerts:        alerts,
		portfolioPort: portfolioPort,
		iamChecker:    iamChecker,
		auditRecorder: auditRecorder,
	}
}

func (h *AcknowledgeAlertHandler) Handle(ctx context.Context, alertID uuid.UUID, actorID uuid.UUID, note *string) (*entity.AlertEvent, error) {
	alert, err := h.alerts.GetByID(ctx, alertID)
	if err != nil {
		return nil, fmt.Errorf("loading alert: %w", err)
	}
	if alert == nil {
		return nil, watchlistdomain.ErrAlertNotFound
	}

	// Scope enforcement.
	if err := h.checkAccess(ctx, alert, actorID); err != nil {
		return nil, err
	}

	if alert.IsAcknowledged() {
		return nil, watchlistdomain.ErrAlertAlreadyAcknowledged
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := h.alerts.Acknowledge(ctx, tx, alertID, actorID, note); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit acknowledgement: %w", err)
	}

	// Reload to get updated fields.
	updated, err := h.alerts.GetByID(ctx, alertID)
	if err != nil {
		return nil, fmt.Errorf("reloading alert: %w", err)
	}

	actorRef := actorID
	h.auditRecorder.Record(ctx, &actorRef, "WATCHLIST_ALERT_ACKNOWLEDGED", "watchlist_alert_event", alertID.String(), "", "", map[string]interface{}{
		"alert_id": alertID.String(),
	})

	return updated, nil
}

func (h *AcknowledgeAlertHandler) checkAccess(ctx context.Context, alert *entity.AlertEvent, actorID uuid.UUID) error {
	switch alert.ScopeType {
	case entity.ScopePersonal:
		if alert.OwnerUserID == nil || *alert.OwnerUserID != actorID {
			return watchlistdomain.ErrForbiddenScope
		}
	case entity.ScopePortfolio:
		if alert.PortfolioID == nil {
			return watchlistdomain.ErrForbiddenScope
		}
		scopeInfo, err := h.portfolioPort.GetPortfolioScope(ctx, *alert.PortfolioID)
		if err != nil || scopeInfo == nil {
			return watchlistdomain.ErrForbiddenScope
		}
		ok, err := h.iamChecker.HasDataPermission(actorID, scopeInfo.FundID.String())
		if err != nil || !ok {
			return watchlistdomain.ErrForbiddenScope
		}
	}
	return nil
}
