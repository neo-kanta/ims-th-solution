package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// CheckRecord is an immutable audit record of one rule evaluation.
// One pre-trade check request produces N CheckRecords (one per applicable rule).
// These are NEVER updated or deleted. Append-only. Enforced at DB level.
type CheckRecord struct {
	ID                  uuid.UUID
	CheckGroupID        uuid.UUID      // groups all records from a single check request
	Timing              vo.CheckTiming `swaggertype:"string"`
	OrderID             *uuid.UUID     // nil for periodic checks
	PortfolioID         uuid.UUID
	ContractID          *uuid.UUID // nil for portfolio-only checks (no fund_id)
	Ticker              string     // empty for portfolio-wide periodic checks
	RuleTypeID          string
	RuleInstanceID      uuid.UUID
	RuleInstanceVersion int
	ParameterSnapshot   json.RawMessage `swaggertype:"object"` // frozen copy of params used
	Verdict             vo.Verdict      `swaggertype:"string"` // raw verdict from rule
	EffectiveSeverity   vo.Severity     `swaggertype:"string"` // binding-level severity applied
	FinalVerdict        vo.Verdict      `swaggertype:"string"` // after severity cap
	Evidence            json.RawMessage `swaggertype:"object"` // structured evidence
	Message             string
	DataSnapshotHash    string // SHA-256 of the DataBundle used
	EvalDurationMs      int64
	CheckedBy           string // actor ID or "system"
	BusinessDate        time.Time
	CheckedAt           time.Time
	CreatedAt           time.Time
}
