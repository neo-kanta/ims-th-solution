package valueobject

// TransactionType is the discriminator for portfolio_transactions rows.
// Stable string values match the SQL CHECK constraint in
// investment__portfolio_transactions.
type TransactionType string

const (
	TransactionTypeBuy          TransactionType = "BUY"
	TransactionTypeSell         TransactionType = "SELL"
	TransactionTypeSubscription TransactionType = "SUBSCRIPTION"
	TransactionTypeRedemption   TransactionType = "REDEMPTION"
	TransactionTypeCashIn       TransactionType = "CASH_IN"
	TransactionTypeCashOut      TransactionType = "CASH_OUT"
	TransactionTypeFee          TransactionType = "FEE"
	TransactionTypeDividend     TransactionType = "DIVIDEND"
	TransactionTypeReversal     TransactionType = "REVERSAL"
)

// IsSecurityTrade returns true for types that require an instrument and a
// quantity/price.
func (t TransactionType) IsSecurityTrade() bool {
	switch t {
	case TransactionTypeBuy, TransactionTypeSell,
		TransactionTypeSubscription, TransactionTypeRedemption:
		return true
	}
	return false
}

// IsCashOnly returns true for types that touch only cash (no instrument).
func (t TransactionType) IsCashOnly() bool {
	switch t {
	case TransactionTypeCashIn, TransactionTypeCashOut,
		TransactionTypeFee, TransactionTypeDividend:
		return true
	}
	return false
}

// IsBuyLike means the type increases position quantity.
func (t TransactionType) IsBuyLike() bool {
	return t == TransactionTypeBuy || t == TransactionTypeSubscription
}

// IsSellLike means the type decreases position quantity.
func (t TransactionType) IsSellLike() bool {
	return t == TransactionTypeSell || t == TransactionTypeRedemption
}

// IsValid returns true for any recognised transaction type.
func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionTypeBuy, TransactionTypeSell,
		TransactionTypeSubscription, TransactionTypeRedemption,
		TransactionTypeCashIn, TransactionTypeCashOut,
		TransactionTypeFee, TransactionTypeDividend,
		TransactionTypeReversal:
		return true
	}
	return false
}
