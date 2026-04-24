package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// RuleBinding maps a rule instance to a scope with a severity override.
// Aggregate root — owns its own lifecycle, effective dating.
//
// Conflict resolution: when two bindings for the same rule type hit the same
// portfolio, most-specific-scope wins (portfolio > contract > fund_category > global).
// Equal specificity: Priority field breaks the tie (lower = higher priority).
type RuleBinding struct {
	ID              uuid.UUID
	RuleInstanceID  uuid.UUID
	RuleSetID       *uuid.UUID // nil if directly bound, not via a rule set
	Scope           vo.Scope
	Severity        vo.Severity // overrides rule type's default severity
	Priority        int         // lower number = higher priority
	EffectiveWindow vo.EffectiveWindow
	IsActive        bool
	CreatedBy       uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
