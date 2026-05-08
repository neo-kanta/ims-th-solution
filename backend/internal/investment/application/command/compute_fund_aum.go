package command

import (
	"context"
	"errors"
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

// ComputeFundAUMRequest is the input for the fund-level AUM aggregation.
type ComputeFundAUMRequest struct {
	FundID       uuid.UUID
	BusinessDate time.Time
	ActorID      uuid.UUID
}

// ComputeFundAUMResult bundles the persisted snapshot plus the per-portfolio
// breakdown the caller may surface to the operator.
type ComputeFundAUMResult struct {
	Snapshot       *entity.AUMSnapshot
	PortfolioCount int
	PortfolioAUMs  []PortfolioAUMBreakdown
	Idempotent     bool
}

// PortfolioAUMBreakdown is one row in the fund AUM aggregation.
type PortfolioAUMBreakdown struct {
	PortfolioID  uuid.UUID
	AUM          decimal.Decimal
	ValuationCcy string
	PriceSetHash string
}

// ComputeFundAUMHandler aggregates today's portfolio valuation snapshots into
// a single fund-level AUM row.
//
// Invariants enforced:
//   - Fund must exist and be alive.
//   - Every active portfolio under the fund must have a valuation_snapshot at
//     (portfolio_id, business_date, source=INTERNAL). Missing → ErrIncompleteFundValuation
//     with reason MISSING_PORTFOLIO_SNAPSHOT.
//   - All portfolio snapshots must share the same price_set_hash. Mismatch
//     → ErrIncompleteFundValuation with reason PRICE_SET_HASH_MISMATCH.
//   - All portfolio snapshots must agree on valuation_ccy = fund.base_currency.
//     Mismatch → ErrIncompleteFundValuation with reason VALUATION_CCY_MISMATCH.
//   - Idempotent on (scope_type=FUND, scope_id=fund_id, business_date,
//     source=INTERNAL): re-running returns the existing snapshot with
//     Idempotent=true rather than inserting a duplicate.
type ComputeFundAUMHandler struct {
	pool       *pgxpool.Pool
	funds      domain.FundRepository
	portfolios domain.PortfolioRepository
	valuation  domain.ValuationRepository
	audit      contract.AuditLogger
	now        func() time.Time
}

// NewComputeFundAUMHandler wires the handler. now may be nil; defaults to UTC.
func NewComputeFundAUMHandler(
	pool *pgxpool.Pool,
	funds domain.FundRepository,
	portfolios domain.PortfolioRepository,
	valuation domain.ValuationRepository,
	audit contract.AuditLogger,
	now func() time.Time,
) *ComputeFundAUMHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &ComputeFundAUMHandler{
		pool:       pool,
		funds:      funds,
		portfolios: portfolios,
		valuation:  valuation,
		audit:      audit,
		now:        now,
	}
}

// Handle aggregates portfolio AUM into a fund-level AUM snapshot.
func (h *ComputeFundAUMHandler) Handle(
	ctx context.Context,
	req ComputeFundAUMRequest,
) (*ComputeFundAUMResult, error) {
	if h == nil {
		return nil, fmt.Errorf("compute fund aum handler not initialised")
	}
	if req.FundID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "fund_id", Detail: "is required"}
	}
	if req.BusinessDate.IsZero() {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "business_date", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return nil, &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	bd := req.BusinessDate.UTC().Truncate(24 * time.Hour)

	// 1. Fund must exist and be alive.
	fund, err := h.funds.GetByID(ctx, req.FundID)
	if err != nil {
		return nil, fmt.Errorf("loading fund: %w", err)
	}
	if fund == nil || fund.DeletedAt != nil {
		return nil, &domain.ErrFundNotFound{FundID: req.FundID.String()}
	}

	// 2. Idempotency check — bail out early if today's snapshot already
	// exists. The (scope, business_date, source) tuple is the unique key.
	existing, err := h.valuation.GetAUM(ctx, vo.AumScopeFund, fund.ID, bd, vo.ValuationSourceInternal)
	if err != nil {
		return nil, fmt.Errorf("loading existing aum: %w", err)
	}
	if existing != nil {
		return &ComputeFundAUMResult{
			Snapshot:       existing,
			PortfolioCount: 0,
			Idempotent:     true,
		}, nil
	}

	// 3. List active portfolios under the fund.
	activeStatus := vo.PortfolioStatusActive
	portfolios, _, err := h.portfolios.List(ctx, domain.PortfolioListFilter{
		FundID: &fund.ID,
		Status: &activeStatus,
		Limit:  10_000,
	})
	if err != nil {
		return nil, fmt.Errorf("listing fund portfolios: %w", err)
	}
	if len(portfolios) == 0 {
		return nil, &domain.ErrIncompleteFundValuation{
			FundID:       fund.ID.String(),
			BusinessDate: bd.Format("2006-01-02"),
			Reason:       "MISSING_PORTFOLIO_SNAPSHOT",
			Detail:       "fund has no active portfolios",
		}
	}

	// 4. Pull each portfolio's snapshot, validating consistency.
	totalAUM := decimal.Zero
	priceSetHash := ""
	breakdowns := make([]PortfolioAUMBreakdown, 0, len(portfolios))
	for _, p := range portfolios {
		snap, snapErr := h.valuation.GetByPortfolioBusinessDate(ctx, p.ID, bd, vo.ValuationSourceInternal)
		if snapErr != nil {
			return nil, fmt.Errorf("loading valuation for portfolio %s: %w", p.ID, snapErr)
		}
		if snap == nil {
			return nil, &domain.ErrIncompleteFundValuation{
				FundID:       fund.ID.String(),
				BusinessDate: bd.Format("2006-01-02"),
				Reason:       "MISSING_PORTFOLIO_SNAPSHOT",
				Detail:       fmt.Sprintf("portfolio %s has no valuation snapshot", p.ID),
			}
		}
		if snap.ValuationCcy != fund.BaseCurrency {
			return nil, &domain.ErrIncompleteFundValuation{
				FundID:       fund.ID.String(),
				BusinessDate: bd.Format("2006-01-02"),
				Reason:       "VALUATION_CCY_MISMATCH",
				Detail: fmt.Sprintf(
					"portfolio %s valuation_ccy %q does not match fund base_currency %q",
					p.ID, snap.ValuationCcy, fund.BaseCurrency,
				),
			}
		}
		if priceSetHash == "" {
			priceSetHash = snap.PriceSetHash
		} else if priceSetHash != snap.PriceSetHash {
			return nil, &domain.ErrIncompleteFundValuation{
				FundID:       fund.ID.String(),
				BusinessDate: bd.Format("2006-01-02"),
				Reason:       "PRICE_SET_HASH_MISMATCH",
				Detail: fmt.Sprintf(
					"portfolio %s price_set_hash %q differs from previously seen %q",
					p.ID, snap.PriceSetHash, priceSetHash,
				),
			}
		}
		totalAUM = totalAUM.Add(snap.AUM)
		breakdowns = append(breakdowns, PortfolioAUMBreakdown{
			PortfolioID:  p.ID,
			AUM:          snap.AUM,
			ValuationCcy: snap.ValuationCcy,
			PriceSetHash: snap.PriceSetHash,
		})
	}

	// 5. Persist the fund-level AUM row in a single tx, with another idempotency
	// guard inside the tx in case a parallel runner won the race.
	now := h.now()
	snap := &entity.AUMSnapshot{
		ID:           uuid.New(),
		ScopeType:    vo.AumScopeFund,
		ScopeID:      fund.ID,
		BusinessDate: bd,
		AUM:          totalAUM,
		ValuationCcy: fund.BaseCurrency,
		Source:       vo.ValuationSourceInternal,
		CreatedAt:    now,
		CreatedBy:    req.ActorID,
	}

	idempotent := false
	err = withTransaction(ctx, h.pool, func(tx pgx.Tx) error {
		raced, raceErr := h.valuation.GetAUM(ctx, vo.AumScopeFund, fund.ID, bd, vo.ValuationSourceInternal)
		if raceErr != nil {
			return fmt.Errorf("re-checking aum inside tx: %w", raceErr)
		}
		if raced != nil {
			snap = raced
			idempotent = true
			return nil
		}
		if insertErr := h.valuation.InsertAUM(ctx, tx, snap); insertErr != nil {
			// Defensive: a concurrent inserter could have raced past our pre-check.
			// Re-read once more before propagating.
			if isUniqueViolation(insertErr) {
				raced, raceErr := h.valuation.GetAUM(ctx, vo.AumScopeFund, fund.ID, bd, vo.ValuationSourceInternal)
				if raceErr == nil && raced != nil {
					snap = raced
					idempotent = true
					return nil
				}
			}
			return fmt.Errorf("inserting fund aum snapshot: %w", insertErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if !idempotent {
		_ = h.audit.LogAction(contract.AuditEntry{
			ActorID:      req.ActorID.String(),
			Action:       "INVESTMENT_FUND_AUM_COMPUTED",
			Module:       "investment",
			ResourceType: "INVESTMENT_FUND_AUM",
			ResourceID:   snap.ID.String(),
			Details: map[string]any{
				"fund_id":         fund.ID,
				"aum":             snap.AUM.String(),
				"valuation_ccy":   snap.ValuationCcy,
				"portfolio_count": len(breakdowns),
				"price_set_hash":  priceSetHash,
			},
			BusinessDate: bd,
		})
	}

	return &ComputeFundAUMResult{
		Snapshot:       snap,
		PortfolioCount: len(breakdowns),
		PortfolioAUMs:  breakdowns,
		Idempotent:     idempotent,
	}, nil
}

// isUniqueViolation matches a Postgres unique-violation SQLSTATE without
// importing pgconn here. Kept narrow — the broader ErrIs API returns true for
// any pgx error, which is too loose.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	type sqlStateProvider interface{ SQLState() string }
	var pe sqlStateProvider
	if errors.As(err, &pe) {
		return pe.SQLState() == "23505"
	}
	return false
}
