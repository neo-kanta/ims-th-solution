package policy

import (
	"strings"
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// MinInvestmentAnalysisLength is the minimum character length required on the
// investment_analysis field. Picked to discourage one-line drafts.
const MinInvestmentAnalysisLength = 25

// ResearchReportInput is the shape the policy validates. Keeping it
// transport-free lets command handlers and tests share the same call.
type ResearchReportInput struct {
	ReportNo           string
	ReportDate         time.Time
	OwnerUserID        uuid.UUID
	AuthorUserID       uuid.UUID
	InstrumentCode     string
	Recommendation     vo.Recommendation
	InvestmentAnalysis string
}

// ResearchReportValidationError describes a single failing field.
type ResearchReportValidationError struct {
	Field  string
	Detail string
}

// ValidateResearchReport runs the basic required-field rules and returns the
// first failing constraint as a typed error, or nil when the input is valid.
//
// Trailing whitespace on string fields is ignored (Go strings can carry
// trimmable padding from frontends), but the policy itself does not mutate
// the input — callers should normalise the strings before persisting.
func ValidateResearchReport(in ResearchReportInput) *ResearchReportValidationError {
	if in.ReportDate.IsZero() {
		return &ResearchReportValidationError{Field: "report_date", Detail: "is required"}
	}
	if in.OwnerUserID == uuid.Nil {
		return &ResearchReportValidationError{Field: "owner_user_id", Detail: "is required"}
	}
	if in.AuthorUserID == uuid.Nil {
		return &ResearchReportValidationError{Field: "author_user_id", Detail: "is required"}
	}
	if strings.TrimSpace(in.InstrumentCode) == "" {
		return &ResearchReportValidationError{Field: "instrument_code", Detail: "is required"}
	}
	if !in.Recommendation.IsValid() {
		return &ResearchReportValidationError{Field: "recommendation", Detail: "must be BUY, SELL, or HOLD"}
	}
	analysis := strings.TrimSpace(in.InvestmentAnalysis)
	if analysis == "" {
		return &ResearchReportValidationError{Field: "investment_analysis", Detail: "is required"}
	}
	if len(analysis) < MinInvestmentAnalysisLength {
		return &ResearchReportValidationError{
			Field:  "investment_analysis",
			Detail: "must be at least 25 characters",
		}
	}
	return nil
}
