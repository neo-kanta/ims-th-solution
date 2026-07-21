package transport

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// PortfolioContractAdapter implements contract.PortfolioComplianceContract —
// the Portfolio Compliance V2 surface keyed by portfolio_id with fund_id
// optional. It embeds ComplianceContractAdapter so CheckProposedOrder /
// SimulateProposedOrder are inherited unchanged: those methods already treat
// ContractID as optional (uuid.Nil means "no contract"), which is exactly the
// V2 semantics.
//
// Like ComplianceContractAdapter, this is the ONLY compliance type external
// modules (investment) are permitted to hold a reference to for the V2
// surface. Internal errors are translated to the pkg/contract sentinel errors
// before crossing the boundary — investment classifies failures with
// errors.Is/errors.As against those sentinels, never against
// internal/compliance error types.
type PortfolioContractAdapter struct {
	*ComplianceContractAdapter

	postTrade            *command.RunPostTradeCheckHandler
	instanceRepo         domain.RuleInstanceRepository
	bindingRepo          domain.RuleBindingRepository
	breachRepo           domain.BreachRepository
	createBindingCmd     *command.CreateRuleBindingHandler
	deactivateBindingCmd *command.DeactivateRuleBindingHandler
}

// NewPortfolioContractAdapter wires the adapter. Callers typically do not
// construct this directly — compliance.Module returns it via
// PortfolioContractAdapter().
func NewPortfolioContractAdapter(
	checkerAdapter *ComplianceContractAdapter,
	postTrade *command.RunPostTradeCheckHandler,
	instanceRepo domain.RuleInstanceRepository,
	bindingRepo domain.RuleBindingRepository,
	breachRepo domain.BreachRepository,
	createBindingCmd *command.CreateRuleBindingHandler,
	deactivateBindingCmd *command.DeactivateRuleBindingHandler,
) *PortfolioContractAdapter {
	return &PortfolioContractAdapter{
		ComplianceContractAdapter: checkerAdapter,
		postTrade:                 postTrade,
		instanceRepo:              instanceRepo,
		bindingRepo:               bindingRepo,
		breachRepo:                breachRepo,
		createBindingCmd:          createBindingCmd,
		deactivateBindingCmd:      deactivateBindingCmd,
	}
}

var _ contract.PortfolioComplianceContract = (*PortfolioContractAdapter)(nil)

// ─────────────────────────────────────────────────────────────────────────────
// Post-trade
// ─────────────────────────────────────────────────────────────────────────────

// RunPortfolioPostTradeCheck runs a post-trade scan scoped to a single
// portfolio. req.FundID may be uuid.Nil — the scan then runs at GLOBAL +
// PORTFOLIO scope only.
func (a *PortfolioContractAdapter) RunPortfolioPostTradeCheck(
	ctx context.Context,
	req contract.PortfolioPostTradeRequest,
) (*contract.ProposedOrderResult, error) {
	if a == nil || a.postTrade == nil {
		return nil, fmt.Errorf("portfolio contract adapter not initialised")
	}
	resp, err := a.postTrade.Handle(ctx, command.PostTradeCheckRequest{
		PortfolioID:  req.PortfolioID,
		ContractID:   req.FundID,
		BusinessDate: req.BusinessDate,
		Actor:        req.Actor,
	})
	if err != nil {
		return nil, err
	}
	out := &contract.ProposedOrderResult{
		CheckGroupID:   resp.CheckGroupID,
		Verdict:        mapVerdictToContract(resp.Verdict),
		RulesEvaluated: resp.RulesEvaluated,
	}
	for _, b := range resp.Breaches {
		out.Breaches = append(out.Breaches, contract.ProposedOrderBreach{
			BreachID:    b.BreachID,
			RuleTypeID:  b.RuleTypeID,
			Verdict:     mapVerdictToContract(b.Verdict),
			Severity:    string(b.Severity),
			Message:     b.Message,
			Overridable: b.Overridable,
		})
	}
	return out, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Rule catalog + binding administration
// ─────────────────────────────────────────────────────────────────────────────

// ListPortfolioRules returns every active rule instance, annotated with its
// current binding to portfolioID when one exists (active or inactive — the
// most recent one wins so a deactivated binding is still visible for audit).
func (a *PortfolioContractAdapter) ListPortfolioRules(
	ctx context.Context,
	portfolioID uuid.UUID,
) ([]contract.PortfolioRuleCatalogEntry, error) {
	if a == nil || a.instanceRepo == nil || a.bindingRepo == nil {
		return nil, fmt.Errorf("portfolio contract adapter not initialised")
	}

	active := true
	instances, _, err := a.instanceRepo.List(ctx, domain.RuleInstanceFilter{IsActive: &active, Limit: 500})
	if err != nil {
		return nil, fmt.Errorf("listing rule instances: %w", err)
	}

	scopeType := vo.ScopePortfolio
	bindings, _, err := a.bindingRepo.List(ctx, domain.BindingFilter{
		ScopeType: &scopeType,
		ScopeID:   &portfolioID,
		Limit:     500,
	})
	if err != nil {
		return nil, fmt.Errorf("listing rule bindings: %w", err)
	}

	latestByInstance := map[uuid.UUID]entity.RuleBinding{}
	for _, b := range bindings {
		cur, ok := latestByInstance[b.RuleInstanceID]
		if !ok || (b.IsActive && !cur.IsActive) ||
			(b.IsActive == cur.IsActive && b.CreatedAt.After(cur.CreatedAt)) {
			latestByInstance[b.RuleInstanceID] = b
		}
	}

	// Best-effort: surface the current parameter snapshots so the portfolio
	// settings UI can render thresholds without a second round-trip. A
	// lookup failure here must not fail the whole catalog listing.
	instanceIDs := make([]uuid.UUID, 0, len(instances))
	for _, inst := range instances {
		instanceIDs = append(instanceIDs, inst.ID)
	}
	versions, err := a.instanceRepo.GetCurrentVersions(ctx, instanceIDs)
	if err != nil {
		versions = nil
	}

	out := make([]contract.PortfolioRuleCatalogEntry, 0, len(instances))
	for _, inst := range instances {
		entry := contract.PortfolioRuleCatalogEntry{
			RuleInstanceID: inst.ID,
			RuleTypeID:     inst.RuleTypeID,
			Name:           inst.Name,
			Description:    inst.Description,
			IsActive:       inst.IsActive,
		}
		if version := versions[inst.ID]; version != nil && len(version.Parameters) > 0 {
			entry.Parameters = version.Parameters
		}
		if b, ok := latestByInstance[inst.ID]; ok {
			entry.Binding = &contract.PortfolioRuleBindingView{
				BindingID:     b.ID,
				Severity:      string(b.Severity),
				Priority:      b.Priority,
				IsActive:      b.IsActive,
				EffectiveFrom: b.EffectiveWindow.ValidFrom,
				EffectiveTo:   b.EffectiveWindow.ValidTo,
			}
		}
		out = append(out, entry)
	}
	return out, nil
}

// BindPortfolioRule binds a rule instance to a portfolio (scope_type=PORTFOLIO).
func (a *PortfolioContractAdapter) BindPortfolioRule(
	ctx context.Context,
	req contract.PortfolioRuleBindingRequest,
) (*contract.PortfolioRuleBindingView, error) {
	if a == nil || a.createBindingCmd == nil {
		return nil, fmt.Errorf("portfolio contract adapter not initialised")
	}

	pid := req.PortfolioID
	binding, err := a.createBindingCmd.Handle(ctx, command.CreateRuleBindingRequest{
		RuleInstanceID: req.RuleInstanceID,
		ScopeType:      vo.ScopePortfolio,
		ScopeID:        &pid,
		Severity:       vo.Severity(req.Severity),
		Priority:       req.Priority,
		EffectiveFrom:  req.EffectiveFrom,
		EffectiveTo:    req.EffectiveTo,
		CreatedBy:      req.ActorID,
	})
	if err != nil {
		return nil, mapBindingError(err)
	}

	return &contract.PortfolioRuleBindingView{
		BindingID:     binding.ID,
		Severity:      string(binding.Severity),
		Priority:      binding.Priority,
		IsActive:      binding.IsActive,
		EffectiveFrom: binding.EffectiveWindow.ValidFrom,
		EffectiveTo:   binding.EffectiveWindow.ValidTo,
	}, nil
}

// DeactivatePortfolioRuleBinding deactivates a binding after verifying it
// actually belongs to portfolioID — a caller cannot deactivate another
// portfolio's binding by guessing its UUID.
func (a *PortfolioContractAdapter) DeactivatePortfolioRuleBinding(
	ctx context.Context,
	portfolioID, bindingID uuid.UUID,
) error {
	if a == nil || a.bindingRepo == nil || a.deactivateBindingCmd == nil {
		return fmt.Errorf("portfolio contract adapter not initialised")
	}
	existing, err := a.bindingRepo.GetByID(ctx, bindingID)
	if err != nil {
		return fmt.Errorf("loading rule binding: %w", err)
	}
	if existing == nil || existing.Scope.Type != vo.ScopePortfolio ||
		existing.Scope.ID == nil || *existing.Scope.ID != portfolioID {
		// Binding does not exist, or belongs to a different scope/portfolio —
		// report NotFound either way so callers cannot probe for other
		// portfolios' binding IDs.
		return fmt.Errorf("%w: %s", contract.ErrPortfolioBindingNotFound, bindingID)
	}
	if err := a.deactivateBindingCmd.Handle(ctx, command.DeactivateRuleBindingRequest{BindingID: bindingID}); err != nil {
		return mapBindingError(err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Breaches
// ─────────────────────────────────────────────────────────────────────────────

// ListPortfolioBreaches lists breaches for a single portfolio.
func (a *PortfolioContractAdapter) ListPortfolioBreaches(
	ctx context.Context,
	portfolioID uuid.UUID,
	filter contract.PortfolioBreachFilter,
) ([]contract.PortfolioBreachView, error) {
	if a == nil || a.breachRepo == nil {
		return nil, fmt.Errorf("portfolio contract adapter not initialised")
	}

	domainFilter := domain.BreachFilter{
		PortfolioID: &portfolioID,
		RuleTypeID:  filter.RuleTypeID,
		DateFrom:    filter.DateFrom,
		DateTo:      filter.DateTo,
		Offset:      filter.Offset,
		Limit:       filter.Limit,
	}
	if filter.Status != nil {
		s := entity.BreachStatus(*filter.Status)
		domainFilter.Status = &s
	}

	breaches, _, err := a.breachRepo.List(ctx, domainFilter)
	if err != nil {
		return nil, fmt.Errorf("listing portfolio breaches: %w", err)
	}

	out := make([]contract.PortfolioBreachView, 0, len(breaches))
	for _, b := range breaches {
		out = append(out, contract.PortfolioBreachView{
			BreachID:     b.ID,
			RuleTypeID:   b.RuleTypeID,
			Severity:     string(b.Severity),
			Verdict:      string(b.Verdict),
			Status:       string(b.Status),
			Message:      b.Message,
			BusinessDate: b.BusinessDate,
			CreatedAt:    b.CreatedAt,
		})
	}
	return out, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// error mapping
// ─────────────────────────────────────────────────────────────────────────────

// mapBindingError translates internal/compliance error types raised by the
// binding CRUD commands into the exported pkg/contract sentinels so callers
// across the module boundary never need to import internal/compliance.
func mapBindingError(err error) error {
	var notFound *domain.ErrRuleInstanceNotFound
	var inactive *domain.ErrRuleInstanceInactive
	var dup *domain.ErrDuplicateActiveBinding
	var invalidCreate *command.ErrInvalidCreateBindingRequest
	var bindingNotFound *domain.ErrBindingNotFound

	switch {
	case errors.As(err, &notFound):
		return fmt.Errorf("%w: %s", contract.ErrPortfolioRuleNotFound, err.Error())
	case errors.As(err, &inactive):
		return fmt.Errorf("%w: %s", contract.ErrPortfolioRuleInactive, err.Error())
	case errors.As(err, &dup):
		return fmt.Errorf("%w: %s", contract.ErrPortfolioBindingDuplicate, err.Error())
	case errors.As(err, &invalidCreate):
		return fmt.Errorf("%w: %s", contract.ErrPortfolioBindingInvalid, err.Error())
	case errors.As(err, &bindingNotFound):
		return fmt.Errorf("%w: %s", contract.ErrPortfolioBindingNotFound, err.Error())
	default:
		return err
	}
}
