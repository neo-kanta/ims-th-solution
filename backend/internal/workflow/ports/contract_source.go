package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ContractSourcePort supplies the contracts that the workflow scheduler should
// evaluate for a business date.
//
// The current PoC adapter reads known workflow contracts from workflow tables.
// A production adapter should come from the contract/fund master module.
type ContractSourcePort interface {
	ListActiveContracts(ctx context.Context, businessDate time.Time) ([]uuid.UUID, error)
}
