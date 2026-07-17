package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// ─── fake binding repo ──────────────────────────────────────────────────────

type fakeRuleBindingRepo struct {
	bindings   map[uuid.UUID]entity.RuleBinding
	createErr  error
	deactivate []uuid.UUID
}

func newFakeRuleBindingRepo() *fakeRuleBindingRepo {
	return &fakeRuleBindingRepo{bindings: map[uuid.UUID]entity.RuleBinding{}}
}

func (f *fakeRuleBindingRepo) Create(_ context.Context, b *entity.RuleBinding) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.bindings[b.ID] = *b
	return nil
}

func (f *fakeRuleBindingRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.RuleBinding, error) {
	if b, ok := f.bindings[id]; ok {
		return &b, nil
	}
	return nil, nil
}

func (f *fakeRuleBindingRepo) List(_ context.Context, filter domain.BindingFilter) ([]entity.RuleBinding, int64, error) {
	var out []entity.RuleBinding
	for _, b := range f.bindings {
		if filter.RuleInstanceID != nil && b.RuleInstanceID != *filter.RuleInstanceID {
			continue
		}
		if filter.ScopeType != nil && b.Scope.Type != *filter.ScopeType {
			continue
		}
		if filter.ScopeID != nil {
			if b.Scope.ID == nil || *b.Scope.ID != *filter.ScopeID {
				continue
			}
		}
		if filter.IsActive != nil && b.IsActive != *filter.IsActive {
			continue
		}
		out = append(out, b)
	}
	return out, int64(len(out)), nil
}

func (f *fakeRuleBindingRepo) Deactivate(_ context.Context, id uuid.UUID) error {
	f.deactivate = append(f.deactivate, id)
	if b, ok := f.bindings[id]; ok {
		b.IsActive = false
		f.bindings[id] = b
	}
	return nil
}

func (f *fakeRuleBindingRepo) ResolveApplicable(_ context.Context, _ []vo.Scope, _ time.Time) ([]domain.ResolvedBinding, error) {
	return nil, nil
}

var _ domain.RuleBindingRepository = (*fakeRuleBindingRepo)(nil)

// ─── helpers ────────────────────────────────────────────────────────────────

func seededInstanceRepo(active bool) (*fakeRuleInstanceRepo, uuid.UUID) {
	repo := newFakeRuleInstanceRepo()
	id := uuid.New()
	repo.instances[id] = entity.RuleInstance{ID: id, RuleTypeID: "cash.availability", IsActive: active}
	return repo, id
}

func validCreateBindingReq(ruleInstanceID, portfolioID uuid.UUID) command.CreateRuleBindingRequest {
	return command.CreateRuleBindingRequest{
		RuleInstanceID: ruleInstanceID,
		ScopeType:      vo.ScopePortfolio,
		ScopeID:        &portfolioID,
		Severity:       vo.SeverityBlock,
		EffectiveFrom:  time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		CreatedBy:      uuid.New(),
	}
}

// ─── happy path ─────────────────────────────────────────────────────────────

func TestCreateRuleBinding_HappyPath(t *testing.T) {
	instRepo, instID := seededInstanceRepo(true)
	bindRepo := newFakeRuleBindingRepo()
	h := command.NewCreateRuleBindingHandler(bindRepo, instRepo)

	portfolioID := uuid.New()
	binding, err := h.Handle(context.Background(), validCreateBindingReq(instID, portfolioID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if binding.Scope.Type != vo.ScopePortfolio {
		t.Fatalf("expected PORTFOLIO scope, got %s", binding.Scope.Type)
	}
	if binding.Scope.ID == nil || *binding.Scope.ID != portfolioID {
		t.Fatalf("expected scope_id=%s, got %v", portfolioID, binding.Scope.ID)
	}
	if !binding.IsActive {
		t.Fatal("expected new binding to be active")
	}
	if binding.Priority != 100 {
		t.Fatalf("expected default priority 100, got %d", binding.Priority)
	}
}

func TestCreateRuleBinding_UnknownRuleInstance_NotFound(t *testing.T) {
	instRepo := newFakeRuleInstanceRepo()
	bindRepo := newFakeRuleBindingRepo()
	h := command.NewCreateRuleBindingHandler(bindRepo, instRepo)

	_, err := h.Handle(context.Background(), validCreateBindingReq(uuid.New(), uuid.New()))
	var notFound *domain.ErrRuleInstanceNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected ErrRuleInstanceNotFound, got %T: %v", err, err)
	}
}

func TestCreateRuleBinding_InactiveRuleInstance_Rejected(t *testing.T) {
	instRepo, instID := seededInstanceRepo(false)
	bindRepo := newFakeRuleBindingRepo()
	h := command.NewCreateRuleBindingHandler(bindRepo, instRepo)

	_, err := h.Handle(context.Background(), validCreateBindingReq(instID, uuid.New()))
	var inactive *domain.ErrRuleInstanceInactive
	if !errors.As(err, &inactive) {
		t.Fatalf("expected ErrRuleInstanceInactive, got %T: %v", err, err)
	}
}

func TestCreateRuleBinding_DuplicateActivePortfolioBinding_Rejected(t *testing.T) {
	instRepo, instID := seededInstanceRepo(true)
	bindRepo := newFakeRuleBindingRepo()
	h := command.NewCreateRuleBindingHandler(bindRepo, instRepo)

	portfolioID := uuid.New()
	req := validCreateBindingReq(instID, portfolioID)
	if _, err := h.Handle(context.Background(), req); err != nil {
		t.Fatalf("first bind should succeed: %v", err)
	}

	_, err := h.Handle(context.Background(), req)
	var dup *domain.ErrDuplicateActiveBinding
	if !errors.As(err, &dup) {
		t.Fatalf("expected ErrDuplicateActiveBinding on second bind, got %T: %v", err, err)
	}
}

func TestCreateRuleBinding_SameRuleDifferentPortfolio_Allowed(t *testing.T) {
	instRepo, instID := seededInstanceRepo(true)
	bindRepo := newFakeRuleBindingRepo()
	h := command.NewCreateRuleBindingHandler(bindRepo, instRepo)

	if _, err := h.Handle(context.Background(), validCreateBindingReq(instID, uuid.New())); err != nil {
		t.Fatalf("first bind: %v", err)
	}
	if _, err := h.Handle(context.Background(), validCreateBindingReq(instID, uuid.New())); err != nil {
		t.Fatalf("second bind to a different portfolio must succeed: %v", err)
	}
}

func TestCreateRuleBinding_GlobalScopeRequiresNilScopeID(t *testing.T) {
	instRepo, instID := seededInstanceRepo(true)
	bindRepo := newFakeRuleBindingRepo()
	h := command.NewCreateRuleBindingHandler(bindRepo, instRepo)

	pid := uuid.New()
	req := command.CreateRuleBindingRequest{
		RuleInstanceID: instID,
		ScopeType:      vo.ScopeGlobal,
		ScopeID:        &pid,
		Severity:       vo.SeverityBlock,
		EffectiveFrom:  time.Now().UTC(),
		CreatedBy:      uuid.New(),
	}
	_, err := h.Handle(context.Background(), req)
	var invalid *command.ErrInvalidCreateBindingRequest
	if !errors.As(err, &invalid) {
		t.Fatalf("expected ErrInvalidCreateBindingRequest, got %T: %v", err, err)
	}
	if invalid.Field != "scope_id" {
		t.Fatalf("expected field scope_id, got %q", invalid.Field)
	}
}

// ─── deactivate ─────────────────────────────────────────────────────────────

func TestDeactivateRuleBinding_HappyPath(t *testing.T) {
	instRepo, instID := seededInstanceRepo(true)
	bindRepo := newFakeRuleBindingRepo()
	createH := command.NewCreateRuleBindingHandler(bindRepo, instRepo)
	binding, err := createH.Handle(context.Background(), validCreateBindingReq(instID, uuid.New()))
	if err != nil {
		t.Fatalf("seed bind: %v", err)
	}

	deactivateH := command.NewDeactivateRuleBindingHandler(bindRepo)
	if err := deactivateH.Handle(context.Background(), command.DeactivateRuleBindingRequest{BindingID: binding.ID}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bindRepo.deactivate) != 1 || bindRepo.deactivate[0] != binding.ID {
		t.Fatalf("expected Deactivate called with %s, got %v", binding.ID, bindRepo.deactivate)
	}
}

func TestDeactivateRuleBinding_NotFound(t *testing.T) {
	bindRepo := newFakeRuleBindingRepo()
	h := command.NewDeactivateRuleBindingHandler(bindRepo)

	err := h.Handle(context.Background(), command.DeactivateRuleBindingRequest{BindingID: uuid.New()})
	var notFound *domain.ErrBindingNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected ErrBindingNotFound, got %T: %v", err, err)
	}
}
