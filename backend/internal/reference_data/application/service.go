// Package application contains the reference_data service. It is the single
// owner of canonical IMS security identity and the provider-symbol resolution
// logic — other modules consume it through the SecurityResolver interface in
// reference_data/domain.
package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
)

// Sentinel errors emitted by the service. Transport handlers translate these
// into HTTP status codes.
var (
	ErrInvalidRequest      = errors.New("invalid reference data request")
	ErrSecurityNotFound    = errors.New("security not found")
	ErrCandidateNotFound   = errors.New("unmapped candidate not found")
	ErrMappingNotFound     = errors.New("provider mapping not found")
	ErrSecurityConflict    = errors.New("security conflict")
	ErrCandidateNotPending = errors.New("candidate is not in REVIEW_REQUIRED state")
)

// Service is the reference_data application service.
type Service struct {
	repo domain.SecurityRepository
	now  func() time.Time
}

// NewService constructs the Service.
func NewService(repo domain.SecurityRepository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

// CreateSecurityInput is the application-level input for CreateSecurity.
// AutoBuildIMSSymbol asks the service to derive ims_symbol when the caller
// did not provide one.
type CreateSecurityInput struct {
	IMSSymbol          string
	DisplaySymbol      string
	PrimaryIdentifier  string
	Name               string
	AssetType          domain.AssetType
	Currency           string
	CountryCode        string
	ExchangeMIC        string
	ISIN               string
	CUSIP              string
	FIGI               string
	Status             domain.SecurityStatus
	AutoBuildIMSSymbol bool
}

// UpdateSecurityInput captures the patchable subset of Security fields.
type UpdateSecurityInput struct {
	ID                string
	DisplaySymbol     *string
	PrimaryIdentifier *string
	Name              *string
	AssetType         *domain.AssetType
	Currency          *string
	CountryCode       *string
	ExchangeMIC       *string
	ISIN              *string
	CUSIP             *string
	FIGI              *string
	Status            *domain.SecurityStatus
}

// AddProviderMappingInput captures provider-mapping creation fields.
type AddProviderMappingInput struct {
	SecurityID        string
	ProviderCode      string
	ProviderSymbol    string
	ProviderExchange  string
	ProviderAssetType string
	ProviderCurrency  string
	Priority          int
	ConfidenceScore   *decimal.Decimal
	IsPrimary         bool
}

// SearchSecurities forwards to the repository with applied validation.
func (s *Service) SearchSecurities(ctx context.Context, filter domain.SecuritySearchFilter) ([]domain.Security, error) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	filter.Query = strings.TrimSpace(filter.Query)
	filter.Provider = strings.TrimSpace(filter.Provider)
	return s.repo.SearchSecurities(ctx, filter)
}

// GetSecurityByID returns a security by canonical UUID.
func (s *Service) GetSecurityByID(ctx context.Context, id string) (*domain.Security, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%w: security id is required", ErrInvalidRequest)
	}
	sec, err := s.repo.GetSecurityByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sec == nil {
		return nil, ErrSecurityNotFound
	}
	return sec, nil
}

// GetSecurityByIMSSymbol returns a security by ims_symbol.
func (s *Service) GetSecurityByIMSSymbol(ctx context.Context, imsSymbol string) (*domain.Security, error) {
	imsSymbol = strings.ToUpper(strings.TrimSpace(imsSymbol))
	if imsSymbol == "" {
		return nil, fmt.Errorf("%w: ims_symbol is required", ErrInvalidRequest)
	}
	return s.repo.GetSecurityByIMSSymbol(ctx, imsSymbol)
}

// GetSecurityByDisplaySymbol returns the first ACTIVE security with matching display_symbol.
func (s *Service) GetSecurityByDisplaySymbol(ctx context.Context, displaySymbol string) (*domain.Security, error) {
	displaySymbol = strings.TrimSpace(displaySymbol)
	if displaySymbol == "" {
		return nil, fmt.Errorf("%w: display_symbol is required", ErrInvalidRequest)
	}
	return s.repo.GetSecurityByDisplaySymbol(ctx, displaySymbol)
}

// CreateSecurity registers a new canonical security.
func (s *Service) CreateSecurity(ctx context.Context, in CreateSecurityInput) (*domain.Security, error) {
	if !in.AssetType.IsValid() {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, domain.ErrInvalidAssetType)
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidRequest)
	}
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	in.CountryCode = strings.ToUpper(strings.TrimSpace(in.CountryCode))
	in.ExchangeMIC = strings.ToUpper(strings.TrimSpace(in.ExchangeMIC))
	in.ISIN = strings.ToUpper(strings.TrimSpace(in.ISIN))
	if err := domain.ValidateCurrency(in.Currency); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	if err := domain.ValidateCountryCode(in.CountryCode); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	if err := domain.ValidateISIN(in.ISIN); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}

	in.DisplaySymbol = strings.TrimSpace(in.DisplaySymbol)
	if in.DisplaySymbol == "" {
		in.DisplaySymbol = strings.TrimSpace(in.PrimaryIdentifier)
	}
	if err := domain.ValidateDisplaySymbol(in.DisplaySymbol); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}

	in.IMSSymbol = strings.ToUpper(strings.TrimSpace(in.IMSSymbol))
	if in.IMSSymbol == "" {
		if !in.AutoBuildIMSSymbol {
			return nil, fmt.Errorf("%w: ims_symbol is required", ErrInvalidRequest)
		}
		built, err := domain.BuildIMSSymbol(domain.BuildIMSSymbolInput{
			AssetType:   in.AssetType,
			CountryCode: in.CountryCode,
			ExchangeMIC: in.ExchangeMIC,
			Ticker:      in.DisplaySymbol,
			ISIN:        in.ISIN,
			Fallback:    in.DisplaySymbol,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
		}
		in.IMSSymbol = built
	}
	if err := domain.ValidateIMSSymbol(in.IMSSymbol); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}

	status := in.Status
	if status == "" {
		status = domain.SecurityStatusActive
	}
	if !status.IsValid() {
		return nil, fmt.Errorf("%w: invalid status %q", ErrInvalidRequest, status)
	}

	// Duplicate check.
	if existing, err := s.repo.GetSecurityByIMSSymbol(ctx, in.IMSSymbol); err == nil && existing != nil {
		return nil, fmt.Errorf("%w: ims_symbol %q already exists", ErrSecurityConflict, in.IMSSymbol)
	}

	created, err := s.repo.CreateSecurity(ctx, domain.Security{
		IMSSymbol:         in.IMSSymbol,
		DisplaySymbol:     in.DisplaySymbol,
		PrimaryIdentifier: strings.TrimSpace(in.PrimaryIdentifier),
		Name:              strings.TrimSpace(in.Name),
		AssetType:         in.AssetType,
		Currency:          in.Currency,
		CountryCode:       in.CountryCode,
		ExchangeMIC:       in.ExchangeMIC,
		ISIN:              in.ISIN,
		CUSIP:             strings.ToUpper(strings.TrimSpace(in.CUSIP)),
		FIGI:              strings.ToUpper(strings.TrimSpace(in.FIGI)),
		Status:            status,
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// UpdateSecurity applies a patch update to a canonical security.
func (s *Service) UpdateSecurity(ctx context.Context, in UpdateSecurityInput) (*domain.Security, error) {
	existing, err := s.GetSecurityByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if in.DisplaySymbol != nil {
		ds := strings.TrimSpace(*in.DisplaySymbol)
		if err := domain.ValidateDisplaySymbol(ds); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
		}
		existing.DisplaySymbol = ds
	}
	if in.PrimaryIdentifier != nil {
		existing.PrimaryIdentifier = strings.TrimSpace(*in.PrimaryIdentifier)
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidRequest)
		}
		existing.Name = name
	}
	if in.AssetType != nil {
		if !in.AssetType.IsValid() {
			return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, domain.ErrInvalidAssetType)
		}
		existing.AssetType = *in.AssetType
	}
	if in.Currency != nil {
		curr := strings.ToUpper(strings.TrimSpace(*in.Currency))
		if err := domain.ValidateCurrency(curr); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
		}
		existing.Currency = curr
	}
	if in.CountryCode != nil {
		code := strings.ToUpper(strings.TrimSpace(*in.CountryCode))
		if err := domain.ValidateCountryCode(code); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
		}
		existing.CountryCode = code
	}
	if in.ExchangeMIC != nil {
		existing.ExchangeMIC = strings.ToUpper(strings.TrimSpace(*in.ExchangeMIC))
	}
	if in.ISIN != nil {
		isin := strings.ToUpper(strings.TrimSpace(*in.ISIN))
		if err := domain.ValidateISIN(isin); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
		}
		existing.ISIN = isin
	}
	if in.CUSIP != nil {
		existing.CUSIP = strings.ToUpper(strings.TrimSpace(*in.CUSIP))
	}
	if in.FIGI != nil {
		existing.FIGI = strings.ToUpper(strings.TrimSpace(*in.FIGI))
	}
	if in.Status != nil {
		if !in.Status.IsValid() {
			return nil, fmt.Errorf("%w: invalid status %q", ErrInvalidRequest, *in.Status)
		}
		existing.Status = *in.Status
	}

	return s.repo.UpdateSecurity(ctx, *existing)
}

// AddProviderMapping attaches a provider symbol to a canonical security.
func (s *Service) AddProviderMapping(ctx context.Context, in AddProviderMappingInput) (*domain.ProviderMapping, error) {
	if strings.TrimSpace(in.SecurityID) == "" {
		return nil, fmt.Errorf("%w: security_id is required", ErrInvalidRequest)
	}
	in.ProviderCode = strings.ToLower(strings.TrimSpace(in.ProviderCode))
	in.ProviderSymbol = strings.TrimSpace(in.ProviderSymbol)
	if in.ProviderCode == "" || in.ProviderSymbol == "" {
		return nil, fmt.Errorf("%w: provider_code and provider_symbol are required", ErrInvalidRequest)
	}
	in.ProviderCurrency = strings.ToUpper(strings.TrimSpace(in.ProviderCurrency))
	if err := domain.ValidateCurrency(in.ProviderCurrency); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	if _, err := s.GetSecurityByID(ctx, in.SecurityID); err != nil {
		return nil, err
	}

	priority := in.Priority
	if priority <= 0 {
		priority = 100
	}
	confidence := decimal.NewFromInt(100)
	if in.ConfidenceScore != nil {
		confidence = *in.ConfidenceScore
	}

	return s.repo.AddProviderMapping(ctx, domain.ProviderMapping{
		SecurityID:        in.SecurityID,
		ProviderCode:      in.ProviderCode,
		ProviderSymbol:    in.ProviderSymbol,
		ProviderExchange:  strings.TrimSpace(in.ProviderExchange),
		ProviderAssetType: strings.TrimSpace(in.ProviderAssetType),
		ProviderCurrency:  in.ProviderCurrency,
		Priority:          priority,
		ConfidenceScore:   confidence,
		MappingStatus:     domain.MappingStatusActive,
		IsPrimary:         in.IsPrimary,
	})
}

// RemoveProviderMapping soft-deletes a mapping by marking it INACTIVE.
func (s *Service) RemoveProviderMapping(ctx context.Context, mappingID string) error {
	mappingID = strings.TrimSpace(mappingID)
	if mappingID == "" {
		return fmt.Errorf("%w: mapping_id is required", ErrInvalidRequest)
	}
	existing, err := s.repo.GetProviderMapping(ctx, mappingID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrMappingNotFound
	}
	return s.repo.SetProviderMappingStatus(ctx, mappingID, domain.MappingStatusInactive)
}

// ListProviderMappings returns all mappings (any status) for a security.
func (s *Service) ListProviderMappings(ctx context.Context, securityID string) ([]domain.ProviderMapping, error) {
	if _, err := s.GetSecurityByID(ctx, securityID); err != nil {
		return nil, err
	}
	return s.repo.ListProviderMappingsBySecurity(ctx, securityID)
}

// ResolveSecurityByProviderSymbol implements the resolver contract.
func (s *Service) ResolveSecurityByProviderSymbol(ctx context.Context, providerCode, providerSymbol string) (*domain.Security, error) {
	providerCode = strings.ToLower(strings.TrimSpace(providerCode))
	providerSymbol = strings.TrimSpace(providerSymbol)
	if providerCode == "" || providerSymbol == "" {
		return nil, fmt.Errorf("%w: provider_code and provider_symbol are required", ErrInvalidRequest)
	}
	mapping, err := s.repo.GetProviderMappingByProviderSymbol(ctx, providerCode, providerSymbol)
	if err != nil {
		return nil, err
	}
	if mapping == nil || mapping.MappingStatus != domain.MappingStatusActive {
		return nil, nil
	}
	return s.repo.GetSecurityByID(ctx, mapping.SecurityID)
}

// ResolveProviderSymbol returns the active mapping for a security/provider pair.
func (s *Service) ResolveProviderSymbol(ctx context.Context, securityID, providerCode string) (*domain.ProviderMapping, error) {
	if strings.TrimSpace(securityID) == "" {
		return nil, fmt.Errorf("%w: security_id is required", ErrInvalidRequest)
	}
	providerCode = strings.ToLower(strings.TrimSpace(providerCode))
	if providerCode == "" {
		return nil, fmt.Errorf("%w: provider_code is required", ErrInvalidRequest)
	}
	return s.repo.GetActiveProviderMappingForSecurity(ctx, securityID, providerCode)
}

// CreateUnmappedCandidate enqueues an unresolved provider symbol for review.
// If an open candidate already exists, the existing one is returned.
func (s *Service) CreateUnmappedCandidate(ctx context.Context, candidate domain.UnmappedSecurityCandidate) (*domain.UnmappedSecurityCandidate, error) {
	candidate.ProviderCode = strings.ToLower(strings.TrimSpace(candidate.ProviderCode))
	candidate.ProviderSymbol = strings.TrimSpace(candidate.ProviderSymbol)
	if candidate.ProviderCode == "" || candidate.ProviderSymbol == "" {
		return nil, fmt.Errorf("%w: provider_code and provider_symbol are required", ErrInvalidRequest)
	}
	candidate.ProviderCurrency = strings.ToUpper(strings.TrimSpace(candidate.ProviderCurrency))
	if err := domain.ValidateCurrency(candidate.ProviderCurrency); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	candidate.ISIN = strings.ToUpper(strings.TrimSpace(candidate.ISIN))
	if err := domain.ValidateISIN(candidate.ISIN); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}

	// If an open candidate already exists, return it (idempotent).
	if existing, err := s.repo.GetOpenCandidateByProviderSymbol(ctx, candidate.ProviderCode, candidate.ProviderSymbol); err == nil && existing != nil {
		return existing, nil
	}

	// If ISIN matches a canonical security, suggest it for review.
	if candidate.SuggestedSecurityID == nil && candidate.ISIN != "" {
		if found, err := s.repo.GetSecurityByISIN(ctx, candidate.ISIN); err == nil && found != nil {
			id := found.ID
			candidate.SuggestedSecurityID = &id
		}
	}

	candidate.CandidateStatus = domain.CandidateStatusReviewRequired
	if candidate.RawPayload == nil {
		candidate.RawPayload = map[string]any{}
	}
	return s.repo.CreateUnmappedCandidate(ctx, candidate)
}

// ListUnmappedCandidates returns candidates filtered by status/provider/batch.
func (s *Service) ListUnmappedCandidates(ctx context.Context, filter domain.UnmappedCandidateFilter) ([]domain.UnmappedSecurityCandidate, error) {
	if filter.Limit <= 0 || filter.Limit > 500 {
		filter.Limit = 100
	}
	if filter.Status == "" {
		filter.Status = domain.CandidateStatusReviewRequired
	}
	return s.repo.ListUnmappedCandidates(ctx, filter)
}

// MapUnmappedCandidate marks a candidate MAPPED and creates an ACTIVE provider
// mapping pointing to the chosen security.
func (s *Service) MapUnmappedCandidate(ctx context.Context, candidateID, securityID string) error {
	candidateID = strings.TrimSpace(candidateID)
	securityID = strings.TrimSpace(securityID)
	if candidateID == "" || securityID == "" {
		return fmt.Errorf("%w: candidate_id and security_id are required", ErrInvalidRequest)
	}
	candidate, err := s.repo.GetUnmappedCandidate(ctx, candidateID)
	if err != nil {
		return err
	}
	if candidate == nil {
		return ErrCandidateNotFound
	}
	if candidate.CandidateStatus != domain.CandidateStatusReviewRequired {
		return ErrCandidateNotPending
	}
	if _, err := s.GetSecurityByID(ctx, securityID); err != nil {
		return err
	}
	// Add (or re-activate) the provider mapping.
	if _, err := s.repo.AddProviderMapping(ctx, domain.ProviderMapping{
		SecurityID:        securityID,
		ProviderCode:      candidate.ProviderCode,
		ProviderSymbol:    candidate.ProviderSymbol,
		ProviderExchange:  candidate.ProviderExchange,
		ProviderAssetType: candidate.ProviderAssetType,
		ProviderCurrency:  candidate.ProviderCurrency,
		Priority:          100,
		ConfidenceScore:   decimal.NewFromInt(100),
		MappingStatus:     domain.MappingStatusActive,
		IsPrimary:         false,
	}); err != nil {
		return err
	}
	return s.repo.UpdateUnmappedCandidateStatus(ctx, candidateID, domain.CandidateStatusMapped, "")
}

// RejectUnmappedCandidate marks the candidate REJECTED with the given reason.
func (s *Service) RejectUnmappedCandidate(ctx context.Context, candidateID, reason string) error {
	candidateID = strings.TrimSpace(candidateID)
	if candidateID == "" {
		return fmt.Errorf("%w: candidate_id is required", ErrInvalidRequest)
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidRequest)
	}
	candidate, err := s.repo.GetUnmappedCandidate(ctx, candidateID)
	if err != nil {
		return err
	}
	if candidate == nil {
		return ErrCandidateNotFound
	}
	if candidate.CandidateStatus != domain.CandidateStatusReviewRequired {
		return ErrCandidateNotPending
	}
	return s.repo.UpdateUnmappedCandidateStatus(ctx, candidateID, domain.CandidateStatusRejected, reason)
}

// Ensure the application service satisfies the resolver contract.
var _ domain.SecurityResolver = (*Service)(nil)
