package ports

import (
	"context"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// InstrumentClassificationPort provides instrument classification data.
type InstrumentClassificationPort interface {
	GetClassifications(ctx context.Context, tickers []string) (*spi.ClassificationSnapshot, error)
}
