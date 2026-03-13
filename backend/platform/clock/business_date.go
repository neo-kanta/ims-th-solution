package clock

import "time"

// BangkokLocation is the Thai timezone (UTC+7).
var BangkokLocation *time.Location

func init() {
	var err error
	BangkokLocation, err = time.LoadLocation("Asia/Bangkok")
	if err != nil {
		// Fallback: manually create UTC+7 fixed zone
		BangkokLocation = time.FixedZone("Asia/Bangkok", 7*60*60)
	}
}

// IsWeekend returns true if the given date falls on Saturday or Sunday.
func IsWeekend(date time.Time) bool {
	day := date.Weekday()
	return day == time.Saturday || day == time.Sunday
}

// IsHoliday checks if the given date is in the list of holiday dates.
// Holiday dates should be loaded from the reference_data module.
func IsHoliday(date time.Time, holidays []time.Time) bool {
	d := truncateToDate(date)
	for _, h := range holidays {
		if truncateToDate(h).Equal(d) {
			return true
		}
	}
	return false
}

// NextBusinessDay returns the next Thai business day after the given date.
// It skips weekends and holidays.
func NextBusinessDay(date time.Time, holidays []time.Time) time.Time {
	next := date.AddDate(0, 0, 1)
	for IsWeekend(next) || IsHoliday(next, holidays) {
		next = next.AddDate(0, 0, 1)
	}
	return truncateToDate(next)
}

// PreviousBusinessDay returns the previous Thai business day before the given date.
func PreviousBusinessDay(date time.Time, holidays []time.Time) time.Time {
	prev := date.AddDate(0, 0, -1)
	for IsWeekend(prev) || IsHoliday(prev, holidays) {
		prev = prev.AddDate(0, 0, -1)
	}
	return truncateToDate(prev)
}

// IsBusinessDay returns true if the given date is a Thai business day
// (not a weekend and not a holiday).
func IsBusinessDay(date time.Time, holidays []time.Time) bool {
	return !IsWeekend(date) && !IsHoliday(date, holidays)
}

// ToThaiTime converts a UTC time to Thailand local time (UTC+7).
func ToThaiTime(t time.Time) time.Time {
	return t.In(BangkokLocation)
}

// truncateToDate strips time components, keeping only the date.
func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
