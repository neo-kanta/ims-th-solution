package valueobject_test

import (
	"testing"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

func TestPortfolioStatus_IsOpenForBusiness(t *testing.T) {
	cases := []struct {
		status   vo.PortfolioStatus
		wantOpen bool
	}{
		{vo.PortfolioStatusActive, true},
		{vo.PortfolioStatusDraft, false},
		{vo.PortfolioStatusPendingApproval, false},
		{vo.PortfolioStatusRejected, false},
		{vo.PortfolioStatusSuspended, false},
		{vo.PortfolioStatusPaused, false},
		{vo.PortfolioStatusClosed, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.status), func(t *testing.T) {
			t.Parallel()
			got := tc.status.IsOpenForBusiness()
			if got != tc.wantOpen {
				t.Errorf("IsOpenForBusiness(%q) = %v, want %v", tc.status, got, tc.wantOpen)
			}
		})
	}
}

func TestPortfolioStatus_IsValid(t *testing.T) {
	valid := []vo.PortfolioStatus{
		vo.PortfolioStatusDraft,
		vo.PortfolioStatusPendingApproval,
		vo.PortfolioStatusActive,
		vo.PortfolioStatusPaused,
		vo.PortfolioStatusSuspended,
		vo.PortfolioStatusRejected,
		vo.PortfolioStatusClosed,
	}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("IsValid(%q) = false, want true", s)
		}
	}
	if vo.PortfolioStatus("BOGUS").IsValid() {
		t.Error("IsValid(BOGUS) = true, want false")
	}
}
