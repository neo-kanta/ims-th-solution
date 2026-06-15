package valueobject

import "github.com/google/uuid"

// ActorType distinguishes human operators from automated system actors.
type ActorType string

const (
	ActorTypeHuman  ActorType = "HUMAN"
	ActorTypeSystem ActorType = "SYSTEM"
)

// ActorContext carries the identity of whoever triggered a workflow command.
// It is populated at the transport layer from the JWT claims and request
// metadata, then threaded through to command handlers and transition records.
type ActorContext struct {
	UserID      uuid.UUID
	Username    string // snapshot of display name at transition time
	AccountCode string // equals Username in current JWT (no separate accountCode field)
	ActorType   ActorType
	RequestID   string   // X-Request-Id header value from chi middleware
	Roles       []string // group/role codes from JWT "rls" claim
	// IsAdminOverride is set by the new daily transport when Admin group bypasses
	// the configured approver check. Logged immutably in workflow__transition_log.
	IsAdminOverride bool
}
