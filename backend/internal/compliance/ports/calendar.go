package ports

import (
	"context"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// CalendarPort provides business day and holiday calendar data.
type CalendarPort interface {
	GetCalendar(ctx context.Context, jurisdiction string, from, to time.Time) (*spi.CalendarSnapshot, error)
}
