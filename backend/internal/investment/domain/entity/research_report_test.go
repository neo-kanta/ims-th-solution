package entity_test

import (
	"testing"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

func draftReport() *entity.ResearchReport {
	return &entity.ResearchReport{
		ReportStatus: vo.ReportStatusDraft,
		ReviewStatus: vo.ReviewStatusNotSubmitted,
	}
}

func TestResearchReport_CanUpdate(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(r *entity.ResearchReport)
		expect bool
	}{
		{"draft not_submitted", func(r *entity.ResearchReport) {}, true},
		{"draft submitted", func(r *entity.ResearchReport) { r.ReviewStatus = vo.ReviewStatusSubmitted }, true},
		{
			"review completed locks update",
			func(r *entity.ResearchReport) { r.ReviewStatus = vo.ReviewStatusReviewCompleted },
			false,
		},
		{
			"deleted blocks update",
			func(r *entity.ResearchReport) { d := time.Now(); r.DeletedAt = &d },
			false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := draftReport()
			tc.setup(r)
			if got := r.CanUpdate(); got != tc.expect {
				t.Fatalf("CanUpdate: want %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestResearchReport_CanDelete(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(r *entity.ResearchReport)
		expect bool
	}{
		{"draft not_submitted", func(r *entity.ResearchReport) {}, true},
		{
			"submitted blocks delete",
			func(r *entity.ResearchReport) { r.ReviewStatus = vo.ReviewStatusSubmitted },
			false,
		},
		{
			"review completed blocks delete",
			func(r *entity.ResearchReport) { r.ReviewStatus = vo.ReviewStatusReviewCompleted },
			false,
		},
		{
			"deleted blocks delete",
			func(r *entity.ResearchReport) { d := time.Now(); r.DeletedAt = &d },
			false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := draftReport()
			tc.setup(r)
			if got := r.CanDelete(); got != tc.expect {
				t.Fatalf("CanDelete: want %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestResearchReport_CanSubmit(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(r *entity.ResearchReport)
		expect bool
	}{
		{"not_submitted is submittable", func(r *entity.ResearchReport) {}, true},
		{
			"submitted is not re-submittable",
			func(r *entity.ResearchReport) { r.ReviewStatus = vo.ReviewStatusSubmitted },
			false,
		},
		{
			"review completed is not submittable",
			func(r *entity.ResearchReport) { r.ReviewStatus = vo.ReviewStatusReviewCompleted },
			false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := draftReport()
			tc.setup(r)
			if got := r.CanSubmit(); got != tc.expect {
				t.Fatalf("CanSubmit: want %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestResearchReport_CanCancelSubmit(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(r *entity.ResearchReport)
		expect bool
	}{
		{
			"not_submitted is not cancellable",
			func(r *entity.ResearchReport) {},
			false,
		},
		{
			"submitted is cancellable",
			func(r *entity.ResearchReport) { r.ReviewStatus = vo.ReviewStatusSubmitted },
			true,
		},
		{
			"review completed is not cancellable",
			func(r *entity.ResearchReport) { r.ReviewStatus = vo.ReviewStatusReviewCompleted },
			false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := draftReport()
			tc.setup(r)
			if got := r.CanCancelSubmit(); got != tc.expect {
				t.Fatalf("CanCancelSubmit: want %v, got %v", tc.expect, got)
			}
		})
	}
}
