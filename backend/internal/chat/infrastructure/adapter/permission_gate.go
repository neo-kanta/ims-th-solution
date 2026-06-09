package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/service"
)

// iamFunctionPermPort is the structural slice of IAM the chat module needs.
// Declaring it here (rather than importing iam) keeps the chat module's
// boundary intact — *iam.Module satisfies this shape, so main.go can pass it
// directly without chat ever importing iam's internal package.
type iamFunctionPermPort interface {
	HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error)
}

// NewPermissionGate adapts an IAM function-permission port to the chat
// module's PermissionGate, converting the uuid user id to the string id IAM
// expects.
func NewPermissionGate(iam iamFunctionPermPort) service.PermissionGate {
	return &permissionGate{iam: iam}
}

type permissionGate struct {
	iam iamFunctionPermPort
}

func (g *permissionGate) HasFunctionPermission(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	return g.iam.HasFunctionPermission(ctx, userID.String(), code)
}
