package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ListItemsInput filters and paginates watchlist items.
type ListItemsInput struct {
	ActorID         uuid.UUID
	ScopeType       *entity.ScopeType
	PortfolioID     *uuid.UUID
	SecurityID      *uuid.UUID
	IncludeDisabled bool
	Limit           int
	Offset          int
}

// ListItemsOutput wraps the paginated result.
type ListItemsOutput struct {
	Items  []*entity.WatchlistItem
	Total  int
	Limit  int
	Offset int
}

// ListItemsHandler returns items visible to the caller after scope enforcement.
type ListItemsHandler struct {
	items         repository.WatchlistItemRepository
	portfolioPort watchlistdomain.PortfolioScopePort
	iamChecker    contract.PermissionChecker
}

func NewListItemsHandler(
	items repository.WatchlistItemRepository,
	portfolioPort watchlistdomain.PortfolioScopePort,
	iamChecker contract.PermissionChecker,
) *ListItemsHandler {
	return &ListItemsHandler{
		items:         items,
		portfolioPort: portfolioPort,
		iamChecker:    iamChecker,
	}
}

func (h *ListItemsHandler) Handle(ctx context.Context, input ListItemsInput) (*ListItemsOutput, error) {
	if input.PortfolioID != nil && (input.ScopeType == nil || *input.ScopeType != entity.ScopePortfolio) {
		return nil, fmt.Errorf("%w: portfolio_id requires scope_type=PORTFOLIO", watchlistdomain.ErrInvalidQuery)
	}

	filter := repository.WatchlistItemFilter{
		ScopeType:       input.ScopeType,
		SecurityID:      input.SecurityID,
		IncludeDisabled: input.IncludeDisabled,
		Limit:           input.Limit,
		Offset:          input.Offset,
	}

	// Unscoped: return personal items owned by actor UNION portfolio items accessible via funds.
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

	// Portfolio scope: resolve accessible portfolios.
	if input.ScopeType != nil && *input.ScopeType == entity.ScopePortfolio {
		if input.PortfolioID != nil {
			// Single portfolio — check data permission.
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
			filter.OwnerUserID = nil // not used for portfolio scope
		} else {
			// All portfolios the actor can access.
			contracts, err := h.iamChecker.GetAccessibleContracts(input.ActorID)
			if err != nil {
				return nil, fmt.Errorf("loading accessible contracts: %w", err)
			}
			if len(contracts) == 0 {
				return &ListItemsOutput{Items: []*entity.WatchlistItem{}, Limit: filter.Limit, Offset: filter.Offset}, nil
			}
			// "*" means all access — no portfolio ID restriction.
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
					return &ListItemsOutput{Items: []*entity.WatchlistItem{}, Limit: filter.Limit, Offset: filter.Offset}, nil
				}
				// watchlist_items.portfolio_id != fund_id; the repository joins portfolios
				// to restrict results to portfolios belonging to accessible funds.
				filter.FundIDs = fundIDs
			}
			filter.OwnerUserID = nil
		}
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	items, total, err := h.items.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing items: %w", err)
	}
	return &ListItemsOutput{
		Items:  items,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}
