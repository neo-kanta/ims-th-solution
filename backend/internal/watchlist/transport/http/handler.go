package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/application/query"
	appsvc "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/application/service"
	watchlistdomain "github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// Handler holds all watchlist HTTP handlers.
type Handler struct {
	createItem    *command.CreateItemHandler
	updateItem    *command.UpdateItemHandler
	deleteItem    *command.DeleteItemHandler
	ackAlert      *command.AcknowledgeAlertHandler
	listItems     *query.ListItemsHandler
	listAlerts    *query.ListAlertsHandler
	evaluator     *appsvc.EvaluatorService
	rules         repository.ThresholdRuleRepository
	securityPort  watchlistdomain.SecurityPort
	quotePort     watchlistdomain.QuotePort
	portfolioPort watchlistdomain.PortfolioScopePort
	userLookup    watchlistdomain.UserLookupPort
}

func NewHandler(
	createItem *command.CreateItemHandler,
	updateItem *command.UpdateItemHandler,
	deleteItem *command.DeleteItemHandler,
	ackAlert *command.AcknowledgeAlertHandler,
	listItems *query.ListItemsHandler,
	listAlerts *query.ListAlertsHandler,
	evaluator *appsvc.EvaluatorService,
	rules repository.ThresholdRuleRepository,
	securityPort watchlistdomain.SecurityPort,
	quotePort watchlistdomain.QuotePort,
	portfolioPort watchlistdomain.PortfolioScopePort,
	userLookup watchlistdomain.UserLookupPort,
) *Handler {
	return &Handler{
		createItem:    createItem,
		updateItem:    updateItem,
		deleteItem:    deleteItem,
		ackAlert:      ackAlert,
		listItems:     listItems,
		listAlerts:    listAlerts,
		evaluator:     evaluator,
		rules:         rules,
		securityPort:  securityPort,
		quotePort:     quotePort,
		portfolioPort: portfolioPort,
		userLookup:    userLookup,
	}
}

// ─── A. List Watchlist Items ─────────────────────────────────────────────────

// ListItems handles GET /watchlists.
// @Summary List watchlist items
// @Tags Watchlist
// @Security BearerAuth
// @Produce json
// @Param scope_type query string false "PERSONAL or PORTFOLIO"
// @Param portfolio_id query string false "Portfolio UUID (PORTFOLIO scope only)"
// @Param security_id query string false "Canonical security UUID"
// @Param include_disabled query bool false "Include disabled items (default false)"
// @Param include_thresholds query bool false "Include threshold rules (default true)"
// @Param include_quote query bool false "Include live quote (default true)"
// @Param limit query int false "Page size (1-200, default 50)"
// @Param offset query int false "Page offset (default 0)"
// @Success 200 {object} httputil.SuccessResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /watchlists [get]
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeWatchlistError(w, http.StatusUnauthorized, "not authenticated", "UNAUTHORIZED", nil)
		return
	}

	q := r.URL.Query()

	var scopeType *entity.ScopeType
	if s := q.Get("scope_type"); s != "" {
		if s != "PERSONAL" && s != "PORTFOLIO" {
			writeWatchlistError(w, http.StatusBadRequest, "invalid scope_type", "VALIDATION_ERROR",
				map[string]interface{}{"field": "scope_type", "allowed_values": []string{"PERSONAL", "PORTFOLIO"}})
			return
		}
		st := entity.ScopeType(s)
		scopeType = &st
	}

	var portfolioID *uuid.UUID
	if s := q.Get("portfolio_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid portfolio_id", "VALIDATION_ERROR",
				map[string]interface{}{"field": "portfolio_id"})
			return
		}
		portfolioID = &id
	}
	if portfolioID != nil && (scopeType == nil || *scopeType != entity.ScopePortfolio) {
		writeWatchlistError(w, http.StatusBadRequest, "portfolio_id requires scope_type=PORTFOLIO", "WATCHLIST_INVALID_QUERY",
			map[string]interface{}{"field": "portfolio_id", "required_scope_type": "PORTFOLIO"})
		return
	}

	var securityID *uuid.UUID
	if s := q.Get("security_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid security_id", "VALIDATION_ERROR",
				map[string]interface{}{"field": "security_id"})
			return
		}
		securityID = &id
	}

	limit, offset := pagination(q.Get("limit"), q.Get("offset"))
	includeDisabled := q.Get("include_disabled") == "true"
	includeThresholds := q.Get("include_thresholds") != "false" // default true
	includeQuote := q.Get("include_quote") != "false"           // default true

	out, err := h.listItems.Handle(r.Context(), query.ListItemsInput{
		ActorID:         actor,
		ScopeType:       scopeType,
		PortfolioID:     portfolioID,
		SecurityID:      securityID,
		IncludeDisabled: includeDisabled,
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		writeWatchlistDomainError(w, err, nil)
		return
	}

	// Batch-lookup creators for enrichment (best-effort; nil map on error is safe).
	var creatorIDs []uuid.UUID
	for _, item := range out.Items {
		creatorIDs = append(creatorIDs, item.CreatedBy)
	}
	users := h.lookupUsers(r.Context(), creatorIDs...)

	items := make([]WatchlistItemResponse, 0, len(out.Items))
	for _, item := range out.Items {
		// Security descriptor.
		var sec *SecurityDescriptor
		if secInfo, _ := h.securityPort.GetSecurityByID(r.Context(), item.SecurityID.String()); secInfo != nil {
			sec = securityDescriptorFromInfo(secInfo)
		}

		// Portfolio descriptor.
		var port *PortfolioDescriptor
		if item.PortfolioID != nil {
			if pInfo, _ := h.portfolioPort.GetPortfolioScope(r.Context(), *item.PortfolioID); pInfo != nil {
				port = portfolioDescriptorFromScope(pInfo)
			}
		}

		// Threshold rules.
		var ruleList []*entity.ThresholdRule
		if includeThresholds {
			ruleList, _ = h.rules.ListByItemID(r.Context(), item.ID)
		}

		// Quote.
		var qs *QuoteSnapshot
		if includeQuote {
			if pm, _ := h.securityPort.ResolveProviderSymbol(r.Context(), item.SecurityID.String(), h.quotePort.PrimaryProviderName()); pm != nil {
				if q, _ := h.quotePort.GetLatestQuote(r.Context(), pm.ProviderSymbol); q != nil {
					qs = quoteSnapshotFromInfo(q)
				}
			}
		}

		var createdByUser *UserDescriptor
		if info, ok := users[item.CreatedBy]; ok {
			createdByUser = &UserDescriptor{UserID: info.UserID.String(), DisplayName: info.DisplayName}
		}

		items = append(items, watchlistItemResponse(item, ruleList, includeThresholds, sec, port, qs, createdByUser))
	}

	httputil.OK(w, ListItemsResponseData{
		Items: items,
		Pagination: PaginationMeta{
			Limit:  out.Limit,
			Offset: out.Offset,
			Total:  out.Total,
		},
	})
}

// ─── B. Create Watchlist Item ─────────────────────────────────────────────────

// CreateItem handles POST /watchlists/items.
// @Summary Create a watchlist item
// @Tags Watchlist
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body CreateItemRequest true "Create payload"
// @Success 201 {object} httputil.SuccessResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /watchlists/items [post]
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeWatchlistError(w, http.StatusUnauthorized, "not authenticated", "UNAUTHORIZED", nil)
		return
	}

	var body CreateItemRequest
	if err := decodeJSON(r, &body); err != nil {
		writeWatchlistError(w, http.StatusBadRequest, "invalid request body", "VALIDATION_ERROR", nil)
		return
	}

	if body.ScopeType == "" {
		writeWatchlistError(w, http.StatusBadRequest, "scope_type is required", "VALIDATION_ERROR",
			map[string]interface{}{"field": "scope_type"})
		return
	}
	if body.ScopeType != "PERSONAL" && body.ScopeType != "PORTFOLIO" {
		writeWatchlistError(w, http.StatusBadRequest, "invalid scope_type", "VALIDATION_ERROR",
			map[string]interface{}{"field": "scope_type", "allowed_values": []string{"PERSONAL", "PORTFOLIO"}})
		return
	}

	secID, err := uuid.Parse(body.SecurityID)
	if err != nil || body.SecurityID == "" {
		writeWatchlistError(w, http.StatusBadRequest, "security_id is required", "VALIDATION_ERROR",
			map[string]interface{}{"field": "security_id"})
		return
	}

	var portfolioID *uuid.UUID
	if body.ScopeType == "PORTFOLIO" {
		if body.PortfolioID == nil || *body.PortfolioID == "" {
			writeWatchlistError(w, http.StatusBadRequest, "portfolio_id is required for PORTFOLIO scope", "VALIDATION_ERROR",
				map[string]interface{}{"field": "portfolio_id"})
			return
		}
		pid, err := uuid.Parse(*body.PortfolioID)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid portfolio_id", "VALIDATION_ERROR",
				map[string]interface{}{"field": "portfolio_id"})
			return
		}
		portfolioID = &pid
	}

	rules, err := parseThresholdRuleInputs(body.ThresholdRules)
	if err != nil {
		writeWatchlistError(w, http.StatusBadRequest, err.Error(), "WATCHLIST_INVALID_THRESHOLD", nil)
		return
	}

	pinned := false
	if body.Pinned != nil {
		pinned = *body.Pinned
	}

	out, err := h.createItem.Handle(r.Context(), command.CreateItemInput{
		ScopeType:      entity.ScopeType(body.ScopeType),
		PortfolioID:    portfolioID,
		SecurityID:     secID,
		Pinned:         pinned,
		Note:           body.Note,
		ThresholdRules: rules,
		ActorID:        actor,
	})
	if err != nil {
		writeWatchlistDomainError(w, err, map[string]interface{}{
			"scope_type":  body.ScopeType,
			"security_id": body.SecurityID,
		})
		return
	}

	sec, _ := h.securityPort.GetSecurityByID(r.Context(), out.Item.SecurityID.String())
	var port *PortfolioDescriptor
	if out.Item.PortfolioID != nil {
		if pInfo, _ := h.portfolioPort.GetPortfolioScope(r.Context(), *out.Item.PortfolioID); pInfo != nil {
			port = portfolioDescriptorFromScope(pInfo)
		}
	}
	users := h.lookupUsers(r.Context(), out.Item.CreatedBy)
	var createdByUser *UserDescriptor
	if info, ok := users[out.Item.CreatedBy]; ok {
		createdByUser = &UserDescriptor{UserID: info.UserID.String(), DisplayName: info.DisplayName}
	}

	httputil.Created(w, watchlistItemResponse(out.Item, out.Rules, true, securityDescriptorFromInfo(sec), port, nil, createdByUser))
}

// ─── C. Update Watchlist Item ─────────────────────────────────────────────────

// UpdateItem handles PATCH /watchlists/items/{id}.
// @Summary Update a watchlist item
// @Tags Watchlist
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Item UUID"
// @Param payload body UpdateItemRequest true "Update payload"
// @Success 200 {object} httputil.SuccessResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /watchlists/items/{id} [patch]
func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeWatchlistError(w, http.StatusUnauthorized, "not authenticated", "UNAUTHORIZED", nil)
		return
	}

	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeWatchlistError(w, http.StatusBadRequest, "invalid item id", "VALIDATION_ERROR",
			map[string]interface{}{"field": "id"})
		return
	}

	var body UpdateItemRequest
	if err := decodeJSON(r, &body); err != nil {
		writeWatchlistError(w, http.StatusBadRequest, "invalid request body", "VALIDATION_ERROR", nil)
		return
	}

	var statusPtr *entity.ItemStatus
	if body.Status != nil {
		if *body.Status != "ACTIVE" && *body.Status != "DISABLED" {
			writeWatchlistError(w, http.StatusBadRequest, "invalid status", "VALIDATION_ERROR",
				map[string]interface{}{"field": "status", "allowed_values": []string{"ACTIVE", "DISABLED"}})
			return
		}
		s := entity.ItemStatus(*body.Status)
		statusPtr = &s
	}

	var rulesPtr *[]command.ThresholdRuleInput
	if body.ThresholdRules != nil {
		parsed, err := parseThresholdRuleInputs(*body.ThresholdRules)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, err.Error(), "WATCHLIST_INVALID_THRESHOLD", nil)
			return
		}
		rulesPtr = &parsed
	}

	out, err := h.updateItem.Handle(r.Context(), command.UpdateItemInput{
		ItemID:         itemID,
		Pinned:         body.Pinned,
		Note:           body.Note,
		Status:         statusPtr,
		ThresholdRules: rulesPtr,
		ActorID:        actor,
	})
	if err != nil {
		writeWatchlistDomainError(w, err, map[string]interface{}{"id": itemID.String()})
		return
	}

	sec, _ := h.securityPort.GetSecurityByID(r.Context(), out.Item.SecurityID.String())
	var port *PortfolioDescriptor
	if out.Item.PortfolioID != nil {
		if pInfo, _ := h.portfolioPort.GetPortfolioScope(r.Context(), *out.Item.PortfolioID); pInfo != nil {
			port = portfolioDescriptorFromScope(pInfo)
		}
	}
	users := h.lookupUsers(r.Context(), out.Item.CreatedBy)
	var createdByUser *UserDescriptor
	if info, ok := users[out.Item.CreatedBy]; ok {
		createdByUser = &UserDescriptor{UserID: info.UserID.String(), DisplayName: info.DisplayName}
	}

	httputil.OK(w, watchlistItemResponse(out.Item, out.Rules, true, securityDescriptorFromInfo(sec), port, nil, createdByUser))
}

// ─── D. Delete Watchlist Item ─────────────────────────────────────────────────

// DeleteItem handles DELETE /watchlists/items/{id}.
// @Summary Delete a watchlist item
// @Tags Watchlist
// @Security BearerAuth
// @Produce json
// @Param id path string true "Item UUID"
// @Success 204
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /watchlists/items/{id} [delete]
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeWatchlistError(w, http.StatusUnauthorized, "not authenticated", "UNAUTHORIZED", nil)
		return
	}

	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeWatchlistError(w, http.StatusBadRequest, "invalid item id", "VALIDATION_ERROR",
			map[string]interface{}{"field": "id"})
		return
	}

	if err := h.deleteItem.Handle(r.Context(), itemID, actor); err != nil {
		writeWatchlistDomainError(w, err, map[string]interface{}{"id": itemID.String()})
		return
	}

	httputil.NoContent(w)
}

// ─── E. List Alert Events ─────────────────────────────────────────────────────

// ListAlerts handles GET /watchlists/alerts.
// @Summary List alert events
// @Tags Watchlist
// @Security BearerAuth
// @Produce json
// @Param scope_type query string false "PERSONAL or PORTFOLIO"
// @Param portfolio_id query string false "Portfolio UUID"
// @Param security_id query string false "Canonical security UUID"
// @Param rule_id query string false "Threshold rule UUID"
// @Param acknowledged query bool false "Filter by acknowledgement state"
// @Param created_from query string false "RFC3339 inclusive lower bound"
// @Param created_to query string false "RFC3339 inclusive upper bound"
// @Param limit query int false "Page size (1-200, default 50)"
// @Param offset query int false "Page offset (default 0)"
// @Success 200 {object} httputil.SuccessResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Router /watchlists/alerts [get]
func (h *Handler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeWatchlistError(w, http.StatusUnauthorized, "not authenticated", "UNAUTHORIZED", nil)
		return
	}

	q := r.URL.Query()

	var scopeType *entity.ScopeType
	if s := q.Get("scope_type"); s != "" {
		if s != "PERSONAL" && s != "PORTFOLIO" {
			writeWatchlistError(w, http.StatusBadRequest, "invalid scope_type", "VALIDATION_ERROR",
				map[string]interface{}{"field": "scope_type", "allowed_values": []string{"PERSONAL", "PORTFOLIO"}})
			return
		}
		st := entity.ScopeType(s)
		scopeType = &st
	}

	var portfolioID *uuid.UUID
	if s := q.Get("portfolio_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid portfolio_id", "VALIDATION_ERROR",
				map[string]interface{}{"field": "portfolio_id"})
			return
		}
		portfolioID = &id
	}
	if portfolioID != nil && (scopeType == nil || *scopeType != entity.ScopePortfolio) {
		writeWatchlistError(w, http.StatusBadRequest, "portfolio_id requires scope_type=PORTFOLIO", "WATCHLIST_INVALID_QUERY",
			map[string]interface{}{"field": "portfolio_id", "required_scope_type": "PORTFOLIO"})
		return
	}

	var securityID *uuid.UUID
	if s := q.Get("security_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid security_id", "VALIDATION_ERROR",
				map[string]interface{}{"field": "security_id"})
			return
		}
		securityID = &id
	}

	var ruleID *uuid.UUID
	if s := q.Get("rule_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid rule_id", "VALIDATION_ERROR",
				map[string]interface{}{"field": "rule_id"})
			return
		}
		ruleID = &id
	}

	var acknowledged *bool
	if s := q.Get("acknowledged"); s != "" {
		b := s == "true"
		acknowledged = &b
	}

	var createdFrom, createdTo *time.Time
	if s := q.Get("created_from"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid created_from timestamp", "VALIDATION_ERROR",
				map[string]interface{}{"field": "created_from", "expected_format": "RFC3339"})
			return
		}
		createdFrom = &t
	}
	if s := q.Get("created_to"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid created_to timestamp", "VALIDATION_ERROR",
				map[string]interface{}{"field": "created_to", "expected_format": "RFC3339"})
			return
		}
		createdTo = &t
	}

	limit, offset := pagination(q.Get("limit"), q.Get("offset"))

	out, err := h.listAlerts.Handle(r.Context(), query.ListAlertsInput{
		ActorID:      actor,
		ScopeType:    scopeType,
		PortfolioID:  portfolioID,
		SecurityID:   securityID,
		RuleID:       ruleID,
		Acknowledged: acknowledged,
		CreatedFrom:  createdFrom,
		CreatedTo:    createdTo,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		writeWatchlistDomainError(w, err, nil)
		return
	}

	// Batch-lookup user IDs referenced in alert events.
	var alertUserIDs []uuid.UUID
	for _, ev := range out.Items {
		alertUserIDs = append(alertUserIDs, ev.CreatedByUserID)
		if ev.AcknowledgedBy != nil {
			alertUserIDs = append(alertUserIDs, *ev.AcknowledgedBy)
		}
	}
	alertUsers := h.lookupUsers(r.Context(), alertUserIDs...)

	alerts := make([]AlertEventResponse, 0, len(out.Items))
	for _, ev := range out.Items {
		var sec *SecurityDescriptor
		if secInfo, _ := h.securityPort.GetSecurityByID(r.Context(), ev.SecurityID.String()); secInfo != nil {
			sec = securityDescriptorFromInfo(secInfo)
		}
		var port *PortfolioDescriptor
		if ev.PortfolioID != nil {
			if pInfo, _ := h.portfolioPort.GetPortfolioScope(r.Context(), *ev.PortfolioID); pInfo != nil {
				port = portfolioDescriptorFromScope(pInfo)
			}
		}
		var createdByUser, acknowledgedByUser *UserDescriptor
		if info, ok := alertUsers[ev.CreatedByUserID]; ok {
			createdByUser = &UserDescriptor{UserID: info.UserID.String(), DisplayName: info.DisplayName}
		}
		if ev.AcknowledgedBy != nil {
			if info, ok := alertUsers[*ev.AcknowledgedBy]; ok {
				acknowledgedByUser = &UserDescriptor{UserID: info.UserID.String(), DisplayName: info.DisplayName}
			}
		}
		alerts = append(alerts, alertEventResponse(ev, sec, port, createdByUser, acknowledgedByUser))
	}

	httputil.OK(w, ListAlertsResponseData{
		Items: alerts,
		Pagination: PaginationMeta{
			Limit:  out.Limit,
			Offset: out.Offset,
			Total:  out.Total,
		},
	})
}

// ─── F. Acknowledge Alert ─────────────────────────────────────────────────────

// AcknowledgeAlert handles POST /watchlists/alerts/{id}/acknowledge.
// @Summary Acknowledge an alert event
// @Tags Watchlist
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Alert event UUID"
// @Param payload body AcknowledgeAlertRequest false "Acknowledgement payload"
// @Success 200 {object} httputil.SuccessResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /watchlists/alerts/{id}/acknowledge [post]
func (h *Handler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeWatchlistError(w, http.StatusUnauthorized, "not authenticated", "UNAUTHORIZED", nil)
		return
	}

	alertID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeWatchlistError(w, http.StatusBadRequest, "invalid alert id", "VALIDATION_ERROR",
			map[string]interface{}{"field": "id"})
		return
	}

	var body AcknowledgeAlertRequest
	_ = decodeJSON(r, &body) // body is optional

	if body.Note != nil && len(*body.Note) > 1000 {
		writeWatchlistError(w, http.StatusBadRequest, "note exceeds maximum length of 1000 characters", "VALIDATION_ERROR",
			map[string]interface{}{"field": "note", "max_length": 1000})
		return
	}

	ev, err := h.ackAlert.Handle(r.Context(), alertID, actor, body.Note)
	if err != nil {
		writeWatchlistDomainError(w, err, map[string]interface{}{"id": alertID.String()})
		return
	}

	var sec *SecurityDescriptor
	if secInfo, _ := h.securityPort.GetSecurityByID(r.Context(), ev.SecurityID.String()); secInfo != nil {
		sec = securityDescriptorFromInfo(secInfo)
	}
	var port *PortfolioDescriptor
	if ev.PortfolioID != nil {
		if pInfo, _ := h.portfolioPort.GetPortfolioScope(r.Context(), *ev.PortfolioID); pInfo != nil {
			port = portfolioDescriptorFromScope(pInfo)
		}
	}
	var ackUserIDs []uuid.UUID
	ackUserIDs = append(ackUserIDs, ev.CreatedByUserID)
	if ev.AcknowledgedBy != nil {
		ackUserIDs = append(ackUserIDs, *ev.AcknowledgedBy)
	}
	ackUsers := h.lookupUsers(r.Context(), ackUserIDs...)
	var createdByUser, acknowledgedByUser *UserDescriptor
	if info, ok := ackUsers[ev.CreatedByUserID]; ok {
		createdByUser = &UserDescriptor{UserID: info.UserID.String(), DisplayName: info.DisplayName}
	}
	if ev.AcknowledgedBy != nil {
		if info, ok := ackUsers[*ev.AcknowledgedBy]; ok {
			acknowledgedByUser = &UserDescriptor{UserID: info.UserID.String(), DisplayName: info.DisplayName}
		}
	}

	httputil.OK(w, alertEventResponse(ev, sec, port, createdByUser, acknowledgedByUser))
}

// ─── G. Manual Evaluation ─────────────────────────────────────────────────────

// ManualEvaluate handles POST /watchlists/evaluate.
// @Summary Manually trigger rule evaluation
// @Tags Watchlist
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body EvaluateRequest false "Evaluation filter"
// @Success 200 {object} httputil.SuccessResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /watchlists/evaluate [post]
func (h *Handler) ManualEvaluate(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFromRequest(r)
	if !ok {
		writeWatchlistError(w, http.StatusUnauthorized, "not authenticated", "UNAUTHORIZED", nil)
		return
	}

	var body EvaluateRequest
	if r.ContentLength > 0 {
		if err := decodeJSON(r, &body); err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid request body", "VALIDATION_ERROR", nil)
			return
		}
	}

	filter := appsvc.EvaluateFilter{}
	if body.DryRun != nil {
		filter.DryRun = *body.DryRun
	}
	if body.ScopeType != nil {
		if *body.ScopeType != "PERSONAL" && *body.ScopeType != "PORTFOLIO" {
			writeWatchlistError(w, http.StatusBadRequest, "invalid scope_type", "VALIDATION_ERROR",
				map[string]interface{}{"field": "scope_type", "allowed_values": []string{"PERSONAL", "PORTFOLIO"}})
			return
		}
		st := entity.ScopeType(*body.ScopeType)
		filter.ScopeType = &st
	}
	if body.PortfolioID != nil {
		if filter.ScopeType == nil {
			writeWatchlistError(w, http.StatusBadRequest, "scope_type must be PORTFOLIO when portfolio_id is provided", "VALIDATION_ERROR",
				map[string]interface{}{"field": "scope_type", "required_value": "PORTFOLIO"})
			return
		}
		id, err := uuid.Parse(*body.PortfolioID)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid portfolio_id", "VALIDATION_ERROR",
				map[string]interface{}{"field": "portfolio_id"})
			return
		}
		filter.PortfolioID = &id
	}
	if body.SecurityID != nil {
		id, err := uuid.Parse(*body.SecurityID)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid security_id", "VALIDATION_ERROR",
				map[string]interface{}{"field": "security_id"})
			return
		}
		filter.SecurityID = &id
	}
	if body.ItemID != nil {
		id, err := uuid.Parse(*body.ItemID)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid item_id", "VALIDATION_ERROR",
				map[string]interface{}{"field": "item_id"})
			return
		}
		filter.ItemID = &id
	}
	if body.RuleID != nil {
		id, err := uuid.Parse(*body.RuleID)
		if err != nil {
			writeWatchlistError(w, http.StatusBadRequest, "invalid rule_id", "VALIDATION_ERROR",
				map[string]interface{}{"field": "rule_id"})
			return
		}
		filter.RuleID = &id
	}

	actorRef := actor
	out, err := h.evaluator.Evaluate(r.Context(), filter, &actorRef)
	if err != nil {
		writeWatchlistDomainError(w, err, nil)
		return
	}

	results := make([]EvaluateRuleResultResponse, 0, len(out.Results))
	for _, res := range out.Results {
		var staleReason *string
		if res.StaleReason != "" {
			sr := res.StaleReason
			staleReason = &sr
		}
		results = append(results, EvaluateRuleResultResponse{
			RuleID:             res.RuleID.String(),
			WatchlistItemID:    res.WatchlistItemID.String(),
			SecurityID:         res.SecurityID.String(),
			PreviousState:      string(res.PreviousState),
			ComputedState:      string(res.ComputedState),
			WouldCreateAlert:   res.WouldCreateAlert,
			NotificationStatus: string(res.NotificationStatus),
			QuoteStatus:        res.QuoteStatus,
			ObservedPrice:      res.ObservedPrice,
			ThresholdValue:     res.ThresholdValue,
			Stale:              res.Stale,
			StaleReason:        staleReason,
		})
	}

	httputil.OK(w, EvaluateResponseData{
		DryRun:           out.DryRun,
		RulesEvaluated:   out.RulesEvaluated,
		AlertsCreated:    out.AlertsCreated,
		AlertsSuppressed: out.AlertsSuppressed,
		RulesSkipped:     out.RulesSkipped,
		ProviderFailures: out.ProviderFailures,
		Results:          results,
	})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func actorFromRequest(r *http.Request) (uuid.UUID, bool) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func pagination(limitStr, offsetStr string) (int, int) {
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func writeWatchlistError(w http.ResponseWriter, status int, message, code string, details interface{}) {
	httputil.JSON(w, status, httputil.ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	})
}

func writeWatchlistDomainError(w http.ResponseWriter, err error, details interface{}) {
	switch {
	case errors.Is(err, watchlistdomain.ErrItemNotFound):
		writeWatchlistError(w, http.StatusNotFound, "watchlist item not found", "WATCHLIST_ITEM_NOT_FOUND", details)
	case errors.Is(err, watchlistdomain.ErrAlertNotFound):
		writeWatchlistError(w, http.StatusNotFound, "alert event not found", "WATCHLIST_ALERT_NOT_FOUND", details)
	case errors.Is(err, watchlistdomain.ErrForbiddenScope):
		writeWatchlistError(w, http.StatusForbidden, "access is forbidden for this scope", "WATCHLIST_FORBIDDEN_SCOPE", details)
	case errors.Is(err, watchlistdomain.ErrDuplicateWatchlistItem):
		writeWatchlistError(w, http.StatusConflict, "watchlist item already exists for this scope and security", "WATCHLIST_DUPLICATE_ITEM", details)
	case errors.Is(err, watchlistdomain.ErrInvalidQuery):
		writeWatchlistError(w, http.StatusBadRequest, err.Error(), "WATCHLIST_INVALID_QUERY", details)
	case errors.Is(err, watchlistdomain.ErrInvalidSecurity):
		writeWatchlistError(w, http.StatusUnprocessableEntity, "invalid or inactive security", "WATCHLIST_INVALID_SECURITY", details)
	case errors.Is(err, watchlistdomain.ErrInvalidThreshold):
		writeWatchlistError(w, http.StatusBadRequest, err.Error(), "WATCHLIST_INVALID_THRESHOLD", details)
	case errors.Is(err, watchlistdomain.ErrAlertAlreadyAcknowledged):
		writeWatchlistError(w, http.StatusConflict, "alert is already acknowledged", "WATCHLIST_ALERT_ALREADY_ACKNOWLEDGED", details)
	case errors.Is(err, watchlistdomain.ErrRuleDisabled):
		writeWatchlistError(w, http.StatusConflict, "rule is disabled", "WATCHLIST_RULE_DISABLED", details)
	case errors.Is(err, watchlistdomain.ErrDuplicateThresholdRule):
		writeWatchlistError(w, http.StatusConflict, "duplicate threshold rule", "WATCHLIST_DUPLICATE_THRESHOLD", details)
	case errors.Is(err, watchlistdomain.ErrStaleQuote):
		writeWatchlistError(w, http.StatusConflict, "quote data is stale and cannot be used for evaluation", "WATCHLIST_STALE_QUOTE", details)
	case errors.Is(err, watchlistdomain.ErrProviderUnavailable):
		writeWatchlistError(w, http.StatusServiceUnavailable, "market data provider is temporarily unavailable", "WATCHLIST_PROVIDER_UNAVAILABLE", details)
	default:
		writeWatchlistError(w, http.StatusInternalServerError, "an unexpected error occurred", "INTERNAL_ERROR", nil)
	}
}

// lookupUsers batch-fetches user display info for the given IDs (best-effort).
// Returns a nil map when userLookup is not configured or the lookup fails.
func (h *Handler) lookupUsers(ctx context.Context, ids ...uuid.UUID) map[uuid.UUID]*watchlistdomain.UserInfo {
	if h.userLookup == nil || len(ids) == 0 {
		return nil
	}
	// Deduplicate nil/zero UUIDs.
	seen := make(map[uuid.UUID]struct{}, len(ids))
	var uniq []uuid.UUID
	for _, id := range ids {
		if id != uuid.Nil {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				uniq = append(uniq, id)
			}
		}
	}
	users, _ := h.userLookup.GetUsersByIDs(ctx, uniq)
	return users
}

func parseThresholdRuleInputs(reqs []ThresholdRuleRequest) ([]command.ThresholdRuleInput, error) {
	result := make([]command.ThresholdRuleInput, 0, len(reqs))
	for i, req := range reqs {
		if req.MetricType != nil && *req.MetricType != "" && *req.MetricType != "MARKET_PRICE" {
			return nil, errors.New("rules[" + strconv.Itoa(i) + "].metric_type must be MARKET_PRICE")
		}
		if req.Direction != "ABOVE" && req.Direction != "BELOW" {
			return nil, errors.New("rules[" + strconv.Itoa(i) + "].direction must be ABOVE or BELOW")
		}
		val, err := decimal.NewFromString(req.ThresholdValue)
		if err != nil || !val.IsPositive() {
			return nil, errors.New("rules[" + strconv.Itoa(i) + "].threshold_value must be a positive decimal string")
		}

		cooldown := 60
		if req.CooldownMinutes != nil {
			cooldown = *req.CooldownMinutes
			if cooldown < 0 {
				return nil, errors.New("rules[" + strconv.Itoa(i) + "].cooldown_minutes must be >= 0")
			}
		}
		status := entity.RuleStatusEnabled
		if req.Status != nil {
			switch *req.Status {
			case "ENABLED":
				status = entity.RuleStatusEnabled
			case "DISABLED":
				status = entity.RuleStatusDisabled
			default:
				return nil, errors.New("rules[" + strconv.Itoa(i) + "].status must be ENABLED or DISABLED")
			}
		}

		var ruleID *uuid.UUID
		if req.ID != nil && *req.ID != "" {
			id, err := uuid.Parse(*req.ID)
			if err != nil {
				return nil, errors.New("rules[" + strconv.Itoa(i) + "].id must be a valid UUID")
			}
			ruleID = &id
		}

		result = append(result, command.ThresholdRuleInput{
			ID:              ruleID,
			Direction:       entity.Direction(req.Direction),
			ThresholdValue:  val,
			Currency:        req.Currency,
			CooldownMinutes: cooldown,
			Status:          status,
		})
	}
	return result, nil
}
