package domain

import (
	"fmt"
	"net/http"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
)

// ErrInvalidResearchReportRequest signals a validation failure on a research
// report request at the application boundary. Distinct from
// ErrInvalidDecisionRequest so the wrong aggregate's error type does not
// surface in research handlers.
type ErrInvalidResearchReportRequest struct {
	Field  string
	Detail string
}

func (e *ErrInvalidResearchReportRequest) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("invalid research report request: %s", e.Detail)
	}
	return fmt.Sprintf("invalid research report request: %s %s", e.Field, e.Detail)
}

// ErrorCode implements errcode.Coded.
func (*ErrInvalidResearchReportRequest) ErrorCode() string { return errcode.CodeInvalidRequest }

// HTTPStatus implements errcode.HTTPStatus.
func (*ErrInvalidResearchReportRequest) HTTPStatus() int { return http.StatusBadRequest }

// ErrorDetails implements errcode.Detailed.
func (e *ErrInvalidResearchReportRequest) ErrorDetails() map[string]any {
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

// ErrResearchReportNotFound signals a research report lookup miss.
type ErrResearchReportNotFound struct {
	ReportID string
}

func (e *ErrResearchReportNotFound) Error() string {
	return fmt.Sprintf("research report not found: %s", e.ReportID)
}

// ErrResearchReportNoAlreadyExists is raised when creating a report with a
// report_no that is already in use among non-deleted rows.
type ErrResearchReportNoAlreadyExists struct {
	ReportNo string
}

func (e *ErrResearchReportNoAlreadyExists) Error() string {
	return fmt.Sprintf("research report number %q already exists", e.ReportNo)
}

// ErrResearchReportCannotUpdate is raised when the report cannot be updated
// in its current lifecycle state — typically because the review has been
// completed or the row is soft-deleted.
type ErrResearchReportCannotUpdate struct {
	ReportID     string
	ReviewStatus string
	Reason       string
}

func (e *ErrResearchReportCannotUpdate) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("research report %s cannot be updated: %s", e.ReportID, e.Reason)
	}
	return fmt.Sprintf(
		"research report %s cannot be updated in review status %q",
		e.ReportID, e.ReviewStatus,
	)
}

// ErrResearchReportCannotDelete is raised when soft-deleting is blocked
// because the report has already been submitted or its review completed.
type ErrResearchReportCannotDelete struct {
	ReportID     string
	ReviewStatus string
}

func (e *ErrResearchReportCannotDelete) Error() string {
	return fmt.Sprintf(
		"research report %s cannot be deleted while review_status is %q",
		e.ReportID, e.ReviewStatus,
	)
}

// ErrResearchReportCannotSubmit is raised when submit is invalid for the
// report's current state.
type ErrResearchReportCannotSubmit struct {
	ReportID     string
	ReviewStatus string
}

func (e *ErrResearchReportCannotSubmit) Error() string {
	return fmt.Sprintf(
		"research report %s cannot be submitted from review_status %q",
		e.ReportID, e.ReviewStatus,
	)
}

// ErrResearchReportCannotCancelSubmit is raised when cancel-submit is invalid
// for the report's current state — i.e., not currently SUBMITTED.
type ErrResearchReportCannotCancelSubmit struct {
	ReportID     string
	ReviewStatus string
}

func (e *ErrResearchReportCannotCancelSubmit) Error() string {
	return fmt.Sprintf(
		"research report %s cannot cancel submission from review_status %q",
		e.ReportID, e.ReviewStatus,
	)
}

// ErrResearchReportCannotInvalidate is raised when invalidation is blocked —
// typically because the report has been soft-deleted or is already
// invalidated.
type ErrResearchReportCannotInvalidate struct {
	ReportID     string
	ReportStatus string
	ReviewStatus string
	Reason       string
}

func (e *ErrResearchReportCannotInvalidate) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("research report %s cannot be invalidated: %s", e.ReportID, e.Reason)
	}
	return fmt.Sprintf(
		"research report %s cannot be invalidated from report_status %q / review_status %q",
		e.ReportID, e.ReportStatus, e.ReviewStatus,
	)
}

// ErrorCode implements errcode.Coded.
func (*ErrResearchReportCannotInvalidate) ErrorCode() string {
	return errcode.CodeResearchReportInvalidateBlocked
}

// HTTPStatus implements errcode.HTTPStatus.
func (*ErrResearchReportCannotInvalidate) HTTPStatus() int { return http.StatusConflict }
