package query

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ListAlertsInput filters and paginates alert events.
type ListAlertsInput struct {
	ActorID      uuid.UUID
	ScopeType    *entity.ScopeType
	PortfolioID  *uuid.UUID
	SecurityID   *uuid.UUID
	RuleID       *uuid.UUID
	Acknowledged *bool
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	Limit        int
	Offset       int
}

// ListAlertsOutput wraps the paginated result.
type ListAlertsOutput struct {
	Items  []*entity.AlertEvent
	Total  int
	Limit  int
	Offset int
}

// ListAlertsHandler returns alert events visible to the caller.
type ListAlertsHandler struct {
	alerts        repository.AlertEventRepository
	portfolioPort watchlistdomain.PortfolioScopePort
	iamChecker    contract.PermissionChecker
}

func NewListAlertsHandler(
	alerts repository.AlertEventRepository,
	portfolioPort watchlistdomain.PortfolioScopePort,
	iamChecker contract.PermissionChecker,
) *ListAlertsHandler {
	return &ListAlertsHandler{
		alerts:        alerts,
		portfolioPort: portfolioPort,
		iamChecker:    iamChecker,
	}
}

func (h *ListAlertsHandler) Handle(ctx context.Context, input ListAlertsInput) (*ListAlertsOutput, error) {
	if input.PortfolioID != nil && (input.ScopeType == nil || *input.ScopeType != entity.ScopePortfolio) {
		return nil, fmt.Errorf("%w: portfolio_id requires scope_type=PORTFOLIO", watchlistdomain.ErrInvalidQuery)
	}

	filter := repository.AlertEventFilter{
		SecurityID:   input.SecurityID,
		RuleID:       input.RuleID,
		ScopeType:    input.ScopeType,
		Acknowledged: input.Acknowledged,
		CreatedFrom:  input.CreatedFrom,
		CreatedTo:    input.CreatedTo,
		Limit:        input.Limit,
		Offset:       input.Offset,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Unscoped: return personal alerts owned by actor UNION portfolio alerts accessible via funds.
	if input.ScopeType == nil {
		actorCopy := input.ActorID
		contracts, err := h.iamChecker.GetAccessibleContracts(input.ActorID)
		if err != nil {
			return nil, fmt.Errorf("loading accessible contracts: %w", err)
		}
		hasWildcard := false
		var fundIDs []uuid.UUID
		for _, c := range contracts {
			if c == "*" {
				hasWildcard = true
				break
			}
			if id, err := uuid.Parse(c); err == nil {
				fundIDs = append(fundIDs, id)
			}
		}
		filter.UnionPersonalOwnerID = &actorCopy
		if hasWildcard {
			filter.UnionPortfolioAll = true
		} else {
			filter.FundIDs = fundIDs
		}
	}

	// Personal scope (explicit): restrict to actor's own rows.
	if input.ScopeType != nil && *input.ScopeType == entity.ScopePersonal {
		actorCopy := input.ActorID
		filter.OwnerUserID = &actorCopy
	}

	// Portfolio scope enforcement.
	if input.ScopeType != nil && *input.ScopeType == entity.ScopePortfolio {
		filter.OwnerUserID = nil
		if input.PortfolioID != nil {
			scopeInfo, err := h.portfolioPort.GetPortfolioScope(ctx, *input.PortfolioID)
			if err != nil || scopeInfo == nil {
				return nil, watchlistdomain.ErrForbiddenScope
			}
			ok, err := h.iamChecker.HasDataPermission(input.ActorID, scopeInfo.FundID.String())
			if err != nil || !ok {
				return nil, fmt.Errorf("%w: portfolio_id=%s", watchlistdomain.ErrForbiddenScope, input.PortfolioID.String())
			}
			pid := *input.PortfolioID
			filter.PortfolioIDs = []uuid.UUID{pid}
		} else {
			// All portfolios the actor can access — restrict by accessible fund IDs.
			contracts, err := h.iamChecker.GetAccessibleContracts(input.ActorID)
			if err != nil {
				return nil, fmt.Errorf("loading accessible contracts: %w", err)
			}
			if len(contracts) == 0 {
				return &ListAlertsOutput{Items: []*entity.AlertEvent{}, Limit: filter.Limit, Offset: filter.Offset}, nil
			}
			hasWildcard := false
			for _, c := range contracts {
				if c == "*" {
					hasWildcard = true
					break
				}
			}
			if !hasWildcard {
				var fundIDs []uuid.UUID
				for _, c := range contracts {
					if id, err := uuid.Parse(c); err == nil {
						fundIDs = append(fundIDs, id)
					}
				}
				if len(fundIDs) == 0 {
					return &ListAlertsOutput{Items: []*entity.AlertEvent{}, Limit: filter.Limit, Offset: filter.Offset}, nil
				}
				filter.FundIDs = fundIDs
			}
		}
	}

	alerts, total, err := h.alerts.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing alerts: %w", err)
	}
	return &ListAlertsOutput{
		Items:  alerts,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}
