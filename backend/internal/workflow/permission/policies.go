// Package permission declares the permission codes owned by the workflow module.
//
// These constants are the single source of truth for permission names. The
// permissions module seeds these codes into permissions__function_permissions;
// transport middleware uses them when gating routes.
package permission

// Permission codes owned by the workflow module.
//
// Naming convention: WORKFLOW_<ACTION>. Codes are stable contracts — renaming
// one is a breaking change for DB seed data and any UI that references them.
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
)

// All returns every permission code owned by this module.
// Used by the permissions module seeder.
func All() []string {
	return []string{
		CodeView,
		CodeOpenDay,
		CodeApprove,
		CodeCancelDayStart,
		CodeCancelApproval,
		CodeCloseTransactions,
		CodeCancelTransactionClose,
		CodeCloseAccounting,
		CodeRollbackAccountingClose,
		CodeRunScheduler,
	}
}
