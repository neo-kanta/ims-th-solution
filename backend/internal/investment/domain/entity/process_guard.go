package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// WorkflowDaySetting is the working-window configuration used by the
// investment process guard. It is read from workflow__day_settings but kept as
// an investment-side read model to preserve module boundaries at the Go layer.
type WorkflowDaySetting struct {
	ID                      uuid.UUID
	Name                    string
	Timezone                string
	WorkStartTime           time.Duration
	WorkEndTime             time.Duration
	RequiresManagerApproval bool
	AllowHighLevelOverride  bool
	BlockOnRejection        bool
	SkipNonBusinessDays     bool
	EffectiveFrom           time.Time
	EffectiveTo             *time.Time
}

// ProcessAssignmentMatch describes the assignment that authorizes a user for a
// process step. AssignmentType is GROUP or USER.
type ProcessAssignmentMatch struct {
	AssignmentID   uuid.UUID
	AssignmentType string
	ProcessStep    vo.ProcessStepKey
	GroupID        *uuid.UUID
	GroupName      string
	UserID         uuid.UUID
	ScopeType      string
	ScopeID        *uuid.UUID
	CanExecute     bool
	Priority       int
}

// BlockingControlDecision describes an active rejection that blocks investment
// work for the day or process step.
type BlockingControlDecision struct {
	ID             uuid.UUID
	WorkflowDayID  *uuid.UUID
	ContractID     uuid.UUID
	BusinessDate   time.Time
	DecisionScope  string
	ProcessStep    *vo.ProcessStepKey
	DecisionLevel  string
	DecisionStatus string
	Reason         string
	DecidedBy      *uuid.UUID
	DecidedAt      time.Time
}
