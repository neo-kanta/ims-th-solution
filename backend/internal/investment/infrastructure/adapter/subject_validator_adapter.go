package adapter

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// ResearchReportSubjectValidator implements contract.ApprovalSubjectValidator
// for the RESEARCH_REPORT subject type. It verifies the report is still in
// SUBMITTED state and has not been deleted or invalidated since the approval
// request was created.
type ResearchReportSubjectValidator struct {
	repo domain.ResearchReportRepository
}

// NewResearchReportSubjectValidator wires the validator to the report repository.
func NewResearchReportSubjectValidator(repo domain.ResearchReportRepository) *ResearchReportSubjectValidator {
	return &ResearchReportSubjectValidator{repo: repo}
}

// ValidateSubjectApprovable implements contract.ApprovalSubjectValidator.
func (v *ResearchReportSubjectValidator) ValidateSubjectApprovable(ctx context.Context, subjectID uuid.UUID) error {
	r, err := v.repo.GetByID(ctx, subjectID)
	if err != nil {
		return err
	}
	if r == nil || r.IsDeleted() || r.IsInvalidated() || r.ReviewStatus != vo.ReviewStatusSubmitted {
		return errors.New("research report is no longer in a submittable state")
	}
	return nil
}

// DecisionSubjectValidator implements contract.ApprovalSubjectValidator for
// the INVESTMENT_DECISION subject type. It verifies the decision is still in
// PENDING_APPROVAL state since the approval request was created.
type DecisionSubjectValidator struct {
	repo domain.DecisionRepository
}

// NewDecisionSubjectValidator wires the validator to the decision repository.
func NewDecisionSubjectValidator(repo domain.DecisionRepository) *DecisionSubjectValidator {
	return &DecisionSubjectValidator{repo: repo}
}

// ValidateSubjectApprovable implements contract.ApprovalSubjectValidator.
func (v *DecisionSubjectValidator) ValidateSubjectApprovable(ctx context.Context, subjectID uuid.UUID) error {
	d, err := v.repo.GetByID(ctx, subjectID)
	if err != nil {
		return err
	}
	if d == nil || d.Status != vo.DecisionLifecyclePendingApproval {
		return errors.New("investment decision is no longer pending approval")
	}
	return nil
}

// CashRequestSubjectValidator implements contract.ApprovalSubjectValidator for
// the CASH_TRANSACTION subject type. It verifies the cash request is still
// PENDING — an approver cannot act on a request the submitter already cancelled
// (or one that was materialized concurrently).
type CashRequestSubjectValidator struct {
	repo domain.PortfolioCashRequestRepository
}

// NewCashRequestSubjectValidator wires the validator.
func NewCashRequestSubjectValidator(repo domain.PortfolioCashRequestRepository) *CashRequestSubjectValidator {
	return &CashRequestSubjectValidator{repo: repo}
}

// ValidateSubjectApprovable implements contract.ApprovalSubjectValidator.
func (v *CashRequestSubjectValidator) ValidateSubjectApprovable(ctx context.Context, subjectID uuid.UUID) error {
	c, err := v.repo.GetByID(ctx, subjectID)
	if err != nil {
		return err
	}
	if c == nil || c.Status != vo.CashRequestStatusPending {
		return errors.New("cash request is no longer pending approval")
	}
	return nil
}

// ComplianceReleaseSubjectValidator implements contract.ApprovalSubjectValidator
// for the COMPLIANCE_RELEASE subject type. It verifies the decision is still in
// PENDING_COMPLIANCE_RELEASE state — not CANCELLED or advanced concurrently.
type ComplianceReleaseSubjectValidator struct {
	repo domain.DecisionRepository
}

// NewComplianceReleaseSubjectValidator wires the validator.
func NewComplianceReleaseSubjectValidator(repo domain.DecisionRepository) *ComplianceReleaseSubjectValidator {
	return &ComplianceReleaseSubjectValidator{repo: repo}
}

// ValidateSubjectApprovable implements contract.ApprovalSubjectValidator.
func (v *ComplianceReleaseSubjectValidator) ValidateSubjectApprovable(ctx context.Context, subjectID uuid.UUID) error {
	d, err := v.repo.GetByID(ctx, subjectID)
	if err != nil {
		return err
	}
	if d == nil || d.Status != vo.DecisionLifecyclePendingComplianceRelease {
		return errors.New("decision is no longer pending compliance release")
	}
	return nil
}
