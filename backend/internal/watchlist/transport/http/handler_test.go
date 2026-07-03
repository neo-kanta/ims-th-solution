package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// ── Test 3: metric_type validation ───────────────────────────────────────────

// Test 3 (P1-3): parseThresholdRuleInputs rejects unsupported metric_type values.
func TestParseThresholdRuleInputs_RejectsUnsupportedMetricType(t *testing.T) {
	badType := "NAV"
	reqs := []ThresholdRuleRequest{
		{MetricType: &badType, Direction: "ABOVE", ThresholdValue: "100"},
	}
	_, err := parseThresholdRuleInputs(reqs)
	if err == nil {
		t.Fatal("expected error for unsupported metric_type, got nil")
	}
	if !strings.Contains(err.Error(), "metric_type") {
		t.Errorf("error message %q should mention metric_type", err.Error())
	}
}

func TestParseThresholdRuleInputs_AcceptsMarketPrice(t *testing.T) {
	mp := "MARKET_PRICE"
	reqs := []ThresholdRuleRequest{
		{MetricType: &mp, Direction: "ABOVE", ThresholdValue: "100"},
	}
	out, err := parseThresholdRuleInputs(reqs)
	if err != nil {
		t.Fatalf("unexpected error for MARKET_PRICE: %v", err)
	}
	if len(out) != 1 {
		t.Errorf("expected 1 result, got %d", len(out))
	}
}

func TestParseThresholdRuleInputs_NilMetricTypeAccepted(t *testing.T) {
	reqs := []ThresholdRuleRequest{
		{Direction: "BELOW", ThresholdValue: "50"},
	}
	if _, err := parseThresholdRuleInputs(reqs); err != nil {
		t.Fatalf("unexpected error with nil metric_type: %v", err)
	}
}

func TestParseThresholdRuleInputs_RejectsInvalidStatus(t *testing.T) {
	status := "PAUSED"
	reqs := []ThresholdRuleRequest{
		{Direction: "ABOVE", ThresholdValue: "100", Status: &status},
	}
	_, err := parseThresholdRuleInputs(reqs)
	if err == nil {
		t.Fatal("expected invalid status error, got nil")
	}
	if !strings.Contains(err.Error(), "status") {
		t.Errorf("error message %q should mention status", err.Error())
	}
}

func TestParseThresholdRuleInputs_RejectsNegativeCooldown(t *testing.T) {
	cooldown := -1
	reqs := []ThresholdRuleRequest{
		{Direction: "ABOVE", ThresholdValue: "100", CooldownMinutes: &cooldown},
	}
	_, err := parseThresholdRuleInputs(reqs)
	if err == nil {
		t.Fatal("expected negative cooldown error, got nil")
	}
	if !strings.Contains(err.Error(), "cooldown_minutes") {
		t.Errorf("error message %q should mention cooldown_minutes", err.Error())
	}
}

// ── Test 9: portfolio descriptor includes display_name ────────────────────────

// Test 9 (P1-3): portfolioDescriptorFromScope populates display_name.
func TestPortfolioDescriptorFromScope_IncludesDisplayName(t *testing.T) {
	info := &watchlistdomain.PortfolioScopeInfo{
		PortfolioID:   uuid.New(),
		PortfolioCode: "PORT01",
		PortfolioName: "Growth Portfolio",
		FundID:        uuid.New(),
		FundCode:      "FUND01",
		FundName:      "Growth Fund",
	}
	desc := portfolioDescriptorFromScope(info)
	if desc == nil {
		t.Fatal("expected non-nil descriptor")
	}
	if desc.DisplayName == "" {
		t.Error("DisplayName is empty; expected non-empty composite name")
	}
	if !strings.Contains(desc.DisplayName, info.PortfolioCode) {
		t.Errorf("DisplayName %q does not contain PortfolioCode %q", desc.DisplayName, info.PortfolioCode)
	}
	if !strings.Contains(desc.DisplayName, info.PortfolioName) {
		t.Errorf("DisplayName %q does not contain PortfolioName %q", desc.DisplayName, info.PortfolioName)
	}
}

func TestPortfolioDescriptorFromScope_NilReturnsNil(t *testing.T) {
	if got := portfolioDescriptorFromScope(nil); got != nil {
		t.Errorf("expected nil for nil input, got %+v", got)
	}
}

// ── Test 13: unauthenticated code is UNAUTHORIZED ────────────────────────────

// Test 13 (P3-1): handler returns code "UNAUTHORIZED" (not "WATCHLIST_UNAUTHENTICATED")
// when no auth context is present.
func TestListItems_Unauthenticated_ReturnsUNAUTHORIZED(t *testing.T) {
	h := &Handler{} // all nil fields — returns early before using any of them

	req := httptest.NewRequest(http.MethodGet, "/watchlists", nil)
	// Intentionally no UserClaims in context.
	w := httptest.NewRecorder()

	h.ListItems(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}

	var body httputil.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.Code != "UNAUTHORIZED" {
		t.Errorf("code = %q, want \"UNAUTHORIZED\"", body.Code)
	}
}

func TestCreateItem_Unauthenticated_ReturnsUNAUTHORIZED(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/watchlists/items", strings.NewReader("{}"))
	w := httptest.NewRecorder()

	h.CreateItem(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	var body httputil.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.Code != "UNAUTHORIZED" {
		t.Errorf("code = %q, want \"UNAUTHORIZED\"", body.Code)
	}
}

func TestCreateItem_InvalidThresholdPayload_ReturnsWatchlistInvalidThreshold(t *testing.T) {
	actorID := uuid.New()
	status := "PAUSED"
	body := `{"scope_type":"PERSONAL","security_id":"` + uuid.New().String() + `","threshold_rules":[{"direction":"ABOVE","threshold_value":"100","status":"` + status + `"}]}`
	req := withActor(httptest.NewRequest(http.MethodPost, "/watchlists/items", strings.NewReader(body)), actorID)
	w := httptest.NewRecorder()

	h := &Handler{}
	h.CreateItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
	var resp httputil.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != "WATCHLIST_INVALID_THRESHOLD" {
		t.Errorf("code = %q, want WATCHLIST_INVALID_THRESHOLD", resp.Code)
	}
}

func TestListItems_PortfolioIDWithoutPortfolioScope_ReturnsWatchlistInvalidQuery(t *testing.T) {
	actorID := uuid.New()
	req := withActor(httptest.NewRequest(http.MethodGet, "/watchlists?portfolio_id="+uuid.New().String(), nil), actorID)
	w := httptest.NewRecorder()

	h := &Handler{}
	h.ListItems(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
	var resp httputil.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != "WATCHLIST_INVALID_QUERY" {
		t.Errorf("code = %q, want WATCHLIST_INVALID_QUERY", resp.Code)
	}
}

func TestListAlerts_PortfolioIDWithoutPortfolioScope_ReturnsWatchlistInvalidQuery(t *testing.T) {
	actorID := uuid.New()
	req := withActor(httptest.NewRequest(http.MethodGet, "/watchlists/alerts?portfolio_id="+uuid.New().String(), nil), actorID)
	w := httptest.NewRecorder()

	h := &Handler{}
	h.ListAlerts(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", w.Code, w.Body.String())
	}
	var resp httputil.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != "WATCHLIST_INVALID_QUERY" {
		t.Errorf("code = %q, want WATCHLIST_INVALID_QUERY", resp.Code)
	}
}

func withActor(req *http.Request, actorID uuid.UUID) *http.Request {
	claims := &middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: actorID.String()},
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	return req.WithContext(ctx)
}
