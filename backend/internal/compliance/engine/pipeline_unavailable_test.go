package engine

import (
	"testing"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

func TestComplianceVerdicts_UnavailableAlwaysBlocks(t *testing.T) {
	t.Parallel()

	for _, severity := range []vo.Severity{vo.SeverityMonitor, vo.SeverityWarn, vo.SeverityBlock} {
		severity := severity
		t.Run(string(severity), func(t *testing.T) {
			t.Parallel()
			raw, final := complianceVerdicts(vo.ComplianceStatusUnavailable, vo.VerdictPass, severity)
			if raw != vo.VerdictBlock || final != vo.VerdictBlock {
				t.Fatalf("UNAVAILABLE with %s severity = raw %s / final %s, want BLOCK / BLOCK", severity, raw, final)
			}
		})
	}
}

func TestComplianceVerdicts_EvaluatedRetainsSeverityPolicy(t *testing.T) {
	t.Parallel()

	raw, final := complianceVerdicts(vo.ComplianceStatusEvaluated, vo.VerdictBlock, vo.SeverityMonitor)
	if raw != vo.VerdictBlock || final != vo.VerdictPass {
		t.Fatalf("evaluated MONITOR = raw %s / final %s, want BLOCK / PASS", raw, final)
	}
}
