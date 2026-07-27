package domain

import (
	"fmt"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
)

// ErrFundNotFound signals a fund lookup miss.
type ErrFundNotFound struct{ FundID string }

func (e *ErrFundNotFound) Error() string {
	return fmt.Sprintf("investment fund not found: %s", e.FundID)
}

// ErrPortfolioNotFound signals a portfolio lookup miss.
type ErrPortfolioNotFound struct{ PortfolioID string }

func (e *ErrPortfolioNotFound) Error() string {
	return fmt.Sprintf("investment portfolio not found: %s", e.PortfolioID)
}

// ErrInstrumentNotFound signals an instrument lookup miss.
type ErrInstrumentNotFound struct{ InstrumentID string }

func (e *ErrInstrumentNotFound) Error() string {
	return fmt.Sprintf("investment instrument not found: %s", e.InstrumentID)
}

// ErrTransactionNotFound signals a ledger row lookup miss.
type ErrTransactionNotFound struct{ TransactionID string }

func (e *ErrTransactionNotFound) Error() string {
	return fmt.Sprintf("investment transaction not found: %s", e.TransactionID)
}

// ErrCashRequestNotFound signals a cash-request lookup miss.
type ErrCashRequestNotFound struct{ RequestID string }

func (e *ErrCashRequestNotFound) Error() string {
	return fmt.Sprintf("cash request not found: %s", e.RequestID)
}

// ErrCashRequestNotCancellable signals a cancel attempt on a request that is
// no longer PENDING.
type ErrCashRequestNotCancellable struct {
	RequestID     string
	CurrentStatus string
}

func (e *ErrCashRequestNotCancellable) Error() string {
	return fmt.Sprintf("cash request %s cannot be cancelled from status %s", e.RequestID, e.CurrentStatus)
}

// ErrCashRequestForbidden signals that the actor is not authorised to act on
// the cash request (e.g. cancel by a non-submitter). Maps to HTTP 403.
type ErrCashRequestForbidden struct {
	RequestID string
	Reason    string
}

func (e *ErrCashRequestForbidden) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("not authorised to act on cash request %s", e.RequestID)
	}
	return fmt.Sprintf("not authorised to act on cash request %s: %s", e.RequestID, e.Reason)
}

// ErrPostPreconditionFailed wraps a violation enum from the policy layer.
type ErrPostPreconditionFailed struct {
	Violation string
	Detail    string
}

func (e *ErrPostPreconditionFailed) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("transaction post precondition failed: %s", e.Violation)
	}
	return fmt.Sprintf("transaction post precondition failed: %s: %s", e.Violation, e.Detail)
}

// ErrTransactionAlreadyReversed prevents reverse-of-reverse and double-reverse.
type ErrTransactionAlreadyReversed struct {
	TransactionID string
}

func (e *ErrTransactionAlreadyReversed) Error() string {
	return fmt.Sprintf("transaction %s has already been reversed", e.TransactionID)
}

// ErrCannotReverseReversal blocks reversing a row that itself is a reversal.
type ErrCannotReverseReversal struct{ TransactionID string }

func (e *ErrCannotReverseReversal) Error() string {
	return fmt.Sprintf("transaction %s is itself a reversal and cannot be reversed", e.TransactionID)
}

// ErrPortfolioVersionMismatch is raised on optimistic-lock conflicts.
type ErrPortfolioVersionMismatch struct {
	PortfolioID     string
	ExpectedVersion int
	ActualVersion   int
}

func (e *ErrPortfolioVersionMismatch) Error() string {
	return fmt.Sprintf(
		"portfolio %s version mismatch: expected %d, got %d",
		e.PortfolioID, e.ExpectedVersion, e.ActualVersion,
	)
}

// ErrFundVersionMismatch mirrors ErrPortfolioVersionMismatch for funds.
type ErrFundVersionMismatch struct {
	FundID          string
	ExpectedVersion int
	ActualVersion   int
}

func (e *ErrFundVersionMismatch) Error() string {
	return fmt.Sprintf(
		"fund %s version mismatch: expected %d, got %d",
		e.FundID, e.ExpectedVersion, e.ActualVersion,
	)
}

// ErrCodeAlreadyExists is raised when a unique business code (fund/portfolio
// code, instrument ticker) collides with an existing live row.
type ErrCodeAlreadyExists struct {
	Resource string
	Code     string
}

func (e *ErrCodeAlreadyExists) Error() string {
	return fmt.Sprintf("%s code %q already exists", e.Resource, e.Code)
}

// ErrAmbiguousPortfolioCode is raised by PortfolioRepository.GetByCode when
// more than one alive portfolio shares the requested code. The DB enforces
// global active-code uniqueness via uq_inv_portfolios_code_alive (migration
// 20260703000001_investment__portfolio_v2_hardening), so this should be
// unreachable in a correctly-migrated database — GetByCode still refuses to
// guess which portfolio the caller means, as defense-in-depth against
// pre-constraint legacy data or a future regression, rather than silently
// returning one.
type ErrAmbiguousPortfolioCode struct {
	Code  string
	Count int
}

func (e *ErrAmbiguousPortfolioCode) Error() string {
	return fmt.Sprintf(
		"portfolio code %q is ambiguous: %d alive portfolios share it; global code uniqueness is not yet enforced",
		e.Code, e.Count,
	)
}

// ErrFundHasActivePortfolios prevents soft-deleting a fund that still owns
// active portfolios.
type ErrFundHasActivePortfolios struct {
	FundID string
	Count  int
}

func (e *ErrFundHasActivePortfolios) Error() string {
	return fmt.Sprintf(
		"fund %s cannot be deleted: %d active portfolio(s) remain",
		e.FundID, e.Count,
	)
}

// ErrPortfolioHasOpenActivity prevents soft-deleting a portfolio with
// non-zero positions or recent same-day transactions.
type ErrPortfolioHasOpenActivity struct {
	PortfolioID string
	Detail      string
}

func (e *ErrPortfolioHasOpenActivity) Error() string {
	return fmt.Sprintf(
		"portfolio %s cannot be deleted: %s",
		e.PortfolioID, e.Detail,
	)
}

// ErrIncompleteFundValuation is raised by ComputeFundAUMHandler when not every
// portfolio under the fund has a valuation_snapshot for the requested
// business date, or when the existing snapshots disagree on price_set_hash /
// valuation_ccy. The Reason field is one of:
//
//	"MISSING_PORTFOLIO_SNAPSHOT" — at least one portfolio has no snapshot for the date.
//	"PRICE_SET_HASH_MISMATCH"    — snapshots span more than one price set.
//	"VALUATION_CCY_MISMATCH"     — snapshots span more than one valuation currency.
type ErrIncompleteFundValuation struct {
	FundID       string
	BusinessDate string
	Reason       string
	Detail       string
}

func (e *ErrIncompleteFundValuation) Error() string {
	return fmt.Sprintf(
		"fund %s incomplete valuation for %s (%s): %s",
		e.FundID, e.BusinessDate, e.Reason, e.Detail,
	)
}

// ErrorCode implements errcode.Coded.
func (*ErrIncompleteFundValuation) ErrorCode() string { return errcode.CodeIncompleteFundValuation }

// ErrorDetails implements errcode.Detailed.
func (e *ErrIncompleteFundValuation) ErrorDetails() map[string]any {
	return map[string]any{
		"fund_id":       e.FundID,
		"business_date": e.BusinessDate,
		"reason":        e.Reason,
	}
}
