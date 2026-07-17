package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// BreachStatus tracks the lifecycle of a compliance breach.
type BreachStatus string

const (
	BreachStatusOpen       BreachStatus = "OPEN"
	BreachStatusOverridden BreachStatus = "OVERRIDDEN"
	BreachStatusResolved   BreachStatus = "RESOLVED" // auto-closed on subsequent pass
)

// Breach is a persisted violation record created when a rule returns BLOCK or WARN.
type Breach struct {
	ID             uuid.UUID
	CheckRecordID  uuid.UUID
	CheckGroupID   uuid.UUID
	PortfolioID    uuid.UUID
	ContractID     *uuid.UUID // nil for portfolio-only checks (no fund_id)
	RuleTypeID     string
	RuleInstanceID uuid.UUID
	Severity       vo.Severity  `swaggertype:"string"`
	Verdict        vo.Verdict   `swaggertype:"string"`
	Status         BreachStatus `swaggertype:"string"`
	Evidence       vo.Evidence  `swaggertype:"object"`
	Message        string
	BusinessDate   time.Time
	CreatedAt      time.Time
	ResolvedAt     *time.Time
	ResolvedBy     *uuid.UUID
}
