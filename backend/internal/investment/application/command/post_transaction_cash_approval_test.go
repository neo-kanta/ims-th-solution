package command

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// cashGateHarness wires a PostTransactionHandler with a real projector and the
// Stage 2 cash-approval dependencies (cash-request repo, approval submitter,
// approval canceller) on top of the shared postFixture repos.
type cashGateHarness struct {
	fixture   *postFixture
	handler   *PostTransactionHandler
	cashRepo  *fakeCashRepo
	submitter *fakeCashApprovalSubmitter
	canceller *fakeCashApprovalCanceller
	status    *fakeApprovalStatus
}

func newCashGateHarness(t *testing.T) *cashGateHarness {
	t.Helper()
	return newCashGateHarnessWithWorkflow(t, &allowWorkflow{})
}

func newCashGateHarnessWithWorkflow(t *testing.T, workflow contract.WorkflowStateProvider) *cashGateHarness {
	t.Helper()
	f := newPostFixture()
	projector := service.NewPortfolioProjectorWithTransactions(
		postPositionRepo{position: f.position}, f.cash, f.txns,
	)
	h := NewPostTransactionHandler(
		nil,
		postPortfolioRepo{portfolio: f.portfolio},
		postFundRepo{fund: f.fund},
		postInstrumentRepo{instrument: f.instrument},
		postPositionRepo{position: f.position},
		f.cash,
		f.txns,
		projector,
		workflow,
		passCompliance{},
		nilAudit{},
		func() time.Time { return f.businessDate.Add(9 * time.Hour) },
	)
	h.withTx = func(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
		return fn(nil)
	}
	cashRepo := newFakeCashRepo()
	submitter := &fakeCashApprovalSubmitter{}
	canceller := &fakeCashApprovalCanceller{}
	status := &fakeApprovalStatus{}
	h.SetCashRequestRepository(cashRepo)
	h.SetApprovalSubmitter(submitter)
	h.SetApprovalCanceller(canceller)
	h.SetApprovalStatusProvider(status)
	return &cashGateHarness{fixture: f, handler: h, cashRepo: cashRepo, submitter: submitter, canceller: canceller, status: status}
}

func cashReq(f *postFixture, amount int64) PostTransactionRequest {
	gross := decimal.NewFromInt(amount)
	return PostTransactionRequest{
		PortfolioID:     f.portfolio.ID,
		TransactionType: vo.TransactionTypeCashIn,
		Currency:        "THB",
		GrossAmount:     &gross,
		BusinessDate:    f.businessDate,
		ActorID:         uuid.New(),
		Reason:          "test cash in",
	}
}

// ── §6: LIVE cash → PENDING request, no txn ──────────────────────────────────

func TestLiveCashCreatesPendingRequestNoTxn(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	res, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.NoError(t, err)
	require.True(t, res.Pending)
	require.NotNil(t, res.CashRequest)
	require.Nil(t, res.Transaction)
	require.Equal(t, vo.CashRequestStatusPending, res.CashRequest.Status)
	require.NotNil(t, res.CashRequest.ApprovalRequestID)
	require.Equal(t, 0, h.fixture.txns.inserts, "no ledger row must be posted for a pending LIVE cash movement")
	require.Equal(t, 1, h.cashRepo.creates)
	require.Equal(t, 1, h.submitter.calls)
	require.Equal(t, "PORTFOLIO_CASH_TRANSACTION", h.submitter.last.ProcessType)
	require.Equal(t, "CASH_TRANSACTION", h.submitter.last.SubjectType)
	require.Equal(t, res.CashRequest.ID, h.submitter.last.SubjectID)
	require.Equal(t, "FUND", h.submitter.last.ContractType)
	require.NotNil(t, h.submitter.last.ContractID)
}

// ── §6: SIMULATION cash → posts immediately (regression) ─────────────────────

func TestSimulationCashPostsImmediately(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeSimulation

	res, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.NoError(t, err)
	require.False(t, res.Pending)
	require.NotNil(t, res.Transaction)
	require.Equal(t, 1, h.fixture.txns.inserts)
	require.Equal(t, 0, h.cashRepo.creates)
	require.Equal(t, 0, h.submitter.calls)
}

// ── §6: MODEL cash → blocked ─────────────────────────────────────────────────

func TestModelCashBlocked(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeModel

	_, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	var precondition *domain.ErrPostPreconditionFailed
	require.ErrorAs(t, err, &precondition)
	require.Equal(t, string(policy.PostViolationModelLedgerBlocked), precondition.Violation)
	require.Equal(t, 0, h.fixture.txns.inserts)
	require.Equal(t, 0, h.cashRepo.creates)
	require.Equal(t, 0, h.submitter.calls)
}

// ── §6: LIVE cash on a closed/stale business day → blocked before the gate ────
//
// G2 item 8. The workflow-day check runs in the SHARED prepareTransaction
// step, before the cash-only branch even decides LIVE vs SIMULATION vs
// MODEL — so a closed business day must block a LIVE cash submission the
// same way it blocks a security trade, and must do so BEFORE any cash
// request row or approval submission is created. This is fund-BOUND only
// (the fixture portfolio has FundID set): a fund-less portfolio's workflow
// gate is a separate, already-tracked G6 gap (see
// docs/MANAGER/TASKS.md P0-F / this goal's G6 package;
// prepareTransaction's fund==nil branch treats the gate as always-open by
// design, not a bug this test covers).

func TestLiveCashOnClosedBusinessDay_BlockedBeforeApprovalSubmit(t *testing.T) {
	h := newCashGateHarnessWithWorkflow(t, closedWorkflow{})
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	_, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.Error(t, err)
	require.Equal(t, 0, h.cashRepo.creates, "no cash request row on a closed business day")
	require.Equal(t, 0, h.submitter.calls, "no approval submission on a closed business day")
}

// ── §6: LIVE BUY → posts immediately (out of gate scope, regression) ─────────

func TestLiveBuyPostsImmediately(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(5)

	res, err := h.handler.Handle(context.Background(), PostTransactionRequest{
		PortfolioID:     h.fixture.portfolio.ID,
		TransactionType: vo.TransactionTypeBuy,
		InstrumentID:    &h.fixture.instrument.ID,
		Quantity:        &qty,
		Price:           &price,
		Currency:        "THB",
		BusinessDate:    h.fixture.businessDate,
		ActorID:         uuid.New(),
	})
	require.NoError(t, err)
	require.False(t, res.Pending)
	require.NotNil(t, res.Transaction)
	require.Equal(t, 1, h.fixture.txns.inserts)
	require.Equal(t, 0, h.cashRepo.creates)
	require.Equal(t, 0, h.submitter.calls)
}

// ── §6: APPROVED callback → materializes exactly one txn ─────────────────────

func TestApprovedCallbackMaterializesTxn(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	res, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.NoError(t, err)
	require.Equal(t, 0, h.fixture.txns.inserts)

	err = h.handler.ApplyCashRequestApproval(context.Background(), res.CashRequest.ID, true, "ok")
	require.NoError(t, err)
	require.Equal(t, 1, h.fixture.txns.inserts, "exactly one ledger row materialized on approval")

	stored, _ := h.cashRepo.GetByID(context.Background(), res.CashRequest.ID)
	require.Equal(t, vo.CashRequestStatusApproved, stored.Status)
	require.NotNil(t, stored.ResultingTxnID)
	require.NotNil(t, stored.DecidedAt)
}

// ── §6: REJECTED callback → no txn ───────────────────────────────────────────

func TestRejectedCallbackNoTxn(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	res, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.NoError(t, err)

	err = h.handler.ApplyCashRequestApproval(context.Background(), res.CashRequest.ID, false, "denied")
	require.NoError(t, err)
	require.Equal(t, 0, h.fixture.txns.inserts)

	stored, _ := h.cashRepo.GetByID(context.Background(), res.CashRequest.ID)
	require.Equal(t, vo.CashRequestStatusRejected, stored.Status)
	require.Nil(t, stored.ResultingTxnID)
}

// ── §6: duplicate approval callback → idempotent, single txn ─────────────────

func TestDuplicateApprovalIsIdempotent(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	res, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.NoError(t, err)

	require.NoError(t, h.handler.ApplyCashRequestApproval(context.Background(), res.CashRequest.ID, true, ""))
	require.NoError(t, h.handler.ApplyCashRequestApproval(context.Background(), res.CashRequest.ID, true, ""))
	require.Equal(t, 1, h.fixture.txns.inserts, "duplicate approval callback must not double-post")

	stored, _ := h.cashRepo.GetByID(context.Background(), res.CashRequest.ID)
	require.Equal(t, vo.CashRequestStatusApproved, stored.Status)
}

// ── §6: cancel PENDING by submitter → CANCELLED, approval cancelled, no txn ───

func TestCancelPendingBySubmitter(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	req := cashReq(h.fixture, 1000)

	res, err := h.handler.Handle(context.Background(), req)
	require.NoError(t, err)

	cr, err := h.handler.CancelCashRequest(context.Background(), res.CashRequest.ID, req.ActorID, "changed my mind")
	require.NoError(t, err)
	require.Equal(t, vo.CashRequestStatusCancelled, cr.Status)
	require.Equal(t, 1, h.canceller.calls)
	require.Equal(t, "CASH_TRANSACTION", h.canceller.lastSubjectType)
	require.Equal(t, res.CashRequest.ID, h.canceller.lastSubjectID)
	require.Equal(t, 0, h.fixture.txns.inserts)
}

// ── §6: cancel by non-submitter → rejected ───────────────────────────────────

func TestCancelByNonSubmitterRejected(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	res, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.NoError(t, err)

	_, err = h.handler.CancelCashRequest(context.Background(), res.CashRequest.ID, uuid.New(), "")
	var forbidden *domain.ErrCashRequestForbidden
	require.ErrorAs(t, err, &forbidden)
	require.Equal(t, 0, h.canceller.calls)

	stored, _ := h.cashRepo.GetByID(context.Background(), res.CashRequest.ID)
	require.Equal(t, vo.CashRequestStatusPending, stored.Status)
}

// ── §6: cancel non-PENDING → rejected ────────────────────────────────────────

func TestCancelNonPendingRejected(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	req := cashReq(h.fixture, 1000)

	res, err := h.handler.Handle(context.Background(), req)
	require.NoError(t, err)
	require.NoError(t, h.handler.ApplyCashRequestApproval(context.Background(), res.CashRequest.ID, true, ""))

	_, err = h.handler.CancelCashRequest(context.Background(), res.CashRequest.ID, req.ActorID, "")
	var notCancel *domain.ErrCashRequestNotCancellable
	require.ErrorAs(t, err, &notCancel)
}

// ── §6: materialization re-check BLOCK → no txn, all-or-nothing rollback ──────

func TestMaterializationRecheckBlockRollsBack(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	res, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.NoError(t, err)

	// State changes between submit and approval: the portfolio is deactivated,
	// so the authoritative re-check at materialization must block and post
	// nothing.
	h.fixture.portfolio.Status = vo.PortfolioStatusClosed

	err = h.handler.ApplyCashRequestApproval(context.Background(), res.CashRequest.ID, true, "")
	var precondition *domain.ErrPostPreconditionFailed
	require.ErrorAs(t, err, &precondition)
	require.Equal(t, 0, h.fixture.txns.inserts, "no ledger row on a blocked re-check")

	stored, _ := h.cashRepo.GetByID(context.Background(), res.CashRequest.ID)
	require.Equal(t, vo.CashRequestStatusPending, stored.Status, "request stays PENDING, not APPROVED, when re-check blocks")
	require.Nil(t, stored.ResultingTxnID)
}

// ── §6: fund-less LIVE portfolio → whole flow works (scope = portfolio id) ────

func TestFundlessLiveCashFlow(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	h.fixture.portfolio.FundID = nil // fund-less

	res, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.NoError(t, err)
	require.True(t, res.Pending)
	require.Nil(t, res.CashRequest.FundID)
	require.Equal(t, "COMPANY", h.submitter.last.ContractType)
	require.Nil(t, h.submitter.last.ContractID)

	err = h.handler.ApplyCashRequestApproval(context.Background(), res.CashRequest.ID, true, "")
	require.NoError(t, err)
	require.Equal(t, 1, h.fixture.txns.inserts)

	stored, _ := h.cashRepo.GetByID(context.Background(), res.CashRequest.ID)
	require.Equal(t, vo.CashRequestStatusApproved, stored.Status)
	require.NotNil(t, stored.ResultingTxnID)
}

// ── §7 fail-closed: LIVE cash with no approval gate wired → rejected, no txn ──

func TestLiveCashFailsClosedWithoutApprovalGate(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	h.handler.SetApprovalSubmitter(nil) // simulate unwired approval engine

	_, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	var precondition *domain.ErrPostPreconditionFailed
	require.ErrorAs(t, err, &precondition)
	require.Equal(t, string(policy.PostViolationApprovalUnavailable), precondition.Violation)
	require.Equal(t, 0, h.fixture.txns.inserts)
	require.Equal(t, 0, h.cashRepo.creates)
}

// ── G2 item 2: request-level idempotency ─────────────────────────────────────

// keyedCashReq returns a LIVE cash-in request carrying an idempotency key with a
// fresh actor. To model the SAME client retrying, reuse the returned value (or
// copy its ActorID); two separate keyedCashReq calls model DIFFERENT actors.
func keyedCashReq(f *postFixture, amount int64, key string) PostTransactionRequest {
	r := cashReq(f, amount)
	r.ActorID = uuid.New()
	r.IdempotencyKey = key
	return r
}

// A response/update timeout followed by a retry with the SAME key must not
// create a second cash request or a second approval submission.
func TestIdempotentRetrySameKeyReturnsSameRequest(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	req := keyedCashReq(h.fixture, 1000, "idem-key-abc")

	first, err := h.handler.Handle(context.Background(), req)
	require.NoError(t, err)
	require.True(t, first.Pending)
	require.NotNil(t, first.CashRequest.ApprovalRequestID)

	second, err := h.handler.Handle(context.Background(), req)
	require.NoError(t, err)
	require.True(t, second.Pending)

	require.Equal(t, first.CashRequest.ID, second.CashRequest.ID, "retry returns the SAME request id")
	require.Equal(t, 1, h.cashRepo.creates, "no second cash request created")
	require.Equal(t, 1, h.submitter.calls, "no second approval submission")
	require.Equal(t, 0, h.fixture.txns.inserts)
}

// A concurrent retry that loses the insert race (partial unique index) resolves
// to the winner instead of surfacing a 500.
func TestConcurrentDuplicateKeyResolvesToWinner(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	req := keyedCashReq(h.fixture, 1000, "race-key")

	// First submit wins and links.
	first, err := h.handler.Handle(context.Background(), req)
	require.NoError(t, err)

	// Second submit: force the idempotency lookup to MISS once (row not yet
	// visible), so it proceeds to Create and trips the unique index, then the
	// post-violation fetch resolves to the winner.
	h.cashRepo.missKeyLookups = 1
	second, err := h.handler.Handle(context.Background(), req)
	require.NoError(t, err)

	require.Equal(t, first.CashRequest.ID, second.CashRequest.ID)
	require.Equal(t, 1, h.cashRepo.creates, "the losing insert did not add a row")
	require.Equal(t, 1, h.submitter.calls, "no second approval submission")
}

// G2 Finding 3: even WITHOUT a client key, a genuine retry of the byte-identical
// request dedups via the server-derived key ("auto:" || fingerprint). NOTE: the
// fixed ActorID in cashReq models the SAME submitter retrying — the fingerprint
// (which includes submitter_id) is therefore identical across both calls.
func TestNoKeyIdenticalRetryDedups(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	actor := uuid.New()

	req := cashReq(h.fixture, 1000)
	req.ActorID = actor
	r1, err := h.handler.Handle(context.Background(), req)
	require.NoError(t, err)

	// Same payload, same actor, still no client key → same fingerprint → dedup.
	req2 := cashReq(h.fixture, 1000)
	req2.ActorID = actor
	req2.Reason = req.Reason
	r2, err := h.handler.Handle(context.Background(), req2)
	require.NoError(t, err)

	require.Equal(t, r1.CashRequest.ID, r2.CashRequest.ID, "identical no-key retry returns the same request")
	require.Equal(t, 1, h.cashRepo.creates, "no second cash request")
	require.Equal(t, 1, h.submitter.calls, "no second approval submission")
}

// A no-key submission with a DIFFERENT payload (different amount → different
// fingerprint → different derived key) is a genuinely new request, not a dedup.
func TestNoKeyDifferentPayloadCreatesTwo(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	actor := uuid.New()

	r1req := cashReq(h.fixture, 1000)
	r1req.ActorID = actor
	r1, err := h.handler.Handle(context.Background(), r1req)
	require.NoError(t, err)

	r2req := cashReq(h.fixture, 2000) // different amount
	r2req.ActorID = actor
	r2, err := h.handler.Handle(context.Background(), r2req)
	require.NoError(t, err)

	require.NotEqual(t, r1.CashRequest.ID, r2.CashRequest.ID)
	require.Equal(t, 2, h.cashRepo.creates)
	require.Equal(t, 2, h.submitter.calls)
}

// Two DIFFERENT actors submitting the same no-key payload are two distinct
// requests: submitter_id is part of the fingerprint, so their derived keys
// differ and neither resumes/leaks the other's request.
func TestNoKeyDifferentActorsCreateTwo(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	r1req := cashReq(h.fixture, 1000)
	r1req.ActorID = uuid.New()
	r1, err := h.handler.Handle(context.Background(), r1req)
	require.NoError(t, err)

	r2req := cashReq(h.fixture, 1000)
	r2req.ActorID = uuid.New()
	r2req.Reason = r1req.Reason
	r2, err := h.handler.Handle(context.Background(), r2req)
	require.NoError(t, err)

	require.NotEqual(t, r1.CashRequest.ID, r2.CashRequest.ID)
	require.Equal(t, 2, h.cashRepo.creates)
}

func TestOverlongIdempotencyKeyRejected(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	req := keyedCashReq(h.fixture, 1000, strings.Repeat("x", maxIdempotencyKeyLen+1))

	_, err := h.handler.Handle(context.Background(), req)
	var invalid *domain.ErrInvalidDecisionRequest
	require.ErrorAs(t, err, &invalid)
	require.Equal(t, "idempotency_key", invalid.Field)
	require.Equal(t, 0, h.cashRepo.creates)
	require.Equal(t, 0, h.submitter.calls)
}

// Reusing the same key for a DIFFERENT payload is rejected, not silently
// resolved to the original movement (money-path safety).
func TestSameKeyDifferentPayloadRejected(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	actor := uuid.New()
	first := keyedCashReq(h.fixture, 1000, "reused-key")
	first.ActorID = actor
	_, err := h.handler.Handle(context.Background(), first)
	require.NoError(t, err)

	// Same key, same actor, different amount → fingerprint mismatch.
	second := keyedCashReq(h.fixture, 9999, "reused-key")
	second.ActorID = actor
	_, err = h.handler.Handle(context.Background(), second)
	var invalid *domain.ErrInvalidDecisionRequest
	require.ErrorAs(t, err, &invalid)
	require.Equal(t, "idempotency_key", invalid.Field)
	require.Contains(t, invalid.Detail, "different cash movement")

	require.Equal(t, 1, h.cashRepo.creates, "mismatched retry created no new request")
	require.Equal(t, 1, h.submitter.calls)
}

// Reusing the same client key by a DIFFERENT actor is rejected — never resume or
// disclose another user's request.
func TestSameKeyDifferentActorRejected(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	first := keyedCashReq(h.fixture, 1000, "shared-key")
	_, err := h.handler.Handle(context.Background(), first)
	require.NoError(t, err)

	// Different actor, same (portfolio, key).
	second := keyedCashReq(h.fixture, 1000, "shared-key") // keyedCashReq assigns a fresh ActorID
	_, err = h.handler.Handle(context.Background(), second)
	var invalid *domain.ErrInvalidDecisionRequest
	require.ErrorAs(t, err, &invalid)
	require.Equal(t, "idempotency_key", invalid.Field)
	require.Contains(t, invalid.Detail, "another user")

	require.Equal(t, 1, h.cashRepo.creates, "no new request for the second actor")
	require.Equal(t, 1, h.submitter.calls)
}

// A client key starting with the reserved "auto:" prefix is rejected so a client
// cannot forge a collision with the server-derived no-key path.
func TestReservedPrefixKeyRejected(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	req := keyedCashReq(h.fixture, 1000, "auto:deadbeef")

	_, err := h.handler.Handle(context.Background(), req)
	var invalid *domain.ErrInvalidDecisionRequest
	require.ErrorAs(t, err, &invalid)
	require.Equal(t, "idempotency_key", invalid.Field)
	require.Equal(t, 0, h.cashRepo.creates)
	require.Equal(t, 0, h.submitter.calls)
}

// A retry (same key) AFTER the request reached a terminal state (here APPROVED)
// returns the terminal request and produces NO new approval and NO new financial
// effect — the key is spent.
func TestTerminalRequestRetryNoNewEffect(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	req := keyedCashReq(h.fixture, 1000, "spent-key")

	first, err := h.handler.Handle(context.Background(), req)
	require.NoError(t, err)
	require.NoError(t, h.handler.ApplyCashRequestApproval(context.Background(), first.CashRequest.ID, true, ""))
	require.Equal(t, 1, h.fixture.txns.inserts)

	// Retry the same key after APPROVED (terminal).
	second, err := h.handler.Handle(context.Background(), req)
	require.NoError(t, err)
	require.True(t, second.Pending)
	require.Equal(t, first.CashRequest.ID, second.CashRequest.ID)
	require.Equal(t, vo.CashRequestStatusApproved, second.CashRequest.Status)
	require.Equal(t, 1, h.submitter.calls, "no second approval submission")
	require.Equal(t, 1, h.fixture.txns.inserts, "no second ledger row")
}

// Timeout-after-commit: a FRESH submit returns an error even though the approval
// actually committed. The handler must verify (via the status provider), find the
// committed approval, and RE-LINK — never compensate/cancel a committed approval.
func TestAmbiguousSubmitErrorButApprovalCommittedRelinks(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive

	// The submit "fails" (client-observed timeout) but the engine committed the
	// approval, which the status provider reports.
	h.submitter.err = errors.New("context deadline exceeded")
	h.status.stage = &contract.ApprovalStageInfo{RequestID: uuid.New(), Status: "PENDING_APPROVAL"}

	res, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.NoError(t, err, "committed approval re-linked, not treated as a failure")
	require.NotNil(t, res.CashRequest.ApprovalRequestID)
	require.Equal(t, 0, h.canceller.calls, "must NOT cancel a committed approval")

	list, _, _ := h.cashRepo.ListByPortfolio(context.Background(), h.fixture.portfolio.ID, domain.CashRequestListFilter{})
	require.Len(t, list, 1)
	require.Equal(t, vo.CashRequestStatusPending, list[0].Status, "row stays PENDING (linked), not cancelled")
}

// computeCashRequestFingerprint must be scale-invariant for money values so an
// honest retry sending "1000" vs "1000.00" hashes identically.
func TestFingerprintDeterministicAcrossScale(t *testing.T) {
	actor := uuid.New()
	portfolio := uuid.New()
	vd := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)

	a := computeCashRequestFingerprint(actor, portfolio, vo.TransactionTypeCashIn,
		decimal.RequireFromString("1000"), decimal.RequireFromString("0"), "THB", vd, "memo")
	b := computeCashRequestFingerprint(actor, portfolio, vo.TransactionTypeCashIn,
		decimal.RequireFromString("1000.00"), decimal.RequireFromString("0.000"), "thb", vd.Add(6*time.Hour), " memo ")
	require.Equal(t, a, b, "scale/case/time-of-day/whitespace must not change the fingerprint")

	// A different amount changes the fingerprint.
	c := computeCashRequestFingerprint(actor, portfolio, vo.TransactionTypeCashIn,
		decimal.RequireFromString("1000.01"), decimal.RequireFromString("0"), "THB", vd, "memo")
	require.NotEqual(t, a, c)
}

// ── G2 item 3: robust create/submit correlation ──────────────────────────────

// Step-3 (link) failure on a KEYED request leaves the row PENDING+NULL and is
// reconciled on retry by re-linking to the already-submitted approval — never a
// second approval request.
func TestKeyedLinkFailureThenRetryRelinks(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	req := keyedCashReq(h.fixture, 1000, "link-fail-key")

	// First attempt: submit succeeds, but recording the approval id (update #1)
	// fails — the row stays PENDING with a NULL approval_request_id.
	h.cashRepo.failUpdateN = 1
	h.cashRepo.updateErr = errors.New("boom: update timeout")
	_, err := h.handler.Handle(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "retry to reconcile")
	require.Equal(t, 1, h.submitter.calls)

	stored, _ := h.cashRepo.GetByID(context.Background(), reqIDForKey(t, h.cashRepo, "link-fail-key"))
	require.Equal(t, vo.CashRequestStatusPending, stored.Status)
	require.Nil(t, stored.ApprovalRequestID, "approval id not recorded after link failure")

	// The approval status provider reports the approval the first attempt already
	// submitted. On retry the reconcile path re-links instead of re-submitting.
	h.status.stage = &contract.ApprovalStageInfo{RequestID: uuid.New(), Status: "PENDING_APPROVAL"}
	res, err := h.handler.Handle(context.Background(), req)
	require.NoError(t, err)
	require.True(t, res.Pending)
	require.NotNil(t, res.CashRequest.ApprovalRequestID)

	require.Equal(t, 1, h.submitter.calls, "retry must NOT create a second approval request")
	require.Equal(t, 1, h.status.calls, "retry looked up the existing approval")
	require.Equal(t, 1, h.cashRepo.creates)
}

// A keyed PENDING+NULL row cannot be reconciled without the approval status
// provider — the reconcile path fails closed rather than blindly re-submitting.
func TestReconcileWithoutStatusProviderFailsClosed(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	h.handler.SetApprovalStatusProvider(nil)
	req := keyedCashReq(h.fixture, 1000, "no-status-key")

	h.cashRepo.failUpdateN = 1
	h.cashRepo.updateErr = errors.New("boom")
	_, err := h.handler.Handle(context.Background(), req)
	require.Error(t, err)
	require.Equal(t, 1, h.submitter.calls)

	// Retry: no status provider → fail closed, no second submit.
	_, err = h.handler.Handle(context.Background(), req)
	var precondition *domain.ErrPostPreconditionFailed
	require.ErrorAs(t, err, &precondition)
	require.Equal(t, 1, h.submitter.calls, "must not re-submit when it cannot reconcile")
}

// Submit (step-2) failure compensates by cancelling the staged row; when the
// compensating cancel ALSO fails, BOTH errors are surfaced (the `_ =` fix).
func TestSubmitFailureCompensationSurfacesBothErrors(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	h.submitter.err = errors.New("approval engine down")
	// The compensating cancel is the first Update — make it fail too.
	h.cashRepo.failUpdateN = 1
	h.cashRepo.updateErr = errors.New("cancel db error")

	_, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.Error(t, err)
	require.Contains(t, err.Error(), "approval engine down", "root submit failure surfaced")
	require.Contains(t, err.Error(), "cancel db error", "compensation failure surfaced, not swallowed")
	require.Equal(t, 0, h.fixture.txns.inserts)
	require.Equal(t, 1, h.cashRepo.creates)
}

// Submit (step-2) failure with a successful compensating cancel leaves the row
// CANCELLED (never a silent orphan PENDING).
func TestSubmitFailureCancelsRow(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	h.submitter.err = errors.New("approval engine down")

	res, err := h.handler.Handle(context.Background(), cashReq(h.fixture, 1000))
	require.Error(t, err)
	require.Nil(t, res)

	// Exactly one row exists and it is CANCELLED, not a lingering PENDING orphan.
	list, _, _ := h.cashRepo.ListByPortfolio(context.Background(), h.fixture.portfolio.ID, domain.CashRequestListFilter{})
	require.Len(t, list, 1)
	require.Equal(t, vo.CashRequestStatusCancelled, list[0].Status)
}

// G2 Finding 3+4: a NO-KEY step-3 (link) failure is now retry-reconcilable
// because the request carries a server-derived key. The row is left PENDING+NULL
// (never cancelled — that would strand the committed approval) and an identical
// retry re-links to the same approval instead of creating a second.
func TestNoKeyLinkFailureLeavesReconcilablePending(t *testing.T) {
	h := newCashGateHarness(t)
	h.fixture.portfolio.PortfolioType = vo.PortfolioTypeLive
	actor := uuid.New()

	// First attempt: submit succeeds, link update fails → PENDING+NULL.
	h.cashRepo.failUpdateN = 1
	h.cashRepo.updateErr = errors.New("link update timeout")
	req := cashReq(h.fixture, 1000)
	req.ActorID = actor
	res, err := h.handler.Handle(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, res)
	require.Contains(t, err.Error(), "retry to reconcile")
	require.Equal(t, 0, h.canceller.calls, "no compensation cancel — a committed approval must not be stranded")

	list, _, _ := h.cashRepo.ListByPortfolio(context.Background(), h.fixture.portfolio.ID, domain.CashRequestListFilter{})
	require.Len(t, list, 1)
	require.Equal(t, vo.CashRequestStatusPending, list[0].Status, "row stays PENDING for reconciliation, not cancelled")
	require.Nil(t, list[0].ApprovalRequestID)

	// Retry the identical no-key request: the status provider reports the
	// approval the first attempt already submitted → re-link, no second submit.
	h.cashRepo.failUpdateN = 0
	h.status.stage = &contract.ApprovalStageInfo{RequestID: uuid.New(), Status: "PENDING_APPROVAL"}
	req2 := cashReq(h.fixture, 1000)
	req2.ActorID = actor
	req2.Reason = req.Reason
	res2, err := h.handler.Handle(context.Background(), req2)
	require.NoError(t, err)
	require.NotNil(t, res2.CashRequest.ApprovalRequestID)
	require.Equal(t, 1, h.submitter.calls, "retry must NOT create a second approval request")
	require.Equal(t, 1, h.cashRepo.creates, "retry reused the same row")
}

// reqIDForKey finds the stored request id for an idempotency key.
func reqIDForKey(t *testing.T, r *fakeCashRepo, key string) uuid.UUID {
	t.Helper()
	for _, c := range r.store {
		if c.IdempotencyKey != nil && *c.IdempotencyKey == key {
			return c.ID
		}
	}
	t.Fatalf("no stored request for key %q", key)
	return uuid.Nil
}

// ── fakes ────────────────────────────────────────────────────────────────────

type fakeCashRepo struct {
	store   map[uuid.UUID]*entity.PortfolioCashRequest
	creates int
	updates int

	// Test knobs.
	getByKeyErr error // GetByIdempotencyKey returns this when set.
	// missKeyLookups makes the next N GetByIdempotencyKey calls return (nil,nil)
	// even when a matching row exists, simulating the read-then-insert race that
	// trips the partial unique index.
	missKeyLookups int
	// failUpdateN makes the r.updates == failUpdateN'th Update call return
	// updateErr, simulating a step-3 link or compensation failure precisely.
	failUpdateN int
	updateErr   error
}

// pgUniqueErr implements the SQLState() interface that command.isUniqueViolation
// inspects, so the fake can reproduce a Postgres partial-unique-index violation.
type pgUniqueErr struct{}

func (pgUniqueErr) Error() string    { return "duplicate key value violates unique constraint" }
func (pgUniqueErr) SQLState() string { return "23505" }

func newFakeCashRepo() *fakeCashRepo {
	return &fakeCashRepo{store: map[uuid.UUID]*entity.PortfolioCashRequest{}}
}

func (r *fakeCashRepo) Create(_ context.Context, _ pgx.Tx, c *entity.PortfolioCashRequest) error {
	// Enforce the partial unique index on (portfolio_id, idempotency_key) so
	// dedup/race behaviour is realistic. NULL keys never collide.
	if c.IdempotencyKey != nil {
		for _, e := range r.store {
			if e.PortfolioID == c.PortfolioID && e.IdempotencyKey != nil && *e.IdempotencyKey == *c.IdempotencyKey {
				return pgUniqueErr{}
			}
		}
	}
	r.creates++
	cp := *c
	r.store[c.ID] = &cp
	return nil
}

func (r *fakeCashRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.PortfolioCashRequest, error) {
	c, ok := r.store[id]
	if !ok {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

func (r *fakeCashRepo) GetByIdempotencyKey(_ context.Context, portfolioID uuid.UUID, key string) (*entity.PortfolioCashRequest, error) {
	if r.getByKeyErr != nil {
		return nil, r.getByKeyErr
	}
	if r.missKeyLookups > 0 {
		r.missKeyLookups--
		return nil, nil
	}
	for _, c := range r.store {
		if c.PortfolioID == portfolioID && c.IdempotencyKey != nil && *c.IdempotencyKey == key {
			cp := *c
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *fakeCashRepo) GetByFingerprint(_ context.Context, portfolioID uuid.UUID, fingerprint string) (*entity.PortfolioCashRequest, error) {
	var best *entity.PortfolioCashRequest
	for _, c := range r.store {
		if c.PortfolioID != portfolioID || c.RequestFingerprint == nil || *c.RequestFingerprint != fingerprint {
			continue
		}
		if best == nil || c.SubmittedAt.Before(best.SubmittedAt) {
			best = c
		}
	}
	if best == nil {
		return nil, nil
	}
	cp := *best
	return &cp, nil
}

func (r *fakeCashRepo) GetForUpdate(ctx context.Context, _ pgx.Tx, id uuid.UUID) (*entity.PortfolioCashRequest, error) {
	return r.GetByID(ctx, id)
}

func (r *fakeCashRepo) Update(_ context.Context, _ pgx.Tx, c *entity.PortfolioCashRequest) error {
	if _, ok := r.store[c.ID]; !ok {
		return errors.New("no row to update")
	}
	r.updates++
	if r.failUpdateN > 0 && r.updates == r.failUpdateN {
		return r.updateErr
	}
	c.Version++
	cp := *c
	r.store[c.ID] = &cp
	return nil
}

func (r *fakeCashRepo) ListByPortfolio(_ context.Context, portfolioID uuid.UUID, f domain.CashRequestListFilter) ([]*entity.PortfolioCashRequest, int, error) {
	out := []*entity.PortfolioCashRequest{}
	for _, c := range r.store {
		if c.PortfolioID != portfolioID {
			continue
		}
		if f.Status != nil && c.Status != *f.Status {
			continue
		}
		cp := *c
		out = append(out, &cp)
	}
	return out, len(out), nil
}

var _ domain.PortfolioCashRequestRepository = (*fakeCashRepo)(nil)

type fakeCashApprovalSubmitter struct {
	calls int
	err   error
	last  contract.ApprovalSubmission
}

func (s *fakeCashApprovalSubmitter) SubmitForApproval(_ context.Context, in contract.ApprovalSubmission) (*contract.ApprovalSubmissionResult, error) {
	s.calls++
	s.last = in
	if s.err != nil {
		return nil, s.err
	}
	return &contract.ApprovalSubmissionResult{RequestID: uuid.New(), RequestNumber: "CR-1", Status: "PENDING_APPROVAL"}, nil
}

type fakeCashApprovalCanceller struct {
	calls           int
	err             error
	lastSubjectType string
	lastSubjectID   uuid.UUID
}

func (c *fakeCashApprovalCanceller) CancelApprovalBySubject(_ context.Context, subjectType string, subjectID, _ uuid.UUID) error {
	c.calls++
	c.lastSubjectType = subjectType
	c.lastSubjectID = subjectID
	return c.err
}

type fakeApprovalStatus struct {
	calls int
	stage *contract.ApprovalStageInfo
	err   error
	last  uuid.UUID
}

func (s *fakeApprovalStatus) GetApprovalStage(_ context.Context, _ string, subjectID uuid.UUID) (*contract.ApprovalStageInfo, error) {
	s.calls++
	s.last = subjectID
	return s.stage, s.err
}
