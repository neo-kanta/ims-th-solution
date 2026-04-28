package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
)

// NopInvestmentQueryAdapter satisfies ports.InvestmentQueryPort by always
// returning zero transactions with no pending unreviewed items.
//
// Consequence in Batch 1: every APPROVE request hits the zero-transaction path.
// Approvals will succeed only when the caller provides zeroTransactionAttestation=true
// and a reason of at least 30 characters — which is the correct business behaviour
// for a zero-transaction day.
//
// Production replacement: wire the real investment module query here once
// the investment module implements GetTransactionSummary.
type NopInvestmentQueryAdapter struct{}

func (a *NopInvestmentQueryAdapter) GetTransactionSummary(
	_ context.Context,
	_ uuid.UUID,
	_ time.Time,
) (ports.TransactionSummary, error) {
	return ports.TransactionSummary{Count: 0, HasPendingUnreviewed: false}, nil
}

// compile-time interface check
var _ ports.InvestmentQueryPort = (*NopInvestmentQueryAdapter)(nil)
