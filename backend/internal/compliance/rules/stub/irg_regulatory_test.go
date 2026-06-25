package stub

import (
	"context"
	"testing"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

func TestThaiSECRule_ReturnsWarn(t *testing.T) {
	rule := &ThaiSECRule{}
	result, err := rule.Evaluate(context.Background(), spi.CheckInput{}, spi.DataBundle{}, spi.ParameterSet{})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if result.Verdict != vo.VerdictWarn {
		t.Errorf("expected VerdictWarn, got %q — stub must not silently pass unimplemented regulatory rules", result.Verdict)
	}
}

func TestThaiSECRule_EvidenceStatus_NotConfigured(t *testing.T) {
	rule := &ThaiSECRule{}
	result, err := rule.Evaluate(context.Background(), spi.CheckInput{}, spi.DataBundle{}, spi.ParameterSet{})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	status, ok := result.Evidence.Metrics["status"]
	if !ok {
		t.Fatal("Evidence.Metrics must contain 'status' key")
	}
	if status != "NOT_CONFIGURED" {
		t.Errorf("expected Evidence.Metrics[status] = 'NOT_CONFIGURED', got %q", status)
	}
}
