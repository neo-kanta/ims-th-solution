package adapter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
)

// ─────────────────────────────────────────────────────────────────────────────
// stub fund repository — satisfies domain.FundRepository
// ─────────────────────────────────────────────────────────────────────────────

type stubFundRepo struct {
	byContractCode map[string]*entity.Fund
	err            error
}

func newStubFundRepo() *stubFundRepo { return &stubFundRepo{byContractCode: map[string]*entity.Fund{}} }

func (s *stubFundRepo) addFund(contractCode string, id uuid.UUID) {
	s.byContractCode[contractCode] = &entity.Fund{
		ID:           id,
		Code:         "FUND-" + contractCode,
		Name:         "Test Fund " + contractCode,
		BaseCurrency: "THB",
		Status:       vo.FundStatusActive,
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func (s *stubFundRepo) GetByContractCode(_ context.Context, contractCode string) (*entity.Fund, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.byContractCode[contractCode], nil
}

func (s *stubFundRepo) Create(_ context.Context, _ pgx.Tx, _ *entity.Fund) error { return nil }
func (s *stubFundRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Fund, error) {
	return nil, nil
}
func (s *stubFundRepo) GetByCode(_ context.Context, _ string) (*entity.Fund, error) {
	return nil, nil
}
func (s *stubFundRepo) List(_ context.Context, _ domain.FundListFilter) ([]*entity.Fund, int, error) {
	return nil, 0, nil
}
func (s *stubFundRepo) Update(_ context.Context, _ pgx.Tx, _ *entity.Fund) error { return nil }
func (s *stubFundRepo) SoftDelete(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ int, _ uuid.UUID) error {
	return nil
}
func (s *stubFundRepo) CountActivePortfolios(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

var _ domain.FundRepository = (*stubFundRepo)(nil)

// ─────────────────────────────────────────────────────────────────────────────
// tests
// ─────────────────────────────────────────────────────────────────────────────

func TestContractCatalogAdapter_ResolvesActiveContractCode(t *testing.T) {
	repo := newStubFundRepo()
	id := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	repo.addFund("ABC", id)

	a := adapter.NewContractCatalogAdapter(repo)
	got, err := a.ResolveContractCode(context.Background(), "ABC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != id {
		t.Errorf("want %s, got %s", id, got)
	}
}

func TestContractCatalogAdapter_EmptyCodeReturnsValidationError(t *testing.T) {
	a := adapter.NewContractCatalogAdapter(newStubFundRepo())
	_, err := a.ResolveContractCode(context.Background(), "")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	code, _ := errcode.CodeOf(err)
	if code != errcode.CodeInvalidRequest {
		t.Errorf("want %s error code, got %s", errcode.CodeInvalidRequest, code)
	}
}

func TestContractCatalogAdapter_UnknownCodeReturnsNotFound(t *testing.T) {
	a := adapter.NewContractCatalogAdapter(newStubFundRepo())
	_, err := a.ResolveContractCode(context.Background(), "UNKNOWN")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	code, _ := errcode.CodeOf(err)
	if code != errcode.CodeContractNotFound {
		t.Errorf("want %s error code, got %s", errcode.CodeContractNotFound, code)
	}
}

func TestContractCatalogAdapter_NeverReturnsNilUUIDOnSuccess(t *testing.T) {
	repo := newStubFundRepo()
	repo.addFund("XYZ", uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000002"))
	a := adapter.NewContractCatalogAdapter(repo)

	got, err := a.ResolveContractCode(context.Background(), "XYZ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == uuid.Nil {
		t.Error("want non-nil UUID, got uuid.Nil")
	}
}

func TestContractCatalogAdapter_RepositoryErrorPropagates(t *testing.T) {
	repo := newStubFundRepo()
	repo.err = errors.New("db connection lost")
	a := adapter.NewContractCatalogAdapter(repo)

	_, err := a.ResolveContractCode(context.Background(), "ABC")
	if err == nil {
		t.Fatal("want error, got nil")
	}
}

func TestContractCatalogAdapter_ImplementsContractCatalog(t *testing.T) {
	// Compile-time assertion: ContractCatalogAdapter must implement contract.ContractCatalog.
	var _ contract.ContractCatalog = adapter.NewContractCatalogAdapter(newStubFundRepo())
}
