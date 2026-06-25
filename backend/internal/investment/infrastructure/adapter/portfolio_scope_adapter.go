package adapter

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// PortfolioScopeAdapter implements contract.PortfolioScopeResolver by resolving
// a portfolio ID to its fund and returning the descriptor fields needed by the
// watchlist module for IAM data-permission checks and response hydration.
type PortfolioScopeAdapter struct {
	portfolios domain.PortfolioRepository
	funds      domain.FundRepository
}

// NewPortfolioScopeAdapter wires the adapter.
func NewPortfolioScopeAdapter(portfolios domain.PortfolioRepository, funds domain.FundRepository) *PortfolioScopeAdapter {
	return &PortfolioScopeAdapter{portfolios: portfolios, funds: funds}
}

// GetPortfolioScope returns scope info for the given portfolio.
// Returns nil, nil when the portfolio or its parent fund does not exist.
func (a *PortfolioScopeAdapter) GetPortfolioScope(ctx context.Context, portfolioID uuid.UUID) (*contract.PortfolioScopeInfo, error) {
	p, err := a.portfolios.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("portfolio scope: portfolio lookup: %w", err)
	}
	if p == nil {
		return nil, nil
	}
	f, err := a.funds.GetByID(ctx, p.FundID)
	if err != nil {
		return nil, fmt.Errorf("portfolio scope: fund lookup: %w", err)
	}
	if f == nil {
		return nil, nil
	}
	return &contract.PortfolioScopeInfo{
		PortfolioID:   p.ID,
		PortfolioCode: p.Code,
		PortfolioName: p.Name,
		FundID:        f.ID,
		FundCode:      f.Code,
		FundName:      f.Name,
	}, nil
}

var _ contract.PortfolioScopeResolver = (*PortfolioScopeAdapter)(nil)
