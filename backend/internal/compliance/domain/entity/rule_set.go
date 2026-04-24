package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// RuleSet is a named bundle of rule instances that can be bound as a unit.
// E.g. "AIMC Equity Mandate Baseline" containing 15 standard rules.
type RuleSet struct {
	ID          uuid.UUID
	Name        string // unique, human-readable
	Description string
	Members     []RuleSetMember
	IsActive    bool
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RuleSetMember is a rule instance within a set, with set-level overrides.
type RuleSetMember struct {
	RuleInstanceID uuid.UUID
	Severity       vo.Severity
	Priority       int
}
