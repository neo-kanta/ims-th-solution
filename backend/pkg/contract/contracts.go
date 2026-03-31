package contract

import (
	"time"

	"github.com/google/uuid"
)

// WorkflowStateProvider defines the cross-module interface for querying workflow state.
type WorkflowStateProvider interface {
	IsTradeAllowed(contractID string, businessDate time.Time) (bool, error)
}

// PermissionChecker defines the cross-module interface for permission verification.
type PermissionChecker interface {
	HasFunctionPermission(userID uuid.UUID, permissionCode string) (bool, error)
	HasDataPermission(userID uuid.UUID, contractID string) (bool, error)
	GetAccessibleContracts(userID uuid.UUID) ([]string, error)
}

// LeaveChecker defines the cross-module interface for querying leave status.
// IAM does NOT own leave state — other modules implement this.
type LeaveChecker interface {
	IsOnLeave(userID uuid.UUID, date time.Time) (bool, error)
}

// AuditLogger defines the cross-module interface for recording audit events.
type AuditLogger interface {
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
