package valueobject

// OrderSide identifies the direction of a trade order.
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// IsValid returns true if the side is a known value.
func (s OrderSide) IsValid() bool {
	return s == OrderSideBuy || s == OrderSideSell
}
