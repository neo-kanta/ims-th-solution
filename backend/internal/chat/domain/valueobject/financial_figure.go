package valueobject

// This file defines the value types used by the post-generation numeric
// financial validator (domain/policy/numeric_validator.go). The validator is
// the structural guarantee behind the module's core rule: every financial
// figure in a final answer must trace to a current-turn MCP tool result (raw
// data tool OR deterministic calculation tool). Figures that cannot be traced
// are treated as unverified and the answer is blocked.

// FigureKind classifies a detected financial figure so the validator can apply
// the right matching and context rules. Dates are validated ONLY when they
// appear in a financial-date context (see FinancialDateContext).
type FigureKind string

const (
	FigureMoney       FigureKind = "money"        // currency-adjacent amount
	FigurePrice       FigureKind = "price"        // price / NAV per unit
	FigureNAV         FigureKind = "nav"          // net asset value
	FigureQuantity    FigureKind = "quantity"     // units / shares
	FigurePercentage  FigureKind = "percentage"   // N%
	FigureRatio       FigureKind = "ratio"        // ROI / exposure ratio
	FigureMarketValue FigureKind = "market_value" // market value / AUM
	FigurePnL         FigureKind = "pnl"          // profit / loss
	FigureBalance     FigureKind = "balance"      // cash balance
	FigureAmount      FigureKind = "amount"       // decimal/thousands amount, kind not otherwise known
	FigureDate        FigureKind = "date"         // financial-context date only
)

// FigureSource records where a figure was traced to. "none" means the figure
// could not be matched against any current-turn tool result.
type FigureSource string

const (
	SourceToolRaw     FigureSource = "tool_raw"    // matched a raw data tool result
	SourceCalculation FigureSource = "calculation" // matched a deterministic calc tool result
	SourceNone        FigureSource = "none"        // untraceable -> unverified
)

// FinancialFigure is one numeric (or date) value detected in the assistant's
// final answer that the validator must account for.
type FinancialFigure struct {
	// Raw is the exact substring matched in the answer (e.g. "10.25", "12.5%",
	// "1,234.56", "2026-06-05").
	Raw string
	// Canonical is the normalized form used for matching. For numbers this is
	// the exact rational string (commas/percent stripped); for dates it is the
	// YYYY-MM-DD prefix.
	Canonical string
	// Kind classifies the figure.
	Kind FigureKind
	// Start/End are byte offsets of Raw within the answer (for redaction/UX).
	Start int
	End   int
}

// IsDate reports whether the figure is a (financial-context) date.
func (f FinancialFigure) IsDate() bool { return f.Kind == FigureDate }

// ProvenanceBinding links a detected figure to the tool result that grounds it.
// Bindings are persisted into the assistant message's provenance_map so an
// auditor can see which tool each cited figure came from.
type ProvenanceBinding struct {
	Figure string       `json:"figure"`
	Source FigureSource `json:"source"`
	// ToolName is the tool whose result contained the value (best-effort; the
	// first matching current-turn tool).
	ToolName string `json:"tool_name,omitempty"`
}

// ValidationViolation is one figure that could not be traced to any source.
type ValidationViolation struct {
	Figure string     `json:"figure"`
	Kind   FigureKind `json:"kind"`
	Reason string     `json:"reason"`
}

// ValidationDecision is the validator's verdict for the whole answer.
type ValidationDecision string

const (
	// DecisionAllow: every detected financial figure traced to a source.
	DecisionAllow ValidationDecision = "allow"
	// DecisionBlock: one or more figures were unverified. Per the financial
	// safety policy we prefer blocking (safe fallback) over inline redaction,
	// because automatic redaction cannot guarantee the remaining answer stays
	// useful and not misleading.
	DecisionBlock ValidationDecision = "block"
)

// ValidationResult is the full outcome the agent loop acts on.
type ValidationResult struct {
	Decision   ValidationDecision    `json:"decision"`
	FinalText  string                `json:"-"` // text to show the user (original on allow, safe fallback on block)
	Violations []ValidationViolation `json:"violations,omitempty"`
	Bindings   []ProvenanceBinding   `json:"bindings,omitempty"`
	// Checked is the number of financial figures evaluated (0 means the answer
	// had no financial figures at all — always an allow).
	Checked int `json:"checked"`
}

// OK reports whether the answer is safe to show as-is.
func (r ValidationResult) OK() bool { return r.Decision == DecisionAllow }
