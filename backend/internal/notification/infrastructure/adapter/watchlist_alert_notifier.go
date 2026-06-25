package adapter

import (
	"context"
	"fmt"

	"github.com/neo-kanta/ims-th-solution/backend/internal/notification/application/service"
	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/enum"
)

// WatchlistAlertNotifier bridges the watchlist module's WatchlistAlertNotifier
// interface onto the notification module's persistence. Failures are returned so
// the evaluator can record notification_status = FAILED.
type WatchlistAlertNotifier struct {
	svc *service.Service
}

// NewWatchlistAlertNotifier wires the notifier.
func NewWatchlistAlertNotifier(svc *service.Service) *WatchlistAlertNotifier {
	return &WatchlistAlertNotifier{svc: svc}
}

func (n *WatchlistAlertNotifier) NotifyThresholdBreached(ctx context.Context, input watchlistdomain.WatchlistAlertNotificationInput) error {
	if n == nil || n.svc == nil {
		return nil
	}
	alertID := input.AlertEventID
	body := fmt.Sprintf(
		"Price %s threshold breached: %s %s (observed: %s, threshold: %s)",
		input.Direction,
		input.SecuritySymbol,
		input.SecurityName,
		input.ObservedPrice.StringFixed(2),
		input.ThresholdValue.StringFixed(2),
	)
	title := fmt.Sprintf("Watchlist alert: %s %s", input.SecuritySymbol, input.Direction)
	return n.svc.Create(ctx, service.CreateInput{
		RecipientUserID:   input.RecipientUserID,
		Category:          string(enum.NotificationEventAlertThresholdBreached),
		EventType:         string(enum.NotificationEventAlertThresholdBreached),
		IdempotencyKey:    input.IdempotencyKey,
		Title:             title,
		Body:              body,
		SourceModule:      "watchlist",
		SourceType:        "watchlist_alert_event",
		SourceID:          &alertID,
		BusinessType:      "WATCHLIST_ALERT",
		BusinessLabel:     "Watchlist alert",
		BusinessID:        &alertID,
		ActionLabel:       "View alert",
		ActionURL:         fmt.Sprintf("/watchlists/alerts/%s", input.AlertEventID.String()),
	})
}

var _ watchlistdomain.WatchlistAlertNotifier = (*WatchlistAlertNotifier)(nil)
