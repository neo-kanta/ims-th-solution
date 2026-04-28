package spi

import vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"

// RuleCategory classifies rule types for organizational and reporting purposes.
type RuleCategory string

const (
	CategoryRegulatory  RuleCategory = "REGULATORY"
	CategoryMandate     RuleCategory = "MANDATE"
	CategoryHouse       RuleCategory = "HOUSE"
	CategoryClient      RuleCategory = "CLIENT"
	CategoryRestriction RuleCategory = "RESTRICTION"
	CategoryRatio       RuleCategory = "RATIO"
	CategoryTemporal    RuleCategory = "TEMPORAL"
	CategoryBehavioral  RuleCategory = "BEHAVIORAL"
)

// RuleMetadata is the stable identity and capabilities of a rule type.
type RuleMetadata struct {
	TypeID           string           `json:"type_id"`
	Version          string           `json:"version"`
	Category         RuleCategory     `json:"category"`
	DefaultSeverity  vo.Severity      `json:"default_severity"`
	SupportedTimings []vo.CheckTiming `json:"supported_timings"`
	SupportedScopes  []vo.ScopeType   `json:"supported_scopes"`
	Overridable      bool             `json:"overridable"`
	Description      string           `json:"description"`
}
