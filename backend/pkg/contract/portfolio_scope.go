package contract

import (
	"context"

	"github.com/google/uuid"
)

// PortfolioScopeInfo carries the fields needed for IAM data-permission checks
// and portfolio descriptor hydration in the watchlist module.
type PortfolioScopeInfo struct {
	PortfolioID   uuid.UUID
	PortfolioCode string
	PortfolioName string
	FundID        uuid.UUID
	FundCode      string
	FundName      string
}

// PortfolioScopeResolver resolves a portfolio ID to its IAM data scope and
// descriptor fields. Implemented by the investment module.
type PortfolioScopeResolver interface {
	// GetPortfolioScope returns scope info for the given portfolio. Returns nil,
	// nil when the portfolio does not exist.
	GetPortfolioScope(ctx context.Context, portfolioID uuid.UUID) (*PortfolioScopeInfo, error)
}
