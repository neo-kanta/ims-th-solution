package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

// StuckDayCutoffDays is the number of business days an open workflow day
// may sit before the watcher considers it "stuck". Tunable; kept short for
// the demo so a single overnight pause produces a visible warning.
const StuckDayCutoffDays = 1

// StuckDayWatcher scans for workflow days that have been in DAY_OPEN or
// MANAGER_APPROVED past the configured cutoff and emits one notification
// per (operator, day, state) tuple via the injected OperatorNotifier.
//
// The watcher NEVER performs a state transition — operator action is the
// only legitimate way to advance a stuck day; the watcher only raises the
// alarm.
type StuckDayWatcher struct {
	dayRepo  StuckDayLister
	notifier ports.OperatorNotifier
	clock    clock.Clock
}

// StuckDayLister is the narrow read port the watcher needs. Satisfied by
// the workflow day repository's ListByState method, which already exists.
type StuckDayLister interface {
	ListByState(ctx context.Context, state vo.WorkflowState, businessDate time.Time, offset, limit int) ([]*entity.WorkflowDay, int64, error)
}

func NewStuckDayWatcher(dayRepo StuckDayLister, notifier ports.OperatorNotifier, clk clock.Clock) *StuckDayWatcher {
	if notifier == nil {
		notifier = ports.NopOperatorNotifier{}
	}
	if clk == nil {
		clk = clock.RealClock{}
	}
	return &StuckDayWatcher{dayRepo: dayRepo, notifier: notifier, clock: clk}
}

// RunOnce scans every business date from cutoff back through a reasonable
// window (here 14 days) so the watcher catches days that were missed by an
// earlier interval. The notifier's CreateIfAbsent dedupe keeps the inbox
// quiet across repeat ticks.
func (w *StuckDayWatcher) RunOnce(ctx context.Context) error {
	if w == nil || w.dayRepo == nil {
		return nil
	}
	now := w.clock.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Walk back from the cutoff day. A 14-day window catches multi-day stalls
	// without unbounded growth; pages of 100 keep memory bounded.
	for offsetDays := StuckDayCutoffDays; offsetDays <= StuckDayCutoffDays+14; offsetDays++ {
		businessDate := today.AddDate(0, 0, -offsetDays)
		if err := w.scanState(ctx, vo.StateDayOpen, businessDate, offsetDays); err != nil {
			return err
		}
		if err := w.scanState(ctx, vo.StateManagerApproved, businessDate, offsetDays); err != nil {
			return err
		}
	}
	return nil
}

func (w *StuckDayWatcher) scanState(ctx context.Context, state vo.WorkflowState, businessDate time.Time, ageDays int) error {
	const pageSize = 100
	offset := 0
	for {
		days, _, err := w.dayRepo.ListByState(ctx, state, businessDate, offset, pageSize)
		if err != nil {
			return fmt.Errorf("listing stuck days in %s for %s: %w",
				state, businessDate.Format("2006-01-02"), err)
		}
		if len(days) == 0 {
			return nil
		}
		for _, d := range days {
			warn := ports.StuckDayWarning{
				WorkflowDayID: d.ID,
				ContractID:    d.ContractID,
				BusinessDate:  d.BusinessDate,
				CurrentState:  string(d.CurrentState),
				OpenedBy:      d.OpenedBy,
				ApprovedBy:    d.ManagerApprovedBy,
				AgeDays:       ageDays,
			}
			if err := w.notifier.NotifyStuckDay(ctx, warn); err != nil {
				slog.Warn("stuck day notifier failed",
					"contract_id", d.ContractID,
					"business_date", d.BusinessDate.Format("2006-01-02"),
					"state", d.CurrentState,
					"error", err,
				)
				// Continue — one failure should not abort the scan.
			}
		}
		if len(days) < pageSize {
			return nil
		}
		offset += pageSize
	}
}

// Static guard: domain.WorkflowDayRepository must remain compatible with the
// narrow StuckDayLister so production wiring can pass it directly.
var _ StuckDayLister = (domain.WorkflowDayRepository)(nil)
