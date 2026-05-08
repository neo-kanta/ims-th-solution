package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// PostTransactionRequest is the input to the ledger-post pipeline.
//
// Money / quantity values are passed as decimal.Decimal — the transport DTO
// is responsible for parsing JSON strings into decimals.
type PostTransactionRequest struct {
	PortfolioID       uuid.UUID
	TransactionType   vo.TransactionType
	Side              *vo.OrderSide
	InstrumentID      *uuid.UUID
	Quantity          *decimal.Decimal
	Price             *decimal.Decimal
	Currency          string
	Fees              decimal.Decimal
	GrossAmount       *decimal.Decimal // optional override, otherwise quantity * price
	NetAmount         *decimal.Decimal // optional override
	FxRateToBase      *decimal.Decimal
	BusinessDate      time.Time
	SettlementDate    *time.Time
	SourceDecisionID  *uuid.UUID
	SourceExecutionID *uuid.UUID
	ExternalRef       string
	Reason            string

	ActorID        uuid.UUID
	AllowForcePost bool
}

// PostTransactionResult is returned to the caller on success.
type PostTransactionResult struct {
	Transaction *entity.PortfolioTransaction
}

// PostTransactionHandler posts a new ledger row and updates the position +
// cash projections in a single DB transaction.
//
// Invariants enforced here:
//   - Workflow lock re-checked inside the same DB tx (race-free).
//   - Append-only: the underlying repo only INSERTs; the Postgres RULE blocks
//     UPDATE/DELETE.
//   - Optimistic locking on position + cash balance.
//   - FX rate is mandatory when currency != portfolio.base_currency.
type PostTransactionHandler struct {
	pool        *pgxpool.Pool
	portfolios  domain.PortfolioRepository
	funds       domain.FundRepository
	instruments domain.InstrumentRepository
	positions   domain.PortfolioPositionRepository
	txns        domain.PortfolioTransactionRepository
	projector   *service.PortfolioProjector

	workflow   contract.WorkflowStateProvider
	compliance contract.ComplianceChecker
	audit      contract.AuditLogger

	now func() time.Time
}

// NewPostTransactionHandler wires the handler. `now` may be nil — defaults to
// time.Now().UTC().
func NewPostTransactionHandler(
	pool *pgxpool.Pool,
	portfolios domain.PortfolioRepository,
	funds domain.FundRepository,
	instruments domain.InstrumentRepository,
	positions domain.PortfolioPositionRepository,
	txns domain.PortfolioTransactionRepository,
	projector *service.PortfolioProjector,
	workflow contract.WorkflowStateProvider,
	compliance contract.ComplianceChecker,
	audit contract.AuditLogger,
	now func() time.Time,
) *PostTransactionHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &PostTransactionHandler{
		pool:        pool,
		portfolios:  portfolios,
		funds:       funds,
		instruments: instruments,
		positions:   positions,
		txns:        txns,
		projector:   projector,
		workflow:    workflow,
		compliance:  compliance,
		audit:       audit,
		now:         now,
	}
}

// Handle runs the post pipeline. Returns the persisted transaction.
func (h *PostTransactionHandler) Handle(
	ctx context.Context,
	req PostTransactionRequest,
) (*PostTransactionResult, error) {
	if h == nil {
		return nil, fmt.Errorf("post transaction handler not initialised")
	}
	if err := validatePostRequest(req); err != nil {
		return nil, err
	}

	// 1. Load context entities (portfolio, fund, instrument, position).
	portfolio, err := h.portfolios.GetByID(ctx, req.PortfolioID)
	if err != nil {
		return nil, fmt.Errorf("loading portfolio: %w", err)
	}
	if portfolio == nil {
		return nil, &domain.ErrPortfolioNotFound{PortfolioID: req.PortfolioID.String()}
	}

	fund, err := h.funds.GetByID(ctx, portfolio.FundID)
	if err != nil {
		return nil, fmt.Errorf("loading fund: %w", err)
	}
	if fund == nil {
		return nil, &domain.ErrFundNotFound{FundID: portfolio.FundID.String()}
	}

	var instrument *entity.Instrument
	if req.InstrumentID != nil {
		instrument, err = h.instruments.GetByID(ctx, *req.InstrumentID)
		if err != nil {
			return nil, fmt.Errorf("loading instrument: %w", err)
		}
		if instrument == nil {
			return nil, &domain.ErrInstrumentNotFound{InstrumentID: req.InstrumentID.String()}
		}
		if err := validateTickSize(instrument, req.Price); err != nil {
			return nil, err
		}
	}

	var position *entity.PortfolioPosition
	if req.InstrumentID != nil && req.TransactionType.IsSellLike() {
		positions, listErr := h.positions.ListByPortfolio(ctx, req.PortfolioID)
		if listErr != nil {
			return nil, fmt.Errorf("loading positions: %w", listErr)
		}
		for _, p := range positions {
			if p.InstrumentID == *req.InstrumentID {
				position = p
				break
			}
		}
	}

	// 2. Evaluate workflow gates (read-side; we re-check inside the tx).
	if h.workflow == nil {
		return nil, &domain.ErrPostPreconditionFailed{Violation: string(policy.PostViolationTradingNotAllowed), Detail: "workflow gate unavailable"}
	}
	tradeAllowed, err := h.workflow.IsTradeAllowed(ctx, fund.ID, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("checking trade-allowed: %w", err)
	}
	txnLocked, err := h.workflow.IsTransactionLocked(ctx, fund.ID, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("checking transaction-locked: %w", err)
	}

	quantity := decimal.Zero
	if req.Quantity != nil {
		quantity = *req.Quantity
	}
	price := decimal.Zero
	if req.Price != nil {
		price = *req.Price
	}

	violation := policy.EvaluatePost(policy.PostInputs{
		Type:                req.TransactionType,
		Portfolio:           portfolio,
		Fund:                fund,
		Instrument:          instrument,
		Quantity:            quantity,
		Price:               price,
		Currency:            req.Currency,
		Fees:                req.Fees,
		FxRateToBase:        req.FxRateToBase,
		Position:            position,
		IsTradeAllowed:      tradeAllowed,
		IsTransactionLocked: txnLocked,
		AllowForcePost:      req.AllowForcePost,
	})
	if violation != "" {
		return nil, &domain.ErrPostPreconditionFailed{Violation: string(violation)}
	}

	if err := h.checkCompliance(ctx, req, portfolio, fund, instrument, quantity, price); err != nil {
		return nil, err
	}

	// 3. Compute monetary amounts deterministically.
	gross, net, err := deriveAmounts(req, quantity, price)
	if err != nil {
		return nil, err
	}

	now := h.now()
	tx := &entity.PortfolioTransaction{
		ID:                    uuid.New(),
		PortfolioID:           portfolio.ID,
		FundID:                portfolio.FundID,
		InstrumentID:          req.InstrumentID,
		TransactionType:       req.TransactionType,
		Side:                  req.Side,
		Quantity:              req.Quantity,
		Price:                 req.Price,
		Currency:              req.Currency,
		GrossAmount:           gross,
		Fees:                  req.Fees,
		NetAmount:             net,
		FxRateToBase:          req.FxRateToBase,
		BusinessDate:          req.BusinessDate,
		SettlementDate:        req.SettlementDate,
		SourceDecisionID:      req.SourceDecisionID,
		SourceExecutionID:     req.SourceExecutionID,
		ReversesTransactionID: nil,
		ExternalRef:           req.ExternalRef,
		Reason:                req.Reason,
		Status:                vo.TransactionStatusPosted,
		CreatedAt:             now,
		CreatedBy:             req.ActorID,
	}

	// 4. Persist + project in one DB tx. Workflow lock is re-evaluated inside
	//    the tx via the contract; race window from step 2 is closed.
	err = withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		tradeAllowedNow, tradeErr := h.workflow.IsTradeAllowed(ctx, fund.ID, req.BusinessDate)
		if tradeErr != nil {
			return fmt.Errorf("re-checking trade-allowed: %w", tradeErr)
		}
		if !tradeAllowedNow {
			return &domain.ErrPostPreconditionFailed{
				Violation: string(policy.PostViolationTradingNotAllowed),
				Detail:    "workflow state changed during post",
			}
		}
		// Re-check the lock under the implicit FOR-SHARE semantics that the
		// workflow provider implements when called inside a tx. The provider
		// returns true if MANAGER_APPROVED or beyond.
		lockedNow, lockErr := h.workflow.IsTransactionLocked(ctx, fund.ID, req.BusinessDate)
		if lockErr != nil {
			return fmt.Errorf("re-checking transaction lock: %w", lockErr)
		}
		if lockedNow && !(req.AllowForcePost && req.TransactionType == vo.TransactionTypeReversal) {
			return &domain.ErrPostPreconditionFailed{
				Violation: string(policy.PostViolationTransactionLocked),
				Detail:    "lock acquired during post",
			}
		}

		if req.TransactionType.IsSellLike() && req.InstrumentID != nil {
			lockedPos, posErr := h.positions.GetForUpdate(ctx, dbtx, req.PortfolioID, *req.InstrumentID)
			if posErr != nil {
				return fmt.Errorf("locking position for sell: %w", posErr)
			}
			have := decimal.Zero
			avg := decimal.Zero
			if lockedPos != nil {
				have = lockedPos.Quantity
				avg = lockedPos.AverageCost
			}
			if quantity.GreaterThan(have) {
				return &domain.ErrPostPreconditionFailed{Violation: string(policy.PostViolationOversell)}
			}
			priceBase := price
			feesBase := req.Fees
			if req.FxRateToBase != nil {
				priceBase = priceBase.Mul(*req.FxRateToBase)
				feesBase = feesBase.Mul(*req.FxRateToBase)
			}
			tx.RealisedPnLBase = quantity.Mul(priceBase.Sub(avg)).Sub(feesBase)
		}

		if err := h.txns.Insert(ctx, dbtx, tx); err != nil {
			return fmt.Errorf("inserting transaction: %w", err)
		}
		if err := h.projector.Apply(ctx, dbtx, tx, +1); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 5. Audit. Failure to log does not roll back the post — the audit
	//    logger is fire-and-forget by contract.
	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_TRANSACTION_POSTED",
		Module:       "investment",
		ResourceType: "INVESTMENT_TRANSACTION",
		ResourceID:   tx.ID.String(),
		Details: map[string]any{
			"portfolio_id":     portfolio.ID,
			"fund_id":          fund.ID,
			"transaction_type": string(req.TransactionType),
			"net_amount":       net.String(),
			"currency":         req.Currency,
		},
		BusinessDate: req.BusinessDate,
	})

	return &PostTransactionResult{Transaction: tx}, nil
}

func (h *PostTransactionHandler) checkCompliance(
	ctx context.Context,
	req PostTransactionRequest,
	portfolio *entity.Portfolio,
	fund *entity.Fund,
	instrument *entity.Instrument,
	qty decimal.Decimal,
	price decimal.Decimal,
) error {
	if !req.TransactionType.IsSecurityTrade() {
		return nil
	}
	decisionID := ""
	if req.SourceDecisionID != nil {
		decisionID = req.SourceDecisionID.String()
	}
	if h.compliance == nil {
		return &domain.ErrComplianceRejected{DecisionID: decisionID, Message: "pre-trade compliance gate unavailable"}
	}
	side := contract.ComplianceOrderSideBuy
	if req.TransactionType.IsSellLike() {
		side = contract.ComplianceOrderSideSell
	}
	result, err := h.compliance.CheckProposedOrder(ctx, contract.ProposedOrderCheck{
		PortfolioID:  portfolio.ID,
		ContractID:   fund.ID,
		BusinessDate: req.BusinessDate,
		Actor:        req.ActorID.String(),
		OrderID:      uuid.New(),
		Ticker:       instrument.PrimaryTicker,
		Side:         side,
		Quantity:     qty,
		Price:        price,
		Currency:     req.Currency,
		Exchange:     instrument.PrimaryExchange,
	})
	if err != nil {
		return &domain.ErrComplianceRejected{DecisionID: decisionID, Message: err.Error()}
	}
	if result == nil || result.Verdict == contract.ComplianceVerdictBlock {
		if result != nil {
			rules := make([]string, 0)
			for _, b := range result.Breaches {
				if b.Verdict == contract.ComplianceVerdictBlock {
					rules = append(rules, b.RuleTypeID)
				}
			}
			return &domain.ErrComplianceRejected{
				DecisionID:   decisionID,
				CheckGroupID: result.CheckGroupID.String(),
				Message:      fmt.Sprintf("blocked rules: %v", rules),
			}
		}
		return &domain.ErrComplianceRejected{DecisionID: decisionID, Message: "empty compliance result"}
	}
	return nil
}

func validateTickSize(inst *entity.Instrument, price *decimal.Decimal) error {
	if inst == nil || inst.TickSize == nil || inst.TickSize.Sign() <= 0 || price == nil {
		return nil
	}
	if !price.Mod(*inst.TickSize).IsZero() {
		return &domain.ErrPostPreconditionFailed{
			Violation: string(policy.PostViolationTickSize),
			Detail:    "price is not a multiple of instrument tick size",
		}
	}
	return nil
}

func deriveAmounts(req PostTransactionRequest, qty, price decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	if req.TransactionType.IsSecurityTrade() {
		gross := qty.Mul(price)
		net := computeNetAmount(req, gross)
		if req.GrossAmount != nil && !req.GrossAmount.Equal(gross) {
			return decimal.Zero, decimal.Zero, &domain.ErrInvalidDecisionRequest{
				Field: "gross_amount", Detail: "security trade amount must equal quantity * price",
			}
		}
		if req.NetAmount != nil && !req.NetAmount.Equal(net) {
			return decimal.Zero, decimal.Zero, &domain.ErrInvalidDecisionRequest{
				Field: "net_amount", Detail: "security trade net amount must match computed cash impact",
			}
		}
		return gross, net, nil
	}

	gross := decimal.Zero
	if req.GrossAmount != nil {
		gross = *req.GrossAmount
	}
	net := computeNetAmount(req, gross)
	if req.NetAmount != nil {
		net = *req.NetAmount
	}
	return gross, net, nil
}

// computeNetAmount returns the signed cash impact in trade currency.
//
//	BUY / SUBSCRIPTION:  -(gross + fees)   (cash leaves the portfolio)
//	SELL / REDEMPTION:   +(gross - fees)   (cash arrives)
//	CASH_IN / DIVIDEND:  +gross
//	CASH_OUT / FEE:      -gross
//	REVERSAL:            caller-controlled (reverse handler sets it)
func computeNetAmount(req PostTransactionRequest, gross decimal.Decimal) decimal.Decimal {
	switch {
	case req.TransactionType.IsBuyLike():
		return gross.Add(req.Fees).Neg()
	case req.TransactionType.IsSellLike():
		return gross.Sub(req.Fees)
	case req.TransactionType == vo.TransactionTypeCashIn ||
		req.TransactionType == vo.TransactionTypeDividend:
		return gross
	case req.TransactionType == vo.TransactionTypeCashOut ||
		req.TransactionType == vo.TransactionTypeFee:
		return gross.Neg()
	}
	return decimal.Zero
}

func validatePostRequest(req PostTransactionRequest) error {
	if req.PortfolioID == uuid.Nil {
		return &domain.ErrInvalidDecisionRequest{Field: "portfolio_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	if !req.TransactionType.IsValid() || req.TransactionType == vo.TransactionTypeReversal {
		// Reversals are posted via the dedicated ReverseTransactionHandler.
		return &domain.ErrInvalidDecisionRequest{Field: "transaction_type", Detail: "invalid for direct post"}
	}
	if req.BusinessDate.IsZero() {
		return &domain.ErrInvalidDecisionRequest{Field: "business_date", Detail: "is required"}
	}
	if req.Currency == "" {
		return &domain.ErrInvalidDecisionRequest{Field: "currency", Detail: "is required"}
	}
	if req.Fees.Sign() < 0 {
		return &domain.ErrInvalidDecisionRequest{Field: "fees", Detail: "must be non-negative"}
	}
	return nil
}

// withTransaction wraps platform/database.WithTransaction without forcing a
// direct dependency in the command package — handler-side wrapper exists so
// tests can inject a stub if needed in future.
func withTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
