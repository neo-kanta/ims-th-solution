package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// Validation limits.
const (
	MaxIMSSymbolLength     = 128
	MaxDisplaySymbolLength = 128
	ISINLength             = 12
	CurrencyLength         = 3
	CountryCodeLength      = 2
)

// Common validation errors.
var (
	ErrInvalidIMSSymbol     = errors.New("invalid ims_symbol")
	ErrInvalidDisplaySymbol = errors.New("invalid display_symbol")
	ErrInvalidISIN          = errors.New("invalid isin")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrInvalidCountryCode   = errors.New("invalid country_code")
	ErrInvalidAssetType     = errors.New("invalid asset_type")
)

// BuildIMSSymbolInput captures the fields BuildIMSSymbol needs to deterministic-
// ally produce a canonical IMS symbol.
type BuildIMSSymbolInput struct {
	AssetType   AssetType
	CountryCode string
	ExchangeMIC string
	Ticker      string
	ISIN        string
	BaseCcy     string
	QuoteCcy    string
	Fallback    string
}

// BuildIMSSymbol produces the canonical ims_symbol for a security.
//
// Rules:
//   - Equity: COUNTRY_EQ_MIC_TICKER          (TH_EQ_XBKK_KBANK)
//   - ETF:    COUNTRY_ETF_MIC_TICKER         (US_ETF_XNYS_SPY)
//   - Bond:   COUNTRY_BOND_ISIN              (TH_BOND_TH0623053C09)
//   - FX:     FX_BASEQUOTE                   (FX_USDTHB)
//   - Otherwise / unknown: SEC_<normalized fallback>
func BuildIMSSymbol(in BuildIMSSymbolInput) (string, error) {
	switch in.AssetType {
	case AssetTypeEquity, AssetTypeETF:
		country := strings.ToUpper(strings.TrimSpace(in.CountryCode))
		mic := strings.ToUpper(strings.TrimSpace(in.ExchangeMIC))
		ticker := strings.ToUpper(strings.TrimSpace(in.Ticker))
		if country == "" || mic == "" || ticker == "" {
			return "", fmt.Errorf("%w: equity/etf requires country, exchange MIC, and ticker", ErrInvalidIMSSymbol)
		}
		marker := "EQ"
		if in.AssetType == AssetTypeETF {
			marker = "ETF"
		}
		return enforceIMSSymbol(strings.Join([]string{country, marker, mic, ticker}, "_"))
	case AssetTypeBond:
		country := strings.ToUpper(strings.TrimSpace(in.CountryCode))
		isin := strings.ToUpper(strings.TrimSpace(in.ISIN))
		if country == "" || isin == "" {
			return "", fmt.Errorf("%w: bond requires country and ISIN", ErrInvalidIMSSymbol)
		}
		if len(isin) != ISINLength {
			return "", fmt.Errorf("%w: ISIN must be 12 characters", ErrInvalidIMSSymbol)
		}
		return enforceIMSSymbol(strings.Join([]string{country, "BOND", isin}, "_"))
	case AssetTypeFX:
		base := strings.ToUpper(strings.TrimSpace(in.BaseCcy))
		quote := strings.ToUpper(strings.TrimSpace(in.QuoteCcy))
		if base == "" || quote == "" {
			return "", fmt.Errorf("%w: fx requires base and quote currency", ErrInvalidIMSSymbol)
		}
		if len(base) != CurrencyLength || len(quote) != CurrencyLength {
			return "", fmt.Errorf("%w: fx currencies must be 3 letters each", ErrInvalidIMSSymbol)
		}
		return enforceIMSSymbol("FX_" + base + quote)
	default:
		fallback := strings.TrimSpace(in.Fallback)
		if fallback == "" {
			fallback = strings.TrimSpace(in.Ticker)
		}
		if fallback == "" {
			return "", fmt.Errorf("%w: cannot derive symbol without a fallback ticker", ErrInvalidIMSSymbol)
		}
		return enforceIMSSymbol("SEC_" + normalizeFallback(fallback))
	}
}

// ValidateIMSSymbol enforces the canonical-symbol rules independent of build.
func ValidateIMSSymbol(symbol string) error {
	if symbol == "" {
		return fmt.Errorf("%w: empty", ErrInvalidIMSSymbol)
	}
	if len(symbol) > MaxIMSSymbolLength {
		return fmt.Errorf("%w: max %d chars", ErrInvalidIMSSymbol, MaxIMSSymbolLength)
	}
	upper := strings.ToUpper(symbol)
	if upper != symbol {
		return fmt.Errorf("%w: must be uppercase", ErrInvalidIMSSymbol)
	}
	for _, r := range symbol {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return fmt.Errorf("%w: whitespace/control not allowed", ErrInvalidIMSSymbol)
		}
	}
	return nil
}

// ValidateDisplaySymbol enforces the display-symbol rules.
func ValidateDisplaySymbol(symbol string) error {
	if symbol == "" {
		return fmt.Errorf("%w: empty", ErrInvalidDisplaySymbol)
	}
	if len(symbol) > MaxDisplaySymbolLength {
		return fmt.Errorf("%w: max %d chars", ErrInvalidDisplaySymbol, MaxDisplaySymbolLength)
	}
	for _, r := range symbol {
		if unicode.IsControl(r) {
			return fmt.Errorf("%w: control chars not allowed", ErrInvalidDisplaySymbol)
		}
	}
	return nil
}

// ValidateISIN enforces 12-character ISO 6166 format when not empty.
func ValidateISIN(isin string) error {
	if isin == "" {
		return nil
	}
	if len(isin) != ISINLength {
		return fmt.Errorf("%w: must be %d characters", ErrInvalidISIN, ISINLength)
	}
	for _, r := range isin {
		if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return fmt.Errorf("%w: must be uppercase alphanumeric", ErrInvalidISIN)
		}
	}
	return nil
}

// ValidateCurrency enforces 3-letter uppercase ISO 4217 when not empty.
func ValidateCurrency(currency string) error {
	if currency == "" {
		return nil
	}
	if len(currency) != CurrencyLength {
		return fmt.Errorf("%w: must be %d letters", ErrInvalidCurrency, CurrencyLength)
	}
	for _, r := range currency {
		if !(r >= 'A' && r <= 'Z') {
			return fmt.Errorf("%w: must be uppercase letters", ErrInvalidCurrency)
		}
	}
	return nil
}

// ValidateCountryCode enforces 2-letter uppercase ISO 3166-1 alpha-2 when not empty.
func ValidateCountryCode(code string) error {
	if code == "" {
		return nil
	}
	if len(code) != CountryCodeLength {
		return fmt.Errorf("%w: must be %d letters", ErrInvalidCountryCode, CountryCodeLength)
	}
	for _, r := range code {
		if !(r >= 'A' && r <= 'Z') {
			return fmt.Errorf("%w: must be uppercase letters", ErrInvalidCountryCode)
		}
	}
	return nil
}

func enforceIMSSymbol(symbol string) (string, error) {
	if err := ValidateIMSSymbol(symbol); err != nil {
		return "", err
	}
	return symbol, nil
}

func normalizeFallback(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	out := strings.Builder{}
	out.Grow(len(s))
	for _, r := range s {
		switch {
		case (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			out.WriteRune(r)
		default:
			out.WriteRune('_')
		}
	}
	// collapse repeated underscores and trim leading/trailing
	collapsed := strings.Builder{}
	prevUnderscore := false
	for _, r := range out.String() {
		if r == '_' {
			if prevUnderscore {
				continue
			}
			prevUnderscore = true
		} else {
			prevUnderscore = false
		}
		collapsed.WriteRune(r)
	}
	return strings.Trim(collapsed.String(), "_")
}
