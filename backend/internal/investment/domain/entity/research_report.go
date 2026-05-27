package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// ResearchReport is the aggregate root for an analyst-authored research
// report. Mutability is gated by the soft-delete flag and the review status;
// see CanUpdate / CanDelete / CanSubmit / CanCancelSubmit for the rules.
type ResearchReport struct {
	ID                   uuid.UUID
	ReportNo             string
	ReportDate           time.Time
	EffectiveDate        *time.Time
	OwnerUserID          uuid.UUID
	AuthorUserID         uuid.UUID
	ApplicableContractID *uuid.UUID

	InstrumentType string
	InstrumentCode string
	InstrumentName string
	Market         string
	Currency       string

	Recommendation vo.Recommendation
	ReportTitle    string

	CompanyOverview    string
	CompanyOutlook     string
	ESGComment         string
	FinancialStatus    string
	InvestmentAnalysis string

	RejectionReason    string
	PostSubmissionNote string

	ReportStatus vo.ReportStatus
	ReviewStatus vo.ReviewStatus

	CreatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedAt time.Time
	UpdatedBy *uuid.UUID
	DeletedAt *time.Time
}

// IsDeleted reports whether the report has been soft-deleted.
func (r *ResearchReport) IsDeleted() bool {
	return r != nil && r.DeletedAt != nil
}

// CanUpdate returns true when the report may be modified.
//
// Rules:
//   - deleted reports cannot be updated
//   - reports whose review has already been completed cannot be updated
func (r *ResearchReport) CanUpdate() bool {
	if r == nil || r.IsDeleted() {
		return false
	}
	if r.ReviewStatus == vo.ReviewStatusReviewCompleted {
		return false
	}
	return true
}

// CanDelete returns true when the report may be soft-deleted.
//
// Rules:
//   - deleted reports cannot be deleted again
//   - submitted or reviewed reports cannot be deleted (only NOT_SUBMITTED)
func (r *ResearchReport) CanDelete() bool {
	if r == nil || r.IsDeleted() {
		return false
	}
	return r.ReviewStatus == vo.ReviewStatusNotSubmitted
}

// CanSubmit returns true when the report may be moved from NOT_SUBMITTED to
// SUBMITTED. Deleted reports cannot be submitted.
func (r *ResearchReport) CanSubmit() bool {
	if r == nil || r.IsDeleted() {
		return false
	}
	return r.ReviewStatus == vo.ReviewStatusNotSubmitted
}

// CanCancelSubmit returns true when the report may be returned from SUBMITTED
// back to NOT_SUBMITTED. Once the review has been completed, cancelling is no
// longer permitted.
func (r *ResearchReport) CanCancelSubmit() bool {
	if r == nil || r.IsDeleted() {
		return false
	}
	return r.ReviewStatus == vo.ReviewStatusSubmitted
}
