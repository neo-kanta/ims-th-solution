package errors

import "fmt"

// BusinessError represents a domain-level business rule violation.
// These errors are safe to display to the API consumer.
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewBusinessError creates a new BusinessError.
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{Code: code, Message: message}
}

// ValidationError represents one or more field-level validation failures.
type ValidationError struct {
	Errors []FieldError `json:"errors"`
}

// FieldError represents a single field validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s — %s", e.Errors[0].Field, e.Errors[0].Message)
}

// NewValidationError creates a ValidationError from field errors.
func NewValidationError(errors ...FieldError) *ValidationError {
	return &ValidationError{Errors: errors}
}

// NotFoundError represents a resource that was not found.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// NewNotFoundError creates a new NotFoundError.
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{Resource: resource, ID: id}
}

// ForbiddenError represents an authorization failure.
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

// NewForbiddenError creates a new ForbiddenError.
func NewForbiddenError(message string) *ForbiddenError {
	return &ForbiddenError{Message: message}
}
