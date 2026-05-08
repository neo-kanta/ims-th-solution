package command

import (
	"context"
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

// CreatePortfolioRequest creates a new Portfolio under an existing Fund.
type CreatePortfolioRequest struct {
	FundID            uuid.UUID
	Code              string
	Name              string
	Description       string
	BaseCurrency      string
	ValuationCurrency string
	StrategyCode      string
	StyleID           *uuid.UUID
	ManagerUserID     *uuid.UUID
	Benchmark         string
	RiskProfile       vo.RiskProfile
	InceptionDate     time.Time
	HasUnits          bool
	TaxLotMethod      vo.TaxLotMethod
	ActorID           uuid.UUID
}

// UpdatePortfolioRequest applies metadata changes.
type UpdatePortfolioRequest struct {
	PortfolioID     uuid.UUID
	ExpectedVersion int
	Name            *string
	Description     *string
	StrategyCode    *string
	StyleID         *uuid.UUID
	ManagerUserID   *uuid.UUID
	Benchmark       *string
	RiskProfile     *vo.RiskProfile
	Status          *vo.PortfolioStatus
	ActorID         uuid.UUID
}

// PortfolioCommandHandler bundles the small portfolio CRUD operations.
type PortfolioCommandHandler struct {
	pool       *pgxpool.Pool
	portfolios domain.PortfolioRepository
	funds      domain.FundRepository
	audit      contract.AuditLogger
	now        func() time.Time
}

// NewPortfolioCommandHandler wires the handler.
func NewPortfolioCommandHandler(
	pool *pgxpool.Pool,
	portfolios domain.PortfolioRepository,
	funds domain.FundRepository,
	audit contract.AuditLogger,
	now func() time.Time,
) *PortfolioCommandHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &PortfolioCommandHandler{pool: pool, portfolios: portfolios, funds: funds, audit: audit, now: now}
}

// Create persists a new Portfolio. (FundID, Code) must be unique among alive rows.
func (h *PortfolioCommandHandler) Create(ctx context.Context, req CreatePortfolioRequest) (*entity.Portfolio, error) {
	if req.FundID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "fund_id", Detail: "is required"}
	}
	if req.Code == "" || req.Name == "" || req.BaseCurrency == "" || req.ValuationCurrency == "" {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "code/name/currencies", Detail: "are required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}

	fund, err := h.funds.GetByID(ctx, req.FundID)
	if err != nil {
		return nil, fmt.Errorf("loading fund: %w", err)
	}
	if fund == nil {
		return nil, &domain.ErrFundNotFound{FundID: req.FundID.String()}
	}
	if !fund.IsActive() {
		return nil, &domain.ErrPostPreconditionFailed{Violation: "FUND_INACTIVE"}
	}

	existing, err := h.portfolios.GetByFundCode(ctx, req.FundID, req.Code)
	if err != nil {
		return nil, fmt.Errorf("checking existing portfolio code: %w", err)
	}
	if existing != nil {
		return nil, &domain.ErrCodeAlreadyExists{Resource: "portfolio", Code: req.Code}
	}

	taxMethod := req.TaxLotMethod
	if taxMethod == "" {
		taxMethod = vo.TaxLotMethodAverage
	}
	if !taxMethod.IsValid() {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "tax_lot_method", Detail: "invalid"}
	}

	now := h.now()
	actor := req.ActorID
	p := &entity.Portfolio{
		ID:                uuid.New(),
		FundID:            req.FundID,
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		BaseCurrency:      req.BaseCurrency,
		ValuationCurrency: req.ValuationCurrency,
		StrategyCode:      req.StrategyCode,
		StyleID:           req.StyleID,
		ManagerUserID:     req.ManagerUserID,
		Benchmark:         req.Benchmark,
		RiskProfile:       req.RiskProfile,
		InceptionDate:     req.InceptionDate,
		Status:            vo.PortfolioStatusActive,
		HasUnits:          req.HasUnits,
		TaxLotMethod:      taxMethod,
		Version:           1,
		CreatedAt:         now,
		UpdatedAt:         now,
		CreatedBy:         &actor,
		UpdatedBy:         &actor,
	}

	if err := withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		return h.portfolios.Create(ctx, dbtx, p)
	}); err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_PORTFOLIO_CREATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_PORTFOLIO",
		ResourceID:   p.ID.String(),
		Details:      map[string]any{"fund_id": p.FundID, "code": p.Code, "name": p.Name},
		BusinessDate: now,
	})

	return p, nil
}

// Update applies metadata changes to an existing Portfolio.
func (h *PortfolioCommandHandler) Update(ctx context.Context, req UpdatePortfolioRequest) (*entity.Portfolio, error) {
	if req.PortfolioID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "portfolio_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}

	p, err := h.portfolios.GetByID(ctx, req.PortfolioID)
	if err != nil {
		return nil, fmt.Errorf("loading portfolio: %w", err)
	}
	if p == nil {
		return nil, &domain.ErrPortfolioNotFound{PortfolioID: req.PortfolioID.String()}
	}
	if p.Version != req.ExpectedVersion {
		return nil, &domain.ErrPortfolioVersionMismatch{
			PortfolioID: p.ID.String(), ExpectedVersion: req.ExpectedVersion, ActualVersion: p.Version,
		}
	}

	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.StrategyCode != nil {
		p.StrategyCode = *req.StrategyCode
	}
	if req.StyleID != nil {
		p.StyleID = req.StyleID
	}
	if req.ManagerUserID != nil {
		p.ManagerUserID = req.ManagerUserID
	}
	if req.Benchmark != nil {
		p.Benchmark = *req.Benchmark
	}
	if req.RiskProfile != nil {
		p.RiskProfile = *req.RiskProfile
	}
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, &domain.ErrInvalidDecisionRequest{Field: "status", Detail: "invalid"}
		}
		p.Status = *req.Status
	}

	now := h.now()
	actor := req.ActorID
	p.UpdatedAt = now
	p.UpdatedBy = &actor
	p.Version = req.ExpectedVersion + 1

	if err := withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		return h.portfolios.Update(ctx, dbtx, p)
	}); err != nil {
		return nil, err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      req.ActorID.String(),
		Action:       "INVESTMENT_PORTFOLIO_UPDATED",
		Module:       "investment",
		ResourceType: "INVESTMENT_PORTFOLIO",
		ResourceID:   p.ID.String(),
		BusinessDate: now,
	})

	return p, nil
}

// SoftDelete marks a Portfolio as deleted, refusing if it has open activity.
func (h *PortfolioCommandHandler) SoftDelete(ctx context.Context, portfolioID uuid.UUID, expectedVersion int, actorID uuid.UUID) error {
	if portfolioID == uuid.Nil {
		return &domain.ErrInvalidDecisionRequest{Field: "portfolio_id", Detail: "is required"}
	}
	hasActivity, detail, err := h.portfolios.HasOpenActivity(ctx, portfolioID, h.now())
	if err != nil {
		return fmt.Errorf("checking open activity: %w", err)
	}
	if hasActivity {
		return &domain.ErrPortfolioHasOpenActivity{PortfolioID: portfolioID.String(), Detail: detail}
	}

	if err := withTransaction(ctx, h.pool, func(dbtx pgx.Tx) error {
		return h.portfolios.SoftDelete(ctx, dbtx, portfolioID, expectedVersion, actorID)
	}); err != nil {
		return err
	}

	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      actorID.String(),
		Action:       "INVESTMENT_PORTFOLIO_DELETED",
		Module:       "investment",
		ResourceType: "INVESTMENT_PORTFOLIO",
		ResourceID:   portfolioID.String(),
		BusinessDate: h.now(),
	})
	return nil
}
