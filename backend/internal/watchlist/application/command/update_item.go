package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// txStarter abstracts pgxpool.Pool for unit-testability.
type txStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// UpdateItemInput is the input to UpdateItemHandler.
type UpdateItemInput struct {
	ItemID         uuid.UUID
	Pinned         *bool
	Note           *string
	Status         *entity.ItemStatus
	ThresholdRules *[]ThresholdRuleInput // nil = unchanged
	ActorID        uuid.UUID
}

// UpdateItemOutput holds the updated item and active rules.
type UpdateItemOutput struct {
	Item  *entity.WatchlistItem
	Rules []*entity.ThresholdRule
}

// UpdateItemHandler updates item metadata and replaces threshold rules.
type UpdateItemHandler struct {
	pool          txStarter
	items         repository.WatchlistItemRepository
	rules         repository.ThresholdRuleRepository
	portfolioPort watchlistdomain.PortfolioScopePort
	iamChecker    contract.PermissionChecker
	auditRecorder watchlistdomain.WatchlistAuditRecorder
}

func NewUpdateItemHandler(
	pool txStarter,
	items repository.WatchlistItemRepository,
	rules repository.ThresholdRuleRepository,
	portfolioPort watchlistdomain.PortfolioScopePort,
	iamChecker contract.PermissionChecker,
	auditRecorder watchlistdomain.WatchlistAuditRecorder,
) *UpdateItemHandler {
	return &UpdateItemHandler{
		pool:          pool,
		items:         items,
		rules:         rules,
		portfolioPort: portfolioPort,
		iamChecker:    iamChecker,
		auditRecorder: auditRecorder,
	}
}

func (h *UpdateItemHandler) Handle(ctx context.Context, input UpdateItemInput) (*UpdateItemOutput, error) {
	item, err := h.items.GetByID(ctx, input.ItemID)
	if err != nil {
		return nil, fmt.Errorf("loading item: %w", err)
	}
	if item == nil {
		return nil, watchlistdomain.ErrItemNotFound
	}

	// Scope enforcement.
	if err := h.checkAccess(ctx, item, input.ActorID); err != nil {
		return nil, err
	}

	// Apply field updates.
	actorPtr := &input.ActorID
	item.UpdatedBy = actorPtr
	if input.Pinned != nil {
		item.Pinned = *input.Pinned
	}
	if input.Note != nil {
		item.Note = input.Note
	}
	if input.Status != nil {
		item.Status = *input.Status
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := h.items.Update(ctx, tx, item); err != nil {
		return nil, fmt.Errorf("updating item: %w", err)
	}

	// Rule upsert: when threshold_rules is present it is the full desired set.
	// Rules with an ID reference an existing rule to update in-place;
	// rules without an ID create a new rule; existing rules not referenced are soft-disabled.
	// Captured outside the block so they can be referenced in post-commit audit.
	var activeRules []*entity.ThresholdRule
	var disabledRules []*entity.ThresholdRule
	var createdRules []*entity.ThresholdRule
	var updatedRules []*entity.ThresholdRule

	if input.ThresholdRules != nil {
		existing, err := h.rules.ListByItemID(ctx, item.ID)
		if err != nil {
			return nil, fmt.Errorf("loading existing rules: %w", err)
		}

		existingByID := make(map[uuid.UUID]*entity.ThresholdRule, len(existing))
		for _, r := range existing {
			existingByID[r.ID] = r
		}
		referencedIDs := make(map[uuid.UUID]struct{})

		for i, tr := range *input.ThresholdRules {
			if err := validateThresholdRuleInput(i, tr); err != nil {
				return nil, err
			}

			if tr.ID != nil {
				if ex, ok := existingByID[*tr.ID]; ok {
					// Update existing rule. Reset last_state if crossing parameters change.
					keyChanged := ex.Direction != tr.Direction ||
						!ex.ThresholdValue.Equal(tr.ThresholdValue) ||
						(ex.Currency == nil) != (tr.Currency == nil) ||
						(ex.Currency != nil && tr.Currency != nil && *ex.Currency != *tr.Currency)
					ex.Direction = tr.Direction
					ex.ThresholdValue = tr.ThresholdValue
					ex.Currency = tr.Currency
					ex.CooldownMinutes = tr.CooldownMinutes
					ex.Status = tr.Status
					ex.UpdatedBy = &input.ActorID
					if keyChanged {
						ex.LastState = entity.RuleStateUnknown
					}
					if err := h.rules.Update(ctx, tx, ex); err != nil {
						return nil, err
					}
					updatedRules = append(updatedRules, ex)
					referencedIDs[*tr.ID] = struct{}{}
					continue
				}
				return nil, fmt.Errorf("%w: rules[%d].id does not belong to watchlist item", watchlistdomain.ErrInvalidThreshold, i)
			}

			// No matching existing rule: create new.
			newRule := &entity.ThresholdRule{
				ID:              uuid.New(),
				WatchlistItemID: item.ID,
				MetricType:      entity.MetricTypeMarketPrice,
				Direction:       tr.Direction,
				ThresholdValue:  tr.ThresholdValue,
				Currency:        tr.Currency,
				CooldownMinutes: tr.CooldownMinutes,
				Status:          tr.Status,
				LastState:       entity.RuleStateUnknown,
				CreatedBy:       input.ActorID,
			}
			if err := h.rules.Create(ctx, tx, newRule); err != nil {
				return nil, err
			}
			createdRules = append(createdRules, newRule)
		}

		// Soft-disable existing rules not referenced by any incoming rule.
		for _, r := range existing {
			if _, referenced := referencedIDs[r.ID]; !referenced {
				if err := h.rules.SoftDisable(ctx, tx, r.ID, input.ActorID); err != nil {
					if !errors.Is(err, watchlistdomain.ErrRuleDisabled) {
						return nil, err
					}
				}
				disabledRules = append(disabledRules, r)
			}
		}

		activeRules = append(updatedRules, createdRules...)
	} else {
		activeRules, err = h.rules.ListByItemID(ctx, item.ID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update item: %w", err)
	}

	actorRef := input.ActorID
	now := time.Now().UTC()
	h.auditRecorder.Record(ctx, &actorRef, "WATCHLIST_ITEM_UPDATED", "watchlist_item", item.ID.String(), "", "", map[string]interface{}{
		"event_time": now,
	})
	for _, r := range updatedRules {
		r := r
		h.auditRecorder.Record(ctx, &actorRef, "WATCHLIST_THRESHOLD_UPDATED", "watchlist_threshold_rule", r.ID.String(), "", "", map[string]interface{}{
			"event_time": now,
		})
	}
	for _, r := range createdRules {
		r := r
		h.auditRecorder.Record(ctx, &actorRef, "WATCHLIST_THRESHOLD_CREATED", "watchlist_threshold_rule", r.ID.String(), "", "", map[string]interface{}{
			"event_time": now,
		})
	}
	for _, r := range disabledRules {
		r := r
		h.auditRecorder.Record(ctx, &actorRef, "WATCHLIST_THRESHOLD_DISABLED", "watchlist_threshold_rule", r.ID.String(), "", "", map[string]interface{}{
			"event_time": now,
		})
	}

	return &UpdateItemOutput{Item: item, Rules: activeRules}, nil
}

func (h *UpdateItemHandler) checkAccess(ctx context.Context, item *entity.WatchlistItem, actorID uuid.UUID) error {
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
