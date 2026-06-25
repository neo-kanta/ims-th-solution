package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// NotificationStatus tracks alert delivery outcome.
type NotificationStatus string

const (
	NotificationStatusPending    NotificationStatus = "PENDING"
	NotificationStatusCreated    NotificationStatus = "CREATED"
	NotificationStatusSuppressed NotificationStatus = "SUPPRESSED"
	NotificationStatusFailed     NotificationStatus = "FAILED"
	NotificationStatusSkipped    NotificationStatus = "SKIPPED"
)

// AckState is derived from acknowledgement fields.
type AckState string

const (
	AckStateUnacknowledged AckState = "UNACKNOWLEDGED"
	AckStateAcknowledged   AckState = "ACKNOWLEDGED"
)

// AlertEvent is an immutable record of a NON_BREACHED → BREACHED crossing.
type AlertEvent struct {
	ID                uuid.UUID
	WatchlistItemID   uuid.UUID
	ThresholdRuleID   uuid.UUID
	ScopeType         ScopeType
	OwnerUserID       *uuid.UUID
	PortfolioID       *uuid.UUID
	SecurityID        uuid.UUID
	CreatedByUserID   uuid.UUID
	Direction         Direction
	PreviousState     RuleState
	CurrentState      RuleState
	ObservedPrice     decimal.Decimal
	ThresholdValue    decimal.Decimal
	Currency          *string
	QuoteProvider     *string
	ObservedAt        time.Time
	EvaluatedAt       time.Time
	Stale             bool
	StaleReason       *string
	IdempotencyKey    string
	NotificationStatus NotificationStatus
	NotificationID    *uuid.UUID
	NotificationError *string
	AcknowledgedBy    *uuid.UUID
	AcknowledgedAt    *time.Time
	AcknowledgementNote *string
	CreatedAt         time.Time
}

func (a *AlertEvent) AckState() AckState {
	if a.AcknowledgedAt != nil {
		return AckStateAcknowledged
	}
	return AckStateUnacknowledged
}

func (a *AlertEvent) IsAcknowledged() bool {
	return a.AcknowledgedAt != nil
}
