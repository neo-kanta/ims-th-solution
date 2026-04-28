package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
)

// PostgresSchedulerRepository implements domain.SchedulerRepository.
type PostgresSchedulerRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresSchedulerRepository creates the scheduler repository.
func NewPostgresSchedulerRepository(pool *pgxpool.Pool) *PostgresSchedulerRepository {
	return &PostgresSchedulerRepository{pool: pool}
}

// TryAcquireSchedulerLock attempts a PostgreSQL advisory lock and keeps the
// owning connection checked out until Release is called.
func (r *PostgresSchedulerRepository) TryAcquireSchedulerLock(
	ctx context.Context,
	lockKey int64,
) (domain.SchedulerLock, bool, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("acquiring scheduler lock connection: %w", err)
	}

	var acquired bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, lockKey).Scan(&acquired); err != nil {
		conn.Release()
		return nil, false, fmt.Errorf("acquiring scheduler advisory lock: %w", err)
	}
	if !acquired {
		conn.Release()
		return nil, false, nil
	}
	return &postgresSchedulerLock{conn: conn, lockKey: lockKey}, true, nil
}

// ListActiveScheduleRules returns enabled schedule rules for businessDate.
func (r *PostgresSchedulerRepository) ListActiveScheduleRules(
	ctx context.Context,
	businessDate time.Time,
) ([]*entity.ScheduleRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			r.id,
			r.day_setting_id,
			r.name,
			r.action,
			r.trigger_time_local::text,
			r.timezone,
			COALESCE(array_to_string(r.days_of_week, ','), ''),
			r.skip_holidays,
			r.priority,
			r.effective_from::text,
			COALESCE(r.effective_to::text, '')
		FROM workflow__schedule_rules r
		JOIN workflow__day_settings s ON s.id = r.day_setting_id
		WHERE r.is_enabled = true
		  AND s.is_active = true
		  AND r.effective_from <= $1
		  AND (r.effective_to IS NULL OR r.effective_to >= $1)
		  AND s.effective_from <= $1
		  AND (s.effective_to IS NULL OR s.effective_to >= $1)
		  AND (
			(r.action = 'OPEN_DAY' AND s.auto_start_enabled = true)
			OR (r.action = 'END_DAY' AND s.auto_end_enabled = true)
		  )
		ORDER BY r.priority ASC, r.trigger_time_local ASC, r.name ASC`,
		truncateDateUTC(businessDate),
	)
	if err != nil {
		return nil, fmt.Errorf("listing active workflow schedule rules: %w", err)
	}
	defer rows.Close()

	var rules []*entity.ScheduleRule
	for rows.Next() {
		rule, err := scanScheduleRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

// CreateSchedulerRun inserts a RUNNING audit header.
func (r *PostgresSchedulerRepository) CreateSchedulerRun(ctx context.Context, run *entity.SchedulerRun) error {
	summary, err := json.Marshal(run.Summary)
	if err != nil {
		return fmt.Errorf("marshalling scheduler run summary: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO workflow__scheduler_runs (
			id, scheduler_name, rule_id, action,
			business_date, timezone, started_at, status,
			triggered_by, locked_by, summary, error_message, created_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11, $12, $13
		)`,
		run.ID, run.SchedulerName, run.RuleID, schedulerActionValue(run.Action),
		truncateDateUTC(run.BusinessDate), run.Timezone, run.StartedAt, string(run.Status),
		run.TriggeredBy, run.LockedBy, summary, run.ErrorMessage, run.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating workflow scheduler run: %w", err)
	}
	return nil
}

// FinishSchedulerRun updates an audit header with its terminal result.
func (r *PostgresSchedulerRepository) FinishSchedulerRun(ctx context.Context, run *entity.SchedulerRun) error {
	summary, err := json.Marshal(run.Summary)
	if err != nil {
		return fmt.Errorf("marshalling scheduler run summary: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE workflow__scheduler_runs
		SET
			finished_at = $2,
			status = $3,
			summary = $4,
			error_message = $5
		WHERE id = $1`,
		run.ID, run.FinishedAt, string(run.Status), summary, run.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("finishing workflow scheduler run: %w", err)
	}
	return nil
}

// InsertSchedulerRunItem inserts one per-contract audit detail row.
func (r *PostgresSchedulerRepository) InsertSchedulerRunItem(ctx context.Context, item *entity.SchedulerRunItem) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO workflow__scheduler_run_items (
			id, scheduler_run_id, rule_id,
			contract_id, business_date, action, status,
			skip_reason, error_message,
			workflow_day_id, transition_id, created_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7,
			$8, $9,
			$10, $11, $12
		)`,
		item.ID, item.SchedulerRunID, item.RuleID,
		item.ContractID, truncateDateUTC(item.BusinessDate), string(item.Action), string(item.Status),
		item.SkipReason, item.ErrorMessage,
		item.WorkflowDayID, item.TransitionID, item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("creating workflow scheduler run item: %w", err)
	}
	return nil
}

func scanScheduleRule(rows pgx.Rows) (*entity.ScheduleRule, error) {
	var rule entity.ScheduleRule
	var action string
	var triggerTime string
	var daysOfWeek string
	var effectiveFrom string
	var effectiveTo string

	err := rows.Scan(
		&rule.ID,
		&rule.DaySettingID,
		&rule.Name,
		&action,
		&triggerTime,
		&rule.Timezone,
		&daysOfWeek,
		&rule.SkipHolidays,
		&rule.Priority,
		&effectiveFrom,
		&effectiveTo,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning workflow schedule rule: %w", err)
	}

	trigger, err := parsePostgresTimeOfDay(triggerTime)
	if err != nil {
		return nil, fmt.Errorf("parsing schedule rule trigger time %q: %w", triggerTime, err)
	}
	from, err := time.Parse("2006-01-02", effectiveFrom)
	if err != nil {
		return nil, fmt.Errorf("parsing schedule rule effective_from %q: %w", effectiveFrom, err)
	}

	rule.Action = entity.SchedulerAction(action)
	rule.TriggerTimeLocal = trigger
	rule.DaysOfWeek = parseDaysOfWeek(daysOfWeek)
	rule.EffectiveFrom = from
	if effectiveTo != "" {
		to, err := time.Parse("2006-01-02", effectiveTo)
		if err != nil {
			return nil, fmt.Errorf("parsing schedule rule effective_to %q: %w", effectiveTo, err)
		}
		rule.EffectiveTo = &to
	}
	return &rule, nil
}

func parsePostgresTimeOfDay(value string) (time.Duration, error) {
	parts := strings.Split(value, ":")
	if len(parts) < 2 {
		return 0, fmt.Errorf("expected HH:MM[:SS]")
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	second := 0
	if len(parts) >= 3 {
		secPart := strings.Split(parts[2], ".")[0]
		second, err = strconv.Atoi(secPart)
		if err != nil {
			return 0, err
		}
	}
	return time.Duration(hour)*time.Hour +
		time.Duration(minute)*time.Minute +
		time.Duration(second)*time.Second, nil
}

func parseDaysOfWeek(value string) []int {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	days := make([]int, 0, len(parts))
	for _, part := range parts {
		day, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil {
			days = append(days, day)
		}
	}
	return days
}

func truncateDateUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

var _ domain.SchedulerRepository = (*PostgresSchedulerRepository)(nil)

type postgresSchedulerLock struct {
	conn    *pgxpool.Conn
	lockKey int64
}

func (l *postgresSchedulerLock) Release(context.Context) error {
	defer l.conn.Release()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var unlocked bool
	if err := l.conn.QueryRow(ctx, `SELECT pg_advisory_unlock($1)`, l.lockKey).Scan(&unlocked); err != nil {
		return fmt.Errorf("releasing scheduler advisory lock: %w", err)
	}
	if !unlocked {
		return fmt.Errorf("scheduler advisory lock %d was not held", l.lockKey)
	}
	return nil
}

func schedulerActionValue(action *entity.SchedulerAction) any {
	if action == nil {
		return nil
	}
	return string(*action)
}

// PostgresWorkflowSchedulerContractSource reads active contracts from the
// temporary workflow__scheduler_contracts bridge table. Replace this adapter
// with the real contract/fund master source once that module exists.
type PostgresWorkflowSchedulerContractSource struct {
	pool *pgxpool.Pool
}

// NewPostgresWorkflowSchedulerContractSource creates the temporary bridge source.
func NewPostgresWorkflowSchedulerContractSource(pool *pgxpool.Pool) *PostgresWorkflowSchedulerContractSource {
	return &PostgresWorkflowSchedulerContractSource{pool: pool}
}

// ListActiveContracts returns active contracts effective for businessDate.
func (s *PostgresWorkflowSchedulerContractSource) ListActiveContracts(
	ctx context.Context,
	businessDate time.Time,
) ([]uuid.UUID, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT contract_id, is_active, effective_from::text, COALESCE(effective_to::text, '')
		FROM workflow__scheduler_contracts
		ORDER BY contract_id`)
	if err != nil {
		return nil, fmt.Errorf("listing scheduler contract bridge records: %w", err)
	}
	defer rows.Close()

	var records []schedulerContractRecord
	for rows.Next() {
		record, err := scanSchedulerContractRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return filterActiveSchedulerContractIDs(records, truncateDateUTC(businessDate)), nil
}

type schedulerContractRecord struct {
	ContractID    uuid.UUID
	IsActive      bool
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}

func scanSchedulerContractRecord(rows pgx.Rows) (schedulerContractRecord, error) {
	var record schedulerContractRecord
	var effectiveFrom string
	var effectiveTo string

	err := rows.Scan(&record.ContractID, &record.IsActive, &effectiveFrom, &effectiveTo)
	if err != nil {
		return schedulerContractRecord{}, fmt.Errorf("scanning scheduler contract bridge record: %w", err)
	}
	from, err := time.Parse("2006-01-02", effectiveFrom)
	if err != nil {
		return schedulerContractRecord{}, fmt.Errorf("parsing scheduler contract effective_from %q: %w", effectiveFrom, err)
	}
	record.EffectiveFrom = from
	if effectiveTo != "" {
		to, err := time.Parse("2006-01-02", effectiveTo)
		if err != nil {
			return schedulerContractRecord{}, fmt.Errorf("parsing scheduler contract effective_to %q: %w", effectiveTo, err)
		}
		record.EffectiveTo = &to
	}
	return record, nil
}

func filterActiveSchedulerContractIDs(records []schedulerContractRecord, businessDate time.Time) []uuid.UUID {
	contractIDs := make([]uuid.UUID, 0, len(records))
	for _, record := range records {
		if isSchedulerContractActive(record, businessDate) {
			contractIDs = append(contractIDs, record.ContractID)
		}
	}
	return contractIDs
}

func isSchedulerContractActive(record schedulerContractRecord, businessDate time.Time) bool {
	date := truncateDateUTC(businessDate)
	if !record.IsActive {
		return false
	}
	if truncateDateUTC(record.EffectiveFrom).After(date) {
		return false
	}
	if record.EffectiveTo != nil && truncateDateUTC(*record.EffectiveTo).Before(date) {
		return false
	}
	return true
}
