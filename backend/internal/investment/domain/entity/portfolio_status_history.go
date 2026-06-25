package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PortfolioStatusHistory is an immutable append-only record of a single
// portfolio lifecycle transition. Backed by investment__portfolio_status_history.
type PortfolioStatusHistory struct {
	ID          uuid.UUID
	PortfolioID uuid.UUID
	FromStatus  *vo.PortfolioStatus // nil for the initial creation event
	ToStatus    vo.PortfolioStatus
	ActorID     *uuid.UUID // nil for system-initiated transitions (e.g. approval callback)
	Reason      *string
	CreatedAt   time.Time
}
