package httptransport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
	refdomain "github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
)

func TestScreenWatchlistReturnsCanonicalRowDTO(t *testing.T) {
	repo := &stubRepository{}
	resolver := &stubResolver{
		securities: []refdomain.Security{{
			ID:            "sec-kbank",
			IMSSymbol:     "TH_EQ_XBKK_KBANK",
			DisplaySymbol: "KBANK.BK",
			Name:          "Kasikornbank PCL",
			AssetType:     refdomain.AssetTypeEquity,
			Currency:      "THB",
			ExchangeMIC:   "XBKK",
			Status:        refdomain.SecurityStatusActive,
			ProviderMappings: []refdomain.ProviderMapping{
				{ProviderCode: "yahoo", MappingStatus: refdomain.MappingStatusActive},
				{ProviderCode: "alpha_vantage", MappingStatus: refdomain.MappingStatusActive},
			},
		}},
	}
	// Plant a latest snapshot for display symbol matching.
	repo.latestQuoteFor = map[string]*domain.Quote{
		"KBANK.BK": {
			Symbol: "KBANK.BK",
			Price:  decimal.RequireFromString("142.50"),
			Change: decimal.RequireFromString("1.50"),
			Volume: 12400000,
			AsOf:   time.Now().Add(-1 * time.Minute).UTC(),
		},
	}

	service := application.NewService(application.Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, nil, repo, repo, nil)
	service.SetSecurityResolver(resolver)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/market-data/screen/watchlist", nil)
	rec := httptest.NewRecorder()
	handler.GetScreenWatchlist(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var resp struct {
		Data ScreenWatchlistResponseDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data.Items, 1)
	row := resp.Data.Items[0]
	require.Equal(t, "TH_EQ_XBKK_KBANK", row.IMSSymbol)
	require.Equal(t, "KBANK.BK", row.DisplaySymbol)
	require.Equal(t, "EQUITY", row.AssetType)
	require.Equal(t, "THB", row.Currency)
	require.NotNil(t, row.LastPrice)
	require.Equal(t, "142.5", *row.LastPrice)
	require.Equal(t, application.FreshnessFresh, row.FreshnessStatus)
	require.Equal(t, application.DataQualityOK, row.DataQualityStatus)
	require.Len(t, row.ProviderBadges, 2)
	require.Equal(t, "Yahoo", findBadgeLabel(row.ProviderBadges, "yahoo"))
	require.Equal(t, "Alpha", findBadgeLabel(row.ProviderBadges, "alpha_vantage"))
}

func TestScreenSearchReturnsActionHint(t *testing.T) {
	repo := &stubRepository{}
	resolver := &stubResolver{
		securities: []refdomain.Security{
			{
				ID:               "sec-mapped",
				IMSSymbol:        "TH_EQ_XBKK_KBANK",
				DisplaySymbol:    "KBANK.BK",
				Name:             "Kasikornbank",
				AssetType:        refdomain.AssetTypeEquity,
				Status:           refdomain.SecurityStatusActive,
				ProviderMappings: []refdomain.ProviderMapping{{ProviderCode: "yahoo", MappingStatus: refdomain.MappingStatusActive}},
			},
			{
				ID:            "sec-unmapped",
				IMSSymbol:     "TH_EQ_XBKK_NEW",
				DisplaySymbol: "NEW.BK",
				Name:          "New Co",
				AssetType:     refdomain.AssetTypeEquity,
				Status:        refdomain.SecurityStatusActive,
			},
		},
	}
	service := application.NewService(application.Config{
		PrimaryProvider:  domain.ProviderAlphaVantage,
		FallbackProvider: domain.ProviderYahoo,
		HTTPTimeout:      time.Second,
		CacheTTL:         time.Minute,
	}, nil, repo, repo, nil)
	service.SetSecurityResolver(resolver)
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/market-data/screen/search?query=BK", nil)
	rec := httptest.NewRecorder()
	handler.GetScreenSearch(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Data ScreenSearchResponseDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data.Items, 2)
	hints := map[string]string{}
	for _, item := range resp.Data.Items {
		hints[item.IMSSymbol] = item.ActionHint
	}
	require.Equal(t, "ADD_TO_WATCHLIST", hints["TH_EQ_XBKK_KBANK"])
	require.Equal(t, "MAP_SYMBOL", hints["TH_EQ_XBKK_NEW"])
}

func findBadgeLabel(badges []ProviderBadgeDTO, code string) string {
	for _, b := range badges {
		if b.ProviderCode == code {
			return b.Label
		}
	}
	return ""
}

// ---- stubResolver — minimal SecurityResolver for screen tests ----

type stubResolver struct {
	mu         sync.Mutex
	securities []refdomain.Security
}

func (r *stubResolver) SearchSecurities(ctx context.Context, filter refdomain.SecuritySearchFilter) ([]refdomain.Security, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []refdomain.Security{}
	for _, s := range r.securities {
		if filter.Query != "" && !contains(s.DisplaySymbol, filter.Query) && !contains(s.Name, filter.Query) {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *stubResolver) GetSecurityByID(ctx context.Context, id string) (*refdomain.Security, error) {
	for _, s := range r.securities {
		if s.ID == id {
			out := s
			return &out, nil
		}
	}
	return nil, nil
}
func (r *stubResolver) GetSecurityByIMSSymbol(ctx context.Context, _ string) (*refdomain.Security, error) {
	return nil, nil
}
func (r *stubResolver) GetSecurityByDisplaySymbol(ctx context.Context, _ string) (*refdomain.Security, error) {
	return nil, nil
}
func (r *stubResolver) ResolveSecurityByProviderSymbol(_ context.Context, _, _ string) (*refdomain.Security, error) {
	return nil, nil
}
func (r *stubResolver) ResolveProviderSymbol(_ context.Context, _, _ string) (*refdomain.ProviderMapping, error) {
	return nil, nil
}
func (r *stubResolver) CreateUnmappedCandidate(_ context.Context, c refdomain.UnmappedSecurityCandidate) (*refdomain.UnmappedSecurityCandidate, error) {
	c.ID = uuid.NewString()
	return &c, nil
}
func (r *stubResolver) ListUnmappedCandidates(_ context.Context, _ refdomain.UnmappedCandidateFilter) ([]refdomain.UnmappedSecurityCandidate, error) {
	return nil, nil
}
func (r *stubResolver) MapUnmappedCandidate(_ context.Context, _, _ string) error    { return nil }
func (r *stubResolver) RejectUnmappedCandidate(_ context.Context, _, _ string) error { return nil }

func contains(s, q string) bool {
	if q == "" {
		return true
	}
	if len(q) == 0 {
		return true
	}
	return indexFold(s, q) >= 0
}

func indexFold(s, substr string) int {
	if substr == "" {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			a := s[i+j]
			b := substr[j]
			if a >= 'A' && a <= 'Z' {
				a += 'a' - 'A'
			}
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
