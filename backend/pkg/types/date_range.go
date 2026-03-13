package types

import "time"

// DateRange represents a range between two dates (inclusive).
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// Contains checks if a given date falls within the range (inclusive).
func (dr DateRange) Contains(date time.Time) bool {
	d := truncateToDate(date)
	from := truncateToDate(dr.From)
	to := truncateToDate(dr.To)
	return !d.Before(from) && !d.After(to)
}

// Overlaps checks if two date ranges overlap.
func (dr DateRange) Overlaps(other DateRange) bool {
	return !truncateToDate(dr.From).After(truncateToDate(other.To)) &&
		!truncateToDate(dr.To).Before(truncateToDate(other.From))
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
