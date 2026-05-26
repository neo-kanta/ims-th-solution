package httptransport

import (
	"net/http"
	"strconv"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// MarketDataScreenRowDTO is the frontend-safe shape for a Market Data screen
// row. Decimal fields are emitted as strings to preserve precision. The
// frontend handles currency symbol, color, and any human formatting.
type MarketDataScreenRowDTO struct {
	SecurityID        string             `json:"security_id"`
	IMSSymbol         string             `json:"ims_symbol"`
	DisplaySymbol     string             `json:"display_symbol"`
	Name              string             `json:"name"`
	AssetType         string             `json:"asset_type"`
	Currency          string             `json:"currency,omitempty"`
	ExchangeMIC       string             `json:"exchange_mic,omitempty"`
	LastPrice         *string            `json:"last_price,omitempty"`
	ChangeAmount      *string            `json:"change_amount,omitempty"`
	ChangePercent     *string            `json:"change_percent,omitempty"`
	YieldToMaturity   *string            `json:"yield_to_maturity,omitempty"`
	Volume            *int64             `json:"volume,omitempty"`
	LastUpdateAt      *time.Time         `json:"last_update_at,omitempty"`
	FreshnessStatus   string             `json:"freshness_status"`
	DataQualityStatus string             `json:"data_quality_status"`
	ProviderBadges    []ProviderBadgeDTO `json:"provider_badges"`
	Watching          bool               `json:"watching"`
	Pinned            bool               `json:"pinned"`
	MappingStatus     string             `json:"mapping_status,omitempty"`
}

// ProviderBadgeDTO is the frontend-safe provider badge.
type ProviderBadgeDTO struct {
	ProviderCode string `json:"provider_code"`
	Label        string `json:"label"`
}

// MarketDataScreenSearchItemDTO is the search-row shape; extends the row with
// search-specific fields.
type MarketDataScreenSearchItemDTO struct {
	MarketDataScreenRowDTO
	HasActiveMapping bool   `json:"has_active_mapping"`
	ActionHint       string `json:"action_hint"`
}

// FreshnessStatusDTO and DataQualityStatusDTO are string-typed aliases the
// frontend can switch on without needing schema knowledge.
type FreshnessStatusDTO string
type DataQualityStatusDTO string

// ScreenWatchlistResponseDTO wraps the watchlist screen response.
type ScreenWatchlistResponseDTO struct {
	Items []MarketDataScreenRowDTO `json:"items"`
}

// ScreenSearchResponseDTO wraps the search screen response.
type ScreenSearchResponseDTO struct {
	Items []MarketDataScreenSearchItemDTO `json:"items"`
}

// GetScreenWatchlist handles GET /market-data/screen/watchlist.
// @Summary Market Data screen — watchlist
// @Description Canonical securities joined with their latest snapshot data. Frontend-safe DTO.
// @Tags MarketData
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Maximum rows (default 100, max 500)"
// @Success 200 {object} ScreenWatchlistResponseDTO
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /market-data/screen/watchlist [get]
func (h *Handler) GetScreenWatchlist(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	rows, err := h.service.ListScreenWatchlist(r.Context(), limit)
	if err != nil {
		httputil.InternalError(w, "failed to load watchlist")
		return
	}
	out := make([]MarketDataScreenRowDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, toScreenRowDTO(row))
	}
	httputil.OK(w, ScreenWatchlistResponseDTO{Items: out})
}

// GetScreenSearch handles GET /market-data/screen/search.
// @Summary Market Data screen — search
// @Description Search canonical securities with latest snapshot data and add-to-watchlist hint.
// @Tags MarketData
// @Security BearerAuth
// @Produce json
// @Param query query string false "Free-text query"
// @Param limit query int false "Maximum rows (default 50, max 200)"
// @Success 200 {object} ScreenSearchResponseDTO
// @Router /market-data/screen/search [get]
func (h *Handler) GetScreenSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("query")
	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	rows, err := h.service.SearchScreen(r.Context(), q, limit)
	if err != nil {
		httputil.InternalError(w, "failed to search market data screen")
		return
	}
	out := make([]MarketDataScreenSearchItemDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, MarketDataScreenSearchItemDTO{
			MarketDataScreenRowDTO: toScreenRowDTO(row.ScreenRow),
			HasActiveMapping:       row.HasActiveMapping,
			ActionHint:             row.ActionHint,
		})
	}
	httputil.OK(w, ScreenSearchResponseDTO{Items: out})
}

func toScreenRowDTO(row application.ScreenRow) MarketDataScreenRowDTO {
	dto := MarketDataScreenRowDTO{
		SecurityID:        row.SecurityID,
		IMSSymbol:         row.IMSSymbol,
		DisplaySymbol:     row.DisplaySymbol,
		Name:              row.Name,
		AssetType:         row.AssetType,
		Currency:          row.Currency,
		ExchangeMIC:       row.ExchangeMIC,
		LastUpdateAt:      row.LastUpdateAt,
		FreshnessStatus:   row.FreshnessStatus,
		DataQualityStatus: row.DataQualityStatus,
		Watching:          row.Watching,
		Pinned:            row.Pinned,
		MappingStatus:     row.MappingStatus,
		ProviderBadges:    make([]ProviderBadgeDTO, 0, len(row.ProviderBadges)),
	}
	if row.LastPrice != nil {
		s := row.LastPrice.String()
		dto.LastPrice = &s
	}
	if row.ChangeAmount != nil {
		s := row.ChangeAmount.String()
		dto.ChangeAmount = &s
	}
	if row.ChangePercent != nil {
		s := row.ChangePercent.String()
		dto.ChangePercent = &s
	}
	if row.YieldToMaturity != nil {
		s := row.YieldToMaturity.String()
		dto.YieldToMaturity = &s
	}
	if row.Volume != nil {
		v := *row.Volume
		dto.Volume = &v
	}
	for _, badge := range row.ProviderBadges {
		dto.ProviderBadges = append(dto.ProviderBadges, ProviderBadgeDTO{
			ProviderCode: badge.ProviderCode,
			Label:        badge.Label,
		})
	}
	return dto
}
