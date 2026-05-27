package policy_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

func validInput() policy.ResearchReportInput {
	return policy.ResearchReportInput{
		ReportNo:           "RR-2026-0001",
		ReportDate:         time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
		OwnerUserID:        uuid.New(),
		AuthorUserID:       uuid.New(),
		InstrumentCode:     "PTT",
		Recommendation:     vo.RecommendationBuy,
		InvestmentAnalysis: "This is a comprehensive analysis of the issuer over twenty-five chars.",
	}
}

func TestValidateResearchReport_HappyPath(t *testing.T) {
	if err := policy.ValidateResearchReport(validInput()); err != nil {
		t.Fatalf("expected nil error, got %+v", err)
	}
}

func TestValidateResearchReport_MissingFields(t *testing.T) {
	cases := []struct {
		name      string
		mutate    func(in *policy.ResearchReportInput)
		wantField string
	}{
		{
			name:      "zero report_date",
			mutate:    func(in *policy.ResearchReportInput) { in.ReportDate = time.Time{} },
			wantField: "report_date",
		},
		{
			name:      "zero owner_user_id",
			mutate:    func(in *policy.ResearchReportInput) { in.OwnerUserID = uuid.Nil },
			wantField: "owner_user_id",
		},
		{
			name:      "zero author_user_id",
			mutate:    func(in *policy.ResearchReportInput) { in.AuthorUserID = uuid.Nil },
			wantField: "author_user_id",
		},
		{
			name:      "blank instrument_code",
			mutate:    func(in *policy.ResearchReportInput) { in.InstrumentCode = "   " },
			wantField: "instrument_code",
		},
		{
			name:      "invalid recommendation",
			mutate:    func(in *policy.ResearchReportInput) { in.Recommendation = vo.Recommendation("STRONG_BUY") },
			wantField: "recommendation",
		},
		{
			name:      "blank investment_analysis",
			mutate:    func(in *policy.ResearchReportInput) { in.InvestmentAnalysis = "   \t  " },
			wantField: "investment_analysis",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validInput()
			tc.mutate(&in)
			err := policy.ValidateResearchReport(in)
			if err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if err.Field != tc.wantField {
				t.Fatalf("wrong field: want %q, got %q (detail=%q)", tc.wantField, err.Field, err.Detail)
			}
		})
	}
}

func TestValidateResearchReport_AnalysisMinLength(t *testing.T) {
	short := "Too short."
	in := validInput()
	in.InvestmentAnalysis = short
	err := policy.ValidateResearchReport(in)
	if err == nil {
		t.Fatal("expected min-length validation error, got nil")
	}
	if err.Field != "investment_analysis" {
		t.Fatalf("wrong field: want investment_analysis, got %q", err.Field)
	}
	if !strings.Contains(err.Detail, "25") {
		t.Fatalf("expected detail to mention min length, got %q", err.Detail)
	}
}

func TestValidateResearchReport_AnalysisExactBoundary(t *testing.T) {
	in := validInput()
	in.InvestmentAnalysis = strings.Repeat("a", policy.MinInvestmentAnalysisLength)
	if err := policy.ValidateResearchReport(in); err != nil {
		t.Fatalf("expected exact-boundary length to pass, got %+v", err)
	}
}
