package query

import (
	"sort"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/domain"
)

func sortTasksByPriority(tasks []domain.DashboardTask) {
	order := map[domain.Priority]int{
		domain.PriorityHigh:   0,
		domain.PriorityMedium: 1,
		domain.PriorityLow:    2,
		domain.PriorityInfo:   3,
	}
	sort.SliceStable(tasks, func(i, j int) bool {
		pi := order[tasks[i].Priority]
		pj := order[tasks[j].Priority]
		if pi != pj {
			return pi < pj
		}
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})
}

func buildSummary(tasks []domain.DashboardTask) domain.TaskSummary {
	s := domain.TaskSummary{
		Total:      len(tasks),
		ByModule:   make(map[string]int),
		ByPriority: make(map[string]int),
	}
	for _, t := range tasks {
		s.ByModule[t.Module]++
		s.ByPriority[string(t.Priority)]++
		if t.Priority == domain.PriorityHigh {
			s.HighPriority++
		}
	}
	return s
}

func workflowTaskLabel(state string) string {
	switch state {
	case "NOT_STARTED":
		return "Day not opened"
	case "DAY_OPEN":
		return "Awaiting manager approval"
	case "MANAGER_APPROVED":
		return "Awaiting transaction close"
	case "TRANSACTION_CLOSED":
		return "Awaiting accounting close"
	default:
		return state
	}
}

func workflowTaskPriority(state string) domain.Priority {
	switch state {
	case "NOT_STARTED":
		return domain.PriorityMedium
	case "DAY_OPEN":
		return domain.PriorityHigh
	case "MANAGER_APPROVED":
		return domain.PriorityMedium
	case "TRANSACTION_CLOSED":
		return domain.PriorityMedium
	default:
		return domain.PriorityInfo
	}
}

func compliancePriority(severity string) domain.Priority {
	switch severity {
	case "BLOCK", "REQUIRE_APPROVAL":
		return domain.PriorityHigh
	case "WARN":
		return domain.PriorityMedium
	default:
		return domain.PriorityInfo
	}
}
