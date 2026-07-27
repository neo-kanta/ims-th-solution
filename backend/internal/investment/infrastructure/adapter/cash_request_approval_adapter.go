package adapter

import (
	"context"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// CashRequestApprovalCallback implements contract.ApprovalSubjectCallback for
// the CASH_TRANSACTION subject type. When a LIVE cash movement's approval
// reaches a final decision the approval engine invokes this callback, which
// materializes the real ledger transaction (on approve) or marks the request
// rejected (on reject) via the shared post pipeline.
type CashRequestApprovalCallback struct {
	postTxn *command.PostTransactionHandler
}

// NewCashRequestApprovalCallback wires the callback to the post-transaction
// handler that owns the cash-request lifecycle.
func NewCashRequestApprovalCallback(postTxn *command.PostTransactionHandler) *CashRequestApprovalCallback {
	return &CashRequestApprovalCallback{postTxn: postTxn}
}

// OnApprovalDecision implements contract.ApprovalSubjectCallback.
func (c *CashRequestApprovalCallback) OnApprovalDecision(ctx context.Context, decision contract.ApprovalDecision) error {
	if c == nil || c.postTxn == nil {
		return nil
	}
	if decision.SubjectType != "CASH_TRANSACTION" {
		return nil
	}
	return c.postTxn.ApplyCashRequestApproval(ctx, decision.SubjectID, decision.Approved, decision.Reason)
}

var _ contract.ApprovalSubjectCallback = (*CashRequestApprovalCallback)(nil)
