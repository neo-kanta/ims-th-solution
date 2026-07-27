package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PortfolioCashRequest is the mutable staging entity for a LIVE-portfolio cash
// movement (CASH_IN/CASH_OUT/FEE/DIVIDEND) that must clear the approval engine
// before it is materialized into the append-only ledger.
//
// Unlike PortfolioTransaction this entity is MUTABLE — its status advances
// through PENDING → APPROVED/REJECTED/CANCELLED. The real ledger row is created
// only on approval and its id is recorded in ResultingTxnID.
type PortfolioCashRequest struct {
	ID          uuid.UUID
	PortfolioID uuid.UUID
	// FundID is nil for a fund-less portfolio; the data-scope key falls back to
	// PortfolioID in that case.
	FundID *uuid.UUID

	TransactionType vo.TransactionType
	// Amount is a positive magnitude; the signed cash impact is derived from
	// TransactionType at materialization.
	Amount    decimal.Decimal
	Currency  string
	Fees      decimal.Decimal
	ValueDate time.Time
	Memo      string

	Status vo.CashRequestStatus

	ApprovalRequestID *uuid.UUID
	ResultingTxnID    *uuid.UUID

	// IdempotencyKey is the effective dedup key. It is the client-supplied
	// Idempotency-Key header when present, otherwise a server-derived key
	// ("auto:" || RequestFingerprint) so that even a no-key retry of a
	// byte-identical request collapses onto this request via the partial unique
	// index on (PortfolioID, IdempotencyKey). nil only for legacy rows created
	// before the derived-key change.
	IdempotencyKey *string

	// RequestFingerprint is a deterministic hex sha256 over the canonical
	// request payload {submitter_id, portfolio_id, transaction_type, amount,
	// fees, currency, value_date, memo}. Written once at creation and NEVER
	// mutated (status/memo changes must not touch it). Used to reject a keyed
	// retry that carries a DIFFERENT movement, and to derive IdempotencyKey for
	// the no-key path. nil only for legacy rows created before the column
	// existed (those fall back to a field-by-field comparison).
	RequestFingerprint *string

	SubmittedBy uuid.UUID
	SubmittedAt time.Time
	DecidedBy   *uuid.UUID
	DecidedAt   *time.Time

	Version int

	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	UpdatedBy uuid.UUID
}

// IsPending reports whether the request is still awaiting a decision.
func (r *PortfolioCashRequest) IsPending() bool {
	return r != nil && r.Status == vo.CashRequestStatusPending
}

// IsMaterialized reports whether the approved request has already produced a
// real ledger transaction. Used as the idempotency guard in the approval
// callback.
func (r *PortfolioCashRequest) IsMaterialized() bool {
	return r != nil && r.ResultingTxnID != nil
}

// ScopeID returns the data-permission scope key: the fund id when the portfolio
// is fund-bound, otherwise the portfolio's own id (fund-optional fallback).
func (r *PortfolioCashRequest) ScopeID() uuid.UUID {
	if r == nil {
		return uuid.Nil
	}
	if r.FundID != nil {
		return *r.FundID
	}
	return r.PortfolioID
}
