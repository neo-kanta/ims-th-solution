package adapter

import (
	"context"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// PortfolioApprovalCallback implements contract.ApprovalSubjectCallback for the
// PORTFOLIO subject type. Invoked by the approval module when a
// PORTFOLIO_ONBOARDING request reaches a final decision: approved → ACTIVE,
// rejected → REJECTED.
type PortfolioApprovalCallback struct {
	portfolios *command.PortfolioCommandHandler
}

// NewPortfolioApprovalCallback wires the callback to the portfolio command handler.
func NewPortfolioApprovalCallback(portfolios *command.PortfolioCommandHandler) *PortfolioApprovalCallback {
	return &PortfolioApprovalCallback{portfolios: portfolios}
}

// OnApprovalDecision implements contract.ApprovalSubjectCallback.
func (c *PortfolioApprovalCallback) OnApprovalDecision(ctx context.Context, decision contract.ApprovalDecision) error {
	if c == nil || c.portfolios == nil {
		return nil
	}
	if decision.SubjectType != "PORTFOLIO" {
		return nil
	}
	return c.portfolios.ApplyApprovalDecision(ctx, decision.SubjectID, decision.Approved, decision.Reason)
}

var _ contract.ApprovalSubjectCallback = (*PortfolioApprovalCallback)(nil)
