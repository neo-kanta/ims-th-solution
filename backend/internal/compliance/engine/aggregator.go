package engine

import (
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
)

// AggregateVerdict returns the dominating verdict from all records.
// BLOCK > WARN > PASS. All rules are evaluated — no short-circuit.
func AggregateVerdict(records []entity.CheckRecord) vo.Verdict {
	result := vo.VerdictPass
	for i := range records {
		if records[i].FinalVerdict.Dominates(result) {
			result = records[i].FinalVerdict
		}
	}
	return result
}
