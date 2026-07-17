package transport

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/engine"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

type noApplicableBindingRepo struct{}

func (*noApplicableBindingRepo) ResolveApplicable(context.Context, []vo.Scope, time.Time) ([]domain.ResolvedBinding, error) {
	return nil, nil
}
func (*noApplicableBindingRepo) Create(context.Context, *entity.RuleBinding) error { return nil }
func (*noApplicableBindingRepo) GetByID(context.Context, uuid.UUID) (*entity.RuleBinding, error) {
	return nil, nil
}
func (*noApplicableBindingRepo) List(context.Context, domain.BindingFilter) ([]entity.RuleBinding, int64, error) {
	return nil, 0, nil
}
func (*noApplicableBindingRepo) Deactivate(context.Context, uuid.UUID) error { return nil }

type noWriteCheckRepo struct{}

func (*noWriteCheckRepo) Create(context.Context, *entity.CheckRecord) error { return nil }
func (*noWriteCheckRepo) CreateBatch(context.Context, []entity.CheckRecord) error {
	panic("no check record may be written when compliance is not configured")
}
func (*noWriteCheckRepo) GetByID(context.Context, uuid.UUID) (*entity.CheckRecord, error) {
	return nil, nil
}
func (*noWriteCheckRepo) GetByGroupID(context.Context, uuid.UUID) ([]entity.CheckRecord, error) {
	return nil, nil
}
func (*noWriteCheckRepo) List(context.Context, domain.CheckRecordFilter) ([]entity.CheckRecord, int64, error) {
	return nil, 0, nil
}

type noWriteBreachRepo struct{}

func (*noWriteBreachRepo) Create(context.Context, *entity.Breach) error {
	panic("no breach may be written when compliance is not configured")
}
func (*noWriteBreachRepo) GetByID(context.Context, uuid.UUID) (*entity.Breach, error) {
	return nil, nil
}
func (*noWriteBreachRepo) List(context.Context, domain.BreachFilter) ([]entity.Breach, int64, error) {
	return nil, 0, nil
}
func (*noWriteBreachRepo) UpdateStatus(context.Context, uuid.UUID, entity.BreachStatus, *uuid.UUID, *time.Time) error {
	return nil
}

func TestComplianceContractAdapter_NotConfiguredStatusSurvivesBoundary(t *testing.T) {
	t.Parallel()
	registry := spi.NewRegistry()
	pipeline := engine.NewPipeline(
		registry,
		&noApplicableBindingRepo{},
		&noWriteCheckRepo{},
		&noWriteBreachRepo{},
		engine.NewFetcher(nil, nil, nil, nil, nil, nil, nil, nil),
	)
	preTrade := command.NewRunPreTradeCheckHandler(pipeline, registry)
	adapter := NewComplianceContractAdapter(preTrade, pipeline, registry)

	result, err := adapter.CheckProposedOrder(context.Background(), contract.ProposedOrderCheck{
		PortfolioID:  uuid.New(),
		ContractID:   uuid.New(),
		BusinessDate: time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC),
		Actor:        "portfolio-manager",
		OrderID:      uuid.New(),
		Ticker:       "PTT",
		Side:         contract.ComplianceOrderSideBuy,
		Quantity:     decimal.NewFromInt(100),
		Price:        decimal.NewFromInt(35),
		Currency:     "THB",
		Exchange:     "SET",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != contract.ComplianceStatusNotConfigured {
		t.Fatalf("status = %q, want %q", result.Status, contract.ComplianceStatusNotConfigured)
	}
	if result.Status == contract.ComplianceStatusEvaluated {
		t.Fatal("zero applicable bindings must not cross the contract as an evaluated PASS")
	}
	if result.RulesEvaluated != 0 {
		t.Fatalf("rules_evaluated = %d, want 0", result.RulesEvaluated)
	}
	if result.CheckGroupID == uuid.Nil {
		t.Fatal("check_group_id must remain available for correlation")
	}
}
