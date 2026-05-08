package policy_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

func activePortfolio() *entity.Portfolio {
	return &entity.Portfolio{
		ID:           uuid.New(),
		Status:       vo.PortfolioStatusActive,
		BaseCurrency: "THB",
	}
}
func activeFund() *entity.Fund {
	return &entity.Fund{ID: uuid.New(), Status: vo.FundStatusActive}
}
func activeInstrument(currency string, lot int) *entity.Instrument {
	return &entity.Instrument{
		ID:         uuid.New(),
		Status:     vo.InstrumentStatusActive,
		IsTradable: true,
		Currency:   currency,
		LotSize:    lot,
	}
}
func positionOf(qty string) *entity.PortfolioPosition {
	return &entity.PortfolioPosition{Quantity: d(qty)}
}

func TestEvaluatePost(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	_ = now

	cases := []struct {
		name string
		in   policy.PostInputs
		want policy.PostPreconditionViolation
	}{
		{
			name: "valid buy in base currency",
			in: policy.PostInputs{
				Type:           vo.TransactionTypeBuy,
				Portfolio:      activePortfolio(),
				Fund:           activeFund(),
				Instrument:     activeInstrument("THB", 1),
				Quantity:       d("100"),
				Price:          d("50"),
				Currency:       "THB",
				IsTradeAllowed: true,
			},
			want: "",
		},
		{
			name: "buy in foreign ccy without FX rejected",
			in: policy.PostInputs{
				Type:           vo.TransactionTypeBuy,
				Portfolio:      activePortfolio(),
				Fund:           activeFund(),
				Instrument:     activeInstrument("USD", 1),
				Quantity:       d("100"),
				Price:          d("50"),
				Currency:       "USD",
				IsTradeAllowed: true,
			},
			want: policy.PostViolationFXMissing,
		},
		{
			name: "trading not allowed",
			in: policy.PostInputs{
				Type:           vo.TransactionTypeBuy,
				Portfolio:      activePortfolio(),
				Fund:           activeFund(),
				Instrument:     activeInstrument("THB", 1),
				Quantity:       d("100"),
				Price:          d("50"),
				Currency:       "THB",
				IsTradeAllowed: false,
			},
			want: policy.PostViolationTradingNotAllowed,
		},
		{
			name: "lock blocks normal post",
			in: policy.PostInputs{
				Type:                vo.TransactionTypeBuy,
				Portfolio:           activePortfolio(),
				Fund:                activeFund(),
				Instrument:          activeInstrument("THB", 1),
				Quantity:            d("100"),
				Price:               d("50"),
				Currency:            "THB",
				IsTradeAllowed:      true,
				IsTransactionLocked: true,
			},
			want: policy.PostViolationTransactionLocked,
		},
		{
			name: "force-post bypass for reversal",
			in: policy.PostInputs{
				Type:                vo.TransactionTypeReversal,
				Portfolio:           activePortfolio(),
				Fund:                activeFund(),
				Currency:            "THB",
				IsTradeAllowed:      true,
				IsTransactionLocked: true,
				AllowForcePost:      true,
			},
			want: "",
		},
		{
			name: "oversell rejected",
			in: policy.PostInputs{
				Type:           vo.TransactionTypeSell,
				Portfolio:      activePortfolio(),
				Fund:           activeFund(),
				Instrument:     activeInstrument("THB", 1),
				Quantity:       d("100"),
				Price:          d("50"),
				Currency:       "THB",
				Position:       positionOf("10"),
				IsTradeAllowed: true,
			},
			want: policy.PostViolationOversell,
		},
		{
			name: "lot size violation",
			in: policy.PostInputs{
				Type:           vo.TransactionTypeBuy,
				Portfolio:      activePortfolio(),
				Fund:           activeFund(),
				Instrument:     activeInstrument("THB", 100),
				Quantity:       d("150"),
				Price:          d("50"),
				Currency:       "THB",
				IsTradeAllowed: true,
			},
			want: policy.PostViolationLotSize,
		},
		{
			name: "currency mismatch with instrument",
			in: policy.PostInputs{
				Type:           vo.TransactionTypeBuy,
				Portfolio:      activePortfolio(),
				Fund:           activeFund(),
				Instrument:     activeInstrument("THB", 1),
				Quantity:       d("100"),
				Price:          d("50"),
				Currency:       "USD",
				IsTradeAllowed: true,
				FxRateToBase:   ptrDecimal(d("36")),
			},
			want: policy.PostViolationCurrencyMismatch,
		},
		{
			name: "negative fees rejected",
			in: policy.PostInputs{
				Type:           vo.TransactionTypeBuy,
				Portfolio:      activePortfolio(),
				Fund:           activeFund(),
				Instrument:     activeInstrument("THB", 1),
				Quantity:       d("100"),
				Price:          d("50"),
				Fees:           d("-1"),
				Currency:       "THB",
				IsTradeAllowed: true,
			},
			want: policy.PostViolationFeesNegative,
		},
		{
			name: "inactive instrument rejected",
			in: policy.PostInputs{
				Type:      vo.TransactionTypeBuy,
				Portfolio: activePortfolio(),
				Fund:      activeFund(),
				Instrument: &entity.Instrument{
					Status: vo.InstrumentStatusDelisted, IsTradable: false, Currency: "THB",
				},
				Quantity:       d("100"),
				Price:          d("50"),
				Currency:       "THB",
				IsTradeAllowed: true,
			},
			want: policy.PostViolationInstrumentNotTradable,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := policy.EvaluatePost(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func ptrDecimal(d decimal.Decimal) *decimal.Decimal { return &d }
