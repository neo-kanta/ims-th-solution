// Package ports declares external-dependency interfaces consumed by the
// workflow module. Real implementations are injected from other modules
// (reference_data, investment) via main.go or integration adapters.
// NOP stubs in infrastructure/adapter/ are used for Batch 1.
package ports

import (
	"context"
	"time"
)

// HolidayCalendarPort provides business-day semantics for the Asia/Bangkok
// market calendar. Implemented by the reference_data module in production;
// backed by a NOP weekend-only adapter in Batch 1.
type HolidayCalendarPort interface {
	// IsBusinessDay returns whether date is a valid Thai trading day and,
	// when it is not, the name of the holiday or "WEEKEND".
	IsBusinessDay(ctx context.Context, date time.Time) (isBusinessDay bool, holidayName string, err error)

	// PreviousBusinessDay returns the nearest prior business day before date.
	PreviousBusinessDay(ctx context.Context, date time.Time) (time.Time, error)
}
