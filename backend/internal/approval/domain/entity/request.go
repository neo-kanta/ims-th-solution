package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// ApprovalRequest is the approval instance for a single submitted business object.
type ApprovalRequest struct {
	ID                 uuid.UUID
	RequestNumber      string
	ProcessType        vo.ProcessType
	ProcessConfigID    *uuid.UUID
	SubjectType        vo.SubjectType
	SubjectID          uuid.UUID
	SubjectTitle       string
	SubjectReference   string
	ContractID         *uuid.UUID
	PortfolioID        *uuid.UUID
	SubmitterID        uuid.UUID
	SubmittedAt        *time.Time
	CurrentStageNumber int
	Status             vo.RequestStatus
	FinalDecisionBy    *uuid.UUID
	FinalDecisionAt    *time.Time
	RejectionReason    string
	CreatedAt          time.Time
	UpdatedAt          time.Time

	// SubmitterName is populated by read queries (joined from iam_users).
	SubmitterName string
}

// ApprovalTask is a per-approver work item for an approval request stage.
type ApprovalTask struct {
	ID                  uuid.UUID
	ApprovalRequestID   uuid.UUID
	StageNumber         int
	AssignedUserID      *uuid.UUID
	AssignedGroupID     *uuid.UUID
	AssignedTeamID      *uuid.UUID
	DelegatedFromUserID *uuid.UUID
	Status              vo.TaskStatus
	ActedBy             *uuid.UUID
	ActedAt             *time.Time
	ActionComment       string
	IsDelegatedAction   bool
	DueAt               *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time

	// Read-projection helpers (joined from iam_users / requests).
	AssignedUserName string
}

// ApprovalEvent is an immutable timeline / audit entry for a request.
type ApprovalEvent struct {
	ID                  uuid.UUID
	ApprovalRequestID   uuid.UUID
	EventType           vo.EventType
	StageNumber         *int
	ActorUserID         *uuid.UUID
	DelegatedFromUserID *uuid.UUID
	Comment             string
	Metadata            map[string]any
	CreatedAt           time.Time

	// ActorName is populated by read queries.
	ActorName string
}

// ApprovalSignatureRecord is a digital stamp/signature record.
type ApprovalSignatureRecord struct {
	ID                uuid.UUID
	ApprovalRequestID uuid.UUID
	StageNumber       int
	SignerUserID      uuid.UUID
	SignerDisplayName string
	SignerTitle       *string
	SignedAt          time.Time
	IsProxySignature  bool
	ProxyForUserID    *uuid.UUID
	SignatureLabel    vo.SignatureLabel
	CreatedAt         time.Time
}
