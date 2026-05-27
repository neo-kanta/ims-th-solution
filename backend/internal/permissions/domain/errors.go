package domain

import (
	"fmt"
	"net/http"
)

const (
	CodeInvalidRequest       = "PERMISSION_INVALID_REQUEST"
	CodeNotFound             = "PERMISSION_NOT_FOUND"
	CodeForbidden            = "PERMISSION_FORBIDDEN"
	CodeConflict             = "PERMISSION_CONFLICT"
	CodeInvalidTransition    = "PERMISSION_INVALID_TRANSITION"
	CodeChecksFailed         = "PERMISSION_CHECKS_FAILED"
	CodeApprovalNotAllowed   = "PERMISSION_APPROVAL_NOT_ALLOWED"
	CodeMergeNotAllowed      = "PERMISSION_MERGE_NOT_ALLOWED"
	CodeRolePriorityDenied   = "PERMISSION_ROLE_PRIORITY_DENIED"
	CodeAssignmentOutOfScope = "PERMISSION_ASSIGNMENT_OUT_OF_SCOPE"
)

type PermissionError struct {
	Code    string
	Message string
	Status  int
	Details map[string]any
}

func (e *PermissionError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *PermissionError) ErrorCode() string {
	if e == nil {
		return CodeInvalidRequest
	}
	return e.Code
}

func (e *PermissionError) HTTPStatus() int {
	if e == nil || e.Status == 0 {
		return http.StatusBadRequest
	}
	return e.Status
}

func (e *PermissionError) ErrorDetails() map[string]any {
	if e == nil {
		return nil
	}
	return e.Details
}

func Invalid(message string) *PermissionError {
	return &PermissionError{Code: CodeInvalidRequest, Message: message, Status: http.StatusBadRequest}
}

func NotFound(resource string, id any) *PermissionError {
	return &PermissionError{
		Code:    CodeNotFound,
		Message: fmt.Sprintf("%s not found: %v", resource, id),
		Status:  http.StatusNotFound,
	}
}

func Forbidden(message string) *PermissionError {
	return &PermissionError{Code: CodeForbidden, Message: message, Status: http.StatusForbidden}
}

func Conflict(message string) *PermissionError {
	return &PermissionError{Code: CodeConflict, Message: message, Status: http.StatusConflict}
}

func InvalidTransition(message string) *PermissionError {
	return &PermissionError{Code: CodeInvalidTransition, Message: message, Status: http.StatusUnprocessableEntity}
}

func ChecksFailed(message string, details map[string]any) *PermissionError {
	return &PermissionError{Code: CodeChecksFailed, Message: message, Status: http.StatusUnprocessableEntity, Details: details}
}

func ApprovalNotAllowed(message string) *PermissionError {
	return &PermissionError{Code: CodeApprovalNotAllowed, Message: message, Status: http.StatusForbidden}
}

func MergeNotAllowed(message string) *PermissionError {
	return &PermissionError{Code: CodeMergeNotAllowed, Message: message, Status: http.StatusUnprocessableEntity}
}
