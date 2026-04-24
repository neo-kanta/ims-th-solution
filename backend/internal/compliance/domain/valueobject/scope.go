package valueobject

import "github.com/google/uuid"

// ScopeType defines the level at which a rule binding applies.
type ScopeType string

const (
	ScopeGlobal         ScopeType = "GLOBAL"
	ScopeJurisdiction   ScopeType = "JURISDICTION"
	ScopeFundCategory   ScopeType = "FUND_CATEGORY"
	ScopeContract       ScopeType = "CONTRACT"
	ScopePortfolio      ScopeType = "PORTFOLIO"
	ScopeAssetClass     ScopeType = "ASSET_CLASS"
	ScopeInstrumentType ScopeType = "INSTRUMENT_TYPE"
)

// Specificity returns the conflict-resolution priority.
// Higher = more specific = wins when two bindings conflict.
func (s ScopeType) Specificity() int {
	return scopeSpecificity[s]
}

var scopeSpecificity = map[ScopeType]int{
	ScopeGlobal:         0,
	ScopeJurisdiction:   10,
	ScopeFundCategory:   20,
	ScopeAssetClass:     30,
	ScopeInstrumentType: 30,
	ScopeContract:       40,
	ScopePortfolio:      50,
}

// Scope identifies where a rule binding applies.
type Scope struct {
	Type ScopeType  `json:"type"`
	ID   *uuid.UUID `json:"id,omitempty"` // nil for GLOBAL
}
