package entity

import (
	"time"

	"github.com/google/uuid"
)

// ApprovalStatus tracks the lifecycle of a single approval record.
type ApprovalStatus string

const (
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRevoked  ApprovalStatus = "REVOKED" // set by CANCEL_APPROVAL (Batch 2)
)

// ApprovalRecord captures one manager approval for a workflow day.
//
// Design intent: this entity is the extensibility anchor for future maker-checker
// and multi-approver flows. Today exactly one record is required per day; future
// policy configuration will raise that threshold without changing this schema.
//
// Zero-transaction days: when IsZeroTransaction is true, AttestationReason must
// be at least 30 characters (enforced by the approval policy, not this struct).
type ApprovalRecord struct {
	ID            uuid.UUID
	WorkflowDayID uuid.UUID
	ContractID    uuid.UUID
	BusinessDate  time.Time

	// Approver snapshot — captured at approval time, immutable thereafter
	ApproverID       uuid.UUID
	ApproverUsername string
	ApproverRole     string

	ApprovalStatus ApprovalStatus

	// Zero-transaction attestation fields
	IsZeroTransaction bool
	AttestationReason *string // non-nil when IsZeroTransaction = true

	ApprovedAt time.Time
	RevokedAt  *time.Time // populated by CANCEL_APPROVAL (Batch 2)
	RevokedBy  *uuid.UUID // populated by CANCEL_APPROVAL (Batch 2)

	Notes *string
}
