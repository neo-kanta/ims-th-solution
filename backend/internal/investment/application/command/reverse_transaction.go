package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ReverseTransactionRequest is the input to the reverse pipeline.
//
// Reasons are mandatory — every correction creates an auditable explanation.
type ReverseTransactionRequest struct {
	OriginalTransactionID uuid.UUID
	BusinessDate          time.Time
	Reason                string
	ActorID               uuid.UUID
	AllowForcePost        bool
}

// ReverseTransactionResult returns the new (REVERSAL) ledger row.
type ReverseTransactionResult struct {
	Reversal *entity.PortfolioTransaction
	Original *entity.PortfolioTransaction
}

// ReverseTransactionHandler posts a REVERSAL row for an existing transaction
// and applies the inverse projection in a single DB tx.
type ReverseTransactionHandler struct {
	pool      *pgxpool.Pool
	txns      domain.PortfolioTransactionRepository
	projector *service.PortfolioProjector
	workflow  contract.WorkflowStateProvider
	audit     contract.AuditLogger
	now       func() time.Time
}

// NewReverseTransactionHandler wires the handler.
func NewReverseTransactionHandler(
	pool *pgxpool.Pool,
	txns domain.PortfolioTransactionRepository,
	projector *service.PortfolioProjector,
	workflow contract.WorkflowStateProvider,
	audit contract.AuditLogger,
	now func() time.Time,
) *ReverseTransactionHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &ReverseTransactionHandler{
		pool:      pool,
		txns:      txns,
		projector: projector,
		workflow:  workflow,
		audit:     audit,
		now:       now,
	}
}

// Handle runs the reverse pipeline.
func (h *ReverseTransactionHandler) Handle(
	ctx context.Context,
	req ReverseTransactionRequest,
) (*ReverseTransactionResult, error) {
	if h == nil {
		return nil, fmt.Errorf("reverse transaction handler not initialised")
	}
	if req.OriginalTransactionID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "transaction_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	if req.Reason == "" {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "reason", Detail: "is required"}
	}
	if req.BusinessDate.IsZero() {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "business_date", Detail: "is required"}
	}

	original, err := h.txns.GetByID(ctx, req.OriginalTransactionID)
	if err != nil {
		return nil, fmt.Errorf("loading original transaction: %w", err)
	}
	if original == nil {
		return nil, &domain.ErrTransactionNotFound{TransactionID: req.OriginalTransactionID.String()}
	}
	if original.IsReversal() {
		return nil, &domain.ErrCannotReverseReversal{TransactionID: original.ID.String()}
	}
	already, err := h.txns.HasReversal(ctx, original.ID)
	if err != nil {
		return nil, fmt.Errorf("checking existing reversal: %w", err)
	}
	if already {
		return nil, &domain.ErrTransactionAlreadyReversed{TransactionID: original.ID.String()}
	}

	now := h.now()
	rev := &entity.PortfolioTransaction{
		ID:                    uuid.New(),
		PortfolioID:           original.PortfolioID,
		FundID:                original.FundID,
		InstrumentID:          original.InstrumentID,
		TransactionType:       vo.TransactionTypeReversal,
		Side:                  original.Side,
		Quantity:              original.Quantity,
		Price:                 original.Price,
		Currency:              original.Currency,
		GrossAmount:           original.GrossAmount,
		Fees:                  original.Fees,
		NetAmount:             original.NetAmount.Neg(),
		RealisedPnLBase:       original.RealisedPnLBase.Neg(),
		FxRateToBase:          original.FxRateToBase,
		BusinessDate:          req.BusinessDate,
		ReversesTransactionID: &original.ID,
		Reason:                req.Reason,
		Status:                vo.TransactionStatusPosted,
		CreatedAt:             now,
		CreatedBy:             req.ActorID,
	}

	err = withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		// Cross-check workflow lock at insert-time. FORCE_POST may bypass. No
		// fund-scoped lock to check for a fund-less transaction.
		if original.FundID != nil {
			locked, lockErr := h.workflow.IsTransactionLocked(ctx, *original.FundID, req.BusinessDate)
			if lockErr != nil {
				return fmt.Errorf("re-checking transaction lock: %w", lockErr)
			}
			if locked && !req.AllowForcePost {
				return &domain.ErrPostPreconditionFailed{
					Violation: "TRANSACTION_LOCKED",
					Detail:    "reversal requires force-post when day is locked",
				}
			}
		}

		if err := h.txns.Insert(ctx, dbtx, rev); err != nil {
			return fmt.Errorf("inserting reversal: %w", err)
		}
		if err := h.projector.ApplyReversal(ctx, dbtx, rev, original); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_TRANSACTION_REVERSED",
		Module:       "investment",
		ResourceType: "INVESTMENT_TRANSACTION",
		ResourceID:   rev.ID.String(),
		Details: map[string]any{
			"original_id":  original.ID,
			"portfolio_id": original.PortfolioID,
			"reason":       req.Reason,
		},
		BusinessDate: req.BusinessDate,
	})

	return &ReverseTransactionResult{Reversal: rev, Original: original}, nil
}
