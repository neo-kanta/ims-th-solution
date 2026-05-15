// Package permission declares the permission codes owned by the iam module.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "iam"

// Permission codes owned by the iam module.
const (
	CodeUserView       = "IAM_USER_VIEW"
	CodeUserCreate     = "IAM_USER_CREATE"
	CodeUserUpdate     = "IAM_USER_UPDATE"
	CodeUserDeactivate = "IAM_USER_DEACTIVATE"
	CodeAuditView      = "IAM_AUDIT_VIEW"
)

// Provider implements contract.PermissionCatalog for the iam module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of iam permission definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeUserView, Name: "IAM User View", Description: "Read access to user accounts and their session metadata."},
		{Code: CodeUserCreate, Name: "IAM User Create", Description: "Create new user accounts."},
		{Code: CodeUserUpdate, Name: "IAM User Update", Description: "Update user profile and identity attributes."},
		{Code: CodeUserDeactivate, Name: "IAM User Deactivate", Description: "Deactivate or reactivate a user account."},
		{Code: CodeAuditView, Name: "IAM Audit View", Description: "Read and export the iam audit trail."},
	}
}
