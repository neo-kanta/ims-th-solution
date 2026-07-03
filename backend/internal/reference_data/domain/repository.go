package domain

import "context"

// SecurityRepository persists and retrieves canonical IMS securities, their
// provider mappings, and the unmapped candidate review queue.
type SecurityRepository interface {
	// Securities

	CreateSecurity(ctx context.Context, security Security) (*Security, error)
	UpdateSecurity(ctx context.Context, security Security) (*Security, error)
	GetSecurityByID(ctx context.Context, id string) (*Security, error)
	GetSecurityByIMSSymbol(ctx context.Context, imsSymbol string) (*Security, error)
	GetSecurityByDisplaySymbol(ctx context.Context, displaySymbol string) (*Security, error)
	GetSecurityByISIN(ctx context.Context, isin string) (*Security, error)
	SearchSecurities(ctx context.Context, filter SecuritySearchFilter) ([]Security, error)

	// Provider mappings

	AddProviderMapping(ctx context.Context, mapping ProviderMapping) (*ProviderMapping, error)
	GetProviderMapping(ctx context.Context, id string) (*ProviderMapping, error)
	ListProviderMappingsBySecurity(ctx context.Context, securityID string) ([]ProviderMapping, error)
	GetProviderMappingByProviderSymbol(ctx context.Context, providerCode, providerSymbol string) (*ProviderMapping, error)
	GetActiveProviderMappingForSecurity(ctx context.Context, securityID, providerCode string) (*ProviderMapping, error)
	SetProviderMappingStatus(ctx context.Context, mappingID string, status MappingStatus) error

	// Unmapped candidates

	CreateUnmappedCandidate(ctx context.Context, candidate UnmappedSecurityCandidate) (*UnmappedSecurityCandidate, error)
	GetUnmappedCandidate(ctx context.Context, id string) (*UnmappedSecurityCandidate, error)
	ListUnmappedCandidates(ctx context.Context, filter UnmappedCandidateFilter) ([]UnmappedSecurityCandidate, error)
	UpdateUnmappedCandidateStatus(ctx context.Context, id string, status string, rejectedReason string) error
	GetOpenCandidateByProviderSymbol(ctx context.Context, providerCode, providerSymbol string) (*UnmappedSecurityCandidate, error)
}
