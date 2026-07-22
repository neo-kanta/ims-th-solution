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

// CreatePortfolioRequest creates a new Portfolio under an existing Fund.
type CreatePortfolioRequest struct {
	FundID            uuid.UUID
	PortfolioType     vo.PortfolioType
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

// PortfolioCommandHandler bundles the portfolio CRUD and lifecycle operations.
type PortfolioCommandHandler struct {
	pool          *pgxpool.Pool
	portfolios    domain.PortfolioRepository
	statusHistory domain.PortfolioStatusHistoryRepository
	funds         domain.FundRepository
	audit         contract.AuditLogger
	now           func() time.Time
}

// NewPortfolioCommandHandler wires the handler.
// statusHistory may be nil — when nil, history writes are skipped (tests / early bring-up).
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

// SetStatusHistoryRepository wires the status-history repo post-construction.
// Called from module.go after the repo is created so the constructor signature
// stays backward-compatible.
func (h *PortfolioCommandHandler) SetStatusHistoryRepository(r domain.PortfolioStatusHistoryRepository) {
	if h != nil {
		h.statusHistory = r
	}
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

	// Belt-and-suspenders global uniqueness check. The database enforces
	// global (not just per-fund) active-code uniqueness via the
	// uq_inv_portfolios_code_alive unique index (migration
	// 20260703000001_investment__portfolio_v2_hardening), but the
	// GetByFundCode check above only catches a same-fund collision. Without
	// this check, a cross-fund code collision would reach the INSERT and
	// surface as a raw unique-constraint violation (undifferentiated 500)
	// instead of the typed ErrCodeAlreadyExists conflict every other
	// duplicate-code path in this handler already returns. This is
	// defense-in-depth — the DB constraint is authoritative and still
	// applies even if this check is ever bypassed or races a concurrent
	// insert.
	globalExisting, err := h.portfolios.GetByCode(ctx, req.Code)
	if err != nil {
		var ambiguous *domain.ErrAmbiguousPortfolioCode
		if errors.As(err, &ambiguous) {
			return nil, &domain.ErrCodeAlreadyExists{Resource: "portfolio", Code: req.Code}
		}
		return nil, fmt.Errorf("checking global portfolio code uniqueness: %w", err)
	}
	if globalExisting != nil {
		return nil, &domain.ErrCodeAlreadyExists{Resource: "portfolio", Code: req.Code}
	}

	taxMethod := req.TaxLotMethod
	if taxMethod == "" {
		taxMethod = vo.TaxLotMethodAverage
	}
	if !taxMethod.IsValid() {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "tax_lot_method", Detail: "invalid"}
	}

	portfolioType := req.PortfolioType
	if portfolioType == "" {
		portfolioType = vo.PortfolioTypeLive
	}
	if !portfolioType.IsValid() {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "portfolio_type", Detail: "invalid"}
	}

	now := h.now()
	actor := req.ActorID
	p := &entity.Portfolio{
		ID:                uuid.New(),
		FundID:            req.FundID,
		PortfolioType:     portfolioType,
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

// ApplyApprovalDecision is called by the approval module when a
// PORTFOLIO_ONBOARDING approval request reaches a final decision.
// approved=true → ACTIVE; approved=false → REJECTED.
// Writes both the portfolio status update and a history row in one transaction.
func (h *PortfolioCommandHandler) ApplyApprovalDecision(
	ctx context.Context,
	portfolioID uuid.UUID,
	approved bool,
	reason string,
) error {
	p, err := h.portfolios.GetByID(ctx, portfolioID)
	if err != nil {
		return fmt.Errorf("loading portfolio for approval decision: %w", err)
	}
	if p == nil {
		return &domain.ErrPortfolioNotFound{PortfolioID: portfolioID.String()}
	}
	if p.Status != vo.PortfolioStatusPendingApproval {
		return fmt.Errorf("portfolio %s is in status %s, expected PENDING_APPROVAL", portfolioID, p.Status)
	}

	now := h.now()
	fromStatus := p.Status
	if approved {
		p.Status = vo.PortfolioStatusActive
	} else {
		p.Status = vo.PortfolioStatusRejected
	}
	p.UpdatedAt = now

	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}

	if err := withTransaction(ctx, h.pool, func(tx pgx.Tx) error {
		if err := h.portfolios.Update(ctx, tx, p); err != nil {
			return err
		}
		if h.statusHistory != nil {
			histEntry := &entity.PortfolioStatusHistory{
				ID:          uuid.New(),
				PortfolioID: portfolioID,
				FromStatus:  &fromStatus,
				ToStatus:    p.Status,
				ActorID:     nil, // system-initiated via approval engine
				Reason:      reasonPtr,
				CreatedAt:   now,
			}
			if err := h.statusHistory.Append(ctx, tx, histEntry); err != nil {
				return fmt.Errorf("appending portfolio status history: %w", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	action := "INVESTMENT_PORTFOLIO_APPROVED"
	details := map[string]any{"fund_id": p.FundID, "code": p.Code}
	if !approved {
		action = "INVESTMENT_PORTFOLIO_REJECTED"
		details["reason"] = reason
	}
	_ = h.audit.LogAction(contract.AuditEntry{
		ActorID:      "approval-engine",
		Action:       action,
		Module:       "investment",
		ResourceType: "INVESTMENT_PORTFOLIO",
		ResourceID:   portfolioID.String(),
		Details:      details,
		BusinessDate: now,
	})
	return nil
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
