package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// PermissionPort is the subset of IAM consumed by the approval module for
// function-permission and contract data-scope checks. It is satisfied by the
// IAM module (whose method signatures already match).
type PermissionPort interface {
	HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error)
	HasDataPermission(ctx context.Context, userID string, scopeID string) (bool, error)
}

// UserInfo is a minimal user projection used for signature display and inbox.
type UserInfo struct {
	ID          uuid.UUID
	DisplayName string
	Title       string
}

// UserDirectory resolves user display information for signature/stamp records.
// Implemented by reading iam_users (read-only) without importing IAM internals.
type UserDirectory interface {
	GetUser(ctx context.Context, id uuid.UUID) (*UserInfo, error)
}

// LeaveChecker reports whether a user is on leave on a given date. The approval
// engine uses it to skip on-leave approvers and to trigger delegation.
//
// When no leave module is wired, NopLeaveChecker reports "never on leave".
type LeaveChecker interface {
	IsOnLeave(ctx context.Context, userID uuid.UUID, asOf time.Time) (bool, error)
}

// Delegate describes a resolved proxy approver for an original approver.
type Delegate struct {
	DelegateUserID uuid.UUID
}

// DelegateResolver resolves a valid delegate for an original approver on a
// contract at a point in time. Returns (nil, nil) when no delegate exists.
//
// When no leave/delegation module is wired, NopDelegateResolver always returns
// (nil, nil) — approval still works; the original approver remains responsible.
type DelegateResolver interface {
	ResolveDelegate(ctx context.Context, originalUserID uuid.UUID, contractID *uuid.UUID, asOf time.Time) (*Delegate, error)
}

// Notifier emits approval notifications. When no notification backend is wired,
// NopNotifier discards calls; approval events still capture the timeline.
type Notifier interface {
	NotifyApprovalTaskCreated(ctx context.Context, req *entity.ApprovalRequest, task *entity.ApprovalTask)
	NotifyApprovalCompleted(ctx context.Context, req *entity.ApprovalRequest)
	NotifyApprovalRejected(ctx context.Context, req *entity.ApprovalRequest)
}

// AuditPort records cross-system audit events through the shared audit recorder.
type AuditPort interface {
	Record(ctx context.Context, actorID *uuid.UUID, eventType, targetType, targetID string, metadata map[string]any)
}

// SubjectSync notifies the owning business module of a final approval decision
// so it can update the underlying business object (e.g. research report status).
// Implemented per subject type by the owning module and registered at wire-up.
type SubjectSync interface {
	OnApproved(ctx context.Context, subjectType vo.SubjectType, subjectID uuid.UUID, requestID uuid.UUID) error
	OnRejected(ctx context.Context, subjectType vo.SubjectType, subjectID uuid.UUID, requestID uuid.UUID, reason string) error
}

// SubjectValidator is called inside an approve/reject transaction to verify
// the underlying business object is still in a state that permits approval.
// Register one per subject type via RegisterSubjectValidator; the approval
// engine calls it before recording the action so a concurrent cancellation or
// invalidation is detected before the approval is committed.
type SubjectValidator interface {
	ValidateSubjectApprovable(ctx context.Context, subjectID uuid.UUID) error
}

// NopSubjectValidator always reports the subject as approvable.
type NopSubjectValidator struct{}

// ValidateSubjectApprovable is a no-op.
func (NopSubjectValidator) ValidateSubjectApprovable(context.Context, uuid.UUID) error { return nil }

// ─────────────────────────────────────────────────────────────────────────────
// Safe no-op default implementations
// ─────────────────────────────────────────────────────────────────────────────

// NopLeaveChecker reports that no user is ever on leave.
type NopLeaveChecker struct{}

// IsOnLeave always returns false.
func (NopLeaveChecker) IsOnLeave(context.Context, uuid.UUID, time.Time) (bool, error) {
	return false, nil
}

// NopDelegateResolver never resolves a delegate.
type NopDelegateResolver struct{}

// ResolveDelegate always returns (nil, nil).
func (NopDelegateResolver) ResolveDelegate(context.Context, uuid.UUID, *uuid.UUID, time.Time) (*Delegate, error) {
	return nil, nil
}

// NopNotifier discards every notification.
type NopNotifier struct{}

// NotifyApprovalTaskCreated is a no-op.
func (NopNotifier) NotifyApprovalTaskCreated(context.Context, *entity.ApprovalRequest, *entity.ApprovalTask) {
}

// NotifyApprovalCompleted is a no-op.
func (NopNotifier) NotifyApprovalCompleted(context.Context, *entity.ApprovalRequest) {}

// NotifyApprovalRejected is a no-op.
func (NopNotifier) NotifyApprovalRejected(context.Context, *entity.ApprovalRequest) {}

// NopAudit discards every audit record.
type NopAudit struct{}

// Record is a no-op.
func (NopAudit) Record(context.Context, *uuid.UUID, string, string, string, map[string]any) {}

// ─────────────────────────────────────────────────────────────────────────────
// Subject access port
// ─────────────────────────────────────────────────────────────────────────────

// ApprovalAction identifies the write operation an actor intends to perform on
// an approval request. Passed to CanActOnApprovalSubject so the business module
// can apply role- or action-specific rules (e.g. only a fund manager may revoke).
type ApprovalAction string

const (
	ApprovalActionApprove  ApprovalAction = "APPROVE"
	ApprovalActionReject   ApprovalAction = "REJECT"
	ApprovalActionRevoke   ApprovalAction = "REVOKE"
	ApprovalActionWithdraw ApprovalAction = "WITHDRAW"
	ApprovalActionCancel   ApprovalAction = "CANCEL"
)

// ApprovalSubjectAccessPort is implemented per subject type by the owning
// business module. The approval engine calls it to authorise every read and
// write without inspecting business-specific keys (contract_id, fund_id,
// portfolio_id). Register one per subject type via
// RuntimeService.RegisterSubjectAccessPort.
//
// Return nil to allow. To deny, return an error that wraps
// contract.ErrSubjectAccessDenied — the engine maps it to ErrForbidden (403).
// Return a RAW (non-wrapping) error for infrastructure failures (nil
// dependency, repository/IAM error, context timeout) so the engine surfaces a
// 5xx instead of a spurious 403. When no port is registered for a subject type
// the engine fails closed (ErrForbidden).
type ApprovalSubjectAccessPort interface {
	CanViewApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType vo.SubjectType, subjectID uuid.UUID) error
	CanSubmitApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType vo.SubjectType, subjectID uuid.UUID) error
	CanActOnApprovalSubject(ctx context.Context, actorID uuid.UUID, subjectType vo.SubjectType, subjectID uuid.UUID, action ApprovalAction) error
}

// NopSubjectAccessPort permits all access. Use in tests where subject-level
// authorisation is not under test so the rest of the service behaves normally.
type NopSubjectAccessPort struct{}

func (NopSubjectAccessPort) CanViewApprovalSubject(context.Context, uuid.UUID, vo.SubjectType, uuid.UUID) error {
	return nil
}
func (NopSubjectAccessPort) CanSubmitApprovalSubject(context.Context, uuid.UUID, vo.SubjectType, uuid.UUID) error {
	return nil
}
func (NopSubjectAccessPort) CanActOnApprovalSubject(context.Context, uuid.UUID, vo.SubjectType, uuid.UUID, ApprovalAction) error {
	return nil
}
