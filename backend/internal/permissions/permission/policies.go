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

// Fine-grained permission-management codes stored in the new
// permission_function_definitions catalog. These are intentionally not emitted
// by Provider.Permissions because the legacy catalog table currently enforces
// uppercase codes for existing modules.
const (
	CodeUsersView          = "permission.users.view"
	CodeGroupsView         = "permission.groups.view"
	CodeRolesView          = "permission.roles.view"
	CodeFunctionRightsView = "permission.function_rights.view"
	CodeDataRightsView     = "permission.data_rights.view"

	CodeChangeRequestCreate  = "permission.change_request.create"
	CodeChangeRequestSubmit  = "permission.change_request.submit"
	CodeChangeRequestReview  = "permission.change_request.review"
	CodeChangeRequestApprove = "permission.change_request.approve"
	CodeChangeRequestMerge   = "permission.change_request.merge"
	CodeChangeRequestReject  = "permission.change_request.reject"
	CodeChangeRequestClose   = "permission.change_request.close"
	CodeChangeRequestCancel  = "permission.change_request.cancel"

	CodeAuditView            = "permission.audit.view"
	CodeAuditExport          = "permission.audit.export"
	CodeNotificationView     = "permission.notification.view"
	CodeNotificationEdit     = "permission.notification.edit"
	CodeLabelsManage         = "permission.change_request.create"
	CodeApprovalSettingsView = "permission.change_request.review"
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
