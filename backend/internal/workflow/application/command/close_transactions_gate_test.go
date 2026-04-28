package command

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

func TestEvaluatePostTradeGate_NilVerification_Passes(t *testing.T) {
	// Nil verifier result (verifier disabled or no scan) must not block.
	if err := EvaluatePostTradeGate(uuid.New(), time.Now(), nil); err != nil {
		t.Fatalf("nil verification must pass, got %v", err)
	}
}

func TestEvaluatePostTradeGate_NoBlock_Passes(t *testing.T) {
	v := &contract.PostTradeVerificationResult{
		CheckGroupID:      uuid.New(),
		Verdict:           contract.ComplianceVerdictPass,
		HasBlockingBreach: false,
	}
	if err := EvaluatePostTradeGate(uuid.New(), time.Now(), v); err != nil {
		t.Fatalf("PASS verdict must not block: %v", err)
	}
}

func TestEvaluatePostTradeGate_Warn_Passes(t *testing.T) {
	// WARN-only scan — HasBlockingBreach is false.
	v := &contract.PostTradeVerificationResult{
		CheckGroupID:      uuid.New(),
		Verdict:           contract.ComplianceVerdictWarn,
		HasBlockingBreach: false,
		Breaches: []contract.PostTradeBreach{{
			BreachID: uuid.New(), Verdict: contract.ComplianceVerdictWarn,
			Severity: "WARN", Message: "exposure elevated",
		}},
	}
	if err := EvaluatePostTradeGate(uuid.New(), time.Now(), v); err != nil {
		t.Fatalf("WARN must not block close: %v", err)
	}
}

func TestEvaluatePostTradeGate_BlockRefusesTransition(t *testing.T) {
	contractID := uuid.New()
	date := time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC)
	groupID := uuid.New()
	v := &contract.PostTradeVerificationResult{
		CheckGroupID:      groupID,
		Verdict:           contract.ComplianceVerdictBlock,
		HasBlockingBreach: true,
		Breaches: []contract.PostTradeBreach{
			{BreachID: uuid.New(), Verdict: contract.ComplianceVerdictBlock, Severity: "BLOCK", Message: "blacklist hit"},
			{BreachID: uuid.New(), Verdict: contract.ComplianceVerdictBlock, Severity: "BLOCK", Message: "rating too low"},
			{BreachID: uuid.New(), Verdict: contract.ComplianceVerdictWarn, Severity: "WARN", Message: "sector 45%"},
		},
	}
	err := EvaluatePostTradeGate(contractID, date, v)
	if err == nil {
		t.Fatal("expected BLOCK gate error")
	}
	var blocked *domain.ErrPostTradeBreachesBlockClose
	if !errors.As(err, &blocked) {
		t.Fatalf("want *ErrPostTradeBreachesBlockClose, got %T", err)
	}
	if blocked.ContractID != contractID.String() {
		t.Fatalf("contract ID not propagated")
	}
	if blocked.CheckGroupID != groupID.String() {
		t.Fatalf("check group ID not propagated")
	}
	if blocked.BreachCount != 2 {
		t.Fatalf("want 2 BLOCK breaches counted, got %d", blocked.BreachCount)
	}
	if blocked.BusinessDate != "2026-04-24" {
		t.Fatalf("date format wrong: %s", blocked.BusinessDate)
	}
}

func TestEvaluatePostTradeGate_HasBlockingTrue_NoBlockEntries_StillBlocks(t *testing.T) {
	// Defensive: if an implementation sets HasBlockingBreach=true but the slice
	// has no BLOCK entries (e.g. summary-only response), we still refuse the close.
	v := &contract.PostTradeVerificationResult{
		CheckGroupID:      uuid.New(),
		Verdict:           contract.ComplianceVerdictBlock,
		HasBlockingBreach: true,
		Breaches:          []contract.PostTradeBreach{},
	}
	err := EvaluatePostTradeGate(uuid.New(), time.Now(), v)
	var blocked *domain.ErrPostTradeBreachesBlockClose
	if !errors.As(err, &blocked) {
		t.Fatalf("want gate error, got %v", err)
	}
	if blocked.BreachCount < 1 {
		t.Fatalf("want at least 1 block reported, got %d", blocked.BreachCount)
	}
}
