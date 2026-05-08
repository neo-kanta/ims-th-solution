// Package permission declares the permission codes owned by the permissions
// management module itself.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "permissions"

// Permission codes owned by the permissions module.
const (
	CodeView   = "PERMISSIONS_VIEW"
	CodeManage = "PERMISSIONS_MANAGE"
)

// Provider implements contract.PermissionCatalog for the permissions module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of permissions-management
// permission definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeView, Name: "Permissions View", Description: "Read access to groups, function rights and data rights."},
		{Code: CodeManage, Name: "Permissions Manage", Description: "Create / update / delete groups, function rights and data rights."},
	}
}
