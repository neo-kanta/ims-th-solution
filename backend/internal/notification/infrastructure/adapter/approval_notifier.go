package adapter

import (
	"context"
	"fmt"

	approvaldomain "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	approvalentity "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/enum"
)

// ApprovalNotifier bridges the approval module's domain.Notifier interface
// onto the notification module's persistence. Failures are swallowed — the
// approval engine treats notification delivery as best-effort.
type ApprovalNotifier struct {
	svc *service.Service
}

func NewApprovalNotifier(svc *service.Service) *ApprovalNotifier {
	return &ApprovalNotifier{svc: svc}
}

func (n *ApprovalNotifier) NotifyApprovalTaskCreated(ctx context.Context, req *approvalentity.ApprovalRequest, task *approvalentity.ApprovalTask) {
	if n == nil || n.svc == nil || req == nil || task == nil {
		return
	}
	if task.AssignedUserID == nil {
		return
	}
	id := req.ID
	idempotencyKey := fmt.Sprintf("%s:%s:task:%s",
		string(enum.NotificationEventApprovalTaskAssigned),
		req.ID.String(),
		task.AssignedUserID.String(),
	)
	_ = n.svc.Create(ctx, service.CreateInput{
		RecipientUserID:   *task.AssignedUserID,
		Category:          string(enum.NotificationEventApprovalTaskAssigned),
		EventType:         string(enum.NotificationEventApprovalTaskAssigned),
		IdempotencyKey:    idempotencyKey,
		Title:             fmt.Sprintf("Approval request %s needs your action", req.RequestNumber),
		Body:              fmt.Sprintf("Subject: %s (%s)", req.SubjectTitle, req.SubjectType),
		Link:              fmt.Sprintf("/approval/requests/%s", req.ID),
		SourceModule:      "approval",
		SourceType:        "APPROVAL_REQUEST",
		SourceID:          &id,
		BusinessType:      "APPROVAL_REQUEST",
		BusinessLabel:     "Approval request",
		BusinessID:        &id,
		BusinessReference: req.RequestNumber,
		BusinessTitle:     req.SubjectTitle,
		ActionLabel:       "Open approval request",
		ActionURL:         fmt.Sprintf("/approval/requests/%s", req.ID),
	})
}

func (n *ApprovalNotifier) NotifyApprovalCompleted(ctx context.Context, req *approvalentity.ApprovalRequest) {
	if n == nil || n.svc == nil || req == nil {
		return
	}
	id := req.ID
	idempotencyKey := fmt.Sprintf("%s:%s:%s",
		string(enum.NotificationEventApprovalCompleted),
		req.ID.String(),
		req.SubmitterID.String(),
	)
	_ = n.svc.Create(ctx, service.CreateInput{
		RecipientUserID:   req.SubmitterID,
		Category:          string(enum.NotificationEventApprovalCompleted),
		EventType:         string(enum.NotificationEventApprovalCompleted),
		IdempotencyKey:    idempotencyKey,
		Title:             fmt.Sprintf("Approval request %s was approved", req.RequestNumber),
		Body:              fmt.Sprintf("Subject: %s", req.SubjectTitle),
		Link:              fmt.Sprintf("/approval/requests/%s", req.ID),
		SourceModule:      "approval",
		SourceType:        "APPROVAL_REQUEST",
		SourceID:          &id,
		BusinessType:      "APPROVAL_REQUEST",
		BusinessLabel:     "Approval request",
		BusinessID:        &id,
		BusinessReference: req.RequestNumber,
		BusinessTitle:     req.SubjectTitle,
		ActionLabel:       "View approval request",
		ActionURL:         fmt.Sprintf("/approval/requests/%s", req.ID),
	})
}

func (n *ApprovalNotifier) NotifyApprovalRejected(ctx context.Context, req *approvalentity.ApprovalRequest) {
	if n == nil || n.svc == nil || req == nil {
		return
	}
	id := req.ID
	idempotencyKey := fmt.Sprintf("%s:%s:%s",
		string(enum.NotificationEventApprovalRejected),
		req.ID.String(),
		req.SubmitterID.String(),
	)
	_ = n.svc.Create(ctx, service.CreateInput{
		RecipientUserID:   req.SubmitterID,
		Category:          string(enum.NotificationEventApprovalRejected),
		EventType:         string(enum.NotificationEventApprovalRejected),
		IdempotencyKey:    idempotencyKey,
		Title:             fmt.Sprintf("Approval request %s was rejected", req.RequestNumber),
		Body:              fmt.Sprintf("Subject: %s. Reason: %s", req.SubjectTitle, req.RejectionReason),
		Link:              fmt.Sprintf("/approval/requests/%s", req.ID),
		SourceModule:      "approval",
		SourceType:        "APPROVAL_REQUEST",
		SourceID:          &id,
		BusinessType:      "APPROVAL_REQUEST",
		BusinessLabel:     "Approval request",
		BusinessID:        &id,
		BusinessReference: req.RequestNumber,
		BusinessTitle:     req.SubjectTitle,
		ActionLabel:       "View approval request",
		ActionURL:         fmt.Sprintf("/approval/requests/%s", req.ID),
	})
}

var _ approvaldomain.Notifier = (*ApprovalNotifier)(nil)
