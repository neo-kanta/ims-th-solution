package types

import (
	"fmt"
	"math/big"
)

// Money represents a monetary amount with currency, using arbitrary-precision arithmetic.
// Financial calculations must NEVER use float64.
type Money struct {
	Amount   *big.Rat `json:"amount"`
	Currency string   `json:"currency"`
}

// NewMoney creates a Money value from a string amount and currency code.
func NewMoney(amount string, currency string) (Money, error) {
	rat := new(big.Rat)
	if _, ok := rat.SetString(amount); !ok {
		return Money{}, fmt.Errorf("invalid money amount: %s", amount)
	}
	return Money{Amount: rat, Currency: currency}, nil
}

// ZeroMoney returns a zero-value Money for the given currency.
func ZeroMoney(currency string) Money {
	return Money{Amount: new(big.Rat), Currency: currency}
}

// String returns the string representation of the money value.
func (m Money) String() string {
	if m.Amount == nil {
		return "0 " + m.Currency
	}
	return m.Amount.FloatString(4) + " " + m.Currency
}
