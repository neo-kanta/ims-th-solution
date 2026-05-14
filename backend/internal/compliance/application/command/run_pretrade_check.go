// Package command contains write-side application use cases for the compliance module.
package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/engine"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// PreTradeCheckRequest is the inbound request from the order management layer.
type PreTradeCheckRequest struct {
	CheckGroupID uuid.UUID
	PortfolioID  uuid.UUID
	ContractID   uuid.UUID
	BusinessDate time.Time
	Actor        string

	OrderID  uuid.UUID
	Ticker   string
	Side     vo.OrderSide
	Quantity decimal.Decimal
	Price    decimal.Decimal
	Currency string
	Exchange string
}

// PreTradeCheckResponse is the result returned to the OMS / caller.
type PreTradeCheckResponse struct {
	CheckGroupID    uuid.UUID       `json:"check_group_id"`
	Verdict         vo.Verdict      `json:"verdict" swaggertype:"string"`
	RulesEvaluated  int             `json:"rules_evaluated"`
	TotalDurationMs int64           `json:"total_duration_ms"`
	Breaches        []BreachSummary `json:"breaches,omitempty"`
}

// BreachSummary is a compact representation for the caller.
type BreachSummary struct {
	BreachID    uuid.UUID   `json:"breach_id"`
	RuleTypeID  string      `json:"rule_type_id"`
	Severity    vo.Severity `json:"severity" swaggertype:"string"`
	Verdict     vo.Verdict  `json:"verdict" swaggertype:"string"`
	Message     string      `json:"message"`
	Overridable bool        `json:"overridable"`
}

// RunPreTradeCheckHandler executes the pre-trade compliance pipeline.
type RunPreTradeCheckHandler struct {
	pipeline *engine.Pipeline
	registry *spi.RuleRegistry
}

// NewRunPreTradeCheckHandler creates the handler.
func NewRunPreTradeCheckHandler(pipeline *engine.Pipeline, registry *spi.RuleRegistry) *RunPreTradeCheckHandler {
	return &RunPreTradeCheckHandler{pipeline: pipeline, registry: registry}
}

// Handle runs the pre-trade check and returns a verdict.
// Returns an error only for infrastructure failures (DB, network).
// A BLOCK verdict is returned in the response, not as an error.
func (h *RunPreTradeCheckHandler) Handle(ctx context.Context, req PreTradeCheckRequest) (*PreTradeCheckResponse, error) {
	if err := validatePreTradeRequest(req); err != nil {
		return nil, fmt.Errorf("invalid pre-trade request: %w", err)
	}

	checkGroupID := req.CheckGroupID
	if checkGroupID == uuid.Nil {
		checkGroupID = uuid.New()
	}

	input := spi.CheckInput{
		CheckGroupID: checkGroupID,
		Timing:       vo.TimingPreTrade,
		BusinessDate: req.BusinessDate.UTC(),
		Actor:        req.Actor,
		PortfolioID:  req.PortfolioID,
		ContractID:   req.ContractID,
		ProposedOrder: &spi.ProposedOrder{
			OrderID:  req.OrderID,
			Ticker:   req.Ticker,
			Side:     req.Side,
			Quantity: req.Quantity,
			Price:    req.Price,
			Currency: req.Currency,
			Exchange: req.Exchange,
		},
	}

	scopes := buildScopes(req.PortfolioID, req.ContractID)

	output, err := h.pipeline.RunCheck(ctx, input, scopes)
	if err != nil {
		return nil, fmt.Errorf("running pre-trade check: %w", err)
	}

	resp := &PreTradeCheckResponse{
		CheckGroupID:    checkGroupID,
		Verdict:         output.FinalVerdict,
		RulesEvaluated:  output.RulesEvaluated,
		TotalDurationMs: output.TotalDurationMs,
	}

	for _, b := range output.Breaches {
		overridable := true
		if meta, ok := h.registry.Get(b.RuleTypeID); ok {
			overridable = meta.Metadata().Overridable
		}
		resp.Breaches = append(resp.Breaches, BreachSummary{
			BreachID:    b.ID,
			RuleTypeID:  b.RuleTypeID,
			Severity:    b.Severity,
			Verdict:     b.Verdict,
			Message:     b.Message,
			Overridable: overridable,
		})
	}

	return resp, nil
}

// --- helpers ---

func validatePreTradeRequest(req PreTradeCheckRequest) error {
	if req.PortfolioID == uuid.Nil {
		return fmt.Errorf("portfolio_id is required")
	}
	if req.ContractID == uuid.Nil {
		return fmt.Errorf("contract_id is required")
	}
	if req.Ticker == "" {
		return fmt.Errorf("ticker is required")
	}
	if req.Quantity.IsZero() || req.Quantity.IsNegative() {
		return fmt.Errorf("quantity must be positive")
	}
	if req.Price.IsZero() || req.Price.IsNegative() {
		return fmt.Errorf("price must be positive")
	}
	if req.BusinessDate.IsZero() {
		return fmt.Errorf("business_date is required")
	}
	return nil
}

// buildScopes returns scopes from most-specific to least-specific.
// The pipeline sorts them further by Specificity() anyway.
func buildScopes(portfolioID, contractID uuid.UUID) []vo.Scope {
	scopes := []vo.Scope{
		{Type: vo.ScopeGlobal, ID: nil},
	}
	if contractID != uuid.Nil {
		cid := contractID
		scopes = append(scopes, vo.Scope{Type: vo.ScopeContract, ID: &cid})
	}
	if portfolioID != uuid.Nil {
		pid := portfolioID
		scopes = append(scopes, vo.Scope{Type: vo.ScopePortfolio, ID: &pid})
	}
	return scopes
}
