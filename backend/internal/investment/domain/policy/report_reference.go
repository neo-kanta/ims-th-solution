package policy

import (
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// ReportReferenceViolation enumerates the reasons a research report cannot be
// referenced by an investment decision. Stable strings — included in audit
// records and surfaced to the API as ErrDecisionReferenceInvalid.Reason.
type ReportReferenceViolation string

const (
	ViolationReportMissing             ReportReferenceViolation = "report_missing"
	ViolationReportDeleted             ReportReferenceViolation = "report_deleted"
	ViolationReportInvalidated         ReportReferenceViolation = "report_invalidated"
	ViolationReportNotApproved         ReportReferenceViolation = "report_not_approved"
	ViolationReportRejected            ReportReferenceViolation = "report_rejected"
	ViolationReportExpired             ReportReferenceViolation = "report_expired"
	ViolationReportContractMismatch    ReportReferenceViolation = "report_contract_mismatch"
	ViolationReportRecommendationMismatch ReportReferenceViolation = "report_recommendation_mismatch"
)

// ReportReferenceInput is the call-site payload for CanReferenceResearchReport.
// The Report pointer may be nil — the policy returns ViolationReportMissing.
//
// ContractID is the decision's contract scope; the policy enforces that the
// report's ApplicableContractID, when set, matches. A nil ApplicableContractID
// on the report means it is a company-wide report and is acceptable for any
// fund/contract in scope of the caller (further data-permission checks happen
// at the application/repository layer, not the policy).
type ReportReferenceInput struct {
	Report     *entity.ResearchReport
	ContractID uuid.UUID
	Side       vo.OrderSide
	BusinessDate time.Time
}

// CanReferenceResearchReport applies the reference-rule policy and returns
// the first failing rule. Returns nil when the reference is acceptable.
//
// Rules enforced (configuration-driven extensions are deliberately TODO):
//
//   * Report must exist (non-nil) and not be soft-deleted.
//   * Report must be ACTIVE (review completed → approved by approval engine).
//     Reports in DRAFT/EXPIRED/REJECTED state are unacceptable.
//   * Report must be valid on the business date — past effective_date when set.
//   * Report's contract scope (when set) must match the decision contract.
//   * Recommendation must match the proposed side (BUY decision needs a
//     BUY/HOLD report; SELL decision needs a SELL/HOLD report). The HOLD
//     recommendation is permissive because it does not advocate direction.
//
// Future configurable rules (latest-report-only, own-report-only,
// company-latest-direction, dedicated-contract) are intentionally NOT
// hardcoded into the handler today — they belong on a configuration table
// not yet present. Add them here when the configuration ships rather than
// at the call site.
func CanReferenceResearchReport(in ReportReferenceInput) ReportReferenceViolation {
	r := in.Report
	if r == nil {
		return ViolationReportMissing
	}
	if r.IsDeleted() {
		return ViolationReportDeleted
	}
	// Invalidated reports are terminal — refuse before checking any other rule
	// so the caller sees a precise "this report was invalidated" message.
	if r.IsInvalidated() {
		return ViolationReportInvalidated
	}
	switch r.ReportStatus {
	case vo.ReportStatusRejected:
		return ViolationReportRejected
	case vo.ReportStatusExpired:
		return ViolationReportExpired
	case vo.ReportStatusInvalidated:
		// Belt-and-braces: handled above via IsInvalidated, kept here for
		// future readers tracing the switch coverage.
		return ViolationReportInvalidated
	case vo.ReportStatusActive:
		// fall through to the rest of the checks
	default:
		// DRAFT (or any unknown future value) is not approved.
		return ViolationReportNotApproved
	}
	if r.ReviewStatus != vo.ReviewStatusReviewCompleted {
		// Defence in depth: ACTIVE without review completed should not occur
		// in practice, but the approval engine is the only legitimate path to
		// flip status, so we double-check.
		return ViolationReportNotApproved
	}
	if r.EffectiveDate != nil && !in.BusinessDate.IsZero() {
		if in.BusinessDate.Before(*r.EffectiveDate) {
			// effective date hasn't started yet
			return ViolationReportExpired
		}
	}
	if r.ApplicableContractID != nil && in.ContractID != uuid.Nil {
		if *r.ApplicableContractID != in.ContractID {
			return ViolationReportContractMismatch
		}
	}
	if in.Side != "" && r.Recommendation != "" && r.Recommendation != vo.RecommendationHold {
		if in.Side == vo.OrderSideBuy && r.Recommendation == vo.RecommendationSell {
			return ViolationReportRecommendationMismatch
		}
		if in.Side == vo.OrderSideSell && r.Recommendation == vo.RecommendationBuy {
			return ViolationReportRecommendationMismatch
		}
	}
	return ""
}
