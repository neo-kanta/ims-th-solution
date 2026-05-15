package entity

import (
	"time"

	"github.com/google/uuid"
)

// AssetClass is the top-level investment-domain taxonomy row.
type AssetClass struct {
	ID           uuid.UUID
	Code         string
	Name         string
	Description  string
	DisplayOrder int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AssetSubtype is a child classification under an asset class.
type AssetSubtype struct {
	ID           uuid.UUID
	AssetClassID uuid.UUID
	Code         string
	Name         string
	Description  string
	DisplayOrder int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Region groups countries for portfolio reporting.
type Region struct {
	ID        uuid.UUID
	Code      string
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Country is a sovereign issuance jurisdiction.
type Country struct {
	ID        uuid.UUID
	ISOCode   string
	Name      string
	RegionID  *uuid.UUID
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Sector is a GICS-style industry classification node. Self-referencing via
// ParentID forms the (Sector → Industry Group → Industry → Sub-Industry)
// hierarchy.
type Sector struct {
	ID        uuid.UUID
	Code      string
	Name      string
	ParentID  *uuid.UUID
	Level     int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// FundCategory describes the investment style of a Fund or fund-like portfolio.
type FundCategory struct {
	ID           uuid.UUID
	Code         string
	Name         string
	AssetClassID *uuid.UUID
	Description  string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// InvestmentStyle is an optional categorical investment style.
type InvestmentStyle struct {
	ID        uuid.UUID
	Code      string
	Name      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
