package adapter

import (
	"context"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

// NopHolidayCalendarAdapter satisfies ports.HolidayCalendarPort using only
// weekend detection (no public holiday data). Used in Batch 1 until the
// reference_data module provides a real calendar.
//
// Production replacement: wrap reference_data.HolidayService in an adapter
// that implements this interface and inject it from main.go.
type NopHolidayCalendarAdapter struct{}

// IsBusinessDay returns false on Saturday and Sunday; true otherwise.
// holidayName is "WEEKEND" when false; "" when true.
func (a *NopHolidayCalendarAdapter) IsBusinessDay(
	_ context.Context,
	date time.Time,
) (bool, string, error) {
	if clock.IsWeekend(date) {
		return false, "WEEKEND", nil
	}
	return true, "", nil
}

// PreviousBusinessDay returns the most recent weekday before date.
func (a *NopHolidayCalendarAdapter) PreviousBusinessDay(
	_ context.Context,
	date time.Time,
) (time.Time, error) {
	// Use platform/clock with no holiday list (weekend-only awareness).
	return clock.PreviousBusinessDay(date, nil), nil
}

// compile-time interface check
var _ ports.HolidayCalendarPort = (*NopHolidayCalendarAdapter)(nil)
