package errcode

import (
	"errors"
	"net/http"
)

// Coded is implemented by domain errors that carry a stable error_code.
// The HTTP envelope writer prefers this interface over heuristic mapping.
type Coded interface {
	error
	ErrorCode() string
}

// Detailed is implemented by domain errors that want to attach structured
// context to the envelope's `details` field. Implementations MUST return a
// map that is safe to publish to API consumers — never include credentials,
// stack traces, or internal IDs that the caller is not authorised to see.
type Detailed interface {
	ErrorDetails() map[string]any
}

// HTTPStatus is implemented by domain errors that have a strong opinion on
// the HTTP status code. When omitted, the envelope writer maps the error
// code to a status via DefaultStatus.
type HTTPStatus interface {
	HTTPStatus() int
}

// CodeOf walks the unwrap chain looking for the nearest Coded error and
// returns its code. When no Coded error is found, CodeInternal is returned.
//
// nil → ("", false) so the caller can decide whether to short-circuit.
func CodeOf(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	var c Coded
	if errors.As(err, &c) {
		return c.ErrorCode(), true
	}
	return CodeInternal, true
}

// DetailsOf returns a copy of the structured details attached to err, or
// nil when none are attached. A copy is returned so the envelope writer
// cannot accidentally mutate caller state.
func DetailsOf(err error) map[string]any {
	if err == nil {
		return nil
	}
	var d Detailed
	if !errors.As(err, &d) {
		return nil
	}
	src := d.ErrorDetails()
	if len(src) == 0 {
		return nil
	}
	cp := make(map[string]any, len(src))
	for k, v := range src {
		cp[k] = v
	}
	return cp
}

// StatusOf returns the HTTP status the error wants, falling back to
// DefaultStatus when none is declared.
func StatusOf(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var hs HTTPStatus
	if errors.As(err, &hs) {
		if s := hs.HTTPStatus(); s > 0 {
			return s
		}
	}
	if code, ok := CodeOf(err); ok {
		return DefaultStatus(code)
	}
	return http.StatusInternalServerError
}

// DefaultStatus maps a code to its canonical HTTP status. Unknown codes
// default to 500.
func DefaultStatus(code string) int {
	switch code {
	case CodeInvalidRequest:
		return http.StatusBadRequest

	case CodeInstrumentNotFound,
		CodeFundNotFound,
		CodePortfolioNotFound,
		CodeTransactionNotFound,
		CodeDecisionNotFound,
		CodeExecutionNotFound,
		CodeConfirmationNotFound,
		CodeContractNotFound:
		return http.StatusNotFound

	case CodeVersionMismatch,
		CodeCodeAlreadyTaken,
		CodeDecisionNotDraft,
		CodeDecisionLifecycle,
		CodeExecutionLifecycle,
		CodeConfirmationLifecycle,
		CodeAlreadyReversed,
		CodeCannotReverseReversal,
		CodeFundHasActivePortfolios,
		CodePortfolioHasOpenActivity,
		CodeResearchReportInvalidateBlocked:
		return http.StatusConflict

	case CodeOversell,
		CodeWorkflowTradeNotAllowed,
		CodeWorkflowLocked,
		CodeComplianceRejected,
		CodePostTradeBlocked,
		CodeInstrumentNotMapped,
		CodePriceCurrencyMismatch,
		CodeUnitisedNotSupported,
		CodeIncompleteFundValuation,
		CodeDecisionReferenceBad,
		CodeConfirmationMismatch,
		CodeClosePendingConfirm:
		return http.StatusUnprocessableEntity

	case CodeForbidden:
		return http.StatusForbidden

	case CodeInternal:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}
