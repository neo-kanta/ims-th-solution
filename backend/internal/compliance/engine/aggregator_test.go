package engine_test

import (
	"testing"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/engine"
)

func TestAggregateVerdict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		records []entity.CheckRecord
		want    vo.Verdict
	}{
		{
			name:    "empty records → PASS",
			records: nil,
			want:    vo.VerdictPass,
		},
		{
			name: "all PASS → PASS",
			records: []entity.CheckRecord{
				{FinalVerdict: vo.VerdictPass},
				{FinalVerdict: vo.VerdictPass},
			},
			want: vo.VerdictPass,
		},
		{
			name: "one WARN among PASSes → WARN",
			records: []entity.CheckRecord{
				{FinalVerdict: vo.VerdictPass},
				{FinalVerdict: vo.VerdictWarn},
				{FinalVerdict: vo.VerdictPass},
			},
			want: vo.VerdictWarn,
		},
		{
			name: "one BLOCK among WARNs → BLOCK",
			records: []entity.CheckRecord{
				{FinalVerdict: vo.VerdictWarn},
				{FinalVerdict: vo.VerdictBlock},
				{FinalVerdict: vo.VerdictPass},
			},
			want: vo.VerdictBlock,
		},
		{
			name: "multiple BLOCKs → BLOCK",
			records: []entity.CheckRecord{
				{FinalVerdict: vo.VerdictBlock},
				{FinalVerdict: vo.VerdictBlock},
			},
			want: vo.VerdictBlock,
		},
		{
			name: "BLOCK does not short-circuit — all records evaluated",
			records: []entity.CheckRecord{
				{FinalVerdict: vo.VerdictBlock},
				{FinalVerdict: vo.VerdictPass}, // still counted
				{FinalVerdict: vo.VerdictPass}, // still counted
			},
			want: vo.VerdictBlock,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := engine.AggregateVerdict(tc.records)
			if got != tc.want {
				t.Errorf("AggregateVerdict() = %v, want %v", got, tc.want)
			}
		})
	}
}
