package policy

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

func approvedReport(t time.Time, side vo.Recommendation) *entity.ResearchReport {
	return &entity.ResearchReport{
		ID:             uuid.New(),
		ReportNo:       "RR-OK",
		ReportDate:     t,
		Recommendation: side,
		ReportStatus:   vo.ReportStatusActive,
		ReviewStatus:   vo.ReviewStatusReviewCompleted,
	}
}

func TestCanReferenceResearchReport_HappyPath(t *testing.T) {
	t.Parallel()
	bdate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	r := approvedReport(bdate, vo.RecommendationBuy)
	got := CanReferenceResearchReport(ReportReferenceInput{
		Report: r, ContractID: uuid.New(), Side: vo.OrderSideBuy, BusinessDate: bdate,
	})
	require.Equal(t, ReportReferenceViolation(""), got)
}

func TestCanReferenceResearchReport_MissingReport(t *testing.T) {
	t.Parallel()
	got := CanReferenceResearchReport(ReportReferenceInput{Report: nil})
	require.Equal(t, ViolationReportMissing, got)
}

func TestCanReferenceResearchReport_RejectedNotAcceptable(t *testing.T) {
	t.Parallel()
	r := approvedReport(time.Now(), vo.RecommendationBuy)
	r.ReportStatus = vo.ReportStatusRejected
	got := CanReferenceResearchReport(ReportReferenceInput{Report: r, Side: vo.OrderSideBuy})
	require.Equal(t, ViolationReportRejected, got)
}

func TestCanReferenceResearchReport_ExpiredNotAcceptable(t *testing.T) {
	t.Parallel()
	r := approvedReport(time.Now(), vo.RecommendationBuy)
	r.ReportStatus = vo.ReportStatusExpired
	got := CanReferenceResearchReport(ReportReferenceInput{Report: r, Side: vo.OrderSideBuy})
	require.Equal(t, ViolationReportExpired, got)
}

func TestCanReferenceResearchReport_DraftNotAcceptable(t *testing.T) {
	t.Parallel()
	r := approvedReport(time.Now(), vo.RecommendationBuy)
	r.ReportStatus = vo.ReportStatusDraft
	r.ReviewStatus = vo.ReviewStatusNotSubmitted
	got := CanReferenceResearchReport(ReportReferenceInput{Report: r, Side: vo.OrderSideBuy})
	require.Equal(t, ViolationReportNotApproved, got)
}

func TestCanReferenceResearchReport_EffectiveDateNotYet(t *testing.T) {
	t.Parallel()
	r := approvedReport(time.Now(), vo.RecommendationBuy)
	future := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	r.EffectiveDate = &future
	bd := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	got := CanReferenceResearchReport(ReportReferenceInput{Report: r, BusinessDate: bd, Side: vo.OrderSideBuy})
	require.Equal(t, ViolationReportExpired, got)
}

func TestCanReferenceResearchReport_WrongContract(t *testing.T) {
	t.Parallel()
	otherContract := uuid.New()
	r := approvedReport(time.Now(), vo.RecommendationBuy)
	r.ApplicableContractID = &otherContract
	got := CanReferenceResearchReport(ReportReferenceInput{
		Report: r, ContractID: uuid.New(), Side: vo.OrderSideBuy,
	})
	require.Equal(t, ViolationReportContractMismatch, got)
}

func TestCanReferenceResearchReport_OppositeDirection(t *testing.T) {
	t.Parallel()
	r := approvedReport(time.Now(), vo.RecommendationSell)
	got := CanReferenceResearchReport(ReportReferenceInput{
		Report: r, ContractID: uuid.New(), Side: vo.OrderSideBuy,
	})
	require.Equal(t, ViolationReportRecommendationMismatch, got)
}

func TestCanReferenceResearchReport_HoldIsPermissive(t *testing.T) {
	t.Parallel()
	r := approvedReport(time.Now(), vo.RecommendationHold)
	// HOLD: either direction is acceptable.
	require.Empty(t, CanReferenceResearchReport(ReportReferenceInput{
		Report: r, ContractID: uuid.New(), Side: vo.OrderSideBuy,
	}))
	require.Empty(t, CanReferenceResearchReport(ReportReferenceInput{
		Report: r, ContractID: uuid.New(), Side: vo.OrderSideSell,
	}))
}
