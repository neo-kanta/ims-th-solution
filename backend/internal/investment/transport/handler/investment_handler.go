// Package handler holds HTTP handlers for the investment module.
//
// Handlers are intentionally thin: parse → call command/query → format.
// They MUST NOT contain business rules; the policy + application layers do.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	invperm "github.com/neo-kanta/ims-th-solution/backend/internal/investment/permission"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// InvestmentHandler bundles the investment HTTP endpoints.
type InvestmentHandler struct {
	pc contract.PermissionChecker

	funds       domain.FundRepository
	portfolios  domain.PortfolioRepository
	instruments domain.InstrumentRepository
	positions   domain.PortfolioPositionRepository
	cash        domain.CashLedgerRepository
	txns        domain.PortfolioTransactionRepository
	prices      domain.PriceSnapshotRepository
	valuation   domain.ValuationRepository
	taxonomy    domain.AssetTaxonomyRepository

	fundCmd       *command.FundCommandHandler
	portfolioCmd  *command.PortfolioCommandHandler
	instrumentCmd *command.InstrumentCommandHandler
	postTxn       *command.PostTransactionHandler
	reverseTxn    *command.ReverseTransactionHandler
	postPrice     *command.PostPriceSnapshotHandler
	fundAUM       *command.ComputeFundAUMHandler
	valuationRun  *service.ValuationRunner
	fundNAVQuery  *query.GetLatestFundNAVHandler
	fundAllocQuery *query.GetFundAllocationHandler
	fundNAVHistory *query.GetFundNAVHistoryHandler
}

// NewInvestmentHandler wires every command/query for the investment module.
func NewInvestmentHandler(
	pc contract.PermissionChecker,
	funds domain.FundRepository,
	portfolios domain.PortfolioRepository,
	instruments domain.InstrumentRepository,
	positions domain.PortfolioPositionRepository,
	cash domain.CashLedgerRepository,
	txns domain.PortfolioTransactionRepository,
	prices domain.PriceSnapshotRepository,
	valuation domain.ValuationRepository,
	taxonomy domain.AssetTaxonomyRepository,
	fundCmd *command.FundCommandHandler,
	portfolioCmd *command.PortfolioCommandHandler,
	instrumentCmd *command.InstrumentCommandHandler,
	postTxn *command.PostTransactionHandler,
	reverseTxn *command.ReverseTransactionHandler,
	postPrice *command.PostPriceSnapshotHandler,
	fundAUM *command.ComputeFundAUMHandler,
	valuationRun *service.ValuationRunner,
	fundNAVQuery *query.GetLatestFundNAVHandler,
	fundAllocQuery *query.GetFundAllocationHandler,
	fundNAVHistory *query.GetFundNAVHistoryHandler,
) *InvestmentHandler {
	return &InvestmentHandler{
		pc:    pc,
		funds: funds, portfolios: portfolios, instruments: instruments,
		positions: positions, cash: cash, txns: txns,
		prices: prices, valuation: valuation, taxonomy: taxonomy,
		fundCmd: fundCmd, portfolioCmd: portfolioCmd, instrumentCmd: instrumentCmd,
		postTxn: postTxn, reverseTxn: reverseTxn, postPrice: postPrice,
		fundAUM:         fundAUM,
		valuationRun:    valuationRun,
		fundNAVQuery:    fundNAVQuery,
		fundAllocQuery:  fundAllocQuery,
		fundNAVHistory:  fundNAVHistory,
	}
}

// Shared transport helpers (actorID, parseUUIDParam, parseDate, parseDateOpt,
// parseDecimalOpt, parseDecimalRequiredOrZero, paginationParams) live in
// helpers.go so the research-report handler — and any future split handler —
// can reuse them without an implicit same-file dependency.

// writeDomainError maps domain errors to HTTP statuses uniformly.
func writeDomainError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var (
		invalid            *domain.ErrInvalidDecisionRequest
		fundNF             *domain.ErrFundNotFound
		portNF             *domain.ErrPortfolioNotFound
		instNF             *domain.ErrInstrumentNotFound
		txnNF              *domain.ErrTransactionNotFound
		precondition       *domain.ErrPostPreconditionFailed
		alreadyRev         *domain.ErrTransactionAlreadyReversed
		cantRevRev         *domain.ErrCannotReverseReversal
		fundVer            *domain.ErrFundVersionMismatch
		portVer            *domain.ErrPortfolioVersionMismatch
		codeExists         *domain.ErrCodeAlreadyExists
		fundActive         *domain.ErrFundHasActivePortfolios
		portActivity       *domain.ErrPortfolioHasOpenActivity
		complianceRejected *domain.ErrComplianceRejected
	)
	switch {
	case errors.As(err, &invalid):
		httputil.BadRequest(w, err.Error())
	case errors.As(err, &fundNF), errors.As(err, &portNF),
		errors.As(err, &instNF), errors.As(err, &txnNF):
		httputil.NotFound(w, err.Error())
	case errors.As(err, &precondition), errors.As(err, &fundActive),
		errors.As(err, &portActivity), errors.As(err, &alreadyRev),
		errors.As(err, &cantRevRev):
		httputil.UnprocessableEntity(w, err.Error())
	case errors.As(err, &complianceRejected):
		httputil.UnprocessableEntity(w, err.Error())
	case errors.As(err, &fundVer), errors.As(err, &portVer):
		httputil.Conflict(w, err.Error())
	case errors.As(err, &codeExists):
		httputil.Conflict(w, err.Error())
	default:
		httputil.InternalError(w, err.Error())
	}
}

// ─── Fund handlers ───────────────────────────────────────────────────────────

// ListFunds handles GET /investment/funds.
// @Summary List Funds
// @Description List funds visible to the authenticated user.
// @Tags Investment - Funds
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.FundListResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds [get]
func (h *InvestmentHandler) ListFunds(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationParams(r)
	filter := domain.FundListFilter{Page: page, Limit: limit, AccessibleFundIDs: accessibleFundIDs(r.Context(), h.pc)}

	funds, total, err := h.funds.List(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.FundResponse, 0, len(funds))
	for _, f := range funds {
		out = append(out, response.FromFund(f))
	}
	httputil.OK(w, response.PaginatedResponse[response.FundResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

// GetFund handles GET /investment/funds/{id}.
// @Summary Get Fund
// @Description Retrieve one fund by ID.
// @Tags Investment - Funds
// @Security BearerAuth
// @Produce json
// @Param id path string true "Fund UUID"
// @Success 200 {object} response.FundResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds/{id} [get]
func (h *InvestmentHandler) GetFund(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	f, err := h.funds.GetByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if f == nil {
		httputil.NotFound(w, "fund not found")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, id) {
		httputil.Forbidden(w, "no access to this fund")
		return
	}
	httputil.OK(w, response.FromFund(f))
}

// CreateFund handles POST /investment/funds.
// @Summary Create Fund
// @Description Create a fund record.
// @Tags Investment - Funds
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.CreateFundRequest true "Fund create payload"
// @Success 201 {object} response.FundResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds [post]
func (h *InvestmentHandler) CreateFund(w http.ResponseWriter, r *http.Request) {
	var req request.CreateFundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	inception, err := parseDate(req.InceptionDate)
	if err != nil {
		httputil.BadRequest(w, "invalid inception_date")
		return
	}
	f, err := h.fundCmd.Create(r.Context(), command.CreateFundRequest{
		Code:                   req.Code,
		Name:                   req.Name,
		ShortName:              req.ShortName,
		FundCategoryID:         req.FundCategoryID,
		BaseCurrency:           req.BaseCurrency,
		InceptionDate:          inception,
		ManagerUserID:          req.ManagerUserID,
		Benchmark:              req.Benchmark,
		RiskProfile:            vo.RiskProfile(req.RiskProfile),
		HasUnits:               req.HasUnits,
		RequirePretradePreview: req.RequirePretradePreview,
		ExternalPAMRef:         req.ExternalPAMRef,
		ActorID:                actor,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.Created(w, response.FromFund(f))
}

// UpdateFund handles PUT /investment/funds/{id}.
// @Summary Update Fund
// @Description Update mutable fields on a fund using optimistic version control.
// @Tags Investment - Funds
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Fund UUID"
// @Param request body request.UpdateFundRequest true "Fund update payload"
// @Success 200 {object} response.FundResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds/{id} [put]
func (h *InvestmentHandler) UpdateFund(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	var req request.UpdateFundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, id) {
		httputil.Forbidden(w, "no access to this fund")
		return
	}
	cmdReq := command.UpdateFundRequest{
		FundID:                 id,
		ExpectedVersion:        req.ExpectedVersion,
		Name:                   req.Name,
		ShortName:              req.ShortName,
		FundCategoryID:         req.FundCategoryID,
		ManagerUserID:          req.ManagerUserID,
		Benchmark:              req.Benchmark,
		ExternalPAMRef:         req.ExternalPAMRef,
		RequirePretradePreview: req.RequirePretradePreview,
		ActorID:                actor,
	}
	if req.RiskProfile != nil {
		rp := vo.RiskProfile(*req.RiskProfile)
		cmdReq.RiskProfile = &rp
	}
	if req.Status != nil {
		st := vo.FundStatus(*req.Status)
		cmdReq.Status = &st
	}

	f, err := h.fundCmd.Update(r.Context(), cmdReq)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.OK(w, response.FromFund(f))
}

// DeleteFund handles DELETE /investment/funds/{id}.
// @Summary Delete Fund
// @Description Soft-delete a fund using optimistic version control.
// @Tags Investment - Funds
// @Security BearerAuth
// @Accept json
// @Param id path string true "Fund UUID"
// @Param request body request.DeleteFundRequest false "Fund delete payload"
// @Success 204 "No Content"
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds/{id} [delete]
func (h *InvestmentHandler) DeleteFund(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	var req request.DeleteFundRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, id) {
		httputil.Forbidden(w, "no access to this fund")
		return
	}
	if err := h.fundCmd.SoftDelete(r.Context(), id, req.ExpectedVersion, actor); err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.NoContent(w)
}

// ─── Portfolio handlers ──────────────────────────────────────────────────────

// ListPortfolios handles GET /investment/portfolios.
// @Summary List Portfolios
// @Description List portfolios visible to the authenticated user.
// @Tags Investment - Portfolios
// @Security BearerAuth
// @Produce json
// @Param fund_id query string false "Fund UUID"
// @Param status query string false "Portfolio status"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.PortfolioListResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios [get]
func (h *InvestmentHandler) ListPortfolios(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationParams(r)
	filter := domain.PortfolioListFilter{Page: page, Limit: limit}

	if v := r.URL.Query().Get("fund_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.FundID = &id
		}
	}
	if v := r.URL.Query().Get("status"); v != "" {
		s := vo.PortfolioStatus(v)
		filter.Status = &s
	}

	// Data permission: scope to the funds the user can access.
	filter.AccessibleFundIDs = accessibleFundIDs(r.Context(), h.pc)

	ps, total, err := h.portfolios.List(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.PortfolioResponse, 0, len(ps))
	for _, p := range ps {
		out = append(out, response.FromPortfolio(p))
	}
	httputil.OK(w, response.PaginatedResponse[response.PortfolioResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

// GetPortfolio handles GET /investment/portfolios/{id}.
// @Summary Get Portfolio
// @Description Retrieve one portfolio by ID.
// @Tags Investment - Portfolios
// @Security BearerAuth
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Success 200 {object} response.PortfolioResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id} [get]
func (h *InvestmentHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	p, err := h.portfolios.GetByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if p == nil {
		httputil.NotFound(w, "portfolio not found")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, p.FundID) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	httputil.OK(w, response.FromPortfolio(p))
}

// CreatePortfolio handles POST /investment/portfolios.
// @Summary Create Portfolio
// @Description Create a portfolio under a fund.
// @Tags Investment - Portfolios
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.CreatePortfolioRequest true "Portfolio create payload"
// @Success 201 {object} response.PortfolioResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios [post]
func (h *InvestmentHandler) CreatePortfolio(w http.ResponseWriter, r *http.Request) {
	var req request.CreatePortfolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	inception, err := parseDate(req.InceptionDate)
	if err != nil {
		httputil.BadRequest(w, "invalid inception_date")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, req.FundID) {
		httputil.Forbidden(w, "no access to target fund")
		return
	}
	taxMethod := vo.TaxLotMethod(req.TaxLotMethod)
	if taxMethod == "" {
		taxMethod = vo.TaxLotMethodAverage
	}

	p, err := h.portfolioCmd.Create(r.Context(), command.CreatePortfolioRequest{
		FundID:            req.FundID,
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		BaseCurrency:      req.BaseCurrency,
		ValuationCurrency: req.ValuationCurrency,
		StrategyCode:      req.StrategyCode,
		StyleID:           req.StyleID,
		ManagerUserID:     req.ManagerUserID,
		Benchmark:         req.Benchmark,
		RiskProfile:       vo.RiskProfile(req.RiskProfile),
		InceptionDate:     inception,
		HasUnits:          req.HasUnits,
		TaxLotMethod:      taxMethod,
		ActorID:           actor,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.Created(w, response.FromPortfolio(p))
}

// UpdatePortfolio handles PUT /investment/portfolios/{id}.
// @Summary Update Portfolio
// @Description Update mutable fields on a portfolio using optimistic version control.
// @Tags Investment - Portfolios
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Param request body request.UpdatePortfolioRequest true "Portfolio update payload"
// @Success 200 {object} response.PortfolioResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id} [put]
func (h *InvestmentHandler) UpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	var req request.UpdatePortfolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, id) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	cmdReq := command.UpdatePortfolioRequest{
		PortfolioID:     id,
		ExpectedVersion: req.ExpectedVersion,
		Name:            req.Name,
		Description:     req.Description,
		StrategyCode:    req.StrategyCode,
		StyleID:         req.StyleID,
		ManagerUserID:   req.ManagerUserID,
		Benchmark:       req.Benchmark,
		ActorID:         actor,
	}
	if req.RiskProfile != nil {
		rp := vo.RiskProfile(*req.RiskProfile)
		cmdReq.RiskProfile = &rp
	}
	if req.Status != nil {
		st := vo.PortfolioStatus(*req.Status)
		cmdReq.Status = &st
	}
	p, err := h.portfolioCmd.Update(r.Context(), cmdReq)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.OK(w, response.FromPortfolio(p))
}

// DeletePortfolio handles DELETE /investment/portfolios/{id}.
// @Summary Delete Portfolio
// @Description Soft-delete a portfolio using optimistic version control.
// @Tags Investment - Portfolios
// @Security BearerAuth
// @Accept json
// @Param id path string true "Portfolio UUID"
// @Param request body request.DeletePortfolioRequest false "Portfolio delete payload"
// @Success 204 "No Content"
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id} [delete]
func (h *InvestmentHandler) DeletePortfolio(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	var req request.DeletePortfolioRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, id) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	if err := h.portfolioCmd.SoftDelete(r.Context(), id, req.ExpectedVersion, actor); err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.NoContent(w)
}

// ─── Holdings, transactions, cash ─────────────────────────────────────────────

// ListHoldings handles GET /investment/portfolios/{id}/holdings.
// @Summary List Portfolio Holdings
// @Description List current holdings for a portfolio.
// @Tags Investment - Portfolios
// @Security BearerAuth
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Success 200 {array} response.HoldingResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id}/holdings [get]
func (h *InvestmentHandler) ListHoldings(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, id) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	rows, err := h.positions.ListByPortfolio(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.HoldingResponse, 0, len(rows))
	for _, p := range rows {
		out = append(out, response.FromPosition(p))
	}
	httputil.OK(w, out)
}

// ListTransactions handles GET /investment/portfolios/{id}/transactions.
// @Summary List Portfolio Transactions
// @Description List transactions for a portfolio with optional instrument and date filters.
// @Tags Investment - Ledger
// @Security BearerAuth
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Param instrument_id query string false "Instrument UUID"
// @Param from query string false "Start business date (YYYY-MM-DD)"
// @Param to query string false "End business date (YYYY-MM-DD)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.TransactionListResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id}/transactions [get]
func (h *InvestmentHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, id) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	page, limit := paginationParams(r)
	filter := domain.TransactionListFilter{
		PortfolioID: &id,
		Page:        page,
		Limit:       limit,
	}
	if v := r.URL.Query().Get("instrument_id"); v != "" {
		if iid, err := uuid.Parse(v); err == nil {
			filter.InstrumentID = &iid
		}
	}
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := parseDate(v); err == nil {
			filter.BusinessFrom = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := parseDate(v); err == nil {
			filter.BusinessTo = &t
		}
	}
	rows, total, err := h.txns.List(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.TransactionResponse, 0, len(rows))
	for _, t := range rows {
		out = append(out, response.FromTransaction(t))
	}
	httputil.OK(w, response.PaginatedResponse[response.TransactionResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

// ListCashBalances handles GET /investment/portfolios/{id}/cash.
// @Summary List Portfolio Cash Balances
// @Description List cash balances by currency for a portfolio.
// @Tags Investment - Portfolios
// @Security BearerAuth
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Success 200 {array} response.CashBalanceResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id}/cash [get]
func (h *InvestmentHandler) ListCashBalances(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, id) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	bals, err := h.cash.ListBalances(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.CashBalanceResponse, 0, len(bals))
	for _, b := range bals {
		out = append(out, response.FromCashBalance(b))
	}
	httputil.OK(w, out)
}

// PostTransaction handles POST /investment/portfolios/{id}/transactions.
// @Summary Post Portfolio Transaction
// @Description Post a buy, sell, cash, or other portfolio transaction into the ledger.
// @Tags Investment - Ledger
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Param request body request.PostTransactionRequest true "Transaction post payload"
// @Success 201 {object} response.TransactionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id}/transactions [post]
func (h *InvestmentHandler) PostTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, id) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	cmdReq, ok := h.parsePostTransactionCommand(w, r, id)
	if !ok {
		return
	}

	res, err := h.postTxn.Handle(r.Context(), cmdReq)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.Created(w, response.FromTransaction(res.Transaction))
}

// SimulateTransaction handles POST /investment/portfolios/{id}/transactions/simulate.
// @Summary Simulate Portfolio Transaction
// @Description Run the post preconditions and pre-trade compliance checks, then preview ledger cash and position impact without mutating investment tables.
// @Tags Investment - Ledger
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Param request body request.PostTransactionRequest true "Transaction simulation payload"
// @Success 200 {object} response.TransactionSimulationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id}/transactions/simulate [post]
func (h *InvestmentHandler) SimulateTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, id) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	cmdReq, ok := h.parsePostTransactionCommand(w, r, id)
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

func (h *InvestmentHandler) parsePostTransactionCommand(
	w http.ResponseWriter,
	r *http.Request,
	portfolioID uuid.UUID,
) (command.PostTransactionRequest, bool) {
	var req request.PostTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return command.PostTransactionRequest{}, false
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return command.PostTransactionRequest{}, false
	}
	bizDate, err := parseDate(req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, "invalid business_date")
		return command.PostTransactionRequest{}, false
	}
	settlement, err := parseDateOpt(req.SettlementDate)
	if err != nil {
		httputil.BadRequest(w, "invalid settlement_date")
		return command.PostTransactionRequest{}, false
	}
	qty, err := parseDecimalOpt(req.Quantity)
	if err != nil {
		httputil.BadRequest(w, "invalid quantity")
		return command.PostTransactionRequest{}, false
	}
	price, err := parseDecimalOpt(req.Price)
	if err != nil {
		httputil.BadRequest(w, "invalid price")
		return command.PostTransactionRequest{}, false
	}
	gross, err := parseDecimalOpt(req.GrossAmount)
	if err != nil {
		httputil.BadRequest(w, "invalid gross_amount")
		return command.PostTransactionRequest{}, false
	}
	net, err := parseDecimalOpt(req.NetAmount)
	if err != nil {
		httputil.BadRequest(w, "invalid net_amount")
		return command.PostTransactionRequest{}, false
	}
	fx, err := parseDecimalOpt(req.FxRateToBase)
	if err != nil {
		httputil.BadRequest(w, "invalid fx_rate_to_base")
		return command.PostTransactionRequest{}, false
	}
	fees, err := parseDecimalRequiredOrZero(req.Fees)
	if err != nil {
		httputil.BadRequest(w, "invalid fees")
		return command.PostTransactionRequest{}, false
	}

	cmdReq := command.PostTransactionRequest{
		PortfolioID:       portfolioID,
		TransactionType:   vo.TransactionType(req.TransactionType),
		InstrumentID:      req.InstrumentID,
		Quantity:          qty,
		Price:             price,
		Currency:          req.Currency,
		Fees:              fees,
		GrossAmount:       gross,
		NetAmount:         net,
		FxRateToBase:      fx,
		BusinessDate:      bizDate,
		SettlementDate:    settlement,
		SourceDecisionID:  req.SourceDecisionID,
		SourceExecutionID: req.SourceExecutionID,
		ExternalRef:       req.ExternalRef,
		Reason:            req.Reason,
		ActorID:           actor,
		AllowForcePost:    req.ForcePost && hasPermission(r.Context(), h.pc, actor, invperm.CodeLedgerForcePost),
	}
	if req.Side != "" {
		s := vo.OrderSide(req.Side)
		cmdReq.Side = &s
	}
	return cmdReq, true
}

// ReverseTransaction handles POST /investment/portfolios/{id}/transactions/{txnId}/reverse.
// @Summary Reverse Portfolio Transaction
// @Description Post a reversal transaction for an existing portfolio transaction.
// @Tags Investment - Ledger
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Param txnId path string true "Transaction UUID"
// @Param request body request.ReverseTransactionRequest true "Transaction reversal payload"
// @Success 201 {object} response.TransactionResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id}/transactions/{txnId}/reverse [post]
func (h *InvestmentHandler) ReverseTransaction(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	txnID, err := parseUUIDParam(r, "txnId")
	if err != nil {
		httputil.BadRequest(w, "invalid transaction id")
		return
	}
	original, err := h.txns.GetByID(r.Context(), txnID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if original == nil || original.PortfolioID != portfolioID {
		httputil.NotFound(w, "transaction not found")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, original.FundID) {
		httputil.Forbidden(w, "no access to this transaction")
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

// ─── Instruments ─────────────────────────────────────────────────────────────

// ListInstruments handles GET /investment/instruments.
// @Summary List Instruments
// @Description List tradable and reference instruments with optional filters.
// @Tags Investment - Instruments
// @Security BearerAuth
// @Produce json
// @Param asset_class_id query string false "Asset class UUID"
// @Param search query string false "Search text"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.InstrumentListResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/instruments [get]
func (h *InvestmentHandler) ListInstruments(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationParams(r)
	filter := domain.InstrumentListFilter{Page: page, Limit: limit}
	if v := r.URL.Query().Get("asset_class_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.AssetClassID = &id
		}
	}
	if v := r.URL.Query().Get("search"); v != "" {
		filter.Search = v
	}

	rows, total, err := h.instruments.List(r.Context(), filter)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.InstrumentResponse, 0, len(rows))
	for _, i := range rows {
		out = append(out, response.FromInstrument(i))
	}
	httputil.OK(w, response.PaginatedResponse[response.InstrumentResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

// GetInstrument handles GET /investment/instruments/{id}.
// @Summary Get Instrument
// @Description Retrieve one instrument by ID.
// @Tags Investment - Instruments
// @Security BearerAuth
// @Produce json
// @Param id path string true "Instrument UUID"
// @Success 200 {object} response.InstrumentResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/instruments/{id} [get]
func (h *InvestmentHandler) GetInstrument(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid instrument id")
		return
	}
	i, err := h.instruments.GetByID(r.Context(), id)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if i == nil {
		httputil.NotFound(w, "instrument not found")
		return
	}
	httputil.OK(w, response.FromInstrument(i))
}

// CreateInstrument handles POST /investment/instruments.
// @Summary Create Instrument
// @Description Create an instrument master record.
// @Tags Investment - Instruments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.CreateInstrumentRequest true "Instrument create payload"
// @Success 201 {object} response.InstrumentResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/instruments [post]
func (h *InvestmentHandler) CreateInstrument(w http.ResponseWriter, r *http.Request) {
	var req request.CreateInstrumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	tick, err := parseDecimalOpt(req.TickSize)
	if err != nil {
		httputil.BadRequest(w, "invalid tick_size")
		return
	}
	inst, err := h.instrumentCmd.Create(r.Context(), command.CreateInstrumentRequest{
		PrimaryTicker:   req.PrimaryTicker,
		Name:            req.Name,
		AssetClassID:    req.AssetClassID,
		AssetSubtypeID:  req.AssetSubtypeID,
		Currency:        req.Currency,
		CountryID:       req.CountryID,
		RegionID:        req.RegionID,
		PrimaryExchange: req.PrimaryExchange,
		SectorID:        req.SectorID,
		FundCategoryID:  req.FundCategoryID,
		LotSize:         req.LotSize,
		TickSize:        tick,
		Attributes:      req.Attributes,
		ActorID:         actor,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.Created(w, response.FromInstrument(inst))
}

// UpdateInstrument handles PUT /investment/instruments/{id}.
// @Summary Update Instrument
// @Description Update mutable fields on an instrument master record.
// @Tags Investment - Instruments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Instrument UUID"
// @Param request body request.UpdateInstrumentRequest true "Instrument update payload"
// @Success 200 {object} response.InstrumentResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/instruments/{id} [put]
func (h *InvestmentHandler) UpdateInstrument(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid instrument id")
		return
	}
	var req request.UpdateInstrumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	cmdReq := command.UpdateInstrumentRequest{
		InstrumentID:    id,
		Name:            req.Name,
		PrimaryExchange: req.PrimaryExchange,
		SectorID:        req.SectorID,
		FundCategoryID:  req.FundCategoryID,
		LotSize:         req.LotSize,
		IsTradable:      req.IsTradable,
		Attributes:      req.Attributes,
		ActorID:         actor,
	}
	if req.TickSize != nil {
		t, err := decimal.NewFromString(*req.TickSize)
		if err != nil {
			httputil.BadRequest(w, "invalid tick_size")
			return
		}
		cmdReq.TickSize = &t
	}
	if req.Status != nil {
		st := vo.InstrumentStatus(*req.Status)
		cmdReq.Status = &st
	}
	inst, err := h.instrumentCmd.Update(r.Context(), cmdReq)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.OK(w, response.FromInstrument(inst))
}

// ─── Reference / taxonomy ────────────────────────────────────────────────────

// ListAssetClasses handles GET /investment/reference/asset-classes.
// @Summary List Asset Classes
// @Description List active asset classes.
// @Tags Investment - Reference
// @Security BearerAuth
// @Produce json
// @Success 200 {array} entity.AssetClass
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/reference/asset-classes [get]
func (h *InvestmentHandler) ListAssetClasses(w http.ResponseWriter, r *http.Request) {
	rows, err := h.taxonomy.ListAssetClasses(r.Context(), false)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.OK(w, rows)
}

// ListAssetSubtypes handles GET /investment/reference/asset-subtypes.
// @Summary List Asset Subtypes
// @Description List active asset subtypes, optionally filtered by asset class.
// @Tags Investment - Reference
// @Security BearerAuth
// @Produce json
// @Param asset_class_id query string false "Asset class UUID"
// @Success 200 {array} entity.AssetSubtype
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/reference/asset-subtypes [get]
func (h *InvestmentHandler) ListAssetSubtypes(w http.ResponseWriter, r *http.Request) {
	var class *uuid.UUID
	if v := r.URL.Query().Get("asset_class_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			class = &id
		}
	}
	rows, err := h.taxonomy.ListAssetSubtypes(r.Context(), class, false)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.OK(w, rows)
}

// ListSectors handles GET /investment/reference/sectors.
// @Summary List Sectors
// @Description List investment sector taxonomy rows.
// @Tags Investment - Reference
// @Security BearerAuth
// @Produce json
// @Success 200 {array} entity.Sector
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/reference/sectors [get]
func (h *InvestmentHandler) ListSectors(w http.ResponseWriter, r *http.Request) {
	rows, err := h.taxonomy.ListSectors(r.Context(), nil, nil)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.OK(w, rows)
}

// ListFundCategories handles GET /investment/reference/fund-categories.
// @Summary List Fund Categories
// @Description List active fund categories.
// @Tags Investment - Reference
// @Security BearerAuth
// @Produce json
// @Success 200 {array} entity.FundCategory
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/reference/fund-categories [get]
func (h *InvestmentHandler) ListFundCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.taxonomy.ListFundCategories(r.Context())
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.OK(w, rows)
}

// ListRegions handles GET /investment/reference/regions.
// @Summary List Regions
// @Description List active regions.
// @Tags Investment - Reference
// @Security BearerAuth
// @Produce json
// @Success 200 {array} entity.Region
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/reference/regions [get]
func (h *InvestmentHandler) ListRegions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.taxonomy.ListRegions(r.Context())
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.OK(w, rows)
}

// ListCountries handles GET /investment/reference/countries.
// @Summary List Countries
// @Description List active countries.
// @Tags Investment - Reference
// @Security BearerAuth
// @Produce json
// @Success 200 {array} entity.Country
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/reference/countries [get]
func (h *InvestmentHandler) ListCountries(w http.ResponseWriter, r *http.Request) {
	rows, err := h.taxonomy.ListCountries(r.Context(), nil)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.OK(w, rows)
}

// ListInvestmentStyles handles GET /investment/reference/investment-styles.
// @Summary List Investment Styles
// @Description List active investment styles.
// @Tags Investment - Reference
// @Security BearerAuth
// @Produce json
// @Success 200 {array} entity.InvestmentStyle
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/reference/investment-styles [get]
func (h *InvestmentHandler) ListInvestmentStyles(w http.ResponseWriter, r *http.Request) {
	rows, err := h.taxonomy.ListInvestmentStyles(r.Context())
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.OK(w, rows)
}

// ─── Pricing & Valuation ──────────────────────────────────────────────────────

// PostPriceSnapshot handles POST /investment/instruments/{id}/prices.
// @Summary Post Price Snapshot
// @Description Capture a price snapshot for an instrument and business date.
// @Tags Investment - Pricing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Instrument UUID"
// @Param request body request.PostPriceSnapshotRequest true "Price snapshot payload"
// @Success 201 {object} response.PriceResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/instruments/{id}/prices [post]
func (h *InvestmentHandler) PostPriceSnapshot(w http.ResponseWriter, r *http.Request) {
	instID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid instrument id")
		return
	}
	var req request.PostPriceSnapshotRequest
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
	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		httputil.BadRequest(w, "invalid price")
		return
	}
	snap, err := h.postPrice.Handle(r.Context(), command.PostPriceSnapshotRequest{
		InstrumentID: instID,
		BusinessDate: bizDate,
		Price:        price,
		Currency:     req.Currency,
		PriceSource:  req.PriceSource,
		ProviderRef:  req.ProviderRef,
		IsStale:      req.IsStale,
		StaleReason:  req.StaleReason,
		ActorID:      actor,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.Created(w, response.FromPrice(snap))
}

// RunValuation handles POST /investment/portfolios/{id}/valuations/run.
// @Summary Run Portfolio Valuation
// @Description Run a valuation snapshot for a portfolio and business date.
// @Tags Investment - Valuation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Param request body request.RunValuationRequest true "Valuation run payload"
// @Success 201 {object} response.ValuationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id}/valuations/run [post]
func (h *InvestmentHandler) RunValuation(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	var req request.RunValuationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, portfolioID) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	bizDate, err := parseDate(req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, "invalid business_date")
		return
	}

	rates := map[string]decimal.Decimal{}
	for k, v := range req.FxRates {
		d, err := decimal.NewFromString(v)
		if err != nil {
			httputil.BadRequest(w, "invalid fx_rate for "+k)
			return
		}
		rates[k] = d
	}
	var totalUnits *decimal.Decimal
	if req.TotalUnits != "" {
		d, err := decimal.NewFromString(req.TotalUnits)
		if err != nil {
			httputil.BadRequest(w, "invalid total_units")
			return
		}
		totalUnits = &d
	}

	res, err := h.valuationRun.Run(r.Context(), service.RunRequest{
		PortfolioID:  portfolioID,
		BusinessDate: bizDate,
		FxRates:      rates,
		TotalUnits:   totalUnits,
		ActorID:      actor,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.Created(w, response.FromValuation(res.Valuation))
}

// GetLatestValuation handles GET /investment/portfolios/{id}/valuations/latest.
// @Summary Get Latest Portfolio Valuation
// @Description Retrieve the latest internal valuation snapshot for a portfolio.
// @Tags Investment - Valuation
// @Security BearerAuth
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Success 200 {object} response.ValuationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id}/valuations/latest [get]
func (h *InvestmentHandler) GetLatestValuation(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, portfolioID) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	v, err := h.valuation.GetLatest(r.Context(), portfolioID, vo.ValuationSourceInternal)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	if v == nil {
		httputil.NotFound(w, "no valuation snapshots available")
		return
	}
	httputil.OK(w, response.FromValuation(v))
}

// GetLatestFundNAV handles GET /investment/funds/{id}/nav/latest.
//
// Aggregates the latest internal valuation snapshot of every portfolio under
// the fund into a single fund-level NAV/AUM view. This is the read-only
// counterpart to POST /investment/funds/{id}/aum/compute — it never writes
// an AUM/NAV snapshot row, so it is safe to call from cockpit cards.
// @Summary Get Latest Fund NAV
// @Description Aggregate latest per-portfolio valuations into a fund-level NAV view (read-only; no side effects).
// @Tags Investment - Valuation
// @Security BearerAuth
// @Produce json
// @Param id path string true "Fund UUID"
// @Success 200 {object} response.FundNAVResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds/{id}/nav/latest [get]
func (h *InvestmentHandler) GetLatestFundNAV(w http.ResponseWriter, r *http.Request) {
	if h.fundNAVQuery == nil {
		httputil.InternalError(w, "fund nav query not wired")
		return
	}
	fundID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, fundID) {
		httputil.Forbidden(w, "no access to this fund")
		return
	}
	result, err := h.fundNAVQuery.Handle(r.Context(), query.GetLatestFundNAVRequest{FundID: fundID})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if !result.HasAnySnapshot {
		httputil.NotFound(w, "no valuation snapshots available for this fund")
		return
	}
	httputil.OK(w, response.FromFundNAV(result))
}

// GetFundAllocation handles GET /investment/funds/{id}/allocation.
//
// Aggregates positions × instruments × latest prices into four breakdowns
// (asset class / sector / country / currency) plus the fund's total NAV
// and cash. Read-only; no state mutation.
// @Summary Get Fund Allocation
// @Description Compute asset-class, sector, country and currency breakdowns for a fund (read-only).
// @Tags Investment - Valuation
// @Security BearerAuth
// @Produce json
// @Param id path string true "Fund UUID"
// @Success 200 {object} response.FundAllocationResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds/{id}/allocation [get]
func (h *InvestmentHandler) GetFundAllocation(w http.ResponseWriter, r *http.Request) {
	if h.fundAllocQuery == nil {
		httputil.InternalError(w, "fund allocation query not wired")
		return
	}
	fundID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, fundID) {
		httputil.Forbidden(w, "no access to this fund")
		return
	}
	result, err := h.fundAllocQuery.Handle(r.Context(), query.GetFundAllocationRequest{FundID: fundID})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if !result.HasAnySnapshot {
		httputil.NotFound(w, "no positions or valuations available for this fund")
		return
	}
	httputil.OK(w, response.FromFundAllocation(result))
}

// GetFundNAVHistory handles GET /investment/funds/{id}/nav-history?range=1M|3M|6M|1Y|5Y|YTD.
//
// For unitised funds, the series carries one NAV-per-unit point per business
// day. For non-unitised funds, the series carries AUM only. The dominant
// portfolio (highest current AUM under the fund) drives the trend so the
// chart stays comparable to the latest NAV KPI.
// @Summary Get Fund NAV History
// @Description Time series of NAV-per-unit (unitised funds) or AUM (non-unitised) over a range.
// @Tags Investment - Valuation
// @Security BearerAuth
// @Produce json
// @Param id path string true "Fund UUID"
// @Param range query string false "Window: 1M, 3M, 6M, 1Y, 5Y, YTD (default 3M)"
// @Success 200 {object} response.FundNAVHistoryResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds/{id}/nav-history [get]
func (h *InvestmentHandler) GetFundNAVHistory(w http.ResponseWriter, r *http.Request) {
	if h.fundNAVHistory == nil {
		httputil.InternalError(w, "fund nav history query not wired")
		return
	}
	fundID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, fundID) {
		httputil.Forbidden(w, "no access to this fund")
		return
	}
	rng := query.NAVHistoryRange(strings.ToUpper(r.URL.Query().Get("range")))
	result, err := h.fundNAVHistory.Handle(r.Context(), query.GetFundNAVHistoryRequest{
		FundID: fundID,
		Range:  rng,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	httputil.OK(w, response.FromFundNAVHistory(result))
}

// ListValuations handles GET /investment/portfolios/{id}/valuations.
// @Summary List Portfolio Valuations
// @Description List valuation snapshots for a portfolio with optional date filters.
// @Tags Investment - Valuation
// @Security BearerAuth
// @Produce json
// @Param id path string true "Portfolio UUID"
// @Param from query string false "Start business date (YYYY-MM-DD)"
// @Param to query string false "End business date (YYYY-MM-DD)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Success 200 {object} response.ValuationListResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/portfolios/{id}/valuations [get]
func (h *InvestmentHandler) ListValuations(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, portfolioID) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	page, limit := paginationParams(r)
	from, to := dateRangeParams(r)

	rows, total, err := h.valuation.List(r.Context(), portfolioID, from, to, page, limit)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.ValuationResponse, 0, len(rows))
	for _, v := range rows {
		out = append(out, response.FromValuation(v))
	}
	httputil.OK(w, response.PaginatedResponse[response.ValuationResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

func (h *InvestmentHandler) ListNAV(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, portfolioID) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	page, limit := paginationParams(r)
	from, to := dateRangeParams(r)

	rows, total, err := h.valuation.ListNAV(r.Context(), portfolioID, from, to, page, limit)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.NAVResponse, 0, len(rows))
	for _, n := range rows {
		out = append(out, response.FromNAV(n))
	}
	httputil.OK(w, response.PaginatedResponse[response.NAVResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

func (h *InvestmentHandler) ListPortfolioAUM(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, portfolioID) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	h.listAUM(w, r, vo.AumScopePortfolio, portfolioID)
}
func (h *InvestmentHandler) ListFundAUM(w http.ResponseWriter, r *http.Request) {
	fundID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, fundID) {
		httputil.Forbidden(w, "no access to this fund")
		return
	}
	h.listAUM(w, r, vo.AumScopeFund, fundID)
}

// ComputeFundAUM aggregates per-portfolio AUM into a fund-level snapshot for
// the requested business date. Idempotent on (fund, date, source=INTERNAL).
// @Summary Compute Fund AUM
// @Description Aggregate portfolio AUM snapshots into a fund-level snapshot for a business date.
// @Tags Investment - Valuation
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Fund UUID"
// @Param request body request.ComputeFundAUMRequest true "Fund AUM compute payload"
// @Success 200 {object} response.ComputeFundAUMResponse
// @Success 201 {object} response.ComputeFundAUMResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 401 {object} httputil.ErrorResponse
// @Failure 403 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Failure 422 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /investment/funds/{id}/aum/compute [post]
func (h *InvestmentHandler) ComputeFundAUM(w http.ResponseWriter, r *http.Request) {
	if h.fundAUM == nil {
		httputil.InternalError(w, "fund aum compute not wired")
		return
	}
	fundID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid fund id")
		return
	}
	var req request.ComputeFundAUMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}
	if !hasFundAccess(r.Context(), h.pc, fundID) {
		httputil.Forbidden(w, "no access to this fund")
		return
	}
	bizDate, err := parseDate(req.BusinessDate)
	if err != nil {
		httputil.BadRequest(w, "invalid business_date")
		return
	}

	res, err := h.fundAUM.Handle(r.Context(), command.ComputeFundAUMRequest{
		FundID:       fundID,
		BusinessDate: bizDate,
		ActorID:      actor,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	out := response.ComputeFundAUMResponse{
		Snapshot:       response.FromAUM(res.Snapshot),
		PortfolioCount: res.PortfolioCount,
		Idempotent:     res.Idempotent,
	}
	if !res.Idempotent {
		httputil.Created(w, out)
	} else {
		httputil.OK(w, out)
	}
}

func (h *InvestmentHandler) listAUM(w http.ResponseWriter, r *http.Request, scope vo.AumScopeType, scopeID uuid.UUID) {
	page, limit := paginationParams(r)
	from, to := dateRangeParams(r)
	rows, total, err := h.valuation.ListAUM(r.Context(), scope, scopeID, from, to, page, limit)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	out := make([]response.AUMResponse, 0, len(rows))
	for _, a := range rows {
		out = append(out, response.FromAUM(a))
	}
	httputil.OK(w, response.PaginatedResponse[response.AUMResponse]{
		Items: out, Total: total, Page: page, Limit: limit,
	})
}

// PortfolioSummary returns headline figures for a portfolio detail view.
func (h *InvestmentHandler) PortfolioSummary(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := parseUUIDParam(r, "id")
	if err != nil {
		httputil.BadRequest(w, "invalid portfolio id")
		return
	}
	if !checkPortfolioAccess(r.Context(), h.pc, h.portfolios, portfolioID) {
		httputil.Forbidden(w, "no access to this portfolio")
		return
	}
	positions, err := h.positions.ListByPortfolio(r.Context(), portfolioID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	cash, err := h.cash.ListBalances(r.Context(), portfolioID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	val, err := h.valuation.GetLatest(r.Context(), portfolioID, vo.ValuationSourceInternal)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	cashOut := make([]response.CashBalanceResponse, 0, len(cash))
	for _, b := range cash {
		cashOut = append(cashOut, response.FromCashBalance(b))
	}
	nonZero := 0
	for _, p := range positions {
		if !p.IsZero() {
			nonZero++
		}
	}
	summary := response.PortfolioSummaryResponse{
		PortfolioID:     portfolioID,
		HoldingCount:    len(positions),
		NonZeroHoldings: nonZero,
		CashBalances:    cashOut,
	}
	if val != nil {
		v := response.FromValuation(val)
		summary.LatestValuation = &v
	}
	httputil.OK(w, summary)
}

// ─── Permission helpers ───────────────────────────────────────────────────────

// hasPermission checks a single function code; on error or false, returns false.
func hasPermission(_ context.Context, pc contract.PermissionChecker, userID uuid.UUID, code string) bool {
	if pc == nil {
		return false
	}
	ok, err := pc.HasFunctionPermission(userID, code)
	if err != nil {
		return false
	}
	return ok
}

// accessibleFundIDs returns the list of fund UUIDs the caller can access.
// Returns an empty slice on missing/failed permissions, which the repos treat as no access.
func accessibleFundIDs(ctx context.Context, pc contract.PermissionChecker) []uuid.UUID {
	claims := middleware.GetUserClaims(ctx)
	if claims == nil || pc == nil {
		return []uuid.UUID{}
	}
	uid, err := uuid.Parse(claims.Subject)
	if err != nil {
		return []uuid.UUID{}
	}
	contracts, err := pc.GetAccessibleContracts(uid)
	if err != nil {
		return []uuid.UUID{}
	}
	out := make([]uuid.UUID, 0, len(contracts))
	for _, c := range contracts {
		if c == "*" {
			return nil
		}
		if id, err := uuid.Parse(c); err == nil {
			out = append(out, id)
		}
	}
	return out
}

// hasFundAccess reports whether the caller can see a single fund.
func hasFundAccess(ctx context.Context, pc contract.PermissionChecker, fundID uuid.UUID) bool {
	claims := middleware.GetUserClaims(ctx)
	if claims == nil || pc == nil {
		return false
	}
	uid, err := uuid.Parse(claims.Subject)
	if err != nil {
		return false
	}
	ok, err := pc.HasDataPermission(uid, fundID.String())
	if err != nil {
		return false
	}
	return ok
}

// checkPortfolioAccess looks up the portfolio's fund and validates access.
func checkPortfolioAccess(ctx context.Context, pc contract.PermissionChecker, portfolios domain.PortfolioRepository, portfolioID uuid.UUID) bool {
	p, err := portfolios.GetByID(ctx, portfolioID)
	if err != nil || p == nil {
		return false
	}
	return hasFundAccess(ctx, pc, p.FundID)
}

func dateRangeParams(r *http.Request) (time.Time, time.Time) {
	to := time.Now().UTC()
	from := to.AddDate(-1, 0, 0)
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := parseDate(v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := parseDate(v); err == nil {
			to = t
		}
	}
	return from, to
}

// _ silences unused-entity import in case the build configuration drops it.
var _ = (*entity.Fund)(nil)
