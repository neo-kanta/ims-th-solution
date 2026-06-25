package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// RecordConfirmationRequest creates a brand-new PENDING_REVIEW trade
// confirmation row tied to an existing execution.
type RecordConfirmationRequest struct {
	ExecutionID       uuid.UUID
	ConfirmedQuantity *decimal.Decimal
	ConfirmedAmount   *decimal.Decimal
	ConfirmedPrice    *decimal.Decimal
	BrokerReference   string
	ImportBatchID     *uuid.UUID
	ActorID           uuid.UUID
}

// ResolveConfirmationRequest flips a PENDING/MISMATCHED row to its terminal
// state. Target = MATCHED needs no reason; MISMATCHED/REVIEWED require one.
type ResolveConfirmationRequest struct {
	ConfirmationID    uuid.UUID
	TargetStatus      vo.TradeConfirmationStatus
	DiscrepancyReason string
	ActorID           uuid.UUID
}

type TradeConfirmationCommandHandler struct {
	pool          *pgxpool.Pool
	executions    domain.ExecutionRepository
	confirmations domain.TradeConfirmationRepository
	audit         contract.AuditLogger
	now           func() time.Time
	runTx         func(ctx context.Context, fn func(pgx.Tx) error) error
}

func NewTradeConfirmationCommandHandler(
	pool *pgxpool.Pool,
	executions domain.ExecutionRepository,
	confirmations domain.TradeConfirmationRepository,
	audit contract.AuditLogger,
	now func() time.Time,
) *TradeConfirmationCommandHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	h := &TradeConfirmationCommandHandler{
		pool:          pool,
		executions:    executions,
		confirmations: confirmations,
		audit:         audit,
		now:           now,
	}
	h.runTx = func(ctx context.Context, fn func(pgx.Tx) error) error {
		return withTransaction(ctx, pool, fn)
	}
	return h
}

func (h *TradeConfirmationCommandHandler) Record(ctx context.Context, req RecordConfirmationRequest) (*entity.TradeConfirmation, error) {
	if req.ExecutionID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "execution_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	exec, err := h.executions.GetByID(ctx, req.ExecutionID)
	if err != nil {
		return nil, fmt.Errorf("loading execution: %w", err)
	}
	if exec == nil {
		return nil, &domain.ErrExecutionNotFound{ExecutionID: req.ExecutionID.String()}
	}
	if exec.Status == vo.ExecutionStatusCancelled {
		return nil, &domain.ErrExecutionLifecycle{
			ExecutionID:   exec.ID.String(),
			CurrentStatus: string(exec.Status),
			Detail:        "cannot confirm a cancelled execution",
		}
	}

	now := h.now()
	c := &entity.TradeConfirmation{
		ID:                uuid.New(),
		ExecutionID:       exec.ID,
		DecisionID:        exec.DecisionID,
		FundID:            exec.FundID,
		PortfolioID:       exec.PortfolioID,
		ContractID:        exec.ContractID,
		BusinessDate:      exec.BusinessDate,
		ConfirmedQuantity: req.ConfirmedQuantity,
		ConfirmedAmount:   req.ConfirmedAmount,
		ConfirmedPrice:    req.ConfirmedPrice,
		Currency:          exec.Currency,
		BrokerReference:   strings.TrimSpace(req.BrokerReference),
		ImportBatchID:     req.ImportBatchID,
		Status:            vo.TradeConfirmationPendingReview,
		CreatedAt:         now,
		CreatedBy:         req.ActorID,
		UpdatedAt:         now,
		UpdatedBy:         req.ActorID,
	}
	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.confirmations.Create(ctx, tx, c)
	}); err != nil {
		return nil, err
	}
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_CONFIRMATION_CREATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_TRADE_CONFIRMATION",
		ResourceID:   c.ID.String(),
		Details: map[string]any{
			"execution_id":    exec.ID.String(),
			"decision_id":     exec.DecisionID.String(),
			"confirmed_qty":   c.ConfirmedQuantity,
			"confirmed_price": c.ConfirmedPrice,
		},
		BusinessDate: now,
	}); err != nil {
		return nil, fmt.Errorf("auditing confirmation create: %w", err)
	}
	return c, nil
}

func (h *TradeConfirmationCommandHandler) Resolve(ctx context.Context, req ResolveConfirmationRequest) (*entity.TradeConfirmation, error) {
	if req.ConfirmationID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "confirmation_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	if !req.TargetStatus.IsValid() ||
		req.TargetStatus == vo.TradeConfirmationPendingReview {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "status", Detail: "invalid target status"}
	}
	c, err := h.confirmations.GetByID(ctx, req.ConfirmationID)
	if err != nil {
		return nil, fmt.Errorf("loading confirmation: %w", err)
	}
	if c == nil {
		return nil, &domain.ErrConfirmationNotFound{ConfirmationID: req.ConfirmationID.String()}
	}
	if c.Status == vo.TradeConfirmationMatched || c.Status == vo.TradeConfirmationReviewed {
		return nil, &domain.ErrConfirmationLifecycle{
			ConfirmationID: c.ID.String(),
			CurrentStatus:  string(c.Status),
			Detail:         "confirmation already resolved",
		}
	}
	reason := strings.TrimSpace(req.DiscrepancyReason)
	if (req.TargetStatus == vo.TradeConfirmationMismatched ||
		req.TargetStatus == vo.TradeConfirmationReviewed) && reason == "" {
		return nil, &domain.ErrConfirmationMismatchReasonRequired{ConfirmationID: c.ID.String()}
	}
	now := h.now()
	c.Status = req.TargetStatus
	if reason != "" {
		c.DiscrepancyReason = reason
	}
	c.ReviewedAt = &now
	c.ReviewedBy = &req.ActorID
	c.UpdatedAt = now
	c.UpdatedBy = req.ActorID
	if err := h.runTx(ctx, func(tx pgx.Tx) error {
		return h.confirmations.Update(ctx, tx, c)
	}); err != nil {
		return nil, err
	}
	action := "INVESTMENT_CONFIRMATION_MATCHED"
	switch c.Status {
	case vo.TradeConfirmationMismatched:
		action = "INVESTMENT_CONFIRMATION_MISMATCHED"
	case vo.TradeConfirmationReviewed:
		action = "INVESTMENT_CONFIRMATION_REVIEWED"
	}
	if err := h.audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       action,
		Module:       "investment",
		ResourceType: "INVESTMENT_TRADE_CONFIRMATION",
		ResourceID:   c.ID.String(),
		Details:      map[string]any{"reason": reason},
		BusinessDate: now,
	}); err != nil {
		return nil, fmt.Errorf("auditing confirmation resolve: %w", err)
	}
	return c, nil
}
