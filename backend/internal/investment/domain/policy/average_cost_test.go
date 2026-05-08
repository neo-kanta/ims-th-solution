package policy_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
)

func d(v string) decimal.Decimal {
	d, err := decimal.NewFromString(v)
	if err != nil {
		panic(err)
	}
	return d
}

func TestApplyBuyAverageCost(t *testing.T) {
	cases := []struct {
		name string
		in   policy.AverageCostInputs
		want policy.AverageCostResult
	}{
		{
			name: "first buy on empty position",
			in: policy.AverageCostInputs{
				OldQuantity:    decimal.Zero,
				OldAverageCost: decimal.Zero,
				BuyQuantity:    d("100"),
				BuyPriceBase:   d("50"),
				BuyFeesBase:    d("0"),
			},
			want: policy.AverageCostResult{
				NewQuantity:    d("100"),
				NewAverageCost: d("50"),
				NewCostBasis:   d("5000"),
			},
		},
		{
			name: "buy with fees raises avg cost",
			in: policy.AverageCostInputs{
				OldQuantity:    decimal.Zero,
				OldAverageCost: decimal.Zero,
				BuyQuantity:    d("100"),
				BuyPriceBase:   d("50"),
				BuyFeesBase:    d("100"),
			},
			want: policy.AverageCostResult{
				NewQuantity:    d("100"),
				NewAverageCost: d("51"),
				NewCostBasis:   d("5100"),
			},
		},
		{
			name: "second buy at higher price increases avg cost proportionally",
			in: policy.AverageCostInputs{
				OldQuantity:    d("100"),
				OldAverageCost: d("50"),
				BuyQuantity:    d("100"),
				BuyPriceBase:   d("60"),
				BuyFeesBase:    d("0"),
			},
			want: policy.AverageCostResult{
				NewQuantity:    d("200"),
				NewAverageCost: d("55"),
				NewCostBasis:   d("11000"),
			},
		},
		{
			name: "buy with zero qty is a no-op",
			in: policy.AverageCostInputs{
				OldQuantity:    d("100"),
				OldAverageCost: d("50"),
				BuyQuantity:    decimal.Zero,
				BuyPriceBase:   d("60"),
				BuyFeesBase:    d("10"),
			},
			want: policy.AverageCostResult{
				NewQuantity:    d("100"),
				NewAverageCost: d("50"),
				NewCostBasis:   d("5000"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := policy.ApplyBuyAverageCost(tc.in)
			assert.True(t, tc.want.NewQuantity.Equal(got.NewQuantity), "qty: want %s got %s", tc.want.NewQuantity, got.NewQuantity)
			assert.True(t, tc.want.NewAverageCost.Equal(got.NewAverageCost), "avg: want %s got %s", tc.want.NewAverageCost, got.NewAverageCost)
			assert.True(t, tc.want.NewCostBasis.Equal(got.NewCostBasis), "basis: want %s got %s", tc.want.NewCostBasis, got.NewCostBasis)
		})
	}
}

func TestApplySellAverageCost(t *testing.T) {
	cases := []struct {
		name             string
		in               policy.SellInputs
		wantQty, wantAvg string
		wantBasis        string
		wantPnL          string
	}{
		{
			name: "partial sell at profit, avg unchanged",
			in: policy.SellInputs{
				OldQuantity:    d("100"),
				OldAverageCost: d("50"),
				SellQuantity:   d("40"),
				SellPriceBase:  d("70"),
				SellFeesBase:   d("0"),
			},
			wantQty: "60", wantAvg: "50", wantBasis: "3000", wantPnL: "800",
		},
		{
			name: "full sell resets avg to zero",
			in: policy.SellInputs{
				OldQuantity:    d("100"),
				OldAverageCost: d("50"),
				SellQuantity:   d("100"),
				SellPriceBase:  d("60"),
				SellFeesBase:   d("0"),
			},
			wantQty: "0", wantAvg: "0", wantBasis: "0", wantPnL: "1000",
		},
		{
			name: "sell with fees reduces realised pnl",
			in: policy.SellInputs{
				OldQuantity:    d("100"),
				OldAverageCost: d("50"),
				SellQuantity:   d("10"),
				SellPriceBase:  d("60"),
				SellFeesBase:   d("5"),
			},
			wantQty: "90", wantAvg: "50", wantBasis: "4500", wantPnL: "95",
		},
		{
			name: "oversell returns inputs unchanged (defensive)",
			in: policy.SellInputs{
				OldQuantity:    d("10"),
				OldAverageCost: d("50"),
				SellQuantity:   d("100"),
				SellPriceBase:  d("60"),
				SellFeesBase:   d("0"),
			},
			wantQty: "10", wantAvg: "50", wantBasis: "500", wantPnL: "0",
		},
		{
			name: "zero qty sell is no-op",
			in: policy.SellInputs{
				OldQuantity:    d("100"),
				OldAverageCost: d("50"),
				SellQuantity:   decimal.Zero,
				SellPriceBase:  d("60"),
				SellFeesBase:   d("0"),
			},
			wantQty: "100", wantAvg: "50", wantBasis: "5000", wantPnL: "0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := policy.ApplySellAverageCost(tc.in)
			assert.True(t, d(tc.wantQty).Equal(got.NewQuantity), "qty: want %s got %s", tc.wantQty, got.NewQuantity)
			assert.True(t, d(tc.wantAvg).Equal(got.NewAverageCost), "avg: want %s got %s", tc.wantAvg, got.NewAverageCost)
			assert.True(t, d(tc.wantBasis).Equal(got.NewCostBasis), "basis: want %s got %s", tc.wantBasis, got.NewCostBasis)
			assert.True(t, d(tc.wantPnL).Equal(got.RealisedPnLBase), "pnl: want %s got %s", tc.wantPnL, got.RealisedPnLBase)
		})
	}
}
