package domain

import (
	"net/http"
	"testing"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
)

func TestInvestmentErrors_ImplementCoded(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		err      errcode.Coded
		wantCode string
		wantHTTP int
	}{
		{"decision-not-found", &ErrDecisionNotFound{DecisionID: "x"}, errcode.CodeDecisionNotFound, http.StatusNotFound},
		{"decision-not-draft", &ErrDecisionNotDraft{DecisionID: "x", CurrentStatus: "SUBMITTED"}, errcode.CodeDecisionNotDraft, http.StatusConflict},
		{"invalid-decision-req", &ErrInvalidDecisionRequest{Field: "actor_id", Detail: "is required"}, errcode.CodeInvalidRequest, http.StatusBadRequest},
		{"invalid-process-guard", &ErrInvalidProcessGuardRequest{Field: "user_id", Detail: "is required"}, errcode.CodeInvalidRequest, http.StatusBadRequest},
		{"compliance-rejected", &ErrComplianceRejected{DecisionID: "d", CheckGroupID: "cg", Message: "blocked"}, errcode.CodeComplianceRejected, http.StatusUnprocessableEntity},
		{"post-trade-blocked", &ErrPostTradeBlock{ContractID: "c", BusinessDate: "2026-05-06", CheckGroupID: "cg", BreachCount: 2}, errcode.CodePostTradeBlocked, http.StatusUnprocessableEntity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.err.ErrorCode(); got != tc.wantCode {
				t.Errorf("%s ErrorCode() = %q, want %q", tc.name, got, tc.wantCode)
			}
			if got := errcode.StatusOf(tc.err); got != tc.wantHTTP {
				t.Errorf("%s StatusOf() = %d, want %d", tc.name, got, tc.wantHTTP)
			}
		})
	}
}

func TestInvestmentErrors_DetailsArePresentAndSafe(t *testing.T) {
	t.Parallel()
	err := &ErrComplianceRejected{DecisionID: "d-1", CheckGroupID: "cg-1", Message: "blocked"}
	det := errcode.DetailsOf(err)
	if det["decision_id"] != "d-1" {
		t.Errorf("decision_id missing: %+v", det)
	}
	if det["check_group_id"] != "cg-1" {
		t.Errorf("check_group_id missing: %+v", det)
	}
	if _, hasMsg := det["message"]; hasMsg {
		t.Errorf("ErrorDetails should not echo human-readable message; that field belongs to envelope.message")
	}
}

func TestInvestmentErrors_InvalidRequestOmitsEmptyDetails(t *testing.T) {
	t.Parallel()
	err := &ErrInvalidDecisionRequest{}
	if det := err.ErrorDetails(); det != nil {
		t.Errorf("expected nil details for fully-empty request, got %+v", det)
	}
}
