// Package permission declares the permission codes owned by the
// leave_delegation module.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "leave_delegation"

// Permission codes owned by the leave_delegation module.
const (
	CodeView   = "LEAVE_VIEW"
	CodeManage = "LEAVE_MANAGE"
)

// Provider implements contract.PermissionCatalog for the leave_delegation module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of leave_delegation permission
// definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeView, Name: "Leave View", Description: "Read access to leave records and delegation assignments."},
		{Code: CodeManage, Name: "Leave Manage", Description: "Create, edit or cancel leave records and configure delegation."},
	}
}
