package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// PostPriceSnapshotRequest is the input for ingesting an instrument price.
type PostPriceSnapshotRequest struct {
	InstrumentID uuid.UUID
	BusinessDate time.Time
	Price        decimal.Decimal
	Currency     string
	PriceSource  string
	ProviderRef  string
	IsStale      bool
	StaleReason  string
	ActorID      uuid.UUID
}

// PostPriceSnapshotHandler appends a single price snapshot.
type PostPriceSnapshotHandler struct {
	pool        *pgxpool.Pool
	prices      domain.PriceSnapshotRepository
	instruments domain.InstrumentRepository
	audit       contract.AuditLogger
	now         func() time.Time
}

// NewPostPriceSnapshotHandler wires the handler.
func NewPostPriceSnapshotHandler(
	pool *pgxpool.Pool,
	prices domain.PriceSnapshotRepository,
	instruments domain.InstrumentRepository,
	audit contract.AuditLogger,
	now func() time.Time,
) *PostPriceSnapshotHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &PostPriceSnapshotHandler{pool: pool, prices: prices, instruments: instruments, audit: audit, now: now}
}

// Handle persists the snapshot.
func (h *PostPriceSnapshotHandler) Handle(ctx context.Context, req PostPriceSnapshotRequest) (*entity.PriceSnapshot, error) {
	if req.InstrumentID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "instrument_id", Detail: "is required"}
	}
	if req.BusinessDate.IsZero() {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "business_date", Detail: "is required"}
	}
	if req.Price.Sign() <= 0 {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "price", Detail: "must be positive"}
	}
	if req.Currency == "" {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "currency", Detail: "is required"}
	}
	if req.PriceSource == "" {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "price_source", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	if h.instruments == nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "instrument_id", Detail: "instrument repository unavailable"}
	}
	inst, err := h.instruments.GetByID(ctx, req.InstrumentID)
	if err != nil {
		return nil, fmt.Errorf("loading instrument: %w", err)
	}
	if inst == nil {
		return nil, &domain.ErrInstrumentNotFound{InstrumentID: req.InstrumentID.String()}
	}
	if req.Currency != inst.Currency {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "currency", Detail: "must match instrument currency"}
	}

	now := h.now()
	snap := &entity.PriceSnapshot{
		ID:           uuid.New(),
		InstrumentID: req.InstrumentID,
		BusinessDate: req.BusinessDate,
		Price:        req.Price,
		Currency:     req.Currency,
		PriceSource:  req.PriceSource,
		ProviderRef:  req.ProviderRef,
		IsStale:      req.IsStale,
		StaleReason:  req.StaleReason,
		CapturedAt:   now,
		CreatedAt:    now,
		CreatedBy:    req.ActorID,
	}

	if err := withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		return h.prices.Insert(ctx, dbtx, snap)
	}); err != nil {
		return nil, fmt.Errorf("inserting price snapshot: %w", err)
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_PRICE_POSTED",
		Module:       "investment",
		ResourceType: "INVESTMENT_PRICE",
		ResourceID:   snap.ID.String(),
		Details: map[string]any{
			"instrument_id": snap.InstrumentID,
			"price":         snap.Price.String(),
			"currency":      snap.Currency,
			"is_stale":      snap.IsStale,
		},
		BusinessDate: req.BusinessDate,
	})

	return snap, nil
}
