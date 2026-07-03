// Package response contains transport-layer response DTOs for the
// reference_data module. Decimal values are emitted as strings to keep
// precision through JSON.
package response

import (
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
)

// SecurityDTO is the HTTP shape of a canonical security.
type SecurityDTO struct {
	SecurityID        string               `json:"security_id"`
	IMSSymbol         string               `json:"ims_symbol"`
	DisplaySymbol     string               `json:"display_symbol"`
	PrimaryIdentifier string               `json:"primary_identifier,omitempty"`
	Name              string               `json:"name"`
	AssetType         string               `json:"asset_type"`
	Currency          string               `json:"currency,omitempty"`
	CountryCode       string               `json:"country_code,omitempty"`
	ExchangeMIC       string               `json:"exchange_mic,omitempty"`
	ISIN              string               `json:"isin,omitempty"`
	CUSIP             string               `json:"cusip,omitempty"`
	FIGI              string               `json:"figi,omitempty"`
	Status            string               `json:"status"`
	ProviderMappings  []ProviderMappingDTO `json:"provider_mappings"`
}

// ProviderMappingDTO is the HTTP shape of a provider mapping.
type ProviderMappingDTO struct {
	MappingID         string `json:"mapping_id"`
	SecurityID        string `json:"security_id"`
	ProviderCode      string `json:"provider_code"`
	ProviderSymbol    string `json:"provider_symbol"`
	ProviderExchange  string `json:"provider_exchange,omitempty"`
	ProviderAssetType string `json:"provider_asset_type,omitempty"`
	ProviderCurrency  string `json:"provider_currency,omitempty"`
	Priority          int    `json:"priority"`
	ConfidenceScore   string `json:"confidence_score"`
	MappingStatus     string `json:"mapping_status"`
	IsPrimary         bool   `json:"is_primary"`
}

// UnmappedCandidateDTO is the HTTP shape of an unmapped candidate.
type UnmappedCandidateDTO struct {
	CandidateID         string     `json:"candidate_id"`
	BatchID             string     `json:"batch_id,omitempty"`
	ProviderCode        string     `json:"provider_code"`
	ProviderSymbol      string     `json:"provider_symbol"`
	ProviderName        string     `json:"provider_name,omitempty"`
	ProviderAssetType   string     `json:"provider_asset_type,omitempty"`
	ProviderExchange    string     `json:"provider_exchange,omitempty"`
	ProviderCurrency    string     `json:"provider_currency,omitempty"`
	ISIN                string     `json:"isin,omitempty"`
	CandidateStatus     string     `json:"candidate_status"`
	SuggestedSecurityID string     `json:"suggested_security_id,omitempty"`
	ConfidenceScore     string     `json:"confidence_score,omitempty"`
	RejectedReason      string     `json:"rejected_reason,omitempty"`
	CreatedAt           time.Time  `json:"created_at,omitempty"`
	ResolvedAt          *time.Time `json:"resolved_at,omitempty"`
}

// SearchResponse wraps a list of securities for GET /reference-data/securities/search.
type SearchResponse struct {
	Items []SecurityDTO `json:"items"`
}

// MappingsResponse wraps a list of mappings for the mappings collection endpoint.
type MappingsResponse struct {
	Items []ProviderMappingDTO `json:"items"`
}

// CandidatesResponse wraps a list of candidates.
type CandidatesResponse struct {
	Items []UnmappedCandidateDTO `json:"items"`
}

// ToSecurityDTO converts a domain.Security to its HTTP shape.
func ToSecurityDTO(sec domain.Security) SecurityDTO {
	mappings := make([]ProviderMappingDTO, 0, len(sec.ProviderMappings))
	for _, m := range sec.ProviderMappings {
		mappings = append(mappings, ToProviderMappingDTO(m))
	}
	return SecurityDTO{
		SecurityID:        sec.ID,
		IMSSymbol:         sec.IMSSymbol,
		DisplaySymbol:     sec.DisplaySymbol,
		PrimaryIdentifier: sec.PrimaryIdentifier,
		Name:              sec.Name,
		AssetType:         string(sec.AssetType),
		Currency:          sec.Currency,
		CountryCode:       sec.CountryCode,
		ExchangeMIC:       sec.ExchangeMIC,
		ISIN:              sec.ISIN,
		CUSIP:             sec.CUSIP,
		FIGI:              sec.FIGI,
		Status:            string(sec.Status),
		ProviderMappings:  mappings,
	}
}

// ToProviderMappingDTO converts a domain.ProviderMapping to its HTTP shape.
func ToProviderMappingDTO(m domain.ProviderMapping) ProviderMappingDTO {
	return ProviderMappingDTO{
		MappingID:         m.ID,
		SecurityID:        m.SecurityID,
		ProviderCode:      m.ProviderCode,
		ProviderSymbol:    m.ProviderSymbol,
		ProviderExchange:  m.ProviderExchange,
		ProviderAssetType: m.ProviderAssetType,
		ProviderCurrency:  m.ProviderCurrency,
		Priority:          m.Priority,
		ConfidenceScore:   m.ConfidenceScore.String(),
		MappingStatus:     string(m.MappingStatus),
		IsPrimary:         m.IsPrimary,
	}
}

// ToCandidateDTO converts a domain.UnmappedSecurityCandidate to its HTTP shape.
// raw_payload is intentionally NOT included; admin/debug endpoints can expose it later.
func ToCandidateDTO(c domain.UnmappedSecurityCandidate) UnmappedCandidateDTO {
	dto := UnmappedCandidateDTO{
		CandidateID:       c.ID,
		ProviderCode:      c.ProviderCode,
		ProviderSymbol:    c.ProviderSymbol,
		ProviderName:      c.ProviderName,
		ProviderAssetType: c.ProviderAssetType,
		ProviderExchange:  c.ProviderExchange,
		ProviderCurrency:  c.ProviderCurrency,
		ISIN:              c.ISIN,
		CandidateStatus:   c.CandidateStatus,
		RejectedReason:    c.RejectedReason,
	}
	if c.BatchID != nil {
		dto.BatchID = *c.BatchID
	}
	if c.SuggestedSecurityID != nil {
		dto.SuggestedSecurityID = *c.SuggestedSecurityID
	}
	if c.ConfidenceScore != nil {
		dto.ConfidenceScore = c.ConfidenceScore.String()
	}
	return dto
}
