package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// CreateFundRequest is the input for creating a Fund (contract).
type CreateFundRequest struct {
	Code           string
	Name           string
	ShortName      string
	FundCategoryID uuid.UUID
	BaseCurrency   string
	InceptionDate  time.Time
	ManagerUserID  *uuid.UUID
	Benchmark      string
	RiskProfile    vo.RiskProfile
	HasUnits       bool
	ExternalPAMRef string
	ActorID        uuid.UUID
}

// UpdateFundRequest is the input for updating mutable Fund metadata.
//
// Only non-nil pointer fields are applied. Optimistic concurrency uses
// ExpectedVersion (typically the value the client last read).
type UpdateFundRequest struct {
	FundID          uuid.UUID
	ExpectedVersion int
	Name            *string
	ShortName       *string
	FundCategoryID  *uuid.UUID
	ManagerUserID   *uuid.UUID
	Benchmark       *string
	RiskProfile     *vo.RiskProfile
	Status          *vo.FundStatus
	ExternalPAMRef  *string
	ActorID         uuid.UUID
}

// FundCommandHandler bundles the small fund CRUD operations.
type FundCommandHandler struct {
	pool  *pgxpool.Pool
	funds domain.FundRepository
	audit contract.AuditLogger
	now   func() time.Time
}

// NewFundCommandHandler wires the handler.
func NewFundCommandHandler(
	pool *pgxpool.Pool,
	funds domain.FundRepository,
	audit contract.AuditLogger,
	now func() time.Time,
) *FundCommandHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &FundCommandHandler{pool: pool, funds: funds, audit: audit, now: now}
}

// Create persists a new Fund. Code must be unique among non-deleted rows.
func (h *FundCommandHandler) Create(ctx context.Context, req CreateFundRequest) (*entity.Fund, error) {
	if req.Code == "" || req.Name == "" || req.BaseCurrency == "" {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "code/name/base_currency", Detail: "are required"}
	}
	if req.FundCategoryID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "fund_category_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}

	existing, err := h.funds.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("checking existing fund code: %w", err)
	}
	if existing != nil {
		return nil, &domain.ErrCodeAlreadyExists{Resource: "fund", Code: req.Code}
	}

	now := h.now()
	actor := req.ActorID
	f := &entity.Fund{
		ID:             uuid.New(),
		Code:           req.Code,
		Name:           req.Name,
		ShortName:      req.ShortName,
		FundCategoryID: req.FundCategoryID,
		BaseCurrency:   req.BaseCurrency,
		InceptionDate:  req.InceptionDate,
		ManagerUserID:  req.ManagerUserID,
		Benchmark:      req.Benchmark,
		RiskProfile:    req.RiskProfile,
		HasUnits:       req.HasUnits,
		ExternalPAMRef: req.ExternalPAMRef,
		Status:         vo.FundStatusActive,
		Version:        1,
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedBy:      &actor,
		UpdatedBy:      &actor,
	}

	if err := withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		return h.funds.Create(ctx, dbtx, f)
	}); err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_FUND_CREATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_FUND",
		ResourceID:   f.ID.String(),
		Details:      map[string]any{"code": f.Code, "name": f.Name},
		BusinessDate: now,
	})

	return f, nil
}

// Update applies the supplied changes to an existing Fund.
func (h *FundCommandHandler) Update(ctx context.Context, req UpdateFundRequest) (*entity.Fund, error) {
	if req.FundID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "fund_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}

	f, err := h.funds.GetByID(ctx, req.FundID)
	if err != nil {
		return nil, fmt.Errorf("loading fund: %w", err)
	}
	if f == nil {
		return nil, &domain.ErrFundNotFound{FundID: req.FundID.String()}
	}
	if f.Version != req.ExpectedVersion {
		return nil, &domain.ErrFundVersionMismatch{
			FundID: f.ID.String(), ExpectedVersion: req.ExpectedVersion, ActualVersion: f.Version,
		}
	}

	if req.Name != nil {
		f.Name = *req.Name
	}
	if req.ShortName != nil {
		f.ShortName = *req.ShortName
	}
	if req.FundCategoryID != nil {
		f.FundCategoryID = *req.FundCategoryID
	}
	if req.ManagerUserID != nil {
		f.ManagerUserID = req.ManagerUserID
	}
	if req.Benchmark != nil {
		f.Benchmark = *req.Benchmark
	}
	if req.RiskProfile != nil {
		f.RiskProfile = *req.RiskProfile
	}
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, &domain.ErrInvalidDecisionRequest{Field: "status", Detail: "invalid"}
		}
		f.Status = *req.Status
	}
	if req.ExternalPAMRef != nil {
		f.ExternalPAMRef = *req.ExternalPAMRef
	}

	now := h.now()
	actor := req.ActorID
	f.UpdatedAt = now
	f.UpdatedBy = &actor
	f.Version = req.ExpectedVersion + 1

	if err := withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		return h.funds.Update(ctx, dbtx, f)
	}); err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_FUND_UPDATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_FUND",
		ResourceID:   f.ID.String(),
		BusinessDate: now,
	})

	return f, nil
}

// SoftDelete marks a Fund as deleted, refusing if it has active portfolios.
func (h *FundCommandHandler) SoftDelete(ctx context.Context, fundID uuid.UUID, expectedVersion int, actorID uuid.UUID) error {
	if fundID == uuid.Nil {
		return &domain.ErrInvalidDecisionRequest{Field: "fund_id", Detail: "is required"}
	}
	count, err := h.funds.CountActivePortfolios(ctx, fundID)
	if err != nil {
		return fmt.Errorf("counting active portfolios: %w", err)
	}
	if count > 0 {
		return &domain.ErrFundHasActivePortfolios{FundID: fundID.String(), Count: count}
	}

	if err := withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		return h.funds.SoftDelete(ctx, dbtx, fundID, expectedVersion, actorID)
	}); err != nil {
		// Surface a typed version mismatch when applicable; otherwise wrap.
		var versionErr *domain.ErrFundVersionMismatch
		if errors.As(err, &versionErr) {
			return err
		}
		return fmt.Errorf("soft-deleting fund: %w", err)
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      actorID.String(),
		Action:       "INVESTMENT_FUND_DELETED",
		Module:       "investment",
		ResourceType: "INVESTMENT_FUND",
		ResourceID:   fundID.String(),
		BusinessDate: h.now(),
	})

	return nil
}
