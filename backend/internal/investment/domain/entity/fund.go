package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// Fund is the legal vehicle ("contract") that owns one or more portfolios.
// Fund.ID is the cross-module contract_id used by workflow / compliance /
// permissions modules — there is no separate contract table.
type Fund struct {
	ID             uuid.UUID
	Code           string
	Name           string
	ShortName      string
	FundCategoryID uuid.UUID
	BaseCurrency   string
	InceptionDate  time.Time
	ManagerUserID  *uuid.UUID
	Benchmark      string
	RiskProfile    vo.RiskProfile
	HasUnits       bool
	// RequirePretradePreview gates the Operation-tab UX. When true the trade
	// ticket must run a pre-trade simulation and surface per-rule verdicts
	// before allowing the post. The backend's post handler still enforces
	// the pre-trade rules independently — this flag only affects the UI flow.
	RequirePretradePreview bool
	// RequireResearchReportForDecision enforces that a research report must be
	// linked before a decision can be submitted for approval. This is a backend
	// policy gate independent of the pre-trade simulation UX flag above.
	RequireResearchReportForDecision bool
	ExternalPAMRef                   string
	Status                 vo.FundStatus
	Version                int

	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedAt *time.Time
}

// IsActive reports whether the fund is operational and not soft-deleted.
func (f *Fund) IsActive() bool {
	return f != nil && f.DeletedAt == nil && f.Status == vo.FundStatusActive
}
