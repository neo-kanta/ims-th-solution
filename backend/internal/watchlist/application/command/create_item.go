package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ThresholdRuleInput carries the user-supplied fields for one threshold rule.
// ID is optional: when set in a PATCH request it references an existing rule to update.
type ThresholdRuleInput struct {
	ID              *uuid.UUID
	Direction       entity.Direction
	ThresholdValue  decimal.Decimal
	Currency        *string
	CooldownMinutes int
	Status          entity.RuleStatus
}

// CreateItemInput is the input to CreateItemHandler.
type CreateItemInput struct {
	ScopeType      entity.ScopeType
	PortfolioID    *uuid.UUID
	SecurityID     uuid.UUID
	Pinned         bool
	Note           *string
	ThresholdRules []ThresholdRuleInput
	ActorID        uuid.UUID
}

// CreateItemOutput holds the created item and its rules.
type CreateItemOutput struct {
	Item  *entity.WatchlistItem
	Rules []*entity.ThresholdRule
}

// CreateItemHandler creates a watchlist item with optional threshold rules.
type CreateItemHandler struct {
	pool          *pgxpool.Pool
	items         repository.WatchlistItemRepository
	rules         repository.ThresholdRuleRepository
	securityPort  watchlistdomain.SecurityPort
	portfolioPort watchlistdomain.PortfolioScopePort
	iamChecker    contract.PermissionChecker
	auditRecorder watchlistdomain.WatchlistAuditRecorder
}

func NewCreateItemHandler(
	pool *pgxpool.Pool,
	items repository.WatchlistItemRepository,
	rules repository.ThresholdRuleRepository,
	securityPort watchlistdomain.SecurityPort,
	portfolioPort watchlistdomain.PortfolioScopePort,
	iamChecker contract.PermissionChecker,
	auditRecorder watchlistdomain.WatchlistAuditRecorder,
) *CreateItemHandler {
	return &CreateItemHandler{
		pool:          pool,
		items:         items,
		rules:         rules,
		securityPort:  securityPort,
		portfolioPort: portfolioPort,
		iamChecker:    iamChecker,
		auditRecorder: auditRecorder,
	}
}

func (h *CreateItemHandler) Handle(ctx context.Context, input CreateItemInput) (*CreateItemOutput, error) {
	// Validate security exists and is active.
	sec, err := h.securityPort.GetSecurityByID(ctx, input.SecurityID.String())
	if err != nil {
		return nil, fmt.Errorf("security lookup: %w", err)
	}
	if sec == nil || sec.Status != "ACTIVE" {
		return nil, watchlistdomain.ErrInvalidSecurity
	}

	// For PORTFOLIO scope, verify data permission.
	if input.ScopeType == entity.ScopePortfolio {
		if input.PortfolioID == nil {
			return nil, fmt.Errorf("%w: portfolio_id required for PORTFOLIO scope", watchlistdomain.ErrForbiddenScope)
		}
		scopeInfo, err := h.portfolioPort.GetPortfolioScope(ctx, *input.PortfolioID)
		if err != nil {
			return nil, fmt.Errorf("portfolio scope: %w", err)
		}
		if scopeInfo == nil {
			return nil, watchlistdomain.ErrForbiddenScope
		}
		ok, err := h.iamChecker.HasDataPermission(input.ActorID, scopeInfo.FundID.String())
		if err != nil || !ok {
			return nil, watchlistdomain.ErrForbiddenScope
		}
	}

	// Validate threshold rules.
	for i, tr := range input.ThresholdRules {
		if err := validateThresholdRuleInput(i, tr); err != nil {
			return nil, err
		}
	}

	// Build item entity.
	itemID := uuid.New()
	var ownerUserID *uuid.UUID
	if input.ScopeType == entity.ScopePersonal {
		ownerUserID = &input.ActorID
	}
	item := &entity.WatchlistItem{
		ID:          itemID,
		ScopeType:   input.ScopeType,
		OwnerUserID: ownerUserID,
		PortfolioID: input.PortfolioID,
		SecurityID:  input.SecurityID,
		Pinned:      input.Pinned,
		Note:        input.Note,
		Status:      entity.ItemStatusActive,
		CreatedBy:   input.ActorID,
	}

	// Build rule entities.
	ruleEntities := make([]*entity.ThresholdRule, 0, len(input.ThresholdRules))
	for _, tr := range input.ThresholdRules {
		cooldown := tr.CooldownMinutes
		if cooldown == 0 {
			cooldown = 60
		}
		status := tr.Status
		if status == "" {
			status = entity.RuleStatusEnabled
		}
		ruleEntities = append(ruleEntities, &entity.ThresholdRule{
			ID:              uuid.New(),
			WatchlistItemID: itemID,
			MetricType:      entity.MetricTypeMarketPrice,
			Direction:       tr.Direction,
			ThresholdValue:  tr.ThresholdValue,
			Currency:        tr.Currency,
			CooldownMinutes: cooldown,
			Status:          status,
			LastState:       entity.RuleStateUnknown,
			CreatedBy:       input.ActorID,
		})
	}

	// Persist in transaction.
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := h.items.Create(ctx, tx, item); err != nil {
		return nil, err
	}
	for _, r := range ruleEntities {
		if err := h.rules.Create(ctx, tx, r); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create item: %w", err)
	}

	// Audit (fire-and-forget).
	now := time.Now().UTC()
	actorID := input.ActorID
	meta := map[string]interface{}{
		"scope_type":  string(input.ScopeType),
		"security_id": input.SecurityID.String(),
		"rules_count": len(ruleEntities),
		"event_time":  now,
	}
	if input.PortfolioID != nil {
		meta["portfolio_id"] = input.PortfolioID.String()
	}
	h.auditRecorder.Record(ctx, &actorID, "WATCHLIST_ITEM_CREATED", "watchlist_item", itemID.String(), "", "", meta)
	for _, r := range ruleEntities {
		rID := r.ID
		h.auditRecorder.Record(ctx, &actorID, "WATCHLIST_THRESHOLD_CREATED", "watchlist_rule", rID.String(), "", "", map[string]interface{}{
			"item_id":         itemID.String(),
			"direction":       string(r.Direction),
			"threshold_value": r.ThresholdValue.String(),
		})
	}

	return &CreateItemOutput{Item: item, Rules: ruleEntities}, nil
}

func validateThresholdRuleInput(i int, tr ThresholdRuleInput) error {
	if tr.Direction != entity.DirectionAbove && tr.Direction != entity.DirectionBelow {
		return fmt.Errorf("%w: rules[%d].direction must be ABOVE or BELOW", watchlistdomain.ErrInvalidThreshold, i)
	}
	if !tr.ThresholdValue.IsPositive() {
		return fmt.Errorf("%w: rules[%d].threshold_value must be > 0", watchlistdomain.ErrInvalidThreshold, i)
	}
	if tr.CooldownMinutes < 0 {
		return fmt.Errorf("%w: rules[%d].cooldown_minutes must be >= 0", watchlistdomain.ErrInvalidThreshold, i)
	}
	if tr.Status != "" && tr.Status != entity.RuleStatusEnabled && tr.Status != entity.RuleStatusDisabled {
		return fmt.Errorf("%w: rules[%d].status must be ENABLED or DISABLED", watchlistdomain.ErrInvalidThreshold, i)
	}
	return nil
}
