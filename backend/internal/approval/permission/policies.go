// Package permission declares the permission codes owned by the approval module.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "approval"

// Permission codes owned by the approval module.
//
// The fine-grained codes below are the source of truth for the generic approval
// engine. The two legacy codes (CodeView / CodeConfig) predate this module and
// are retained for backward compatibility with existing grants; new routes use
// the fine-grained codes.
const (
	// Legacy (retained for backward compatibility).
	CodeView   = "APPROVAL_VIEW"
	CodeConfig = "APPROVAL_CONFIG"

	// Runtime.
	CodeViewInbox   = "APPROVAL_VIEW_INBOX"
	CodeViewRequest = "APPROVAL_VIEW_REQUEST"
	CodeSubmit      = "APPROVAL_SUBMIT"
	CodeApprove     = "APPROVAL_APPROVE"
	CodeReject      = "APPROVAL_REJECT"
	CodeWithdraw    = "APPROVAL_WITHDRAW"
	CodeCancel      = "APPROVAL_CANCEL"
	CodeRevoke      = "APPROVAL_REVOKE"
	CodeAuditView   = "APPROVAL_AUDIT_VIEW"

	// Configuration.
	CodeConfigView    = "APPROVAL_CONFIG_VIEW"
	CodeGroupManage   = "APPROVAL_GROUP_MANAGE"
	CodeTeamManage    = "APPROVAL_TEAM_MANAGE"
	CodeProcessManage = "APPROVAL_PROCESS_MANAGE"
)

// Provider implements contract.PermissionCatalog for the approval module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of approval permission definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeView, Name: "Approval View", Description: "Legacy read access to approval flows (retained for backward compatibility)."},
		{Code: CodeConfig, Name: "Approval Config", Description: "Legacy configure access to approval flows (retained for backward compatibility)."},

		{Code: CodeViewInbox, Name: "Approval View Inbox", Description: "View the personal approval inbox of pending tasks."},
		{Code: CodeViewRequest, Name: "Approval View Request", Description: "View approval requests, timeline and signatures."},
		{Code: CodeSubmit, Name: "Approval Submit", Description: "Submit a business object into the approval workflow."},
		{Code: CodeApprove, Name: "Approval Approve", Description: "Approve an assigned approval task."},
		{Code: CodeReject, Name: "Approval Reject", Description: "Reject an assigned approval task."},
		{Code: CodeWithdraw, Name: "Approval Withdraw", Description: "Withdraw an approval request you submitted."},
		{Code: CodeCancel, Name: "Approval Cancel", Description: "Cancel an in-flight approval request (privileged)."},
		{Code: CodeRevoke, Name: "Approval Revoke", Description: "Revoke a previously approved request, reopening the subject for correction."},
		{Code: CodeAuditView, Name: "Approval Audit View", Description: "View the immutable approval audit timeline."},

		{Code: CodeConfigView, Name: "Approval Config View", Description: "Read approval configuration (groups, teams, processes)."},
		{Code: CodeGroupManage, Name: "Approval Group Manage", Description: "Create and manage approval groups and members."},
		{Code: CodeTeamManage, Name: "Approval Team Manage", Description: "Create and manage approval teams, contracts and members."},
		{Code: CodeProcessManage, Name: "Approval Process Manage", Description: "Create and manage approval process configurations and stages."},
	}
}
