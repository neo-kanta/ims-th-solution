package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// RuleInstance is the aggregate root for a configured compliance rule.
// It references a rule type (by stable string ID from the SPI registry)
// and holds versioned parameters. Parameters are append-only: a change
// creates a new RuleInstanceVersion, the old version is never mutated.
type RuleInstance struct {
	ID              uuid.UUID
	RuleTypeID      string // stable SPI type ID, e.g. "concentration.single_issuer"
	Name            string // human-friendly name
	Description     string
	CurrentVersion  int // pointer to the active version number
	IsActive        bool
	EffectiveWindow vo.EffectiveWindow `swaggertype:"object"`
	CreatedBy       uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
