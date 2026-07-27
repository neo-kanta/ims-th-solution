package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// CashRequestHandler serves the Portfolio V2 LIVE cash-request endpoints:
// the submitter's pending list and the cancel action. Cash requests are
// created implicitly by POST .../transactions (handled in
// portfolio_v2_ledger_handler.go); this handler only reads and cancels them.
type CashRequestHandler struct {
	portfolios   domain.PortfolioRepository
	cashRequests domain.PortfolioCashRequestRepository
	cmd          *command.PostTransactionHandler
	pc           contract.PermissionChecker
}

// NewCashRequestHandler wires the handler.
func NewCashRequestHandler(
	portfolios domain.PortfolioRepository,
	cashRequests domain.PortfolioCashRequestRepository,
	cmd *command.PostTransactionHandler,
) *CashRequestHandler {
	return &CashRequestHandler{portfolios: portfolios, cashRequests: cashRequests, cmd: cmd}
}

// SetPermissionChecker injects the data-scope permission checker. Must be wired
// in production — resolvePortfolioByCode fails closed when pc is nil.
func (h *CashRequestHandler) SetPermissionChecker(pc contract.PermissionChecker) {
	if h != nil {
		h.pc = pc
	}
}

// ListCashRequestsByCode handles GET /api/v2/portfolios/{portfolioCode}/cash-requests.
// @Summary List Portfolio Cash Requests By Code
// @Description List LIVE cash-movement approval requests for a portfolio, resolved by business code (Portfolio V2). Optional status filter (PENDING/APPROVED/REJECTED/CANCELLED).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param status query string false "Cash request status (PENDING, APPROVED, REJECTED, CANCELLED)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.CashRequestListResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/cash-requests [get]
func (h *CashRequestHandler) ListCashRequestsByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	page, limit := paginationParams(r)
	filter := domain.CashRequestListFilter{Page: page, Limit: limit}
	if v := r.URL.Query().Get("status"); v != "" {
		s := vo.CashRequestStatus(v)
		if !s.IsValid() {
			httputil.BadRequest(w, "invalid status (expected PENDING, APPROVED, REJECTED, or CANCELLED)")
			return
		}
		filter.Status = &s
	}
	rows, total, err := h.cashRequests.ListByPortfolio(r.Context(), p.ID, filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.CashRequestResponse, 0, len(rows))
	for _, c := range rows {
		out = append(out, response.FromCashRequest(c))
	}
	httputil.OK(w, response.CashRequestListResponse{Items: out, Total: total, Page: page, Limit: limit})
}

// CancelCashRequestByCode handles
// POST /api/v2/portfolios/{portfolioCode}/cash-requests/{cashRequestId}/cancel.
// @Summary Cancel Portfolio Cash Request By Code
// @Description Cancel a PENDING cash-movement approval request that belongs to the resolved portfolio (Portfolio V2). Only the submitter may cancel; after cancel, approvers can no longer act on it.
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param cashRequestId path string true "Cash request UUID"
// @Param request body request.CancelCashRequestV2Request false "Optional cancellation reason"
// @Success 200 {object} response.CashRequestResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/cash-requests/{cashRequestId}/cancel [post]
func (h *CashRequestHandler) CancelCashRequestByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := resolvePortfolioByCode(w, r, h.portfolios, h.pc)
	if !ok {
		return
	}
	reqID, err := parseUUID(chi.URLParam(r, "cashRequestId"))
	if err != nil {
		httputil.BadRequest(w, "invalid cash request id")
		return
	}
	// Confirm the request belongs to the resolved portfolio before acting —
	// 404 both when missing and when it belongs to another portfolio, so a
	// caller cannot probe for other portfolios' requests.
	existing, err := h.cashRequests.GetByID(r.Context(), reqID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if existing == nil || existing.PortfolioID != p.ID {
		httputil.NotFound(w, "cash request not found")
		return
	}
	var body request.CancelCashRequestV2Request
	// Body is optional; ignore decode errors on an empty body.
	_ = json.NewDecoder(r.Body).Decode(&body)
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	cr, err := h.cmd.CancelCashRequest(r.Context(), reqID, actor, body.Reason)
	if err != nil {
		writeCashRequestError(w, err)
		return
	}
	httputil.OK(w, response.FromCashRequest(cr))
}

// writeCashRequestError maps cash-request domain errors to HTTP statuses.
func writeCashRequestError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var (
		notFound     *domain.ErrCashRequestNotFound
		notCancel    *domain.ErrCashRequestNotCancellable
		forbidden    *domain.ErrCashRequestForbidden
		invalid      *domain.ErrInvalidDecisionRequest
		precondition *domain.ErrPostPreconditionFailed
	)
	switch {
	case errors.As(err, &invalid):
		httputil.BadRequest(w, err.Error())
	case errors.As(err, &notFound):
		httputil.NotFound(w, err.Error())
	case errors.As(err, &forbidden):
		httputil.Forbidden(w, err.Error())
	case errors.As(err, &notCancel), errors.As(err, &precondition):
		httputil.UnprocessableEntity(w, err.Error())
	default:
		httputil.InternalError(w, err.Error())
	}
}
