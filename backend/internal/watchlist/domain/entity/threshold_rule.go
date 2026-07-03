package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MetricType defines what value is compared against the threshold.
type MetricType string

const MetricTypeMarketPrice MetricType = "MARKET_PRICE"

// Direction controls which side triggers a breach.
type Direction string

const (
	DirectionAbove Direction = "ABOVE"
	DirectionBelow Direction = "BELOW"
)

// RuleStatus controls whether a rule participates in evaluation.
type RuleStatus string

const (
	RuleStatusEnabled  RuleStatus = "ENABLED"
	RuleStatusDisabled RuleStatus = "DISABLED"
)

// RuleState tracks the last known crossing state.
type RuleState string

const (
	RuleStateUnknown     RuleState = "UNKNOWN"
	RuleStateNonBreached RuleState = "NON_BREACHED"
	RuleStateBreached    RuleState = "BREACHED"
)

// ThresholdRule is a price alert rule attached to a watchlist item.
type ThresholdRule struct {
	ID                 uuid.UUID
	WatchlistItemID    uuid.UUID
	MetricType         MetricType
	Direction          Direction
	ThresholdValue     decimal.Decimal
	Currency           *string
	CooldownMinutes    int
	Status             RuleStatus
	LastState          RuleState
	LastObservedPrice  *decimal.Decimal
	LastObservedAt     *time.Time
	LastEvaluatedAt    *time.Time
	LastStateChangedAt *time.Time
	LastAlertedAt      *time.Time
	LastQuoteStale     bool
	LastStaleReason    *string
	CreatedBy          uuid.UUID
	UpdatedBy          *uuid.UUID
	DeletedBy          *uuid.UUID
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

func (r *ThresholdRule) IsEligible() bool {
	return r != nil && r.DeletedAt == nil && r.Status == RuleStatusEnabled
}
