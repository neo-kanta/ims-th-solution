package adapter

import (
	"context"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// DecisionApprovalCallback implements contract.ApprovalSubjectCallback for the
// INVESTMENT_DECISION subject type.
type DecisionApprovalCallback struct {
	decisions *command.DecisionCommandHandler
}

func NewDecisionApprovalCallback(decisions *command.DecisionCommandHandler) *DecisionApprovalCallback {
	return &DecisionApprovalCallback{decisions: decisions}
}

func (c *DecisionApprovalCallback) OnApprovalDecision(ctx context.Context, decision contract.ApprovalDecision) error {
	if c == nil || c.decisions == nil {
		return nil
	}
	if decision.SubjectType != "INVESTMENT_DECISION" {
		return nil
	}
	return c.decisions.ApplyApprovalDecision(ctx, decision.SubjectID, decision.Approved, decision.Reason)
}

var _ contract.ApprovalSubjectCallback = (*DecisionApprovalCallback)(nil)
