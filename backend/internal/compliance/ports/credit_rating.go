package ports

import (
	"context"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// CreditRatingPort provides credit rating data per issuer.
type CreditRatingPort interface {
	GetRatings(ctx context.Context, issuers []string) (*spi.CreditRatingSnapshot, error)
}
