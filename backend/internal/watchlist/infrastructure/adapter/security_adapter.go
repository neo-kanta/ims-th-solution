package adapter

import (
	"context"
	"fmt"

	refdata "github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
)

// SecurityAdapter adapts reference_data's SecurityResolver to the watchlist SecurityPort.
type SecurityAdapter struct {
	resolver refdata.SecurityResolver
}

func NewSecurityAdapter(resolver refdata.SecurityResolver) *SecurityAdapter {
	return &SecurityAdapter{resolver: resolver}
}

func (a *SecurityAdapter) GetSecurityByID(ctx context.Context, id string) (*domain.SecurityInfo, error) {
	sec, err := a.resolver.GetSecurityByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("security lookup: %w", err)
	}
	if sec == nil {
		return nil, nil
	}
	return &domain.SecurityInfo{
		ID:            sec.ID,
		IMSSymbol:     sec.IMSSymbol,
		DisplaySymbol: sec.DisplaySymbol,
		Name:          sec.Name,
		AssetType:     string(sec.AssetType),
		Currency:      sec.Currency,
		ExchangeMIC:   sec.ExchangeMIC,
		Status:        string(sec.Status),
	}, nil
}

func (a *SecurityAdapter) ResolveProviderSymbol(ctx context.Context, securityID, providerCode string) (*domain.ProviderMapping, error) {
	pm, err := a.resolver.ResolveProviderSymbol(ctx, securityID, providerCode)
	if err != nil {
		return nil, fmt.Errorf("resolve provider symbol: %w", err)
	}
	if pm == nil {
		return nil, nil
	}
	return &domain.ProviderMapping{
		ProviderCode:   pm.ProviderCode,
		ProviderSymbol: pm.ProviderSymbol,
	}, nil
}
