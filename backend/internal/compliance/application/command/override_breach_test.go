package command_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
)

// fakeBreach mirrors the small slice of breach state that the override flow
// inspects and mutates. It is mutated only while holding fakeOverrideRepo.mu.
type fakeBreach struct {
	ID     uuid.UUID
	Status entity.BreachStatus
}

// fakeOverrideRepo simulates the atomic semantics of
// PostgresOverrideRepository.CommitOverride without a real database:
//
//   - mu models the SELECT ... FOR UPDATE row lock (one writer at a time
//     per breach is sufficient because every test uses a single breach).
//   - breaches maps BreachID → current status.
//   - overrides enforces UNIQUE(breach_id): once a breach has committed
//     an override, any further attempt returns *domain.ErrOverrideAlreadyExists.
//   - If injectInsertErr is set, the "insert" step fails and every earlier
//     mutation must be rolled back (nothing persisted, breach still OPEN).
//   - If injectUpdateErr is set, the update step fails after the insert
//     was staged; both are rolled back together.
//
// It is NOT a general-purpose mock; it exists to exercise the handler's
// contract with the repository.
type fakeOverrideRepo struct {
	mu              sync.Mutex
	breaches        map[uuid.UUID]*fakeBreach
	overrides       map[uuid.UUID]*entity.Override
	commitCalls     int32
	injectInsertErr error
	injectUpdateErr error
	beforeCommit    func() // test hook fired while the fake holds its lock
}

func newFakeRepo() *fakeOverrideRepo {
	return &fakeOverrideRepo{
		breaches:  make(map[uuid.UUID]*fakeBreach),
		overrides: make(map[uuid.UUID]*entity.Override),
	}
}

func (f *fakeOverrideRepo) seedOpenBreach(id uuid.UUID) {
	f.breaches[id] = &fakeBreach{ID: id, Status: entity.BreachStatusOpen}
}

func (f *fakeOverrideRepo) seedBreach(id uuid.UUID, status entity.BreachStatus) {
	f.breaches[id] = &fakeBreach{ID: id, Status: status}
}

// CommitOverride implements domain.OverrideRepository.
func (f *fakeOverrideRepo) CommitOverride(ctx context.Context, o *entity.Override) error {
	atomic.AddInt32(&f.commitCalls, 1)

	// Simulate BEGIN + SELECT ... FOR UPDATE: serialise on the repo mutex.
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.beforeCommit != nil {
		hook := f.beforeCommit
		// One-shot: clear so nested recursive calls during the hook don't loop.
		f.beforeCommit = nil
		hook()
	}

	// Step 1 — lookup under lock.
	br, ok := f.breaches[o.BreachID]
	if !ok {
		return &domain.ErrBreachNotFound{BreachID: o.BreachID.String()}
	}

	// Step 2 — state check.
	if br.Status != entity.BreachStatusOpen {
		return &domain.ErrBreachNotOpen{
			BreachID:      o.BreachID.String(),
			CurrentStatus: string(br.Status),
		}
	}

	// Step 3 — unique-constraint check (fires before the injected error to
	// mirror Postgres raising SQLSTATE 23505 at INSERT time).
	if _, exists := f.overrides[o.BreachID]; exists {
		return &domain.ErrOverrideAlreadyExists{BreachID: o.BreachID.String()}
	}

	// Step 4 — simulated insert failure. Nothing must persist.
	if f.injectInsertErr != nil {
		return f.injectInsertErr
	}

	// Stage the insert but do not publish until both writes succeed.
	stagedOverride := *o

	// Step 5 — simulated update failure. Insert must also be rolled back.
	if f.injectUpdateErr != nil {
		return f.injectUpdateErr
	}

	// Step 6 — commit: publish both mutations atomically under the lock.
	f.overrides[o.BreachID] = &stagedOverride
	br.Status = entity.BreachStatusOverridden
	return nil
}

// GetByBreachID implements domain.OverrideRepository.
func (f *fakeOverrideRepo) GetByBreachID(_ context.Context, breachID uuid.UUID) (*entity.Override, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	o, ok := f.overrides[breachID]
	if !ok {
		return nil, nil
	}
	cp := *o
	return &cp, nil
}

// List implements domain.OverrideRepository.
func (f *fakeOverrideRepo) List(_ context.Context, breachIDs []uuid.UUID) ([]entity.Override, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]entity.Override, 0, len(breachIDs))
	for _, id := range breachIDs {
		if o, ok := f.overrides[id]; ok {
			out = append(out, *o)
		}
	}
	return out, nil
}

// ensure the fake still satisfies the interface.
var _ domain.OverrideRepository = (*fakeOverrideRepo)(nil)

// ─── tests ────────────────────────────────────────────────────────────────────

func TestOverrideBreach_Success(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	breachID := uuid.New()
	repo.seedOpenBreach(breachID)

	h := command.NewOverrideBreachHandler(repo)

	actor := uuid.New()
	out, err := h.Handle(context.Background(), command.OverrideBreachRequest{
		BreachID:     breachID,
		Reason:       "regulatory exemption 2026-Q2",
		OverriddenBy: actor,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected override entity, got nil")
	}
	if out.BreachID != breachID {
		t.Errorf("BreachID mismatch: got %v, want %v", out.BreachID, breachID)
	}
	if out.OverriddenBy != actor {
		t.Errorf("OverriddenBy mismatch: got %v, want %v", out.OverriddenBy, actor)
	}
	if out.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if out.CreatedAt.Location().String() != "UTC" {
		t.Errorf("CreatedAt must be UTC, got %v", out.CreatedAt.Location())
	}

	// Both mutations must be visible.
	if repo.breaches[breachID].Status != entity.BreachStatusOverridden {
		t.Errorf("breach status should be OVERRIDDEN, got %s", repo.breaches[breachID].Status)
	}
	persisted, _ := repo.GetByBreachID(context.Background(), breachID)
	if persisted == nil {
		t.Fatal("override should be persisted")
	}
	if persisted.ID != out.ID {
		t.Error("persisted override ID mismatch")
	}
}

func TestOverrideBreach_BreachNotFound(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo() // no breach seeded
	h := command.NewOverrideBreachHandler(repo)

	_, err := h.Handle(context.Background(), command.OverrideBreachRequest{
		BreachID:     uuid.New(),
		Reason:       "any",
		OverriddenBy: uuid.New(),
	})
	var target *domain.ErrBreachNotFound
	if !errors.As(err, &target) {
		t.Fatalf("expected *ErrBreachNotFound, got %T: %v", err, err)
	}
}

func TestOverrideBreach_BreachNotOpen_Rejected(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	breachID := uuid.New()
	repo.seedBreach(breachID, entity.BreachStatusResolved)

	h := command.NewOverrideBreachHandler(repo)
	_, err := h.Handle(context.Background(), command.OverrideBreachRequest{
		BreachID:     breachID,
		Reason:       "any",
		OverriddenBy: uuid.New(),
	})
	var target *domain.ErrBreachNotOpen
	if !errors.As(err, &target) {
		t.Fatalf("expected *ErrBreachNotOpen, got %T: %v", err, err)
	}
	if target.CurrentStatus != string(entity.BreachStatusResolved) {
		t.Errorf("CurrentStatus: got %s, want RESOLVED", target.CurrentStatus)
	}
	// Nothing persisted.
	if got, _ := repo.GetByBreachID(context.Background(), breachID); got != nil {
		t.Error("override must not be persisted when breach is not OPEN")
	}
}

func TestOverrideBreach_AlreadyOverridden_IsConflict(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	breachID := uuid.New()
	repo.seedOpenBreach(breachID)

	h := command.NewOverrideBreachHandler(repo)
	req := command.OverrideBreachRequest{
		BreachID:     breachID,
		Reason:       "first",
		OverriddenBy: uuid.New(),
	}
	if _, err := h.Handle(context.Background(), req); err != nil {
		t.Fatalf("first override must succeed, got %v", err)
	}

	// Second request — breach now OVERRIDDEN → ErrBreachNotOpen (state check wins).
	req.Reason = "second"
	req.OverriddenBy = uuid.New()
	_, err := h.Handle(context.Background(), req)
	var notOpen *domain.ErrBreachNotOpen
	if !errors.As(err, &notOpen) {
		t.Fatalf("expected *ErrBreachNotOpen on duplicate, got %T: %v", err, err)
	}
}

// TestOverrideBreach_AtomicRollback_OnInsertError verifies that if the INSERT
// fails (e.g. IO error, non-unique error), no partial state is persisted.
func TestOverrideBreach_AtomicRollback_OnInsertError(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	breachID := uuid.New()
	repo.seedOpenBreach(breachID)
	repo.injectInsertErr = errors.New("simulated insert IO error")

	h := command.NewOverrideBreachHandler(repo)
	_, err := h.Handle(context.Background(), command.OverrideBreachRequest{
		BreachID:     breachID,
		Reason:       "any",
		OverriddenBy: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error from injected insert failure")
	}

	// Breach must still be OPEN.
	if repo.breaches[breachID].Status != entity.BreachStatusOpen {
		t.Errorf("breach status should remain OPEN after rollback, got %s",
			repo.breaches[breachID].Status)
	}
	// Override must not be persisted.
	if got, _ := repo.GetByBreachID(context.Background(), breachID); got != nil {
		t.Error("override must not be persisted when insert fails")
	}
}

// TestOverrideBreach_AtomicRollback_OnBreachUpdateError verifies that a
// failure between insert and breach-status-update rolls back the insert too.
func TestOverrideBreach_AtomicRollback_OnBreachUpdateError(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	breachID := uuid.New()
	repo.seedOpenBreach(breachID)
	repo.injectUpdateErr = errors.New("simulated update IO error")

	h := command.NewOverrideBreachHandler(repo)
	_, err := h.Handle(context.Background(), command.OverrideBreachRequest{
		BreachID:     breachID,
		Reason:       "any",
		OverriddenBy: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error from injected update failure")
	}

	if repo.breaches[breachID].Status != entity.BreachStatusOpen {
		t.Errorf("breach status should remain OPEN after rollback, got %s",
			repo.breaches[breachID].Status)
	}
	if got, _ := repo.GetByBreachID(context.Background(), breachID); got != nil {
		t.Error("override must not be persisted when update fails")
	}
}

// TestOverrideBreach_ConcurrentAttempts_OnlyOneWins runs N goroutines trying
// to override the same breach simultaneously. Exactly one must succeed, and
// all losers must see a typed domain error (either NotOpen or AlreadyExists —
// both are acceptable depending on which check the loser hits first).
func TestOverrideBreach_ConcurrentAttempts_OnlyOneWins(t *testing.T) {
	t.Parallel()

	const attempts = 16
	repo := newFakeRepo()
	breachID := uuid.New()
	repo.seedOpenBreach(breachID)

	h := command.NewOverrideBreachHandler(repo)

	var (
		wg            sync.WaitGroup
		successes     int32
		typedFailures int32
	)
	start := make(chan struct{})
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := h.Handle(context.Background(), command.OverrideBreachRequest{
				BreachID:     breachID,
				Reason:       "race",
				OverriddenBy: uuid.New(),
			})
			if err == nil {
				atomic.AddInt32(&successes, 1)
				return
			}
			var notOpen *domain.ErrBreachNotOpen
			var dup *domain.ErrOverrideAlreadyExists
			if errors.As(err, &notOpen) || errors.As(err, &dup) {
				atomic.AddInt32(&typedFailures, 1)
				return
			}
			t.Errorf("unexpected error from concurrent loser: %T: %v", err, err)
		}()
	}
	close(start)
	wg.Wait()

	if successes != 1 {
		t.Errorf("exactly one goroutine should succeed, got %d", successes)
	}
	if typedFailures != attempts-1 {
		t.Errorf("expected %d typed failures, got %d", attempts-1, typedFailures)
	}

	// Repo state should show exactly one persisted override.
	if repo.breaches[breachID].Status != entity.BreachStatusOverridden {
		t.Errorf("breach should be OVERRIDDEN, got %s", repo.breaches[breachID].Status)
	}
	list, _ := repo.List(context.Background(), []uuid.UUID{breachID})
	if len(list) != 1 {
		t.Errorf("expected exactly 1 override persisted, got %d", len(list))
	}
}

func TestOverrideBreach_ValidationErrors(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	h := command.NewOverrideBreachHandler(repo)

	cases := []struct {
		name      string
		req       command.OverrideBreachRequest
		wantField string
	}{
		{"missing breach id", command.OverrideBreachRequest{Reason: "x", OverriddenBy: uuid.New()}, "breach_id"},
		{"missing reason", command.OverrideBreachRequest{BreachID: uuid.New(), OverriddenBy: uuid.New()}, "reason"},
		{"missing actor", command.OverrideBreachRequest{BreachID: uuid.New(), Reason: "x"}, "overridden_by"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := h.Handle(context.Background(), tc.req)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			var invalid *domain.ErrInvalidOverrideRequest
			if !errors.As(err, &invalid) {
				t.Fatalf("expected *ErrInvalidOverrideRequest, got %T: %v", err, err)
			}
			if invalid.Field != tc.wantField {
				t.Errorf("Field: got %q, want %q", invalid.Field, tc.wantField)
			}
		})
	}
	// Handler must not have touched the repo for any of the validation failures.
	if got := atomic.LoadInt32(&repo.commitCalls); got != 0 {
		t.Errorf("CommitOverride must not be called on validation failure, got %d calls", got)
	}
}

// TestOverrideBreach_InternalRepoError_NotValidationError verifies that a
// genuine server-side failure from CommitOverride (simulated DB error) is
// surfaced as-is — it must NOT be mistaken for a validation error, because
// the transport layer relies on the type distinction to pick 500 vs 400.
func TestOverrideBreach_InternalRepoError_NotValidationError(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	breachID := uuid.New()
	repo.seedOpenBreach(breachID)
	repo.injectInsertErr = errors.New("connection reset by peer")

	h := command.NewOverrideBreachHandler(repo)
	_, err := h.Handle(context.Background(), command.OverrideBreachRequest{
		BreachID:     breachID,
		Reason:       "any",
		OverriddenBy: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error")
	}

	// Must NOT match any typed domain error — the transport layer must fall
	// through to its default (500) branch.
	var invalid *domain.ErrInvalidOverrideRequest
	var notFound *domain.ErrBreachNotFound
	var notOpen *domain.ErrBreachNotOpen
	var dup *domain.ErrOverrideAlreadyExists
	if errors.As(err, &invalid) || errors.As(err, &notFound) ||
		errors.As(err, &notOpen) || errors.As(err, &dup) {
		t.Errorf("internal error must not match a typed domain error: %T: %v", err, err)
	}
}
