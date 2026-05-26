package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/transport"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/transport/handler"
)

func newTestRouter() (*chi.Mux, *application.Service) {
	repo := newFakeRepo()
	svc := application.NewService(repo)
	h := handler.NewSecuritiesHandler(svc)
	r := chi.NewRouter()
	transport.RegisterRoutes(r, h)
	return r, svc
}

func TestHTTP_CreateAndGetSecurity(t *testing.T) {
	r, _ := newTestRouter()

	body := strings.NewReader(`{
		"ims_symbol": "TH_EQ_XBKK_KBANK",
		"display_symbol": "KBANK.BK",
		"name": "Kasikornbank PCL",
		"asset_type": "EQUITY",
		"currency": "THB",
		"country_code": "TH",
		"exchange_mic": "XBKK",
		"isin": "TH0016010R14"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/reference-data/securities", body)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var createResp struct {
		Data struct {
			SecurityID string `json:"security_id"`
			IMSSymbol  string `json:"ims_symbol"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	require.NotEmpty(t, createResp.Data.SecurityID)

	// GET by id
	getReq := httptest.NewRequest(http.MethodGet, "/reference-data/securities/"+createResp.Data.SecurityID, nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)
	require.Equal(t, http.StatusOK, getRec.Code)
}

func TestHTTP_CreateSecurityRejectsInvalid(t *testing.T) {
	r, _ := newTestRouter()
	body := strings.NewReader(`{"name": "no symbol", "asset_type": "EQUITY"}`)
	req := httptest.NewRequest(http.MethodPost, "/reference-data/securities", body)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHTTP_UpdateSecurity(t *testing.T) {
	r, svc := newTestRouter()
	sec, err := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol: "TH_EQ_XBKK_KBANK", DisplaySymbol: "KBANK.BK",
		Name: "Old", AssetType: domain.AssetTypeEquity, Currency: "THB",
	})
	require.NoError(t, err)

	body := strings.NewReader(`{"name": "Kasikornbank PCL"}`)
	req := httptest.NewRequest(http.MethodPatch, "/reference-data/securities/"+sec.ID, body)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

func TestHTTP_SearchSecurities(t *testing.T) {
	r, svc := newTestRouter()
	_, _ = svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol: "TH_EQ_XBKK_KBANK", DisplaySymbol: "KBANK.BK",
		Name: "Kasikornbank", AssetType: domain.AssetTypeEquity, Currency: "THB",
	})
	_, _ = svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol: "TH_EQ_XBKK_PTT", DisplaySymbol: "PTT.BK",
		Name: "PTT PCL", AssetType: domain.AssetTypeEquity, Currency: "THB",
	})
	req := httptest.NewRequest(http.MethodGet, "/reference-data/securities/search?query=kbank", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var resp struct {
		Data struct {
			Items []map[string]any `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data.Items, 1)
}

func TestHTTP_AddAndListMappings(t *testing.T) {
	r, svc := newTestRouter()
	sec, _ := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol: "TH_EQ_XBKK_KBANK", DisplaySymbol: "KBANK.BK",
		Name: "Kasikornbank", AssetType: domain.AssetTypeEquity, Currency: "THB",
	})
	body := strings.NewReader(`{
		"provider_code": "yahoo",
		"provider_symbol": "KBANK.BK",
		"priority": 10,
		"is_primary": true
	}`)
	req := httptest.NewRequest(http.MethodPost, "/reference-data/securities/"+sec.ID+"/mappings", body)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	// List
	listReq := httptest.NewRequest(http.MethodGet, "/reference-data/securities/"+sec.ID+"/mappings", nil)
	listRec := httptest.NewRecorder()
	r.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)
}

func TestHTTP_DeleteMappingSoftDeletes(t *testing.T) {
	r, svc := newTestRouter()
	sec, _ := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol: "TH_EQ_XBKK_KBANK", DisplaySymbol: "KBANK.BK",
		Name: "Kasikornbank", AssetType: domain.AssetTypeEquity, Currency: "THB",
	})
	mapping, err := svc.AddProviderMapping(context.Background(), application.AddProviderMappingInput{
		SecurityID: sec.ID, ProviderCode: "yahoo", ProviderSymbol: "KBANK.BK",
	})
	require.NoError(t, err)

	delReq := httptest.NewRequest(http.MethodDelete, "/reference-data/securities/"+sec.ID+"/mappings/"+mapping.ID, nil)
	delRec := httptest.NewRecorder()
	r.ServeHTTP(delRec, delReq)
	require.Equal(t, http.StatusNoContent, delRec.Code)

	mappings, err := svc.ListProviderMappings(context.Background(), sec.ID)
	require.NoError(t, err)
	require.Len(t, mappings, 1)
	require.Equal(t, domain.MappingStatusInactive, mappings[0].MappingStatus)
}

func TestHTTP_CandidateWorkflow(t *testing.T) {
	r, svc := newTestRouter()
	sec, _ := svc.CreateSecurity(context.Background(), application.CreateSecurityInput{
		IMSSymbol: "TH_EQ_XBKK_NEW", DisplaySymbol: "NEW.BK",
		Name: "New Co", AssetType: domain.AssetTypeEquity, Currency: "THB",
	})
	cand, err := svc.CreateUnmappedCandidate(context.Background(), domain.UnmappedSecurityCandidate{
		ProviderCode: "yahoo", ProviderSymbol: "NEW.BK",
	})
	require.NoError(t, err)

	// List
	listReq := httptest.NewRequest(http.MethodGet, "/reference-data/unmapped-candidates", nil)
	listRec := httptest.NewRecorder()
	r.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	// Map
	mapBody := bytes.NewBufferString(`{"security_id":"` + sec.ID + `"}`)
	mapReq := httptest.NewRequest(http.MethodPost, "/reference-data/unmapped-candidates/"+cand.ID+"/map", mapBody)
	mapRec := httptest.NewRecorder()
	r.ServeHTTP(mapRec, mapReq)
	require.Equal(t, http.StatusNoContent, mapRec.Code, mapRec.Body.String())

	// Cannot reject mapped candidate.
	rejBody := bytes.NewBufferString(`{"reason":"already mapped"}`)
	rejReq := httptest.NewRequest(http.MethodPost, "/reference-data/unmapped-candidates/"+cand.ID+"/reject", rejBody)
	rejRec := httptest.NewRecorder()
	r.ServeHTTP(rejRec, rejReq)
	require.Equal(t, http.StatusConflict, rejRec.Code)
}

func TestHTTP_CandidateReject(t *testing.T) {
	r, svc := newTestRouter()
	cand, err := svc.CreateUnmappedCandidate(context.Background(), domain.UnmappedSecurityCandidate{
		ProviderCode: "yahoo", ProviderSymbol: "BAD.BK",
	})
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"reason":"bad symbol"}`)
	req := httptest.NewRequest(http.MethodPost, "/reference-data/unmapped-candidates/"+cand.ID+"/reject", body)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
}

// ---- duplicated tiny fake repo so the handler test file can run in
// isolation (the service-test fakeRepo lives in package application_test) ----

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
	clone := sec
	r.securities[sec.ID] = &clone
	return &clone, nil
}

func (r *fakeRepo) GetSecurityByID(ctx context.Context, id string) (*domain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.securities[id]; ok {
		out := *s
		out.ProviderMappings = r.mappingsForLocked(id)
		return &out, nil
	}
	return nil, nil
}

func (r *fakeRepo) GetSecurityByIMSSymbol(ctx context.Context, imsSymbol string) (*domain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.securities {
		if s.IMSSymbol == imsSymbol {
			out := *s
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
		if q != "" {
			if !strings.Contains(strings.ToLower(s.IMSSymbol), q) &&
				!strings.Contains(strings.ToLower(s.DisplaySymbol), q) &&
				!strings.Contains(strings.ToLower(s.Name), q) {
				continue
			}
		}
		out = append(out, *s)
	}
	return out, nil
}

func (r *fakeRepo) AddProviderMapping(ctx context.Context, m domain.ProviderMapping) (*domain.ProviderMapping, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.mappings {
		if existing.ProviderCode == m.ProviderCode && existing.ProviderSymbol == m.ProviderSymbol {
			existing.SecurityID = m.SecurityID
			existing.MappingStatus = m.MappingStatus
			if m.ConfidenceScore.IsZero() {
				existing.ConfidenceScore = decimal.NewFromInt(100)
			} else {
				existing.ConfidenceScore = m.ConfidenceScore
			}
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
	if m, ok := r.mappings[id]; ok {
		out := *m
		return &out, nil
	}
	return nil, nil
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
	if c, ok := r.candidates[id]; ok {
		out := *c
		return &out, nil
	}
	return nil, nil
}

func (r *fakeRepo) ListUnmappedCandidates(ctx context.Context, filter domain.UnmappedCandidateFilter) ([]domain.UnmappedSecurityCandidate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []domain.UnmappedSecurityCandidate{}
	for _, c := range r.candidates {
		if filter.Status != "" && c.CandidateStatus != filter.Status {
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
