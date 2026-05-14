package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RuleInstanceVersion is an immutable snapshot of parameters at a point in time.
// Created on every parameter change. Never updated or deleted.
type RuleInstanceVersion struct {
	ID             uuid.UUID
	RuleInstanceID uuid.UUID
	VersionNumber  int
	Parameters     json.RawMessage `swaggertype:"object"` // validated against rule type schema at write time
	ChangeReason   string          // why this version was created
	ApprovedBy     *uuid.UUID      // nil if no approval required
	CreatedBy      uuid.UUID
	CreatedAt      time.Time
}
