package query

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/domain"
	integrationperm "github.com/neo-kanta/ims-th-solution/backend/internal/integration/permission"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ErrInvalidValuationScope is returned when the caller supplies a scope
// value other than "company" or "mine".
var ErrInvalidValuationScope = errors.New(`invalid scope: must be "company" or "mine"`)

// GetValuationSummaryHandler aggregates today's AUM and P&L for the
// dashboard, scoped to either the caller's full accessible data set
// ("company") or the funds the caller manages ("mine").
//
// Aggregation math lives in the investment module (behind
// contract.ValuationSummaryProvider) — this handler only resolves identity
// and data scope from the authenticated session and maps the result.
type GetValuationSummaryHandler struct {
	iam       IAMPort
	valuation contract.ValuationSummaryProvider
}

// NewGetValuationSummaryHandler wires the handler. valuation may be nil in
// tests that only exercise the permission/no-data paths.
func NewGetValuationSummaryHandler(iam IAMPort, valuation contract.ValuationSummaryProvider) *GetValuationSummaryHandler {
	return &GetValuationSummaryHandler{iam: iam, valuation: valuation}
}

// Execute resolves the caller's data scope and returns the aggregate
// valuation summary for the requested scope.
//
// username is the JWT-derived identity (auth claims' Username) supplied by
// the transport layer — it is never accepted from client input, so "mine"
// cannot be used to impersonate another user.
func (h *GetValuationSummaryHandler) Execute(
	ctx context.Context,
	userID string,
	username string,
	scopeParam string,
) (*domain.ValuationSummary, error) {
	scope := domain.ValuationScope(scopeParam)
	if scope == "" {
		scope = domain.ValuationScopeCompany
	}
	if scope != domain.ValuationScopeCompany && scope != domain.ValuationScopeMine {
		return nil, ErrInvalidValuationScope
	}

	resolvedUsername := ""
	if scope == domain.ValuationScopeMine {
		resolvedUsername = username
	}

	ok, err := h.iam.HasFunctionPermission(ctx, userID, integrationperm.DashboardView)
	if err != nil {
		return nil, err
	}
	if !ok || h.valuation == nil {
		return &domain.ValuationSummary{Scope: scope, Username: resolvedUsername, DataAvailable: false}, nil
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	accessibleFundIDs, err := h.accessibleFundIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	res, err := h.valuation.GetValuationSummary(ctx, contract.ValuationSummaryRequest{
		Scope:             contract.ValuationSummaryScope(scope),
		UserID:            uid,
		AccessibleFundIDs: accessibleFundIDs,
	})
	if err != nil {
		return nil, err
	}
	if res == nil || !res.DataAvailable {
		return &domain.ValuationSummary{Scope: scope, Username: resolvedUsername, DataAvailable: false}, nil
	}

	return &domain.ValuationSummary{
		Scope:           scope,
		Username:        resolvedUsername,
		BusinessDate:    res.BusinessDate,
		Currency:        res.Currency,
		AUM:             res.AUM,
		TodayPnL:        res.TodayPnL,
		TodayPnLPercent: res.TodayPnLPercent,
		AsOf:            res.AsOf,
		DataAvailable:   true,
	}, nil
}

// accessibleFundIDs converts the caller's data-scope contract list into fund
// UUIDs. nil means "no filter" (the "*" wildcard is present in the caller's
// data scope); a non-nil (possibly empty) slice restricts the aggregate to
// exactly those funds. Mirrors the convention used by the investment
// module's own handlers (accessibleFundIDs in transport/handler).
func (h *GetValuationSummaryHandler) accessibleFundIDs(ctx context.Context, userID string) ([]uuid.UUID, error) {
	contracts, err := h.iam.GetAccessibleContracts(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(contracts))
	for _, c := range contracts {
		if c == "*" {
			return nil, nil
		}
		if id, parseErr := uuid.Parse(c); parseErr == nil {
			out = append(out, id)
		}
	}
	return out, nil
}
