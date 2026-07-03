// Package permission declares the permission codes owned by the watchlist module.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

const ModuleName = "watchlist"

const (
	CodeView     = "WATCHLIST_VIEW"
	CodeManage   = "WATCHLIST_MANAGE"
	CodeAlertAck = "WATCHLIST_ALERT_ACK"
	CodeEvaluate = "WATCHLIST_EVALUATE"
	CodeAdmin    = "WATCHLIST_ADMIN"
)

func All() []string {
	defs := Provider{}.Permissions()
	codes := make([]string, 0, len(defs))
	for _, d := range defs {
		codes = append(codes, d.Code)
	}
	return codes
}

// Provider implements contract.PermissionCatalog for the watchlist module.
type Provider struct{}

func (Provider) Module() string { return ModuleName }

func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeView, Name: "Watchlist View", Description: "List visible watchlist items and alert events."},
		{Code: CodeManage, Name: "Watchlist Manage", Description: "Create, update, disable, and soft-delete watchlist items and threshold rules."},
		{Code: CodeAlertAck, Name: "Watchlist Alert Acknowledge", Description: "Acknowledge visible alert events."},
		{Code: CodeEvaluate, Name: "Watchlist Evaluate", Description: "Manually trigger operational rule evaluation. Normal users must not receive this permission."},
		{Code: CodeAdmin, Name: "Watchlist Admin", Description: "Optional override for approved admin/risk/audit use cases."},
	}
}
