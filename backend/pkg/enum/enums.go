package enum

// WorkflowState represents the daily investment workflow states.
type WorkflowState string

const (
	WorkflowStatePending           WorkflowState = "PENDING"
	WorkflowStateDayStarted        WorkflowState = "DAY_STARTED"
	WorkflowStateManagerApproved   WorkflowState = "MANAGER_APPROVED"
	WorkflowStateTransactionClosed WorkflowState = "TRANSACTION_CLOSED"
	WorkflowStateAccountingClosed  WorkflowState = "ACCOUNTING_CLOSED"
)

// InvestmentAction represents stock investment recommendation types.
type InvestmentAction string

const (
	InvestmentActionBuy  InvestmentAction = "BUY"
	InvestmentActionSell InvestmentAction = "SELL"
	InvestmentActionHold InvestmentAction = "HOLD"
)

// InvestmentAction represents Fund investment types
type FundInvestmentAction string

const (
	FundInvestmentActionSubScription FundInvestmentAction = "SUBSCRIPTION"
	FundInvestmentActionRedemption   FundInvestmentAction = "REDEMPTION"
	FundInvestmentActionSwitch       FundInvestmentAction = "SWITCH"
)

// ApprovalStatus represents the status of an approval request.
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
)

// LeaveType represents types of leave.
type LeaveType string

const (
	LeaveTypeRegular   LeaveType = "REGULAR"
	LeaveTypeTemporary LeaveType = "TEMPORARY"
)

// LeaveStatus represents the status of a leave request.
type LeaveStatus string

const (
	LeaveStatusPending   LeaveStatus = "PENDING"
	LeaveStatusApproved  LeaveStatus = "APPROVED"
	LeaveStatusRejected  LeaveStatus = "REJECTED"
	LeaveStatusCancelled LeaveStatus = "CANCELLED"
)
