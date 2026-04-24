package valueobject

import "time"

// EffectiveWindow defines when a rule instance or binding is active.
// ValidTo == nil means indefinite.
type EffectiveWindow struct {
	ValidFrom time.Time  `json:"valid_from"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
}

// Contains returns true if the given date falls within the window (date-only comparison, UTC).
func (w EffectiveWindow) Contains(date time.Time) bool {
	d := truncateToDate(date)
	if d.Before(truncateToDate(w.ValidFrom)) {
		return false
	}
	if w.ValidTo != nil && d.After(truncateToDate(*w.ValidTo)) {
		return false
	}
	return true
}

// IsIndefinite returns true if the window has no end date.
func (w EffectiveWindow) IsIndefinite() bool {
	return w.ValidTo == nil
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
