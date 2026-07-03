package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/domain"
)

// TaskRepository implements query.TaskRepository using direct SQL against the
// shared pgxpool. It intentionally queries tables owned by other modules (no
// Go-level imports of those modules — the shared PostgreSQL instance is the
// legitimate boundary for this cross-module read model).
type TaskRepository struct {
	pool *pgxpool.Pool
}

// NewTaskRepository creates a new TaskRepository.
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

// Compile-time assertion that TaskRepository satisfies the port.
var _ query.TaskRepository = (*TaskRepository)(nil)

// parseUUIDs converts string UUIDs to uuid.UUID slice; invalid entries are skipped.
func parseUUIDs(ids []string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(ids))
	for _, s := range ids {
		if id, err := uuid.Parse(s); err == nil {
			out = append(out, id)
		}
	}
	return out
}

// FetchResearchTasks returns SUBMITTED research reports visible to the caller.
// Reports with a NULL applicable_contract_id are shown to any user with
// INTEGRATION_DASHBOARD_VIEW (global research).
func (r *TaskRepository) FetchResearchTasks(ctx context.Context, contractIDs []string) ([]domain.DashboardTask, error) {
	ids := parseUUIDs(contractIDs)

	rows, err := r.pool.Query(ctx, `
		SELECT
			id,
			report_no,
			report_title,
			instrument_code,
			applicable_contract_id,
			owner_user_id,
			report_date,
			created_at,
			updated_at
		FROM investment__research_reports
		WHERE review_status = 'SUBMITTED'
		  AND deleted_at IS NULL
		  AND (applicable_contract_id IS NULL OR applicable_contract_id = ANY($1::uuid[]))
		ORDER BY updated_at DESC
		LIMIT 50
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("fetching research tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.DashboardTask
	for rows.Next() {
		var (
			id                   uuid.UUID
			reportNo             string
			reportTitle          string
			instrumentCode       string
			applicableContractID *uuid.UUID
			ownerUserID          uuid.UUID
			reportDate           time.Time
			createdAt            time.Time
			updatedAt            time.Time
		)
		if err := rows.Scan(
			&id,
			&reportNo,
			&reportTitle,
			&instrumentCode,
			&applicableContractID,
			&ownerUserID,
			&reportDate,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning research task row: %w", err)
		}

		contractID := ""
		if applicableContractID != nil {
			contractID = applicableContractID.String()
		}

		desc := fmt.Sprintf("Research report for %s submitted for review.", instrumentCode)
		actionURL := fmt.Sprintf("/investment/research/%s", id.String())

		tasks = append(tasks, domain.DashboardTask{
			TaskID:         "rr-" + id.String(),
			Type:           domain.TaskTypeResearchReview,
			Module:         "investment",
			Title:          fmt.Sprintf("Review: %s", reportTitle),
			Description:    desc,
			Priority:       domain.PriorityMedium,
			Status:         domain.TaskStatusPending,
			BusinessDate:   reportDate.Format("2006-01-02"),
			SourceRecordID: id.String(),
			SourceType:     "research_report",
			ActionURL:      actionURL,
			Reason:         reportNo,
			Subject:        reportNo,
			Severity:       "",
			CanAct:         true,
			AllowedActions: []string{"review"},
			ContractID:     contractID,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating research task rows: %w", err)
	}

	return tasks, nil
}

// FetchWorkflowTasks returns pending workflow day-state tasks for the caller's
// accessible contracts. ACCOUNTING_CLOSED states are excluded — they require
// no user action.
func (r *TaskRepository) FetchWorkflowTasks(ctx context.Context, contractIDs []string) ([]domain.DashboardTask, error) {
	ids := parseUUIDs(contractIDs)
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (contract_id)
			contract_id,
			business_date,
			current_state,
			updated_at
		FROM workflow__day_states
		WHERE contract_id = ANY($1::uuid[])
		  AND current_state NOT IN ('ACCOUNTING_CLOSED')
		ORDER BY contract_id, business_date DESC
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("fetching workflow tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.DashboardTask
	for rows.Next() {
		var (
			contractID   uuid.UUID
			businessDate time.Time
			currentState string
			updatedAt    time.Time
		)
		if err := rows.Scan(&contractID, &businessDate, &currentState, &updatedAt); err != nil {
			return nil, fmt.Errorf("scanning workflow task row: %w", err)
		}

		dateStr := businessDate.Format("2006-01-02")
		label := workflowTaskLabel(currentState)
		priority := workflowTaskPriority(currentState)
		actionURL := fmt.Sprintf("/workflow/contracts/%s", contractID.String())

		tasks = append(tasks, domain.DashboardTask{
			TaskID:         "wf-" + contractID.String() + "-" + dateStr,
			Type:           domain.TaskTypeWorkflowPending,
			Module:         "workflow",
			Title:          label,
			Description:    fmt.Sprintf("Contract %s workflow state: %s for %s.", contractID.String(), currentState, dateStr),
			Priority:       priority,
			Status:         domain.TaskStatusPending,
			BusinessDate:   dateStr,
			SourceRecordID: contractID.String(),
			SourceType:     "workflow_day_state",
			ActionURL:      actionURL,
			Reason:         currentState,
			Subject:        currentState,
			Severity:       "",
			CanAct:         true,
			AllowedActions: []string{"advance"},
			ContractID:     contractID.String(),
			CreatedAt:      updatedAt,
			UpdatedAt:      updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating workflow task rows: %w", err)
	}

	return tasks, nil
}

// FetchWorkflowStates returns the latest workflow state row per accessible contract.
func (r *TaskRepository) FetchWorkflowStates(ctx context.Context, contractIDs []string) ([]domain.WorkflowStateRow, error) {
	ids := parseUUIDs(contractIDs)
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (contract_id)
			contract_id,
			business_date,
			current_state,
			updated_at
		FROM workflow__day_states
		WHERE contract_id = ANY($1::uuid[])
		ORDER BY contract_id, business_date DESC
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("fetching workflow states: %w", err)
	}
	defer rows.Close()

	var states []domain.WorkflowStateRow
	for rows.Next() {
		var (
			contractID   uuid.UUID
			businessDate time.Time
			currentState string
			updatedAt    time.Time
		)
		if err := rows.Scan(&contractID, &businessDate, &currentState, &updatedAt); err != nil {
			return nil, fmt.Errorf("scanning workflow state row: %w", err)
		}
		states = append(states, domain.WorkflowStateRow{
			ContractID:   contractID.String(),
			BusinessDate: businessDate.Format("2006-01-02"),
			CurrentState: currentState,
			UpdatedAt:    updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating workflow state rows: %w", err)
	}

	return states, nil
}

// FetchComplianceTasks returns open compliance breaches for the caller's
// accessible contracts, ordered by severity then recency.
func (r *TaskRepository) FetchComplianceTasks(ctx context.Context, contractIDs []string) ([]domain.DashboardTask, error) {
	ids := parseUUIDs(contractIDs)
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			b.id,
			b.contract_id,
			b.rule_type_id,
			b.severity,
			b.verdict,
			b.message,
			b.business_date,
			cr.timing,
			b.created_at,
			b.updated_at
		FROM compliance_breaches b
		JOIN compliance_check_records cr ON cr.id = b.check_record_id
		WHERE b.status = 'OPEN'
		  AND b.contract_id = ANY($1::uuid[])
		ORDER BY
			CASE b.severity
				WHEN 'BLOCK' THEN 1
				WHEN 'REQUIRE_APPROVAL' THEN 2
				WHEN 'WARN' THEN 3
				ELSE 4
			END,
			b.created_at DESC
		LIMIT 50
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("fetching compliance tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.DashboardTask
	for rows.Next() {
		var (
			id           uuid.UUID
			contractID   uuid.UUID
			ruleTypeID   string
			severity     string
			verdict      string
			message      *string
			businessDate time.Time
			timing       string
			createdAt    time.Time
			updatedAt    time.Time
		)
		if err := rows.Scan(
			&id,
			&contractID,
			&ruleTypeID,
			&severity,
			&verdict,
			&message,
			&businessDate,
			&timing,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning compliance task row: %w", err)
		}

		desc := fmt.Sprintf("%s breach (%s): %s — %s check.", severity, verdict, ruleTypeID, timing)
		if message != nil && *message != "" {
			desc = *message
		}
		dateStr := businessDate.Format("2006-01-02")
		actionURL := fmt.Sprintf("/compliance/breaches/%s", id.String())

		tasks = append(tasks, domain.DashboardTask{
			TaskID:         "cb-" + id.String(),
			Type:           domain.TaskTypeComplianceBreach,
			Module:         "compliance",
			Title:          fmt.Sprintf("Compliance breach: %s (%s)", ruleTypeID, severity),
			Description:    desc,
			Priority:       compliancePriority(severity),
			Status:         domain.TaskStatusPending,
			BusinessDate:   dateStr,
			SourceRecordID: id.String(),
			SourceType:     "compliance_breach",
			ActionURL:      actionURL,
			Reason:         fmt.Sprintf("%s / %s", severity, verdict),
			Subject:        ruleTypeID,
			Severity:       severity,
			CanAct:         true,
			AllowedActions: []string{"resolve", "override"},
			ContractID:     contractID.String(),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating compliance task rows: %w", err)
	}

	return tasks, nil
}

// workflowTaskLabel maps a workflow state code to a human-readable label.
// Kept here (not in the query layer) so the persistence layer stays self-contained.
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

// workflowTaskPriority maps a workflow state to its task priority.
func workflowTaskPriority(state string) domain.Priority {
	switch state {
	case "DAY_OPEN":
		return domain.PriorityHigh
	case "NOT_STARTED", "MANAGER_APPROVED", "TRANSACTION_CLOSED":
		return domain.PriorityMedium
	default:
		return domain.PriorityInfo
	}
}

// compliancePriority maps a breach severity code to a task priority.
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
