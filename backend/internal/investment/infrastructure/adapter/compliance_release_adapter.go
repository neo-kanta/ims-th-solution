package adapter

import (
	"context"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ComplianceReleaseCallback implements contract.ApprovalSubjectCallback for the
// COMPLIANCE_RELEASE subject type. When the approval engine reaches a final
// decision on a compliance-release request, this callback notifies the
// investment module so it can either cancel the decision (on rejection) or
// re-submit it to the standard INVESTMENT_DECISION approval engine (on approval).
type ComplianceReleaseCallback struct {
	decisions *command.DecisionCommandHandler
}

// NewComplianceReleaseCallback wires the callback.
func NewComplianceReleaseCallback(decisions *command.DecisionCommandHandler) *ComplianceReleaseCallback {
	return &ComplianceReleaseCallback{decisions: decisions}
}

// OnApprovalDecision is called by the approval engine when a COMPLIANCE_RELEASE
// approval request reaches a terminal state.
func (c *ComplianceReleaseCallback) OnApprovalDecision(ctx context.Context, decision contract.ApprovalDecision) error {
	if c == nil || c.decisions == nil {
		return nil
	}
	if decision.SubjectType != "COMPLIANCE_RELEASE" {
		return nil
	}
	return c.decisions.ApplyComplianceReleaseDecision(ctx, decision.SubjectID, decision.Approved, decision.Reason)
}

var _ contract.ApprovalSubjectCallback = (*ComplianceReleaseCallback)(nil)
