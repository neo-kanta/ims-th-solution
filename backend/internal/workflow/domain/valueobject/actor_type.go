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
	UserID    uuid.UUID
	Username  string // snapshot of display name at transition time
	ActorType ActorType
	RequestID string   // X-Request-Id header value from chi middleware
	Roles     []string // group/role codes from JWT "rls" claim
}
