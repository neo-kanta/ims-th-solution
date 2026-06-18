package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// StageSnapshot captures the process-stage configuration at the moment a
// request is submitted. In-flight approval uses this snapshot so that
// subsequent config edits cannot retroactively change an active request.
type StageSnapshot struct {
	StageNumber           int    `json:"stage_number"`
	StageName             string `json:"stage_name"`
	ApproverMode          string `json:"approver_mode"`
	ApproverUserID        string `json:"approver_user_id,omitempty"`
	ApprovalGroupID       string `json:"approval_group_id,omitempty"`
	RequiredApprovalCount int    `json:"required_approval_count"`
	IsFinalStage          bool   `json:"is_final_stage"`
	RejectPolicy          string `json:"reject_policy"`
}

// ToProcessStage converts a snapshot entry back to an ApprovalProcessStage for
// use by runtime helpers. ID and timestamps are zeroed — only routing fields matter.
func (s StageSnapshot) ToProcessStage() ApprovalProcessStage {
	st := ApprovalProcessStage{
		StageNumber:           s.StageNumber,
		StageName:             s.StageName,
		ApproverMode:          vo.ApproverMode(s.ApproverMode),
		RequiredApprovalCount: s.RequiredApprovalCount,
		IsFinalStage:          s.IsFinalStage,
		RejectPolicy:          s.RejectPolicy,
	}
	if s.ApproverUserID != "" {
		if uid, err := uuid.Parse(s.ApproverUserID); err == nil {
			st.ApproverUserID = &uid
		}
	}
	if s.ApprovalGroupID != "" {
		if uid, err := uuid.Parse(s.ApprovalGroupID); err == nil {
			st.ApprovalGroupID = &uid
		}
	}
	return st
}

// StageSnapshotFromConfig creates a snapshot from the provided process stages.
func StageSnapshotFromConfig(stages []ApprovalProcessStage) []StageSnapshot {
	out := make([]StageSnapshot, 0, len(stages))
	for _, s := range stages {
		ss := StageSnapshot{
			StageNumber:           s.StageNumber,
			StageName:             s.StageName,
			ApproverMode:          string(s.ApproverMode),
			RequiredApprovalCount: s.RequiredApprovalCount,
			IsFinalStage:          s.IsFinalStage,
			RejectPolicy:          s.RejectPolicy,
		}
		if s.ApproverUserID != nil {
			ss.ApproverUserID = s.ApproverUserID.String()
		}
		if s.ApprovalGroupID != nil {
			ss.ApprovalGroupID = s.ApprovalGroupID.String()
		}
		out = append(out, ss)
	}
	return out
}

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

	// ConfigSnapshot is the process stages frozen at submit time.
	// Used instead of the live config so edits cannot affect in-flight requests.
	ConfigSnapshot []StageSnapshot
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
