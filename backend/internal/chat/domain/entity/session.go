package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// Session is one conversation. The active provider and model are captured at
// creation so audit and history reflect what was actually used.
type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Provider  valueobject.ProviderID
	Model     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
