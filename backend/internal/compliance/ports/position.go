package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// PositionSnapshotPort provides portfolio position data.
type PositionSnapshotPort interface {
	GetSnapshot(ctx context.Context, portfolioID uuid.UUID, asOf time.Time) (*spi.PositionSnapshot, error)
}
