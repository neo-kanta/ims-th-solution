package valueobject

// Recommendation is the analyst recommendation on the researched instrument.
type Recommendation string

const (
	RecommendationBuy  Recommendation = "BUY"
	RecommendationSell Recommendation = "SELL"
	RecommendationHold Recommendation = "HOLD"
)

// IsValid reports whether the recommendation string is one of the accepted
// values.
func (r Recommendation) IsValid() bool {
	switch r {
	case RecommendationBuy, RecommendationSell, RecommendationHold:
		return true
	default:
		return false
	}
}

// ReportStatus is the lifecycle status of a research report itself, separate
// from its review status.
type ReportStatus string

const (
	ReportStatusDraft       ReportStatus = "DRAFT"
	ReportStatusActive      ReportStatus = "ACTIVE"
	ReportStatusExpired     ReportStatus = "EXPIRED"
	ReportStatusRejected    ReportStatus = "REJECTED"
	ReportStatusInvalidated ReportStatus = "INVALIDATED"
)

// IsValid reports whether the status string is one of the accepted values.
func (s ReportStatus) IsValid() bool {
	switch s {
	case ReportStatusDraft, ReportStatusActive, ReportStatusExpired,
		ReportStatusRejected, ReportStatusInvalidated:
		return true
	default:
		return false
	}
}

// ReviewStatus tracks where the report sits in the (simple, PoC-scoped)
// submit → review-completed pipeline.
type ReviewStatus string

const (
	ReviewStatusNotSubmitted    ReviewStatus = "NOT_SUBMITTED"
	ReviewStatusSubmitted       ReviewStatus = "SUBMITTED"
	ReviewStatusReviewCompleted ReviewStatus = "REVIEW_COMPLETED"
	// ReviewStatusInvalidated mirrors ReportStatusInvalidated. The two flip
	// together (DB CHECK enforces) so a single front-end "INVALIDATED" badge
	// is unambiguous regardless of which column the UI consults.
	ReviewStatusInvalidated ReviewStatus = "INVALIDATED"
)

// IsValid reports whether the status string is one of the accepted values.
func (s ReviewStatus) IsValid() bool {
	switch s {
	case ReviewStatusNotSubmitted, ReviewStatusSubmitted,
		ReviewStatusReviewCompleted, ReviewStatusInvalidated:
		return true
	default:
		return false
	}
}
