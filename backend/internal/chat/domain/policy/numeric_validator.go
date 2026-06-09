package policy

// numeric_validator.go is the post-generation numeric financial validator —
// the structural enforcement behind the module's core rule: every financial
// figure in a final answer must trace to a current-turn MCP tool result (a raw
// data tool OR a deterministic calculation tool). Untraceable figures make the
// answer unverified.
//
// It is a PURE function with no infrastructure dependencies, so it is fully
// unit-testable. The agent loop calls ValidateAnswer after the model finishes
// and BEFORE the answer is streamed/persisted as completed.
//
// Design choices (deliberately conservative, but not stupid):
//   - A number is treated as a financial figure ONLY when it is a percentage,
//     currency-adjacent, a decimal, or thousands-separated. Bare integers are
//     treated as harmless counts ("3 steps", "1 portfolio", "top 5 holdings")
//     and are NOT validated — this avoids false blocks while still catching all
//     realistic NAV/price/money/percentage figures, which carry decimals.
//   - Dates are validated ONLY in a financial-date context (NAV/valuation/
//     trade/transaction/settlement/as-of/business date).
//   - Matching is exact on rational value (so "10.2500" matches "10.25") and on
//     the YYYY-MM-DD date prefix.
//   - On any violation the answer is BLOCKED with a safe fallback message. We
//     prefer blocking over inline redaction because automatic redaction cannot
//     guarantee the remaining answer stays useful and not misleading.

import (
	"encoding/json"
	"math/big"
	"regexp"
	"strings"
	"unicode"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// ToolResultSource is one current-turn tool result the validator may trace
// figures to. Raw is the verbatim tool output (JSON text). ToolName drives the
// source classification (calc_* tools are deterministic calculation sources).
type ToolResultSource struct {
	ToolName string
	Raw      string
}

// SafeBlockedMessage is shown when one or more financial figures could not be
// verified. It never reveals which value failed or any internal detail.
const SafeBlockedMessage = "I can't show this answer because one or more financial figures in it " +
	"could not be verified against the IMS system of record. I only report values returned by IMS " +
	"tools. Please try again or narrow your question (for example, name the specific fund or portfolio)."

var (
	// number token: optional sign, digits with optional thousands groups, optional fraction.
	reNumber = regexp.MustCompile(`-?\d{1,3}(?:,\d{3})+(?:\.\d+)?|-?\d+(?:\.\d+)?`)
	// ISO date prefix (date or datetime).
	reISODate = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	// currency codes / symbols used to mark a number as money.
	reCurrency = regexp.MustCompile(`(?i)\b(THB|USD|EUR|JPY|GBP|SGD|CNY|HKD|AUD|CHF|baht)\b|[$฿€£]`)
)

// financial keywords that, near a date, make it a financial-context date.
var dateKeywords = []string{
	"nav date", "valuation date", "trade date", "transaction date",
	"settlement date", "as of", "as-of", "asof", "business date", "dated",
	"value date", "pricing date", "price date",
}

// ValidateAnswer evaluates answer against the current-turn tool results and
// returns a ValidationResult. If answer contains no financial figures the
// result is always an allow (Checked == 0).
func ValidateAnswer(answer string, sources []ToolResultSource) valueobject.ValidationResult {
	numCorpus, dateCorpus := buildCorpus(sources)

	figures := detectFigures(answer)
	res := valueobject.ValidationResult{
		Decision:  valueobject.DecisionAllow,
		FinalText: answer,
		Checked:   len(figures),
	}

	for _, f := range figures {
		var meta sourceMeta
		var found bool
		if f.IsDate() {
			meta, found = dateCorpus[f.Canonical]
		} else {
			meta, found = resolveFigure(f, answer, numCorpus)
		}
		if !found {
			res.Violations = append(res.Violations, valueobject.ValidationViolation{
				Figure: f.Raw,
				Kind:   f.Kind,
				Reason: "value not found in any current-turn tool result",
			})
			continue
		}
		res.Bindings = append(res.Bindings, valueobject.ProvenanceBinding{
			Figure:   f.Raw,
			Source:   meta.source,
			ToolName: meta.toolName,
		})
	}

	if len(res.Violations) > 0 {
		res.Decision = valueobject.DecisionBlock
		res.FinalText = SafeBlockedMessage
	}
	return res
}

// resolveFigure resolves a financial figure against the corpus with optional scaling suffixes and rounding tolerance.
func resolveFigure(f valueobject.FinancialFigure, answer string, numCorpus map[string]sourceMeta) (sourceMeta, bool) {
	// 1. Check exact match first
	if meta, found := numCorpus[f.Canonical]; found {
		return meta, true
	}

	// 2. Try parsing the raw string as a big.Rat
	clean := strings.ReplaceAll(strings.TrimSpace(f.Raw), ",", "")
	claimedVal := new(big.Rat)
	if _, ok := claimedVal.SetString(clean); !ok {
		return sourceMeta{}, false
	}

	// 3. Find multiplier
	multiplier := big.NewRat(1, 1)

	// Check next word for scale suffix
	lowerAnswer := strings.ToLower(answer)
	nextW := nextWord(lowerAnswer, f.End)
	if mult, ok := getMultiplier(nextW); ok {
		multiplier.Mul(multiplier, mult)
	}

	// If it is a percentage, apply a 1/100 factor
	isPercentage := f.Kind == valueobject.FigurePercentage || isPercent(answer, f.End)
	if isPercentage {
		multiplier.Mul(multiplier, big.NewRat(1, 100))
	}

	// Calculate decimal precision of the raw number
	precision := getPrecision(f.Raw)

	// Tolerance = 0.5 * multiplier / 10^precision
	tenPowerPrecision := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision)), nil)
	divisor := new(big.Rat).SetInt(tenPowerPrecision)

	tolerance := new(big.Rat).Mul(big.NewRat(1, 2), multiplier)
	tolerance.Quo(tolerance, divisor)

	// Multiply claimedVal by multiplier to scale it to the corpus unit
	scaledClaimed := new(big.Rat).Mul(claimedVal, multiplier)

	// Search corpus with tolerance
	if meta, found := findInCorpusWithTolerance(numCorpus, scaledClaimed, tolerance); found {
		return meta, true
	}

	return sourceMeta{}, false
}

// getPrecision returns the number of digits after the decimal point.
func getPrecision(raw string) int {
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return 0
	}
	count := 0
	for _, r := range parts[1] {
		if unicode.IsDigit(r) {
			count++
		}
	}
	return count
}

// findInCorpusWithTolerance checks if any value in numCorpus is within tolerance of claimed.
func findInCorpusWithTolerance(numCorpus map[string]sourceMeta, claimed *big.Rat, tolerance *big.Rat) (sourceMeta, bool) {
	lower := new(big.Rat).Sub(claimed, tolerance)
	upper := new(big.Rat).Add(claimed, tolerance)

	for k, meta := range numCorpus {
		val := new(big.Rat)
		if _, ok := val.SetString(k); ok {
			if val.Cmp(lower) >= 0 && val.Cmp(upper) <= 0 {
				return meta, true
			}
		}
	}
	return sourceMeta{}, false
}

// getMultiplier resolves scale suffixes to their numeric multipliers.
func getMultiplier(word string) (*big.Rat, bool) {
	w := strings.ToLower(word)

	// English exact matches first (including abbreviations)
	switch w {
	case "billion", "b":
		return big.NewRat(1000000000, 1), true
	case "million", "m":
		return big.NewRat(1000000, 1), true
	case "thousand", "k":
		return big.NewRat(1000, 1), true
	case "trillion", "t":
		return big.NewRat(1000000000000, 1), true
	}

	// Thai prefix matches (ordered descending by length to match longest prefix first)
	thaiPrefixes := []struct {
		prefix string
		mult   *big.Rat
	}{
		{"ล้านล้าน", big.NewRat(1000000000000, 1)},
		{"แสนล้าน", big.NewRat(100000000000, 1)},
		{"หมื่นล้าน", big.NewRat(10000000000, 1)},
		{"พันล้าน", big.NewRat(1000000000, 1)},
		{"ร้อยล้าน", big.NewRat(100000000, 1)},
		{"สิบล้าน", big.NewRat(10000000, 1)},
		{"ล้าน", big.NewRat(1000000, 1)},
		{"แสน", big.NewRat(100000, 1)},
		{"หมื่น", big.NewRat(10000, 1)},
		{"พัน", big.NewRat(1000, 1)},
	}

	for _, tp := range thaiPrefixes {
		if strings.HasPrefix(w, tp.prefix) {
			return tp.mult, true
		}
	}

	return nil, false
}

// sourceMeta records which tool (and source kind) first supplied a value.
type sourceMeta struct {
	toolName string
	source   valueobject.FigureSource
}

// buildCorpus extracts the set of numeric and date values present anywhere in
// the verbatim tool results. Numbers come from JSON leaves (numbers and
// fully-numeric strings) so we never pull digit runs out of UUIDs/hashes.
func buildCorpus(sources []ToolResultSource) (map[string]sourceMeta, map[string]sourceMeta) {
	nums := map[string]sourceMeta{}
	dates := map[string]sourceMeta{}

	for _, s := range sources {
		src := valueobject.SourceToolRaw
		if strings.HasPrefix(s.ToolName, "calc_") {
			src = valueobject.SourceCalculation
		}
		meta := sourceMeta{toolName: s.ToolName, source: src}

		var v interface{}
		dec := json.NewDecoder(strings.NewReader(s.Raw))
		dec.UseNumber()
		if err := dec.Decode(&v); err != nil {
			// Not valid JSON — fall back to scanning the raw text for numbers
			// and dates so a plain-text tool result still grounds figures.
			collectFromText(s.Raw, meta, nums, dates)
			continue
		}
		walkJSON(v, meta, nums, dates)
	}
	return nums, dates
}

// walkJSON recurses through a decoded JSON value collecting numeric and date
// leaves into the corpora.
func walkJSON(v interface{}, meta sourceMeta, nums, dates map[string]sourceMeta) {
	switch t := v.(type) {
	case map[string]interface{}:
		for _, child := range t {
			walkJSON(child, meta, nums, dates)
		}
	case []interface{}:
		for _, child := range t {
			walkJSON(child, meta, nums, dates)
		}
	case json.Number:
		addNumber(t.String(), meta, nums)
	case string:
		if key, ok := canonicalNumber(t); ok {
			putIfAbsent(nums, key, meta)
		} else {
			extractNumbersFromStringLeaf(t, meta, nums)
		}
		for _, d := range reISODate.FindAllString(t, -1) {
			putIfAbsent(dates, d, meta)
		}
	}
}

// collectFromText is the non-JSON fallback corpus builder.
func collectFromText(text string, meta sourceMeta, nums, dates map[string]sourceMeta) {
	for _, loc := range reNumber.FindAllStringIndex(text, -1) {
		start, end := loc[0], loc[1]
		if isInsideDate(text, start, end) {
			continue
		}
		if isIdentifierEmbedded(text, start, end) {
			continue
		}
		addNumber(text[start:end], meta, nums)
	}
	for _, d := range reISODate.FindAllString(text, -1) {
		putIfAbsent(dates, d, meta)
	}
}

func extractNumbersFromStringLeaf(t string, meta sourceMeta, nums map[string]sourceMeta) {
	for _, loc := range reNumber.FindAllStringIndex(t, -1) {
		start, end := loc[0], loc[1]
		if isInsideDate(t, start, end) {
			continue
		}
		if isIdentifierEmbedded(t, start, end) {
			continue
		}
		addNumber(t[start:end], meta, nums)
	}
}

func addNumber(tok string, meta sourceMeta, nums map[string]sourceMeta) {
	if key, ok := canonicalNumber(tok); ok {
		putIfAbsent(nums, key, meta)
	}
}

func putIfAbsent(m map[string]sourceMeta, key string, meta sourceMeta) {
	if _, exists := m[key]; !exists {
		m[key] = meta
	}
}

// canonicalNumber parses a numeric token (possibly with thousands separators)
// into its exact rational canonical form. Returns false if not numeric.
func canonicalNumber(tok string) (string, bool) {
	clean := strings.ReplaceAll(strings.TrimSpace(tok), ",", "")
	if clean == "" || clean == "-" || clean == "." {
		return "", false
	}
	r := new(big.Rat)
	if _, ok := r.SetString(clean); !ok {
		return "", false
	}
	return r.RatString(), true
}

// detectFigures finds the financial figures (numbers + financial-context dates)
// in the answer that must be validated.
func detectFigures(answer string) []valueobject.FinancialFigure {
	var out []valueobject.FinancialFigure
	lower := strings.ToLower(answer)

	for _, loc := range reNumber.FindAllStringIndex(answer, -1) {
		start, end := loc[0], loc[1]
		raw := answer[start:end]

		// Skip a number that is part of an ISO date (handled separately).
		if isInsideDate(answer, start, end) {
			continue
		}
		// Skip digit-runs that are part of an identifier (UUIDs, codes, hashes)
		// — e.g. "00000000-0000-0000-0000-000000000004" or "GEF-0001000". These
		// are not financial figures and must never be flagged.
		if isIdentifierEmbedded(answer, start, end) {
			continue
		}
		// Skip code-like integers (multi-digit with a leading zero, e.g.
		// "0001000", "0000") — identifiers/codes, never monetary amounts.
		if isCodeLikeInteger(raw) {
			continue
		}
		// Skip list/ordinal markers ("5.", "6)") at the start of a line — these
		// are never financial figures.
		if isListOrdinal(answer, start, end) {
			continue
		}

		kind, include := classifyNumber(answer, lower, start, end, raw)
		if !include {
			continue
		}
		canon, ok := canonicalNumber(raw)
		if !ok {
			continue
		}
		out = append(out, valueobject.FinancialFigure{
			Raw:       raw,
			Canonical: canon,
			Kind:      kind,
			Start:     start,
			End:       end,
		})
	}

	// Financial-context dates.
	for _, loc := range reISODate.FindAllStringIndex(answer, -1) {
		start, end := loc[0], loc[1]
		if !nearAny(lower, start, dateKeywords, 40) {
			continue
		}
		out = append(out, valueobject.FinancialFigure{
			Raw:       answer[start:end],
			Canonical: answer[start:end],
			Kind:      valueobject.FigureDate,
			Start:     start,
			End:       end,
		})
	}
	return out
}

// classifyNumber decides whether a number token is a financial figure to
// validate, and its kind.
//
//   - Percentages, currency-adjacent numbers, decimals and thousands-grouped
//     numbers are always financial figures.
//   - A BARE integer is a financial figure when it is in financial context:
//     a financial unit (units/shares/lots) follows it, OR a financial keyword
//     precedes it (e.g. "NAV is 10", "Quantity is 1000", "AUM is 50000000").
//   - A bare integer immediately followed by a generic count noun
//     (funds/rows/holdings/steps/...) is a COUNT and is ignored — even if a
//     financial keyword appears earlier in the sentence ("top 5 holdings",
//     "in total there are 2 funds").
func classifyNumber(answer, lower string, start, end int, raw string) (valueobject.FigureKind, bool) {
	// percentage: next non-space char is '%'
	if isPercent(answer, end) {
		return valueobject.FigurePercentage, true
	}
	if currencyAdjacent(answer, start, end) {
		return valueobject.FigureMoney, true
	}
	if strings.Contains(raw, ".") || strings.Contains(raw, ",") {
		// A decimal/grouped amount near a known keyword gets a more specific
		// kind; otherwise it is a generic amount. Either way it is validated.
		return keywordKind(lower, start), true
	}

	// Bare integer.
	if unitFollows(lower, end) {
		return valueobject.FigureQuantity, true
	}
	if countNounFollows(lower, end) {
		return valueobject.FigureAmount, false
	}
	if precededByFinancialKeyword(lower, start) {
		return keywordKind(lower, start), true
	}
	return valueobject.FigureAmount, false
}

// financialKeywords are the terms that, when they PRECEDE a bare integer, make
// it a financial figure. Mirrors the agreed keyword list.
var financialKeywords = []string{
	// English
	"net asset value", "nav", "market value", "cost basis", "transaction value",
	"settlement amount", "assets under management", "aum",
	"unrealised_pnl", "unrealized_pnl", "unrealised pnl", "unrealized pnl",
	"p&l", "pnl", "profit", "loss", "roi", "return", "exposure", "allocation",
	"balance", "quantity", "valuation", "cash", "cost", "amount", "total",
	"weight", "ratio", "percentage", "price", "holding", "units", "shares",

	// Thai
	"มูลค่าสินทรัพย์สุทธิ", "มูลค่าตลาด", "ต้นทุน", "กำไร", "ขาดทุน",
	"ผลตอบแทน", "สัดส่วน", "จำนวน", "ปริมาณ", "ราคา", "ยอดคงเหลือ",
	"สินทรัพย์ภายใต้การจัดการ", "เงินสด", "น้ำหนัก", "อัตราส่วน",
}

// financialUnitWords following a number mark it a financial quantity.
var financialUnitWords = map[string]bool{
	// English
	"unit": true, "units": true, "share": true, "shares": true, "lot": true, "lots": true,

	// Thai
	"หน่วย": true,
	"หุ้น":  true,
	"กอง":  true,
}

// countNouns following a number mark it as a plain count, not a figure.
var countNouns = map[string]bool{
	// English
	"step": true, "steps": true, "row": true, "rows": true, "fund": true, "funds": true,
	"portfolio": true, "portfolios": true, "holding": true, "holdings": true,
	"item": true, "items": true, "result": true, "results": true, "section": true,
	"sections": true, "record": true, "records": true, "page": true, "pages": true,
	"column": true, "columns": true, "instrument": true, "instruments": true,
	"account": true, "accounts": true, "entry": true, "entries": true,
	"option": true, "options": true, "transaction": true, "transactions": true,
	"day": true, "days": true, "month": true, "months": true, "year": true, "years": true,
	"week": true, "weeks": true, "hour": true, "hours": true, "minute": true, "minutes": true,
	"second": true, "seconds": true,

	// Thai
	"ขั้นตอน":     true,
	"แถว":        true,
	"กองทุน":     true,
	"พอร์ต":      true,
	"พอร์ตโฟลิโอ": true,
	"รายการ":     true,
	"ผลลัพธ์":     true,
	"หน้า":       true,
	"คอลัมน์":     true,
	"บัญชี":      true,
	"ตัวเลือก":    true,
	"ธุรกรรม":    true,
	"วัน":        true,
	"เดือน":      true,
	"ปี":         true,
	"สัปดาห์":    true,
	"ชั่วโมง":     true,
	"นาที":       true,
	"วินาที":     true,
}

// nextWord returns the next alphabetic word after position end (skipping
// spaces/tabs), lower-cased input assumed.
func nextWord(lower string, end int) string {
	runes := []rune(lower[end:])
	i := 0
	for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t' || runes[i] == '\u00a0') {
		i++
	}
	if i >= len(runes) {
		return ""
	}
	start := i
	for i < len(runes) {
		r := runes[i]
		if unicode.IsLetter(r) || unicode.IsMark(r) {
			i++
		} else {
			break
		}
	}
	return string(runes[start:i])
}

func unitFollows(lower string, end int) bool { return financialUnitWords[nextWord(lower, end)] }

func countNounFollows(lower string, end int) bool { return countNouns[nextWord(lower, end)] }

// precededByFinancialKeyword reports whether a financial keyword appears in the
// IMMEDIATE window before the number (not after — so "top 5 holdings" stays a
// count while "NAV is 10" is a figure). The window is tight (covers phrapings
// like "market value is 200000") so that a keyword elsewhere in the sentence
// does not sweep up an unrelated bare integer.
func precededByFinancialKeyword(lower string, start int) bool {
	runes := []rune(lower)
	runeStart := byteIndexToRuneIndex(runes, start)

	const window = 20
	lo := runeStart - window
	if lo < 0 {
		lo = 0
	}
	ctx := string(runes[lo:runeStart])
	for _, kw := range financialKeywords {
		if strings.Contains(ctx, kw) {
			return true
		}
	}
	return false
}

func isPercent(answer string, end int) bool {
	for i := end; i < len(answer) && i < end+2; i++ {
		c := answer[i]
		if c == ' ' {
			continue
		}
		return c == '%'
	}
	return false
}

func currencyAdjacent(answer string, start, end int) bool {
	runes := []rune(answer)
	runeStart := byteIndexToRuneIndex(runes, start)
	runeEnd := byteIndexToRuneIndex(runes, end)

	lo := runeStart - 8
	if lo < 0 {
		lo = 0
	}
	hi := runeEnd + 8
	if hi > len(runes) {
		hi = len(runes)
	}
	before := string(runes[lo:runeEnd])
	after := string(runes[runeStart:hi])
	return reCurrency.MatchString(before) || reCurrency.MatchString(after)
}

func byteIndexToRuneIndex(runes []rune, byteIdx int) int {
	byteOffset := 0
	for idx, r := range runes {
		if byteOffset >= byteIdx {
			return idx
		}
		byteOffset += len(string(r))
	}
	return len(runes)
}

// keywordKind picks a kind based on a financial keyword near the number.
func keywordKind(lower string, pos int) valueobject.FigureKind {
	switch {
	case nearAny(lower, pos, []string{"nav", "net asset value"}, 30):
		return valueobject.FigureNAV
	case nearAny(lower, pos, []string{"market value", "aum", "assets under management"}, 30):
		return valueobject.FigureMarketValue
	case nearAny(lower, pos, []string{"p&l", "pnl", "profit", "loss", "unrealised", "unrealized", "realised", "realized"}, 30):
		return valueobject.FigurePnL
	case nearAny(lower, pos, []string{"balance", "cash"}, 24):
		return valueobject.FigureBalance
	case nearAny(lower, pos, []string{"price", "average cost", "cost"}, 24):
		return valueobject.FigurePrice
	case nearAny(lower, pos, []string{"quantity", "units", "shares"}, 24):
		return valueobject.FigureQuantity
	case nearAny(lower, pos, []string{"roi", "return", "exposure", "ratio", "allocation"}, 30):
		return valueobject.FigureRatio
	default:
		return valueobject.FigureAmount
	}
}

// nearAny reports whether any keyword appears within window runes before/after pos.
func nearAny(lower string, pos int, keywords []string, window int) bool {
	runes := []rune(lower)
	runePos := byteIndexToRuneIndex(runes, pos)

	lo := runePos - window
	if lo < 0 {
		lo = 0
	}
	hi := runePos + window
	if hi > len(runes) {
		hi = len(runes)
	}
	ctx := string(runes[lo:hi])
	for _, kw := range keywords {
		if strings.Contains(ctx, kw) {
			return true
		}
	}
	return false
}

func isInsideDate(answer string, start, end int) bool {
	for _, loc := range reISODate.FindAllStringIndex(answer, -1) {
		if start >= loc[0] && end <= loc[1] {
			return true
		}
	}
	return false
}

// isIdentifierEmbedded reports whether the number token at [start,end) is part
// of a larger identifier (UUID, code, hash) rather than a standalone figure. A
// digit-run is identifier-embedded when it is adjacent to a letter/underscore,
// or to a hyphen acting as a segment separator inside an identifier — e.g.
// "00000000-0000-0000-0000-000000000004" or "GEF-0001000". This prevents the
// validator from slicing fragments out of fund ids/codes and demanding they be
// "verified".
func isIdentifierEmbedded(answer string, start, end int) bool {
	if start > 0 {
		b := answer[start-1]
		if isAlphaNum(b) || b == '_' {
			return true
		}
		// A hyphen before the token is a separator (not a minus sign) when the
		// character before IT is alphanumeric, i.e. we are mid-identifier.
		if b == '-' && start-2 >= 0 && isAlphaNum(answer[start-2]) {
			return true
		}
	}
	if end < len(answer) {
		a := answer[end]
		if isAlpha(a) || a == '_' {
			return true
		}
		if a == '-' && end+1 < len(answer) && isAlphaNum(answer[end+1]) {
			return true
		}
	}
	return false
}

// isCodeLikeInteger reports whether raw is a multi-digit integer with a leading
// zero (e.g. "0001000", "0000", "-0000") — an identifier/code, never a monetary
// amount. Numbers with a decimal point or thousands separator are exempt.
func isCodeLikeInteger(raw string) bool {
	s := strings.TrimPrefix(raw, "-")
	if strings.ContainsAny(s, ".,") {
		return false
	}
	return len(s) > 1 && s[0] == '0'
}

// isListOrdinal reports whether the number is a list/ordinal marker: a number
// at the start of a line (only whitespace before it on that line) immediately
// followed by '.' or ')'. e.g. "5." or "6)" in a numbered list. Never a figure.
func isListOrdinal(answer string, start, end int) bool {
	if end >= len(answer) || (answer[end] != '.' && answer[end] != ')') {
		return false
	}
	for i := start - 1; i >= 0; i-- {
		c := answer[i]
		if c == '\n' || c == '\r' {
			return true
		}
		if c != ' ' && c != '\t' {
			return false
		}
	}
	return true // start of the answer
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isAlphaNum(c byte) bool {
	return isAlpha(c) || (c >= '0' && c <= '9')
}

// RedactFigures replaces the given figures with a placeholder. It is provided
// for callers that can prove the remaining answer is still useful and not
// misleading; the default policy blocks instead (see ValidateAnswer). Pure and
// position-safe (replaces right-to-left).
func RedactFigures(answer string, figures []valueobject.FinancialFigure, placeholder string) string {
	out := answer
	for i := len(figures) - 1; i >= 0; i-- {
		f := figures[i]
		if f.Start < 0 || f.End > len(out) || f.Start >= f.End {
			continue
		}
		out = out[:f.Start] + placeholder + out[f.End:]
	}
	return out
}
