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
