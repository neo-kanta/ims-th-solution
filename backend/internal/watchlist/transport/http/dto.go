package http

import (
	"time"

	"github.com/google/uuid"

	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
)

// ─── Descriptors ────────────────────────────────────────────────────────────

// SecurityDescriptor carries display fields for a canonical security.
type SecurityDescriptor struct {
	SecurityID    string `json:"security_id"`
	IMSSymbol     string `json:"ims_symbol"`
	DisplaySymbol string `json:"display_symbol"`
	Name          string `json:"name"`
	AssetType     string `json:"asset_type"`
	Currency      string `json:"currency"`
	ExchangeMIC   string `json:"exchange_mic"`
}

func securityDescriptorFromInfo(info *watchlistdomain.SecurityInfo) *SecurityDescriptor {
	if info == nil {
		return nil
	}
	return &SecurityDescriptor{
		SecurityID:    info.ID,
		IMSSymbol:     info.IMSSymbol,
		DisplaySymbol: info.DisplaySymbol,
		Name:          info.Name,
		AssetType:     info.AssetType,
		Currency:      info.Currency,
		ExchangeMIC:   info.ExchangeMIC,
	}
}

// PortfolioDescriptor carries display fields for a portfolio.
type PortfolioDescriptor struct {
	PortfolioID   string `json:"portfolio_id"`
	PortfolioCode string `json:"portfolio_code"`
	PortfolioName string `json:"portfolio_name"`
	FundID        string `json:"fund_id"`
	FundCode      string `json:"fund_code"`
	FundName      string `json:"fund_name"`
	DisplayName   string `json:"display_name"`
}

func portfolioDescriptorFromScope(info *watchlistdomain.PortfolioScopeInfo) *PortfolioDescriptor {
	if info == nil {
		return nil
	}
	displayName := info.PortfolioCode + " – " + info.PortfolioName
	return &PortfolioDescriptor{
		PortfolioID:   info.PortfolioID.String(),
		PortfolioCode: info.PortfolioCode,
		PortfolioName: info.PortfolioName,
		FundID:        info.FundID.String(),
		FundCode:      info.FundCode,
		FundName:      info.FundName,
		DisplayName:   displayName,
	}
}

// UserDescriptor carries display information for a user referenced in responses.
type UserDescriptor struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

// QuoteSnapshot is the market data view included in item responses.
type QuoteSnapshot struct {
	Symbol        string  `json:"symbol"`
	Provider      string  `json:"provider"`
	Price         string  `json:"price"`
	Currency      string  `json:"currency"`
	PreviousClose string  `json:"previous_close"`
	ChangePercent string  `json:"change_percent"`
	EffectiveAt   string  `json:"effective_at"`
	FetchedAt     string  `json:"fetched_at"`
	MarketStatus  string  `json:"market_status"`
	Stale         bool    `json:"stale"`
	StaleReason   *string `json:"stale_reason"`
}

func quoteSnapshotFromInfo(q *watchlistdomain.QuoteInfo) *QuoteSnapshot {
	if q == nil {
		return nil
	}
	var staleReason *string
	if q.StaleReason != "" {
		sr := q.StaleReason
		staleReason = &sr
	}
	return &QuoteSnapshot{
		Symbol:        q.Symbol,
		Provider:      q.Provider,
		Price:         q.Price.StringFixed(8),
		Currency:      q.Currency,
		PreviousClose: q.PreviousClose.StringFixed(8),
		ChangePercent: q.ChangePercent.StringFixed(8),
		EffectiveAt:   q.EffectiveAt.UTC().Format(time.RFC3339),
		FetchedAt:     q.FetchedAt.UTC().Format(time.RFC3339),
		MarketStatus:  q.MarketStatus,
		Stale:         q.Stale,
		StaleReason:   staleReason,
	}
}

// ─── Threshold Rule response ─────────────────────────────────────────────────

// ThresholdRuleResponse is one threshold rule in item responses.
type ThresholdRuleResponse struct {
	ID                 string  `json:"id"`
	MetricType         string  `json:"metric_type"`
	Direction          string  `json:"direction"`
	ThresholdValue     string  `json:"threshold_value"`
	Currency           *string `json:"currency"`
	CooldownMinutes    int     `json:"cooldown_minutes"`
	Status             string  `json:"status"`
	LastState          string  `json:"last_state"`
	LastObservedPrice  *string `json:"last_observed_price"`
	LastObservedAt     *string `json:"last_observed_at"`
	LastEvaluatedAt    *string `json:"last_evaluated_at"`
	LastStateChangedAt *string `json:"last_state_changed_at"`
	LastAlertedAt      *string `json:"last_alerted_at"`
	LastQuoteStale     bool    `json:"last_quote_stale"`
	LastStaleReason    *string `json:"last_stale_reason"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

func thresholdRuleResponse(r *entity.ThresholdRule) ThresholdRuleResponse {
	resp := ThresholdRuleResponse{
		ID:              r.ID.String(),
		MetricType:      string(r.MetricType),
		Direction:       string(r.Direction),
		ThresholdValue:  r.ThresholdValue.StringFixed(8),
		Currency:        r.Currency,
		CooldownMinutes: r.CooldownMinutes,
		Status:          string(r.Status),
		LastState:       string(r.LastState),
		LastQuoteStale:  r.LastQuoteStale,
		LastStaleReason: r.LastStaleReason,
		CreatedAt:       r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       r.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if r.LastObservedPrice != nil {
		s := r.LastObservedPrice.StringFixed(8)
		resp.LastObservedPrice = &s
	}
	if r.LastObservedAt != nil {
		s := r.LastObservedAt.UTC().Format(time.RFC3339)
		resp.LastObservedAt = &s
	}
	if r.LastEvaluatedAt != nil {
		s := r.LastEvaluatedAt.UTC().Format(time.RFC3339)
		resp.LastEvaluatedAt = &s
	}
	if r.LastStateChangedAt != nil {
		s := r.LastStateChangedAt.UTC().Format(time.RFC3339)
		resp.LastStateChangedAt = &s
	}
	if r.LastAlertedAt != nil {
		s := r.LastAlertedAt.UTC().Format(time.RFC3339)
		resp.LastAlertedAt = &s
	}
	return resp
}

// ─── Watchlist Item response ─────────────────────────────────────────────────

// WatchlistItemResponse is the top-level item resource.
type WatchlistItemResponse struct {
	ID              string                   `json:"id"`
	ScopeType       string                   `json:"scope_type"`
	OwnerUserID     *string                  `json:"owner_user_id"`
	CreatedByUserID string                   `json:"created_by_user_id"`
	CreatedByUser   *UserDescriptor          `json:"created_by_user,omitempty"`
	PortfolioID     *string                  `json:"portfolio_id"`
	Portfolio       *PortfolioDescriptor     `json:"portfolio"`
	Security        *SecurityDescriptor      `json:"security"`
	Status          string                   `json:"status"`
	Pinned          bool                     `json:"pinned"`
	Note            *string                  `json:"note"`
	ThresholdRules  *[]ThresholdRuleResponse `json:"threshold_rules"`
	Quote           *QuoteSnapshot           `json:"quote"`
	CreatedAt       string                   `json:"created_at"`
	UpdatedAt       string                   `json:"updated_at"`
}

func watchlistItemResponse(
	item *entity.WatchlistItem,
	rules []*entity.ThresholdRule,
	includeThresholds bool,
	security *SecurityDescriptor,
	portfolio *PortfolioDescriptor,
	quote *QuoteSnapshot,
	createdByUser *UserDescriptor,
) WatchlistItemResponse {
	resp := WatchlistItemResponse{
		ID:              item.ID.String(),
		ScopeType:       string(item.ScopeType),
		CreatedByUserID: item.CreatedBy.String(),
		CreatedByUser:   createdByUser,
		Status:          string(item.Status),
		Pinned:          item.Pinned,
		Note:            item.Note,
		Security:        security,
		Portfolio:       portfolio,
		Quote:           quote,
		CreatedAt:       item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       item.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if item.OwnerUserID != nil {
		s := item.OwnerUserID.String()
		resp.OwnerUserID = &s
	}
	if item.PortfolioID != nil {
		s := item.PortfolioID.String()
		resp.PortfolioID = &s
	}
	if includeThresholds {
		rr := make([]ThresholdRuleResponse, 0, len(rules))
		for _, r := range rules {
			rr = append(rr, thresholdRuleResponse(r))
		}
		resp.ThresholdRules = &rr
	}
	return resp
}

// ─── Alert Event response ────────────────────────────────────────────────────

// AlertEventResponse is the alert resource.
type AlertEventResponse struct {
	ID                   string               `json:"id"`
	WatchlistItemID      string               `json:"watchlist_item_id"`
	ThresholdRuleID      string               `json:"threshold_rule_id"`
	ScopeType            string               `json:"scope_type"`
	OwnerUserID          *string              `json:"owner_user_id"`
	CreatedByUserID      string               `json:"created_by_user_id"`
	CreatedByUser        *UserDescriptor      `json:"created_by_user,omitempty"`
	PortfolioID          *string              `json:"portfolio_id"`
	Portfolio            *PortfolioDescriptor `json:"portfolio"`
	Security             *SecurityDescriptor  `json:"security"`
	Direction            string               `json:"direction"`
	PreviousState        string               `json:"previous_state"`
	CurrentState         string               `json:"current_state"`
	ObservedPrice        string               `json:"observed_price"`
	ThresholdValue       string               `json:"threshold_value"`
	Currency             *string              `json:"currency"`
	QuoteProvider        *string              `json:"quote_provider"`
	ObservedAt           string               `json:"observed_at"`
	EvaluatedAt          string               `json:"evaluated_at"`
	Stale                bool                 `json:"stale"`
	StaleReason          *string              `json:"stale_reason"`
	NotificationStatus   string               `json:"notification_status"`
	AcknowledgementState string               `json:"acknowledgement_state"`
	AcknowledgedBy       *string              `json:"acknowledged_by"`
	AcknowledgedByUser   *UserDescriptor      `json:"acknowledged_by_user,omitempty"`
	AcknowledgedAt       *string              `json:"acknowledged_at"`
	AcknowledgementNote  *string              `json:"acknowledgement_note"`
	CreatedAt            string               `json:"created_at"`
}

func alertEventResponse(
	ev *entity.AlertEvent,
	security *SecurityDescriptor,
	portfolio *PortfolioDescriptor,
	createdByUser *UserDescriptor,
	acknowledgedByUser *UserDescriptor,
) AlertEventResponse {
	resp := AlertEventResponse{
		ID:                   ev.ID.String(),
		WatchlistItemID:      ev.WatchlistItemID.String(),
		ThresholdRuleID:      ev.ThresholdRuleID.String(),
		ScopeType:            string(ev.ScopeType),
		CreatedByUserID:      ev.CreatedByUserID.String(),
		CreatedByUser:        createdByUser,
		Security:             security,
		Portfolio:            portfolio,
		Direction:            string(ev.Direction),
		PreviousState:        string(ev.PreviousState),
		CurrentState:         string(ev.CurrentState),
		ObservedPrice:        ev.ObservedPrice.StringFixed(8),
		ThresholdValue:       ev.ThresholdValue.StringFixed(8),
		Currency:             ev.Currency,
		QuoteProvider:        ev.QuoteProvider,
		ObservedAt:           ev.ObservedAt.UTC().Format(time.RFC3339),
		EvaluatedAt:          ev.EvaluatedAt.UTC().Format(time.RFC3339),
		Stale:                ev.Stale,
		StaleReason:          ev.StaleReason,
		NotificationStatus:   string(ev.NotificationStatus),
		AcknowledgementState: string(ev.AckState()),
		AcknowledgedByUser:   acknowledgedByUser,
		AcknowledgementNote:  ev.AcknowledgementNote,
		CreatedAt:            ev.CreatedAt.UTC().Format(time.RFC3339),
	}
	if ev.OwnerUserID != nil {
		s := ev.OwnerUserID.String()
		resp.OwnerUserID = &s
	}
	if ev.PortfolioID != nil {
		s := ev.PortfolioID.String()
		resp.PortfolioID = &s
	}
	if ev.AcknowledgedBy != nil {
		s := ev.AcknowledgedBy.String()
		resp.AcknowledgedBy = &s
	}
	if ev.AcknowledgedAt != nil {
		s := ev.AcknowledgedAt.UTC().Format(time.RFC3339)
		resp.AcknowledgedAt = &s
	}
	return resp
}

// ─── Pagination ──────────────────────────────────────────────────────────────

type PaginationMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// ─── Request bodies ──────────────────────────────────────────────────────────

// ThresholdRuleRequest is one rule entry in create/update request bodies.
type ThresholdRuleRequest struct {
	ID              *string `json:"id"`
	MetricType      *string `json:"metric_type"`
	Direction       string  `json:"direction"`
	ThresholdValue  string  `json:"threshold_value"`
	Currency        *string `json:"currency"`
	CooldownMinutes *int    `json:"cooldown_minutes"`
	Status          *string `json:"status"`
}

// CreateItemRequest is the POST /watchlists/items body.
type CreateItemRequest struct {
	ScopeType      string                 `json:"scope_type"`
	PortfolioID    *string                `json:"portfolio_id"`
	SecurityID     string                 `json:"security_id"`
	Pinned         *bool                  `json:"pinned"`
	Note           *string                `json:"note"`
	ThresholdRules []ThresholdRuleRequest `json:"threshold_rules"`
}

// UpdateItemRequest is the PATCH /watchlists/items/{id} body.
type UpdateItemRequest struct {
	Pinned         *bool                   `json:"pinned"`
	Note           *string                 `json:"note"`
	Status         *string                 `json:"status"`
	ThresholdRules *[]ThresholdRuleRequest `json:"threshold_rules"`
}

// AcknowledgeAlertRequest is the POST /watchlists/alerts/{id}/acknowledge body.
type AcknowledgeAlertRequest struct {
	Note *string `json:"note"`
}

// EvaluateRequest is the POST /watchlists/evaluate body.
type EvaluateRequest struct {
	ScopeType   *string `json:"scope_type"`
	PortfolioID *string `json:"portfolio_id"`
	SecurityID  *string `json:"security_id"`
	ItemID      *string `json:"item_id"`
	RuleID      *string `json:"rule_id"`
	DryRun      *bool   `json:"dry_run"`
}

// ─── Response bodies ─────────────────────────────────────────────────────────

type ListItemsResponseData struct {
	Items      []WatchlistItemResponse `json:"items"`
	Pagination PaginationMeta          `json:"pagination"`
}

type ListAlertsResponseData struct {
	Items      []AlertEventResponse `json:"items"`
	Pagination PaginationMeta       `json:"pagination"`
}

// EvaluateRuleResultResponse is one per-rule result in the evaluate response.
type EvaluateRuleResultResponse struct {
	RuleID             string  `json:"rule_id"`
	WatchlistItemID    string  `json:"watchlist_item_id"`
	SecurityID         string  `json:"security_id"`
	PreviousState      string  `json:"previous_state"`
	ComputedState      string  `json:"computed_state"`
	WouldCreateAlert   bool    `json:"would_create_alert"`
	NotificationStatus string  `json:"notification_status"`
	QuoteStatus        string  `json:"quote_status"`
	ObservedPrice      *string `json:"observed_price"`
	ThresholdValue     string  `json:"threshold_value"`
	Stale              bool    `json:"stale"`
	StaleReason        *string `json:"stale_reason"`
}

type EvaluateResponseData struct {
	DryRun           bool                         `json:"dry_run"`
	RulesEvaluated   int                          `json:"rules_evaluated"`
	AlertsCreated    int                          `json:"alerts_created"`
	AlertsSuppressed int                          `json:"alerts_suppressed"`
	RulesSkipped     int                          `json:"rules_skipped"`
	ProviderFailures int                          `json:"provider_failures"`
	Results          []EvaluateRuleResultResponse `json:"results"`
}

// uuidStr formats a UUID pointer as a string pointer.
func uuidStr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// timeStr formats a time pointer as RFC3339 string pointer.
func timeStr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}
