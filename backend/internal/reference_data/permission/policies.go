// Package permission declares the permission codes owned by the
// reference_data module. The module currently exposes no operator-grantable
// codes — canonical-security CRUD and unmapped-candidate review are gated by
// the parent /api/v1 auth middleware for now. The Provider is registered with
// cmd/seed so the catalog records the empty contribution explicitly.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "reference_data"

// Provider implements contract.PermissionCatalog for the reference_data module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of reference_data permission
// definitions. Empty by design until operator-grantable codes are added.
func (Provider) Permissions() []contract.PermissionDefinition {
	return nil
}
