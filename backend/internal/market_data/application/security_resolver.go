package application

import (
	"context"

	refdomain "github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
)

// SecurityResolver is the narrow contract market_data needs from
// reference_data. Implementing it on the reference_data application service is
// trivial — but market_data declares only the methods it actually uses, to
// keep the cross-module dependency surface small.
type SecurityResolver interface {
	SearchSecurities(ctx context.Context, filter refdomain.SecuritySearchFilter) ([]refdomain.Security, error)
	GetSecurityByIMSSymbol(ctx context.Context, imsSymbol string) (*refdomain.Security, error)
	GetSecurityByDisplaySymbol(ctx context.Context, displaySymbol string) (*refdomain.Security, error)
	ResolveSecurityByProviderSymbol(ctx context.Context, providerCode, providerSymbol string) (*refdomain.Security, error)
	ResolveProviderSymbol(ctx context.Context, securityID, providerCode string) (*refdomain.ProviderMapping, error)
	CreateUnmappedCandidate(ctx context.Context, candidate refdomain.UnmappedSecurityCandidate) (*refdomain.UnmappedSecurityCandidate, error)
}

// resolvedSymbol is the outcome of attempting canonical-symbol resolution
// before calling a provider.
type resolvedSymbol struct {
	SecurityID     string
	IMSSymbol      string
	DisplaySymbol  string
	ProviderSymbol string // provider symbol to send to the adapter, when known
	HasMapping     bool   // true if (provider_code, provider_symbol) mapping is known
}
