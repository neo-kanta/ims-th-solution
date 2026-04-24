package spi

// DataDependencies declares what data a rule needs.
// The engine unions these from all applicable rules and fetches once.
type DataDependencies struct {
	Positions       bool `json:"positions"`
	NAV             bool `json:"nav"`
	MarketPrices    bool `json:"market_prices"`
	FXRates         bool `json:"fx_rates"`
	Classifications bool `json:"classifications"`
	CreditRatings   bool `json:"credit_ratings"`
	Restrictions    bool `json:"restrictions"`
	TradeHistory    bool `json:"trade_history"`
	Calendar        bool `json:"calendar"`
	PortfolioMeta   bool `json:"portfolio_meta"`

	// TradeHistoryLookbackDays is relevant only if TradeHistory is true.
	// The engine uses the maximum across all rules.
	TradeHistoryLookbackDays int `json:"trade_history_lookback_days,omitempty"`
}

// Union merges two dependency manifests (logical OR of each field).
func (d DataDependencies) Union(other DataDependencies) DataDependencies {
	result := DataDependencies{
		Positions:       d.Positions || other.Positions,
		NAV:             d.NAV || other.NAV,
		MarketPrices:    d.MarketPrices || other.MarketPrices,
		FXRates:         d.FXRates || other.FXRates,
		Classifications: d.Classifications || other.Classifications,
		CreditRatings:   d.CreditRatings || other.CreditRatings,
		Restrictions:    d.Restrictions || other.Restrictions,
		TradeHistory:    d.TradeHistory || other.TradeHistory,
		Calendar:        d.Calendar || other.Calendar,
		PortfolioMeta:   d.PortfolioMeta || other.PortfolioMeta,
	}
	if other.TradeHistoryLookbackDays > d.TradeHistoryLookbackDays {
		result.TradeHistoryLookbackDays = other.TradeHistoryLookbackDays
	} else {
		result.TradeHistoryLookbackDays = d.TradeHistoryLookbackDays
	}
	return result
}
