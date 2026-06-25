package adapter

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ConfirmationGateAdapter implements contract.TradeConfirmationGate by walking
// the investment module's executions + confirmations for the given contract
// and business date.
//
// Blocking conditions:
//   * any execution for the day in PENDING or PARTIALLY_EXECUTED state without
//     a MATCHED/REVIEWED confirmation,
//   * any confirmation in PENDING_REVIEW or MISMATCHED-without-reason state.
//
// Cancelled executions are skipped — they aren't on the books.
type ConfirmationGateAdapter struct {
	executions    domain.ExecutionRepository
	confirmations domain.TradeConfirmationRepository
}

func NewConfirmationGateAdapter(
	executions domain.ExecutionRepository,
	confirmations domain.TradeConfirmationRepository,
) *ConfirmationGateAdapter {
	return &ConfirmationGateAdapter{
		executions:    executions,
		confirmations: confirmations,
	}
}

func (a *ConfirmationGateAdapter) EvaluateClose(
	ctx context.Context,
	contractID uuid.UUID,
	businessDate time.Time,
) (*contract.ConfirmationGateResult, error) {
	if a == nil || a.executions == nil || a.confirmations == nil {
		return &contract.ConfirmationGateResult{}, nil
	}
	out := &contract.ConfirmationGateResult{}

	// Load executions for the (contract, business_date) tuple.
	execs, err := a.executions.ListByContractDate(ctx, contractID, businessDate)
	if err != nil {
		return nil, err
	}
	// Load confirmations once and key by execution_id.
	confs, err := a.confirmations.ListByContractDate(ctx, contractID, businessDate)
	if err != nil {
		return nil, err
	}
	resolvedExec := map[uuid.UUID]bool{}
	for _, c := range confs {
		switch c.Status {
		case vo.TradeConfirmationPendingReview:
			out.PendingCount++
			out.Reasons = append(out.Reasons, "confirmation pending review: "+c.ID.String())
		case vo.TradeConfirmationMismatched:
			if strings.TrimSpace(c.DiscrepancyReason) == "" {
				out.UnresolvedCount++
				out.Reasons = append(out.Reasons, "mismatched confirmation without reason: "+c.ID.String())
			} else {
				resolvedExec[c.ExecutionID] = true
			}
		case vo.TradeConfirmationMatched, vo.TradeConfirmationReviewed:
			resolvedExec[c.ExecutionID] = true
		}
	}
	for _, e := range execs {
		if e.Status == vo.ExecutionStatusCancelled {
			continue
		}
		if !resolvedExec[e.ID] {
			out.PendingCount++
			out.Reasons = append(out.Reasons, "execution missing confirmation: "+e.ID.String())
		}
	}
	return out, nil
}

var _ contract.TradeConfirmationGate = (*ConfirmationGateAdapter)(nil)
