package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	invperm "github.com/neo-kanta/ims-th-solution/backend/internal/investment/permission"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// Portfolio V2 ledger write endpoints (docs/api/portfolio-v2-api-ddd.md,
// Milestone 3 of
// docs/handoff/portfolio-v2-claude-implementation-prompt.md).
//
// request.PostTransactionRequest and request.ReverseTransactionRequest
// already have no fund_id/contract_id field — V1 derives fund_id from the
// resolved portfolio inside the command layer (see
// application/command/post_transaction.go: command.PostTransactionRequest
// has no FundID field either). V2 reuses the exact same request DTOs and
// command handlers, so this file only adds the portfolioCode -> portfolio_id
// resolution step; it never trusts a client-supplied fund_id or contract_id
// because those fields do not exist on the wire format to begin with. Every
// handler in this package (V1 and V2) decodes JSON via json.NewDecoder
// without DisallowUnknownFields, so an extra "fund_id" key in a V2 request
// body is silently ignored, consistent with V1 behavior — not rejected.

// PostTransactionByCode handles POST /api/v2/portfolios/{portfolioCode}/transactions.
// @Summary Post Portfolio Transaction By Code
// @Description Post a buy, sell, cash, or other portfolio transaction into the ledger, resolved by business code (Portfolio V2). Request body must not include fund_id or contract_id. For a LIVE portfolio, a cash movement (CASH_IN/CASH_OUT/FEE/DIVIDEND) is NOT posted immediately — it is staged for approval and returned with HTTP 202 as a pending cash request. MODEL cash movements are rejected (422).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param Idempotency-Key header string false "Optional idempotency key for a LIVE cash movement. A retry with the same key returns the original pending cash request instead of creating a duplicate. Max 255 chars; a missing key means the request is not deduplicated."
// @Param request body request.PostTransactionRequest true "Transaction post payload"
// @Success 201 {object} response.TransactionResponse "Posted immediately (SIMULATION cash, or any BUY/SELL/other non-gated movement)"
// @Success 202 {object} response.CashRequestResponse "LIVE cash movement staged for approval (pending)"
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/transactions [post]
func (h *InvestmentHandler) PostTransactionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	cmdReq, ok := h.parsePostTransactionCommand(w, r, p.ID)
	if !ok {
		return
	}
	res, err := h.postTxn.Handle(r.Context(), cmdReq)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	// A LIVE cash movement is staged for approval, not posted — return 202 with
	// the pending cash request so the frontend can distinguish "pending
	// approval" from "posted".
	if res.Pending {
		httputil.Accepted(w, response.FromCashRequest(res.CashRequest))
		return
	}
	httputil.Created(w, response.FromTransaction(res.Transaction))
}

// SimulateTransactionByCode handles POST /api/v2/portfolios/{portfolioCode}/transactions/simulate.
// @Summary Simulate Portfolio Transaction By Code
// @Description Run the post preconditions and pre-trade compliance checks, then preview ledger cash and position impact without mutating investment tables, resolved by business code (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param request body request.PostTransactionRequest true "Transaction simulation payload"
// @Success 200 {object} response.TransactionSimulationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/transactions/simulate [post]
func (h *InvestmentHandler) SimulateTransactionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	cmdReq, ok := h.parsePostTransactionCommand(w, r, p.ID)
	if !ok {
		return
	}
	res, err := h.postTxn.Simulate(r.Context(), cmdReq)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.OK(w, response.FromSimulation(res))
}

// ReverseTransactionByCode handles POST /api/v2/portfolios/{portfolioCode}/transactions/{transactionId}/reverse.
// @Summary Reverse Portfolio Transaction By Code
// @Description Post a reversal transaction for an existing portfolio transaction, resolved by business code (Portfolio V2).
// @Tags Investment - Portfolios V2
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param portfolioCode path string true "Portfolio code"
// @Param transactionId path string true "Transaction UUID"
// @Param request body request.ReverseTransactionRequest true "Transaction reversal payload"
// @Success 201 {object} response.TransactionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /portfolios/{portfolioCode}/transactions/{transactionId}/reverse [post]
func (h *InvestmentHandler) ReverseTransactionByCode(w http.ResponseWriter, r *http.Request) {
	p, ok := h.resolvePortfolioCode(w, r)
	if !ok {
		return
	}
	txnID, err := parseUUID(chi.URLParam(r, "transactionId"))
	if err != nil {
		httputil.BadRequest(w, "invalid transaction id")
		return
	}
	original, err := h.txns.GetByID(r.Context(), txnID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if original == nil || original.PortfolioID != p.ID {
		httputil.NotFound(w, "transaction not found")
		return
	}
	var req request.ReverseTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	bizDate, err := parseDate(req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, "invalid business_date")
		return
	}
	res, err := h.reverseTxn.Handle(r.Context(), command.ReverseTransactionRequest{
		OriginalTransactionID: txnID,
		BusinessDate:          bizDate,
		Reason:                req.Reason,
		ActorID:               actor,
		AllowForcePost:        req.ForcePost && hasPermission(r.Context(), h.pc, actor, invperm.CodeLedgerForcePost),
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.Created(w, response.FromTransaction(res.Reversal))
}
