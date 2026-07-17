package command

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

func TestLiveComplianceStatusError_UntypedPassFailsClosed(t *testing.T) {
	result := &contract.ProposedOrderResult{
		CheckGroupID:   uuid.New(),
		Verdict:        contract.ComplianceVerdictPass,
		RulesEvaluated: 1,
	}

	err := liveComplianceStatusError(vo.PortfolioTypeLive, result)
	var unavailable *ErrComplianceUnavailable
	if !errors.As(err, &unavailable) {
		t.Fatalf("untyped PASS must fail closed as ErrComplianceUnavailable, got %T: %v", err, err)
	}
}

func TestLiveComplianceStatusError_NilAndUnsupportedVerdictsFailClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		result *contract.ProposedOrderResult
	}{
		{name: "nil result", result: nil},
		{
			name: "evaluated empty verdict",
			result: &contract.ProposedOrderResult{
				CheckGroupID:   uuid.New(),
				Status:         contract.ComplianceStatusEvaluated,
				RulesEvaluated: 1,
			},
		},
		{
			name: "evaluated unknown verdict",
			result: &contract.ProposedOrderResult{
				CheckGroupID:   uuid.New(),
				Status:         contract.ComplianceStatusEvaluated,
				Verdict:        contract.ComplianceVerdict("ALLOW"),
				RulesEvaluated: 1,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := liveComplianceStatusError(vo.PortfolioTypeLive, tt.result)
			var unavailable *ErrComplianceUnavailable
			if !errors.As(err, &unavailable) {
				t.Fatalf("error = %T %v, want *ErrComplianceUnavailable", err, err)
			}
		})
	}
}
