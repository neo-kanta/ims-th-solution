package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// stubPositionRepo, stubCashRepo, stubTxnRepo, and stubValuationRepo are
// minimal domain repository stubs for the Milestone 2 V2 read-endpoint
// tests (docs/handoff/portfolio-v2-claude-implementation-prompt.md). Only
// the methods each handler actually calls are exercised.

type stubPositionRepo struct {
	byPortfolio map[uuid.UUID][]*entity.PortfolioPosition
}

func (r *stubPositionRepo) GetForUpdate(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (*entity.PortfolioPosition, error) {
	return nil, nil
}
func (r *stubPositionRepo) Upsert(context.Context, pgx.Tx, *entity.PortfolioPosition, int) error {
	return nil
}
func (r *stubPositionRepo) ListByPortfolio(_ context.Context, portfolioID uuid.UUID) ([]*entity.PortfolioPosition, error) {
	return r.byPortfolio[portfolioID], nil
}

type stubCashRepo struct {
	byPortfolio map[uuid.UUID][]*entity.CashBalance
}

func (r *stubCashRepo) InsertMovement(context.Context, pgx.Tx, *entity.CashMovement) error {
	return nil
}
func (r *stubCashRepo) GetBalanceForUpdate(context.Context, pgx.Tx, uuid.UUID, string) (*entity.CashBalance, error) {
	return nil, nil
}
func (r *stubCashRepo) UpsertBalance(context.Context, pgx.Tx, *entity.CashBalance, int) error {
	return nil
}
func (r *stubCashRepo) ListBalances(_ context.Context, portfolioID uuid.UUID) ([]*entity.CashBalance, error) {
	return r.byPortfolio[portfolioID], nil
}

type stubTxnRepo struct {
	byPortfolio map[uuid.UUID][]*entity.PortfolioTransaction
}

func (r *stubTxnRepo) Insert(context.Context, pgx.Tx, *entity.PortfolioTransaction) error {
	return nil
}
func (r *stubTxnRepo) GetByID(context.Context, uuid.UUID) (*entity.PortfolioTransaction, error) {
	return nil, nil
}
func (r *stubTxnRepo) List(_ context.Context, filter domain.TransactionListFilter) ([]*entity.PortfolioTransaction, int, error) {
	if filter.PortfolioID == nil {
		return nil, 0, nil
	}
	rows := r.byPortfolio[*filter.PortfolioID]
	return rows, len(rows), nil
}
func (r *stubTxnRepo) ListForPositionReplay(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) ([]*entity.PortfolioTransaction, error) {
	return nil, nil
}
func (r *stubTxnRepo) HasReversal(context.Context, uuid.UUID) (bool, error) { return false, nil }
func (r *stubTxnRepo) SumRealisedPnLBase(context.Context, uuid.UUID, time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

type stubValuationRepo struct {
	latest map[uuid.UUID]*entity.ValuationSnapshot
	list   map[uuid.UUID][]*entity.ValuationSnapshot
}

func (r *stubValuationRepo) Insert(context.Context, pgx.Tx, *entity.ValuationSnapshot) error {
	return nil
}
func (r *stubValuationRepo) GetLatest(_ context.Context, portfolioID uuid.UUID, _ vo.ValuationSource) (*entity.ValuationSnapshot, error) {
	return r.latest[portfolioID], nil
}
func (r *stubValuationRepo) GetByID(context.Context, uuid.UUID) (*entity.ValuationSnapshot, error) {
	return nil, nil
}
func (r *stubValuationRepo) GetByPortfolioBusinessDate(context.Context, uuid.UUID, time.Time, vo.ValuationSource) (*entity.ValuationSnapshot, error) {
	return nil, nil
}
func (r *stubValuationRepo) List(_ context.Context, portfolioID uuid.UUID, _, _ time.Time, _, _ int) ([]*entity.ValuationSnapshot, int, error) {
	rows := r.list[portfolioID]
	return rows, len(rows), nil
}
func (r *stubValuationRepo) InsertNAV(context.Context, pgx.Tx, *entity.NAVSnapshot) error { return nil }
func (r *stubValuationRepo) ListNAV(context.Context, uuid.UUID, time.Time, time.Time, int, int) ([]*entity.NAVSnapshot, int, error) {
	return nil, 0, nil
}
func (r *stubValuationRepo) InsertAUM(context.Context, pgx.Tx, *entity.AUMSnapshot) error { return nil }
func (r *stubValuationRepo) GetAUM(context.Context, vo.AumScopeType, uuid.UUID, time.Time, vo.ValuationSource) (*entity.AUMSnapshot, error) {
	return nil, nil
}
func (r *stubValuationRepo) ListAUM(context.Context, vo.AumScopeType, uuid.UUID, time.Time, time.Time, int, int) ([]*entity.AUMSnapshot, int, error) {
	return nil, 0, nil
}

// v2ReadFixture builds an InvestmentHandler wired with an alive portfolio
// (accessible to the caller) plus whatever read-repo stubs the test needs.
type v2ReadFixture struct {
	handler   *InvestmentHandler
	portfolio *entity.Portfolio
}

func newV2ReadFixture(t *testing.T) *v2ReadFixture {
	t.Helper()
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
	repo := &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{"TH-EQ-01": portfolio}}
	return &v2ReadFixture{
		handler:   &InvestmentHandler{pc: pc, portfolios: repo},
		portfolio: portfolio,
	}
}

func serveV2(h *InvestmentHandler, method, path string, wire func(chi.Router)) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Route("/portfolios/{portfolioCode}", func(r chi.Router) {
		wire(r)
	})
	req := withUserClaims(httptest.NewRequest(method, path, nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestGetHoldingsByCode_ValidCode(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.positions = &stubPositionRepo{byPortfolio: map[uuid.UUID][]*entity.PortfolioPosition{
		f.portfolio.ID: {{ID: uuid.New(), PortfolioID: f.portfolio.ID, InstrumentID: uuid.New()}},
	}}

	w := serveV2(f.handler, http.MethodGet, "/portfolios/TH-EQ-01/holdings", func(r chi.Router) {
		r.Get("/holdings", f.handler.GetHoldingsByCode)
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestGetHoldingsByCode_UnknownCodeReturns404(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.positions = &stubPositionRepo{}

	w := serveV2(f.handler, http.MethodGet, "/portfolios/NOPE/holdings", func(r chi.Router) {
		r.Get("/holdings", f.handler.GetHoldingsByCode)
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestGetHoldingsByCode_AmbiguousCodeReturns409(t *testing.T) {
	t.Parallel()
	pc := &fakeV2PermissionChecker{}
	h := &InvestmentHandler{pc: pc, portfolios: &stubPortfolioRepo{
		byCodeErr: &domain.ErrAmbiguousPortfolioCode{Code: "DUP", Count: 2},
	}, positions: &stubPositionRepo{}}

	w := serveV2(h, http.MethodGet, "/portfolios/DUP/holdings", func(r chi.Router) {
		r.Get("/holdings", h.GetHoldingsByCode)
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body=%s", w.Code, w.Body.String())
	}
}

func TestGetCashByCode_ValidCode(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.cash = &stubCashRepo{byPortfolio: map[uuid.UUID][]*entity.CashBalance{
		f.portfolio.ID: {{ID: uuid.New(), PortfolioID: f.portfolio.ID, Currency: "THB"}},
	}}

	w := serveV2(f.handler, http.MethodGet, "/portfolios/TH-EQ-01/cash", func(r chi.Router) {
		r.Get("/cash", f.handler.GetCashByCode)
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestListTransactionsByCode_ValidCode(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.txns = &stubTxnRepo{byPortfolio: map[uuid.UUID][]*entity.PortfolioTransaction{
		f.portfolio.ID: {{ID: uuid.New(), PortfolioID: f.portfolio.ID, FundID: f.portfolio.FundID}},
	}}

	w := serveV2(f.handler, http.MethodGet, "/portfolios/TH-EQ-01/transactions", func(r chi.Router) {
		r.Get("/transactions", f.handler.ListTransactionsByCode)
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestListTransactionsByCode_UnknownCodeReturns404(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.portfolios = &stubPortfolioRepo{byCode: map[string]*entity.Portfolio{}}
	f.handler.txns = &stubTxnRepo{}

	w := serveV2(f.handler, http.MethodGet, "/portfolios/NOPE/transactions", func(r chi.Router) {
		r.Get("/transactions", f.handler.ListTransactionsByCode)
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestListValuationsByCode_ValidCode(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.valuation = &stubValuationRepo{list: map[uuid.UUID][]*entity.ValuationSnapshot{
		f.portfolio.ID: {{ID: uuid.New(), PortfolioID: f.portfolio.ID}},
	}}

	w := serveV2(f.handler, http.MethodGet, "/portfolios/TH-EQ-01/valuations", func(r chi.Router) {
		r.Get("/valuations", f.handler.ListValuationsByCode)
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestGetLatestValuationByCode_ValidCode(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.valuation = &stubValuationRepo{latest: map[uuid.UUID]*entity.ValuationSnapshot{
		f.portfolio.ID: {ID: uuid.New(), PortfolioID: f.portfolio.ID},
	}}

	w := serveV2(f.handler, http.MethodGet, "/portfolios/TH-EQ-01/valuations/latest", func(r chi.Router) {
		r.Get("/valuations/latest", f.handler.GetLatestValuationByCode)
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}

func TestGetLatestValuationByCode_NoSnapshotReturns404(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.valuation = &stubValuationRepo{}

	w := serveV2(f.handler, http.MethodGet, "/portfolios/TH-EQ-01/valuations/latest", func(r chi.Router) {
		r.Get("/valuations/latest", f.handler.GetLatestValuationByCode)
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", w.Code, w.Body.String())
	}
}

func TestListTransactionsByCode_NonUUIDCodeWorks(t *testing.T) {
	t.Parallel()
	f := newV2ReadFixture(t)
	f.handler.txns = &stubTxnRepo{byPortfolio: map[uuid.UUID][]*entity.PortfolioTransaction{}}

	w := serveV2(f.handler, http.MethodGet, "/portfolios/TH-EQ-01/transactions", func(r chi.Router) {
		r.Get("/transactions", f.handler.ListTransactionsByCode)
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
}
