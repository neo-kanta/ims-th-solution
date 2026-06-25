package entity

import (
	"time"

	"github.com/google/uuid"
)

// ScopeType distinguishes personal and portfolio watchlists.
type ScopeType string

const (
	ScopePersonal  ScopeType = "PERSONAL"
	ScopePortfolio ScopeType = "PORTFOLIO"
)

// ItemStatus controls whether an item is evaluated.
type ItemStatus string

const (
	ItemStatusActive   ItemStatus = "ACTIVE"
	ItemStatusDisabled ItemStatus = "DISABLED"
)

// WatchlistItem is a single security tracked in a personal or portfolio watchlist.
type WatchlistItem struct {
	ID           uuid.UUID
	ScopeType    ScopeType
	OwnerUserID  *uuid.UUID
	PortfolioID  *uuid.UUID
	SecurityID   uuid.UUID
	DisplayOrder *int
	Pinned       bool
	Note         *string
	Status       ItemStatus
	CreatedBy    uuid.UUID
	UpdatedBy    *uuid.UUID
	DeletedBy    *uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func (i *WatchlistItem) IsActive() bool {
	return i != nil && i.DeletedAt == nil && i.Status == ItemStatusActive
}

func (i *WatchlistItem) IsPersonal() bool {
	return i != nil && i.ScopeType == ScopePersonal
}

func (i *WatchlistItem) IsPortfolio() bool {
	return i != nil && i.ScopeType == ScopePortfolio
}
