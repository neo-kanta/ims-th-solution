package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// stubPortfolioRepo is a minimal domain.PortfolioRepository stub for the
// Portfolio V2 handler tests. Only GetByCode is exercised; every other
// method is an unused no-op required to satisfy the interface.
type stubPortfolioRepo struct {
	byCode    map[string]*entity.Portfolio
	byCodeErr error // when set, GetByCode returns this error unconditionally
}

func (r *stubPortfolioRepo) Create(context.Context, pgx.Tx, *entity.Portfolio) error { return nil }
func (r *stubPortfolioRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.Portfolio, error) {
	for _, portfolio := range r.byCode {
		if portfolio != nil && portfolio.ID == id {
			return portfolio, nil
		}
	}
	return nil, nil
}
func (r *stubPortfolioRepo) GetByFundCode(context.Context, uuid.UUID, string) (*entity.Portfolio, error) {
	return nil, nil
}
func (r *stubPortfolioRepo) GetByCode(_ context.Context, code string) (*entity.Portfolio, error) {
	if r.byCodeErr != nil {
		return nil, r.byCodeErr
	}
	return r.byCode[code], nil
}
func (r *stubPortfolioRepo) List(context.Context, domain.PortfolioListFilter) ([]*entity.Portfolio, int, error) {
	return nil, 0, nil
}
func (r *stubPortfolioRepo) Update(context.Context, pgx.Tx, *entity.Portfolio) error { return nil }
func (r *stubPortfolioRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r *stubPortfolioRepo) HasOpenActivity(context.Context, uuid.UUID, time.Time) (bool, string, error) {
	return false, "", nil
}

func newV2PortfolioHandler(pc *fakeV2PermissionChecker, repo domain.PortfolioRepository) *InvestmentHandler {
	return &InvestmentHandler{pc: pc, portfolios: repo}
}

// fakeV2PermissionChecker grants data access only for fund IDs in Allowed.
type fakeV2PermissionChecker struct {
	Allowed map[string]bool
	// Global, when true, makes HasDataPermission grant access to every fund —
	// simulates the real AuthorizationService.HasDataPermission's "*" wildcard
	// scope handling (a caller holding the "*" data-permission scope sees
	// every contract/fund) for single-record GET handlers that call
	// hasFundAccess directly (GetDecision, GetExecution, GetConfirmation).
	Global bool
	// Contracts, when non-nil, is returned as-is by GetAccessibleContracts —
	// e.g. []string{"*"} simulates a global/company-wide data-scope caller
	// for accessibleFundIDs-based list endpoints (ListDecisions,
	// ListApprovalItems, ListPortfoliosV2). Left nil (default), list
	// endpoints see a caller with zero fund access, matching the fail-closed
	// default of a real IAM port with no granted scopes.
	Contracts []string
}

func (c *fakeV2PermissionChecker) HasFunctionPermission(uuid.UUID, string) (bool, error) {
	return true, nil
}

func (c *fakeV2PermissionChecker) HasDataPermission(_ uuid.UUID, contractID string) (bool, error) {
	if c.Global {
		return true, nil
	}
	if c.Allowed == nil {
		return false, nil
	}
	return c.Allowed[contractID], nil
}

func (c *fakeV2PermissionChecker) GetAccessibleContracts(uuid.UUID) ([]string, error) {
	return c.Contracts, nil
}

func withUserClaims(req *http.Request) *http.Request {
	claims := &middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: uuid.NewString()},
	}
	return req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, claims))
}

func serveGetPortfolioByCode(h *InvestmentHandler, path string) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Get("/portfolios/{portfolioCode}", h.GetPortfolioByCode)

	req := withUserClaims(httptest.NewRequest(http.MethodGet, path, nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestGetPortfolioByCode_ValidCodeReturnsPortfolio(t *testing.T) {
	t.Parallel()

	fundID := uuid.New()
	portfolio := &entity.Portfolio{
		ID:                uuid.New(),
		FundID:            fundID,
		Code:              "TH-EQ-01",
		Name:              "Thailand Equity Portfolio",
		BaseCurrency:      "THB",
		ValuationCurrency: "THB",
		Status:            vo.PortfolioStatusActive,
		InceptionDate:     time.Now(),
	}

	pc := &fakeV2PermissionChecker{Allowed: map[string]bool{fundID.String(): true}}
	h := newV2PortfolioHandler(pc, &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{
		"TH-EQ-01": portfolio,
	}})

	w := serveGetPortfolioByCode(h, "/portfolios/TH-EQ-01")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	var envelope struct {
		Data response.PortfolioResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	got := envelope.Data
	if got.Code != "TH-EQ-01" {
		t.Errorf("code = %q, want TH-EQ-01", got.Code)
	}
	if got.ID != portfolio.ID {
		t.Errorf("id = %s, want %s", got.ID, portfolio.ID)
	}
}

func TestGetPortfolioByCode_AmbiguousCodeReturns409(t *testing.T) {
	t.Parallel()

	pc := &fakeV2PermissionChecker{}
	h := newV2PortfolioHandler(pc, &stubPortfolioRepo{
		byCodeErr: &domain.ErrAmbiguousPortfolioCode{Code: "DUP-CODE", Count: 2},
	})

	w := serveGetPortfolioByCode(h, "/portfolios/DUP-CODE")

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body=%s", w.Code, w.Body.String())
	}
}

func TestGetPortfolioByCode_UnknownCodeReturns404(t *testing.T) {
	t.Parallel()

	pc := &fakeV2PermissionChecker{}
	h := newV2PortfolioHandler(pc, &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{}})

	w := serveGetPortfolioByCode(h, "/portfolios/NOT-A-REAL-CODE")

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestGetPortfolioByCode_DoesNotRequireUUID(t *testing.T) {
	t.Parallel()

	// A code that would fail uuid.Parse must still resolve correctly — V2
	// routes are never UUID-shaped, and the handler must not attempt to
	// parse the path param as a UUID.
	fundID := uuid.New()
	portfolio := &entity.Portfolio{
		ID:                uuid.New(),
		FundID:            fundID,
		Code:              "A02-CORE",
		BaseCurrency:      "THB",
		ValuationCurrency: "THB",
		Status:            vo.PortfolioStatusActive,
		InceptionDate:     time.Now(),
	}

	pc := &fakeV2PermissionChecker{Allowed: map[string]bool{fundID.String(): true}}
	h := newV2PortfolioHandler(pc, &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{
		"A02-CORE": portfolio,
	}})

	w := serveGetPortfolioByCode(h, "/portfolios/A02-CORE")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}
