package domain

import (
	"fmt"
	"net/http"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
)

// ErrDecisionLifecycle signals an illegal state transition on an investment
// decision aggregate (e.g. trying to submit an already-cancelled row).
type ErrDecisionLifecycle struct {
	DecisionID    string
	CurrentStatus string
	Detail        string
}

func (e *ErrDecisionLifecycle) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("investment decision %s in status %q: %s", e.DecisionID, e.CurrentStatus, e.Detail)
	}
	return fmt.Sprintf("investment decision %s in status %q cannot perform requested action", e.DecisionID, e.CurrentStatus)
}

func (*ErrDecisionLifecycle) ErrorCode() string { return errcode.CodeDecisionLifecycle }
func (*ErrDecisionLifecycle) HTTPStatus() int   { return http.StatusConflict }
func (e *ErrDecisionLifecycle) ErrorDetails() map[string]any {
	return map[string]any{"decision_id": e.DecisionID, "current_status": e.CurrentStatus}
}

// ErrDecisionReferenceInvalid signals that a research report referenced by a
// decision is not acceptable (missing, not approved, wrong contract, etc.).
type ErrDecisionReferenceInvalid struct {
	DecisionID string
	ReportID   string
	Reason     string
}

func (e *ErrDecisionReferenceInvalid) Error() string {
	return fmt.Sprintf("research report reference invalid: %s", e.Reason)
}

func (*ErrDecisionReferenceInvalid) ErrorCode() string { return errcode.CodeDecisionReferenceBad }
func (*ErrDecisionReferenceInvalid) HTTPStatus() int   { return http.StatusUnprocessableEntity }
func (e *ErrDecisionReferenceInvalid) ErrorDetails() map[string]any {
	d := map[string]any{"reason": e.Reason}
	if e.ReportID != "" {
		d["research_report_id"] = e.ReportID
	}
	if e.DecisionID != "" {
		d["decision_id"] = e.DecisionID
	}
	return d
}

// ErrDecisionNumberConflict signals a duplicate decision_number on create.
type ErrDecisionNumberConflict struct {
	DecisionNumber string
}

func (e *ErrDecisionNumberConflict) Error() string {
	return fmt.Sprintf("decision number already exists: %s", e.DecisionNumber)
}

func (*ErrDecisionNumberConflict) ErrorCode() string { return errcode.CodeCodeAlreadyTaken }
func (*ErrDecisionNumberConflict) HTTPStatus() int   { return http.StatusConflict }
func (e *ErrDecisionNumberConflict) ErrorDetails() map[string]any {
	return map[string]any{"decision_number": e.DecisionNumber}
}

// ─── Execution ──────────────────────────────────────────────────────────

type ErrExecutionNotFound struct {
	ExecutionID string
}

func (e *ErrExecutionNotFound) Error() string {
	return fmt.Sprintf("execution not found: %s", e.ExecutionID)
}
func (*ErrExecutionNotFound) ErrorCode() string { return errcode.CodeExecutionNotFound }
func (*ErrExecutionNotFound) HTTPStatus() int   { return http.StatusNotFound }
func (e *ErrExecutionNotFound) ErrorDetails() map[string]any {
	return map[string]any{"execution_id": e.ExecutionID}
}

type ErrExecutionLifecycle struct {
	ExecutionID   string
	CurrentStatus string
	Detail        string
}

func (e *ErrExecutionLifecycle) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("execution %s in status %q: %s", e.ExecutionID, e.CurrentStatus, e.Detail)
	}
	return fmt.Sprintf("execution %s in status %q cannot perform requested action", e.ExecutionID, e.CurrentStatus)
}
func (*ErrExecutionLifecycle) ErrorCode() string { return errcode.CodeExecutionLifecycle }
func (*ErrExecutionLifecycle) HTTPStatus() int   { return http.StatusConflict }
func (e *ErrExecutionLifecycle) ErrorDetails() map[string]any {
	return map[string]any{"execution_id": e.ExecutionID, "current_status": e.CurrentStatus}
}

// ─── Trade confirmation ─────────────────────────────────────────────────

type ErrConfirmationNotFound struct {
	ConfirmationID string
}

func (e *ErrConfirmationNotFound) Error() string {
	return fmt.Sprintf("trade confirmation not found: %s", e.ConfirmationID)
}
func (*ErrConfirmationNotFound) ErrorCode() string { return errcode.CodeConfirmationNotFound }
func (*ErrConfirmationNotFound) HTTPStatus() int   { return http.StatusNotFound }
func (e *ErrConfirmationNotFound) ErrorDetails() map[string]any {
	return map[string]any{"confirmation_id": e.ConfirmationID}
}

type ErrConfirmationLifecycle struct {
	ConfirmationID string
	CurrentStatus  string
	Detail         string
}

func (e *ErrConfirmationLifecycle) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("trade confirmation %s in status %q: %s", e.ConfirmationID, e.CurrentStatus, e.Detail)
	}
	return fmt.Sprintf("trade confirmation %s in status %q cannot perform requested action", e.ConfirmationID, e.CurrentStatus)
}
func (*ErrConfirmationLifecycle) ErrorCode() string { return errcode.CodeConfirmationLifecycle }
func (*ErrConfirmationLifecycle) HTTPStatus() int   { return http.StatusConflict }
func (e *ErrConfirmationLifecycle) ErrorDetails() map[string]any {
	return map[string]any{"confirmation_id": e.ConfirmationID, "current_status": e.CurrentStatus}
}

type ErrConfirmationMismatchReasonRequired struct {
	ConfirmationID string
}

func (e *ErrConfirmationMismatchReasonRequired) Error() string {
	return "discrepancy_reason is required when marking a confirmation MISMATCHED or REVIEWED"
}
func (*ErrConfirmationMismatchReasonRequired) ErrorCode() string {
	return errcode.CodeConfirmationMismatch
}
func (*ErrConfirmationMismatchReasonRequired) HTTPStatus() int { return http.StatusUnprocessableEntity }
func (e *ErrConfirmationMismatchReasonRequired) ErrorDetails() map[string]any {
	return map[string]any{"confirmation_id": e.ConfirmationID}
}

// ErrClosePendingConfirmations signals that the workflow closing gate found
// executions that lack a resolved confirmation (MATCHED or REVIEWED).
type ErrClosePendingConfirmations struct {
	ContractID      string
	BusinessDate    string
	PendingCount    int
	UnresolvedCount int
}

func (e *ErrClosePendingConfirmations) Error() string {
	return fmt.Sprintf(
		"contract %s business date %s cannot close transactions: %d pending and %d unresolved trade confirmations",
		e.ContractID, e.BusinessDate, e.PendingCount, e.UnresolvedCount,
	)
}
func (*ErrClosePendingConfirmations) ErrorCode() string { return errcode.CodeClosePendingConfirm }
func (*ErrClosePendingConfirmations) HTTPStatus() int   { return http.StatusUnprocessableEntity }
func (e *ErrClosePendingConfirmations) ErrorDetails() map[string]any {
	return map[string]any{
		"contract_id":      e.ContractID,
		"business_date":    e.BusinessDate,
		"pending_count":    e.PendingCount,
		"unresolved_count": e.UnresolvedCount,
	}
}
