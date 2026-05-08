// Package permission declares the permission codes owned by the market_data
// module. The module currently exposes no operator-grantable codes — Phase 5
// ingestion uses the system.market_data system actor and bypasses the function
// permission gate entirely. The Provider is registered with cmd/seed so the
// catalog records the empty contribution explicitly.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "market_data"

// Provider implements contract.PermissionCatalog for the market_data module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of market_data permission
// definitions. Empty by design until Phase 5+ adds operator-grantable codes.
func (Provider) Permissions() []contract.PermissionDefinition {
	return nil
}
