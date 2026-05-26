package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
)

func TestBuildIMSSymbol_Equity(t *testing.T) {
	got, err := domain.BuildIMSSymbol(domain.BuildIMSSymbolInput{
		AssetType:   domain.AssetTypeEquity,
		CountryCode: "TH",
		ExchangeMIC: "XBKK",
		Ticker:      "kbank",
	})
	require.NoError(t, err)
	require.Equal(t, "TH_EQ_XBKK_KBANK", got)
}

func TestBuildIMSSymbol_ETF(t *testing.T) {
	got, err := domain.BuildIMSSymbol(domain.BuildIMSSymbolInput{
		AssetType:   domain.AssetTypeETF,
		CountryCode: "US",
		ExchangeMIC: "XNYS",
		Ticker:      "SPY",
	})
	require.NoError(t, err)
	require.Equal(t, "US_ETF_XNYS_SPY", got)
}

func TestBuildIMSSymbol_Bond(t *testing.T) {
	got, err := domain.BuildIMSSymbol(domain.BuildIMSSymbolInput{
		AssetType:   domain.AssetTypeBond,
		CountryCode: "TH",
		ISIN:        "TH0623053C09",
	})
	require.NoError(t, err)
	require.Equal(t, "TH_BOND_TH0623053C09", got)
}

func TestBuildIMSSymbol_BondRejectsShortISIN(t *testing.T) {
	_, err := domain.BuildIMSSymbol(domain.BuildIMSSymbolInput{
		AssetType:   domain.AssetTypeBond,
		CountryCode: "TH",
		ISIN:        "TH123",
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrInvalidIMSSymbol))
}

func TestBuildIMSSymbol_FX(t *testing.T) {
	got, err := domain.BuildIMSSymbol(domain.BuildIMSSymbolInput{
		AssetType: domain.AssetTypeFX,
		BaseCcy:   "usd",
		QuoteCcy:  "thb",
	})
	require.NoError(t, err)
	require.Equal(t, "FX_USDTHB", got)
}

func TestBuildIMSSymbol_FXRejectsInvalidCurrency(t *testing.T) {
	_, err := domain.BuildIMSSymbol(domain.BuildIMSSymbolInput{
		AssetType: domain.AssetTypeFX,
		BaseCcy:   "US",
		QuoteCcy:  "THB",
	})
	require.Error(t, err)
}

func TestBuildIMSSymbol_UnknownUsesFallback(t *testing.T) {
	got, err := domain.BuildIMSSymbol(domain.BuildIMSSymbolInput{
		AssetType: domain.AssetTypeUnknown,
		Fallback:  "kbank.bk",
	})
	require.NoError(t, err)
	require.Equal(t, "SEC_KBANK_BK", got)
}

func TestBuildIMSSymbol_RejectsWhitespaceInTicker(t *testing.T) {
	_, err := domain.BuildIMSSymbol(domain.BuildIMSSymbolInput{
		AssetType:   domain.AssetTypeEquity,
		CountryCode: "TH",
		ExchangeMIC: "XBKK",
		Ticker:      "KB ANK",
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrInvalidIMSSymbol))
}

func TestValidateISIN(t *testing.T) {
	require.NoError(t, domain.ValidateISIN(""))
	require.NoError(t, domain.ValidateISIN("TH0623053C09"))
	require.ErrorIs(t, domain.ValidateISIN("TH00"), domain.ErrInvalidISIN)
	require.ErrorIs(t, domain.ValidateISIN("th0623053c09"), domain.ErrInvalidISIN)
}

func TestValidateCurrency(t *testing.T) {
	require.NoError(t, domain.ValidateCurrency(""))
	require.NoError(t, domain.ValidateCurrency("THB"))
	require.ErrorIs(t, domain.ValidateCurrency("TH"), domain.ErrInvalidCurrency)
	require.ErrorIs(t, domain.ValidateCurrency("thb"), domain.ErrInvalidCurrency)
	require.ErrorIs(t, domain.ValidateCurrency("THBX"), domain.ErrInvalidCurrency)
}

func TestValidateCountryCode(t *testing.T) {
	require.NoError(t, domain.ValidateCountryCode(""))
	require.NoError(t, domain.ValidateCountryCode("TH"))
	require.ErrorIs(t, domain.ValidateCountryCode("THA"), domain.ErrInvalidCountryCode)
	require.ErrorIs(t, domain.ValidateCountryCode("th"), domain.ErrInvalidCountryCode)
}

func TestAssetType_IsValidIncludesBondAndFX(t *testing.T) {
	require.True(t, domain.AssetTypeBond.IsValid())
	require.True(t, domain.AssetTypeFX.IsValid())
	require.True(t, domain.AssetTypeIndex.IsValid())
	require.False(t, domain.AssetType("BOGUS").IsValid())
}
