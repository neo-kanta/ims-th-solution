package policy

import (
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PostPreconditionViolation enumerates why a proposed transaction is rejected
// before any DB write.
type PostPreconditionViolation string

const (
	PostViolationPortfolioInactive     PostPreconditionViolation = "PORTFOLIO_INACTIVE"
	PostViolationFundInactive          PostPreconditionViolation = "FUND_INACTIVE"
	PostViolationInstrumentNotTradable PostPreconditionViolation = "INSTRUMENT_NOT_TRADABLE"
	PostViolationTradingNotAllowed     PostPreconditionViolation = "TRADING_NOT_ALLOWED"
	PostViolationTransactionLocked     PostPreconditionViolation = "TRANSACTION_LOCKED"
	PostViolationQuantityNotPositive   PostPreconditionViolation = "QUANTITY_NOT_POSITIVE"
	PostViolationPriceNotPositive      PostPreconditionViolation = "PRICE_NOT_POSITIVE"
	PostViolationLotSize               PostPreconditionViolation = "LOT_SIZE_VIOLATED"
	PostViolationTickSize              PostPreconditionViolation = "TICK_SIZE_VIOLATED"
	PostViolationOversell              PostPreconditionViolation = "OVERSELL"
	PostViolationCurrencyMismatch      PostPreconditionViolation = "CURRENCY_MISMATCH"
	PostViolationFXMissing             PostPreconditionViolation = "FX_RATE_MISSING"
	PostViolationFeesNegative          PostPreconditionViolation = "FEES_NEGATIVE"
	// PostViolationModelLedgerBlocked rejects a cash movement on a MODEL
	// portfolio. A MODEL portfolio is a target-allocation template with no real
	// ledger, so cash movements never enter it (confirmed owner policy). This is
	// enforced in the cash-movement gate in the post pipeline, not in
	// EvaluatePost, because it depends on portfolio type + transaction type.
	PostViolationModelLedgerBlocked PostPreconditionViolation = "MODEL_LEDGER_BLOCKED"
	// PostViolationApprovalUnavailable rejects a LIVE cash movement when the
	// approval engine (or its staging repository) is not wired. Fail closed —
	// a LIVE cash movement must never post immediately, bypassing approval.
	PostViolationApprovalUnavailable PostPreconditionViolation = "APPROVAL_GATE_UNAVAILABLE"
)

// PostInputs is the validation input for a proposed transaction. The handler
// loads these from the repository before calling EvaluatePost.
type PostInputs struct {
	Type       vo.TransactionType
	Portfolio  *entity.Portfolio
	Fund       *entity.Fund
	Instrument *entity.Instrument

	Quantity     decimal.Decimal
	Price        decimal.Decimal
	Currency     string
	Fees         decimal.Decimal
	FxRateToBase *decimal.Decimal

	// Position is the CURRENT projection for (portfolio, instrument). May be
	// nil when there is no existing position.
	Position *entity.PortfolioPosition

	// IsTradeAllowed is the result of WorkflowStateProvider.IsTradeAllowed.
	IsTradeAllowed bool
	// IsTransactionLocked is the result of WorkflowStateProvider.IsTransactionLocked.
	IsTransactionLocked bool

	// AllowForcePost is true when the caller has the FORCE_POST permission
	// (only applicable to REVERSAL on a locked day).
	AllowForcePost bool
}

// EvaluatePost returns the first failing precondition, or "" when the inputs
// are valid. Pure function — no I/O. Caller maps the result to a typed error.
func EvaluatePost(in PostInputs) PostPreconditionViolation {
	if in.Portfolio == nil || !in.Portfolio.IsActive() {
		return PostViolationPortfolioInactive
	}
	// Fund is nil for a fund-less portfolio ("Bind with Fund: N") — that is a
	// valid state, not a violation. Only an existing-but-inactive fund is.
	if in.Fund != nil && !in.Fund.IsActive() {
		return PostViolationFundInactive
	}
	if in.Fees.Sign() < 0 {
		return PostViolationFeesNegative
	}

	// Workflow gates apply to all posts. FORCE_POST may only bypass the lock
	// for REVERSAL types, never the day-not-open check.
	if !in.IsTradeAllowed {
		return PostViolationTradingNotAllowed
	}
	if in.IsTransactionLocked {
		if !(in.AllowForcePost && in.Type == vo.TransactionTypeReversal) {
			return PostViolationTransactionLocked
		}
	}

	if in.Type.IsSecurityTrade() {
		if in.Instrument == nil || !in.Instrument.IsTradableNow() {
			return PostViolationInstrumentNotTradable
		}
		if in.Quantity.Sign() <= 0 {
			return PostViolationQuantityNotPositive
		}
		if in.Price.Sign() <= 0 {
			return PostViolationPriceNotPositive
		}
		if in.Instrument.LotSize > 1 {
			lot := decimal.NewFromInt(int64(in.Instrument.LotSize))
			if !in.Quantity.Mod(lot).IsZero() {
				return PostViolationLotSize
			}
		}
		if in.Currency != in.Instrument.Currency {
			return PostViolationCurrencyMismatch
		}

		// Sells must not oversell.
		if in.Type.IsSellLike() {
			have := decimal.Zero
			if in.Position != nil {
				have = in.Position.Quantity
			}
			if in.Quantity.GreaterThan(have) {
				return PostViolationOversell
			}
		}
	}

	// FX is required when the trade currency differs from the portfolio base
	// currency. Silent 1.0 fallback is forbidden.
	if in.Currency != in.Portfolio.BaseCurrency {
		if in.FxRateToBase == nil || in.FxRateToBase.Sign() <= 0 {
			return PostViolationFXMissing
		}
	}

	return ""
}
