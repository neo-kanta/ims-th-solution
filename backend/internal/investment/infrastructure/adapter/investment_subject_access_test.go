package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
)

// ── stubs ──────────────────────────────────────────────────────────────────

type stubResearchRepo struct {
	domain.ResearchReportRepository
	report *entity.ResearchReport
	err    error
}

func (r *stubResearchRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.ResearchReport, error) {
	return r.report, r.err
}

type stubDecisionRepo struct {
	domain.DecisionRepository
	decision *entity.Decision
	err      error
}

func (r *stubDecisionRepo) GetByID(_ context.Context, _ uuid.UUID) (*entity.Decision, error) {
	return r.decision, r.err
}

type stubIAM struct {
	ok  bool
	err error
}

func (s stubIAM) HasFunctionPermission(_ context.Context, _, _ string) (bool, error) {
	return true, nil
}
func (s stubIAM) GetAccessibleContracts(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (s stubIAM) HasDataPermission(_ context.Context, _, _ string) (bool, error) {
	return s.ok, s.err
}

// ── helpers ────────────────────────────────────────────────────────────────

func contractPtr(id uuid.UUID) *uuid.UUID { return &id }

// ── tests ──────────────────────────────────────────────────────────────────

func TestInvestmentSubjectAccessor_NilIAM_DeniesAll(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(
		&stubResearchRepo{report: &entity.ResearchReport{ApplicableContractID: contractPtr(uuid.New())}},
		nil,
		nil, // nil IAM
	)
	err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "RESEARCH_REPORT", uuid.New())
	if err == nil {
		t.Fatal("expected deny when IAM is nil, got nil error")
	}
}

func TestInvestmentSubjectAccessor_NilResearchRepo_Denies(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(nil, nil, stubIAM{ok: true})
	err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "RESEARCH_REPORT", uuid.New())
	if err == nil {
		t.Fatal("expected deny when research repo is nil, got nil error")
	}
}

func TestInvestmentSubjectAccessor_ResearchNotFound_Denies(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(
		&stubResearchRepo{report: nil}, // not found
		nil,
		stubIAM{ok: true},
	)
	err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "RESEARCH_REPORT", uuid.New())
	if err == nil {
		t.Fatal("expected deny when research report is nil, got nil error")
	}
}

func TestInvestmentSubjectAccessor_ResearchRepoError_Denies(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(
		&stubResearchRepo{err: errors.New("db error")},
		nil,
		stubIAM{ok: true},
	)
	err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "RESEARCH_REPORT", uuid.New())
	if err == nil {
		t.Fatal("expected deny on repo error, got nil error")
	}
}

func TestInvestmentSubjectAccessor_ResearchNilContract_Allows(t *testing.T) {
	t.Parallel()
	// ApplicableContractID == nil means no data-scope restriction; allow.
	a := NewInvestmentSubjectAccessor(
		&stubResearchRepo{report: &entity.ResearchReport{ApplicableContractID: nil}},
		nil,
		stubIAM{ok: false}, // IAM would deny, but should never be called
	)
	if err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "RESEARCH_REPORT", uuid.New()); err != nil {
		t.Fatalf("expected allow when report has no contract restriction, got %v", err)
	}
}

func TestInvestmentSubjectAccessor_ResearchWithContract_IAMAllows(t *testing.T) {
	t.Parallel()
	cid := uuid.New()
	a := NewInvestmentSubjectAccessor(
		&stubResearchRepo{report: &entity.ResearchReport{ApplicableContractID: &cid}},
		nil,
		stubIAM{ok: true},
	)
	if err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "RESEARCH_REPORT", uuid.New()); err != nil {
		t.Fatalf("expected allow when IAM grants permission, got %v", err)
	}
}

func TestInvestmentSubjectAccessor_ResearchWithContract_IAMDenies(t *testing.T) {
	t.Parallel()
	cid := uuid.New()
	a := NewInvestmentSubjectAccessor(
		&stubResearchRepo{report: &entity.ResearchReport{ApplicableContractID: &cid}},
		nil,
		stubIAM{ok: false},
	)
	err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "RESEARCH_REPORT", uuid.New())
	if err == nil {
		t.Fatal("expected deny when IAM denies permission, got nil error")
	}
}

func TestInvestmentSubjectAccessor_ResearchWithContract_IAMError_Denies(t *testing.T) {
	t.Parallel()
	cid := uuid.New()
	a := NewInvestmentSubjectAccessor(
		&stubResearchRepo{report: &entity.ResearchReport{ApplicableContractID: &cid}},
		nil,
		stubIAM{err: errors.New("iam unavailable")},
	)
	// Fail-closed: IAM transient error must deny, not allow.
	err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "RESEARCH_REPORT", uuid.New())
	if err == nil {
		t.Fatal("expected deny on IAM error (fail-closed), got nil error")
	}
}

func TestInvestmentSubjectAccessor_NilDecisionRepo_Denies(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(nil, nil, stubIAM{ok: true})
	err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "INVESTMENT_DECISION", uuid.New())
	if err == nil {
		t.Fatal("expected deny when decision repo is nil, got nil error")
	}
}

func TestInvestmentSubjectAccessor_DecisionNotFound_Denies(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(
		nil,
		&stubDecisionRepo{decision: nil},
		stubIAM{ok: true},
	)
	err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "INVESTMENT_DECISION", uuid.New())
	if err == nil {
		t.Fatal("expected deny when decision is nil, got nil error")
	}
}

func TestInvestmentSubjectAccessor_DecisionWithContract_IAMAllows(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(
		nil,
		&stubDecisionRepo{decision: &entity.Decision{FundID: uuid.New()}},
		stubIAM{ok: true},
	)
	if err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "INVESTMENT_DECISION", uuid.New()); err != nil {
		t.Fatalf("expected allow for decision with contract + IAM grant, got %v", err)
	}
}

func TestInvestmentSubjectAccessor_UnknownSubjectType_Denies(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(nil, nil, stubIAM{ok: true})
	err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "UNKNOWN_TYPE", uuid.New())
	if err == nil {
		t.Fatal("expected deny for unknown subject type, got nil error")
	}
}

// TestInvestmentSubjectAccessor_Portfolio_Unsupported_Denies guards Option A:
// PORTFOLIO is not registered in main.go because resolveContractID has no PORTFOLIO case.
// If someone adds a RegisterSubjectAccessPort("PORTFOLIO", ...) before adding that case,
// this test catches the mismatch immediately.
func TestInvestmentSubjectAccessor_Portfolio_Unsupported_Denies(t *testing.T) {
	t.Parallel()
	// Fully-wired accessor (non-nil repos and IAM that would grant) — PORTFOLIO must still deny
	// because resolveContractID hits the default arm and returns "unsupported subject type".
	a := NewInvestmentSubjectAccessor(
		&stubResearchRepo{report: &entity.ResearchReport{ApplicableContractID: contractPtr(uuid.New())}},
		&stubDecisionRepo{decision: &entity.Decision{FundID: uuid.New()}},
		stubIAM{ok: true},
	)
	err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "PORTFOLIO", uuid.New())
	if err == nil {
		t.Fatal("PORTFOLIO must be denied — register it in main.go only after adding case \"PORTFOLIO\" to resolveContractID")
	}
}

// CanSubmitApprovalSubject and CanActOnApprovalSubject delegate to the same
// checkDataPermission path; one representative test each verifies the wiring.

func TestInvestmentSubjectAccessor_CanSubmit_NilIAM_Denies(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(
		&stubResearchRepo{report: &entity.ResearchReport{ApplicableContractID: contractPtr(uuid.New())}},
		nil,
		nil,
	)
	if err := a.CanSubmitApprovalSubject(context.Background(), uuid.New(), "RESEARCH_REPORT", uuid.New()); err == nil {
		t.Fatal("expected deny from CanSubmit with nil IAM")
	}
}

func TestInvestmentSubjectAccessor_CanAct_NilIAM_Denies(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(
		&stubResearchRepo{report: &entity.ResearchReport{ApplicableContractID: contractPtr(uuid.New())}},
		nil,
		nil,
	)
	if err := a.CanActOnApprovalSubject(context.Background(), uuid.New(), "RESEARCH_REPORT", uuid.New(), "approve"); err == nil {
		t.Fatal("expected deny from CanAct with nil IAM")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// COMPLIANCE_RELEASE authorization tests (BLOCKER 3)
// ─────────────────────────────────────────────────────────────────────────────

// iamWithFuncPerm is an IAM stub with independent control over data and function permissions.
type iamWithFuncPerm struct {
	dataOK bool
	funcOK bool
	err    error
}

func (s iamWithFuncPerm) HasFunctionPermission(_ context.Context, _, _ string) (bool, error) {
	return s.funcOK, s.err
}
func (s iamWithFuncPerm) HasDataPermission(_ context.Context, _, _ string) (bool, error) {
	return s.dataOK, s.err
}
func (s iamWithFuncPerm) GetAccessibleContracts(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func TestComplianceRelease_CanAct_BothPermissions_Allowed(t *testing.T) {
	t.Parallel()
	a := NewInvestmentSubjectAccessor(
		nil,
		&stubDecisionRepo{decision: &entity.Decision{FundID: uuid.New()}},
		iamWithFuncPerm{dataOK: true, funcOK: true},
	)
	if err := a.CanActOnApprovalSubject(context.Background(), uuid.New(), "COMPLIANCE_RELEASE", uuid.New(), "approve"); err != nil {
		t.Fatalf("expected allow when both data and function permissions granted, got %v", err)
	}
}

func TestComplianceRelease_CanAct_DataPermissionOnly_Denied(t *testing.T) {
	t.Parallel()
	// Has data permission but NOT the compliance release function permission — must deny.
	a := NewInvestmentSubjectAccessor(
		nil,
		&stubDecisionRepo{decision: &entity.Decision{FundID: uuid.New()}},
		iamWithFuncPerm{dataOK: true, funcOK: false},
	)
	if err := a.CanActOnApprovalSubject(context.Background(), uuid.New(), "COMPLIANCE_RELEASE", uuid.New(), "approve"); err == nil {
		t.Fatal("expected deny when function permission is missing for COMPLIANCE_RELEASE")
	}
}

func TestComplianceRelease_CanAct_NoDataPermission_Denied(t *testing.T) {
	t.Parallel()
	// Has function permission but not data permission — must deny.
	a := NewInvestmentSubjectAccessor(
		nil,
		&stubDecisionRepo{decision: &entity.Decision{FundID: uuid.New()}},
		iamWithFuncPerm{dataOK: false, funcOK: true},
	)
	if err := a.CanActOnApprovalSubject(context.Background(), uuid.New(), "COMPLIANCE_RELEASE", uuid.New(), "approve"); err == nil {
		t.Fatal("expected deny when data permission is missing for COMPLIANCE_RELEASE")
	}
}

func TestComplianceRelease_CanView_DataPermissionOnly_Allowed(t *testing.T) {
	t.Parallel()
	// CanView on COMPLIANCE_RELEASE does NOT require the function permission —
	// only act (approve/reject) does. Viewing is gated by data permission only.
	a := NewInvestmentSubjectAccessor(
		nil,
		&stubDecisionRepo{decision: &entity.Decision{FundID: uuid.New()}},
		iamWithFuncPerm{dataOK: true, funcOK: false}, // no function perm
	)
	if err := a.CanViewApprovalSubject(context.Background(), uuid.New(), "COMPLIANCE_RELEASE", uuid.New()); err != nil {
		t.Fatalf("expected allow for CanView COMPLIANCE_RELEASE with data perm only, got %v", err)
	}
}
