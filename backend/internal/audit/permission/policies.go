// Package permission declares the permission codes owned by the audit module.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "audit"

// Permission codes owned by the audit module.
const (
	CodeView = "AUDIT_VIEW"
)

// Provider implements contract.PermissionCatalog for the audit module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of audit permission definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeView, Name: "Audit View", Description: "Read the cross-module audit trail (separate from IAM_AUDIT_VIEW which scopes the iam-owned audit feed)."},
	}
}
