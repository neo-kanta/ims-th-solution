package valueobject

// FundStatus represents the lifecycle state of a Fund (contract).
type FundStatus string

const (
	FundStatusActive    FundStatus = "ACTIVE"
	FundStatusSuspended FundStatus = "SUSPENDED"
	FundStatusClosed    FundStatus = "CLOSED"
)

// IsValid checks whether s is one of the recognised fund statuses.
func (s FundStatus) IsValid() bool {
	switch s {
	case FundStatusActive, FundStatusSuspended, FundStatusClosed:
		return true
	}
	return false
}

// PortfolioType classifies whether a Portfolio is real, paper, or a
// target-allocation template (docs/api/portfolio-v2-api-ddd.md section 4).
type PortfolioType string

const (
	PortfolioTypeLive       PortfolioType = "LIVE"
	PortfolioTypeSimulation PortfolioType = "SIMULATION"
	PortfolioTypeModel      PortfolioType = "MODEL"
)

func (t PortfolioType) IsValid() bool {
	switch t {
	case PortfolioTypeLive, PortfolioTypeSimulation, PortfolioTypeModel:
		return true
	}
	return false
}

// PortfolioStatus represents the lifecycle state of a Portfolio.
type PortfolioStatus string

const (
	PortfolioStatusDraft           PortfolioStatus = "DRAFT"
	PortfolioStatusPendingApproval PortfolioStatus = "PENDING_APPROVAL"
	PortfolioStatusActive          PortfolioStatus = "ACTIVE"
	PortfolioStatusPaused          PortfolioStatus = "PAUSED" // Phase 1 compat alias for SUSPENDED
	PortfolioStatusSuspended       PortfolioStatus = "SUSPENDED"
	PortfolioStatusRejected        PortfolioStatus = "REJECTED"
	PortfolioStatusClosed          PortfolioStatus = "CLOSED"
)

func (s PortfolioStatus) IsValid() bool {
	switch s {
	case PortfolioStatusDraft, PortfolioStatusPendingApproval,
		PortfolioStatusActive, PortfolioStatusPaused, PortfolioStatusSuspended,
		PortfolioStatusRejected, PortfolioStatusClosed:
		return true
	}
	return false
}

// IsOpenForBusiness reports whether the portfolio can receive new transactions.
// Investment activity starts only after approval completes and status is ACTIVE.
func (s PortfolioStatus) IsOpenForBusiness() bool {
	return s == PortfolioStatusActive
}

// InstrumentStatus represents the tradability state of an instrument.
type InstrumentStatus string

const (
	InstrumentStatusActive    InstrumentStatus = "ACTIVE"
	InstrumentStatusSuspended InstrumentStatus = "SUSPENDED"
	InstrumentStatusDelisted  InstrumentStatus = "DELISTED"
)

func (s InstrumentStatus) IsValid() bool {
	switch s {
	case InstrumentStatusActive, InstrumentStatusSuspended, InstrumentStatusDelisted:
		return true
	}
	return false
}

// TransactionStatus is the post/reverse state of a ledger row.
type TransactionStatus string

const (
	TransactionStatusPosted   TransactionStatus = "POSTED"
	TransactionStatusReversed TransactionStatus = "REVERSED"
)

// RiskProfile is an optional categorical risk indicator.
type RiskProfile string

const (
	RiskProfileLow         RiskProfile = "LOW"
	RiskProfileMedium      RiskProfile = "MEDIUM"
	RiskProfileHigh        RiskProfile = "HIGH"
	RiskProfileSpeculative RiskProfile = "SPECULATIVE"
)

func (r RiskProfile) IsValid() bool {
	switch r {
	case RiskProfileLow, RiskProfileMedium, RiskProfileHigh, RiskProfileSpeculative:
		return true
	}
	return false
}

// TaxLotMethod is currently a placeholder. Only AVERAGE is implemented.
type TaxLotMethod string

const (
	TaxLotMethodAverage TaxLotMethod = "AVERAGE"
	TaxLotMethodFIFO    TaxLotMethod = "FIFO"
	TaxLotMethodLIFO    TaxLotMethod = "LIFO"
	TaxLotMethodSpecID  TaxLotMethod = "SPEC_ID"
)

func (m TaxLotMethod) IsValid() bool {
	switch m {
	case TaxLotMethodAverage, TaxLotMethodFIFO, TaxLotMethodLIFO, TaxLotMethodSpecID:
		return true
	}
	return false
}

// AumScopeType identifies whether an AUM snapshot is at portfolio or fund scope.
type AumScopeType string

const (
	AumScopePortfolio AumScopeType = "PORTFOLIO"
	AumScopeFund      AumScopeType = "FUND"
)

// ValuationSource identifies the origin of a valuation/AUM snapshot.
type ValuationSource string

const (
	ValuationSourceInternal ValuationSource = "INTERNAL"
	ValuationSourcePAM      ValuationSource = "PAM"
	ValuationSourceExternal ValuationSource = "EXTERNAL"
)
