package domain

import "fmt"

// ErrDecisionNotFound signals the requested decision does not exist.
type ErrDecisionNotFound struct {
	DecisionID string
}

func (e *ErrDecisionNotFound) Error() string {
	return fmt.Sprintf("investment decision not found: %s", e.DecisionID)
}

// ErrDecisionNotDraft signals a transition attempt from a non-DRAFT status.
type ErrDecisionNotDraft struct {
	DecisionID    string
	CurrentStatus string
}

func (e *ErrDecisionNotDraft) Error() string {
	return fmt.Sprintf(
		"investment decision %s cannot be submitted from status %q (expected DRAFT)",
		e.DecisionID, e.CurrentStatus,
	)
}

// ErrInvalidDecisionRequest signals a validation failure at the application boundary.
type ErrInvalidDecisionRequest struct {
	Field  string
	Detail string
}

func (e *ErrInvalidDecisionRequest) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("invalid decision request: %s", e.Detail)
	}
	return fmt.Sprintf("invalid decision request: %s: %s", e.Field, e.Detail)
}

// ErrInvalidProcessGuardRequest signals a validation failure for the investment
// process execution guard.
type ErrInvalidProcessGuardRequest struct {
	Field  string
	Detail string
}

func (e *ErrInvalidProcessGuardRequest) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("invalid process guard request: %s", e.Detail)
	}
	return fmt.Sprintf("invalid process guard request: %s: %s", e.Field, e.Detail)
}

// ErrComplianceRejected signals the IRG pre-trade pipeline returned BLOCK.
// The CheckGroupID field lets callers deep-link into the breach / override UI.
type ErrComplianceRejected struct {
	DecisionID   string
	CheckGroupID string
	Message      string
}

func (e *ErrComplianceRejected) Error() string {
	return fmt.Sprintf(
		"investment decision %s rejected by IRG pre-trade check (group %s): %s",
		e.DecisionID, e.CheckGroupID, e.Message,
	)
}

// ErrPostTradeBlock signals the IRG post-trade verifier returned a BLOCKING
// breach and therefore the workflow must refuse the TRANSACTION_CLOSED
// transition. The CheckGroupID lets operators pivot into the breach UI.
//
// Placed in the investment-neutral domain errors file for reuse, but the
// workflow module also defines its own typed error for this condition —
// callers should not import this type from another module.
type ErrPostTradeBlock struct {
	ContractID   string
	BusinessDate string
	CheckGroupID string
	BreachCount  int
}

func (e *ErrPostTradeBlock) Error() string {
	return fmt.Sprintf(
		"contract %s business date %s cannot close transactions: %d BLOCK breach(es) (group %s)",
		e.ContractID, e.BusinessDate, e.BreachCount, e.CheckGroupID,
	)
}
