package query

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ErrInvalidValuationScope is returned when the caller supplies a scope
// value other than "company" or "mine".
var ErrInvalidValuationScope = errors.New(`invalid scope: must be "company" or "mine"`)

// GetValuationSummaryHandler aggregates latest-available AUM and P&L for the
// dashboard, scoped to either the full company ("company") or the portfolios
// the caller manages inside funds they are authorized to access ("mine").
//
// Aggregation math lives in the investment module (behind
// contract.ValuationSummaryProvider) — this handler only resolves identity
// and data scope from the authenticated session and maps the result.
type GetValuationSummaryHandler struct {
	iam               IAMPort
	valuation         contract.ValuationSummaryProvider
	reportingCurrency string
}

type reportingCurrencyProvider interface {
	ReportingCurrency() string
}

// NewGetValuationSummaryHandler wires the handler. valuation may be nil in
// tests that only exercise the permission/no-data paths.
func NewGetValuationSummaryHandler(iam IAMPort, valuation contract.ValuationSummaryProvider) *GetValuationSummaryHandler {
	h := &GetValuationSummaryHandler{iam: iam, valuation: valuation}
	if provider, ok := valuation.(reportingCurrencyProvider); ok {
		h.reportingCurrency = provider.ReportingCurrency()
	}
	return h
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

	if h.valuation == nil {
		return unavailableValuationSummary(scope, resolvedUsername, contract.ValuationSummaryStatusIncomplete, h.reportingCurrency), nil
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// Company AUM is an authenticated company-wide metric by explicit owner
	// policy. A nil fund filter therefore means every active company fund,
	// independent of the caller's portfolio/fund data scope. The "mine" view
	// remains bounded by the caller's effective fund data permissions while the
	// investment provider applies portfolio-manager ownership.
	var accessibleFundIDs []uuid.UUID
	if scope == domain.ValuationScopeMine {
		accessibleFundIDs, err = h.accessibleFundIDs(ctx, userID)
		if err != nil {
			return nil, err
		}
	}

	res, err := h.valuation.GetValuationSummary(ctx, contract.ValuationSummaryRequest{
		Scope:             contract.ValuationSummaryScope(scope),
		UserID:            uid,
		AccessibleFundIDs: accessibleFundIDs,
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return unavailableValuationSummary(scope, resolvedUsername, contract.ValuationSummaryStatusIncomplete, h.reportingCurrency), nil
	}
	status := res.Status
	if status == "" {
		status = contract.ValuationSummaryStatusNoData
		if res.DataAvailable {
			status = contract.ValuationSummaryStatusAvailable
		}
	}
	currency := res.Currency
	if currency == "" {
		currency = h.reportingCurrency
	}

	coverage := res.Coverage
	if scope == domain.ValuationScopeCompany {
		// Company visibility grants the aggregate only. Item-level exclusions can
		// carry fund and portfolio business codes, so omit them for a response
		// available to every authenticated user. Aggregate counts, currencies,
		// dates, and stable reasons remain available for completeness UX.
		coverage.Exclusions = nil
	}

	return &domain.ValuationSummary{
		Scope:           scope,
		Username:        resolvedUsername,
		Status:          status,
		BusinessDate:    res.BusinessDate,
		Currency:        currency,
		AUM:             res.AUM,
		TodayPnL:        res.TodayPnL,
		TodayPnLPercent: res.TodayPnLPercent,
		AsOf:            res.AsOf,
		DataAvailable:   res.DataAvailable,
		Coverage:        coverage,
	}, nil
}

func unavailableValuationSummary(
	scope domain.ValuationScope,
	username string,
	status contract.ValuationSummaryStatus,
	reportingCurrency string,
) *domain.ValuationSummary {
	return &domain.ValuationSummary{
		Scope:         scope,
		Username:      username,
		Status:        status,
		Currency:      reportingCurrency,
		DataAvailable: false,
	}
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
