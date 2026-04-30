package domain

import "fmt"

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
