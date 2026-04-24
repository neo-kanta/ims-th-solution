package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// PortfolioMetadataPort provides portfolio/mandate metadata.
type PortfolioMetadataPort interface {
	GetMetadata(ctx context.Context, portfolioID uuid.UUID) (*spi.PortfolioMetadata, error)
}
