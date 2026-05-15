package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// IAMPermissionPort is the subset of the IAM module's API consumed by this
// adapter. Defining it here keeps the investment module decoupled from
// internal/iam at the type level.
type IAMPermissionPort interface {
	HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error)
	HasDataPermission(ctx context.Context, userID string, scopeID string) (bool, error)
	GetAccessibleContracts(ctx context.Context, userID string) ([]string, error)
}

// PermissionCheckerAdapter wraps an IAMPermissionPort to satisfy
// pkg/contract.PermissionChecker (which uses uuid.UUID and no ctx).
type PermissionCheckerAdapter struct {
	port IAMPermissionPort
}

// NewPermissionCheckerAdapter wires the adapter.
func NewPermissionCheckerAdapter(port IAMPermissionPort) *PermissionCheckerAdapter {
	return &PermissionCheckerAdapter{port: port}
}

// HasFunctionPermission implements contract.PermissionChecker.
func (a *PermissionCheckerAdapter) HasFunctionPermission(userID uuid.UUID, code string) (bool, error) {
	if a == nil || a.port == nil {
		return false, nil
	}
	return a.port.HasFunctionPermission(context.Background(), userID.String(), code)
}

// HasDataPermission implements contract.PermissionChecker.
func (a *PermissionCheckerAdapter) HasDataPermission(userID uuid.UUID, contractID string) (bool, error) {
	if a == nil || a.port == nil {
		return false, nil
	}
	return a.port.HasDataPermission(context.Background(), userID.String(), contractID)
}

// GetAccessibleContracts implements contract.PermissionChecker.
func (a *PermissionCheckerAdapter) GetAccessibleContracts(userID uuid.UUID) ([]string, error) {
	if a == nil || a.port == nil {
		return nil, nil
	}
	return a.port.GetAccessibleContracts(context.Background(), userID.String())
}

var _ contract.PermissionChecker = (*PermissionCheckerAdapter)(nil)
