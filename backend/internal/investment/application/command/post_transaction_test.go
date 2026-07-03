package command

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func TestPostTransactionComplianceUsesLedgerTransactionID(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)
	expectedID := uuid.New()
	checker := &capturingBlockingCompliance{}
	handler := fixture.handler(checker)
	handler.newID = func() uuid.UUID { return expectedID }

	_, err := handler.Handle(context.Background(), PostTransactionRequest{
		PortfolioID:     fixture.portfolio.ID,
		TransactionType: vo.TransactionTypeBuy,
		InstrumentID:    &fixture.instrument.ID,
		Quantity:        &qty,
		Price:           &price,
		Fees:            decimal.NewFromInt(3),
		Currency:        "THB",
		BusinessDate:    fixture.businessDate,
		ActorID:         uuid.New(),
	})

	var rejected *domain.ErrComplianceRejected
	require.ErrorAs(t, err, &rejected)
	require.Equal(t, expectedID, checker.req.OrderID)
	require.Equal(t, decimal.NewFromInt(3), checker.req.Fees)
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

func TestSimulateTransactionDoesNotMutateInvestmentRepos(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)

	res, err := fixture.handler(passCompliance{}).Simulate(context.Background(), PostTransactionRequest{
		PortfolioID:     fixture.portfolio.ID,
		TransactionType: vo.TransactionTypeBuy,
		InstrumentID:    &fixture.instrument.ID,
		Quantity:        &qty,
		Price:           &price,
		Currency:        "THB",
		BusinessDate:    fixture.businessDate,
		ActorID:         uuid.New(),
	})

	require.NoError(t, err)
	require.Equal(t, decimal.NewFromInt(50), res.GrossAmount)
	require.Equal(t, decimal.NewFromInt(-50), res.NetAmount)
	require.Equal(t, 0, fixture.txns.inserts)
	require.Equal(t, 0, fixture.cash.insertMovements)
	require.Equal(t, 0, fixture.cash.upserts)
}

func TestSimulateTransactionComplianceBlockReturnsPreview(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)

	res, err := fixture.handler(&blockingCompliance{}).Simulate(context.Background(), PostTransactionRequest{
		PortfolioID:     fixture.portfolio.ID,
		TransactionType: vo.TransactionTypeBuy,
		InstrumentID:    &fixture.instrument.ID,
		Quantity:        &qty,
		Price:           &price,
		Currency:        "THB",
		BusinessDate:    fixture.businessDate,
		ActorID:         uuid.New(),
	})

	require.NoError(t, err)
	require.NotNil(t, res.Compliance)
	require.Equal(t, contract.ComplianceVerdictBlock, res.Compliance.Verdict)
	require.Len(t, res.Compliance.Breaches, 1)
	require.Equal(t, "BLACKLIST", res.Compliance.Breaches[0].RuleTypeID)
	require.Equal(t, decimal.NewFromInt(50), res.GrossAmount)
	require.Equal(t, decimal.NewFromInt(-50), res.NetAmount)
	require.Equal(t, 0, fixture.txns.inserts)
	require.Equal(t, 0, fixture.cash.insertMovements)
	require.Equal(t, 0, fixture.cash.upserts)
}

func TestSimulateTransactionRequiresDryRunCompliance(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)
	checker := &postOnlyCompliance{}

	_, err := fixture.handler(checker).Simulate(context.Background(), PostTransactionRequest{
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
	require.Contains(t, rejected.Message, "dry-run compliance gate unavailable")
	require.Equal(t, 0, checker.calls)
	require.Equal(t, 0, fixture.txns.inserts)
	require.Equal(t, 0, fixture.cash.insertMovements)
	require.Equal(t, 0, fixture.cash.upserts)
}

func TestSimulateTransactionCashProjectionBuyAndSell(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)

	buy, err := fixture.handler(passCompliance{}).Simulate(context.Background(), PostTransactionRequest{
		PortfolioID:     fixture.portfolio.ID,
		TransactionType: vo.TransactionTypeBuy,
		InstrumentID:    &fixture.instrument.ID,
		Quantity:        &qty,
		Price:           &price,
		Currency:        "THB",
		Fees:            decimal.NewFromInt(1),
		BusinessDate:    fixture.businessDate,
		ActorID:         uuid.New(),
	})
	require.NoError(t, err)
	require.Equal(t, decimal.NewFromInt(-51), buy.Cash.CashImpact)
	require.Equal(t, decimal.NewFromInt(9949), buy.Cash.ProjectedBalance)
	require.Equal(t, decimal.NewFromInt(110), buy.Position.ProjectedQuantity)

	sell, err := fixture.handler(passCompliance{}).Simulate(context.Background(), PostTransactionRequest{
		PortfolioID:     fixture.portfolio.ID,
		TransactionType: vo.TransactionTypeSell,
		InstrumentID:    &fixture.instrument.ID,
		Quantity:        &qty,
		Price:           &price,
		Currency:        "THB",
		Fees:            decimal.NewFromInt(1),
		BusinessDate:    fixture.businessDate,
		ActorID:         uuid.New(),
	})
	require.NoError(t, err)
	require.Equal(t, decimal.NewFromInt(49), sell.Cash.CashImpact)
	require.Equal(t, decimal.NewFromInt(10049), sell.Cash.ProjectedBalance)
	require.Equal(t, decimal.NewFromInt(90), sell.Position.ProjectedQuantity)
}

func TestSimulateTransactionBlockedWhenWorkflowDayClosed(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)
	handler := fixture.handlerWithWorkflow(passCompliance{}, closedWorkflow{})

	_, err := handler.Simulate(context.Background(), PostTransactionRequest{
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
	require.Equal(t, string(policy.PostViolationTradingNotAllowed), precondition.Violation)
	require.Equal(t, 0, fixture.txns.inserts)
}

func TestPostTransactionBlockedWhenWorkflowDayClosed(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)
	handler := fixture.handlerWithWorkflow(passCompliance{}, closedWorkflow{})

	_, err := handler.Handle(context.Background(), PostTransactionRequest{
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
	require.Equal(t, string(policy.PostViolationTradingNotAllowed), precondition.Violation)
	require.Equal(t, 0, fixture.txns.inserts)
}

func TestPostTransactionWorkflowLockMissingDayBlocksBeforeComplianceAndInsert(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)
	checker := &postOnlyCompliance{}
	workflow := &workflowLockProbe{
		readAllowed: true,
		readLocked:  false,
		lock:        &contract.WorkflowDayLock{Exists: false},
	}
	handler := fixture.handlerWithWorkflow(checker, workflow)

	_, err := handler.Handle(context.Background(), PostTransactionRequest{
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
	require.Equal(t, string(policy.PostViolationTradingNotAllowed), precondition.Violation)
	require.Equal(t, 1, workflow.lockCalls)
	require.Equal(t, 0, checker.calls)
	require.Equal(t, 0, fixture.txns.inserts)
	require.Equal(t, 0, fixture.cash.insertMovements)
	require.Equal(t, 0, fixture.cash.upserts)
}

func TestPostTransactionWorkflowLockNonOpenStateBlocksBeforeComplianceAndInsert(t *testing.T) {
	fixture := newPostFixture()
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)
	checker := &postOnlyCompliance{}
	workflow := &workflowLockProbe{
		readAllowed: true,
		readLocked:  false,
		lock:        &contract.WorkflowDayLock{Exists: true, CurrentState: "MANAGER_APPROVED"},
	}
	handler := fixture.handlerWithWorkflow(checker, workflow)

	_, err := handler.Handle(context.Background(), PostTransactionRequest{
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
	require.Equal(t, string(policy.PostViolationTradingNotAllowed), precondition.Violation)
	require.Equal(t, 1, workflow.lockCalls)
	require.Equal(t, 0, checker.calls)
	require.Equal(t, 0, fixture.txns.inserts)
	require.Equal(t, 0, fixture.cash.insertMovements)
	require.Equal(t, 0, fixture.cash.upserts)
}

type postFixture struct {
	portfolio    *entity.Portfolio
	fund         *entity.Fund
	instrument   *entity.Instrument
	position     *entity.PortfolioPosition
	txns         *postTxnRepo
	cash         *postCashRepo
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
		cash:         &postCashRepo{balances: []*entity.CashBalance{{PortfolioID: portfolioID, Currency: "THB", Balance: decimal.NewFromInt(10_000)}}},
		businessDate: businessDate,
	}
}

func (f *postFixture) handler(compliance contract.ComplianceChecker) *PostTransactionHandler {
	return f.handlerWithWorkflow(compliance, &allowWorkflow{})
}

func (f *postFixture) handlerWithWorkflow(
	compliance contract.ComplianceChecker,
	workflow contract.WorkflowStateProvider,
) *PostTransactionHandler {
	h := NewPostTransactionHandler(
		nil,
		postPortfolioRepo{portfolio: f.portfolio},
		postFundRepo{fund: f.fund},
		postInstrumentRepo{instrument: f.instrument},
		postPositionRepo{position: f.position},
		f.cash,
		f.txns,
		nil,
		workflow,
		compliance,
		nilAudit{},
		func() time.Time { return f.businessDate.Add(9 * time.Hour) },
	)
	h.withTx = func(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
		return fn(nil)
	}
	return h
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
func (r postFundRepo) GetByContractCode(context.Context, string) (*entity.Fund, error) {
	return nil, nil
}
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

type postCashRepo struct {
	insertMovements int
	upserts         int
	balances        []*entity.CashBalance
}

func (r *postCashRepo) InsertMovement(context.Context, pgx.Tx, *entity.CashMovement) error {
	r.insertMovements++
	return nil
}
func (r *postCashRepo) GetBalanceForUpdate(context.Context, pgx.Tx, uuid.UUID, string) (*entity.CashBalance, error) {
	if len(r.balances) == 0 {
		return nil, nil
	}
	return r.balances[0], nil
}
func (r *postCashRepo) UpsertBalance(context.Context, pgx.Tx, *entity.CashBalance, int) error {
	r.upserts++
	return nil
}
func (r *postCashRepo) ListBalances(ctx context.Context, portfolioID uuid.UUID) ([]*entity.CashBalance, error) {
	return r.balances, nil
}

type allowWorkflow struct {
	lockCalls int
}

func (*allowWorkflow) IsTradeAllowed(context.Context, uuid.UUID, time.Time) (bool, error) {
	return true, nil
}
func (*allowWorkflow) IsTransactionLocked(context.Context, uuid.UUID, time.Time) (bool, error) {
	return false, nil
}
func (w *allowWorkflow) LockTradeDayForPost(context.Context, pgx.Tx, uuid.UUID, time.Time) (*contract.WorkflowDayLock, error) {
	w.lockCalls++
	return &contract.WorkflowDayLock{Exists: true, CurrentState: contract.WorkflowStateDayOpen}, nil
}

type closedWorkflow struct{}

func (closedWorkflow) IsTradeAllowed(context.Context, uuid.UUID, time.Time) (bool, error) {
	return false, nil
}
func (closedWorkflow) IsTransactionLocked(context.Context, uuid.UUID, time.Time) (bool, error) {
	return true, nil
}

type workflowLockProbe struct {
	readAllowed bool
	readLocked  bool
	lock        *contract.WorkflowDayLock
	lockCalls   int
}

func (w *workflowLockProbe) IsTradeAllowed(context.Context, uuid.UUID, time.Time) (bool, error) {
	return w.readAllowed, nil
}
func (w *workflowLockProbe) IsTransactionLocked(context.Context, uuid.UUID, time.Time) (bool, error) {
	return w.readLocked, nil
}
func (w *workflowLockProbe) LockTradeDayForPost(context.Context, pgx.Tx, uuid.UUID, time.Time) (*contract.WorkflowDayLock, error) {
	w.lockCalls++
	return w.lock, nil
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

func (c blockingCompliance) SimulateProposedOrder(ctx context.Context, req contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	return c.CheckProposedOrder(ctx, req)
}

type passCompliance struct{}

func (passCompliance) CheckProposedOrder(context.Context, contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	return &contract.ProposedOrderResult{CheckGroupID: uuid.New(), Verdict: contract.ComplianceVerdictPass}, nil
}

func (c passCompliance) SimulateProposedOrder(ctx context.Context, req contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	return c.CheckProposedOrder(ctx, req)
}

type postOnlyCompliance struct {
	calls int
}

func (c *postOnlyCompliance) CheckProposedOrder(context.Context, contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	c.calls++
	return &contract.ProposedOrderResult{CheckGroupID: uuid.New(), Verdict: contract.ComplianceVerdictPass}, nil
}

type capturingBlockingCompliance struct {
	req contract.ProposedOrderCheck
}

func (c *capturingBlockingCompliance) CheckProposedOrder(_ context.Context, req contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	c.req = req
	return &contract.ProposedOrderResult{
		CheckGroupID: uuid.New(),
		Verdict:      contract.ComplianceVerdictBlock,
		Breaches: []contract.ProposedOrderBreach{{
			RuleTypeID: "BLACKLIST", Verdict: contract.ComplianceVerdictBlock,
		}},
	}, nil
}

type nilAudit struct{}

func (nilAudit) LogAction(contract.AuditEntry) error { return nil }

func (nilAudit) LogActionStrict(context.Context, contract.AuditEntry) error { return nil }
