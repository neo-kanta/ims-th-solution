package domain

import "errors"

// Sentinel domain errors for the approval module. Transport maps these to HTTP
// statuses uniformly (see transport/handler).
var (
	// ErrValidation indicates a malformed or invalid request payload (400).
	ErrValidation = errors.New("approval: validation error")
	// ErrNotFound indicates a referenced entity does not exist (404).
	ErrNotFound = errors.New("approval: not found")
	// ErrConfigNotFound indicates no approval process is configured (422).
	ErrConfigNotFound = errors.New("approval: no approval process configured for this subject")
	// ErrConflict indicates a state conflict, e.g. stale/duplicate action (409).
	ErrConflict = errors.New("approval: conflict")
	// ErrForbidden indicates the actor is not permitted to perform the action (403).
	ErrForbidden = errors.New("approval: forbidden")
	// ErrSelfApproval indicates a maker attempted to approve their own request (403).
	ErrSelfApproval = errors.New("approval: submitter cannot approve their own request")
	// ErrDuplicateActiveRequest indicates an active request already exists for the subject (409).
	ErrDuplicateActiveRequest = errors.New("approval: an active approval request already exists for this subject")
	// ErrTaskNotPending indicates the task was already actioned (409).
	ErrTaskNotPending = errors.New("approval: task is no longer pending")
	// ErrStaleTask indicates the task belongs to a stage that has already advanced (409).
	ErrStaleTask = errors.New("approval: task belongs to a superseded stage")
	// ErrNotAssigned indicates the actor is not the assignee or a valid delegate (403).
	ErrNotAssigned = errors.New("approval: you are not assigned to this task")
	// ErrRequestNotActionable indicates the request is not in a state that accepts actions (409).
	ErrRequestNotActionable = errors.New("approval: request is not awaiting approval")
)

// DomainError is a typed error carrying a sentinel kind plus a human message.
// It lets handlers map to status codes via errors.Is while preserving detail.
type DomainError struct {
	Kind    error
	Message string
}

// Error implements error.
func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Kind != nil {
		return e.Kind.Error()
	}
	return "approval: error"
}

// Unwrap exposes the sentinel kind for errors.Is.
func (e *DomainError) Unwrap() error { return e.Kind }

// NewError builds a DomainError of the given kind with a custom message.
func NewError(kind error, message string) *DomainError {
	return &DomainError{Kind: kind, Message: message}
}

// Validation is a convenience constructor for validation errors.
func Validation(message string) *DomainError { return NewError(ErrValidation, message) }

// Forbidden is a convenience constructor for forbidden errors.
func Forbidden(message string) *DomainError { return NewError(ErrForbidden, message) }

// Conflict is a convenience constructor for conflict errors.
func Conflict(message string) *DomainError { return NewError(ErrConflict, message) }

// NotFound is a convenience constructor for not-found errors.
func NotFound(message string) *DomainError { return NewError(ErrNotFound, message) }
