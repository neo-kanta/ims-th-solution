// Package command contains write-side application use cases for the investment module.
package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// SubmitDecisionForExecutionRequest is the input to the pre-trade gate.
type SubmitDecisionForExecutionRequest struct {
	DecisionID uuid.UUID
	ActorID    uuid.UUID // authenticated user submitting the decision
}

// SubmitDecisionForExecutionResult is returned to the caller on success.
// On a BLOCK verdict the handler returns a typed *domain.ErrComplianceRejected
// error instead — the HTTP layer maps that to 409.
type SubmitDecisionForExecutionResult struct {
	DecisionID     uuid.UUID
	Status         vo.DecisionStatus
	CheckGroupID   uuid.UUID
	Verdict        contract.ComplianceVerdict
	Breaches       []contract.ProposedOrderBreach
	RulesEvaluated int
	SubmittedAt    time.Time
}

// SubmitDecisionForExecutionHandler promotes a DRAFT investment decision to
// SUBMITTED only after the IRG pre-trade pipeline clears it.
//
// Flow:
//  1. Load the decision; refuse if not in DRAFT.
//  2. Translate to a contract.ProposedOrderCheck and invoke the
//     ComplianceChecker (synchronous).
//  3. On BLOCK: mark the decision BLOCKED, persist the CheckGroupID, return
//     *domain.ErrComplianceRejected. No order is submitted to OMS.
//  4. On PASS/WARN: mark the decision SUBMITTED and persist the CheckGroupID.
//
// Cross-module boundary: ComplianceChecker is the ONLY compliance-side
// dependency. Handler never imports internal/compliance.
type SubmitDecisionForExecutionHandler struct {
	decisions  domain.DecisionRepository
	compliance contract.ComplianceChecker
	now        func() time.Time // injectable clock for deterministic tests
}

// NewSubmitDecisionForExecutionHandler wires the handler. `now` may be nil,
// in which case time.Now().UTC() is used.
func NewSubmitDecisionForExecutionHandler(
	decisions domain.DecisionRepository,
	checker contract.ComplianceChecker,
	now func() time.Time,
) *SubmitDecisionForExecutionHandler {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &SubmitDecisionForExecutionHandler{
		decisions:  decisions,
		compliance: checker,
		now:        now,
	}
}

// Handle executes the pre-trade submit flow. See struct doc for semantics.
func (h *SubmitDecisionForExecutionHandler) Handle(
	ctx context.Context,
	req SubmitDecisionForExecutionRequest,
) (*SubmitDecisionForExecutionResult, error) {
	if h == nil || h.decisions == nil || h.compliance == nil {
		return nil, fmt.Errorf("submit decision handler not initialised")
	}
	if err := validateSubmitRequest(req); err != nil {
		return nil, err
	}

	decision, err := h.decisions.GetByID(ctx, req.DecisionID)
	if err != nil {
		var notFound *domain.ErrDecisionNotFound
		if errors.As(err, &notFound) {
			return nil, err
		}
		return nil, fmt.Errorf("loading decision %s: %w", req.DecisionID, err)
	}
	if decision == nil {
		return nil, &domain.ErrDecisionNotFound{DecisionID: req.DecisionID.String()}
	}
	if !decision.CanSubmitForExecution() {
		return nil, &domain.ErrDecisionNotDraft{
			DecisionID:    decision.ID.String(),
			CurrentStatus: string(decision.Status),
		}
	}

	// Translate optional decimal pointers to the contract's value type.
	// Missing quantity defaults to zero so the IRG pipeline can still
	// evaluate amount-based rules.
	var qty, price decimal.Decimal
	if decision.Quantity != nil {
		qty = *decision.Quantity
	}
	if decision.LimitPrice != nil {
		price = *decision.LimitPrice
	}
	checkReq := contract.ProposedOrderCheck{
		PortfolioID:  decision.PortfolioID,
		ContractID:   decision.ContractID,
		BusinessDate: decision.BusinessDate,
		Actor:        req.ActorID.String(),
		OrderID:      decision.ID,
		Ticker:       decision.InstrumentCode,
		Side:         mapOrderSideToContract(decision.Side),
		Quantity:     qty,
		Price:        price,
		Currency:     decision.Currency,
		Exchange:     decision.Exchange,
	}

	checkResult, err := h.compliance.CheckProposedOrder(ctx, checkReq)
	if err != nil {
		return nil, fmt.Errorf("running pre-trade compliance check: %w", err)
	}
	if checkResult == nil {
		return nil, fmt.Errorf("compliance checker returned nil result")
	}

	now := h.now()

	// BLOCK verdict → mark BLOCKED and refuse submission.
	if checkResult.Verdict == contract.ComplianceVerdictBlock {
		if updateErr := h.decisions.UpdateStatus(
			ctx, decision.ID, vo.DecisionStatusBlocked,
			checkResult.CheckGroupID, req.ActorID, now,
		); updateErr != nil {
			// Persist failure is secondary; still surface the BLOCK to caller.
			return nil, fmt.Errorf("persisting blocked decision status: %w", updateErr)
		}
		msg := summarizeBreaches(checkResult.Breaches)
		return nil, &domain.ErrComplianceRejected{
			DecisionID:   decision.ID.String(),
			CheckGroupID: checkResult.CheckGroupID.String(),
			Message:      msg,
		}
	}

	// PASS / WARN → promote to SUBMITTED. WARN does not block submission but is
	// captured in the breach list for traders to acknowledge downstream.
	if err := h.decisions.UpdateStatus(
		ctx, decision.ID, vo.DecisionStatusSubmitted,
		checkResult.CheckGroupID, req.ActorID, now,
	); err != nil {
		return nil, fmt.Errorf("persisting submitted decision status: %w", err)
	}

	return &SubmitDecisionForExecutionResult{
		DecisionID:     decision.ID,
		Status:         vo.DecisionStatusSubmitted,
		CheckGroupID:   checkResult.CheckGroupID,
		Verdict:        checkResult.Verdict,
		Breaches:       checkResult.Breaches,
		RulesEvaluated: checkResult.RulesEvaluated,
		SubmittedAt:    now,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

func validateSubmitRequest(req SubmitDecisionForExecutionRequest) error {
	if req.DecisionID == uuid.Nil {
		return &domain.ErrInvalidDecisionRequest{Field: "decision_id", Detail: "is required"}
	}
	if req.ActorID == uuid.Nil {
		return &domain.ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}
	}
	return nil
}

func mapOrderSideToContract(s vo.OrderSide) contract.ComplianceOrderSide {
	switch s {
	case vo.OrderSideBuy:
		return contract.ComplianceOrderSideBuy
	case vo.OrderSideSell:
		return contract.ComplianceOrderSideSell
	default:
		return contract.ComplianceOrderSide(string(s))
	}
}

// summarizeBreaches builds a single human-readable string from the first few
// breaches so a single-line error message conveys why a decision was blocked.
func summarizeBreaches(breaches []contract.ProposedOrderBreach) string {
	if len(breaches) == 0 {
		return "blocked by IRG (no breach detail reported)"
	}
	if len(breaches) == 1 {
		return fmt.Sprintf("%s: %s", breaches[0].RuleTypeID, breaches[0].Message)
	}
	return fmt.Sprintf(
		"%s: %s (and %d more breach(es))",
		breaches[0].RuleTypeID, breaches[0].Message, len(breaches)-1,
	)
}
