package service

import (
	"context"

	"github.com/google/uuid"
)

// PermissionGate is the chat module's authorization port. The agent loop calls
// it BEFORE executing any tool, so a user can never invoke a tool they lack
// permission for — even though the downstream REST call would also reject it.
// This is defense-in-depth: the gate is the first, in-process check; the
// MCP-server-to-REST call (with the user's own token) is the authoritative
// second check that also enforces row-level data scoping.
type PermissionGate interface {
	HasFunctionPermission(ctx context.Context, userID uuid.UUID, code string) (bool, error)
}

// AllowAllPermissionGate is a permissive gate for tests/dev only. Production
// wiring MUST supply a real gate backed by IAM.
type AllowAllPermissionGate struct{}

// HasFunctionPermission always allows.
func (AllowAllPermissionGate) HasFunctionPermission(context.Context, uuid.UUID, string) (bool, error) {
	return true, nil
}
