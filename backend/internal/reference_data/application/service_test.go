package application_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
)

func TestCreateSecurity_RequiresIMSSymbolOrAutoBuild(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	_, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		DisplaySymbol: "KBANK.BK",
		Name:          "Kasikornbank PCL",
		AssetType:     domain.AssetTypeEquity,
		Currency:      "THB",
		CountryCode:   "TH",
		ExchangeMIC:   "XBKK",
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, application.ErrInvalidRequest))
}

func TestCreateSecurity_AutoBuildsIMSSymbolForEquity(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	sec, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		DisplaySymbol:      "KBANK.BK",
		Name:               "Kasikornbank PCL",
		AssetType:          domain.AssetTypeEquity,
		Currency:           "THB",
		CountryCode:        "TH",
		ExchangeMIC:        "XBKK",
		ISIN:               "TH0016010R14",
		AutoBuildIMSSymbol: true,
	})
	require.NoError(t, err)
	require.Equal(t, "TH_EQ_XBKK_KBANK.BK", sec.IMSSymbol)
	require.Equal(t, domain.SecurityStatusActive, sec.Status)
}

func TestCreateSecurity_RejectsDuplicateIMSSymbol(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	in := application.CreateSecurityInput{
		IMSSymbol:     "TH_EQ_XBKK_KBANK",
		DisplaySymbol: "KBANK.BK",
		Name:          "Kasikornbank",
		AssetType:     domain.AssetTypeEquity,
		Currency:      "THB",
	}
	_, err := svc.CreateSecurity(context.Background(), in)
	require.NoError(t, err)
	_, err = svc.CreateSecurity(context.Background(), in)
	require.Error(t, err)
	require.True(t, errors.Is(err, application.ErrSecurityConflict))
}

func TestCreateSecurity_RejectsInvalidCurrency(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	// "THBX" is 4 chars (after uppercasing) → invalid by length.
	_, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol:     "TH_EQ_XBKK_KBANK",
		DisplaySymbol: "KBANK.BK",
		Name:          "Kasikornbank",
		AssetType:     domain.AssetTypeEquity,
		Currency:      "THBX",
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, application.ErrInvalidRequest))
}

func TestCreateSecurity_AcceptsBondAndFX(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	bond, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol:     "TH_BOND_TH0623053C09",
		DisplaySymbol: "LB236A",
		Name:          "Thai Govt Bond 2036",
		AssetType:     domain.AssetTypeBond,
		Currency:      "THB",
		CountryCode:   "TH",
		ISIN:          "TH0623053C09",
	})
	require.NoError(t, err)
	require.Equal(t, domain.AssetTypeBond, bond.AssetType)

	fx, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol:     "FX_USDTHB",
		DisplaySymbol: "USDTHB",
		Name:          "USD / THB",
		AssetType:     domain.AssetTypeFX,
		Currency:      "USD",
	})
	require.NoError(t, err)
	require.Equal(t, domain.AssetTypeFX, fx.AssetType)
}

func TestUpdateSecurity_PatchesFields(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	created, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol:     "TH_EQ_XBKK_KBANK",
		DisplaySymbol: "KBANK.BK",
		Name:          "Kasikornbank",
		AssetType:     domain.AssetTypeEquity,
		Currency:      "THB",
	})
	require.NoError(t, err)

	newName := "Kasikornbank PCL"
	suspended := domain.SecurityStatusSuspended
	updated, err := svc.UpdateSecurity(context.Background(), application.UpdateSecurityInput{
		ID:     created.ID,
		Name:   &newName,
		Status: &suspended,
	})
	require.NoError(t, err)
	require.Equal(t, newName, updated.Name)
	require.Equal(t, suspended, updated.Status)
}

func TestAddProviderMapping_MultiProvider(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	sec, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol:     "TH_EQ_XBKK_KBANK",
		DisplaySymbol: "KBANK.BK",
		Name:          "Kasikornbank",
		AssetType:     domain.AssetTypeEquity,
		Currency:      "THB",
	})
	require.NoError(t, err)

	_, err = svc.AddProviderMapping(context.Background(), application.AddProviderMappingInput{
		SecurityID:       sec.ID,
		ProviderCode:     "yahoo",
		ProviderSymbol:   "KBANK.BK",
		ProviderCurrency: "THB",
		IsPrimary:        true,
	})
	require.NoError(t, err)

	_, err = svc.AddProviderMapping(context.Background(), application.AddProviderMappingInput{
		SecurityID:       sec.ID,
		ProviderCode:     "alpha_vantage",
		ProviderSymbol:   "KBANK.BKK",
		ProviderCurrency: "THB",
	})
	require.NoError(t, err)

	mappings, err := svc.ListProviderMappings(context.Background(), sec.ID)
	require.NoError(t, err)
	require.Len(t, mappings, 2)
}

func TestResolveSecurityByProviderSymbol(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	sec, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol:     "TH_EQ_XBKK_KBANK",
		DisplaySymbol: "KBANK.BK",
		Name:          "Kasikornbank",
		AssetType:     domain.AssetTypeEquity,
		Currency:      "THB",
	})
	require.NoError(t, err)

	_, err = svc.AddProviderMapping(context.Background(), application.AddProviderMappingInput{
		SecurityID:     sec.ID,
		ProviderCode:   "yahoo",
		ProviderSymbol: "KBANK.BK",
	})
	require.NoError(t, err)

	got, err := svc.ResolveSecurityByProviderSymbol(context.Background(), "yahoo", "KBANK.BK")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, sec.ID, got.ID)
}

func TestResolveSecurityByProviderSymbol_MissingReturnsNil(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	got, err := svc.ResolveSecurityByProviderSymbol(context.Background(), "yahoo", "WHATEVER")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestCreateUnmappedCandidate_IdempotentWhilePending(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	first, err := svc.CreateUnmappedCandidate(context.Background(), domain.UnmappedSecurityCandidate{
		ProviderCode:   "yahoo",
		ProviderSymbol: "NEW.BK",
	})
	require.NoError(t, err)

	second, err := svc.CreateUnmappedCandidate(context.Background(), domain.UnmappedSecurityCandidate{
		ProviderCode:   "yahoo",
		ProviderSymbol: "NEW.BK",
	})
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)
}

func TestMapUnmappedCandidate_CreatesMappingAndMarksMapped(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	sec, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol:     "TH_EQ_XBKK_NEW",
		DisplaySymbol: "NEW.BK",
		Name:          "New Co",
		AssetType:     domain.AssetTypeEquity,
		Currency:      "THB",
	})
	require.NoError(t, err)

	cand, err := svc.CreateUnmappedCandidate(context.Background(), domain.UnmappedSecurityCandidate{
		ProviderCode:   "yahoo",
		ProviderSymbol: "NEW.BK",
	})
	require.NoError(t, err)

	require.NoError(t, svc.MapUnmappedCandidate(context.Background(), cand.ID, sec.ID))

	mappings, err := svc.ListProviderMappings(context.Background(), sec.ID)
	require.NoError(t, err)
	require.Len(t, mappings, 1)
	require.Equal(t, "yahoo", mappings[0].ProviderCode)

	// Candidate is now MAPPED.
	got := repo.candidates[cand.ID]
	require.Equal(t, domain.CandidateStatusMapped, got.CandidateStatus)
}

func TestRejectUnmappedCandidate_RequiresReason(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	cand, err := svc.CreateUnmappedCandidate(context.Background(), domain.UnmappedSecurityCandidate{
		ProviderCode:   "yahoo",
		ProviderSymbol: "WHATEVER",
	})
	require.NoError(t, err)

	require.Error(t, svc.RejectUnmappedCandidate(context.Background(), cand.ID, ""))
	require.NoError(t, svc.RejectUnmappedCandidate(context.Background(), cand.ID, "bad symbol"))
	require.Equal(t, domain.CandidateStatusRejected, repo.candidates[cand.ID].CandidateStatus)
	require.Equal(t, "bad symbol", repo.candidates[cand.ID].RejectedReason)
}

func TestRejectAlreadyRejectedFails(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	cand, err := svc.CreateUnmappedCandidate(context.Background(), domain.UnmappedSecurityCandidate{
		ProviderCode:   "yahoo",
		ProviderSymbol: "WHATEVER",
	})
	require.NoError(t, err)
	require.NoError(t, svc.RejectUnmappedCandidate(context.Background(), cand.ID, "first reason"))
	err = svc.RejectUnmappedCandidate(context.Background(), cand.ID, "second reason")
	require.Error(t, err)
	require.True(t, errors.Is(err, application.ErrCandidateNotPending))
}

func TestCreateUnmappedCandidate_AttachesSuggestedSecurityByISIN(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	sec, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol:     "TH_BOND_TH0623053C09",
		DisplaySymbol: "LB236A",
		Name:          "Thai Govt Bond",
		AssetType:     domain.AssetTypeBond,
		Currency:      "THB",
		ISIN:          "TH0623053C09",
	})
	require.NoError(t, err)

	cand, err := svc.CreateUnmappedCandidate(context.Background(), domain.UnmappedSecurityCandidate{
		ProviderCode:   "yahoo",
		ProviderSymbol: "BOND_WHATEVER",
		ISIN:           "TH0623053C09",
	})
	require.NoError(t, err)
	require.NotNil(t, cand.SuggestedSecurityID)
	require.Equal(t, sec.ID, *cand.SuggestedSecurityID)
}

func TestSearchSecuritiesFilters(t *testing.T) {
	repo := newFakeRepo()
	svc := application.NewService(repo)

	for i, in := range []application.CreateSecurityInput{
		{IMSSymbol: "TH_EQ_XBKK_KBANK", DisplaySymbol: "KBANK.BK", Name: "Kasikornbank", AssetType: domain.AssetTypeEquity},
		{IMSSymbol: "TH_EQ_XBKK_PTT", DisplaySymbol: "PTT.BK", Name: "PTT PCL", AssetType: domain.AssetTypeEquity},
		{IMSSymbol: "FX_USDTHB", DisplaySymbol: "USDTHB", Name: "USD / THB", AssetType: domain.AssetTypeFX},
	} {
		_ = i
		_, err := svc.CreateSecurity(context.Background(), in)
		require.NoError(t, err)
	}

	results, err := svc.SearchSecurities(context.Background(), domain.SecuritySearchFilter{Query: "kbank"})
	require.NoError(t, err)
	require.Len(t, results, 1)

	fx, err := svc.SearchSecurities(context.Background(), domain.SecuritySearchFilter{AssetType: domain.AssetTypeFX})
	require.NoError(t, err)
	require.Len(t, fx, 1)
}

// ---- fake repository ----

type fakeRepo struct {
	mu         sync.Mutex
	securities map[string]*domain.Security
	mappings   map[string]*domain.ProviderMapping
	candidates map[string]*domain.UnmappedSecurityCandidate
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		securities: map[string]*domain.Security{},
		mappings:   map[string]*domain.ProviderMapping{},
		candidates: map[string]*domain.UnmappedSecurityCandidate{},
	}
}

func (r *fakeRepo) CreateSecurity(ctx context.Context, sec domain.Security) (*domain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	sec.ID = uuid.NewString()
	clone := sec
	r.securities[sec.ID] = &clone
	return &clone, nil
}

func (r *fakeRepo) UpdateSecurity(ctx context.Context, sec domain.Security) (*domain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.securities[sec.ID]
	if !ok {
		return nil, nil
	}
	updated := sec
	r.securities[sec.ID] = &updated
	_ = existing
	return &updated, nil
}

func (r *fakeRepo) GetSecurityByID(ctx context.Context, id string) (*domain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	sec, ok := r.securities[id]
	if !ok {
		return nil, nil
	}
	out := *sec
	out.ProviderMappings = r.mappingsForLocked(id)
	return &out, nil
}

func (r *fakeRepo) GetSecurityByIMSSymbol(ctx context.Context, imsSymbol string) (*domain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.securities {
		if s.IMSSymbol == imsSymbol {
			out := *s
			out.ProviderMappings = r.mappingsForLocked(s.ID)
			return &out, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) GetSecurityByDisplaySymbol(ctx context.Context, displaySymbol string) (*domain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.securities {
		if s.DisplaySymbol == displaySymbol {
			out := *s
			out.ProviderMappings = r.mappingsForLocked(s.ID)
			return &out, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) GetSecurityByISIN(ctx context.Context, isin string) (*domain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.securities {
		if s.ISIN == isin {
			out := *s
			return &out, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) SearchSecurities(ctx context.Context, filter domain.SecuritySearchFilter) ([]domain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []domain.Security{}
	q := strings.ToLower(filter.Query)
	for _, s := range r.securities {
		if filter.AssetType != "" && s.AssetType != filter.AssetType {
			continue
		}
		if filter.Status != "" && s.Status != filter.Status {
			continue
		}
		if q != "" {
			match := strings.Contains(strings.ToLower(s.IMSSymbol), q) ||
				strings.Contains(strings.ToLower(s.DisplaySymbol), q) ||
				strings.Contains(strings.ToLower(s.Name), q) ||
				strings.Contains(strings.ToLower(s.ISIN), q)
			if !match {
				continue
			}
		}
		clone := *s
		clone.ProviderMappings = r.mappingsForLocked(s.ID)
		out = append(out, clone)
	}
	return out, nil
}

func (r *fakeRepo) AddProviderMapping(ctx context.Context, m domain.ProviderMapping) (*domain.ProviderMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// upsert by (provider_code, provider_symbol)
	for _, existing := range r.mappings {
		if existing.ProviderCode == m.ProviderCode && existing.ProviderSymbol == m.ProviderSymbol {
			existing.SecurityID = m.SecurityID
			existing.MappingStatus = m.MappingStatus
			if m.ConfidenceScore.IsZero() {
				existing.ConfidenceScore = decimal.NewFromInt(100)
			} else {
				existing.ConfidenceScore = m.ConfidenceScore
			}
			existing.IsPrimary = m.IsPrimary
			out := *existing
			return &out, nil
		}
	}
	m.ID = uuid.NewString()
	if m.ConfidenceScore.IsZero() {
		m.ConfidenceScore = decimal.NewFromInt(100)
	}
	clone := m
	r.mappings[m.ID] = &clone
	return &clone, nil
}

func (r *fakeRepo) GetProviderMapping(ctx context.Context, id string) (*domain.ProviderMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.mappings[id]
	if !ok {
		return nil, nil
	}
	out := *m
	return &out, nil
}

func (r *fakeRepo) ListProviderMappingsBySecurity(ctx context.Context, securityID string) ([]domain.ProviderMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.mappingsForLocked(securityID), nil
}

func (r *fakeRepo) mappingsForLocked(securityID string) []domain.ProviderMapping {
	out := []domain.ProviderMapping{}
	for _, m := range r.mappings {
		if m.SecurityID == securityID {
			out = append(out, *m)
		}
	}
	return out
}

func (r *fakeRepo) GetProviderMappingByProviderSymbol(ctx context.Context, providerCode, providerSymbol string) (*domain.ProviderMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.mappings {
		if m.ProviderCode == providerCode && m.ProviderSymbol == providerSymbol {
			out := *m
			return &out, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) GetActiveProviderMappingForSecurity(ctx context.Context, securityID, providerCode string) (*domain.ProviderMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.mappings {
		if m.SecurityID == securityID && m.ProviderCode == providerCode && m.MappingStatus == domain.MappingStatusActive {
			out := *m
			return &out, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) SetProviderMappingStatus(ctx context.Context, mappingID string, status domain.MappingStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.mappings[mappingID]; ok {
		m.MappingStatus = status
	}
	return nil
}

func (r *fakeRepo) CreateUnmappedCandidate(ctx context.Context, c domain.UnmappedSecurityCandidate) (*domain.UnmappedSecurityCandidate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c.ID = uuid.NewString()
	clone := c
	r.candidates[c.ID] = &clone
	return &clone, nil
}

func (r *fakeRepo) GetUnmappedCandidate(ctx context.Context, id string) (*domain.UnmappedSecurityCandidate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.candidates[id]
	if !ok {
		return nil, nil
	}
	out := *c
	return &out, nil
}

func (r *fakeRepo) ListUnmappedCandidates(ctx context.Context, filter domain.UnmappedCandidateFilter) ([]domain.UnmappedSecurityCandidate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []domain.UnmappedSecurityCandidate{}
	for _, c := range r.candidates {
		if filter.Status != "" && c.CandidateStatus != filter.Status {
			continue
		}
		if filter.ProviderCode != "" && c.ProviderCode != filter.ProviderCode {
			continue
		}
		out = append(out, *c)
	}
	return out, nil
}

func (r *fakeRepo) UpdateUnmappedCandidateStatus(ctx context.Context, id string, status string, rejectedReason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.candidates[id]; ok {
		c.CandidateStatus = status
		if rejectedReason != "" {
			c.RejectedReason = rejectedReason
		}
	}
	return nil
}

func (r *fakeRepo) GetOpenCandidateByProviderSymbol(ctx context.Context, providerCode, providerSymbol string) (*domain.UnmappedSecurityCandidate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.candidates {
		if c.ProviderCode == providerCode && c.ProviderSymbol == providerSymbol && c.CandidateStatus == domain.CandidateStatusReviewRequired {
			out := *c
			return &out, nil
		}
	}
	return nil, nil
}
