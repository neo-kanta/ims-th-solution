// Package permission declares the permission codes owned by the integration
// module.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "integration"

// Permission codes owned by the integration module.
const (
	DashboardView = "INTEGRATION_DASHBOARD_VIEW"
)

// Provider implements contract.PermissionCatalog for the integration module.
type Provider struct{}

// Module returns the module tag.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of integration permission definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{
			Code:        DashboardView,
			Name:        "Integration Dashboard View",
			Description: "View the personal task feed, workflow states, compliance alerts, and pending approvals on the dashboard.",
		},
	}
}
