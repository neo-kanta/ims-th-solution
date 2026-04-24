package credit

import (
	"strings"
)

// ratingOrdinal is the numeric scale we compare against. Higher values are
// better credit quality. Value 0 is reserved as "invalid / unrecognised" so
// that a zero-valued struct cannot accidentally satisfy a `>= min` check.
//
// The scale is unified across TRIS / Fitch Thailand / S&P / Moody's by mapping
// each agency's notation to the same ordinal slot — callers do not need to
// know which agency issued the rating in order to compare.
type ratingOrdinal struct {
	Code  string // canonical S&P/TRIS form ("AAA", "BBB-", …)
	Value int    // higher = better
	Valid bool
}

// Canonical 22-level ordinal map. The "Code" field is the S&P/TRIS form; we
// map Moody's notation onto the same row (Aaa↔AAA, Baa3↔BBB-, etc.).
var ratingByCode = map[string]int{
	"AAA": 22,
	"AA+": 21, "AA": 20, "AA-": 19,
	"A+": 18, "A": 17, "A-": 16,
	"BBB+": 15, "BBB": 14, "BBB-": 13,
	"BB+": 12, "BB": 11, "BB-": 10,
	"B+": 9, "B": 8, "B-": 7,
	"CCC+": 6, "CCC": 5, "CCC-": 4,
	"CC": 3,
	"C":  2,
	"D":  1,
}

// moodysAliases maps Moody's-style codes onto the canonical S&P code.
var moodysAliases = map[string]string{
	"AAA": "AAA",
	"AA1": "AA+", "AA2": "AA", "AA3": "AA-",
	"A1": "A+", "A2": "A", "A3": "A-",
	"BAA1": "BBB+", "BAA2": "BBB", "BAA3": "BBB-",
	"BA1": "BB+", "BA2": "BB", "BA3": "BB-",
	"B1": "B+", "B2": "B", "B3": "B-",
	"CAA1": "CCC+", "CAA2": "CCC", "CAA3": "CCC-",
	"CA": "CC",
	"C":  "C",
}

// ParseRating maps a free-form rating string to the canonical ordinal.
//
// Accepted forms:
//   - S&P / Fitch / TRIS: "AAA", "AA-", "BBB+", "B", "D" (case-insensitive)
//   - Moody's:            "Aaa", "Aa2", "Baa3", "Ba1", "B2", "Caa1", "Ca", "C"
//
// Trailing qualifiers (".pd", "u", "*", "(stable)") are tolerated by trimming
// everything after the first space or non-rating character run.
func ParseRating(raw string) (ratingOrdinal, bool) {
	clean := canonicaliseRatingCode(raw)
	if clean == "" {
		return ratingOrdinal{}, false
	}

	// Moody's alias first (so "A3" resolves to "A-" not literal "A3").
	if mapped, ok := moodysAliases[clean]; ok {
		clean = mapped
	}

	if v, ok := ratingByCode[clean]; ok {
		return ratingOrdinal{Code: clean, Value: v, Valid: true}, true
	}
	return ratingOrdinal{Code: clean}, false
}

// canonicaliseRatingCode strips whitespace / qualifiers and upper-cases.
func canonicaliseRatingCode(raw string) string {
	s := strings.ToUpper(strings.TrimSpace(raw))
	if s == "" {
		return ""
	}

	if i := strings.IndexByte(s, ' '); i > 0 {
		s = s[:i]
	}

	s = strings.TrimSuffix(s, "U")
	s = strings.TrimSuffix(s, "*")
	s = strings.TrimSuffix(s, ".PD")
	return s
}
