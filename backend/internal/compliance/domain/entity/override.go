package entity

import (
	"time"

	"github.com/google/uuid"
)

// Override is an audited record of a compliance officer overriding a BLOCK verdict.
// Immutable once created. Never updated or deleted.
type Override struct {
	ID            uuid.UUID
	BreachID      uuid.UUID
	Reason        string     // mandatory — justification
	OverriddenBy  uuid.UUID  // compliance officer with IRG_OVERRIDE_BREACH permission
	DelegatedFrom *uuid.UUID // original officer if delegated
	ApprovedBy    *uuid.UUID // second-level approval if required
	CreatedAt     time.Time
}
