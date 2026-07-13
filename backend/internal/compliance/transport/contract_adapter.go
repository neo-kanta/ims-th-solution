// Package transport contains the compliance module's public, cross-module
// contract adapters in addition to its HTTP layer.
//
// ComplianceContractAdapter implements the cross-module interfaces defined in
// pkg/contract (ComplianceChecker, PostTradeVerifier) so that consumers
// (investment, workflow) never import compliance internals.
package transport

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/engine"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ComplianceContractAdapter wraps the compliance pipeline / command handlers and
// exposes them via the stable pkg/contract interfaces.
//
// The adapter is the ONLY compliance type that external modules are permitted
// to hold a reference to. All internal concerns (pipeline, registry, repos)
// stay behind this boundary.
type ComplianceContractAdapter struct {
	preTrade *command.RunPreTradeCheckHandler
	pipeline *engine.Pipeline
	registry *spi.RuleRegistry
}

// NewComplianceContractAdapter wires the adapter. Callers typically do not
// construct this directly — compliance.Module returns it via ContractAdapter().
func NewComplianceContractAdapter(
	preTrade *command.RunPreTradeCheckHandler,
	pipeline *engine.Pipeline,
	registry *spi.RuleRegistry,
) *ComplianceContractAdapter {
	return &ComplianceContractAdapter{
		preTrade: preTrade,
		pipeline: pipeline,
		registry: registry,
	}
}

// Compile-time contract assertions.
var (
	_ contract.ComplianceChecker   = (*ComplianceContractAdapter)(nil)
	_ contract.ComplianceSimulator = (*ComplianceContractAdapter)(nil)
	_ contract.PostTradeVerifier   = (*ComplianceContractAdapter)(nil)
)

// ─────────────────────────────────────────────────────────────────────────────
// ComplianceChecker — pre-trade gate for OMS
// ─────────────────────────────────────────────────────────────────────────────

// CheckProposedOrder delegates to the pre-trade handler and translates the
// compliance-internal response into the contract shape.
func (a *ComplianceContractAdapter) CheckProposedOrder(
	ctx context.Context,
	req contract.ProposedOrderCheck,
) (*contract.ProposedOrderResult, error) {
	return a.checkProposedOrder(ctx, req, true)
}

// SimulateProposedOrder delegates to the pre-trade handler in dry-run mode.
// The rule evaluation and returned evidence match CheckProposedOrder, but no
// compliance_check_records or compliance_breaches rows are inserted.
func (a *ComplianceContractAdapter) SimulateProposedOrder(
	ctx context.Context,
	req contract.ProposedOrderCheck,
) (*contract.ProposedOrderResult, error) {
	return a.checkProposedOrder(ctx, req, false)
}

func (a *ComplianceContractAdapter) checkProposedOrder(
	ctx context.Context,
	req contract.ProposedOrderCheck,
	persist bool,
) (*contract.ProposedOrderResult, error) {
	if a == nil || a.preTrade == nil {
		return nil, fmt.Errorf("compliance contract adapter not initialised")
	}

	internalReq := command.PreTradeCheckRequest{
		CheckGroupID: req.CheckGroupID,
		PortfolioID:  req.PortfolioID,
		ContractID:   req.ContractID,
		BusinessDate: req.BusinessDate,
		Actor:        req.Actor,
		OrderID:      req.OrderID,
		Ticker:       req.Ticker,
		Side:         mapOrderSideFromContract(req.Side),
		Quantity:     req.Quantity,
		Price:        req.Price,
		Fees:         req.Fees,
		Currency:     req.Currency,
		Exchange:     req.Exchange,
	}

	var resp *command.PreTradeCheckResponse
	var err error
	if persist {
		resp, err = a.preTrade.Handle(ctx, internalReq)
	} else {
		resp, err = a.preTrade.HandleDryRun(ctx, internalReq)
	}
	if err != nil {
		return nil, mapPreTradeError(err)
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
// PostTradeVerifier — workflow gate before TRANSACTION_CLOSED
// ─────────────────────────────────────────────────────────────────────────────

// RunPostTradeVerification executes the post-trade pipeline at contract scope
// and reports whether any BLOCK-level breach would forbid closing transactions.
//
// The scan is scope-level (Global + Contract); no single ProposedOrder is
// supplied. Per-portfolio scans remain available via the REST endpoint.
func (a *ComplianceContractAdapter) RunPostTradeVerification(
	ctx context.Context,
	contractID uuid.UUID,
	businessDate time.Time,
) (*contract.PostTradeVerificationResult, error) {
	if a == nil || a.pipeline == nil {
		return nil, fmt.Errorf("compliance contract adapter not initialised")
	}
	if contractID == uuid.Nil {
		return nil, fmt.Errorf("contract_id is required")
	}
	if businessDate.IsZero() {
		return nil, fmt.Errorf("business_date is required")
	}

	checkGroupID := uuid.New()
	input := spi.CheckInput{
		CheckGroupID: checkGroupID,
		Timing:       vo.TimingPostTrade,
		BusinessDate: businessDate.UTC(),
		Actor:        "workflow:close_transactions",
		ContractID:   contractID,
		// PortfolioID left as uuid.Nil — scan is contract-wide.
		ProposedOrder: nil,
	}

	scopes := []vo.Scope{{Type: vo.ScopeGlobal, ID: nil}}
	cid := contractID
	scopes = append(scopes, vo.Scope{Type: vo.ScopeContract, ID: &cid})

	output, err := a.pipeline.RunCheck(ctx, input, scopes)
	if err != nil {
		return nil, fmt.Errorf("running post-trade verification: %w", err)
	}

	result := &contract.PostTradeVerificationResult{
		CheckGroupID: checkGroupID,
		Verdict:      mapVerdictToContract(output.FinalVerdict),
	}
	for _, b := range output.Breaches {
		if b.Verdict == vo.VerdictBlock {
			result.HasBlockingBreach = true
		}
		result.Breaches = append(result.Breaches, contract.PostTradeBreach{
			BreachID:   b.ID,
			RuleTypeID: b.RuleTypeID,
			Verdict:    mapVerdictToContract(b.Verdict),
			Severity:   string(b.Severity),
			Message:    b.Message,
		})
	}
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// mapping helpers
// ─────────────────────────────────────────────────────────────────────────────

// mapPreTradeError translates the internal/compliance validation error type
// into the exported pkg/contract sentinel so callers across the module
// boundary (investment) can classify pre-trade failures with errors.Is
// without importing internal/compliance — same pattern as mapBindingError.
func mapPreTradeError(err error) error {
	var invalid *domain.ErrInvalidPreTradeRequest
	if errors.As(err, &invalid) {
		return fmt.Errorf("%w: %s", contract.ErrInvalidProposedOrder, err.Error())
	}
	return err
}

func mapOrderSideFromContract(s contract.ComplianceOrderSide) vo.OrderSide {
	switch s {
	case contract.ComplianceOrderSideBuy:
		return vo.OrderSideBuy
	case contract.ComplianceOrderSideSell:
		return vo.OrderSideSell
	default:
		return vo.OrderSide(string(s))
	}
}

func mapVerdictToContract(v vo.Verdict) contract.ComplianceVerdict {
	switch v {
	case vo.VerdictPass:
		return contract.ComplianceVerdictPass
	case vo.VerdictWarn:
		return contract.ComplianceVerdictWarn
	case vo.VerdictBlock:
		return contract.ComplianceVerdictBlock
	default:
		return contract.ComplianceVerdict(string(v))
	}
}
