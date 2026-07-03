package entity

import (
	"time"

	"github.com/google/uuid"
)

// ApprovalMode controls how configured approvers are evaluated.
type ApprovalMode string

const (
	// ApprovalModeAnyOf means any one of the listed approvers may approve.
	ApprovalModeAnyOf ApprovalMode = "ANY_OF"
)

// WorkflowApprovalSetting is one approver entry for a given operationType.
// A single operationType may have multiple active settings (any-of semantics).
// Rows are soft-toggled via is_active rather than deleted.
type WorkflowApprovalSetting struct {
	ID                  uuid.UUID
	OperationType       string // canonical API operation name, e.g. "MANAGER_APPROVE"
	ApprovalMode        ApprovalMode
	ApproverAccountCode string
	ApproverUsername    string
	ApproverRole        string
	IsActive            bool
	UpdatedBy           string
	UpdatedAt           time.Time
}
