package adapter

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/permission"
)

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
	research  domain.ResearchReportRepository
	decisions domain.DecisionRepository
	iam       IAMPermissionPort
}

// NewInvestmentSubjectAccessor wires the accessor.
func NewInvestmentSubjectAccessor(
	research domain.ResearchReportRepository,
	decisions domain.DecisionRepository,
	iam IAMPermissionPort,
) *InvestmentSubjectAccessor {
	return &InvestmentSubjectAccessor{research: research, decisions: decisions, iam: iam}
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
// scope) and the compliance release function permission. Fail-closed on any error.
func (a *InvestmentSubjectAccessor) checkComplianceReleaseActPermission(ctx context.Context, actorID uuid.UUID, subjectID uuid.UUID) error {
	if a.iam == nil {
		return errors.New("not found or not accessible")
	}
	if err := a.checkDataPermission(ctx, actorID, "COMPLIANCE_RELEASE", subjectID); err != nil {
		return err
	}
	ok, err := a.iam.HasFunctionPermission(ctx, actorID.String(), permission.CodeComplianceReleaseApprove)
	if err != nil {
		return errors.New("not found or not accessible")
	}
	if !ok {
		return errors.New("not found or not accessible")
	}
	return nil
}

// checkDataPermission resolves the subject's contract and calls IAM.
// Fail-closed: any nil dependency, missing subject, or permission error denies.
func (a *InvestmentSubjectAccessor) checkDataPermission(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID) error {
	if a.iam == nil {
		return errors.New("not found or not accessible")
	}
	contractID, err := a.resolveContractID(ctx, subjectType, subjectID)
	if err != nil {
		return errors.New("not found or not accessible")
	}
	if contractID == uuid.Nil {
		// Subject has no associated contract — no data-scope restriction applies;
		// the IAM function-permission gate at the route level still applies.
		return nil
	}
	ok, permErr := a.iam.HasDataPermission(ctx, actorID.String(), contractID.String())
	if permErr != nil {
		return errors.New("not found or not accessible")
	}
	if !ok {
		return errors.New("not found or not accessible")
	}
	return nil
}

// resolveContractID returns the contract UUID for the given subject, or
// uuid.Nil when the subject legitimately has no associated contract.
// Returns a non-nil error when a required dependency is nil or the subject
// cannot be found — callers must treat those as deny.
func (a *InvestmentSubjectAccessor) resolveContractID(ctx context.Context, subjectType string, subjectID uuid.UUID) (uuid.UUID, error) {
	switch subjectType {
	case "RESEARCH_REPORT":
		if a.research == nil {
			return uuid.Nil, errors.New("research repository unavailable")
		}
		r, err := a.research.GetByID(ctx, subjectID)
		if err != nil {
			return uuid.Nil, err
		}
		if r == nil {
			return uuid.Nil, errors.New("research report not found")
		}
		if r.ApplicableContractID == nil {
			return uuid.Nil, nil
		}
		return *r.ApplicableContractID, nil

	case "INVESTMENT_DECISION", "COMPLIANCE_RELEASE":
		// COMPLIANCE_RELEASE subjects share the same physical object as the decision —
		// the subjectID is the decision UUID. Resolve the contract via the decision repo.
		if a.decisions == nil {
			return uuid.Nil, errors.New("decision repository unavailable")
		}
		d, err := a.decisions.GetByID(ctx, subjectID)
		if err != nil {
			return uuid.Nil, err
		}
		if d == nil {
			return uuid.Nil, errors.New("investment decision not found")
		}
		return d.FundID, nil

	default:
		return uuid.Nil, errors.New("unsupported subject type: " + subjectType)
	}
}
