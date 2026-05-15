// Package permission declares the permission codes owned by the notification module.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "notification"

// Permission codes owned by the notification module.
const (
	CodeConfig = "NOTIFICATION_CONFIG"
)

// Provider implements contract.PermissionCatalog for the notification module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of notification permission definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeConfig, Name: "Notification Config", Description: "Configure notification templates, per-contract notification rules and disabled notifications."},
	}
}
