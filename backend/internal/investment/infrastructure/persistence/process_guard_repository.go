package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PostgresProcessGuardRepository implements the read-side queries required by
// CanExecuteInvestmentProcess.
type PostgresProcessGuardRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProcessGuardRepository(pool *pgxpool.Pool) *PostgresProcessGuardRepository {
	return &PostgresProcessGuardRepository{pool: pool}
}

// GetActiveDaySetting resolves contract-specific settings first, then GLOBAL.
func (r *PostgresProcessGuardRepository) GetActiveDaySetting(
	ctx context.Context,
	contractID uuid.UUID,
	businessDate time.Time,
) (*entity.WorkflowDaySetting, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			id, name, timezone,
			work_start_time::text, work_end_time::text,
			requires_manager_approval,
			allow_high_level_override,
			block_on_rejection,
			skip_non_business_days,
			effective_from,
			effective_to
		FROM workflow__day_settings
		WHERE is_active = true
		  AND effective_from <= $2
		  AND (effective_to IS NULL OR effective_to >= $2)
		  AND (
		      (scope_type = 'CONTRACT' AND scope_id = $1)
		      OR (scope_type = 'GLOBAL' AND scope_id IS NULL)
		  )
		ORDER BY
			CASE scope_type WHEN 'CONTRACT' THEN 0 ELSE 100 END,
			updated_at DESC
		LIMIT 1`,
		contractID, businessDate,
	)

	setting, err := scanWorkflowDaySetting(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying workflow day setting: %w", err)
	}
	return setting, nil
}

func (r *PostgresProcessGuardRepository) FindBlockingControlDecision(
	ctx context.Context,
	contractID uuid.UUID,
	businessDate time.Time,
	processStep vo.ProcessStepKey,
) (*entity.BlockingControlDecision, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			id, workflow_day_id, contract_id, business_date,
			decision_scope, process_step, decision_level, decision_status,
			reason, decided_by, decided_at
		FROM workflow__control_decisions
		WHERE contract_id = $1
		  AND business_date = $2
		  AND blocks_work = true
		  AND decision_status = 'REJECTED'
		  AND superseded_at IS NULL
		  AND (
		      decision_scope IN ('DAY_START', 'INTRADAY', 'DAY_END')
		      OR (decision_scope = 'INVESTMENT_PROCESS' AND process_step = $3)
		  )
		ORDER BY
			CASE decision_scope WHEN 'INVESTMENT_PROCESS' THEN 0 ELSE 10 END,
			decided_at DESC
		LIMIT 1`,
		contractID, businessDate, string(processStep),
	)

	decision, err := scanBlockingControlDecision(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying blocking control decision: %w", err)
	}
	return decision, nil
}

func (r *PostgresProcessGuardRepository) FindUserProcessAssignment(
	ctx context.Context,
	userID uuid.UUID,
	contractID uuid.UUID,
	businessDate time.Time,
	processStep vo.ProcessStepKey,
) (*entity.ProcessAssignmentMatch, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			a.id,
			a.assignment_type,
			s.step_key,
			a.process_group_id,
			COALESCE(g.name, ''),
			$1::uuid AS user_id,
			a.scope_type,
			a.scope_id,
			a.can_execute,
			a.priority
		FROM investment__process_step_assignments a
		JOIN investment__process_steps s
			ON s.id = a.process_step_id
		LEFT JOIN investment__process_groups g
			ON g.id = a.process_group_id
		LEFT JOIN investment__process_group_members gm
			ON gm.group_id = a.process_group_id
			AND gm.user_id = $1
			AND gm.is_active = true
			AND gm.effective_from <= $3
			AND (gm.effective_to IS NULL OR gm.effective_to >= $3)
		WHERE s.step_key = $4
		  AND s.is_active = true
		  AND a.is_active = true
		  AND a.effective_from <= $3
		  AND (a.effective_to IS NULL OR a.effective_to >= $3)
		  AND (
		      (a.scope_type = 'CONTRACT' AND a.scope_id = $2)
		      OR (a.scope_type = 'GLOBAL' AND a.scope_id IS NULL)
		  )
		  AND (
		      (a.assignment_type = 'USER' AND a.user_id = $1)
		      OR (
		          a.assignment_type = 'GROUP'
		          AND gm.user_id IS NOT NULL
		          AND g.is_active = true
		          AND g.effective_from <= $3
		          AND (g.effective_to IS NULL OR g.effective_to >= $3)
		      )
		  )
		ORDER BY
			CASE a.scope_type WHEN 'CONTRACT' THEN 0 ELSE 100 END,
			a.priority ASC,
			CASE a.assignment_type WHEN 'USER' THEN 0 ELSE 10 END,
			a.created_at DESC
		LIMIT 1`,
		userID, contractID, businessDate, string(processStep),
	)

	assignment, err := scanProcessAssignmentMatch(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying process assignment: %w", err)
	}
	return assignment, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanWorkflowDaySetting(s scanner) (*entity.WorkflowDaySetting, error) {
	var setting entity.WorkflowDaySetting
	var startText, endText string

	if err := s.Scan(
		&setting.ID,
		&setting.Name,
		&setting.Timezone,
		&startText,
		&endText,
		&setting.RequiresManagerApproval,
		&setting.AllowHighLevelOverride,
		&setting.BlockOnRejection,
		&setting.SkipNonBusinessDays,
		&setting.EffectiveFrom,
		&setting.EffectiveTo,
	); err != nil {
		return nil, err
	}

	start, err := parsePostgresTimeOfDay(startText)
	if err != nil {
		return nil, fmt.Errorf("parsing work_start_time: %w", err)
	}
	end, err := parsePostgresTimeOfDay(endText)
	if err != nil {
		return nil, fmt.Errorf("parsing work_end_time: %w", err)
	}
	setting.WorkStartTime = start
	setting.WorkEndTime = end
	return &setting, nil
}

func scanBlockingControlDecision(s scanner) (*entity.BlockingControlDecision, error) {
	var d entity.BlockingControlDecision
	var processStep *string
	var reason *string

	if err := s.Scan(
		&d.ID,
		&d.WorkflowDayID,
		&d.ContractID,
		&d.BusinessDate,
		&d.DecisionScope,
		&processStep,
		&d.DecisionLevel,
		&d.DecisionStatus,
		&reason,
		&d.DecidedBy,
		&d.DecidedAt,
	); err != nil {
		return nil, err
	}

	if processStep != nil {
		step := vo.ProcessStepKey(*processStep)
		d.ProcessStep = &step
	}
	if reason != nil {
		d.Reason = *reason
	}
	d.DecidedAt = d.DecidedAt.UTC()
	return &d, nil
}

func scanProcessAssignmentMatch(s scanner) (*entity.ProcessAssignmentMatch, error) {
	var a entity.ProcessAssignmentMatch
	var step string

	if err := s.Scan(
		&a.AssignmentID,
		&a.AssignmentType,
		&step,
		&a.GroupID,
		&a.GroupName,
		&a.UserID,
		&a.ScopeType,
		&a.ScopeID,
		&a.CanExecute,
		&a.Priority,
	); err != nil {
		return nil, err
	}

	a.ProcessStep = vo.ProcessStepKey(step)
	return &a, nil
}

func parsePostgresTimeOfDay(raw string) (time.Duration, error) {
	layouts := []string{"15:04:05.999999", "15:04:05", "15:04"}
	var parsed time.Time
	var err error
	for _, layout := range layouts {
		parsed, err = time.Parse(layout, raw)
		if err == nil {
			return time.Duration(parsed.Hour())*time.Hour +
				time.Duration(parsed.Minute())*time.Minute +
				time.Duration(parsed.Second())*time.Second +
				time.Duration(parsed.Nanosecond()), nil
		}
	}
	return 0, err
}

var _ domain.InvestmentProcessGuardRepository = (*PostgresProcessGuardRepository)(nil)
