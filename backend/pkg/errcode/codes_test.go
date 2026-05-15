package errcode

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestAll_UniqueAndUpperSnake(t *testing.T) {
	t.Parallel()
	seen := make(map[string]struct{})
	for _, c := range All() {
		if _, dup := seen[c]; dup {
			t.Errorf("duplicate code in registry: %q", c)
		}
		seen[c] = struct{}{}
		if c != strings.ToUpper(c) {
			t.Errorf("code %q must be UPPER_SNAKE_CASE", c)
		}
		if strings.ContainsAny(c, " \t-/") {
			t.Errorf("code %q contains forbidden characters", c)
		}
	}
}

func TestAll_CoversBriefMandatedCodes(t *testing.T) {
	t.Parallel()
	mandated := []string{
		"OVERSELL",
		"WORKFLOW_TRADE_NOT_ALLOWED",
		"WORKFLOW_LOCKED",
		"COMPLIANCE_REJECTED",
		"VERSION_MISMATCH",
		"CODE_ALREADY_EXISTS",
		"INSTRUMENT_NOT_MAPPED",
		"PRICE_CURRENCY_MISMATCH",
		"INSTRUMENT_NOT_FOUND",
		"FUND_NOT_FOUND",
		"PORTFOLIO_NOT_FOUND",
		"TRANSACTION_NOT_FOUND",
		"ALREADY_REVERSED",
		"CANNOT_REVERSE_REVERSAL",
		"FUND_HAS_ACTIVE_PORTFOLIOS",
		"PORTFOLIO_HAS_OPEN_ACTIVITY",
		"INVALID_REQUEST",
		"INTERNAL_ERROR",
	}

	known := make(map[string]struct{}, len(All()))
	for _, c := range All() {
		known[c] = struct{}{}
	}
	for _, c := range mandated {
		if _, ok := known[c]; !ok {
			t.Errorf("brief-mandated code %q is missing from errcode.All()", c)
		}
	}
}

func TestDefaultStatus_KnownCodes(t *testing.T) {
	t.Parallel()
	cases := map[string]int{
		CodeInvalidRequest:          http.StatusBadRequest,
		CodeInstrumentNotFound:      http.StatusNotFound,
		CodeDecisionNotFound:        http.StatusNotFound,
		CodeVersionMismatch:         http.StatusConflict,
		CodeAlreadyReversed:         http.StatusConflict,
		CodeFundHasActivePortfolios: http.StatusConflict,
		CodeOversell:                http.StatusUnprocessableEntity,
		CodeWorkflowTradeNotAllowed: http.StatusUnprocessableEntity,
		CodeComplianceRejected:      http.StatusUnprocessableEntity,
		CodeUnitisedNotSupported:    http.StatusUnprocessableEntity,
		CodeInternal:                http.StatusInternalServerError,
	}
	for code, want := range cases {
		if got := DefaultStatus(code); got != want {
			t.Errorf("DefaultStatus(%q) = %d, want %d", code, got, want)
		}
	}
}

func TestDefaultStatus_UnknownDefaultsTo500(t *testing.T) {
	t.Parallel()
	if got := DefaultStatus("FANCY_NEW_CODE"); got != http.StatusInternalServerError {
		t.Errorf("DefaultStatus(unknown) = %d, want 500", got)
	}
}

type codedErr struct {
	code   string
	msg    string
	status int
	det    map[string]any
}

func (e *codedErr) Error() string                { return e.msg }
func (e *codedErr) ErrorCode() string            { return e.code }
func (e *codedErr) ErrorDetails() map[string]any { return e.det }
func (e *codedErr) HTTPStatus() int              { return e.status }

func TestCodeOf_FindsCodedInChain(t *testing.T) {
	t.Parallel()
	inner := &codedErr{code: CodeOversell, msg: "no inventory"}
	wrapped := errors.New("outer: " + inner.Error())
	// errors.As should find the typed error directly.
	gotCode, ok := CodeOf(inner)
	if !ok || gotCode != CodeOversell {
		t.Errorf("CodeOf(inner) = (%q, %v), want (OVERSELL, true)", gotCode, ok)
	}
	// A plain error wraps to no Coded match, so we get INTERNAL_ERROR.
	gotCode, ok = CodeOf(wrapped)
	if !ok || gotCode != CodeInternal {
		t.Errorf("CodeOf(plain) = (%q, %v), want (INTERNAL_ERROR, true)", gotCode, ok)
	}
}

func TestCodeOf_Nil(t *testing.T) {
	t.Parallel()
	if c, ok := CodeOf(nil); ok || c != "" {
		t.Errorf("CodeOf(nil) = (%q, %v), want (\"\", false)", c, ok)
	}
}

func TestStatusOf_PrefersHTTPStatus(t *testing.T) {
	t.Parallel()
	err := &codedErr{code: CodeInvalidRequest, status: http.StatusTeapot}
	if got := StatusOf(err); got != http.StatusTeapot {
		t.Errorf("StatusOf with explicit status = %d, want 418", got)
	}
}

func TestStatusOf_FallsBackToCodeMapping(t *testing.T) {
	t.Parallel()
	err := &codedErr{code: CodeOversell}
	if got := StatusOf(err); got != http.StatusUnprocessableEntity {
		t.Errorf("StatusOf falls back to code mapping = %d, want 422", got)
	}
}

func TestDetailsOf_ReturnsCopy(t *testing.T) {
	t.Parallel()
	src := map[string]any{"min_quantity": 100}
	err := &codedErr{code: CodeOversell, det: src}
	got := DetailsOf(err)
	if got["min_quantity"] != 100 {
		t.Fatalf("DetailsOf returned wrong content: %v", got)
	}
	got["min_quantity"] = 999
	if src["min_quantity"] != 100 {
		t.Errorf("DetailsOf must return a copy; mutation leaked back to caller")
	}
}
