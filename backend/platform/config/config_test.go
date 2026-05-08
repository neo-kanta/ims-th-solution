package config

import (
	"testing"
	"time"
)

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
