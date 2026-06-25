package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
	domainsvc "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/service"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// EvaluateFilter optionally narrows the scope of an evaluation run.
type EvaluateFilter struct {
	ScopeType   *entity.ScopeType
	PortfolioID *uuid.UUID
	SecurityID  *uuid.UUID
	ItemID      *uuid.UUID
	RuleID      *uuid.UUID
	DryRun      bool
}

// RuleEvalResult reports what happened to one rule during evaluation.
type RuleEvalResult struct {
	RuleID             uuid.UUID
	WatchlistItemID    uuid.UUID
	SecurityID         uuid.UUID
	PreviousState      entity.RuleState
	ComputedState      entity.RuleState
	WouldCreateAlert   bool
	NotificationStatus entity.NotificationStatus
	QuoteStatus        string
	ObservedPrice      *string // decimal string, nil when no quote fetched
	ThresholdValue     string
	Stale              bool
	StaleReason        string
}

// EvaluateOutput summarises the result of a full evaluation run.
type EvaluateOutput struct {
	DryRun           bool
	RulesEvaluated   int
	AlertsCreated    int
	AlertsSuppressed int
	RulesSkipped     int
	ProviderFailures int
	Results          []RuleEvalResult
}

// EvaluatorService orchestrates rule evaluation, alert creation, notification,
// and audit recording.
type EvaluatorService struct {
	pool          *pgxpool.Pool
	rules         repository.ThresholdRuleRepository
	items         repository.WatchlistItemRepository
	alerts        repository.AlertEventRepository
	securityPort  watchlistdomain.SecurityPort
	quotePort     watchlistdomain.QuotePort
	portfolioPort watchlistdomain.PortfolioScopePort
	notifier      watchlistdomain.WatchlistAlertNotifier
	auditRecorder watchlistdomain.WatchlistAuditRecorder
	iamChecker    contract.PermissionChecker
	maxStaleAge   time.Duration
}

func NewEvaluatorService(
	pool *pgxpool.Pool,
	rules repository.ThresholdRuleRepository,
	items repository.WatchlistItemRepository,
	alerts repository.AlertEventRepository,
	securityPort watchlistdomain.SecurityPort,
	quotePort watchlistdomain.QuotePort,
	portfolioPort watchlistdomain.PortfolioScopePort,
	notifier watchlistdomain.WatchlistAlertNotifier,
	auditRecorder watchlistdomain.WatchlistAuditRecorder,
	iamChecker contract.PermissionChecker,
) *EvaluatorService {
	return &EvaluatorService{
		pool:          pool,
		rules:         rules,
		items:         items,
		alerts:        alerts,
		securityPort:  securityPort,
		quotePort:     quotePort,
		portfolioPort: portfolioPort,
		notifier:      notifier,
		auditRecorder: auditRecorder,
		iamChecker:    iamChecker,
		maxStaleAge:   domainsvc.DefaultMaxStaleAge,
	}
}

// Evaluate runs the evaluation pipeline for all rules matching the filter.
// When filter.RuleID is set, errors are propagated; otherwise per-rule failures
// are counted as provider failures and skipped.
func (s *EvaluatorService) Evaluate(ctx context.Context, filter EvaluateFilter, actorID *uuid.UUID) (*EvaluateOutput, error) {
	if filter.RuleID != nil {
		r, err := s.rules.GetByID(ctx, *filter.RuleID)
		if err != nil {
			s.recordEvaluationFailed(ctx, actorID, nil, RuleEvalResult{RuleID: *filter.RuleID}, err)
			return nil, err
		}
		if r == nil || r.Status == entity.RuleStatusDisabled {
			s.recordEvaluationFailed(ctx, actorID, r, RuleEvalResult{RuleID: *filter.RuleID}, watchlistdomain.ErrRuleDisabled)
			return nil, watchlistdomain.ErrRuleDisabled
		}
	}

	if filter.PortfolioID != nil && s.iamChecker != nil && actorID != nil {
		scopeInfo, err := s.portfolioPort.GetPortfolioScope(ctx, *filter.PortfolioID)
		if err != nil || scopeInfo == nil {
			return nil, watchlistdomain.ErrForbiddenScope
		}
		ok, err := s.iamChecker.HasDataPermission(*actorID, scopeInfo.FundID.String())
		if err != nil || !ok {
			return nil, watchlistdomain.ErrForbiddenScope
		}
	}

	ruleFilter := repository.EvaluatorRuleFilter{
		ScopeType:   filter.ScopeType,
		PortfolioID: filter.PortfolioID,
		SecurityID:  filter.SecurityID,
		ItemID:      filter.ItemID,
		RuleID:      filter.RuleID,
	}

	eligibleRules, err := s.rules.ListForEvaluation(ctx, ruleFilter)
	if err != nil {
		return nil, fmt.Errorf("loading rules: %w", err)
	}

	output := &EvaluateOutput{DryRun: filter.DryRun}

	for _, rule := range eligibleRules {
		result, err := s.evaluateOne(ctx, rule, filter.DryRun, actorID)
		if err != nil {
			s.recordEvaluationFailed(ctx, actorID, rule, result, err)
			if filter.RuleID != nil {
				return nil, err
			}
			// Stale quote and forbidden scope are expected skip conditions, not provider failures.
			if errors.Is(err, watchlistdomain.ErrStaleQuote) || errors.Is(err, watchlistdomain.ErrForbiddenScope) {
				output.RulesSkipped++
				continue
			}
			output.ProviderFailures++
			output.RulesSkipped++
			continue
		}
		output.RulesEvaluated++
		switch result.NotificationStatus {
		case entity.NotificationStatusCreated:
			output.AlertsCreated++
		case entity.NotificationStatusSuppressed:
			output.AlertsSuppressed++
		}
		output.Results = append(output.Results, result)
	}

	return output, nil
}

func (s *EvaluatorService) recordEvaluationFailed(ctx context.Context, actorID *uuid.UUID, rule *entity.ThresholdRule, result RuleEvalResult, err error) {
	if s.auditRecorder == nil || err == nil {
		return
	}

	ruleID := result.RuleID
	if ruleID == uuid.Nil && rule != nil {
		ruleID = rule.ID
	}
	itemID := result.WatchlistItemID
	if itemID == uuid.Nil && rule != nil {
		itemID = rule.WatchlistItemID
	}

	metadata := map[string]interface{}{
		"error_code": evaluationFailureCode(err),
		"message":    err.Error(),
		"event_time": time.Now().UTC(),
	}
	if ruleID != uuid.Nil {
		metadata["rule_id"] = ruleID.String()
	}
	if itemID != uuid.Nil {
		metadata["watchlist_item_id"] = itemID.String()
	}
	if result.SecurityID != uuid.Nil {
		metadata["security_id"] = result.SecurityID.String()
	}
	if result.QuoteStatus != "" {
		metadata["quote_status"] = result.QuoteStatus
	}
	if result.StaleReason != "" {
		metadata["stale_reason"] = result.StaleReason
	}

	targetID := ""
	if ruleID != uuid.Nil {
		targetID = ruleID.String()
	}
	s.auditRecorder.Record(ctx, actorID, "WATCHLIST_EVALUATION_FAILED",
		"watchlist_threshold_rule", targetID, "", "", metadata)
}

func evaluationFailureCode(err error) string {
	switch {
	case errors.Is(err, watchlistdomain.ErrStaleQuote):
		return "WATCHLIST_STALE_QUOTE"
	case errors.Is(err, watchlistdomain.ErrProviderUnavailable):
		return "WATCHLIST_PROVIDER_UNAVAILABLE"
	case errors.Is(err, watchlistdomain.ErrForbiddenScope):
		return "WATCHLIST_FORBIDDEN_SCOPE"
	case errors.Is(err, watchlistdomain.ErrRuleDisabled):
		return "WATCHLIST_RULE_DISABLED"
	default:
		return "WATCHLIST_EVALUATION_FAILED"
	}
}

// evaluateOne evaluates a single rule.
//
// For non-dry-run: fetches quote first (no lock), then opens a transaction,
// re-reads the rule with SELECT FOR UPDATE (prevents duplicate alerts from
// concurrent evaluators), evaluates, persists, commits, then sends notification.
// For dry-run: evaluates against the snapshot from ListForEvaluation without
// locking, so dry-runs never block the scheduler.
func (s *EvaluatorService) evaluateOne(ctx context.Context, rule *entity.ThresholdRule, dryRun bool, actorID *uuid.UUID) (RuleEvalResult, error) {
	now := time.Now().UTC()
	threshStr := rule.ThresholdValue.StringFixed(8)
	base := RuleEvalResult{
		RuleID:          rule.ID,
		WatchlistItemID: rule.WatchlistItemID,
		ThresholdValue:  threshStr,
		PreviousState:   rule.LastState,
	}

	item, err := s.items.GetByID(ctx, rule.WatchlistItemID)
	if err != nil || item == nil {
		return base, fmt.Errorf("item not found for rule %s", rule.ID)
	}
	base.SecurityID = item.SecurityID

	// Enforce portfolio data permission for actor-triggered evaluations.
	// System-context runs (actorID == nil) bypass this check intentionally.
	if actorID != nil && s.iamChecker != nil && item.ScopeType == entity.ScopePortfolio && item.PortfolioID != nil {
		scopeInfo, scopeErr := s.portfolioPort.GetPortfolioScope(ctx, *item.PortfolioID)
		if scopeErr != nil || scopeInfo == nil {
			return base, watchlistdomain.ErrForbiddenScope
		}
		ok, permErr := s.iamChecker.HasDataPermission(*actorID, scopeInfo.FundID.String())
		if permErr != nil || !ok {
			return base, watchlistdomain.ErrForbiddenScope
		}
	}

	pm, err := s.securityPort.ResolveProviderSymbol(ctx, item.SecurityID.String(), s.quotePort.PrimaryProviderName())
	if err != nil || pm == nil {
		return base, fmt.Errorf("no provider mapping for security %s", item.SecurityID)
	}

	// Fetch quote before acquiring any row lock to avoid holding a lock across a network call.
	quote, err := s.quotePort.GetLatestQuote(ctx, pm.ProviderSymbol)
	if err != nil {
		base.QuoteStatus = "UNAVAILABLE"
		base.NotificationStatus = entity.NotificationStatusSkipped
		return base, err
	}
	if quote == nil {
		base.QuoteStatus = "UNAVAILABLE"
		base.NotificationStatus = entity.NotificationStatusSkipped
		return base, watchlistdomain.ErrProviderUnavailable
	}

	if dryRun {
		// Read-only path: evaluate against the snapshot from ListForEvaluation.
		// No locking — dry-runs must not block the scheduler.
		return s.evaluateDryRun(rule, quote, now, base)
	}

	// Non-dry-run: open a transaction and lock the rule row before evaluating.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return base, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	lockedRule, err := s.rules.GetByIDForUpdate(ctx, tx, rule.ID)
	if err != nil {
		return base, fmt.Errorf("locking rule: %w", err)
	}
	if lockedRule == nil || lockedRule.Status == entity.RuleStatusDisabled {
		// Rule was concurrently disabled between ListForEvaluation and now; skip.
		base.NotificationStatus = entity.NotificationStatusSkipped
		return base, nil
	}

	// Re-evaluate against the freshly-locked rule state (prevents duplicate alerts).
	base.PreviousState = lockedRule.LastState
	evalResult, err := domainsvc.Evaluate(lockedRule, quote, now, s.maxStaleAge)
	if err != nil {
		if errors.Is(err, watchlistdomain.ErrStaleQuote) {
			base.QuoteStatus = "STALE_SKIPPED"
			base.NotificationStatus = entity.NotificationStatusSkipped
			base.Stale = true
			base.StaleReason = quote.StaleReason
			return base, watchlistdomain.ErrStaleQuote
		}
		return base, err
	}

	if quote.Stale {
		base.QuoteStatus = "STALE_ACCEPTED"
	} else {
		base.QuoteStatus = "LIVE"
	}
	base.ComputedState = evalResult.NewState
	base.Stale = evalResult.Stale
	base.StaleReason = evalResult.StaleReason
	priceStr := evalResult.ObservedPrice.StringFixed(8)
	base.ObservedPrice = &priceStr
	base.WouldCreateAlert = evalResult.Outcome == domainsvc.OutcomeAlertCreated

	stateChanged := evalResult.NewState != lockedRule.LastState
	var stateChangedAt *time.Time
	if stateChanged {
		t := now
		stateChangedAt = &t
	} else {
		stateChangedAt = lockedRule.LastStateChangedAt
	}

	observedPrice := evalResult.ObservedPrice
	observedAt := quote.EffectiveAt
	stateUpd := repository.RuleStateUpdate{
		ID:                 lockedRule.ID,
		LastState:          evalResult.NewState,
		LastObservedPrice:  &observedPrice,
		LastObservedAt:     &observedAt,
		LastEvaluatedAt:    now,
		LastStateChangedAt: stateChangedAt,
		LastAlertedAt:      lockedRule.LastAlertedAt,
		LastQuoteStale:     evalResult.Stale,
		UpdatedBy:          actorID,
	}
	if evalResult.StaleReason != "" {
		staleReason := evalResult.StaleReason
		stateUpd.LastStaleReason = &staleReason
	}

	var alertID uuid.UUID
	var idempotencyConflict bool

	if evalResult.Outcome == domainsvc.OutcomeAlertCreated {
		alertID = uuid.New()
		alertedAt := now
		stateUpd.LastAlertedAt = &alertedAt

		var staleReason *string
		if evalResult.StaleReason != "" {
			sr := evalResult.StaleReason
			staleReason = &sr
		}
		quoteProvider := quote.Provider
		var currency *string
		if lockedRule.Currency != nil {
			currency = lockedRule.Currency
		} else if quote.Currency != "" {
			c := quote.Currency
			currency = &c
		}

		alertEvent := &entity.AlertEvent{
			ID:                 alertID,
			WatchlistItemID:    item.ID,
			ThresholdRuleID:    lockedRule.ID,
			ScopeType:          item.ScopeType,
			OwnerUserID:        item.OwnerUserID,
			PortfolioID:        item.PortfolioID,
			SecurityID:         item.SecurityID,
			CreatedByUserID:    item.CreatedBy,
			Direction:          lockedRule.Direction,
			PreviousState:      evalResult.PreviousState,
			CurrentState:       evalResult.NewState,
			ObservedPrice:      evalResult.ObservedPrice,
			ThresholdValue:     lockedRule.ThresholdValue,
			Currency:           currency,
			QuoteProvider:      &quoteProvider,
			ObservedAt:         quote.EffectiveAt,
			EvaluatedAt:        now,
			Stale:              evalResult.Stale,
			StaleReason:        staleReason,
			IdempotencyKey:     evalResult.AlertKey,
			NotificationStatus: entity.NotificationStatusPending,
			CreatedAt:          now,
		}

		insertErr := s.alerts.Insert(ctx, tx, alertEvent)
		if insertErr != nil && !errors.Is(insertErr, watchlistdomain.ErrAlertIdempotencyConflict) {
			return base, insertErr
		}
		idempotencyConflict = errors.Is(insertErr, watchlistdomain.ErrAlertIdempotencyConflict)
	}

	if err := s.rules.UpdateState(ctx, tx, stateUpd); err != nil {
		return base, err
	}
	if err := tx.Commit(ctx); err != nil {
		return base, fmt.Errorf("commit: %w", err)
	}

	if evalResult.Outcome == domainsvc.OutcomeCooldownSuppressed {
		base.NotificationStatus = entity.NotificationStatusSuppressed
		return base, nil
	}
	if evalResult.Outcome != domainsvc.OutcomeAlertCreated {
		base.NotificationStatus = entity.NotificationStatusSkipped
		return base, nil
	}

	if idempotencyConflict {
		base.NotificationStatus = entity.NotificationStatusSuppressed
		return base, nil
	}

	// Post-commit: notify and record audit (best-effort, non-transactional).
	recipientUserID := item.CreatedBy
	if item.ScopeType == entity.ScopePersonal && item.OwnerUserID != nil {
		recipientUserID = *item.OwnerUserID
	}

	var finalStatus entity.NotificationStatus
	if s.notifier != nil {
		secName, secSymbol := "", ""
		if secInfo, err := s.securityPort.GetSecurityByID(ctx, item.SecurityID.String()); err == nil && secInfo != nil {
			secName = secInfo.Name
			secSymbol = secInfo.DisplaySymbol
		}

		notifInput := watchlistdomain.WatchlistAlertNotificationInput{
			RecipientUserID: recipientUserID,
			AlertEventID:    alertID,
			WatchlistItemID: item.ID,
			ThresholdRuleID: lockedRule.ID,
			SecurityID:      item.SecurityID,
			SecurityName:    secName,
			SecuritySymbol:  secSymbol,
			PortfolioID:     item.PortfolioID,
			ScopeType:       string(item.ScopeType),
			Direction:       string(lockedRule.Direction),
			ObservedPrice:   evalResult.ObservedPrice,
			ThresholdValue:  lockedRule.ThresholdValue,
			Stale:           evalResult.Stale,
			StaleReason:     evalResult.StaleReason,
			IdempotencyKey:  evalResult.AlertKey,
		}
		if lockedRule.Currency != nil {
			notifInput.Currency = *lockedRule.Currency
		}

		notifErr := s.notifier.NotifyThresholdBreached(ctx, notifInput)
		if notifErr != nil {
			finalStatus = entity.NotificationStatusFailed
			errMsg := notifErr.Error()
			_ = s.alerts.UpdateNotificationStatus(ctx, alertID, finalStatus, &errMsg)
		} else {
			finalStatus = entity.NotificationStatusCreated
			_ = s.alerts.UpdateNotificationStatus(ctx, alertID, finalStatus, nil)
		}
	} else {
		finalStatus = entity.NotificationStatusSkipped
	}

	s.auditRecorder.Record(ctx, actorID, "WATCHLIST_ALERT_TRIGGERED",
		"watchlist_alert_event", alertID.String(), "", "",
		map[string]interface{}{
			"rule_id":     lockedRule.ID.String(),
			"security_id": item.SecurityID.String(),
			"outcome":     string(finalStatus),
		})

	base.NotificationStatus = finalStatus
	return base, nil
}

func (s *EvaluatorService) evaluateDryRun(rule *entity.ThresholdRule, quote *watchlistdomain.QuoteInfo, now time.Time, base RuleEvalResult) (RuleEvalResult, error) {
	evalResult, err := domainsvc.Evaluate(rule, quote, now, s.maxStaleAge)
	if err != nil {
		if errors.Is(err, watchlistdomain.ErrStaleQuote) {
			base.QuoteStatus = "STALE_SKIPPED"
			base.NotificationStatus = entity.NotificationStatusSkipped
			base.Stale = true
			base.StaleReason = quote.StaleReason
			return base, watchlistdomain.ErrStaleQuote
		}
		return base, err
	}

	if quote.Stale {
		base.QuoteStatus = "STALE_ACCEPTED"
	} else {
		base.QuoteStatus = "LIVE"
	}
	base.ComputedState = evalResult.NewState
	base.Stale = evalResult.Stale
	base.StaleReason = evalResult.StaleReason
	priceStr := evalResult.ObservedPrice.StringFixed(8)
	base.ObservedPrice = &priceStr
	base.WouldCreateAlert = evalResult.Outcome == domainsvc.OutcomeAlertCreated
	if evalResult.Outcome == domainsvc.OutcomeCooldownSuppressed {
		base.NotificationStatus = entity.NotificationStatusSuppressed
	} else {
		base.NotificationStatus = entity.NotificationStatusSkipped
	}
	return base, nil
}
