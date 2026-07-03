package contract

import (
	"context"

	"github.com/google/uuid"
)

// ApprovalSubmission is the cross-module payload used by business modules
// (e.g. investment) to push a subject into the generic approval workflow.
// Keeping it here avoids importing the approval module's internal types.
type ApprovalSubmission struct {
	ProcessType      string // e.g. "INVESTMENT_ANALYSIS_REPORT"
	SubjectType      string // e.g. "RESEARCH_REPORT"
	SubjectID        uuid.UUID
	SubjectTitle     string
	SubjectReference string
	ContractType     string // optional; "FUND" | "DISCRETIONARY" | "COMPANY"
	ContractID       *uuid.UUID
	PortfolioID      *uuid.UUID
	SubmitterID      uuid.UUID
}

// ApprovalSubmissionResult is the minimal result returned to the submitting module.
type ApprovalSubmissionResult struct {
	RequestID     uuid.UUID
	RequestNumber string
	Status        string
}

// ApprovalSubmitter is implemented by the approval module and consumed by
// business modules to create an approval request when a subject is submitted.
type ApprovalSubmitter interface {
	SubmitForApproval(ctx context.Context, in ApprovalSubmission) (*ApprovalSubmissionResult, error)
}

// ApprovalDecision is delivered to a business module when its subject's approval
// request reaches a final decision (approved or rejected).
type ApprovalDecision struct {
	SubjectType string
	SubjectID   uuid.UUID
	RequestID   uuid.UUID
	Approved    bool
	Reason      string
}

// ApprovalSubjectCallback is implemented by a business module (e.g. investment)
// and registered with the approval module so the underlying business object can
// be updated when its approval reaches a final decision.
type ApprovalSubjectCallback interface {
	OnApprovalDecision(ctx context.Context, decision ApprovalDecision) error
}

// ApproverInfo is a lightweight approver identity used for grid display.
type ApproverInfo struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
}

// ApprovalStageInfo summarises the current approval state of a subject for
// read-side enrichment in business modules. The approval engine remains the
// source of truth; business modules use this to label the lifecycle in their
// own response DTOs (e.g. the research-report "derived_review_stage" badge).
//
// CurrentStageNumber is 1-based and only meaningful when Status indicates an
// in-flight request (e.g. "PENDING_APPROVAL"). TotalStages reflects the
// configured stage count for the resolved process configuration at the time
// the request was created.
type ApprovalStageInfo struct {
	// RequestID is the approval request UUID. Zero when no active or
	// historical request exists for the subject.
	RequestID uuid.UUID
	// RequestNumber is the human-readable request number.
	RequestNumber string
	// Status is the request status string, e.g. "DRAFT", "PENDING_APPROVAL",
	// "APPROVED", "REJECTED", "CANCELLED", "WITHDRAWN".
	Status string
	// CurrentStageNumber is 1-based; 0 when the request has no stages.
	CurrentStageNumber int
	// TotalStages is the number of configured stages on the process config.
	TotalStages int
	// CurrentApprovers lists the pending-task assignees at the current stage.
	// Populated by GetApprovalStage for display in approval grids.
	CurrentApprovers []ApproverInfo
	// PreviousApprovers lists who acted on the stage prior to the current one.
	PreviousApprovers []ApproverInfo
}

// ApprovalStatusProvider is implemented by the approval module to expose the
// current approval state of a subject to business modules. Returning (nil, nil)
// means "no approval request exists for this subject"; the consumer should
// treat that as a NOT_SUBMITTED-style display state.
type ApprovalStatusProvider interface {
	GetApprovalStage(ctx context.Context, subjectType string, subjectID uuid.UUID) (*ApprovalStageInfo, error)
}

// ApprovalBatchActor is implemented by the approval module so business modules
// (e.g. investment batch-approve) can approve or reject by approval request ID
// rather than by task ID. The implementation resolves the actor's pending task
// for the request and delegates to the standard approve/reject path.
type ApprovalBatchActor interface {
	ApproveByRequest(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID, comment string) error
	RejectByRequest(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID, reason string) error
}

// ApprovalCanceller is implemented by the approval module so business modules
// can cancel the active approval request for a subject when the subject itself
// is cancelled or withdrawn. Returns nil when no active request exists (no-op).
type ApprovalCanceller interface {
	CancelApprovalBySubject(ctx context.Context, subjectType string, subjectID uuid.UUID, actorID uuid.UUID) error
}

// ApprovalSubjectValidator is implemented by business modules to let the
// approval engine verify that the underlying subject is still in an approvable
// state before recording an approve or reject action.
type ApprovalSubjectValidator interface {
	ValidateSubjectApprovable(ctx context.Context, subjectID uuid.UUID) error
}

// SubjectAccessor is implemented by business modules and registered with the
// approval module (one per subject type) so the approval engine can enforce
// object-level access without inspecting business-specific keys (contract_id,
// fund_id, portfolio_id). The implementation may use any data within its own
// bounded context to make the authorisation decision.
//
// All methods return nil for allow, non-nil for deny. Register via
// Module.RegisterSubjectAccessPort in cmd/server/main.go after both modules
// are constructed.
type SubjectAccessor interface {
	CanViewApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID) error
	CanSubmitApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID) error
	CanActOnApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType string, subjectID uuid.UUID, action string) error
}
