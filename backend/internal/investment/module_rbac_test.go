package investment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	invperm "github.com/neo-kanta/ims-th-solution/backend/internal/investment/permission"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

type rbacPermissionChecker struct {
	allowed map[string]bool
	called  []string
}

func (c *rbacPermissionChecker) HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error) {
	c.called = append(c.called, code)
	return c.allowed[code], nil
}

func TestInvestmentLedgerRBACRejectsPostWithoutPermission(t *testing.T) {
	t.Parallel()
	checker := &rbacPermissionChecker{allowed: map[string]bool{}}
	status := exerciseLedgerRBAC(checker, invperm.CodeLedgerPost, http.MethodPost, "/transactions")

	if status != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", status)
	}
	if len(checker.called) != 1 || checker.called[0] != invperm.CodeLedgerPost {
		t.Fatalf("expected post permission check, got %v", checker.called)
	}
}

func TestInvestmentLedgerRBACRejectsSimulateWithoutPermission(t *testing.T) {
	t.Parallel()
	checker := &rbacPermissionChecker{allowed: map[string]bool{}}
	status := exerciseLedgerRBAC(checker, invperm.CodeLedgerSimulate, http.MethodPost, "/transactions/simulate")

	if status != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", status)
	}
	if len(checker.called) != 1 || checker.called[0] != invperm.CodeLedgerSimulate {
		t.Fatalf("expected simulate permission check, got %v", checker.called)
	}
}

func TestInvestmentLedgerRBACAllowsSimulateWithPermission(t *testing.T) {
	t.Parallel()
	checker := &rbacPermissionChecker{allowed: map[string]bool{invperm.CodeLedgerSimulate: true}}
	status := exerciseLedgerRBAC(checker, invperm.CodeLedgerSimulate, http.MethodPost, "/transactions/simulate")

	if status != http.StatusAccepted {
		t.Fatalf("expected handler reach status 202, got %d", status)
	}
}

func exerciseLedgerRBAC(
	checker *rbacPermissionChecker,
	code string,
	method string,
	path string,
) int {
	r := chi.NewRouter()
	r.With(middleware.RequirePermission(checker, code)).MethodFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	req := httptest.NewRequest(method, path, nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, &middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "00000000-0000-0000-0000-000000000001"},
	}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Result().StatusCode
}
