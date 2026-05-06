// Package permission declares the permission codes owned by the workflow
// module.
//
// Codes are stable contracts — renaming one is a breaking change for DB seed
// data and any UI that references them. The Provider returned by Catalog()
// is registered with cmd/seed to keep permissions_function_definitions in
// sync with this file.
package permission

import "github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

// ModuleName is the catalog tag for rows owned by this module.
const ModuleName = "workflow"

// Permission codes owned by the workflow module.
const (
	CodeView                    = "WORKFLOW_VIEW"
	CodeOpenDay                 = "WORKFLOW_OPEN_DAY"
	CodeApprove                 = "WORKFLOW_APPROVE"
	CodeCancelDayStart          = "WORKFLOW_CANCEL_DAY_START"
	CodeCancelApproval          = "WORKFLOW_CANCEL_APPROVAL"
	CodeCloseTransactions       = "WORKFLOW_CLOSE_TRANSACTIONS"
	CodeCancelTransactionClose  = "WORKFLOW_CANCEL_TRANSACTION_CLOSE"
	CodeCloseAccounting         = "WORKFLOW_CLOSE_ACCOUNTING"
	CodeRollbackAccountingClose = "WORKFLOW_ROLLBACK_ACCOUNTING_CLOSE"
	CodeRunScheduler            = "WORKFLOW_RUN_SCHEDULER"
	CodeExecute                 = "WORKFLOW_EXECUTE"
)

// All returns every permission code owned by this module. Retained for any
// caller that wants only the codes; structured definitions live on Provider.
func All() []string {
	defs := Provider{}.Permissions()
	codes := make([]string, 0, len(defs))
	for _, d := range defs {
		codes = append(codes, d.Code)
	}
	return codes
}

// Provider implements contract.PermissionCatalog for the workflow module.
type Provider struct{}

// Module returns the module tag stored in permissions_function_definitions.
func (Provider) Module() string { return ModuleName }

// Permissions returns the canonical list of workflow permission definitions.
func (Provider) Permissions() []contract.PermissionDefinition {
	return []contract.PermissionDefinition{
		{Code: CodeView, Name: "Workflow View", Description: "Read access to workflow state, day-state queries, and scheduler views."},
		{Code: CodeOpenDay, Name: "Workflow Open Day", Description: "Permission to execute the day-start operation for a contract."},
		{Code: CodeApprove, Name: "Workflow Approve", Description: "Permission to perform the manager-approval transition."},
		{Code: CodeCancelDayStart, Name: "Workflow Cancel Day Start", Description: "Cancel an applied day-start, reverting to the prior state."},
		{Code: CodeCancelApproval, Name: "Workflow Cancel Approval", Description: "Cancel an applied manager approval, reverting to the prior state."},
		{Code: CodeCloseTransactions, Name: "Workflow Close Transactions", Description: "Transition the contract day to TRANSACTION_CLOSED."},
		{Code: CodeCancelTransactionClose, Name: "Workflow Cancel Transaction Close", Description: "Cancel a TRANSACTION_CLOSED transition."},
		{Code: CodeCloseAccounting, Name: "Workflow Close Accounting", Description: "Transition the contract day to ACCOUNTING_CLOSED."},
		{Code: CodeRollbackAccountingClose, Name: "Workflow Rollback Accounting Close", Description: "Roll back an ACCOUNTING_CLOSED transition."},
		{Code: CodeRunScheduler, Name: "Workflow Run Scheduler", Description: "Trigger the workflow scheduler manually."},
		{Code: CodeExecute, Name: "Workflow Execute", Description: "Generic execute capability used by legacy workflow grants."},
	}
}
