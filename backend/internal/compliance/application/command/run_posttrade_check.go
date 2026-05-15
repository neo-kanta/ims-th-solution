package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/engine"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// PostTradeCheckRequest triggers a post-trade or periodic compliance scan.
type PostTradeCheckRequest struct {
	CheckGroupID uuid.UUID // optional; generated if zero
	PortfolioID  uuid.UUID
	ContractID   uuid.UUID
	BusinessDate time.Time
	Actor        string
}

// PostTradeCheckResponse is returned from the post-trade handler.
type PostTradeCheckResponse struct {
	CheckGroupID    uuid.UUID       `json:"check_group_id"`
	Verdict         vo.Verdict      `json:"verdict" swaggertype:"string"`
	RulesEvaluated  int             `json:"rules_evaluated"`
	TotalDurationMs int64           `json:"total_duration_ms"`
	Breaches        []BreachSummary `json:"breaches,omitempty"`
}

// RunPostTradeCheckHandler executes post-trade / periodic compliance scans.
// These are informational: a BLOCK verdict raises a breach for investigation
type RunPostTradeCheckHandler struct {
	pipeline *engine.Pipeline
	registry *spi.RuleRegistry
}

// NewRunPostTradeCheckHandler creates the handler.
func NewRunPostTradeCheckHandler(pipeline *engine.Pipeline, registry *spi.RuleRegistry) *RunPostTradeCheckHandler {
	return &RunPostTradeCheckHandler{pipeline: pipeline, registry: registry}
}

// Handle executes a portfolio-level post-trade check.
func (h *RunPostTradeCheckHandler) Handle(ctx context.Context, req PostTradeCheckRequest) (*PostTradeCheckResponse, error) {
	if req.PortfolioID == uuid.Nil {
		return nil, fmt.Errorf("portfolio_id is required")
	}
	if req.ContractID == uuid.Nil {
		return nil, fmt.Errorf("contract_id is required")
	}
	if req.BusinessDate.IsZero() {
		return nil, fmt.Errorf("business_date is required")
	}

	checkGroupID := req.CheckGroupID
	if checkGroupID == uuid.Nil {
		checkGroupID = uuid.New()
	}

	input := spi.CheckInput{
		CheckGroupID:  checkGroupID,
		Timing:        vo.TimingPostTrade,
		BusinessDate:  req.BusinessDate.UTC(),
		Actor:         req.Actor,
		PortfolioID:   req.PortfolioID,
		ContractID:    req.ContractID,
		ProposedOrder: nil,
	}

	scopes := buildScopes(req.PortfolioID, req.ContractID)

	output, err := h.pipeline.RunCheck(ctx, input, scopes)
	if err != nil {
		return nil, fmt.Errorf("running post-trade check: %w", err)
	}

	resp := &PostTradeCheckResponse{
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
