package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// RestrictionListPort provides blacklist/whitelist/graylist data.
type RestrictionListPort interface {
	GetRestrictions(ctx context.Context, portfolioID uuid.UUID, tickers []string) (*spi.RestrictionSnapshot, error)
}
