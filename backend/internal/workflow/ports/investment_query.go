package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TransactionSummary is the result of querying the investment module for
// trading activity on a specific contract + business date.
type TransactionSummary struct {
	// Count is the total number of investment transactions for the date.
	Count int

	// HasPendingUnreviewed is true when at least one transaction has not yet
	// been reviewed by the fund manager.
	HasPendingUnreviewed bool
}

// InvestmentQueryPort is implemented by the investment module and consumed by
// the workflow module's approval command handler. It answers the question:
// "what is the transaction state for this contract on this date?"
//
// In Batch 1 the NopInvestmentQueryAdapter always returns Count=0,
// HasPendingUnreviewed=false, which exercises the zero-transaction approval path.
type InvestmentQueryPort interface {
	GetTransactionSummary(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (TransactionSummary, error)
}
