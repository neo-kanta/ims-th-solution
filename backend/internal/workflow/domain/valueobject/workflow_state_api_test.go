package valueobject_test

import (
	"testing"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
)

func TestWorkflowStateToAPIName_CanonicalNames(t *testing.T) {
	cases := []struct {
		state    vo.WorkflowState
		expected string
	}{
		{vo.StateNotStarted, "NOT_STARTED"},
		{vo.StateInvestmentDayStarted, "INVESTMENT_DAY_STARTED"},
		{vo.StateManagerApproved, "MANAGER_APPROVED"},
		{vo.StateTransactionClosed, "TRANSACTION_CLOSED"},
		{vo.StateAccountingClosed, "ACCOUNTING_CLOSED"},
	}
	for _, tc := range cases {
		t.Run(string(tc.state), func(t *testing.T) {
			got := tc.state.ToAPIName()
			if got != tc.expected {
				t.Errorf("ToAPIName(%q) = %q, want %q", tc.state, got, tc.expected)
			}
		})
	}
}

func TestWorkflowStateToAPIName_Phase1CompatNames(t *testing.T) {
	if vo.StateDayOpen.ToAPIName() != "INVESTMENT_DAY_STARTED" {
		t.Errorf("StateDayOpen.ToAPIName() = %q, want INVESTMENT_DAY_STARTED", vo.StateDayOpen.ToAPIName())
	}
	if vo.StateManagerApprovedEOD.ToAPIName() != "MANAGER_APPROVED" {
		t.Errorf("StateManagerApprovedEOD.ToAPIName() = %q, want MANAGER_APPROVED", vo.StateManagerApprovedEOD.ToAPIName())
	}
}

func TestWorkflowActionToAPIName(t *testing.T) {
	cases := []struct {
		action   vo.WorkflowAction
		expected string
	}{
		{vo.ActionOpenDay, vo.APIOpStartInvestmentDay},
		{vo.ActionCancelDayStart, vo.APIOpCancelInvestmentDay},
		{vo.ActionApprove, vo.APIOpManagerApprove},
		{vo.ActionCancelApproval, vo.APIOpCancelManagerApproval},
		{vo.ActionCloseTransactions, vo.APIOpCloseTransaction},
		{vo.ActionCancelTransactionClose, vo.APIOpCancelTransactionClose},
		{vo.ActionCloseAccounting, vo.APIOpCloseAccounting},
		{vo.ActionRollbackAccountingClose, vo.APIOpCancelAccountingClose},
	}
	for _, tc := range cases {
		t.Run(string(tc.action), func(t *testing.T) {
			got := tc.action.ToAPIName()
			if got != tc.expected {
				t.Errorf("ToAPIName(%q) = %q, want %q", tc.action, got, tc.expected)
			}
		})
	}
}

func TestParseAPIOperationType_AllValidOperations(t *testing.T) {
	valid := []struct {
		opType   string
		expected vo.WorkflowAction
	}{
		{vo.APIOpStartInvestmentDay, vo.ActionOpenDay},
		{vo.APIOpCancelInvestmentDay, vo.ActionCancelDayStart},
		{vo.APIOpManagerApprove, vo.ActionApprove},
		{vo.APIOpCancelManagerApproval, vo.ActionCancelApproval},
		{vo.APIOpCloseTransaction, vo.ActionCloseTransactions},
		{vo.APIOpCancelTransactionClose, vo.ActionCancelTransactionClose},
		{vo.APIOpCloseAccounting, vo.ActionCloseAccounting},
		{vo.APIOpCancelAccountingClose, vo.ActionRollbackAccountingClose},
	}
	for _, tc := range valid {
		t.Run(tc.opType, func(t *testing.T) {
			action, ok := vo.ParseAPIOperationType(tc.opType)
			if !ok {
				t.Fatalf("ParseAPIOperationType(%q) returned ok=false", tc.opType)
			}
			if action != tc.expected {
				t.Errorf("ParseAPIOperationType(%q) = %q, want %q", tc.opType, action, tc.expected)
			}
		})
	}
}

func TestParseAPIOperationType_InvalidOperation(t *testing.T) {
	invalid := []string{"OPEN_DAY", "APPROVE", "CLOSE_TRANSACTIONS", "UNKNOWN", ""}
	for _, op := range invalid {
		t.Run(op, func(t *testing.T) {
			_, ok := vo.ParseAPIOperationType(op)
			if ok {
				t.Errorf("ParseAPIOperationType(%q) returned ok=true, want false", op)
			}
		})
	}
}

func TestAllowedActions_UsesAPINames(t *testing.T) {
	// Verify AllowedActions returns actions that all have valid ToAPIName mappings
	allStates := []vo.WorkflowState{
		vo.StateNotStarted,
		vo.StateInvestmentDayStarted,
		vo.StateDayOpen,
		vo.StateManagerApprovedEOD,
		vo.StateManagerApproved,
		vo.StateTransactionClosed,
		vo.StateAccountingClosed,
	}
	for _, state := range allStates {
		actions := vo.AllowedActions(state)
		for _, action := range actions {
			apiName := action.ToAPIName()
			if apiName == "" {
				t.Errorf("AllowedActions(%q) contains action %q with empty ToAPIName()", state, action)
			}
			// Round-trip: parse back to internal action
			parsed, ok := vo.ParseAPIOperationType(apiName)
			if !ok {
				t.Errorf("ParseAPIOperationType(%q) returned false for action %q from state %q", apiName, action, state)
			}
			if parsed != action {
				t.Errorf("round-trip failed: action %q → API %q → %q", action, apiName, parsed)
			}
		}
	}
}
