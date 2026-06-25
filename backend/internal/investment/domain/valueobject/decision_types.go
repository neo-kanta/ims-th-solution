package valueobject

// DecisionType classifies the structural shape of an investment decision.
type DecisionType string

const (
	DecisionTypeSingleOrder DecisionType = "SINGLE_ORDER"
	DecisionTypeBasketOrder DecisionType = "BASKET_ORDER"
	DecisionTypeRebalance   DecisionType = "REBALANCE"
	DecisionTypeSwitch      DecisionType = "SWITCH"
)

// DecisionProcessType indicates the operational intent of a decision.
type DecisionProcessType string

const (
	DecisionProcessInvestment DecisionProcessType = "INVESTMENT_DECISION"
	DecisionProcessCancel     DecisionProcessType = "ORDER_CANCEL"
	DecisionProcessAmend      DecisionProcessType = "ORDER_AMEND"
)

// DecisionProductType names the primary asset class of the decision.
type DecisionProductType string

const (
	DecisionProductMutualFund DecisionProductType = "MUTUAL_FUND"
	DecisionProductETF        DecisionProductType = "ETF"
	DecisionProductStock      DecisionProductType = "STOCK"
	DecisionProductBond       DecisionProductType = "BOND"
	DecisionProductCash       DecisionProductType = "CASH"
	DecisionProductMixed      DecisionProductType = "MIXED"
)
