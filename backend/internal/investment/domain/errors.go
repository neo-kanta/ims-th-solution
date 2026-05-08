package domain

import (
	"fmt"
	"net/http"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
)

// ErrDecisionNotFound signals the requested decision does not exist.
type ErrDecisionNotFound struct {
	DecisionID string
}

func (e *ErrDecisionNotFound) Error() string {
	return fmt.Sprintf("investment decision not found: %s", e.DecisionID)
}

// ErrorCode implements errcode.Coded.
func (*ErrDecisionNotFound) ErrorCode() string { return errcode.CodeDecisionNotFound }

// HTTPStatus implements errcode.HTTPStatus.
func (*ErrDecisionNotFound) HTTPStatus() int { return http.StatusNotFound }

// ErrorDetails implements errcode.Detailed.
func (e *ErrDecisionNotFound) ErrorDetails() map[string]any {
	return map[string]any{"decision_id": e.DecisionID}
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

// ErrorCode implements errcode.Coded.
func (*ErrDecisionNotDraft) ErrorCode() string { return errcode.CodeDecisionNotDraft }

// HTTPStatus implements errcode.HTTPStatus.
func (*ErrDecisionNotDraft) HTTPStatus() int { return http.StatusConflict }

// ErrorDetails implements errcode.Detailed.
func (e *ErrDecisionNotDraft) ErrorDetails() map[string]any {
	return map[string]any{
		"decision_id":    e.DecisionID,
		"current_status": e.CurrentStatus,
	}
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

// ErrorCode implements errcode.Coded.
func (*ErrInvalidDecisionRequest) ErrorCode() string { return errcode.CodeInvalidRequest }

// HTTPStatus implements errcode.HTTPStatus.
func (*ErrInvalidDecisionRequest) HTTPStatus() int { return http.StatusBadRequest }

// ErrorDetails implements errcode.Detailed.
func (e *ErrInvalidDecisionRequest) ErrorDetails() map[string]any {
	d := map[string]any{}
	if e.Field != "" {
		d["field"] = e.Field
	}
	if e.Detail != "" {
		d["detail"] = e.Detail
	}
	if len(d) == 0 {
		return nil
	}
	return d
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

// ErrorCode implements errcode.Coded.
func (*ErrInvalidProcessGuardRequest) ErrorCode() string { return errcode.CodeInvalidRequest }

// HTTPStatus implements errcode.HTTPStatus.
func (*ErrInvalidProcessGuardRequest) HTTPStatus() int { return http.StatusBadRequest }

// ErrorDetails implements errcode.Detailed.
func (e *ErrInvalidProcessGuardRequest) ErrorDetails() map[string]any {
	d := map[string]any{}
	if e.Field != "" {
		d["field"] = e.Field
	}
	if e.Detail != "" {
		d["detail"] = e.Detail
	}
	if len(d) == 0 {
		return nil
	}
	return d
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

// ErrorCode implements errcode.Coded.
func (*ErrComplianceRejected) ErrorCode() string { return errcode.CodeComplianceRejected }

// HTTPStatus implements errcode.HTTPStatus.
func (*ErrComplianceRejected) HTTPStatus() int { return http.StatusUnprocessableEntity }

// ErrorDetails implements errcode.Detailed.
func (e *ErrComplianceRejected) ErrorDetails() map[string]any {
	return map[string]any{
		"decision_id":    e.DecisionID,
		"check_group_id": e.CheckGroupID,
	}
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

// ErrorCode implements errcode.Coded.
func (*ErrPostTradeBlock) ErrorCode() string { return errcode.CodePostTradeBlocked }

// HTTPStatus implements errcode.HTTPStatus.
func (*ErrPostTradeBlock) HTTPStatus() int { return http.StatusUnprocessableEntity }

// ErrorDetails implements errcode.Detailed.
func (e *ErrPostTradeBlock) ErrorDetails() map[string]any {
	return map[string]any{
		"contract_id":    e.ContractID,
		"business_date":  e.BusinessDate,
		"check_group_id": e.CheckGroupID,
		"breach_count":   e.BreachCount,
	}
}
