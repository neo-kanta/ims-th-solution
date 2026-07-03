// Package entity holds the approval module's business entities.
package entity

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// ApprovalGroup is a reusable pool of approvers.
type ApprovalGroup struct {
	ID        uuid.UUID
	GroupCode string
	GroupName string
	Remarks   string
	IsActive  bool
	CreatedBy *uuid.UUID
	CreatedAt time.Time
	UpdatedBy *uuid.UUID
	UpdatedAt time.Time
}

// ApprovalGroupMember is a single member entry inside an approval group.
type ApprovalGroupMember struct {
	ID            uuid.UUID
	GroupID       uuid.UUID
	UserID        uuid.UUID
	PriorityOrder int
	MemberType    vo.GroupMemberType
	Status        vo.GroupMemberStatus
	IsActive      bool
	CreatedBy     *uuid.UUID
	CreatedAt     time.Time
	UpdatedBy     *uuid.UUID
	UpdatedAt     time.Time

	// DisplayName is populated by read queries (joined from iam_users); it is
	// not persisted on this table.
	DisplayName string
}

// Eligible reports whether the member may be assigned an approval task.
func (m ApprovalGroupMember) Eligible() bool {
	return m.IsActive && m.Status == vo.GroupMemberStatusApproved
}

// ApprovalTeam is a per contract/fund approval team.
type ApprovalTeam struct {
	ID                uuid.UUID
	TeamCode          string
	TeamName          string
	Remarks           string
	HasCoManager      bool
	MinRequiredStamps int
	MaxAllowedStamps  int
	IsActive          bool
	CreatedBy         *uuid.UUID
	CreatedAt         time.Time
	UpdatedBy         *uuid.UUID
	UpdatedAt         time.Time
}

// ApprovalTeamContract assigns a team to a contract/fund.
type ApprovalTeamContract struct {
	ID            uuid.UUID
	TeamID        uuid.UUID
	ContractID    uuid.UUID
	EffectiveDate time.Time
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ApprovalTeamMember is a member of an approval team.
type ApprovalTeamMember struct {
	ID            uuid.UUID
	TeamID        uuid.UUID
	UserID        uuid.UUID
	MemberType    vo.TeamMemberType
	PriorityOrder int
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time

	DisplayName string
}

// Eligible reports whether the team member may approve (reviewer agents only).
func (m ApprovalTeamMember) Eligible() bool {
	return m.IsActive && m.MemberType == vo.TeamMemberReviewerAgent
}

// ApprovalProcessConfig defines an approval process and its scope.
type ApprovalProcessConfig struct {
	ID                   uuid.UUID
	ProcessCode          string
	ProcessName          string
	ProcessType          vo.ProcessType
	ContractType         vo.ContractType
	ContractID           *uuid.UUID
	EffectiveDate        time.Time
	IsActive             bool
	GroupApprovalEnabled bool
	RequireTeamApproval  bool
	CreatedBy            *uuid.UUID
	CreatedAt            time.Time
	UpdatedBy            *uuid.UUID
	UpdatedAt            time.Time

	// Stages is populated on detail reads.
	Stages []ApprovalProcessStage
}

// ApprovalProcessStage is one ordered stage of an approval process.
type ApprovalProcessStage struct {
	ID                    uuid.UUID
	ProcessConfigID       uuid.UUID
	StageNumber           int
	StageName             string
	ApproverMode          vo.ApproverMode
	ApproverUserID        *uuid.UUID
	ApprovalGroupID       *uuid.UUID
	RequiredApprovalCount int
	IsFinalStage          bool
	RejectPolicy          string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
