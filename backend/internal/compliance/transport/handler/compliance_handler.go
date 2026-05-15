// Package handler contains HTTP handlers for the compliance / IRG module.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	platformmw "github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// ComplianceHandler bundles all IRG HTTP endpoints.
type ComplianceHandler struct {
	preTradeCmd       *command.RunPreTradeCheckHandler
	postTradeCmd      *command.RunPostTradeCheckHandler
	overrideCmd       *command.OverrideBreachHandler
	createInstanceCmd *command.CreateRuleInstanceHandler
	checkGroupQry     *query.GetCheckGroupHandler
	listBreachesQry   *query.ListBreachesHandler
	listInstancesQry  *query.ListRuleInstancesHandler
}

// NewComplianceHandler wires all handlers together.
func NewComplianceHandler(
	preTradeCmd *command.RunPreTradeCheckHandler,
	postTradeCmd *command.RunPostTradeCheckHandler,
	overrideCmd *command.OverrideBreachHandler,
	createInstanceCmd *command.CreateRuleInstanceHandler,
	checkGroupQry *query.GetCheckGroupHandler,
	listBreachesQry *query.ListBreachesHandler,
	listInstancesQry *query.ListRuleInstancesHandler,
) *ComplianceHandler {
	return &ComplianceHandler{
		preTradeCmd:       preTradeCmd,
		postTradeCmd:      postTradeCmd,
		overrideCmd:       overrideCmd,
		createInstanceCmd: createInstanceCmd,
		checkGroupQry:     checkGroupQry,
		listBreachesQry:   listBreachesQry,
		listInstancesQry:  listInstancesQry,
	}
}

// ============================================================
// POST /compliance/checks/pre-trade
// ============================================================

type PreTradeRequest struct {
	PortfolioID  string `json:"portfolio_id"`
	ContractID   string `json:"contract_id"`
	BusinessDate string `json:"business_date"` // "2006-01-02"
	OrderID      string `json:"order_id"`
	Ticker       string `json:"ticker"`
	Side         string `json:"side"` // BUY | SELL
	Quantity     string `json:"quantity"`
	Price        string `json:"price"`
	Currency     string `json:"currency"`
	Exchange     string `json:"exchange"`
}

// RunPreTradeCheck handles POST /compliance/checks/pre-trade.
// @Summary Run Pre-Trade Compliance Check
// @Description Evaluate compliance rules for a proposed order before execution.
// @Tags Compliance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body PreTradeRequest true "Pre-trade check payload"
// @Success 201 {object} command.PreTradeCheckResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /compliance/checks/pre-trade [post]
func (h *ComplianceHandler) RunPreTradeCheck(w http.ResponseWriter, r *http.Request) {
	var req PreTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body: "+err.Error())
		return
	}

	portfolioID, err := uuid.Parse(req.PortfolioID)
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio_id")
		return
	}
	contractID, err := uuid.Parse(req.ContractID)
	if err != nil {
		httputil.BadRequest(w, "invalid contract_id")
		return
	}
	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		httputil.BadRequest(w, "invalid order_id")
		return
	}
	bizDate, err := time.Parse("2006-01-02", req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, "invalid business_date format, expected YYYY-MM-DD")
		return
	}
	qty, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		httputil.BadRequest(w, "invalid quantity")
		return
	}
	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		httputil.BadRequest(w, "invalid price")
		return
	}

	actor := actorFromCtx(r)

	resp, err := h.preTradeCmd.Handle(r.Context(), command.PreTradeCheckRequest{
		CheckGroupID: uuid.New(),
		PortfolioID:  portfolioID,
		ContractID:   contractID,
		BusinessDate: bizDate,
		Actor:        actor,
		OrderID:      orderID,
		Ticker:       req.Ticker,
		Side:         vo.OrderSide(req.Side),
		Quantity:     qty,
		Price:        price,
		Currency:     req.Currency,
		Exchange:     req.Exchange,
	})
	if err != nil {
		httputil.InternalError(w, "pre-trade check failed: "+err.Error())
		return
	}

	// A BLOCK verdict is communicated via the response body, not HTTP 4xx.
	httputil.Created(w, resp)
}

// ============================================================
// POST /compliance/checks/post-trade
// ============================================================

type PostTradeRequest struct {
	PortfolioID  string `json:"portfolio_id"`
	ContractID   string `json:"contract_id"`
	BusinessDate string `json:"business_date"`
}

// RunPostTradeCheck handles POST /compliance/checks/post-trade.
// @Summary Run Post-Trade Compliance Check
// @Description Evaluate compliance rules after trade capture or during periodic replay.
// @Tags Compliance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body PostTradeRequest true "Post-trade check payload"
// @Success 201 {object} command.PostTradeCheckResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /compliance/checks/post-trade [post]
func (h *ComplianceHandler) RunPostTradeCheck(w http.ResponseWriter, r *http.Request) {
	var req PostTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body: "+err.Error())
		return
	}

	portfolioID, err := uuid.Parse(req.PortfolioID)
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio_id")
		return
	}
	contractID, err := uuid.Parse(req.ContractID)
	if err != nil {
		httputil.BadRequest(w, "invalid contract_id")
		return
	}
	bizDate, err := time.Parse("2006-01-02", req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, "invalid business_date format, expected YYYY-MM-DD")
		return
	}

	resp, err := h.postTradeCmd.Handle(r.Context(), command.PostTradeCheckRequest{
		PortfolioID:  portfolioID,
		ContractID:   contractID,
		BusinessDate: bizDate,
		Actor:        actorFromCtx(r),
	})
	if err != nil {
		httputil.InternalError(w, "post-trade check failed: "+err.Error())
		return
	}

	httputil.Created(w, resp)
}

// ============================================================
// GET /compliance/checks/{groupID}
// ============================================================

// GetCheckGroup handles GET /compliance/checks/{groupID}.
// @Summary Get Compliance Check Group
// @Description Retrieve check records and breaches associated with one compliance check group.
// @Tags Compliance
// @Security BearerAuth
// @Produce json
// @Param groupID path string true "Check group UUID"
// @Success 200 {object} query.CheckGroupResult
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /compliance/checks/{groupID} [get]
func (h *ComplianceHandler) GetCheckGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(chi.URLParam(r, "groupID"))
	if err != nil {
		httputil.BadRequest(w, "invalid group_id")
		return
	}

	result, err := h.checkGroupQry.Handle(r.Context(), groupID)
	if err != nil {
		httputil.InternalError(w, "failed to fetch check group")
		return
	}
	if result == nil {
		httputil.NotFound(w, "check group not found")
		return
	}
	httputil.OK(w, result)
}

// ============================================================
// GET /compliance/breaches
// ============================================================

// ListBreaches handles GET /compliance/breaches.
// @Summary List Compliance Breaches
// @Description List compliance breaches with optional portfolio, contract, rule, status, and date filters.
// @Tags Compliance
// @Security BearerAuth
// @Produce json
// @Param portfolio_id query string false "Portfolio UUID"
// @Param contract_id query string false "Contract UUID"
// @Param status query string false "Breach status"
// @Param rule_type_id query string false "Rule type ID"
// @Param date_from query string false "Start business date (YYYY-MM-DD)"
// @Param date_to query string false "End business date (YYYY-MM-DD)"
// @Param offset query int false "Offset (default 0)"
// @Param limit query int false "Limit (default 50, max 200)"
// @Success 200 {object} query.ListBreachesResult
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /compliance/breaches [get]
func (h *ComplianceHandler) ListBreaches(w http.ResponseWriter, r *http.Request) {
	req := query.ListBreachesRequest{
		Offset: parseIntParam(r, "offset", 0),
		Limit:  parseIntParam(r, "limit", 50),
	}

	if v := r.URL.Query().Get("portfolio_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.PortfolioID = &id
		}
	}
	if v := r.URL.Query().Get("contract_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			req.ContractID = &id
		}
	}
	if v := r.URL.Query().Get("status"); v != "" {
		s := entity.BreachStatus(v)
		req.Status = &s
	}
	if v := r.URL.Query().Get("rule_type_id"); v != "" {
		req.RuleTypeID = v
	}
	if v := r.URL.Query().Get("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			req.DateFrom = &t
		}
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			req.DateTo = &t
		}
	}

	result, err := h.listBreachesQry.Handle(r.Context(), req)
	if err != nil {
		httputil.InternalError(w, "failed to list breaches")
		return
	}
	httputil.OK(w, result)
}

// ============================================================
// POST /compliance/breaches/{breachID}/override
// ============================================================

type OverrideRequest struct {
	Reason     string `json:"reason"`
	ApprovedBy string `json:"approved_by,omitempty"`
}

// OverrideBreach handles POST /compliance/breaches/{breachID}/override.
// @Summary Override Compliance Breach
// @Description Override an open, overridable compliance breach with an audit reason.
// @Tags Compliance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param breachID path string true "Breach UUID"
// @Param request body OverrideRequest true "Override payload"
// @Success 201 {object} entity.Override
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /compliance/breaches/{breachID}/override [post]
func (h *ComplianceHandler) OverrideBreach(w http.ResponseWriter, r *http.Request) {
	breachID, err := uuid.Parse(chi.URLParam(r, "breachID"))
	if err != nil {
		httputil.BadRequest(w, "invalid breach_id")
		return
	}

	var req OverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body: "+err.Error())
		return
	}
	if req.Reason == "" {
		httputil.BadRequest(w, "reason is required")
		return
	}

	claims := platformmw.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "authentication required")
		return
	}
	actorUUID, err := uuid.Parse(claims.Subject)
	if err != nil {
		httputil.Unauthorized(w, "invalid user claims")
		return
	}

	overrideReq := command.OverrideBreachRequest{
		BreachID:     breachID,
		Reason:       req.Reason,
		OverriddenBy: actorUUID,
	}
	if req.ApprovedBy != "" {
		if id, err := uuid.Parse(req.ApprovedBy); err == nil {
			overrideReq.ApprovedBy = &id
		}
	}

	override, err := h.overrideCmd.Handle(r.Context(), overrideReq)
	if err != nil {
		writeOverrideError(w, err)
		return
	}
	httputil.Created(w, override)
}

// writeOverrideError maps compliance-override errors to HTTP responses.
//
// Every case is explicit — the default branch handles genuine internal
// failures (DB connection errors, tx begin/commit failures, context deadline,
// unexpected pgx errors) and returns 500, not 400. Misclassifying a server-side
// failure as a client error would make operators miss real outages during
// incident response.
//
//   - *ErrInvalidOverrideRequest → 400 (client-fixable input problem).
//   - *ErrBreachNotFound         → 404.
//   - *ErrBreachNotOpen          → 409 (state conflict; breach no longer OPEN).
//   - *ErrOverrideAlreadyExists  → 409 (duplicate; concurrent request won).
//   - anything else              → 500, and the raw error message is NOT
//     forwarded (it may leak internal details); we log-and-return a
//     generic message instead. Logging is left to the platform logger
//     middleware which already captures handler errors on the request.
func writeOverrideError(w http.ResponseWriter, err error) {
	var invalid *domain.ErrInvalidOverrideRequest
	var notFound *domain.ErrBreachNotFound
	var notOpen *domain.ErrBreachNotOpen
	var dup *domain.ErrOverrideAlreadyExists
	switch {
	case errors.As(err, &invalid):
		httputil.BadRequest(w, err.Error())
	case errors.As(err, &notFound):
		httputil.NotFound(w, err.Error())
	case errors.As(err, &notOpen):
		httputil.Conflict(w, err.Error())
	case errors.As(err, &dup):
		httputil.Conflict(w, err.Error())
	default:
		// Server-side failure: persist the detail in server logs via the
		// recovery/logger middleware, but do not leak internals to the client.
		httputil.InternalError(w, "failed to commit override")
	}
}

// ============================================================
// POST /compliance/rules
// ============================================================

// CreateRuleInstanceRequest is the JSON body accepted by RunCreateRuleInstance.
//
// Field naming follows the snake_case convention used across the compliance
// API. `parameters` is a raw JSON document — the handler does not decode it
// beyond validating that it is well-formed JSON; the application layer then
// checks it against the rule type's ParameterSchema.
type CreateRuleInstanceRequest struct {
	RuleTypeID    string          `json:"rule_type_id"`
	Name          string          `json:"name"`
	Description   string          `json:"description,omitempty"`
	Parameters    json.RawMessage `json:"parameters" swaggertype:"object"`
	EffectiveFrom string          `json:"effective_from"`          // "2006-01-02"
	EffectiveTo   *string         `json:"effective_to,omitempty"`  // "2006-01-02" or null
	IsActive      *bool           `json:"is_active,omitempty"`     // defaults to true
	ChangeReason  string          `json:"change_reason,omitempty"` // "initial creation" if blank
}

// CreateRuleInstance handles POST /compliance/rules.
//
// Permission: IRG_EDIT_RULE_INSTANCE (enforced at the router). The actor UUID
// is taken from the auth context — never from the request body — so callers
// cannot impersonate another user when stamping created_by.
// @Summary Create Compliance Rule Instance
// @Description Create a compliance rule instance and its initial parameter version.
// @Tags Compliance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateRuleInstanceRequest true "Rule instance payload"
// @Success 201 {object} command.CreateRuleInstanceResult
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /compliance/rules [post]
func (h *ComplianceHandler) CreateRuleInstance(w http.ResponseWriter, r *http.Request) {
	var req CreateRuleInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid request body: "+err.Error())
		return
	}

	claims := platformmw.GetUserClaims(r.Context())
	if claims == nil {
		httputil.Unauthorized(w, "authentication required")
		return
	}
	actorID, err := uuid.Parse(claims.Subject)
	if err != nil {
		httputil.Unauthorized(w, "invalid user claims")
		return
	}

	from, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		httputil.BadRequest(w, "invalid effective_from, expected YYYY-MM-DD")
		return
	}
	var to *time.Time
	if req.EffectiveTo != nil && *req.EffectiveTo != "" {
		parsed, err := time.Parse("2006-01-02", *req.EffectiveTo)
		if err != nil {
			httputil.BadRequest(w, "invalid effective_to, expected YYYY-MM-DD")
			return
		}
		to = &parsed
	}

	// Default IsActive to true when the caller omits the field — a common
	// case for typical rule admin flows. Explicit false stages the instance
	// for later activation.
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	result, err := h.createInstanceCmd.Handle(r.Context(), command.CreateRuleInstanceRequest{
		RuleTypeID:    req.RuleTypeID,
		Name:          req.Name,
		Description:   req.Description,
		Parameters:    req.Parameters,
		EffectiveFrom: from,
		EffectiveTo:   to,
		IsActive:      active,
		ChangeReason:  req.ChangeReason,
		CreatedBy:     actorID,
	})
	if err != nil {
		writeCreateRuleInstanceError(w, err)
		return
	}

	httputil.Created(w, result)
}

// writeCreateRuleInstanceError maps domain errors to HTTP responses.
//
// Explicit branches — the default returns 500 with a generic message to avoid
// leaking internal DB details; the typed errors above are client-safe.
//
//   - *ErrInvalidCreateRuleRequest → 400 (missing/bad field).
//   - *ErrRuleTypeNotFound         → 422 (caller referenced an unregistered rule type).
//   - *ErrParameterValidation      → 400 (parameters fail the rule's JSON Schema).
//   - anything else                → 500.
func writeCreateRuleInstanceError(w http.ResponseWriter, err error) {
	var invalid *command.ErrInvalidCreateRuleRequest
	var noType *domain.ErrRuleTypeNotFound
	var badParams *domain.ErrParameterValidation
	switch {
	case errors.As(err, &invalid):
		httputil.BadRequest(w, err.Error())
	case errors.As(err, &noType):
		httputil.UnprocessableEntity(w, err.Error())
	case errors.As(err, &badParams):
		httputil.BadRequest(w, err.Error())
	default:
		httputil.InternalError(w, "failed to create rule instance")
	}
}

// ============================================================
// GET /compliance/rules
// ============================================================

// ListRuleInstances handles GET /compliance/rules.
// @Summary List Compliance Rule Instances
// @Description List configured compliance rule instances with optional type and active filters.
// @Tags Compliance
// @Security BearerAuth
// @Produce json
// @Param rule_type_id query string false "Rule type ID"
// @Param is_active query bool false "Filter by active flag"
// @Param offset query int false "Offset (default 0)"
// @Param limit query int false "Limit (default 50, max 200)"
// @Success 200 {object} query.ListRuleInstancesResult
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /compliance/rules [get]
func (h *ComplianceHandler) ListRuleInstances(w http.ResponseWriter, r *http.Request) {
	req := query.ListRuleInstancesRequest{
		Offset: parseIntParam(r, "offset", 0),
		Limit:  parseIntParam(r, "limit", 50),
	}
	if v := r.URL.Query().Get("rule_type_id"); v != "" {
		req.RuleTypeID = &v
	}
	if v := r.URL.Query().Get("is_active"); v != "" {
		active := v == "true"
		req.IsActive = &active
	}

	result, err := h.listInstancesQry.Handle(r.Context(), req)
	if err != nil {
		httputil.InternalError(w, "failed to list rule instances")
		return
	}
	httputil.OK(w, result)
}

// ============================================================
// helpers
// ============================================================

func actorFromCtx(r *http.Request) string {
	claims := platformmw.GetUserClaims(r.Context())
	if claims != nil {
		return claims.Subject
	}
	return "system"
}

func parseIntParam(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}
