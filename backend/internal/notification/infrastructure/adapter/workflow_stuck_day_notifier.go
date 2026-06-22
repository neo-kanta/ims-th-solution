package adapter

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/application/service"
	workflowports "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/enum"
)

// WorkflowStuckDayNotifier implements workflow ports.OperatorNotifier by
// writing one in-app notification per (operator, stuck day, state) tuple via
// the notification service's CreateIfAbsent dedupe path.
type WorkflowStuckDayNotifier struct {
	svc *service.Service
}

func NewWorkflowStuckDayNotifier(svc *service.Service) *WorkflowStuckDayNotifier {
	return &WorkflowStuckDayNotifier{svc: svc}
}

// dedupeWindowHours suppresses repeat notifications within 23 hours so the
// hourly scheduler doesn't flood the inbox when a workflow day stays stuck.
const dedupeWindowHours = 23

// NotifyStuckDay emits an in-app notification (and, if email is configured,
// an email outbox row) to each accountable user for the stuck workflow day.
func (n *WorkflowStuckDayNotifier) NotifyStuckDay(ctx context.Context, w workflowports.StuckDayWarning) error {
	if n == nil || n.svc == nil {
		return nil
	}

	businessRef := w.BusinessDate.Format("2006-01-02")
	title := fmt.Sprintf(
		"Workflow day stuck in %s for %d day(s)",
		w.CurrentState, w.AgeDays,
	)
	body := fmt.Sprintf(
		"Contract %s • business date %s is still in %s. Please complete or cancel the day.",
		w.ContractID, businessRef, w.CurrentState,
	)
	link := fmt.Sprintf("/workflow/%s?businessDate=%s", w.ContractID, businessRef)
	dayID := w.WorkflowDayID

	recipients := map[uuid.UUID]struct{}{}
	if w.OpenedBy != nil && *w.OpenedBy != uuid.Nil {
		recipients[*w.OpenedBy] = struct{}{}
	}
	if w.ApprovedBy != nil && *w.ApprovedBy != uuid.Nil {
		recipients[*w.ApprovedBy] = struct{}{}
	}

	for rec := range recipients {
		idempotencyKey := fmt.Sprintf("%s:%s:%s:%s",
			string(enum.NotificationEventWorkflowStuckDay),
			w.WorkflowDayID.String(),
			w.CurrentState,
			rec.String(),
		)
		if err := n.svc.CreateIfAbsent(ctx, service.CreateInput{
			RecipientUserID:   rec,
			Category:          string(enum.NotificationEventWorkflowStuckDay),
			EventType:         string(enum.NotificationEventWorkflowStuckDay),
			IdempotencyKey:    idempotencyKey,
			Title:             title,
			Body:              body,
			Link:              link,
			SourceModule:      "workflow",
			SourceType:        "WORKFLOW_DAY",
			SourceID:          &dayID,
			BusinessType:      "WORKFLOW_DAY",
			BusinessLabel:     "Workflow day",
			BusinessID:        &dayID,
			BusinessReference: businessRef,
			BusinessTitle:     body,
			ActionLabel:       "View workflow",
			ActionURL:         link,
		}, dedupeWindowHours); err != nil {
			return fmt.Errorf("emit stuck-day notification: %w", err)
		}
	}
	return nil
}

var _ workflowports.OperatorNotifier = (*WorkflowStuckDayNotifier)(nil)
