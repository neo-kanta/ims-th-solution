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
	ExternalPAMRef string
	Status         vo.FundStatus
	Version        int

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
