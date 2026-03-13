package clock

import "time"

// Clock abstracts time operations for testability.
// Production code uses RealClock; tests use a mock implementation.
type Clock interface {
	// Now returns the current time in UTC.
	Now() time.Time
	// Today returns today's date (UTC, truncated to midnight).
	Today() time.Time
}

// RealClock implements Clock using the system clock.
type RealClock struct{}

// Now returns the current UTC time.
func (RealClock) Now() time.Time {
	return time.Now().UTC()
}

// Today returns today's date at midnight UTC.
func (RealClock) Today() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// FixedClock implements Clock with a fixed time, useful for testing.
type FixedClock struct {
	FixedTime time.Time
}

// Now returns the fixed time.
func (c FixedClock) Now() time.Time {
	return c.FixedTime
}

// Today returns the fixed date at midnight.
func (c FixedClock) Today() time.Time {
	return time.Date(c.FixedTime.Year(), c.FixedTime.Month(), c.FixedTime.Day(), 0, 0, 0, 0, time.UTC)
}
