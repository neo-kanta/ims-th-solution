// Package request contains transport-layer request DTOs for the reference_data
// module. These are intentionally distinct from domain types so HTTP wire
// format and internal models can evolve independently.
package request

// CreateSecurityRequest is the body for POST /reference-data/securities.
type CreateSecurityRequest struct {
	IMSSymbol          string `json:"ims_symbol"`
	DisplaySymbol      string `json:"display_symbol"`
	PrimaryIdentifier  string `json:"primary_identifier"`
	Name               string `json:"name"`
	AssetType          string `json:"asset_type"`
	Currency           string `json:"currency"`
	CountryCode        string `json:"country_code"`
	ExchangeMIC        string `json:"exchange_mic"`
	ISIN               string `json:"isin"`
	CUSIP              string `json:"cusip"`
	FIGI               string `json:"figi"`
	Status             string `json:"status"`
	AutoBuildIMSSymbol bool   `json:"auto_build_ims_symbol"`
}

// UpdateSecurityRequest is the body for PATCH /reference-data/securities/{id}.
// All fields are pointers so PATCH semantics (omit = leave) are explicit.
type UpdateSecurityRequest struct {
	DisplaySymbol     *string `json:"display_symbol"`
	PrimaryIdentifier *string `json:"primary_identifier"`
	Name              *string `json:"name"`
	AssetType         *string `json:"asset_type"`
	Currency          *string `json:"currency"`
	CountryCode       *string `json:"country_code"`
	ExchangeMIC       *string `json:"exchange_mic"`
	ISIN              *string `json:"isin"`
	CUSIP             *string `json:"cusip"`
	FIGI              *string `json:"figi"`
	Status            *string `json:"status"`
}

// AddProviderMappingRequest is the body for POST /reference-data/securities/{id}/mappings.
type AddProviderMappingRequest struct {
	ProviderCode      string  `json:"provider_code"`
	ProviderSymbol    string  `json:"provider_symbol"`
	ProviderExchange  string  `json:"provider_exchange"`
	ProviderAssetType string  `json:"provider_asset_type"`
	ProviderCurrency  string  `json:"provider_currency"`
	Priority          int     `json:"priority"`
	ConfidenceScore   *string `json:"confidence_score"`
	IsPrimary         bool    `json:"is_primary"`
}

// MapCandidateRequest is the body for POST /reference-data/unmapped-candidates/{id}/map.
type MapCandidateRequest struct {
	SecurityID string `json:"security_id"`
}

// RejectCandidateRequest is the body for POST /reference-data/unmapped-candidates/{id}/reject.
type RejectCandidateRequest struct {
	Reason string `json:"reason"`
}
