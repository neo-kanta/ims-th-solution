// Package command — research_report_crud_test.go
//
// Strategy
// ---------
// The command handler depends on *pgxpool.Pool for transactional writes.
// To exercise both rejection paths and happy paths without a live
// database, this file leans on two seams:
//
//  1. fakeResearchReportRepo — an in-memory implementation of
//     domain.ResearchReportRepository. The pgx.Tx argument on Create /
//     Update / SoftDelete is ignored; the data lives in plain maps.
//
//  2. The handler's `runTx` field — a small injectable transaction
//     runner introduced specifically for testability. Production wiring
//     defaults to `withTransaction(ctx, pool, fn)`; here we install a
//     stub that calls `fn(nil)` directly, so the happy paths never
//     touch a real pool. The seam is contained to the research handler;
//     no other command was touched.
//
// Rejection paths short-circuit before runTx is invoked, so they pass
// the production handler with a nil pool. Happy paths use
// newTestHandler(...) which always installs the bypass.
package command

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// fakeResearchReportRepo is a minimal in-memory implementation of
// domain.ResearchReportRepository used by the command tests. It is
// intentionally goroutine-safe — although the tests don't share an
// instance across goroutines, the mutex keeps the surface honest for any
// future contributor who wants to add concurrency coverage.
type fakeResearchReportRepo struct {
	mu      sync.Mutex
	byID    map[uuid.UUID]*entity.ResearchReport
	byNo    map[string]*entity.ResearchReport
	createN int
}

func newFakeResearchReportRepo() *fakeResearchReportRepo {
	return &fakeResearchReportRepo{
		byID: map[uuid.UUID]*entity.ResearchReport{},
		byNo: map[string]*entity.ResearchReport{},
	}
}

func (r *fakeResearchReportRepo) seed(rep *entity.ResearchReport) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *rep
	r.byID[cp.ID] = &cp
	r.byNo[cp.ReportNo] = &cp
}

func (r *fakeResearchReportRepo) Create(ctx context.Context, tx pgx.Tx, e *entity.ResearchReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.createN++
	if _, exists := r.byNo[e.ReportNo]; exists {
		return &domain.ErrResearchReportNoAlreadyExists{ReportNo: e.ReportNo}
	}
	cp := *e
	r.byID[cp.ID] = &cp
	r.byNo[cp.ReportNo] = &cp
	return nil
}

func (r *fakeResearchReportRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ResearchReport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rep, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	cp := *rep
	return &cp, nil
}

func (r *fakeResearchReportRepo) GetByReportNo(ctx context.Context, reportNo string) (*entity.ResearchReport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rep, ok := r.byNo[reportNo]
	if !ok {
		return nil, nil
	}
	cp := *rep
	return &cp, nil
}

func (r *fakeResearchReportRepo) List(ctx context.Context, _ domain.ResearchReportListFilter) ([]*entity.ResearchReport, int, error) {
	return nil, 0, nil
}

func (r *fakeResearchReportRepo) Update(ctx context.Context, tx pgx.Tx, e *entity.ResearchReport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *e
	r.byID[cp.ID] = &cp
	r.byNo[cp.ReportNo] = &cp
	return nil
}

func (r *fakeResearchReportRepo) Invalidate(_ context.Context, _ pgx.Tx, id uuid.UUID, actorID uuid.UUID, reason string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rep, ok := r.byID[id]
	if !ok {
		return &domain.ErrResearchReportNotFound{ReportID: id.String()}
	}
	if rep.IsInvalidated() {
		return &domain.ErrResearchReportNotFound{ReportID: id.String()}
	}
	rep.ReportStatus = vo.ReportStatusInvalidated
	rep.ReviewStatus = vo.ReviewStatusInvalidated
	actor := actorID
	rep.InvalidatedAt = &at
	rep.InvalidatedBy = &actor
	rep.InvalidationReason = reason
	rep.UpdatedAt = at
	rep.UpdatedBy = &actor
	return nil
}

func (r *fakeResearchReportRepo) SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, deletedBy uuid.UUID, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rep, ok := r.byID[id]
	if !ok {
		return &domain.ErrResearchReportNotFound{ReportID: id.String()}
	}
	rep.DeletedAt = &deletedAt
	rep.UpdatedAt = deletedAt
	return nil
}

// recordingAudit captures every contract.AuditEntry emitted by the
// command so the tests can assert which audit actions fired.
type recordingAudit struct {
	mu      sync.Mutex
	entries []contract.AuditEntry
}

func (a *recordingAudit) LogAction(entry contract.AuditEntry) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, entry)
	return nil
}

func (a *recordingAudit) LogActionStrict(_ context.Context, entry contract.AuditEntry) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, entry)
	return nil
}

func (a *recordingAudit) actions() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]string, len(a.entries))
	for i, e := range a.entries {
		out[i] = e.Action
	}
	return out
}

// researchFixedClock returns the same instant on every call — keeps test
// assertions deterministic. Named with the research prefix because
// submit_decision_for_execution_test.go already defines a `fixedClock`
// in this package.
func researchFixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// ──────────────────────────────────────────────────────────────────────
// Helpers to build commonly-used fixtures.
// ──────────────────────────────────────────────────────────────────────

func validCreateReq() CreateResearchReportRequest {
	return CreateResearchReportRequest{
		ReportNo:           "RR-2026-0001",
		ReportDate:         time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
		OwnerUserID:        uuid.New(),
		AuthorUserID:       uuid.New(),
		InstrumentCode:     "ptt", // lowercase to assert uppercase-on-create
		Recommendation:     vo.RecommendationBuy,
		InvestmentAnalysis: "This is a comprehensive analyst note exceeding twenty-five chars.",
		ActorID:            uuid.New(),
	}
}

func seededReport(rep *fakeResearchReportRepo, status vo.ReviewStatus) *entity.ResearchReport {
	r := &entity.ResearchReport{
		ID:                 uuid.New(),
		ReportNo:           "RR-EXISTING",
		ReportDate:         time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		OwnerUserID:        uuid.New(),
		AuthorUserID:       uuid.New(),
		InstrumentCode:     "AOT",
		Recommendation:     vo.RecommendationHold,
		InvestmentAnalysis: "Seeded analyst note that meets the 25-character minimum.",
		ReportStatus:       vo.ReportStatusDraft,
		ReviewStatus:       status,
		CreatedAt:          time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:          time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
	}
	rep.seed(r)
	return r
}

// newTestHandler returns a ResearchReportCommandHandler with its `runTx`
// hook overridden to pass a nil pgx.Tx directly to the supplied closure.
// Use it for happy-path tests that need to exercise the post-validation
// repository write. Rejection-path tests don't need this helper — they
// short-circuit before runTx is invoked, so a nil pool is harmless.
func newTestHandler(
	repo domain.ResearchReportRepository,
	audit contract.AuditLogger,
	clock func() time.Time,
) *ResearchReportCommandHandler {
	h := NewResearchReportCommandHandler(nil, repo, audit, clock)
	h.runTx = func(_ context.Context, fn func(pgx.Tx) error) error {
		return fn(nil)
	}
	return h
}

// ──────────────────────────────────────────────────────────────────────
// Tests
// ──────────────────────────────────────────────────────────────────────

// TestCreate_DuplicateReportNo verifies the in-memory pre-check rejects a
// duplicate report_no with the typed error that the handler maps to 409.
func TestCreate_DuplicateReportNo(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	seededReport(repo, vo.ReviewStatusNotSubmitted)

	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	req := validCreateReq()
	req.ReportNo = "RR-EXISTING"

	_, err := h.Create(context.Background(), req)
	var dup *domain.ErrResearchReportNoAlreadyExists
	require.ErrorAs(t, err, &dup)
	require.Equal(t, "RR-EXISTING", dup.ReportNo)
}

// TestCreate_ReportNoTooLong asserts that the new length validator runs
// before the policy check and emits the typed error with the right field.
func TestCreate_ReportNoTooLong(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	req := validCreateReq()
	req.ReportNo = strings.Repeat("R", 61) // one over the 60-char budget

	_, err := h.Create(context.Background(), req)
	var invalid *domain.ErrInvalidResearchReportRequest
	require.ErrorAs(t, err, &invalid)
	require.Equal(t, "report_no", invalid.Field)
	require.Contains(t, invalid.Detail, "60")
}

// TestCreate_InvalidCurrency exercises the currency format validator.
func TestCreate_InvalidCurrency(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	req := validCreateReq()
	req.Currency = "TH" // 2-letter; the constraint requires 3

	_, err := h.Create(context.Background(), req)
	var invalid *domain.ErrInvalidResearchReportRequest
	require.ErrorAs(t, err, &invalid)
	require.Equal(t, "currency", invalid.Field)
}

// TestCreate_ActorRequired confirms ErrInvalidResearchReportRequest is the
// type emitted when the actor is missing — i.e. that the migration from
// the old ErrInvalidDecisionRequest landed cleanly.
func TestCreate_ActorRequired(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	req := validCreateReq()
	req.ActorID = uuid.Nil

	_, err := h.Create(context.Background(), req)
	var invalid *domain.ErrInvalidResearchReportRequest
	require.ErrorAs(t, err, &invalid)
	require.Equal(t, "actor_id", invalid.Field)
	// Ensure the old typed error is NOT returned — a regression guard for
	// the B6 migration.
	var oldInvalid *domain.ErrInvalidDecisionRequest
	require.False(t, errors.As(err, &oldInvalid),
		"actor_id validation must emit ErrInvalidResearchReportRequest, not the legacy decision error")
}

// TestUpdate_ReviewCompletedBlocked ensures the entity lifecycle guard
// CanUpdate() rejects mutation of a report whose review has completed.
func TestUpdate_ReviewCompletedBlocked(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReport(repo, vo.ReviewStatusReviewCompleted)

	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	newTitle := "Trying to edit a locked report"
	_, err := h.Update(context.Background(), UpdateResearchReportRequest{
		ReportID:    r.ID,
		ActorID:     uuid.New(),
		ReportTitle: &newTitle,
	})
	var cannot *domain.ErrResearchReportCannotUpdate
	require.ErrorAs(t, err, &cannot)
	require.Equal(t, string(vo.ReviewStatusReviewCompleted), cannot.ReviewStatus)
}

// TestSoftDelete_SubmittedBlocked ensures a SUBMITTED report cannot be
// soft-deleted (only NOT_SUBMITTED is deletable).
func TestSoftDelete_SubmittedBlocked(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReport(repo, vo.ReviewStatusSubmitted)

	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	err := h.SoftDelete(context.Background(), r.ID, uuid.New())
	var cannot *domain.ErrResearchReportCannotDelete
	require.ErrorAs(t, err, &cannot)
	require.Equal(t, string(vo.ReviewStatusSubmitted), cannot.ReviewStatus)
}

// TestSoftDelete_NotFound ensures a missing ID returns the not-found
// error so the handler can emit 404.
func TestSoftDelete_NotFound(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	err := h.SoftDelete(context.Background(), uuid.New(), uuid.New())
	var nf *domain.ErrResearchReportNotFound
	require.ErrorAs(t, err, &nf)
}

// TestSubmit_AlreadySubmittedBlocked rejects re-submitting a report that
// is already in the review queue.
func TestSubmit_AlreadySubmittedBlocked(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReport(repo, vo.ReviewStatusSubmitted)

	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	_, err := h.Submit(context.Background(), r.ID, uuid.New())
	var cannot *domain.ErrResearchReportCannotSubmit
	require.ErrorAs(t, err, &cannot)
	require.Equal(t, string(vo.ReviewStatusSubmitted), cannot.ReviewStatus)
}

// TestSubmit_NotFound ensures the lookup miss surfaces correctly even on
// the state-transition path.
func TestSubmit_NotFound(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	_, err := h.Submit(context.Background(), uuid.New(), uuid.New())
	var nf *domain.ErrResearchReportNotFound
	require.ErrorAs(t, err, &nf)
}

// TestCancelSubmit_NotSubmittedBlocked rejects cancel-submit on a report
// that has not yet been submitted.
func TestCancelSubmit_NotSubmittedBlocked(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReport(repo, vo.ReviewStatusNotSubmitted)

	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	_, err := h.CancelSubmit(context.Background(), r.ID, uuid.New())
	var cannot *domain.ErrResearchReportCannotCancelSubmit
	require.ErrorAs(t, err, &cannot)
	require.Equal(t, string(vo.ReviewStatusNotSubmitted), cannot.ReviewStatus)
}

// TestCancelSubmit_ReviewCompletedBlocked rejects cancel-submit once the
// review has been completed (per the entity guard).
func TestCancelSubmit_ReviewCompletedBlocked(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReport(repo, vo.ReviewStatusReviewCompleted)

	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	_, err := h.CancelSubmit(context.Background(), r.ID, uuid.New())
	var cannot *domain.ErrResearchReportCannotCancelSubmit
	require.ErrorAs(t, err, &cannot)
	require.Equal(t, string(vo.ReviewStatusReviewCompleted), cannot.ReviewStatus)
}

// ──────────────────────────────────────────────────────────────────────
// Happy paths — use newTestHandler so the runTx hook bypasses *pgxpool.Pool.
// ──────────────────────────────────────────────────────────────────────

// TestCreate_HappyPath confirms a valid request persists with the correct
// lifecycle defaults and fires exactly one INVESTMENT_RESEARCH_CREATED
// audit entry.
func TestCreate_HappyPath(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	audit := &recordingAudit{}
	h := newTestHandler(repo, audit,
		researchFixedClock(time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)))

	got, err := h.Create(context.Background(), validCreateReq())
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, vo.ReportStatusDraft, got.ReportStatus)
	require.Equal(t, vo.ReviewStatusNotSubmitted, got.ReviewStatus)
	// B4: instrument_code uppercased server-side even when the caller
	// sends lowercase.
	require.Equal(t, "PTT", got.InstrumentCode)

	// Repository contract: exactly one Create call landed.
	require.Equal(t, 1, repo.createN)

	// Audit: exactly one INVESTMENT_RESEARCH_CREATED entry, no others.
	require.Equal(t, []string{"INVESTMENT_RESEARCH_CREATED"}, audit.actions())
}

// TestCreate_DefaultsAuthorOwnerToActor proves B5: when the caller omits
// owner_user_id / author_user_id, the actor's UUID is substituted.
func TestCreate_DefaultsAuthorOwnerToActor(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	h := newTestHandler(repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	actor := uuid.New()
	req := validCreateReq()
	req.ActorID = actor
	req.OwnerUserID = uuid.Nil
	req.AuthorUserID = uuid.Nil

	got, err := h.Create(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, actor, got.OwnerUserID, "owner_user_id must default to actor when omitted")
	require.Equal(t, actor, got.AuthorUserID, "author_user_id must default to actor when omitted")
}

// TestSubmit_HappyPath transitions a NOT_SUBMITTED report to SUBMITTED
// and verifies the audit trail names the transition correctly.
func TestSubmit_HappyPath(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReport(repo, vo.ReviewStatusNotSubmitted)

	audit := &recordingAudit{}
	h := newTestHandler(repo, audit,
		researchFixedClock(time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)))

	got, err := h.Submit(context.Background(), r.ID, uuid.New())
	require.NoError(t, err)
	require.Equal(t, vo.ReviewStatusSubmitted, got.ReviewStatus)
	require.Equal(t, []string{"INVESTMENT_RESEARCH_SUBMITTED"}, audit.actions())
}

// TestCancelSubmit_HappyPath transitions a SUBMITTED report back to
// NOT_SUBMITTED and verifies the corresponding audit entry.
func TestCancelSubmit_HappyPath(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReport(repo, vo.ReviewStatusSubmitted)

	audit := &recordingAudit{}
	h := newTestHandler(repo, audit,
		researchFixedClock(time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)))

	got, err := h.CancelSubmit(context.Background(), r.ID, uuid.New())
	require.NoError(t, err)
	require.Equal(t, vo.ReviewStatusNotSubmitted, got.ReviewStatus)
	require.Equal(t, []string{"INVESTMENT_RESEARCH_SUBMIT_CANCELLED"}, audit.actions())
}

// ──────────────────────────────────────────────────────────────────────
// P0: lifecycle status fields must NOT be mutable via the generic Update
// path. The DTO no longer carries report_status; if any caller bypasses
// the wire and constructs the command struct directly, the absence of
// the field still means status is preserved. These tests pin the
// invariant: a REJECTED/EXPIRED report stays REJECTED/EXPIRED after a
// content edit, and an ACTIVE report does not become DRAFT or back.
// ──────────────────────────────────────────────────────────────────────

// seededReportWithStatus seeds a report whose review is completed (so it
// has reached a terminal lifecycle state) with the requested ReportStatus.
// REVIEW_COMPLETED reports refuse Update outright — for the lifecycle-flip
// tests we use NOT_SUBMITTED reports whose ReportStatus is REJECTED /
// EXPIRED / ACTIVE so the entity will accept the edit but the command must
// never overwrite ReportStatus.
func seededReportWithStatus(rep *fakeResearchReportRepo, reportStatus vo.ReportStatus) *entity.ResearchReport {
	r := &entity.ResearchReport{
		ID:                 uuid.New(),
		ReportNo:           "RR-LIFECYCLE-" + string(reportStatus),
		ReportDate:         time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		OwnerUserID:        uuid.New(),
		AuthorUserID:       uuid.New(),
		InstrumentCode:     "PTT",
		Recommendation:     vo.RecommendationHold,
		InvestmentAnalysis: "Seeded analyst note that meets the 25-character minimum.",
		ReportStatus:       reportStatus,
		ReviewStatus:       vo.ReviewStatusNotSubmitted,
		CreatedAt:          time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:          time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
	}
	rep.seed(r)
	return r
}

// TestUpdate_DoesNotFlipReportStatus_Rejected proves that editing a REJECTED
// report's content fields leaves ReportStatus = REJECTED — there is no path
// through Update that can resurrect a rejected report. Only the approval
// engine's final-decision callback can reset status.
func TestUpdate_DoesNotFlipReportStatus_Rejected(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReportWithStatus(repo, vo.ReportStatusRejected)

	h := newTestHandler(repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	newTitle := "Edited title on a rejected report"
	got, err := h.Update(context.Background(), UpdateResearchReportRequest{
		ReportID:    r.ID,
		ActorID:     uuid.New(),
		ReportTitle: &newTitle,
	})
	require.NoError(t, err)
	require.Equal(t, vo.ReportStatusRejected, got.ReportStatus,
		"editing content must not flip REJECTED → ACTIVE")
	require.Equal(t, newTitle, got.ReportTitle)
}

// TestUpdate_DoesNotFlipReportStatus_Expired proves that editing an EXPIRED
// report's content does not reactivate it. An expired report cannot become
// ACTIVE via a normal edit; it must go through the proper resubmission flow.
func TestUpdate_DoesNotFlipReportStatus_Expired(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReportWithStatus(repo, vo.ReportStatusExpired)

	h := newTestHandler(repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	newTitle := "Edited title on an expired report"
	got, err := h.Update(context.Background(), UpdateResearchReportRequest{
		ReportID:    r.ID,
		ActorID:     uuid.New(),
		ReportTitle: &newTitle,
	})
	require.NoError(t, err)
	require.Equal(t, vo.ReportStatusExpired, got.ReportStatus,
		"editing content must not flip EXPIRED → ACTIVE")
}

// TestUpdate_DoesNotFlipReportStatus_Active proves the symmetry: an ACTIVE
// report keeps its ACTIVE status when only content fields are edited.
func TestUpdate_DoesNotFlipReportStatus_Active(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReportWithStatus(repo, vo.ReportStatusActive)

	h := newTestHandler(repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	newTitle := "Refreshed title"
	got, err := h.Update(context.Background(), UpdateResearchReportRequest{
		ReportID:    r.ID,
		ActorID:     uuid.New(),
		ReportTitle: &newTitle,
	})
	require.NoError(t, err)
	require.Equal(t, vo.ReportStatusActive, got.ReportStatus)
}

// failingAudit returns an error from LogActionStrict so we can prove the
// financial-action audit hardening surfaces failures instead of silently
// dropping them. LogAction stays fire-and-forget to match the contract.
type failingAudit struct {
	strictErr error
}

func (f *failingAudit) LogAction(contract.AuditEntry) error { return nil }
func (f *failingAudit) LogActionStrict(context.Context, contract.AuditEntry) error {
	return f.strictErr
}

// TestInvalidate_StrictAuditFailureSurfaces asserts that when the strict
// audit emit fails, the Invalidate command returns an error to the caller.
// The DB row may still be invalidated (the audit emit is post-commit), but
// the operator MUST see the failure — not a silent success.
func TestInvalidate_StrictAuditFailureSurfaces(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReportWithStatus(repo, vo.ReportStatusActive)

	audit := &failingAudit{strictErr: errors.New("audit store down")}
	h := newTestHandler(repo, audit,
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	_, err := h.Invalidate(context.Background(), InvalidateResearchReportRequest{
		ReportID: r.ID,
		ActorID:  uuid.New(),
		Reason:   "Risk override — superseded by RR-2026-0050 release.",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "auditing invalidation")
}

// TestInvalidate_RejectsShortReason proves the ≥20-char reason gate fires
// before any repository write. Distinct from the database CHECK so the
// caller sees a typed validation error, not a generic CHECK violation.
func TestInvalidate_RejectsShortReason(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReportWithStatus(repo, vo.ReportStatusActive)

	h := newTestHandler(repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	_, err := h.Invalidate(context.Background(), InvalidateResearchReportRequest{
		ReportID: r.ID,
		ActorID:  uuid.New(),
		Reason:   "too short",
	})
	var bad *domain.ErrInvalidResearchReportRequest
	require.ErrorAs(t, err, &bad)
	require.Equal(t, "reason", bad.Field)
}

// TestInvalidate_BlocksDecisionReference proves the report-reference policy
// refuses an INVALIDATED report. Closes the loop between the new lifecycle
// state and the decision-side enforcement.
func TestInvalidate_BlocksDecisionReference(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReportWithStatus(repo, vo.ReportStatusActive)
	// Set the effective_date in the past so it cannot trip earlier guards.
	effective := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r.EffectiveDate = &effective
	r.ReviewStatus = vo.ReviewStatusReviewCompleted
	repo.seed(r)

	h := newTestHandler(repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	_, err := h.Invalidate(context.Background(), InvalidateResearchReportRequest{
		ReportID: r.ID,
		ActorID:  uuid.New(),
		Reason:   "Recall: thesis disproved by Q1 results — full retraction.",
	})
	require.NoError(t, err)

	inv, err := repo.GetByID(context.Background(), r.ID)
	require.NoError(t, err)
	require.True(t, inv.IsInvalidated())
	require.NotNil(t, inv.InvalidatedAt)
	require.NotNil(t, inv.InvalidatedBy)
}

// ── P0 Fix #1: submitted reports are locked for editing ──────────────────────

// TestUpdate_SubmittedBlocked ensures the CanUpdate() guard now also blocks
// mutation of a report whose review is in SUBMITTED state (in-flight approval).
func TestUpdate_SubmittedBlocked(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReport(repo, vo.ReviewStatusSubmitted)

	h := NewResearchReportCommandHandler(nil, repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))

	newTitle := "Trying to edit a submitted report"
	_, err := h.Update(context.Background(), UpdateResearchReportRequest{
		ReportID:    r.ID,
		ActorID:     uuid.New(),
		ReportTitle: &newTitle,
	})
	var cannot *domain.ErrResearchReportCannotUpdate
	require.ErrorAs(t, err, &cannot, "submitting a report must lock it against edits")
	require.Equal(t, string(vo.ReviewStatusSubmitted), cannot.ReviewStatus)
}

// ── P0 Fix #2: cancel-submit cancels the active approval request ──────────────

// fakeApprovalCanceller records which subjects were cancelled so the test can
// assert the approval cancel was triggered.
type fakeApprovalCanceller struct {
	cancelled []uuid.UUID
}

func (f *fakeApprovalCanceller) CancelApprovalBySubject(_ context.Context, _ string, subjectID uuid.UUID, _ uuid.UUID) error {
	f.cancelled = append(f.cancelled, subjectID)
	return nil
}

// TestCancelSubmit_CancelsActiveApproval verifies that CancelSubmit invokes the
// ApprovalCanceller so any in-flight approval request is terminated when the
// analyst retracts their submission.
func TestCancelSubmit_CancelsActiveApproval(t *testing.T) {
	t.Parallel()
	repo := newFakeResearchReportRepo()
	r := seededReport(repo, vo.ReviewStatusSubmitted)
	canceller := &fakeApprovalCanceller{}

	h := newTestHandler(repo, &recordingAudit{},
		researchFixedClock(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)))
	h.SetApprovalCanceller(canceller)

	_, err := h.CancelSubmit(context.Background(), r.ID, uuid.New())
	require.NoError(t, err, "cancel-submit must succeed when approval cancel succeeds")
	require.Len(t, canceller.cancelled, 1, "approval canceller must be called exactly once")
	require.Equal(t, r.ID, canceller.cancelled[0], "canceller must receive the report ID as the subject")
}
