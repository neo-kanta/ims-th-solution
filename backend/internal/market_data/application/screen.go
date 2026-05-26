package application

import (
	"context"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
	refdomain "github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
)

// Freshness buckets for screen rows. The frontend uses them to render badges;
// only the enum string crosses the wire — formatting is the frontend's job.
const (
	FreshnessFresh   = "FRESH"
	FreshnessStale   = "STALE"
	FreshnessExpired = "EXPIRED"
	FreshnessMissing = "MISSING"
)

// Data quality buckets surfaced on screen rows.
const (
	DataQualityOK      = "OK"
	DataQualityWarning = "WARNING"
	DataQualityMissing = "MISSING"
)

// Provider labels for badges. Unknown providers fall back to the upper-cased code.
var providerLabels = map[string]string{
	domain.ProviderAlphaVantage: "Alpha",
	domain.ProviderYahoo:        "Yahoo",
}

// ScreenRow is the canonical row consumed by /market-data/screen/* endpoints.
// Decimal values are kept as decimal.Decimal here and converted to strings at
// the DTO boundary. No currency symbols, colors, or CSS classes — those belong
// to the frontend.
type ScreenRow struct {
	SecurityID        string
	IMSSymbol         string
	DisplaySymbol     string
	Name              string
	AssetType         string
	Currency          string
	ExchangeMIC       string
	LastPrice         *decimal.Decimal
	ChangeAmount      *decimal.Decimal
	ChangePercent     *decimal.Decimal
	YieldToMaturity   *decimal.Decimal
	Volume            *int64
	LastUpdateAt      *time.Time
	FreshnessStatus   string
	DataQualityStatus string
	ProviderBadges    []ScreenProviderBadge
	Watching          bool
	Pinned            bool
	MappingStatus     string
}

// ScreenProviderBadge identifies an active provider mapping for the row.
type ScreenProviderBadge struct {
	ProviderCode string
	Label        string
}

// ScreenSearchRow is the lighter row returned by /market-data/screen/search.
type ScreenSearchRow struct {
	ScreenRow
	HasActiveMapping bool
	ActionHint       string // e.g. "ADD_TO_WATCHLIST", "MAP_SYMBOL"
}

// ListScreenWatchlist returns active canonical securities with their latest
// snapshot data, suitable for the market data screen.
func (s *Service) ListScreenWatchlist(ctx context.Context, limit int) ([]ScreenRow, error) {
	if s.securityResolver == nil {
		return []ScreenRow{}, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	securities, err := s.securityResolver.SearchSecurities(ctx, refdomain.SecuritySearchFilter{
		Status: refdomain.SecurityStatusActive,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]ScreenRow, 0, len(securities))
	for _, sec := range securities {
		out = append(out, s.buildScreenRow(ctx, sec))
	}
	return out, nil
}

// SearchScreen returns canonical securities matching the query, joined with
// the latest snapshot data and a mapping-state action hint.
func (s *Service) SearchScreen(ctx context.Context, query string, limit int) ([]ScreenSearchRow, error) {
	if s.securityResolver == nil {
		return []ScreenSearchRow{}, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	securities, err := s.securityResolver.SearchSecurities(ctx, refdomain.SecuritySearchFilter{
		Query: strings.TrimSpace(query),
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]ScreenSearchRow, 0, len(securities))
	for _, sec := range securities {
		row := s.buildScreenRow(ctx, sec)
		hasActive := false
		for _, m := range sec.ProviderMappings {
			if m.MappingStatus == refdomain.MappingStatusActive {
				hasActive = true
				break
			}
		}
		hint := "ADD_TO_WATCHLIST"
		if !hasActive {
			hint = "MAP_SYMBOL"
		}
		out = append(out, ScreenSearchRow{
			ScreenRow:        row,
			HasActiveMapping: hasActive,
			ActionHint:       hint,
		})
	}
	return out, nil
}

func (s *Service) buildScreenRow(ctx context.Context, sec refdomain.Security) ScreenRow {
	row := ScreenRow{
		SecurityID:    sec.ID,
		IMSSymbol:     sec.IMSSymbol,
		DisplaySymbol: sec.DisplaySymbol,
		Name:          sec.Name,
		AssetType:     string(sec.AssetType),
		Currency:      sec.Currency,
		ExchangeMIC:   sec.ExchangeMIC,
		MappingStatus: mappingStatusForSecurity(sec),
	}

	// Best-effort latest snapshot lookup via the existing repository. Looking
	// up by display_symbol covers the legacy snapshots path; canonical
	// snapshot↔security_id linkage will follow once snapshot writes adopt it.
	if s.repo != nil {
		if q, err := s.repo.GetLatestQuote(ctx, sec.DisplaySymbol); err == nil && q != nil {
			price := q.Price
			row.LastPrice = &price
			if !q.Change.IsZero() {
				ch := q.Change
				row.ChangeAmount = &ch
			}
			if !q.ChangePercent.IsZero() {
				cp := q.ChangePercent
				row.ChangePercent = &cp
			}
			if q.Volume > 0 {
				vol := q.Volume
				row.Volume = &vol
			}
			ts := q.AsOf
			if ts.IsZero() {
				ts = q.CapturedAt
			}
			if !ts.IsZero() {
				row.LastUpdateAt = &ts
			}
			row.FreshnessStatus = freshnessFor(ts, s.now())
			row.DataQualityStatus = DataQualityOK
			if row.FreshnessStatus == FreshnessStale {
				row.DataQualityStatus = DataQualityWarning
			}
		}
	}

	if row.FreshnessStatus == "" {
		row.FreshnessStatus = FreshnessMissing
	}
	if row.DataQualityStatus == "" {
		row.DataQualityStatus = DataQualityMissing
	}

	// Provider badges from canonical mappings (active only).
	for _, m := range sec.ProviderMappings {
		if m.MappingStatus != refdomain.MappingStatusActive {
			continue
		}
		row.ProviderBadges = append(row.ProviderBadges, ScreenProviderBadge{
			ProviderCode: m.ProviderCode,
			Label:        providerLabelFor(m.ProviderCode),
		})
	}
	return row
}

func mappingStatusForSecurity(sec refdomain.Security) string {
	for _, m := range sec.ProviderMappings {
		if m.MappingStatus == refdomain.MappingStatusActive {
			return string(refdomain.MappingStatusActive)
		}
	}
	if len(sec.ProviderMappings) == 0 {
		return string(refdomain.MappingStatusUnmapped)
	}
	return string(refdomain.MappingStatusInactive)
}

func freshnessFor(ts, now time.Time) string {
	if ts.IsZero() {
		return FreshnessMissing
	}
	age := now.Sub(ts)
	switch {
	case age < 15*time.Minute:
		return FreshnessFresh
	case age < 24*time.Hour:
		return FreshnessStale
	default:
		return FreshnessExpired
	}
}

func providerLabelFor(code string) string {
	if l, ok := providerLabels[strings.ToLower(code)]; ok {
		return l
	}
	return strings.ToUpper(code)
}
