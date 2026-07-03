// Package domain contains canonical IMS security identity types owned by the
// reference_data module. Other modules must consume this identity through the
// SecurityResolver interface (also defined here) — direct SQL access to
// securities_master / security_provider_mappings is reserved for this module.
package domain

import (
	"context"

	"github.com/shopspring/decimal"
)

// AssetType is the canonical enumeration of asset types supported by IMS.
type AssetType string

const (
	AssetTypeUnknown    AssetType = "UNKNOWN"
	AssetTypeEquity     AssetType = "EQUITY"
	AssetTypeETF        AssetType = "ETF"
	AssetTypeMutualFund AssetType = "MUTUAL_FUND"
	AssetTypeBond       AssetType = "BOND"
	AssetTypeFX         AssetType = "FX"
	AssetTypeIndex      AssetType = "INDEX"
	AssetTypeCash       AssetType = "CASH"
	AssetTypeDerivative AssetType = "DERIVATIVE"
	AssetTypeOther      AssetType = "OTHER"
)

// AllAssetTypes returns every valid asset type. Kept stable for validation.
func AllAssetTypes() []AssetType {
	return []AssetType{
		AssetTypeUnknown, AssetTypeEquity, AssetTypeETF, AssetTypeMutualFund,
		AssetTypeBond, AssetTypeFX, AssetTypeIndex, AssetTypeCash,
		AssetTypeDerivative, AssetTypeOther,
	}
}

// IsValid reports whether the asset type is one of the supported values.
func (a AssetType) IsValid() bool {
	for _, x := range AllAssetTypes() {
		if x == a {
			return true
		}
	}
	return false
}

// SecurityStatus is the lifecycle status of a canonical security.
type SecurityStatus string

const (
	SecurityStatusActive    SecurityStatus = "ACTIVE"
	SecurityStatusInactive  SecurityStatus = "INACTIVE"
	SecurityStatusSuspended SecurityStatus = "SUSPENDED"
)

// IsValid reports whether the security status is supported.
func (s SecurityStatus) IsValid() bool {
	switch s {
	case SecurityStatusActive, SecurityStatusInactive, SecurityStatusSuspended:
		return true
	}
	return false
}

// MappingStatus is the lifecycle status of a provider-symbol mapping.
type MappingStatus string

const (
	MappingStatusActive         MappingStatus = "ACTIVE"
	MappingStatusInactive       MappingStatus = "INACTIVE"
	MappingStatusUnmapped       MappingStatus = "UNMAPPED"
	MappingStatusConflicted     MappingStatus = "CONFLICTED"
	MappingStatusReviewRequired MappingStatus = "REVIEW_REQUIRED"
)

// IsValid reports whether the mapping status is supported.
func (m MappingStatus) IsValid() bool {
	switch m {
	case MappingStatusActive, MappingStatusInactive, MappingStatusUnmapped,
		MappingStatusConflicted, MappingStatusReviewRequired:
		return true
	}
	return false
}

// Candidate statuses for unmapped security candidates.
const (
	CandidateStatusReviewRequired = "REVIEW_REQUIRED"
	CandidateStatusMapped         = "MAPPED"
	CandidateStatusRejected       = "REJECTED"
	CandidateStatusDuplicate      = "DUPLICATE"
	CandidateStatusConflicted     = "CONFLICTED"
)

// Security is the canonical IMS security identity.
type Security struct {
	ID                string
	IMSSymbol         string
	DisplaySymbol     string
	PrimaryIdentifier string
	Name              string
	AssetType         AssetType
	Currency          string
	CountryCode       string
	ExchangeMIC       string
	ISIN              string
	CUSIP             string
	FIGI              string
	Status            SecurityStatus
	ProviderMappings  []ProviderMapping
}

// ProviderMapping links a provider's symbol/code pair to a canonical Security.
type ProviderMapping struct {
	ID                string
	SecurityID        string
	ProviderCode      string
	ProviderSymbol    string
	ProviderExchange  string
	ProviderAssetType string
	ProviderCurrency  string
	Priority          int
	ConfidenceScore   decimal.Decimal
	MappingStatus     MappingStatus
	IsPrimary         bool
}

// UnmappedSecurityCandidate represents a provider symbol that did not resolve
// to a canonical security and needs human review.
type UnmappedSecurityCandidate struct {
	ID                  string
	BatchID             *string
	ProviderCode        string
	ProviderSymbol      string
	ProviderName        string
	ProviderAssetType   string
	ProviderExchange    string
	ProviderCurrency    string
	ISIN                string
	RawPayload          map[string]any
	CandidateStatus     string
	SuggestedSecurityID *string
	ConfidenceScore     *decimal.Decimal
	RejectedReason      string
}

// SecuritySearchFilter constrains SearchSecurities results.
type SecuritySearchFilter struct {
	Query     string
	AssetType AssetType
	Provider  string
	Status    SecurityStatus
	Limit     int
}

// UnmappedCandidateFilter constrains ListUnmappedCandidates results.
type UnmappedCandidateFilter struct {
	Status       string
	ProviderCode string
	BatchID      string
	Limit        int
}

// SecurityResolver is the read/resolve contract that other modules
// (market_data, investment) consume to translate provider data into
// canonical IMS securities. It deliberately excludes the full CRUD
// surface so cross-module callers cannot mutate identity through it.
type SecurityResolver interface {
	SearchSecurities(ctx context.Context, filter SecuritySearchFilter) ([]Security, error)
	GetSecurityByID(ctx context.Context, id string) (*Security, error)
	GetSecurityByIMSSymbol(ctx context.Context, imsSymbol string) (*Security, error)
	GetSecurityByDisplaySymbol(ctx context.Context, displaySymbol string) (*Security, error)

	ResolveSecurityByProviderSymbol(ctx context.Context, providerCode string, providerSymbol string) (*Security, error)
	ResolveProviderSymbol(ctx context.Context, securityID string, providerCode string) (*ProviderMapping, error)

	CreateUnmappedCandidate(ctx context.Context, candidate UnmappedSecurityCandidate) (*UnmappedSecurityCandidate, error)
	ListUnmappedCandidates(ctx context.Context, filter UnmappedCandidateFilter) ([]UnmappedSecurityCandidate, error)
	MapUnmappedCandidate(ctx context.Context, candidateID string, securityID string) error
	RejectUnmappedCandidate(ctx context.Context, candidateID string, reason string) error
}
