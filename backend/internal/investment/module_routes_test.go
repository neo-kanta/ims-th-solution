package investment

import (
	"context"
	"testing"

	"github.com/go-chi/chi/v5"
)

// fakeIAMPort is a minimal adapter.IAMPermissionPort stub used only to make
// NewModule wire a non-nil m.middlewarePerm (so RegisterRoutes/
// RegisterRoutesV2 don't early-return) — no test in this file exercises
// actual permission decisions.
type fakeIAMPort struct{}

func (fakeIAMPort) HasFunctionPermission(context.Context, string, string) (bool, error) {
	return true, nil
}
func (fakeIAMPort) HasDataPermission(context.Context, string, string) (bool, error) {
	return true, nil
}
func (fakeIAMPort) GetAccessibleContracts(context.Context, string) ([]string, error) {
	return nil, nil
}

// TestRegisterRoutes_NoPanic proves every V1 and V2 route (including the new
// Portfolio V2 endpoints added in this change set: GET/POST /portfolios,
// PATCH /portfolios/{portfolioCode}, GET .../executions[/{executionId}],
// GET .../confirmations[/{confirmationId}], POST .../valuations/run)
// registers on chi's routing tree without a pattern conflict. chi panics
// synchronously at registration time on a conflicting method+pattern, so a
// clean run here is meaningful evidence beyond `go build` succeeding.
func TestRegisterRoutes_NoPanic(t *testing.T) {
	t.Parallel()
	m := NewModule(nil, nil, nil, fakeIAMPort{}, nil)
	if m == nil || m.middlewarePerm == nil {
		t.Fatal("expected a non-nil module with middlewarePerm wired")
	}

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		m.RegisterRoutes(r)
	})
	r.Route("/api/v2", func(r chi.Router) {
		m.RegisterRoutesV2(r)
	})
}
