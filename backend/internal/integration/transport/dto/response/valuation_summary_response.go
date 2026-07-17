package response

import (
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// ValuationSummaryDTO is the API response for
// GET /integration/dashboard/valuation-summary.
//
// Decimal amounts are returned as strings per project convention. When
// DataAvailable is false the numeric fields are omitted; the frontend MUST
// render an explicit "not available" state, never a zero.
type ValuationSummaryDTO struct {
	Scope           string                      `json:"scope"`
	Username        *string                     `json:"username"`
	Status          string                      `json:"status" enums:"AVAILABLE,NO_DATA,INCOMPLETE"`
	BusinessDate    string                      `json:"business_date,omitempty"`
	Currency        string                      `json:"currency,omitempty"`
	AUMToday        string                      `json:"aum_today,omitempty"`
	TodayPnL        string                      `json:"today_pnl,omitempty"`
	TodayPnLPercent *string                     `json:"today_pnl_percent,omitempty"`
	AsOf            string                      `json:"as_of,omitempty"`
	DataAvailable   bool                        `json:"data_available"`
	Coverage        ValuationSummaryCoverageDTO `json:"coverage"`
}

// ValuationSummaryCoverageDTO explains exactly how much of the authorized
// scope contributed to an official aggregate. Exclusions use business codes,
// never internal UUIDs.
type ValuationSummaryCoverageDTO struct {
	TotalFundCount         int                            `json:"total_fund_count"`
	IncludedFundCount      int                            `json:"included_fund_count"`
	ExcludedFundCount      int                            `json:"excluded_fund_count"`
	TotalPortfolioCount    int                            `json:"total_portfolio_count"`
	IncludedPortfolioCount int                            `json:"included_portfolio_count"`
	ExcludedPortfolioCount int                            `json:"excluded_portfolio_count"`
	ExcludedCurrencies     []string                       `json:"excluded_currencies"`
	ExcludedBusinessDates  []string                       `json:"excluded_business_dates"`
	ExclusionReasons       []string                       `json:"exclusion_reasons"`
	Exclusions             []ValuationSummaryExclusionDTO `json:"exclusions"`
}

// ValuationSummaryExclusionDTO identifies an omitted official input and the
// stable machine-readable reason. Date fields are YYYY-MM-DD when present.
type ValuationSummaryExclusionDTO struct {
	FundCode             string `json:"fund_code,omitempty"`
	PortfolioCode        string `json:"portfolio_code,omitempty"`
	Currency             string `json:"currency,omitempty"`
	BusinessDate         string `json:"business_date,omitempty"`
	RequiredBusinessDate string `json:"required_business_date,omitempty"`
	Reason               string `json:"reason"`
}

// FromValuationSummary converts the domain read model to its DTO.
func FromValuationSummary(s *domain.ValuationSummary) ValuationSummaryDTO {
	dto := ValuationSummaryDTO{
		Scope:         string(s.Scope),
		Status:        string(s.Status),
		DataAvailable: s.DataAvailable,
		Coverage:      fromValuationSummaryCoverage(s.Coverage),
	}
	if s.Username != "" {
		dto.Username = &s.Username
	}
	if !s.BusinessDate.IsZero() {
		dto.BusinessDate = s.BusinessDate.Format("2006-01-02")
	}
	dto.Currency = s.Currency
	if !s.DataAvailable {
		return dto
	}
	dto.AUMToday = s.AUM.String()
	dto.TodayPnL = s.TodayPnL.String()
	if s.TodayPnLPercent != nil {
		pct := s.TodayPnLPercent.String()
		dto.TodayPnLPercent = &pct
	}
	if !s.AsOf.IsZero() {
		dto.AsOf = s.AsOf.Format(time.RFC3339)
	}
	return dto
}

func fromValuationSummaryCoverage(c contract.ValuationSummaryCoverage) ValuationSummaryCoverageDTO {
	dto := ValuationSummaryCoverageDTO{
		TotalFundCount:         c.TotalFundCount,
		IncludedFundCount:      c.IncludedFundCount,
		ExcludedFundCount:      c.ExcludedFundCount,
		TotalPortfolioCount:    c.TotalPortfolioCount,
		IncludedPortfolioCount: c.IncludedPortfolioCount,
		ExcludedPortfolioCount: c.ExcludedPortfolioCount,
		ExcludedCurrencies:     append([]string{}, c.ExcludedCurrencies...),
		ExcludedBusinessDates:  make([]string, 0, len(c.ExcludedBusinessDates)),
		ExclusionReasons:       make([]string, 0, len(c.ExclusionReasons)),
		Exclusions:             make([]ValuationSummaryExclusionDTO, 0, len(c.Exclusions)),
	}
	for _, businessDate := range c.ExcludedBusinessDates {
		if !businessDate.IsZero() {
			dto.ExcludedBusinessDates = append(dto.ExcludedBusinessDates, businessDate.Format("2006-01-02"))
		}
	}
	for _, reason := range c.ExclusionReasons {
		dto.ExclusionReasons = append(dto.ExclusionReasons, string(reason))
	}
	for _, exclusion := range c.Exclusions {
		item := ValuationSummaryExclusionDTO{
			FundCode:      exclusion.FundCode,
			PortfolioCode: exclusion.PortfolioCode,
			Currency:      exclusion.Currency,
			Reason:        string(exclusion.Reason),
		}
		if !exclusion.BusinessDate.IsZero() {
			item.BusinessDate = exclusion.BusinessDate.Format("2006-01-02")
		}
		if !exclusion.RequiredBusinessDate.IsZero() {
			item.RequiredBusinessDate = exclusion.RequiredBusinessDate.Format("2006-01-02")
		}
		dto.Exclusions = append(dto.Exclusions, item)
	}
	return dto
}
