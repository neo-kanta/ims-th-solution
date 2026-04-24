package command_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	// Pull in a registered rule type so the registry has something to look up.
	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/credit"
)

// ─── fake repo ──────────────────────────────────────────────────────────────

// fakeRuleInstanceRepo is the minimal slice of the repository the
// create-instance handler interacts with. It records what was persisted so
// tests can assert on the end state (v1 inserted, current_version bumped,
// etc.).
type fakeRuleInstanceRepo struct {
	instances   map[uuid.UUID]entity.RuleInstance
	versions    map[uuid.UUID][]entity.RuleInstanceVersion
	createErr   error
	versionErr  error
	updateCalls []uuid.UUID
}

func newFakeRuleInstanceRepo() *fakeRuleInstanceRepo {
	return &fakeRuleInstanceRepo{
		instances: map[uuid.UUID]entity.RuleInstance{},
		versions:  map[uuid.UUID][]entity.RuleInstanceVersion{},
	}
}

func (f *fakeRuleInstanceRepo) Create(_ context.Context, inst *entity.RuleInstance) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.instances[inst.ID] = *inst
	return nil
}

func (f *fakeRuleInstanceRepo) GetByID(_ context.Context, id uuid.UUID) (*entity.RuleInstance, error) {
	if inst, ok := f.instances[id]; ok {
		return &inst, nil
	}
	return nil, nil
}

func (f *fakeRuleInstanceRepo) List(_ context.Context, _ domain.RuleInstanceFilter) ([]entity.RuleInstance, int64, error) {
	out := make([]entity.RuleInstance, 0, len(f.instances))
	for _, v := range f.instances {
		out = append(out, v)
	}
	return out, int64(len(out)), nil
}

func (f *fakeRuleInstanceRepo) UpdateActive(_ context.Context, id uuid.UUID, active bool) error {
	f.updateCalls = append(f.updateCalls, id)
	if inst, ok := f.instances[id]; ok {
		inst.IsActive = active
		f.instances[id] = inst
	}
	return nil
}

func (f *fakeRuleInstanceRepo) CreateVersion(_ context.Context, v *entity.RuleInstanceVersion) error {
	if f.versionErr != nil {
		return f.versionErr
	}
	f.versions[v.RuleInstanceID] = append(f.versions[v.RuleInstanceID], *v)
	if inst, ok := f.instances[v.RuleInstanceID]; ok {
		inst.CurrentVersion = v.VersionNumber
		f.instances[v.RuleInstanceID] = inst
	}
	return nil
}

func (f *fakeRuleInstanceRepo) GetCurrentVersion(_ context.Context, instanceID uuid.UUID) (*entity.RuleInstanceVersion, error) {
	versions := f.versions[instanceID]
	if len(versions) == 0 {
		return nil, nil
	}
	last := versions[len(versions)-1]
	return &last, nil
}

func (f *fakeRuleInstanceRepo) ListVersions(_ context.Context, instanceID uuid.UUID) ([]entity.RuleInstanceVersion, error) {
	return f.versions[instanceID], nil
}

// sanity: interface satisfaction
var _ domain.RuleInstanceRepository = (*fakeRuleInstanceRepo)(nil)

// ─── helpers ────────────────────────────────────────────────────────────────

func validCreateReq() command.CreateRuleInstanceRequest {
	return command.CreateRuleInstanceRequest{
		RuleTypeID:    "credit.min_rating",
		Name:          "Thai corporate IG floor",
		Description:   "Investment-grade floor for Thai corporates",
		Parameters:    json.RawMessage(`{"min_rating":"BBB-"}`),
		EffectiveFrom: time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC),
		IsActive:      true,
		CreatedBy:     uuid.New(),
	}
}

// ─── happy path ─────────────────────────────────────────────────────────────

func TestCreateRuleInstance_HappyPath(t *testing.T) {
	repo := newFakeRuleInstanceRepo()
	h := command.NewCreateRuleInstanceHandler(repo, spi.GlobalRegistry())

	res, err := h.Handle(context.Background(), validCreateReq())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Instance.ID == uuid.Nil {
		t.Fatalf("expected non-nil instance ID")
	}
	if res.Instance.CurrentVersion != 1 {
		t.Fatalf("expected current_version=1, got %d", res.Instance.CurrentVersion)
	}
	if res.Version.VersionNumber != 1 {
		t.Fatalf("expected version_number=1, got %d", res.Version.VersionNumber)
	}
	if got := string(res.Version.Parameters); got != `{"min_rating":"BBB-"}` {
		t.Fatalf("parameters not persisted verbatim: %s", got)
	}
	if res.Version.ChangeReason != "initial creation" {
		t.Fatalf("expected default change reason, got %q", res.Version.ChangeReason)
	}
	// Instance + v1 both persisted.
	if len(repo.instances) != 1 {
		t.Fatalf("expected 1 instance persisted, got %d", len(repo.instances))
	}
	if len(repo.versions[res.Instance.ID]) != 1 {
		t.Fatalf("expected 1 version persisted, got %d", len(repo.versions[res.Instance.ID]))
	}
}

// ─── validation ─────────────────────────────────────────────────────────────

func TestCreateRuleInstance_Validation(t *testing.T) {
	repo := newFakeRuleInstanceRepo()
	h := command.NewCreateRuleInstanceHandler(repo, spi.GlobalRegistry())
	base := validCreateReq()

	cases := []struct {
		name  string
		mut   func(*command.CreateRuleInstanceRequest)
		field string
	}{
		{"missing rule_type_id", func(r *command.CreateRuleInstanceRequest) { r.RuleTypeID = "" }, "rule_type_id"},
		{"missing name", func(r *command.CreateRuleInstanceRequest) { r.Name = "" }, "name"},
		{"missing parameters", func(r *command.CreateRuleInstanceRequest) { r.Parameters = nil }, "parameters"},
		{"missing effective_from", func(r *command.CreateRuleInstanceRequest) { r.EffectiveFrom = time.Time{} }, "effective_from"},
		{"missing created_by", func(r *command.CreateRuleInstanceRequest) { r.CreatedBy = uuid.Nil }, "created_by"},
		{"effective_to before effective_from", func(r *command.CreateRuleInstanceRequest) {
			t2 := r.EffectiveFrom.Add(-24 * time.Hour)
			r.EffectiveTo = &t2
		}, "effective_to"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := base
			tc.mut(&req)
			_, err := h.Handle(context.Background(), req)
			var invalid *command.ErrInvalidCreateRuleRequest
			if !errors.As(err, &invalid) {
				t.Fatalf("expected ErrInvalidCreateRuleRequest, got %T: %v", err, err)
			}
			if invalid.Field != tc.field {
				t.Fatalf("expected field %q, got %q", tc.field, invalid.Field)
			}
		})
	}
}

// ─── rule type resolution ───────────────────────────────────────────────────

func TestCreateRuleInstance_UnknownRuleType_Returns404Domain(t *testing.T) {
	repo := newFakeRuleInstanceRepo()
	h := command.NewCreateRuleInstanceHandler(repo, spi.GlobalRegistry())
	req := validCreateReq()
	req.RuleTypeID = "does.not.exist"
	_, err := h.Handle(context.Background(), req)
	var notFound *domain.ErrRuleTypeNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("expected ErrRuleTypeNotFound, got %T: %v", err, err)
	}
	if notFound.TypeID != "does.not.exist" {
		t.Fatalf("unexpected TypeID in error: %s", notFound.TypeID)
	}
}

// ─── parameter schema ───────────────────────────────────────────────────────

func TestCreateRuleInstance_InvalidParametersJSON_Blocks(t *testing.T) {
	repo := newFakeRuleInstanceRepo()
	h := command.NewCreateRuleInstanceHandler(repo, spi.GlobalRegistry())
	req := validCreateReq()
	req.Parameters = json.RawMessage(`{this is not json`)
	_, err := h.Handle(context.Background(), req)
	var paramErr *domain.ErrParameterValidation
	if !errors.As(err, &paramErr) {
		t.Fatalf("expected ErrParameterValidation, got %T: %v", err, err)
	}
	if paramErr.RuleTypeID != "credit.min_rating" {
		t.Fatalf("unexpected rule type on error: %s", paramErr.RuleTypeID)
	}
}

// ─── persistence failure paths ──────────────────────────────────────────────

func TestCreateRuleInstance_InstanceInsertFailsClean(t *testing.T) {
	repo := newFakeRuleInstanceRepo()
	repo.createErr = errors.New("pgx: unique violation")
	h := command.NewCreateRuleInstanceHandler(repo, spi.GlobalRegistry())

	_, err := h.Handle(context.Background(), validCreateReq())
	if err == nil {
		t.Fatal("expected error from instance insert")
	}
	if len(repo.versions) != 0 {
		t.Fatal("no version should be written when instance insert fails")
	}
}

func TestCreateRuleInstance_VersionInsertFails_InstanceDeactivated(t *testing.T) {
	repo := newFakeRuleInstanceRepo()
	repo.versionErr = errors.New("boom")
	h := command.NewCreateRuleInstanceHandler(repo, spi.GlobalRegistry())

	_, err := h.Handle(context.Background(), validCreateReq())
	if err == nil {
		t.Fatal("expected error from version insert")
	}
	// Instance was inserted then deactivated.
	if len(repo.instances) != 1 {
		t.Fatalf("expected orphan instance row, got %d", len(repo.instances))
	}
	if len(repo.updateCalls) != 1 {
		t.Fatal("expected exactly one UpdateActive(deactivate) call on the orphan")
	}
	// And deactivation actually flipped the flag.
	for _, inst := range repo.instances {
		if inst.IsActive {
			t.Fatalf("orphan instance must be deactivated, got IsActive=true")
		}
	}
}

// ─── defaults ───────────────────────────────────────────────────────────────

func TestCreateRuleInstance_CustomChangeReasonPreserved(t *testing.T) {
	repo := newFakeRuleInstanceRepo()
	h := command.NewCreateRuleInstanceHandler(repo, spi.GlobalRegistry())
	req := validCreateReq()
	req.ChangeReason = "seed from regulatory filing Q2 2026"
	res, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Version.ChangeReason != "seed from regulatory filing Q2 2026" {
		t.Fatalf("change reason not preserved: %q", res.Version.ChangeReason)
	}
}

func TestCreateRuleInstance_IndefiniteWindow(t *testing.T) {
	repo := newFakeRuleInstanceRepo()
	h := command.NewCreateRuleInstanceHandler(repo, spi.GlobalRegistry())
	req := validCreateReq() // no EffectiveTo
	res, err := h.Handle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Instance.EffectiveWindow.ValidTo != nil {
		t.Fatalf("expected indefinite window, got ValidTo=%v", res.Instance.EffectiveWindow.ValidTo)
	}
}
