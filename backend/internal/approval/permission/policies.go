// Package permission declares the permission codes owned by the approval module.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "approval"

// Permission codes owned by the approval module.
const (
	CodeView   = "APPROVAL_VIEW"
	CodeConfig = "APPROVAL_CONFIG"
)

// Provider implements contract.PermissionCatalog for the approval module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of approval permission definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeView, Name: "Approval View", Description: "Read access to approval flows, groups, teams and signatures."},
		{Code: CodeConfig, Name: "Approval Config", Description: "Configure approval flows, groups, teams and digital-signature requirements."},
	}
}
