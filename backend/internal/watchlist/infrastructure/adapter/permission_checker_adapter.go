package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// IAMPermissionPort is the subset of the IAM module API consumed by the watchlist module.
type IAMPermissionPort interface {
	HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error)
	HasDataPermission(ctx context.Context, userID string, scopeID string) (bool, error)
	GetAccessibleContracts(ctx context.Context, userID string) ([]string, error)
}

// PermissionCheckerAdapter bridges IAMPermissionPort to contract.PermissionChecker.
type PermissionCheckerAdapter struct {
	port IAMPermissionPort
}

func NewPermissionCheckerAdapter(port IAMPermissionPort) *PermissionCheckerAdapter {
	return &PermissionCheckerAdapter{port: port}
}

func (a *PermissionCheckerAdapter) HasFunctionPermission(userID uuid.UUID, code string) (bool, error) {
	if a == nil || a.port == nil {
		return false, nil
	}
	return a.port.HasFunctionPermission(context.Background(), userID.String(), code)
}

func (a *PermissionCheckerAdapter) HasDataPermission(userID uuid.UUID, contractID string) (bool, error) {
	if a == nil || a.port == nil {
		return false, nil
	}
	return a.port.HasDataPermission(context.Background(), userID.String(), contractID)
}

func (a *PermissionCheckerAdapter) GetAccessibleContracts(userID uuid.UUID) ([]string, error) {
	if a == nil || a.port == nil {
		return nil, nil
	}
	return a.port.GetAccessibleContracts(context.Background(), userID.String())
}

var _ contract.PermissionChecker = (*PermissionCheckerAdapter)(nil)
