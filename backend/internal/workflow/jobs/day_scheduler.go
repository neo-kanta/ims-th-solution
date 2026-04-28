package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

const (
	defaultSchedulerName     = "workflow-day-scheduler"
	defaultSchedulerLockKey  = int64(2026042708301730)
	defaultSchedulerTimezone = "Asia/Bangkok"
)

var defaultSystemActorID = uuid.MustParse("a0000000-0000-0000-0000-000000000001")

// WorkflowDayReader is the narrow workflow state read model the scheduler needs.
type WorkflowDayReader interface {
	GetByContractDate(ctx context.Context, contractID uuid.UUID, businessDate time.Time) (*entity.WorkflowDay, error)
}

// OpenDayCommand is implemented by command.OpenDayHandler.
type OpenDayCommand interface {
	Handle(ctx context.Context, req command.OpenDayRequest) (*command.OpenDayResult, error)
}

// DaySchedulerConfig controls operational scheduler identity.
type DaySchedulerConfig struct {
	SchedulerName string
	SystemActorID uuid.UUID
	LockedBy      string
	LockKey       int64
	Timezone      string
}

// DayScheduler evaluates workflow__schedule_rules and writes audit rows.
type DayScheduler struct {
	schedulerRepo domain.SchedulerRepository
	dayRepo       WorkflowDayReader
	contractSrc   ports.ContractSourcePort
	calendar      ports.HolidayCalendarPort
	openDay       OpenDayCommand
	clock         clock.Clock
	config        DaySchedulerConfig
}

// NewDayScheduler creates the workflow day scheduler.
func NewDayScheduler(
	schedulerRepo domain.SchedulerRepository,
	dayRepo WorkflowDayReader,
	contractSrc ports.ContractSourcePort,
	calendar ports.HolidayCalendarPort,
	openDay OpenDayCommand,
	clk clock.Clock,
	cfg DaySchedulerConfig,
) *DayScheduler {
	if cfg.SchedulerName == "" {
		cfg.SchedulerName = defaultSchedulerName
	}
	if cfg.SystemActorID == uuid.Nil {
		cfg.SystemActorID = defaultSystemActorID
	}
	if cfg.LockedBy == "" {
		cfg.LockedBy = cfg.SchedulerName
	}
	if cfg.LockKey == 0 {
		cfg.LockKey = defaultSchedulerLockKey
	}
	if cfg.Timezone == "" {
		cfg.Timezone = defaultSchedulerTimezone
	}
	if clk == nil {
		clk = clock.RealClock{}
	}
	if contractSrc == nil {
		contractSrc = emptyContractSource{}
	}
	return &DayScheduler{
		schedulerRepo: schedulerRepo,
		dayRepo:       dayRepo,
		contractSrc:   contractSrc,
		calendar:      calendar,
		openDay:       openDay,
		clock:         clk,
		config:        cfg,
	}
}

// RunOnce executes one scheduler tick and writes one tick-level audit run.
func (s *DayScheduler) RunOnce(ctx context.Context) ([]*entity.SchedulerRun, error) {
	if s == nil {
		return nil, fmt.Errorf("workflow day scheduler is nil")
	}
	if s.schedulerRepo == nil {
		return nil, fmt.Errorf("workflow scheduler repository is required")
	}

	now := s.clock.Now().UTC()
	loc, err := loadRuleLocation(s.config.Timezone)
	if err != nil {
		return nil, err
	}
	businessDate := businessDateInLocation(now, loc)

	lock, acquired, err := s.schedulerRepo.TryAcquireSchedulerLock(ctx, s.config.LockKey)
	if err != nil {
		run, auditErr := s.recordStandaloneTick(ctx, now, businessDate, entity.SchedulerRunStatusFailed, map[string]any{
			"reason": "LOCK_ACQUIRE_FAILED",
		}, err)
		if auditErr != nil {
			return nil, fmt.Errorf("%w; additionally failed to audit scheduler lock error: %v", err, auditErr)
		}
		return []*entity.SchedulerRun{run}, err
	}
	if !acquired {
		slog.Info("workflow scheduler tick skipped: advisory lock not acquired",
			"scheduler", s.config.SchedulerName,
			"business_date", businessDate.Format("2006-01-02"),
		)
		run, err := s.recordStandaloneTick(ctx, now, businessDate, entity.SchedulerRunStatusSkipped, map[string]any{
			"reason": "SCHEDULER_LOCK_NOT_ACQUIRED",
		}, nil)
		if err != nil {
			return nil, err
		}
		return []*entity.SchedulerRun{run}, nil
	}
	defer func() {
		if err := lock.Release(ctx); err != nil {
			slog.Error("workflow scheduler advisory lock release failed", "error", err)
		}
	}()

	run, err := s.runTick(ctx, now, businessDate)
	if err != nil {
		if run == nil {
			return nil, err
		}
		return []*entity.SchedulerRun{run}, err
	}
	return []*entity.SchedulerRun{run}, nil
}

// Start runs the scheduler immediately and then at interval until ctx is done.
func (s *DayScheduler) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}

	go func() {
		s.logRun(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.logRun(ctx)
			}
		}
	}()
}

func (s *DayScheduler) logRun(ctx context.Context) {
	runs, err := s.RunOnce(ctx)
	if err != nil {
		slog.Error("workflow scheduler run failed", "error", err)
		return
	}
	slog.Info("workflow scheduler run completed", "run_count", len(runs))
}

func (s *DayScheduler) runTick(
	ctx context.Context,
	now time.Time,
	businessDate time.Time,
) (*entity.SchedulerRun, error) {
	run := entity.NewSchedulerTickRun(s.config.SchedulerName, businessDate, s.config.Timezone, now, s.config.LockedBy)
	if err := s.schedulerRepo.CreateSchedulerRun(ctx, run); err != nil {
		return nil, err
	}

	rules, err := s.schedulerRepo.ListActiveScheduleRules(ctx, businessDate)
	if err != nil {
		msg := err.Error()
		finishErr := s.finishRun(ctx, run, entity.SchedulerRunStatusFailed, map[string]any{
			"reason": "LIST_RULES_FAILED",
		}, &msg)
		if finishErr != nil {
			return run, finishErr
		}
		return run, err
	}
	if len(rules) == 0 {
		return run, s.finishRun(ctx, run, entity.SchedulerRunStatusSkipped, map[string]any{
			"reason":     "NO_ACTIVE_RULES",
			"rule_count": 0,
			"checked_at": now.Format(time.RFC3339),
			"timezone":   s.config.Timezone,
		}, nil)
	}

	summary := newSchedulerTickSummary(len(rules), now, s.config.Timezone)
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if err := s.evaluateRule(ctx, run, rule, now, summary); err != nil {
			msg := err.Error()
			finishErr := s.finishRun(ctx, run, entity.SchedulerRunStatusFailed, summary.mapSummary(), &msg)
			if finishErr != nil {
				return run, finishErr
			}
			return run, err
		}
	}

	return run, s.finishRun(ctx, run, summary.status(), summary.mapSummary(), nil)
}

func (s *DayScheduler) evaluateRule(
	ctx context.Context,
	run *entity.SchedulerRun,
	rule *entity.ScheduleRule,
	now time.Time,
	summary *schedulerTickSummary,
) error {
	loc, err := loadRuleLocation(rule.Timezone)
	if err != nil {
		summary.addRuleResult(rule, entity.SchedulerRunStatusFailed, "INVALID_TIMEZONE", nil)
		summary.addRuleStatus(entity.SchedulerRunStatusFailed)
		return nil
	}
	businessDate := businessDateInLocation(now, loc)

	due, skipReason := isRuleDue(rule, now.In(loc))
	if !due {
		summary.addRuleResult(rule, entity.SchedulerRunStatusSkipped, skipReason, map[string]any{
			"trigger_time_local": formatDurationClock(rule.TriggerTimeLocal),
			"checked_at_local":   now.In(loc).Format(time.RFC3339),
		})
		summary.addRuleStatus(entity.SchedulerRunStatusSkipped)
		return nil
	}

	if rule.SkipHolidays && s.calendar != nil {
		isBusinessDay, holidayName, err := s.calendar.IsBusinessDay(ctx, businessDate)
		if err != nil {
			summary.addRuleResult(rule, entity.SchedulerRunStatusFailed, "HOLIDAY_CHECK_FAILED", map[string]any{
				"error": err.Error(),
			})
			summary.addRuleStatus(entity.SchedulerRunStatusFailed)
			return nil
		}
		if !isBusinessDay {
			summary.addRuleResult(rule, entity.SchedulerRunStatusSkipped, "NON_BUSINESS_DAY", map[string]any{
				"holiday_name": holidayName,
			})
			summary.addRuleStatus(entity.SchedulerRunStatusSkipped)
			return nil
		}
	}

	contractIDs, err := s.contractSrc.ListActiveContracts(ctx, businessDate)
	if err != nil {
		summary.addRuleResult(rule, entity.SchedulerRunStatusFailed, "CONTRACT_SOURCE_FAILED", map[string]any{
			"error": err.Error(),
		})
		summary.addRuleStatus(entity.SchedulerRunStatusFailed)
		return nil
	}
	if len(contractIDs) == 0 {
		summary.addRuleResult(rule, entity.SchedulerRunStatusSkipped, "NO_ACTIVE_CONTRACTS", map[string]any{
			"contracts_checked": 0,
		})
		summary.addRuleStatus(entity.SchedulerRunStatusSkipped)
		return nil
	}

	ruleCounts := schedulerRunCounts{ContractsChecked: len(contractIDs)}
	for _, contractID := range contractIDs {
		item := entity.NewSchedulerRunItem(run.ID, rule.ID, contractID, businessDate, rule.Action, s.clock.Now())
		switch rule.Action {
		case entity.SchedulerActionOpenDay:
			s.evaluateOpenDay(ctx, run, item)
		case entity.SchedulerActionEndDay:
			s.evaluateEndDay(ctx, item)
		default:
			failItem(item, fmt.Errorf("unsupported scheduler action %s", rule.Action))
		}

		if err := s.schedulerRepo.InsertSchedulerRunItem(ctx, item); err != nil {
			return err
		}
		ruleCounts.add(item.Status)
		summary.addItem(item.Status)
	}

	summary.addRuleResult(rule, ruleCounts.status(), "", ruleCounts.summary())
	return nil
}

func (s *DayScheduler) recordStandaloneTick(
	ctx context.Context,
	now time.Time,
	businessDate time.Time,
	status entity.SchedulerRunStatus,
	summary map[string]any,
	err error,
) (*entity.SchedulerRun, error) {
	run := entity.NewSchedulerTickRun(s.config.SchedulerName, businessDate, s.config.Timezone, now, s.config.LockedBy)
	if err := s.schedulerRepo.CreateSchedulerRun(ctx, run); err != nil {
		return nil, err
	}
	var errMsg *string
	if err != nil {
		msg := err.Error()
		errMsg = &msg
	}
	if finishErr := s.finishRun(ctx, run, status, summary, errMsg); finishErr != nil {
		return run, finishErr
	}
	return run, nil
}

func (s *DayScheduler) evaluateOpenDay(
	ctx context.Context,
	run *entity.SchedulerRun,
	item *entity.SchedulerRunItem,
) {
	if s.dayRepo == nil {
		failItem(item, fmt.Errorf("workflow day repository is required"))
		return
	}
	if s.openDay == nil {
		skipItem(item, "OPEN_DAY_COMMAND_NOT_CONFIGURED")
		return
	}

	day, err := s.dayRepo.GetByContractDate(ctx, item.ContractID, item.BusinessDate)
	if err != nil {
		failItem(item, err)
		return
	}
	if day != nil {
		item.WorkflowDayID = &day.ID
		skipItem(item, "ALREADY_STARTED")
		return
	}

	result, err := s.openDay.Handle(ctx, command.OpenDayRequest{
		ContractID:   item.ContractID,
		BusinessDate: item.BusinessDate,
		Actor: vo.ActorContext{
			UserID:    s.config.SystemActorID,
			Username:  s.config.SchedulerName,
			ActorType: vo.ActorTypeSystem,
			RequestID: fmt.Sprintf("%s/%s/%s", s.config.SchedulerName, run.ID.String(), item.ContractID.String()),
		},
	})
	if err != nil {
		var exists *domain.ErrWorkflowDayExists
		if errors.As(err, &exists) {
			skipItem(item, "ALREADY_STARTED")
			return
		}
		failItem(item, err)
		return
	}

	item.Status = entity.SchedulerItemStatusSuccess
	item.WorkflowDayID = &result.WorkflowDayID
	item.TransitionID = &result.TransitionID
}

func (s *DayScheduler) evaluateEndDay(ctx context.Context, item *entity.SchedulerRunItem) {
	if s.dayRepo == nil {
		failItem(item, fmt.Errorf("workflow day repository is required"))
		return
	}

	day, err := s.dayRepo.GetByContractDate(ctx, item.ContractID, item.BusinessDate)
	if err != nil {
		failItem(item, err)
		return
	}
	if day == nil {
		skipItem(item, "NOT_STARTED")
		return
	}
	item.WorkflowDayID = &day.ID

	switch day.CurrentState {
	case vo.StateAccountingClosed:
		skipItem(item, "ALREADY_ENDED")
	case vo.StateTransactionClosed:
		skipItem(item, "ACCOUNTING_CLOSE_NOT_AUTOMATED")
	case vo.StateManagerApproved:
		skipItem(item, "END_DAY_COMMAND_NOT_CONFIGURED")
	case vo.StateDayOpen:
		skipItem(item, "MANAGER_APPROVAL_REQUIRED")
	default:
		skipItem(item, "END_DAY_NOT_AVAILABLE")
	}
}

func (s *DayScheduler) finishRun(
	ctx context.Context,
	run *entity.SchedulerRun,
	status entity.SchedulerRunStatus,
	summary map[string]any,
	errorMessage *string,
) error {
	finishedAt := s.clock.Now().UTC()
	run.FinishedAt = &finishedAt
	run.Status = status
	run.Summary = summary
	run.ErrorMessage = errorMessage
	return s.schedulerRepo.FinishSchedulerRun(ctx, run)
}

func isRuleDue(rule *entity.ScheduleRule, localNow time.Time) (bool, string) {
	if !containsISODay(rule.DaysOfWeek, localNow) {
		return false, "NOT_SCHEDULED_DAY"
	}
	localMidnight := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, localNow.Location())
	triggerAt := localMidnight.Add(rule.TriggerTimeLocal)
	if localNow.Before(triggerAt) {
		return false, "BEFORE_TRIGGER_TIME"
	}
	return true, ""
}

func containsISODay(days []int, t time.Time) bool {
	if len(days) == 0 {
		return true
	}
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	for _, day := range days {
		if day == weekday {
			return true
		}
	}
	return false
}

func loadRuleLocation(timezone string) (*time.Location, error) {
	if timezone == "" {
		return clock.BangkokLocation, nil
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("loading scheduler timezone %q: %w", timezone, err)
	}
	return loc, nil
}

func businessDateInLocation(now time.Time, loc *time.Location) time.Time {
	local := now.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func formatDurationClock(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	hour := totalSeconds / 3600
	minute := (totalSeconds % 3600) / 60
	second := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

func skipItem(item *entity.SchedulerRunItem, reason string) {
	item.Status = entity.SchedulerItemStatusSkipped
	item.SkipReason = &reason
}

func failItem(item *entity.SchedulerRunItem, err error) {
	msg := err.Error()
	item.Status = entity.SchedulerItemStatusFailed
	item.ErrorMessage = &msg
}

type schedulerRunCounts struct {
	ContractsChecked int
	Succeeded        int
	Skipped          int
	Failed           int
}

func (c *schedulerRunCounts) add(status entity.SchedulerItemStatus) {
	switch status {
	case entity.SchedulerItemStatusSuccess:
		c.Succeeded++
	case entity.SchedulerItemStatusSkipped:
		c.Skipped++
	case entity.SchedulerItemStatusFailed:
		c.Failed++
	}
}

func (c schedulerRunCounts) status() entity.SchedulerRunStatus {
	if c.Failed > 0 && (c.Succeeded > 0 || c.Skipped > 0) {
		return entity.SchedulerRunStatusPartialSuccess
	}
	if c.Failed > 0 {
		return entity.SchedulerRunStatusFailed
	}
	if c.Succeeded > 0 {
		return entity.SchedulerRunStatusSuccess
	}
	return entity.SchedulerRunStatusSkipped
}

func (c schedulerRunCounts) summary() map[string]any {
	return map[string]any{
		"contracts_checked": c.ContractsChecked,
		"succeeded":         c.Succeeded,
		"skipped":           c.Skipped,
		"failed":            c.Failed,
	}
}

type schedulerTickSummary struct {
	RuleCount        int
	CheckedAt        time.Time
	Timezone         string
	ContractsChecked int
	Succeeded        int
	Skipped          int
	Failed           int
	RuleResults      []map[string]any
	primaryReason    string
}

func newSchedulerTickSummary(ruleCount int, checkedAt time.Time, timezone string) *schedulerTickSummary {
	return &schedulerTickSummary{
		RuleCount:   ruleCount,
		CheckedAt:   checkedAt.UTC(),
		Timezone:    timezone,
		RuleResults: make([]map[string]any, 0, ruleCount),
	}
}

func (s *schedulerTickSummary) addRuleResult(
	rule *entity.ScheduleRule,
	status entity.SchedulerRunStatus,
	reason string,
	extra map[string]any,
) {
	result := map[string]any{
		"rule_id":   rule.ID.String(),
		"rule_name": rule.Name,
		"action":    string(rule.Action),
		"status":    string(status),
	}
	if reason != "" {
		result["reason"] = reason
		if s.primaryReason == "" {
			s.primaryReason = reason
		}
	}
	for key, value := range extra {
		result[key] = value
	}
	s.RuleResults = append(s.RuleResults, result)
}

func (s *schedulerTickSummary) addRuleStatus(status entity.SchedulerRunStatus) {
	switch status {
	case entity.SchedulerRunStatusSuccess:
		s.Succeeded++
	case entity.SchedulerRunStatusSkipped:
		s.Skipped++
	case entity.SchedulerRunStatusFailed:
		s.Failed++
	case entity.SchedulerRunStatusPartialSuccess:
		s.Succeeded++
		s.Failed++
	}
}

func (s *schedulerTickSummary) addItem(status entity.SchedulerItemStatus) {
	s.ContractsChecked++
	switch status {
	case entity.SchedulerItemStatusSuccess:
		s.Succeeded++
	case entity.SchedulerItemStatusSkipped:
		s.Skipped++
	case entity.SchedulerItemStatusFailed:
		s.Failed++
	}
}

func (s schedulerTickSummary) status() entity.SchedulerRunStatus {
	if s.Failed > 0 && (s.Succeeded > 0 || s.Skipped > 0) {
		return entity.SchedulerRunStatusPartialSuccess
	}
	if s.Failed > 0 {
		return entity.SchedulerRunStatusFailed
	}
	if s.Succeeded > 0 {
		return entity.SchedulerRunStatusSuccess
	}
	return entity.SchedulerRunStatusSkipped
}

func (s schedulerTickSummary) mapSummary() map[string]any {
	summary := map[string]any{
		"rule_count":        s.RuleCount,
		"contracts_checked": s.ContractsChecked,
		"succeeded":         s.Succeeded,
		"skipped":           s.Skipped,
		"failed":            s.Failed,
		"checked_at":        s.CheckedAt.Format(time.RFC3339),
		"timezone":          s.Timezone,
		"rule_results":      s.RuleResults,
	}
	if s.primaryReason != "" && s.Succeeded == 0 {
		summary["reason"] = s.primaryReason
	}
	return summary
}

type emptyContractSource struct{}

func (emptyContractSource) ListActiveContracts(context.Context, time.Time) ([]uuid.UUID, error) {
	return nil, nil
}
