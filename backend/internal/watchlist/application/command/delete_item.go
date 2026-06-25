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

// DeleteItemHandler soft-deletes a watchlist item.
type DeleteItemHandler struct {
	pool          *pgxpool.Pool
	items         repository.WatchlistItemRepository
	portfolioPort watchlistdomain.PortfolioScopePort
	iamChecker    contract.PermissionChecker
	auditRecorder watchlistdomain.WatchlistAuditRecorder
}

func NewDeleteItemHandler(
	pool *pgxpool.Pool,
	items repository.WatchlistItemRepository,
	portfolioPort watchlistdomain.PortfolioScopePort,
	iamChecker contract.PermissionChecker,
	auditRecorder watchlistdomain.WatchlistAuditRecorder,
) *DeleteItemHandler {
	return &DeleteItemHandler{
		pool:          pool,
		items:         items,
		portfolioPort: portfolioPort,
		iamChecker:    iamChecker,
		auditRecorder: auditRecorder,
	}
}

func (h *DeleteItemHandler) Handle(ctx context.Context, itemID uuid.UUID, actorID uuid.UUID) error {
	item, err := h.items.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("loading item: %w", err)
	}
	if item == nil {
		return watchlistdomain.ErrItemNotFound
	}

	// Scope enforcement.
	if err := h.checkAccess(ctx, item, actorID); err != nil {
		return err
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := h.items.SoftDelete(ctx, tx, itemID, actorID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete item: %w", err)
	}

	actorRef := actorID
	h.auditRecorder.Record(ctx, &actorRef, "WATCHLIST_ITEM_DELETED", "watchlist_item", itemID.String(), "", "", map[string]interface{}{
		"scope_type": string(item.ScopeType),
	})
	return nil
}

func (h *DeleteItemHandler) checkAccess(ctx context.Context, item *entity.WatchlistItem, actorID uuid.UUID) error {
	switch item.ScopeType {
	case entity.ScopePersonal:
		if item.OwnerUserID == nil || *item.OwnerUserID != actorID {
			return watchlistdomain.ErrForbiddenScope
		}
	case entity.ScopePortfolio:
		if item.PortfolioID == nil {
			return watchlistdomain.ErrForbiddenScope
		}
		scopeInfo, err := h.portfolioPort.GetPortfolioScope(ctx, *item.PortfolioID)
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
