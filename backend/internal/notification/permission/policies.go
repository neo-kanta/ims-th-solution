// Package permission declares the permission codes owned by the notification module.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "notification"

// Permission codes owned by the notification module.
const (
	CodeConfig = "NOTIFICATION_CONFIG"
	CodeView   = "NOTIFICATION_VIEW"
	CodeEdit   = "NOTIFICATION_EDIT"
	CodeRetry  = "NOTIFICATION_RETRY"
	CodeTest   = "NOTIFICATION_TEST"
	CodeHealth = "NOTIFICATION_HEALTH"
)

// Provider implements contract.PermissionCatalog for the notification module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of notification permission definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeConfig, Name: "Notification Config", Description: "Configure notification templates, per-contract notification rules and disabled notifications."},
		{Code: CodeView, Name: "Notification View", Description: "Read notification and email outbox operational data."},
		{Code: CodeEdit, Name: "Notification Edit", Description: "General notification administration."},
		{Code: CodeRetry, Name: "Notification Retry", Description: "Retry failed or dead email outbox rows."},
		{Code: CodeTest, Name: "Notification Test", Description: "Send demo or test email through the outbox worker."},
		{Code: CodeHealth, Name: "Notification Health", Description: "Read email health and config status without secrets."},
	}
}
