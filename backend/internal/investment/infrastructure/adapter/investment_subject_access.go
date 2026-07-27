package adapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/permission"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// deniedAccess builds a GENUINE access-denial error wrapping the cross-module
// contract.ErrSubjectAccessDenied sentinel. The approval engine maps an error
// that wraps this sentinel to 403 Forbidden. The message is intentionally vague
// to avoid leaking subject existence. Infrastructure failures must NOT use this
// helper — they return the raw underlying error so the engine can surface a 5xx.
func deniedAccess() error {
	return fmt.Errorf("not found or not accessible: %w", contract.ErrSubjectAccessDenied)
}

// InvestmentSubjectAccessor implements contract.SubjectAccessor for the three
// investment-owned subject types: RESEARCH_REPORT, INVESTMENT_DECISION, and
// COMPLIANCE_RELEASE. For COMPLIANCE_RELEASE the subject ID is the decision ID;
// the same contract-level data-permission gate applies so only actors with data
// access to the decision's fund can act on the compliance release approval.
//
// It resolves the subject's associated contract and delegates the access
// decision to the IAM data-permission gate (HasDataPermission). The approval
// engine calls these methods without knowing anything about fund_id,
// portfolio_id, or other business-specific keys.
type InvestmentSubjectAccessor struct {
	research     domain.ResearchReportRepository
	decisions    domain.DecisionRepository
	cashRequests domain.PortfolioCashRequestRepository
	iam          IAMPermissionPort
}

// NewInvestmentSubjectAccessor wires the accessor. The CASH_TRANSACTION
// repository is injected separately via SetCashRequestRepository to keep this
// constructor's signature stable for existing callers/tests.
func NewInvestmentSubjectAccessor(
	research domain.ResearchReportRepository,
	decisions domain.DecisionRepository,
	iam IAMPermissionPort,
) *InvestmentSubjectAccessor {
	return &InvestmentSubjectAccessor{research: research, decisions: decisions, iam: iam}
}

// SetCashRequestRepository injects the cash-request repository used to resolve
// data-scope for CASH_TRANSACTION approval subjects.
func (a *InvestmentSubjectAccessor) SetCashRequestRepository(r domain.PortfolioCashRequestRepository) {
	if a != nil {
		a.cashRequests = r
	}
}

// CanViewApprovalSubject allows viewing when the actor has data-permission on
// the subject's associated contract. Returns forbidden for unknown subject types.
func (a *InvestmentSubjectAccessor) CanViewApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID) error {
	return a.checkDataPermission(ctx, actorID, subjectType, subjectID)
}

// CanSubmitApprovalSubject allows submitting when the actor has data-permission
// on the subject's associated contract.
func (a *InvestmentSubjectAccessor) CanSubmitApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID) error {
	return a.checkDataPermission(ctx, actorID, subjectType, subjectID)
}

// CanActOnApprovalSubject allows approve/reject/revoke/cancel. For
// COMPLIANCE_RELEASE, the actor must hold both data permission on the decision's
// contract AND the INVESTMENT_COMPLIANCE_RELEASE_APPROVE function permission.
// Other subject types require data permission only.
func (a *InvestmentSubjectAccessor) CanActOnApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID, _ string) error {
	if subjectType == "COMPLIANCE_RELEASE" {
		return a.checkComplianceReleaseActPermission(ctx, actorID, subjectID)
	}
	return a.checkDataPermission(ctx, actorID, subjectType, subjectID)
}

// checkComplianceReleaseActPermission enforces both data permission (contract
// scope) and the compliance release function permission. Fail-closed on any
// error, but distinguishes a genuine denial (wraps ErrSubjectAccessDenied → 403)
// from an infrastructure failure (raw error → 5xx).
func (a *InvestmentSubjectAccessor) checkComplianceReleaseActPermission(ctx context.Context, actorID uuid.UUID, subjectID uuid.UUID) error {
	if a.iam == nil {
		// Missing required dependency is a wiring/infrastructure failure, not a
		// user-level denial: return a raw error so it surfaces as 5xx.
		return errors.New("investment subject accessor: IAM permission port not configured")
	}
	if err := a.checkDataPermission(ctx, actorID, "COMPLIANCE_RELEASE", subjectID); err != nil {
		// Already classified (sentinel-wrapped denial, or raw infrastructure error).
		return err
	}
	ok, err := a.iam.HasFunctionPermission(ctx, actorID.String(), permission.CodeComplianceReleaseApprove)
	if err != nil {
		return err // infrastructure failure — surface raw so it maps to 5xx
	}
	if !ok {
		return deniedAccess() // genuine denial — actor lacks the function permission
	}
	return nil
}

// checkDataPermission resolves the subject's contract and calls IAM.
// Fail-closed on every error path, but classifies the outcome: a genuine denial
// (subject not found, unsupported type, or IAM returning !ok) wraps
// ErrSubjectAccessDenied (→ 403); an infrastructure failure (nil dependency, or
// a non-nil error from the repository or IAM) returns the raw error (→ 5xx).
func (a *InvestmentSubjectAccessor) checkDataPermission(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID) error {
	if a.iam == nil {
		return errors.New("investment subject accessor: IAM permission port not configured")
	}
	contractID, err := a.resolveContractID(ctx, subjectType, subjectID)
	if err != nil {
		// resolveContractID already classifies: sentinel-wrapped denial or raw
		// infrastructure error. Propagate unchanged.
		return err
	}
	if contractID == uuid.Nil {
		// Subject has no associated contract — no data-scope restriction applies;
		// the IAM function-permission gate at the route level still applies.
		return nil
	}
	ok, permErr := a.iam.HasDataPermission(ctx, actorID.String(), contractID.String())
	if permErr != nil {
		return permErr // infrastructure failure — surface raw so it maps to 5xx
	}
	if !ok {
		return deniedAccess() // genuine denial — actor lacks data permission
	}
	return nil
}

// resolveContractID returns the contract UUID for the given subject, or
// uuid.Nil when the subject legitimately has no associated contract.
//
// Error classification for callers:
//   - a nil required dependency, or a non-nil error from a repository GetByID,
//     is an infrastructure failure and is returned RAW (→ 5xx).
//   - a subject that cannot be found (GetByID returns (nil, nil)) and an
//     unsupported subject type are genuine denials and wrap
//     ErrSubjectAccessDenied via deniedAccess() (→ 403).
func (a *InvestmentSubjectAccessor) resolveContractID(ctx context.Context, subjectType string, subjectID uuid.UUID) (uuid.UUID, error) {
	switch subjectType {
	case "RESEARCH_REPORT":
		if a.research == nil {
			return uuid.Nil, errors.New("investment subject accessor: research repository not configured")
		}
		r, err := a.research.GetByID(ctx, subjectID)
		if err != nil {
			return uuid.Nil, err // infrastructure failure
		}
		if r == nil {
			return uuid.Nil, deniedAccess() // subject not found — deny
		}
		if r.ApplicableContractID == nil {
			return uuid.Nil, nil
		}
		return *r.ApplicableContractID, nil

	case "INVESTMENT_DECISION", "COMPLIANCE_RELEASE":
		// COMPLIANCE_RELEASE subjects share the same physical object as the decision —
		// the subjectID is the decision UUID. Resolve the contract via the decision repo.
		if a.decisions == nil {
			return uuid.Nil, errors.New("investment subject accessor: decision repository not configured")
		}
		d, err := a.decisions.GetByID(ctx, subjectID)
		if err != nil {
			return uuid.Nil, err // infrastructure failure
		}
		if d == nil {
			return uuid.Nil, deniedAccess() // subject not found — deny
		}
		// A fund-less decision has no fund to scope against — fall back to its
		// own portfolio's id as the data-permission scope key, matching every
		// other fund-less scope check in this module (portfolioScopeID in
		// transport/handler/investment_handler.go). This keeps access
		// controlled rather than falling open to "no restriction."
		if d.FundID == nil {
			return d.PortfolioID, nil
		}
		return *d.FundID, nil

	case "CASH_TRANSACTION":
		// The subjectID is the cash request UUID. Resolve its portfolio → fund
		// data-scope key, falling back to the portfolio id for a fund-less
		// portfolio (same fund-optional convention as decisions above).
		if a.cashRequests == nil {
			return uuid.Nil, errors.New("investment subject accessor: cash request repository not configured")
		}
		c, err := a.cashRequests.GetByID(ctx, subjectID)
		if err != nil {
			return uuid.Nil, err // infrastructure failure
		}
		if c == nil {
			return uuid.Nil, deniedAccess() // subject not found — deny
		}
		if c.FundID == nil {
			return c.PortfolioID, nil
		}
		return *c.FundID, nil

	default:
		return uuid.Nil, deniedAccess() // unsupported subject type — deny
	}
}
