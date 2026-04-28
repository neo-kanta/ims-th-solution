package jobs

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/platform/clock"
)

func TestDayScheduler_RecordsSkippedRunBeforeTrigger(t *testing.T) {
	repo := &fakeSchedulerRepo{
		rules:        []*entity.ScheduleRule{testScheduleRule(entity.SchedulerActionOpenDay, 8*time.Hour+30*time.Minute)},
		lockAcquired: true,
	}
	scheduler := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{},
		fakeContractSource{},
		fakeHolidayCalendar{},
		&fakeOpenDayCommand{},
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 8, 0)},
		DaySchedulerConfig{},
	)

	runs, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected one run, got %d", len(runs))
	}
	if runs[0].Status != entity.SchedulerRunStatusSkipped {
		t.Fatalf("expected skipped run, got %s", runs[0].Status)
	}
	if got := runs[0].Summary["reason"]; got != "BEFORE_TRIGGER_TIME" {
		t.Fatalf("expected BEFORE_TRIGGER_TIME, got %v", got)
	}
	if len(repo.items) != 0 {
		t.Fatalf("expected no per-contract items before trigger, got %d", len(repo.items))
	}
}

func TestDayScheduler_RecordsSkippedRunOnNonScheduledDay(t *testing.T) {
	rule := testScheduleRule(entity.SchedulerActionOpenDay, 8*time.Hour+30*time.Minute)
	rule.DaysOfWeek = []int{1}
	repo := &fakeSchedulerRepo{
		rules:        []*entity.ScheduleRule{rule},
		lockAcquired: true,
	}
	scheduler := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{},
		fakeContractSource{},
		fakeHolidayCalendar{},
		&fakeOpenDayCommand{},
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 9, 0)},
		DaySchedulerConfig{},
	)

	runs, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runs[0].Status != entity.SchedulerRunStatusSkipped {
		t.Fatalf("expected skipped run, got %s", runs[0].Status)
	}
	if got := runs[0].Summary["reason"]; got != "NOT_SCHEDULED_DAY" {
		t.Fatalf("expected NOT_SCHEDULED_DAY, got %v", got)
	}
}

func TestDayScheduler_RecordsSkippedRunOnNonBusinessDay(t *testing.T) {
	repo := &fakeSchedulerRepo{
		rules:        []*entity.ScheduleRule{testScheduleRule(entity.SchedulerActionOpenDay, 8*time.Hour+30*time.Minute)},
		lockAcquired: true,
	}
	scheduler := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{},
		fakeContractSource{},
		fakeHolidayCalendar{isBusinessDay: false, holidayName: "WEEKEND"},
		&fakeOpenDayCommand{},
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 9, 0)},
		DaySchedulerConfig{},
	)

	runs, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runs[0].Status != entity.SchedulerRunStatusSkipped {
		t.Fatalf("expected skipped run, got %s", runs[0].Status)
	}
	if got := runs[0].Summary["reason"]; got != "NON_BUSINESS_DAY" {
		t.Fatalf("expected NON_BUSINESS_DAY, got %v", got)
	}
}

func TestDayScheduler_RecordsSkippedRunWhenNoActiveRules(t *testing.T) {
	repo := &fakeSchedulerRepo{lockAcquired: true}
	scheduler := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{},
		fakeContractSource{},
		fakeHolidayCalendar{},
		&fakeOpenDayCommand{},
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 9, 0)},
		DaySchedulerConfig{},
	)

	runs, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected one tick run, got %d", len(runs))
	}
	if runs[0].Status != entity.SchedulerRunStatusSkipped {
		t.Fatalf("expected skipped run, got %s", runs[0].Status)
	}
	if got := runs[0].Summary["reason"]; got != "NO_ACTIVE_RULES" {
		t.Fatalf("expected NO_ACTIVE_RULES, got %v", got)
	}
}

func TestDayScheduler_OpensDayWhenDueAndNotStarted(t *testing.T) {
	contractID := uuid.New()
	workflowDayID := uuid.New()
	transitionID := uuid.New()
	repo := &fakeSchedulerRepo{
		rules:        []*entity.ScheduleRule{testScheduleRule(entity.SchedulerActionOpenDay, 8*time.Hour+30*time.Minute)},
		lockAcquired: true,
	}
	scheduler := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{},
		fakeContractSource{contractIDs: []uuid.UUID{contractID}},
		fakeHolidayCalendar{},
		&fakeOpenDayCommand{
			result: &command.OpenDayResult{
				WorkflowDayID: workflowDayID,
				TransitionID:  transitionID,
			},
		},
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 9, 0)},
		DaySchedulerConfig{},
	)

	runs, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runs[0].Status != entity.SchedulerRunStatusSuccess {
		t.Fatalf("expected success run, got %s", runs[0].Status)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected one item, got %d", len(repo.items))
	}
	item := repo.items[0]
	if item.Status != entity.SchedulerItemStatusSuccess {
		t.Fatalf("expected success item, got %s", item.Status)
	}
	if item.WorkflowDayID == nil || *item.WorkflowDayID != workflowDayID {
		t.Fatalf("workflow day ID not captured")
	}
	if item.TransitionID == nil || *item.TransitionID != transitionID {
		t.Fatalf("transition ID not captured")
	}
}

func TestDayScheduler_RecordsNoActiveContractsAudit(t *testing.T) {
	repo := &fakeSchedulerRepo{
		rules:        []*entity.ScheduleRule{testScheduleRule(entity.SchedulerActionOpenDay, 8*time.Hour+30*time.Minute)},
		lockAcquired: true,
	}
	scheduler := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{},
		fakeContractSource{},
		fakeHolidayCalendar{},
		&fakeOpenDayCommand{},
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 9, 0)},
		DaySchedulerConfig{},
	)

	runs, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runs[0].Status != entity.SchedulerRunStatusSkipped {
		t.Fatalf("expected skipped run, got %s", runs[0].Status)
	}
	if got := runs[0].Summary["reason"]; got != "NO_ACTIVE_CONTRACTS" {
		t.Fatalf("expected NO_ACTIVE_CONTRACTS, got %v", got)
	}
	if len(repo.items) != 0 {
		t.Fatalf("expected no per-contract items, got %d", len(repo.items))
	}
}

func TestDayScheduler_EndDayAuditsManagerApprovalRequirement(t *testing.T) {
	contractID := uuid.New()
	businessDate := dateUTC(2026, 4, 24)
	day := &entity.WorkflowDay{
		ID:           uuid.New(),
		ContractID:   contractID,
		BusinessDate: businessDate,
		CurrentState: vo.StateDayOpen,
	}
	repo := &fakeSchedulerRepo{
		rules:        []*entity.ScheduleRule{testScheduleRule(entity.SchedulerActionEndDay, 17*time.Hour+30*time.Minute)},
		lockAcquired: true,
	}
	scheduler := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{days: map[string]*entity.WorkflowDay{
			dayKey(contractID, businessDate): day,
		}},
		fakeContractSource{contractIDs: []uuid.UUID{contractID}},
		fakeHolidayCalendar{},
		&fakeOpenDayCommand{},
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 18, 0)},
		DaySchedulerConfig{},
	)

	runs, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runs[0].Status != entity.SchedulerRunStatusSkipped {
		t.Fatalf("expected skipped run, got %s", runs[0].Status)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected one item, got %d", len(repo.items))
	}
	item := repo.items[0]
	if item.Status != entity.SchedulerItemStatusSkipped {
		t.Fatalf("expected skipped item, got %s", item.Status)
	}
	if item.SkipReason == nil || *item.SkipReason != "MANAGER_APPROVAL_REQUIRED" {
		t.Fatalf("expected MANAGER_APPROVAL_REQUIRED, got %v", item.SkipReason)
	}
}

func TestDayScheduler_LockNotAcquiredSkipsRuleEvaluation(t *testing.T) {
	repo := &fakeSchedulerRepo{
		rules:        []*entity.ScheduleRule{testScheduleRule(entity.SchedulerActionOpenDay, 8*time.Hour+30*time.Minute)},
		lockAcquired: false,
	}
	openDay := &fakeOpenDayCommand{}
	scheduler := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{},
		fakeContractSource{contractIDs: []uuid.UUID{uuid.New()}},
		fakeHolidayCalendar{},
		openDay,
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 9, 0)},
		DaySchedulerConfig{},
	)

	runs, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listRulesCalled {
		t.Fatalf("scheduler must not evaluate rules when lock is not acquired")
	}
	if openDay.callCount() != 0 {
		t.Fatalf("expected no open day attempts, got %d", openDay.callCount())
	}
	if runs[0].Status != entity.SchedulerRunStatusSkipped {
		t.Fatalf("expected skipped run, got %s", runs[0].Status)
	}
	if got := runs[0].Summary["reason"]; got != "SCHEDULER_LOCK_NOT_ACQUIRED" {
		t.Fatalf("expected SCHEDULER_LOCK_NOT_ACQUIRED, got %v", got)
	}
}

func TestDayScheduler_LockAcquiredRunsRules(t *testing.T) {
	contractID := uuid.New()
	repo := &fakeSchedulerRepo{
		rules:        []*entity.ScheduleRule{testScheduleRule(entity.SchedulerActionOpenDay, 8*time.Hour+30*time.Minute)},
		lockAcquired: true,
	}
	openDay := &fakeOpenDayCommand{}
	scheduler := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{},
		fakeContractSource{contractIDs: []uuid.UUID{contractID}},
		fakeHolidayCalendar{},
		openDay,
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 9, 0)},
		DaySchedulerConfig{},
	)

	if _, err := scheduler.RunOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.listRulesCalled {
		t.Fatalf("expected scheduler to evaluate rules")
	}
	if openDay.callCount() != 1 {
		t.Fatalf("expected one open day attempt, got %d", openDay.callCount())
	}
}

func TestDayScheduler_ConcurrentInstancesDoNotDuplicateOpenDayAttempts(t *testing.T) {
	contractID := uuid.New()
	repo := newSharedLockSchedulerRepo([]*entity.ScheduleRule{
		testScheduleRule(entity.SchedulerActionOpenDay, 8*time.Hour+30*time.Minute),
	})
	openDay := &blockingOpenDayCommand{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}

	first := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{},
		fakeContractSource{contractIDs: []uuid.UUID{contractID}},
		fakeHolidayCalendar{},
		openDay,
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 9, 0)},
		DaySchedulerConfig{},
	)
	second := NewDayScheduler(
		repo,
		&fakeWorkflowDayReader{},
		fakeContractSource{contractIDs: []uuid.UUID{contractID}},
		fakeHolidayCalendar{},
		openDay,
		clock.FixedClock{FixedTime: bangkokTimeUTC(2026, 4, 24, 9, 0)},
		DaySchedulerConfig{},
	)

	errCh := make(chan error, 1)
	go func() {
		_, err := first.RunOnce(context.Background())
		errCh <- err
	}()

	<-openDay.started
	runs, err := second.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("second scheduler returned error: %v", err)
	}
	if got := runs[0].Summary["reason"]; got != "SCHEDULER_LOCK_NOT_ACQUIRED" {
		t.Fatalf("expected second scheduler lock skip, got %v", got)
	}
	if openDay.callCount() != 1 {
		t.Fatalf("expected only one open day attempt while lock held, got %d", openDay.callCount())
	}

	close(openDay.release)
	if err := <-errCh; err != nil {
		t.Fatalf("first scheduler returned error: %v", err)
	}
	if openDay.callCount() != 1 {
		t.Fatalf("expected one total open day attempt, got %d", openDay.callCount())
	}
}

type fakeSchedulerRepo struct {
	mu              sync.Mutex
	rules           []*entity.ScheduleRule
	runs            []*entity.SchedulerRun
	items           []*entity.SchedulerRunItem
	lockAcquired    bool
	listRulesCalled bool
}

func (r *fakeSchedulerRepo) TryAcquireSchedulerLock(context.Context, int64) (domain.SchedulerLock, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.lockAcquired {
		return nil, false, nil
	}
	return fakeSchedulerLock{}, true, nil
}

func (r *fakeSchedulerRepo) ListActiveScheduleRules(context.Context, time.Time) ([]*entity.ScheduleRule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listRulesCalled = true
	return r.rules, nil
}

func (r *fakeSchedulerRepo) CreateSchedulerRun(_ context.Context, run *entity.SchedulerRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs = append(r.runs, run)
	return nil
}

func (r *fakeSchedulerRepo) FinishSchedulerRun(_ context.Context, run *entity.SchedulerRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return nil
}

func (r *fakeSchedulerRepo) InsertSchedulerRunItem(_ context.Context, item *entity.SchedulerRunItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, item)
	return nil
}

type fakeSchedulerLock struct{}

func (fakeSchedulerLock) Release(context.Context) error {
	return nil
}

type sharedLockSchedulerRepo struct {
	fakeSchedulerRepo
	lockHeld bool
}

func newSharedLockSchedulerRepo(rules []*entity.ScheduleRule) *sharedLockSchedulerRepo {
	return &sharedLockSchedulerRepo{
		fakeSchedulerRepo: fakeSchedulerRepo{
			rules:        rules,
			lockAcquired: true,
		},
	}
}

func (r *sharedLockSchedulerRepo) TryAcquireSchedulerLock(context.Context, int64) (domain.SchedulerLock, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.lockHeld {
		return nil, false, nil
	}
	r.lockHeld = true
	return sharedFakeSchedulerLock{repo: r}, true, nil
}

type sharedFakeSchedulerLock struct {
	repo *sharedLockSchedulerRepo
}

func (l sharedFakeSchedulerLock) Release(context.Context) error {
	l.repo.mu.Lock()
	defer l.repo.mu.Unlock()
	l.repo.lockHeld = false
	return nil
}

type fakeWorkflowDayReader struct {
	days map[string]*entity.WorkflowDay
	err  error
}

func (r *fakeWorkflowDayReader) GetByContractDate(
	_ context.Context,
	contractID uuid.UUID,
	businessDate time.Time,
) (*entity.WorkflowDay, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.days[dayKey(contractID, businessDate)], nil
}

type fakeContractSource struct {
	contractIDs []uuid.UUID
	err         error
}

func (s fakeContractSource) ListActiveContracts(context.Context, time.Time) ([]uuid.UUID, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.contractIDs, nil
}

type fakeHolidayCalendar struct {
	isBusinessDay bool
	holidayName   string
	err           error
}

func (c fakeHolidayCalendar) IsBusinessDay(context.Context, time.Time) (bool, string, error) {
	if c.err != nil {
		return false, "", c.err
	}
	if !c.isBusinessDay && c.holidayName != "" {
		return false, c.holidayName, nil
	}
	return true, "", nil
}

func (c fakeHolidayCalendar) PreviousBusinessDay(_ context.Context, date time.Time) (time.Time, error) {
	return date.AddDate(0, 0, -1), nil
}

type fakeOpenDayCommand struct {
	mu     sync.Mutex
	result *command.OpenDayResult
	err    error
	calls  int
}

func (c *fakeOpenDayCommand) Handle(context.Context, command.OpenDayRequest) (*command.OpenDayResult, error) {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()
	if c.err != nil {
		return nil, c.err
	}
	if c.result != nil {
		return c.result, nil
	}
	return &command.OpenDayResult{
		WorkflowDayID: uuid.New(),
		TransitionID:  uuid.New(),
	}, nil
}

func (c *fakeOpenDayCommand) callCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

type blockingOpenDayCommand struct {
	mu      sync.Mutex
	started chan struct{}
	release chan struct{}
	calls   int
	once    sync.Once
}

func (c *blockingOpenDayCommand) Handle(context.Context, command.OpenDayRequest) (*command.OpenDayResult, error) {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()
	c.once.Do(func() { close(c.started) })
	<-c.release
	return &command.OpenDayResult{
		WorkflowDayID: uuid.New(),
		TransitionID:  uuid.New(),
	}, nil
}

func (c *blockingOpenDayCommand) callCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func testScheduleRule(action entity.SchedulerAction, trigger time.Duration) *entity.ScheduleRule {
	return &entity.ScheduleRule{
		ID:               uuid.New(),
		DaySettingID:     uuid.New(),
		Name:             string(action),
		Action:           action,
		TriggerTimeLocal: trigger,
		Timezone:         "Asia/Bangkok",
		DaysOfWeek:       []int{1, 2, 3, 4, 5},
		SkipHolidays:     true,
		Priority:         100,
		EffectiveFrom:    dateUTC(2026, 1, 1),
	}
}

func bangkokTimeUTC(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, clock.BangkokLocation).UTC()
}

func dateUTC(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func dayKey(contractID uuid.UUID, businessDate time.Time) string {
	return fmt.Sprintf("%s/%s", contractID.String(), businessDate.Format("2006-01-02"))
}
