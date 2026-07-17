package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

type ExecutionHandler struct {
	executions domain.ExecutionRepository
	decisions  domain.DecisionRepository
	portfolios domain.PortfolioRepository
	cmd        *command.ExecutionCommandHandler
}

func NewExecutionHandler(repo domain.ExecutionRepository, cmd *command.ExecutionCommandHandler) *ExecutionHandler {
	return &ExecutionHandler{executions: repo, cmd: cmd}
}

// SetDecisionRepository wires the decision repository post-construction so
// the Portfolio V2 (portfolioCode) route
// POST /portfolios/{portfolioCode}/decisions/{decisionId}/executions in
// portfolio_v2_execution_handler.go can verify the decision belongs to the
// resolved portfolio before creating an execution.
func (h *ExecutionHandler) SetDecisionRepository(r domain.DecisionRepository) {
	if h != nil {
		h.decisions = r
	}
}

// SetPortfolioRepository wires the portfolio repository post-construction so
// the Portfolio V2 (portfolioCode) routes can resolve portfolioCode ->
// portfolio_id.
func (h *ExecutionHandler) SetPortfolioRepository(r domain.PortfolioRepository) {
	if h != nil {
		h.portfolios = r
	}
}

// ListExecutions handles GET /investment/executions?decision_id=&fund_id=&business_date=.
func (h *ExecutionHandler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if v := q.Get("decision_id"); v != "" {
		id, err := parseUUID(v)
		if err != nil {
			httputil.BadRequest(w, "invalid decision_id")
			return
		}
		items, err := h.executions.ListByDecision(r.Context(), id)
		if err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		out := make([]response.ExecutionResponse, 0, len(items))
		for _, e := range items {
			out = append(out, response.FromExecution(e))
		}
		httputil.OK(w, map[string]any{"items": out})
		return
	}
	if v := q.Get("fund_id"); v != "" {
		id, err := parseUUID(v)
		if err != nil {
			httputil.BadRequest(w, "invalid fund_id")
			return
		}
		bd, err := parseDate(q.Get("business_date"))
		if err != nil {
			httputil.BadRequest(w, "business_date required (YYYY-MM-DD)")
			return
		}
		items, err := h.executions.ListByFundDate(r.Context(), id, bd)
		if err != nil {
			httputil.InternalError(w, err.Error())
			return
		}
		out := make([]response.ExecutionResponse, 0, len(items))
		for _, e := range items {
			out = append(out, response.FromExecution(e))
		}
		httputil.OK(w, map[string]any{"items": out})
		return
	}
	httputil.BadRequest(w, "decision_id or (fund_id+business_date) is required")
}

func (h *ExecutionHandler) GetExecution(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid execution id")
		return
	}
	e, err := h.executions.GetByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if e == nil {
		httputil.NotFound(w, "execution not found")
		return
	}
	httputil.OK(w, response.FromExecution(e))
}

func (h *ExecutionHandler) CreateExecution(w http.ResponseWriter, r *http.Request) {
	var req request.CreateExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	q, err := parseDecimalOpt(req.OrderedQuantity)
	if err != nil {
		httputil.BadRequest(w, "invalid ordered_quantity")
		return
	}
	a, err := parseDecimalOpt(req.OrderedAmount)
	if err != nil {
		httputil.BadRequest(w, "invalid ordered_amount")
		return
	}
	e, err := h.cmd.Create(r.Context(), command.CreateExecutionRequest{
		DecisionID:      req.DecisionID,
		OrderedQuantity: q,
		OrderedAmount:   a,
		BrokerReference: req.BrokerReference,
		ActorID:         actor,
	})
	if err != nil {
		writeExecutionError(w, err)
		return
	}
	httputil.Created(w, response.FromExecution(e))
}

func (h *ExecutionHandler) FillExecution(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid execution id")
		return
	}
	var req request.FillExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	q, err := parseDecimalOpt(req.ExecutedQuantity)
	if err != nil {
		httputil.BadRequest(w, "invalid executed_quantity")
		return
	}
	a, err := parseDecimalOpt(req.ExecutedAmount)
	if err != nil {
		httputil.BadRequest(w, "invalid executed_amount")
		return
	}
	p, err := parseDecimalOpt(req.ExecutionPrice)
	if err != nil {
		httputil.BadRequest(w, "invalid execution_price")
		return
	}
	status := vo.ExecutionStatus(req.Status)
	e, err := h.cmd.Fill(r.Context(), command.FillExecutionRequest{
		ExecutionID:      id,
		ExecutedQuantity: q,
		ExecutedAmount:   a,
		ExecutionPrice:   p,
		Status:           status,
		BrokerReference:  req.BrokerReference,
		ActorID:          actor,
	})
	if err != nil {
		writeExecutionError(w, err)
		return
	}
	httputil.OK(w, response.FromExecution(e))
}

func (h *ExecutionHandler) CancelExecution(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid execution id")
		return
	}
	var req request.CancelExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	e, err := h.cmd.Cancel(r.Context(), command.CancelExecutionRequest{
		ExecutionID: id, Reason: req.Reason, ActorID: actor,
	})
	if err != nil {
		writeExecutionError(w, err)
		return
	}
	httputil.OK(w, response.FromExecution(e))
}

func writeExecutionError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var (
		invalid     *domain.ErrInvalidDecisionRequest
		decision    *domain.ErrDecisionNotFound
		lifecycle   *domain.ErrDecisionLifecycle
		execNot     *domain.ErrExecutionNotFound
		execLife    *domain.ErrExecutionLifecycle
		compBlocked *domain.ErrComplianceRejected
	)
	switch {
	case errors.As(err, &invalid):
		httputil.BadRequest(w, err.Error())
	case errors.As(err, &decision), errors.As(err, &execNot):
		httputil.NotFound(w, err.Error())
	case errors.As(err, &lifecycle), errors.As(err, &execLife):
		httputil.Conflict(w, err.Error())
	case errors.As(err, &compBlocked):
		httputil.UnprocessableEntity(w, err.Error())
	default:
		httputil.InternalError(w, err.Error())
	}
}
