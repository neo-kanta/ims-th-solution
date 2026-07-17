package response

import (
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/domain"
)

// ValuationSummaryDTO is the API response for
// GET /integration/dashboard/valuation-summary.
//
// Decimal amounts are returned as strings per project convention. When
// DataAvailable is false the numeric fields are empty strings — the
// frontend MUST render an explicit "not available" state, never a zero.
type ValuationSummaryDTO struct {
	Scope           string  `json:"scope"`
	Username        *string `json:"username"`
	BusinessDate    string  `json:"business_date,omitempty"`
	Currency        string  `json:"currency,omitempty"`
	AUMToday        string  `json:"aum_today,omitempty"`
	TodayPnL        string  `json:"today_pnl,omitempty"`
	TodayPnLPercent *string `json:"today_pnl_percent,omitempty"`
	AsOf            string  `json:"as_of,omitempty"`
	DataAvailable   bool    `json:"data_available"`
}

// FromValuationSummary converts the domain read model to its DTO.
func FromValuationSummary(s *domain.ValuationSummary) ValuationSummaryDTO {
	dto := ValuationSummaryDTO{
		Scope:         string(s.Scope),
		DataAvailable: s.DataAvailable,
	}
	if s.Username != "" {
		dto.Username = &s.Username
	}
	if !s.DataAvailable {
		return dto
	}
	dto.BusinessDate = s.BusinessDate.Format("2006-01-02")
	dto.Currency = s.Currency
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
