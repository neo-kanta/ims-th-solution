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

	// Invalidation metadata. Populated together when the report is invalidated;
	// the database CHECK constraint chk_inv_research_invalidation_coherent
	// rejects partial population.
	InvalidatedAt      *time.Time
	InvalidatedBy      *uuid.UUID
	InvalidationReason string

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

// IsInvalidated reports whether the report has been invalidated.
// The DB CHECK keeps ReportStatus and ReviewStatus aligned, so either
// column is authoritative.
func (r *ResearchReport) IsInvalidated() bool {
	if r == nil {
		return false
	}
	return r.ReportStatus == vo.ReportStatusInvalidated ||
		r.ReviewStatus == vo.ReviewStatusInvalidated
}

// CanInvalidate returns true when the report may be moved to INVALIDATED.
// Rules:
//   - deleted reports cannot be invalidated (the row is gone from the demo view).
//   - already-invalidated reports cannot be invalidated again.
//   - reports may be invalidated from any other lifecycle state, including
//     ACTIVE / REVIEW_COMPLETED — invalidation is the safety hatch for a
//     report that turned out to be wrong after approval.
func (r *ResearchReport) CanInvalidate() bool {
	if r == nil || r.IsDeleted() {
		return false
	}
	return !r.IsInvalidated()
}

// CanUpdate returns true when the report may be modified.
//
// Rules:
//   - deleted reports cannot be updated
//   - invalidated reports cannot be updated
//   - reports in SUBMITTED or REVIEW_COMPLETED state cannot be updated
//     (submitted reports are locked while their approval is in-flight)
func (r *ResearchReport) CanUpdate() bool {
	if r == nil || r.IsDeleted() || r.IsInvalidated() {
		return false
	}
	if r.ReviewStatus == vo.ReviewStatusSubmitted || r.ReviewStatus == vo.ReviewStatusReviewCompleted {
		return false
	}
	return true
}

// CanDelete returns true when the report may be soft-deleted.
//
// Rules:
//   - deleted reports cannot be deleted again
//   - invalidated reports cannot be deleted (the invalidation record must
//     remain visible for the audit trail)
//   - submitted or reviewed reports cannot be deleted (only NOT_SUBMITTED)
func (r *ResearchReport) CanDelete() bool {
	if r == nil || r.IsDeleted() || r.IsInvalidated() {
		return false
	}
	return r.ReviewStatus == vo.ReviewStatusNotSubmitted
}

// CanSubmit returns true when the report may be moved from NOT_SUBMITTED to
// SUBMITTED. Deleted or invalidated reports cannot be submitted.
func (r *ResearchReport) CanSubmit() bool {
	if r == nil || r.IsDeleted() || r.IsInvalidated() {
		return false
	}
	return r.ReviewStatus == vo.ReviewStatusNotSubmitted
}

// CanCancelSubmit returns true when the report may be returned from SUBMITTED
// back to NOT_SUBMITTED. Once the review has been completed or the report has
// been invalidated, cancelling is no longer permitted.
func (r *ResearchReport) CanCancelSubmit() bool {
	if r == nil || r.IsDeleted() || r.IsInvalidated() {
		return false
	}
	return r.ReviewStatus == vo.ReviewStatusSubmitted
}
