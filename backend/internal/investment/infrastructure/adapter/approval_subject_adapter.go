package adapter

import (
	"context"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ResearchApprovalCallback implements contract.ApprovalSubjectCallback for the
// RESEARCH_REPORT subject type. The approval module invokes it when a research
// report's approval request reaches a final decision, and it drives the
// report's lifecycle accordingly.
type ResearchApprovalCallback struct {
	research *command.ResearchReportCommandHandler
}

// NewResearchApprovalCallback wires the callback to the research command handler.
func NewResearchApprovalCallback(research *command.ResearchReportCommandHandler) *ResearchApprovalCallback {
	return &ResearchApprovalCallback{research: research}
}

// OnApprovalDecision implements contract.ApprovalSubjectCallback.
func (c *ResearchApprovalCallback) OnApprovalDecision(ctx context.Context, decision contract.ApprovalDecision) error {
	if c == nil || c.research == nil {
		return nil
	}
	if decision.SubjectType != "RESEARCH_REPORT" {
		return nil
	}
	return c.research.ApplyApprovalDecision(ctx, decision.SubjectID, decision.Approved, decision.Reason)
}

var _ contract.ApprovalSubjectCallback = (*ResearchApprovalCallback)(nil)
