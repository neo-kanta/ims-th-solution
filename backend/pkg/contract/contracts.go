package contract

import "time"

// WorkflowStateProvider defines the cross-module interface for querying workflow state.
// Other modules (e.g., investment) use this to check if trading is allowed.
type WorkflowStateProvider interface {
	// IsTradeAllowed returns true if the given contract on the given business date
	// is in a state that allows investment transactions
	// (i.e., day has started, manager has NOT yet approved).
	IsTradeAllowed(contractID string, businessDate time.Time) (bool, error)
}

// PermissionChecker defines the cross-module interface for permission verification.
type PermissionChecker interface {
	// HasFunctionPermission checks if the user has the given function permission.
	HasFunctionPermission(userID string, permissionCode string) (bool, error)

	// HasDataPermission checks if the user has access to the given contract.
	HasDataPermission(userID string, contractID string) (bool, error)

	// GetAccessibleContracts returns the list of contract IDs the user can access.
	GetAccessibleContracts(userID string) ([]string, error)
}

// AuditLogger defines the cross-module interface for recording audit events.
type AuditLogger interface {
	// LogAction records an audit trail entry for a business action.
	LogAction(entry AuditEntry) error
}

// AuditEntry represents a single audit log record.
type AuditEntry struct {
	ActorID      string      `json:"actor_id"`
	Action       string      `json:"action"`
	Module       string      `json:"module"`
	ResourceType string      `json:"resource_type"`
	ResourceID   string      `json:"resource_id"`
	Details      interface{} `json:"details,omitempty"`
	BusinessDate time.Time   `json:"business_date"`
}
