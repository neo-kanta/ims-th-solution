package event

import (
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// IRGCheckCompleted is emitted after every pre-trade or post-trade check.
type IRGCheckCompleted struct {
	CheckGroupID uuid.UUID      `json:"check_group_id"`
	Timing       vo.CheckTiming `json:"timing"`
	PortfolioID  uuid.UUID      `json:"portfolio_id"`
	ContractID   uuid.UUID      `json:"contract_id"`
	OrderID      *uuid.UUID     `json:"order_id,omitempty"`
	FinalVerdict vo.Verdict     `json:"final_verdict"`
	RulesChecked int            `json:"rules_checked"`
	Breaches     int            `json:"breaches"`
	Warnings     int            `json:"warnings"`
	DurationMs   int64          `json:"duration_ms"`
	BusinessDate time.Time      `json:"business_date"`
	CheckedAt    time.Time      `json:"checked_at"`
}

// IRGBreachRaised is emitted when a new breach is created.
type IRGBreachRaised struct {
	BreachID       uuid.UUID   `json:"breach_id"`
	CheckGroupID   uuid.UUID   `json:"check_group_id"`
	PortfolioID    uuid.UUID   `json:"portfolio_id"`
	ContractID     uuid.UUID   `json:"contract_id"`
	RuleTypeID     string      `json:"rule_type_id"`
	RuleInstanceID uuid.UUID   `json:"rule_instance_id"`
	Severity       vo.Severity `json:"severity"`
	Verdict        vo.Verdict  `json:"verdict"`
	Message        string      `json:"message"`
	BusinessDate   time.Time   `json:"business_date"`
}

// IRGBreachOverridden is emitted when a breach is overridden by a compliance officer.
type IRGBreachOverridden struct {
	OverrideID   uuid.UUID `json:"override_id"`
	BreachID     uuid.UUID `json:"breach_id"`
	OverriddenBy uuid.UUID `json:"overridden_by"`
	Reason       string    `json:"reason"`
	OccurredAt   time.Time `json:"occurred_at"`
}
