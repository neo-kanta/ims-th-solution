package adapter

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// PortfolioScopeAdapter adapts contract.PortfolioScopeResolver to the watchlist PortfolioScopePort.
type PortfolioScopeAdapter struct {
	resolver contract.PortfolioScopeResolver
}

func NewPortfolioScopeAdapter(resolver contract.PortfolioScopeResolver) *PortfolioScopeAdapter {
	return &PortfolioScopeAdapter{resolver: resolver}
}

func (a *PortfolioScopeAdapter) GetPortfolioScope(ctx context.Context, portfolioID uuid.UUID) (*domain.PortfolioScopeInfo, error) {
	info, err := a.resolver.GetPortfolioScope(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("portfolio scope: %w", err)
	}
	if info == nil {
		return nil, nil
	}
	return &domain.PortfolioScopeInfo{
		PortfolioID:   info.PortfolioID,
		PortfolioCode: info.PortfolioCode,
		PortfolioName: info.PortfolioName,
		FundID:        info.FundID,
		FundCode:      info.FundCode,
		FundName:      info.FundName,
	}, nil
}
