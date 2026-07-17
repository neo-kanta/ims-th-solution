package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/request"
)

// These tests cover the portion of Milestone 3
// (docs/handoff/portfolio-v2-claude-implementation-prompt.md) that is
// reachable without a live Postgres pool: portfolioCode resolution fails
// before any of the write handlers touch h.postTxn/h.reverseTxn (both nil
// in these fixtures), so 404/409 behavior is fully exercised here. The
// happy-path simulate/post/reverse round trip and the "extra fund_id is
// derived server-side, never trusted" guarantee are covered end-to-end
// against a live DB by tests/e2e/portfolio_v2_ledger_test.go
// (TestE2E_PortfolioV2_LedgerWrites), since PostTransactionHandler and
// ReverseTransactionHandler require a real pgx transaction.

func TestSimulateTransactionByCode_UnknownCodeReturns404(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.portfolios = &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{}}

	w := serveV2(f.handler, http.MethodPost, "/portfolios/NOPE/transactions/simulate", func(r chi.Router) {
		r.Post("/transactions/simulate", f.handler.SimulateTransactionByCode)
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestPostTransactionByCode_UnknownCodeReturns404(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.portfolios = &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{}}

	w := serveV2(f.handler, http.MethodPost, "/portfolios/NOPE/transactions", func(r chi.Router) {
		r.Post("/transactions", f.handler.PostTransactionByCode)
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestReverseTransactionByCode_UnknownCodeReturns404(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.portfolios = &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{}}

	w := serveV2(f.handler, http.MethodPost, "/portfolios/NOPE/transactions/00000000-0000-0000-0000-000000000000/reverse", func(r chi.Router) {
		r.Post("/transactions/{transactionId}/reverse", f.handler.ReverseTransactionByCode)
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestPostTransactionByCode_AmbiguousCodeReturns409(t *testing.T) {
	t.Parallel()
	pc := &fakeV2PermissionChecker{}
	h := &InvestmentHandler{pc: pc, portfolios: &stubPortfolioRepo{
		byCodeErr: &domain.ErrAmbiguousPortfolioCode{Code: "DUP", Count: 2},
	}}

	w := serveV2(h, http.MethodPost, "/portfolios/DUP/transactions", func(r chi.Router) {
		r.Post("/transactions", h.PostTransactionByCode)
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body=%s", w.Code, w.Body.String())
	}
}

func TestSimulateTransactionByCode_AmbiguousCodeReturns409(t *testing.T) {
	t.Parallel()
	pc := &fakeV2PermissionChecker{}
	h := &InvestmentHandler{pc: pc, portfolios: &stubPortfolioRepo{
		byCodeErr: &domain.ErrAmbiguousPortfolioCode{Code: "DUP", Count: 2},
	}}

	w := serveV2(h, http.MethodPost, "/portfolios/DUP/transactions/simulate", func(r chi.Router) {
		r.Post("/transactions/simulate", h.SimulateTransactionByCode)
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body=%s", w.Code, w.Body.String())
	}
}

// TestPostTransactionRequest_HasNoLegacyIdentityFields locks in the
// Milestone 3 rule that V2 request bodies must not accept fund_id or
// contract_id: request.PostTransactionRequest and
// request.ReverseTransactionRequest carry no such field, so the JSON
// decoder silently drops any client-supplied fund_id/contract_id key
// instead of trusting it. A future accidental field addition would flip
// this test red.
func TestPostTransactionRequest_HasNoLegacyIdentityFields(t *testing.T) {
	t.Parallel()
	buf, err := json.Marshal(request.PostTransactionRequest{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(buf)
	for _, forbidden := range []string{"fund_id", "contract_id"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("request.PostTransactionRequest JSON contains %q; V2 write bodies must not carry legacy identity fields", forbidden)
		}
	}
}

func TestReverseTransactionRequest_HasNoLegacyIdentityFields(t *testing.T) {
	t.Parallel()
	buf, err := json.Marshal(request.ReverseTransactionRequest{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(buf)
	for _, forbidden := range []string{"fund_id", "contract_id"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("request.ReverseTransactionRequest JSON contains %q; V2 write bodies must not carry legacy identity fields", forbidden)
		}
	}
}
