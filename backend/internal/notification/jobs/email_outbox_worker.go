// Package jobs holds background workers for the notification module.
package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/domain/entity"
)

// EmailOutboxWorker periodically claims due outbox rows and delivers them via
// the configured EmailSender. Multiple backend instances are safe because the
// claim query uses FOR UPDATE SKIP LOCKED.
type EmailOutboxWorker struct {
	repo         domain.EmailOutboxRepository
	sender       domain.EmailSender
	interval     time.Duration
	batchSize    int
	staleTimeout time.Duration
	retryPolicy  []time.Duration
	workerID     string
}

// NewEmailOutboxWorker creates a worker. interval, batchSize, staleTimeout,
// and retryPolicy come directly from AppConfig.
func NewEmailOutboxWorker(
	repo domain.EmailOutboxRepository,
	sender domain.EmailSender,
	interval time.Duration,
	batchSize int,
	staleTimeout time.Duration,
	retryPolicy []time.Duration,
) *EmailOutboxWorker {
	return &EmailOutboxWorker{
		repo:         repo,
		sender:       sender,
		interval:     interval,
		batchSize:    batchSize,
		staleTimeout: staleTimeout,
		retryPolicy:  retryPolicy,
		workerID:     fmt.Sprintf("worker-%s", uuid.New().String()[:8]),
	}
}

// Start runs the worker loop until ctx is cancelled. Call in a goroutine.
func (w *EmailOutboxWorker) Start(ctx context.Context) {
	slog.Info("email outbox worker started", "worker_id", w.workerID, "interval", w.interval)

	// Recover stale rows on startup before the first claim cycle.
	w.recoverStale(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("email outbox worker stopped", "worker_id", w.workerID)
			return
		case <-ticker.C:
			w.recoverStale(ctx)
			w.processBatch(ctx)
		}
	}
}

func (w *EmailOutboxWorker) recoverStale(ctx context.Context) {
	if w.staleTimeout <= 0 {
		return
	}
	nextDelay := w.nextRetryDelay(0)
	n, err := w.repo.RecoverStale(ctx, w.staleTimeout, nextDelay)
	if err != nil {
		slog.Warn("email outbox: stale recovery error", "worker_id", w.workerID, "err", err)
		return
	}
	if n > 0 {
		slog.Info("email outbox: recovered stale rows", "worker_id", w.workerID, "recovered", n)
	}
}

func (w *EmailOutboxWorker) processBatch(ctx context.Context) {
	batchSize := w.batchSize
	if batchSize <= 0 {
		batchSize = 25
	}
	rows, err := w.repo.ClaimBatch(ctx, batchSize, w.workerID)
	if err != nil {
		slog.Warn("email outbox: claim batch error", "worker_id", w.workerID, "err", err)
		return
	}
	for _, row := range rows {
		w.sendRow(ctx, row)
	}
}

func (w *EmailOutboxWorker) sendRow(ctx context.Context, row *entity.EmailOutbox) {
	msg := domain.EmailMessage{
		ToEmail:  row.ToEmail,
		ToName:   row.ToName,
		Subject:  row.Subject,
		BodyText: row.BodyText,
		BodyHTML: row.BodyHTML,
	}

	providerMsgID, err := w.sender.Send(ctx, msg)
	if err != nil {
		nextDelay := w.nextRetryDelay(row.Attempts + 1)
		slog.Warn("email outbox: send failed",
			"worker_id", w.workerID,
			"outbox_id", row.ID,
			"attempts", row.Attempts+1,
			"err", err,
		)
		if markErr := w.repo.MarkFailed(ctx, row.ID, err.Error(), nextDelay); markErr != nil {
			slog.Error("email outbox: mark failed error", "outbox_id", row.ID, "err", markErr)
		}
		return
	}

	slog.Info("email outbox: sent",
		"worker_id", w.workerID,
		"outbox_id", row.ID,
		"to", row.ToEmail,
		"event_type", row.EventType,
	)
	if markErr := w.repo.MarkSent(ctx, row.ID, providerMsgID); markErr != nil {
		slog.Error("email outbox: mark sent error", "outbox_id", row.ID, "err", markErr)
	}
}

// nextRetryDelay returns the delay before the next send attempt.
// attempt is the number of attempts already made (0-indexed for next delay).
func (w *EmailOutboxWorker) nextRetryDelay(attempt int) time.Duration {
	if len(w.retryPolicy) == 0 {
		return time.Minute // safe fallback
	}
	idx := attempt
	if idx < 0 {
		idx = 0
	}
	if idx >= len(w.retryPolicy) {
		idx = len(w.retryPolicy) - 1
	}
	return w.retryPolicy[idx]
}
