package command

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

func TestPostTransactionComplianceBlockStopsBeforeInsert(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)

	_, err := fixture.handler(&blockingCompliance{}).Handle(context.Background(), PostTransactionRequest{
		PortfolioID:     fixture.portfolio.ID,
		TransactionType: vo.TransactionTypeBuy,
		InstrumentID:    &fixture.instrument.ID,
		Quantity:        &qty,
		Price:           &price,
		Currency:        "THB",
		BusinessDate:    fixture.businessDate,
		ActorID:         uuid.New(),
	})

	var rejected *domain.ErrComplianceRejected
	require.ErrorAs(t, err, &rejected)
	require.Equal(t, 0, fixture.txns.inserts)
}

func TestPostTransactionRejectsSecurityAmountOverrideMismatch(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)
	badGross := decimal.NewFromInt(51)

	_, err := fixture.handler(passCompliance{}).Handle(context.Background(), PostTransactionRequest{
		PortfolioID:     fixture.portfolio.ID,
		TransactionType: vo.TransactionTypeBuy,
		InstrumentID:    &fixture.instrument.ID,
		Quantity:        &qty,
		Price:           &price,
		Currency:        "THB",
		GrossAmount:     &badGross,
		BusinessDate:    fixture.businessDate,
		ActorID:         uuid.New(),
	})

	var invalid *domain.ErrInvalidDecisionRequest
	require.ErrorAs(t, err, &invalid)
	require.Equal(t, "gross_amount", invalid.Field)
	require.Equal(t, 0, fixture.txns.inserts)
}

func TestPostTransactionRejectsPriceOffTick(t *testing.T) {
	fixture := newPostFixture()
	tick := decimal.RequireFromString("0.05")
	fixture.instrument.TickSize = &tick
	qty := decimal.NewFromInt(10)
	price := decimal.RequireFromString("10.03")

	_, err := fixture.handler(passCompliance{}).Handle(context.Background(), PostTransactionRequest{
		PortfolioID:     fixture.portfolio.ID,
		TransactionType: vo.TransactionTypeBuy,
		InstrumentID:    &fixture.instrument.ID,
		Quantity:        &qty,
		Price:           &price,
		Currency:        "THB",
		BusinessDate:    fixture.businessDate,
		ActorID:         uuid.New(),
	})

	var precondition *domain.ErrPostPreconditionFailed
	require.ErrorAs(t, err, &precondition)
	require.Equal(t, string(policy.PostViolationTickSize), precondition.Violation)
	require.Equal(t, 0, fixture.txns.inserts)
}

type postFixture struct {
	portfolio    *entity.Portfolio
	fund         *entity.Fund
	instrument   *entity.Instrument
	position     *entity.PortfolioPosition
	txns         *postTxnRepo
	businessDate time.Time
}

func newPostFixture() *postFixture {
	fundID := uuid.New()
	portfolioID := uuid.New()
	instrumentID := uuid.New()
	businessDate := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)
	return &postFixture{
		portfolio: &entity.Portfolio{
			ID: portfolioID, FundID: fundID, BaseCurrency: "THB", ValuationCurrency: "THB",
			Status: vo.PortfolioStatusActive,
		},
		fund: &entity.Fund{ID: fundID, BaseCurrency: "THB", Status: vo.FundStatusActive},
		instrument: &entity.Instrument{
			ID: instrumentID, PrimaryTicker: "ABC", Currency: "THB",
			IsTradable: true, Status: vo.InstrumentStatusActive,
		},
		position:     &entity.PortfolioPosition{PortfolioID: portfolioID, InstrumentID: instrumentID, Quantity: decimal.NewFromInt(100)},
		txns:         &postTxnRepo{},
		businessDate: businessDate,
	}
}

func (f *postFixture) handler(compliance contract.ComplianceChecker) *PostTransactionHandler {
	return NewPostTransactionHandler(
		nil,
		postPortfolioRepo{portfolio: f.portfolio},
		postFundRepo{fund: f.fund},
		postInstrumentRepo{instrument: f.instrument},
		postPositionRepo{position: f.position},
		f.txns,
		nil,
		allowWorkflow{},
		compliance,
		nilAudit{},
		func() time.Time { return f.businessDate.Add(9 * time.Hour) },
	)
}

type postPortfolioRepo struct{ portfolio *entity.Portfolio }

func (r postPortfolioRepo) Create(context.Context, pgx.Tx, *entity.Portfolio) error { return nil }
func (r postPortfolioRepo) GetByID(context.Context, uuid.UUID) (*entity.Portfolio, error) {
	return r.portfolio, nil
}
func (r postPortfolioRepo) GetByFundCode(context.Context, uuid.UUID, string) (*entity.Portfolio, error) {
	return nil, nil
}
func (r postPortfolioRepo) List(context.Context, domain.PortfolioListFilter) ([]*entity.Portfolio, int, error) {
	return nil, 0, nil
}
func (r postPortfolioRepo) Update(context.Context, pgx.Tx, *entity.Portfolio) error { return nil }
func (r postPortfolioRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r postPortfolioRepo) HasOpenActivity(context.Context, uuid.UUID, time.Time) (bool, string, error) {
	return false, "", nil
}

type postFundRepo struct{ fund *entity.Fund }

func (r postFundRepo) Create(context.Context, pgx.Tx, *entity.Fund) error { return nil }
func (r postFundRepo) GetByID(context.Context, uuid.UUID) (*entity.Fund, error) {
	return r.fund, nil
}
func (r postFundRepo) GetByCode(context.Context, string) (*entity.Fund, error) { return nil, nil }
func (r postFundRepo) List(context.Context, domain.FundListFilter) ([]*entity.Fund, int, error) {
	return nil, 0, nil
}
func (r postFundRepo) Update(context.Context, pgx.Tx, *entity.Fund) error { return nil }
func (r postFundRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, int, uuid.UUID) error {
	return nil
}
func (r postFundRepo) CountActivePortfolios(context.Context, uuid.UUID) (int, error) { return 0, nil }

type postInstrumentRepo struct{ instrument *entity.Instrument }

func (r postInstrumentRepo) Create(context.Context, pgx.Tx, *entity.Instrument) error { return nil }
func (r postInstrumentRepo) GetByID(context.Context, uuid.UUID) (*entity.Instrument, error) {
	return r.instrument, nil
}
func (r postInstrumentRepo) List(context.Context, domain.InstrumentListFilter) ([]*entity.Instrument, int, error) {
	return nil, 0, nil
}
func (r postInstrumentRepo) Update(context.Context, pgx.Tx, *entity.Instrument) error { return nil }
func (r postInstrumentRepo) SoftDelete(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) error {
	return nil
}

type postPositionRepo struct{ position *entity.PortfolioPosition }

func (r postPositionRepo) GetForUpdate(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (*entity.PortfolioPosition, error) {
	return r.position, nil
}
func (r postPositionRepo) Upsert(context.Context, pgx.Tx, *entity.PortfolioPosition, int) error {
	return nil
}
func (r postPositionRepo) ListByPortfolio(context.Context, uuid.UUID) ([]*entity.PortfolioPosition, error) {
	if r.position == nil {
		return nil, nil
	}
	return []*entity.PortfolioPosition{r.position}, nil
}

type postTxnRepo struct{ inserts int }

func (r *postTxnRepo) Insert(context.Context, pgx.Tx, *entity.PortfolioTransaction) error {
	r.inserts++
	return nil
}
func (r *postTxnRepo) GetByID(context.Context, uuid.UUID) (*entity.PortfolioTransaction, error) {
	return nil, nil
}
func (r *postTxnRepo) List(context.Context, domain.TransactionListFilter) ([]*entity.PortfolioTransaction, int, error) {
	return nil, 0, nil
}
func (r *postTxnRepo) ListForPositionReplay(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) ([]*entity.PortfolioTransaction, error) {
	return nil, nil
}
func (r *postTxnRepo) HasReversal(context.Context, uuid.UUID) (bool, error) { return false, nil }
func (r *postTxnRepo) SumRealisedPnLBase(context.Context, uuid.UUID, time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

type allowWorkflow struct{}

func (allowWorkflow) IsTradeAllowed(context.Context, uuid.UUID, time.Time) (bool, error) {
	return true, nil
}
func (allowWorkflow) IsTransactionLocked(context.Context, uuid.UUID, time.Time) (bool, error) {
	return false, nil
}

type blockingCompliance struct{}

func (blockingCompliance) CheckProposedOrder(context.Context, contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	return &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Verdict:      contract.ComplianceVerdictBlock,
		Breaches: []contract.ProposedOrderBreach{{
			RuleTypeID: "BLACKLIST", Verdict: contract.ComplianceVerdictBlock,
		}},
	}, nil
}

type passCompliance struct{}

func (passCompliance) CheckProposedOrder(context.Context, contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	return &contract.ProposedOrderResult{CheckGroupID: uuid.New(), Verdict: contract.ComplianceVerdictPass}, nil
}

type nilAudit struct{}

func (nilAudit) LogAction(contract.AuditEntry) error { return nil }
