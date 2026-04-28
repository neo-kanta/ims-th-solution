// Package valueobject holds immutable value types used across the investment domain.
package valueobject

// OrderSide identifies the direction of a trade.
type OrderSide string

const (
	OrderSideBuy    OrderSide = "BUY"    // SUBSCRIPTION
	OrderSideSell   OrderSide = "SELL"   // REDEMPTION
	OrderSideSwitch OrderSide = "SWITCH" // SWITCH
)

// IsValid returns true for recognised order sides.
func (s OrderSide) IsValid() bool {
	return s == OrderSideBuy || s == OrderSideSell
}
