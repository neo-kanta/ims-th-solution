package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadReportingCurrency_ValidProductionConfiguration(t *testing.T) {
	setProductionConfigEnv(t)
	t.Setenv("REPORTING_CURRENCY", "USD")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ReportingCurrency != "USD" {
		t.Fatalf("ReportingCurrency = %q, want USD", cfg.ReportingCurrency)
	}
}

func TestLoadReportingCurrency_MissingFailsProductionStartup(t *testing.T) {
	setProductionConfigEnv(t)
	t.Setenv("REPORTING_CURRENCY", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "REPORTING_CURRENCY is required") {
		t.Fatalf("Load() error = %v, want missing REPORTING_CURRENCY error", err)
	}
}

func TestLoadReportingCurrency_InvalidFailsProductionStartup(t *testing.T) {
	for _, code := range []string{"thb", "ZZZ", "US"} {
		t.Run(code, func(t *testing.T) {
			setProductionConfigEnv(t)
			t.Setenv("REPORTING_CURRENCY", code)

			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), "REPORTING_CURRENCY") {
				t.Fatalf("Load() error = %v, want invalid REPORTING_CURRENCY error", err)
			}
		})
	}
}

func setProductionConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_JWT_SECRET", "production-test-jwt-secret")
	t.Setenv("MFA_ENCRYPTION_KEY", strings.Repeat("0", 64))
	t.Setenv("ALPHA_VANTAGE_API_KEY", "production-test-market-data-key")
}

func TestParseStaleThresholds_Defaults(t *testing.T) {
	t.Parallel()
	got := parseStaleThresholds("DOES_NOT_EXIST", "EQUITY=24h,FIXED_INCOME=72h,FUND=24h")
	cases := map[string]time.Duration{
		"EQUITY":       24 * time.Hour,
		"FIXED_INCOME": 72 * time.Hour,
		"FUND":         24 * time.Hour,
	}
	for code, want := range cases {
		if got[code] != want {
			t.Errorf("threshold for %s = %v, want %v", code, got[code], want)
		}
	}
}

func TestParseStaleThresholds_IgnoresMalformed(t *testing.T) {
	t.Parallel()
	got := parseStaleThresholds("DOES_NOT_EXIST", "EQUITY=24h,GARBAGE,=10s,KEY=,KEY2=notaduration")
	if got["EQUITY"] != 24*time.Hour {
		t.Errorf("valid pair lost: %+v", got)
	}
	if len(got) != 1 {
		t.Errorf("malformed entries should be ignored; got %+v", got)
	}
}

func TestParseStaleThresholds_OverrideViaEnv(t *testing.T) {
	t.Setenv("VALUATION_STALE_THRESHOLDS", "EQUITY=1h")
	got := parseStaleThresholds("VALUATION_STALE_THRESHOLDS", "EQUITY=24h")
	if got["EQUITY"] != time.Hour {
		t.Errorf("env override ignored; got %+v", got)
	}
}
