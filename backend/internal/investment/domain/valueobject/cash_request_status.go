package valueobject

// CashRequestStatus is the lifecycle status of a LIVE-portfolio cash-movement
// approval request (investment__portfolio_cash_requests). Stable string values
// match the SQL CHECK constraint chk_inv_cash_req_status.
type CashRequestStatus string

const (
	// CashRequestStatusPending is the initial state after submission — awaiting
	// an approval decision.
	CashRequestStatusPending CashRequestStatus = "PENDING"
	// CashRequestStatusApproved means the request was approved AND its real
	// ledger transaction was materialized (resulting_txn_id is set).
	CashRequestStatusApproved CashRequestStatus = "APPROVED"
	// CashRequestStatusRejected means an approver rejected the request; no
	// ledger transaction was posted.
	CashRequestStatusRejected CashRequestStatus = "REJECTED"
	// CashRequestStatusCancelled means the submitter cancelled the request
	// before an approver acted; no ledger transaction was posted.
	CashRequestStatusCancelled CashRequestStatus = "CANCELLED"
)

// IsTerminal reports whether the request can no longer transition.
func (s CashRequestStatus) IsTerminal() bool {
	switch s {
	case CashRequestStatusApproved, CashRequestStatusRejected, CashRequestStatusCancelled:
		return true
	}
	return false
}

// IsValid reports whether the value is a recognised cash-request status.
func (s CashRequestStatus) IsValid() bool {
	switch s {
	case CashRequestStatusPending, CashRequestStatusApproved,
		CashRequestStatusRejected, CashRequestStatusCancelled:
		return true
	}
	return false
}
