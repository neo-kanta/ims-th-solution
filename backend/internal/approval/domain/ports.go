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
