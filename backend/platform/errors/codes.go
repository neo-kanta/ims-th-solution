package errors

// Standard error codes used across the application.
// These codes are returned in API error responses for client identification.
const (
	// Authentication & Authorization
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeTokenExpired = "TOKEN_EXPIRED"
	CodeTokenInvalid = "TOKEN_INVALID"

	// Validation
	CodeValidationFailed = "VALIDATION_FAILED"
	CodeInvalidInput     = "INVALID_INPUT"

	// Resources
	CodeNotFound  = "NOT_FOUND"
	CodeConflict  = "CONFLICT"
	CodeDuplicate = "DUPLICATE"

	// Business Rules
	CodeBusinessRule       = "BUSINESS_RULE_VIOLATION"
	CodeWorkflowViolation  = "WORKFLOW_VIOLATION"
	CodePermissionDenied   = "PERMISSION_DENIED"
	CodeDataScopeViolation = "DATA_SCOPE_VIOLATION"

	// Workflow-specific
	CodeDayNotStarted          = "DAY_NOT_STARTED"
	CodeManagerAlreadyApproved = "MANAGER_ALREADY_APPROVED"
	CodeTransactionLocked      = "TRANSACTION_LOCKED"

	// System
	CodeInternal = "INTERNAL_ERROR"
	CodeTimeout  = "TIMEOUT"
)
