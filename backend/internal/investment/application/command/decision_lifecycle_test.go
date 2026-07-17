package command

// Tests for DecisionCommandHandler.Submit with the research-report reference
// rule. The policy itself is unit-tested in domain/policy/report_reference_test.go;
// these tests verify that the handler correctly threads the decision's
// ResearchReportID through the policy and returns ErrDecisionReferenceInvalid
// on a violation — i.e. the enforcement is wired, not just the rule.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// ── P1 Fix #4: BatchApprove/BatchReject — permission checked before disclosure ─

// fakeBatchPermissionChecker allows callers to specify per-contract access.
type fakeBatchPermissionChecker struct {
	allowed map[uuid.UUID]bool
}

func (f *fakeBatchPermissionChecker) HasDataPermission(_ context.Context, _ string, scopeID string) (bool, error) {
	id, err := uuid.Parse(scopeID)
	if err != nil {
		return false, nil
	}
	return f.allowed[id], nil
}

// fakeBatchActor records ApproveByRequest calls; errors can be injected via err.
type fakeBatchActor struct {
	err      error
	approved []uuid.UUID
}

func (f *fakeBatchActor) ApproveByRequest(_ context.Context, requestID uuid.UUID, _ uuid.UUID, _ string) error {
	if f.err != nil {
		return f.err
	}
	f.approved = append(f.approved, requestID)
	return nil
}
func (f *fakeBatchActor) RejectByRequest(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ string) error {
	return f.err
}

// TestBatchApprove_PermissionBeforeDisclosure verifies that:
//  1. A nonexistent decision number returns "decision not found".
//  2. An existing decision in an unauthorized contract also returns "decision
//     not found" — indistinguishable from nonexistent, preventing existence
//     inference by unauthorized callers.
//  3. An authorized decision proceeds to the approval actor.
func TestBatchApprove_PermissionBeforeDisclosure(t *testing.T) {
	authorizedCID := uuid.New()
	unauthorizedCID := uuid.New()
	approvalReqID := uuid.New()

	authorizedDecision := &entity.Decision{
		ID:                uuid.New(),
		FundID:            authorizedCID,
		DecisionNumber:    "DEC-ALLOWED",
		ApprovalRequestID: &approvalReqID,
		Status:            vo.DecisionLifecyclePendingApproval,
		BusinessDate:      time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC),
		CreatedBy:         uuid.New(),
		CreatedAt:         time.Now().UTC(),
	}
	unauthorizedDecision := &entity.Decision{
		ID:                uuid.New(),
		FundID:            unauthorizedCID,
		DecisionNumber:    "DEC-DENIED",
		ApprovalRequestID: &approvalReqID,
		Status:            vo.DecisionLifecyclePendingApproval,
		BusinessDate:      time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC),
		CreatedBy:         uuid.New(),
		CreatedAt:         time.Now().UTC(),
	}

	decRepo := newFakeDecisionRepo(authorizedDecision, unauthorizedDecision)
	actor := &fakeBatchActor{}
	perms := &fakeBatchPermissionChecker{allowed: map[uuid.UUID]bool{authorizedCID: true}}

	h := NewDecisionBatchApprovalHandler(decRepo, actor)
	h.SetPermissionChecker(perms)

	results, err := h.BatchApprove(context.Background(), BatchApproveRequest{
		DecisionNos: []string{"DEC-ALLOWED", "DEC-DENIED", "DEC-NONEXISTENT"},
		ActorID:     uuid.New(),
		Comment:     "batch test",
	})
	if err != nil {
		t.Fatalf("unexpected fatal error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Authorized decision must succeed.
	if !results[0].OK {
		t.Errorf("DEC-ALLOWED: expected OK, got error %q", results[0].Error)
	}

	// Unauthorized decision must return the same "not found" as nonexistent.
	if results[1].OK || results[1].Error != "decision not found" {
		t.Errorf("DEC-DENIED: expected 'decision not found', got OK=%v error=%q", results[1].OK, results[1].Error)
	}
	if results[2].OK || results[2].Error != "decision not found" {
		t.Errorf("DEC-NONEXISTENT: expected 'decision not found', got OK=%v error=%q", results[2].OK, results[2].Error)
	}
	// Both "denied" and "nonexistent" must return identical errors (existence-safe).
	if results[1].Error != results[2].Error {
		t.Errorf("unauthorized (%q) and nonexistent (%q) must produce identical errors to prevent existence inference",
			results[1].Error, results[2].Error)
	}
}

// buildDraftDecisionWithReport constructs a DRAFT decision that references the
// given report ID. The decision repo and report repo are pre-seeded.
func buildDraftDecisionWithReport(t *testing.T) (
	d *entity.Decision,
	rep *entity.ResearchReport,
	decRepo *fakeDecisionRepo,
	repRepo *fakeResearchReportRepo,
) {
	t.Helper()
	repID := uuid.New()
	cid := uuid.New()

	rep = &entity.ResearchReport{
		ID:           repID,
		ReportStatus: vo.ReportStatusDraft,
		ReviewStatus: vo.ReviewStatusNotSubmitted,
	}

	d = &entity.Decision{
		ID:               uuid.New(),
		FundID:           cid,
		PortfolioID:      uuid.New(),
		InstrumentCode:   "PTT",
		Side:             vo.OrderSideBuy,
		Status:           vo.DecisionLifecycleDraft,
		ResearchReportID: &repID,
		BusinessDate:     time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC),
		CreatedBy:        uuid.New(),
		CreatedAt:        time.Now().UTC(),
	}

	decRepo = newFakeDecisionRepo(d)
	repRepo = newFakeResearchReportRepo()
	repRepo.seed(rep)
	return
}

func TestSubmitDecision_UnapprovedReport_Rejected(t *testing.T) {
	d, _, decRepo, repRepo := buildDraftDecisionWithReport(t)
	audit := &recordingAudit{}

	// workflow=nil → workflow gate skipped; approval=nil → no approval submission.
	h := NewDecisionCommandHandler(nil, decRepo, repRepo, nil, audit, nil)

	_, err := h.Submit(context.Background(), d.ID, uuid.New())
	if err == nil {
		t.Fatal("expected ErrDecisionReferenceInvalid, got nil")
	}
	var refErr *domain.ErrDecisionReferenceInvalid
	if !errors.As(err, &refErr) {
		t.Fatalf("expected *ErrDecisionReferenceInvalid, got %T: %v", err, err)
	}
	if refErr.DecisionID == "" {
		t.Fatal("ErrDecisionReferenceInvalid must carry the decision ID")
	}
}

func TestSubmitDecision_NoReport_Skips_ReferenceCheck(t *testing.T) {
	// A decision without a ResearchReportID must skip the reference policy.
	// We inject a runTx bypass (same package) so the test doesn't need a DB.
	d := &entity.Decision{
		ID:             uuid.New(),
		FundID:         uuid.New(),
		PortfolioID:    uuid.New(),
		InstrumentCode: "PTT",
		Side:           vo.OrderSideBuy,
		Status:         vo.DecisionLifecycleDraft,
		// ResearchReportID intentionally absent
		BusinessDate: time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC),
		CreatedBy:    uuid.New(),
		CreatedAt:    time.Now().UTC(),
	}
	decRepo := newFakeDecisionRepo(d)
	audit := &recordingAudit{}
	h := NewDecisionCommandHandler(nil, decRepo, newFakeResearchReportRepo(), nil, audit, nil)
	h.runTx = func(ctx context.Context, fn func(pgx.Tx) error) error { return fn(nil) }

	_, err := h.Submit(context.Background(), d.ID, uuid.New())

	var refErr *domain.ErrDecisionReferenceInvalid
	if errors.As(err, &refErr) {
		t.Fatalf("no report → reference check must be skipped, got ErrDecisionReferenceInvalid: %v", err)
	}
	if err != nil {
		t.Fatalf("unexpected error after bypassing runTx: %v", err)
	}
}

// ── P0 Fix #3: cancelling a decision also cancels the active approval request ─

// TestCancel_CancelsActiveApproval verifies that Cancel() on a PENDING_APPROVAL
// decision calls the ApprovalCanceller so the in-flight approval request is
// terminated alongside the decision lifecycle transition.
func TestCancel_CancelsActiveApproval(t *testing.T) {
	d := &entity.Decision{
		ID:              uuid.New(),
		FundID:          uuid.New(),
		PortfolioID:     uuid.New(),
		InstrumentCode:  "PTT",
		Side:            vo.OrderSideBuy,
		Status:          vo.DecisionLifecyclePendingApproval,
		BusinessDate:    time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC),
		CreatedBy:       uuid.New(),
		SubmitterUserID: uuid.New(),
		CreatedAt:       time.Now().UTC(),
	}
	decRepo := newFakeDecisionRepo(d)
	audit := &recordingAudit{}
	canceller := &fakeApprovalCanceller{}

	h := NewDecisionCommandHandler(nil, decRepo, newFakeResearchReportRepo(), nil, audit, nil)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error { return fn(nil) }
	h.SetApprovalCanceller(canceller)

	_, err := h.Cancel(context.Background(), CancelDecisionRequest{
		DecisionID: d.ID,
		Reason:     "Retracting pending decision to revise the trade.",
		ActorID:    uuid.New(),
	})
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if len(canceller.cancelled) != 1 || canceller.cancelled[0] != d.ID {
		t.Fatalf("approval canceller must be invoked with the decision ID; got %v", canceller.cancelled)
	}
}
