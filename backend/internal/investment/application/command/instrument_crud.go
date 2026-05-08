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
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// CreateInstrumentRequest registers a new tradable security.
type CreateInstrumentRequest struct {
	PrimaryTicker   string
	Name            string
	AssetClassID    uuid.UUID
	AssetSubtypeID  uuid.UUID
	Currency        string
	CountryID       uuid.UUID
	RegionID        *uuid.UUID
	PrimaryExchange string
	SectorID        *uuid.UUID
	FundCategoryID  *uuid.UUID
	LotSize         int
	TickSize        *decimal.Decimal
	Attributes      map[string]any
	ActorID         uuid.UUID
}

// UpdateInstrumentRequest patches mutable instrument metadata.
type UpdateInstrumentRequest struct {
	InstrumentID    uuid.UUID
	Name            *string
	PrimaryExchange *string
	SectorID        *uuid.UUID
	FundCategoryID  *uuid.UUID
	LotSize         *int
	TickSize        *decimal.Decimal
	IsTradable      *bool
	Status          *vo.InstrumentStatus
	Attributes      map[string]any
	ActorID         uuid.UUID
}

// InstrumentCommandHandler bundles instrument CRUD operations.
type InstrumentCommandHandler struct {
	pool        *pgxpool.Pool
	instruments domain.InstrumentRepository
	audit       contract.AuditLogger
	now         func() time.Time
}

// NewInstrumentCommandHandler wires the handler.
func NewInstrumentCommandHandler(
	pool *pgxpool.Pool,
	instruments domain.InstrumentRepository,
	audit contract.AuditLogger,
	now func() time.Time,
) *InstrumentCommandHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &InstrumentCommandHandler{pool: pool, instruments: instruments, audit: audit, now: now}
}

// Create persists a new Instrument.
func (h *InstrumentCommandHandler) Create(ctx context.Context, req CreateInstrumentRequest) (*entity.Instrument, error) {
	if req.PrimaryTicker == "" || req.Name == "" || req.Currency == "" {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "ticker/name/currency", Detail: "are required"}
	}
	if req.AssetClassID == uuid.Nil || req.AssetSubtypeID == uuid.Nil || req.CountryID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "classification", Detail: "asset_class/subtype/country are required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	if req.LotSize <= 0 {
		req.LotSize = 1
	}

	now := h.now()
	actor := req.ActorID
	attrs := req.Attributes
	if attrs == nil {
		attrs = map[string]any{}
	}

	inst := &entity.Instrument{
		ID:              uuid.New(),
		PrimaryTicker:   req.PrimaryTicker,
		Name:            req.Name,
		AssetClassID:    req.AssetClassID,
		AssetSubtypeID:  req.AssetSubtypeID,
		Currency:        req.Currency,
		CountryID:       req.CountryID,
		RegionID:        req.RegionID,
		PrimaryExchange: req.PrimaryExchange,
		SectorID:        req.SectorID,
		FundCategoryID:  req.FundCategoryID,
		LotSize:         req.LotSize,
		TickSize:        req.TickSize,
		IsTradable:      true,
		Status:          vo.InstrumentStatusActive,
		Attributes:      attrs,
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       &actor,
		UpdatedBy:       &actor,
	}

	if err := withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		return h.instruments.Create(ctx, dbtx, inst)
	}); err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_INSTRUMENT_CREATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_INSTRUMENT",
		ResourceID:   inst.ID.String(),
		Details:      map[string]any{"ticker": inst.PrimaryTicker},
		BusinessDate: now,
	})

	return inst, nil
}

// Update applies metadata changes to an existing instrument.
func (h *InstrumentCommandHandler) Update(ctx context.Context, req UpdateInstrumentRequest) (*entity.Instrument, error) {
	if req.InstrumentID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "instrument_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}

	inst, err := h.instruments.GetByID(ctx, req.InstrumentID)
	if err != nil {
		return nil, fmt.Errorf("loading instrument: %w", err)
	}
	if inst == nil {
		return nil, &domain.ErrInstrumentNotFound{InstrumentID: req.InstrumentID.String()}
	}

	if req.Name != nil {
		inst.Name = *req.Name
	}
	if req.PrimaryExchange != nil {
		inst.PrimaryExchange = *req.PrimaryExchange
	}
	if req.SectorID != nil {
		inst.SectorID = req.SectorID
	}
	if req.FundCategoryID != nil {
		inst.FundCategoryID = req.FundCategoryID
	}
	if req.LotSize != nil {
		inst.LotSize = *req.LotSize
	}
	if req.TickSize != nil {
		inst.TickSize = req.TickSize
	}
	if req.IsTradable != nil {
		inst.IsTradable = *req.IsTradable
	}
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, &domain.ErrInvalidDecisionRequest{Field: "status", Detail: "invalid"}
		}
		inst.Status = *req.Status
	}
	if req.Attributes != nil {
		inst.Attributes = req.Attributes
	}

	now := h.now()
	actor := req.ActorID
	inst.UpdatedAt = now
	inst.UpdatedBy = &actor

	if err := withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		return h.instruments.Update(ctx, dbtx, inst)
	}); err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_INSTRUMENT_UPDATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_INSTRUMENT",
		ResourceID:   inst.ID.String(),
		BusinessDate: now,
	})

	return inst, nil
}
