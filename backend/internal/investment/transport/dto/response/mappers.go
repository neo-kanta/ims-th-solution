package response

import (
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
)

// FromFund converts an entity.Fund into a FundResponse.
func FromFund(f *entity.Fund) FundResponse {
	if f == nil {
		return FundResponse{}
	}
	return FundResponse{
		ID:                     f.ID,
		Code:                   f.Code,
		Name:                   f.Name,
		ShortName:              f.ShortName,
		FundCategoryID:         f.FundCategoryID,
		BaseCurrency:           f.BaseCurrency,
		InceptionDate:          FormatDate(f.InceptionDate),
		ManagerUserID:          f.ManagerUserID,
		Benchmark:              f.Benchmark,
		RiskProfile:            string(f.RiskProfile),
		HasUnits:               f.HasUnits,
		RequirePretradePreview: f.RequirePretradePreview,
		ExternalPAMRef:         f.ExternalPAMRef,
		Status:                 string(f.Status),
		Version:                f.Version,
		CreatedAt:              f.CreatedAt,
		UpdatedAt:              f.UpdatedAt,
	}
}

// FromPortfolio converts an entity.Portfolio into a PortfolioResponse.
func FromPortfolio(p *entity.Portfolio) PortfolioResponse {
	if p == nil {
		return PortfolioResponse{}
	}
	return PortfolioResponse{
		ID:                p.ID,
		FundID:            p.FundID,
		PortfolioType:     string(p.PortfolioType),
		Code:              p.Code,
		Name:              p.Name,
		Description:       p.Description,
		BaseCurrency:      p.BaseCurrency,
		ValuationCurrency: p.ValuationCurrency,
		StrategyCode:      p.StrategyCode,
		StyleID:           p.StyleID,
		ManagerUserID:     p.ManagerUserID,
		Benchmark:         p.Benchmark,
		RiskProfile:       string(p.RiskProfile),
		InceptionDate:     FormatDate(p.InceptionDate),
		Status:            string(p.Status),
		HasUnits:          p.HasUnits,
		TaxLotMethod:      string(p.TaxLotMethod),
		Version:           p.Version,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

// FromInstrument converts an entity.Instrument into an InstrumentResponse.
func FromInstrument(i *entity.Instrument) InstrumentResponse {
	if i == nil {
		return InstrumentResponse{}
	}
	tick := ""
	if i.TickSize != nil {
		tick = i.TickSize.String()
	}
	return InstrumentResponse{
		ID:              i.ID,
		PrimaryTicker:   i.PrimaryTicker,
		Name:            i.Name,
		AssetClassID:    i.AssetClassID,
		AssetSubtypeID:  i.AssetSubtypeID,
		Currency:        i.Currency,
		CountryID:       i.CountryID,
		RegionID:        i.RegionID,
		PrimaryExchange: i.PrimaryExchange,
		SectorID:        i.SectorID,
		FundCategoryID:  i.FundCategoryID,
		LotSize:         i.LotSize,
		TickSize:        tick,
		IsTradable:      i.IsTradable,
		Status:          string(i.Status),
		Attributes:      i.Attributes,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
	}
}

// FromPosition converts an entity.PortfolioPosition into a HoldingResponse.
func FromPosition(p *entity.PortfolioPosition) HoldingResponse {
	if p == nil {
		return HoldingResponse{}
	}
	return HoldingResponse{
		InstrumentID:      p.InstrumentID,
		Quantity:          p.Quantity.String(),
		AverageCost:       p.AverageCost.String(),
		CostBasis:         p.CostBasis.String(),
		LastBusinessDate:  FormatDatePtr(p.LastBusinessDate),
		LastTransactionID: p.LastTransactionID,
		Version:           p.Version,
	}
}

// FromTransaction converts an entity.PortfolioTransaction into a TransactionResponse.
func FromTransaction(t *entity.PortfolioTransaction) TransactionResponse {
	if t == nil {
		return TransactionResponse{}
	}
	side := ""
	if t.Side != nil {
		side = string(*t.Side)
	}
	return TransactionResponse{
		ID:              t.ID,
		PortfolioID:     t.PortfolioID,
		FundID:          t.FundID,
		InstrumentID:    t.InstrumentID,
		TransactionType: string(t.TransactionType),
		Side:            side,
		Quantity:        FormatDecimal(t.Quantity),
		Price:           FormatDecimal(t.Price),
		Currency:        t.Currency,
		GrossAmount:     t.GrossAmount.String(),
		Fees:            t.Fees.String(),
		NetAmount:       t.NetAmount.String(),
		FxRateToBase:    FormatDecimal(t.FxRateToBase),
		BusinessDate:    FormatDate(t.BusinessDate),
		SettlementDate: func() string {
			if t.SettlementDate == nil {
				return ""
			}
			return t.SettlementDate.Format("2006-01-02")
		}(),
		SourceDecisionID:      t.SourceDecisionID,
		ReversesTransactionID: t.ReversesTransactionID,
		ExternalRef:           t.ExternalRef,
		Reason:                t.Reason,
		Status:                string(t.Status),
		CreatedAt:             t.CreatedAt,
		CreatedBy:             t.CreatedBy,
	}
}

// FromCashRequest converts an entity.PortfolioCashRequest into its HTTP DTO.
func FromCashRequest(c *entity.PortfolioCashRequest) CashRequestResponse {
	if c == nil {
		return CashRequestResponse{}
	}
	return CashRequestResponse{
		ID:                c.ID,
		PortfolioID:       c.PortfolioID,
		FundID:            c.FundID,
		TransactionType:   string(c.TransactionType),
		Amount:            c.Amount.String(),
		Currency:          c.Currency,
		Fees:              c.Fees.String(),
		ValueDate:         FormatDate(c.ValueDate),
		Memo:              c.Memo,
		Status:            string(c.Status),
		ApprovalRequestID: c.ApprovalRequestID,
		ResultingTxnID:    c.ResultingTxnID,
		SubmittedBy:       c.SubmittedBy,
		SubmittedAt:       c.SubmittedAt,
		DecidedBy:         c.DecidedBy,
		DecidedAt:         c.DecidedAt,
		Version:           c.Version,
	}
}

// FromCashBalance converts an entity.CashBalance into a CashBalanceResponse.
func FromCashBalance(b *entity.CashBalance) CashBalanceResponse {
	if b == nil {
		return CashBalanceResponse{}
	}
	return CashBalanceResponse{
		Currency:         b.Currency,
		Balance:          b.Balance.String(),
		LastBusinessDate: FormatDatePtr(b.LastBusinessDate),
		Version:          b.Version,
	}
}

// FromSimulation converts a dry-run command result into the HTTP DTO.
func FromSimulation(s *command.SimulateTransactionResult) TransactionSimulationResponse {
	if s == nil {
		return TransactionSimulationResponse{}
	}
	out := TransactionSimulationResponse{
		PortfolioID:     s.PortfolioID,
		TransactionType: string(s.TransactionType),
		InstrumentID:    s.InstrumentID,
		GrossAmount:     s.GrossAmount.String(),
		NetAmount:       s.NetAmount.String(),
		Cash: CashProjectionResponse{
			Currency:         s.Cash.Currency,
			CurrentBalance:   s.Cash.CurrentBalance.String(),
			CashImpact:       s.Cash.CashImpact.String(),
			ProjectedBalance: s.Cash.ProjectedBalance.String(),
		},
	}
	if s.Position != nil {
		out.Position = &PositionProjectionResponse{
			InstrumentID:         s.Position.InstrumentID,
			CurrentQuantity:      s.Position.CurrentQuantity.String(),
			CurrentAverageCost:   s.Position.CurrentAverageCost.String(),
			CurrentCostBasis:     s.Position.CurrentCostBasis.String(),
			ProjectedQuantity:    s.Position.ProjectedQuantity.String(),
			ProjectedAverageCost: s.Position.ProjectedAverageCost.String(),
			ProjectedCostBasis:   s.Position.ProjectedCostBasis.String(),
		}
	}
	if s.Compliance != nil {
		cp := &CompliancePreviewResponse{
			CheckGroupID:   s.Compliance.CheckGroupID,
			Verdict:        string(s.Compliance.Verdict),
			RulesEvaluated: s.Compliance.RulesEvaluated,
		}
		for _, b := range s.Compliance.Breaches {
			cp.Breaches = append(cp.Breaches, ComplianceBreachPreviewResponse{
				BreachID:    b.BreachID,
				RuleTypeID:  b.RuleTypeID,
				Verdict:     string(b.Verdict),
				Severity:    b.Severity,
				Message:     b.Message,
				Overridable: b.Overridable,
			})
		}
		out.Compliance = cp
	}
	return out
}

// FromValuation converts an entity.ValuationSnapshot (with optional lines).
func FromValuation(v *entity.ValuationSnapshot) ValuationResponse {
	if v == nil {
		return ValuationResponse{}
	}
	resp := ValuationResponse{
		ID:             v.ID,
		PortfolioID:    v.PortfolioID,
		BusinessDate:   FormatDate(v.BusinessDate),
		ValuationCcy:   v.ValuationCcy,
		MarketValue:    v.MarketValue.String(),
		CostBasis:      v.CostBasis.String(),
		UnrealisedPnL:  v.UnrealisedPnL.String(),
		RealisedPnL:    v.RealisedPnL.String(),
		ROI:            FormatDecimal(v.ROI),
		AUM:            v.AUM.String(),
		CashBalance:    v.CashBalance.String(),
		PriceSetHash:   v.PriceSetHash,
		HasStaleInputs: v.HasStaleInputs,
		IsIndicative:   v.IsIndicative,
		Source:         string(v.Source),
		CreatedAt:      v.CreatedAt,
	}
	if len(v.HoldingLines) > 0 {
		lines := make([]ValuationLineResponse, 0, len(v.HoldingLines))
		for i := range v.HoldingLines {
			ln := &v.HoldingLines[i]
			lines = append(lines, ValuationLineResponse{
				InstrumentID:    ln.InstrumentID,
				PriceSnapshotID: ln.PriceSnapshotID,
				Quantity:        ln.Quantity.String(),
				PriceInQuoteCcy: ln.PriceInQuoteCcy.String(),
				QuoteCurrency:   ln.QuoteCurrency,
				FxRate:          ln.FxRateToValuationCcy.String(),
				MarketValue:     ln.MarketValue.String(),
				CostBasis:       ln.CostBasis.String(),
				UnrealisedPnL:   ln.UnrealisedPnL.String(),
				IsStale:         ln.IsStale,
			})
		}
		resp.HoldingLines = lines
	}
	return resp
}

// FromNAV converts an entity.NAVSnapshot.
func FromNAV(n *entity.NAVSnapshot) NAVResponse {
	if n == nil {
		return NAVResponse{}
	}
	return NAVResponse{
		ID:                  n.ID,
		PortfolioID:         n.PortfolioID,
		BusinessDate:        FormatDate(n.BusinessDate),
		TotalUnits:          n.TotalUnits.String(),
		NAVPerUnit:          n.NAVPerUnit.String(),
		ValuationSnapshotID: n.ValuationSnapshotID,
		IsIndicative:        n.IsIndicative,
		CreatedAt:           n.CreatedAt,
	}
}

// FromFundNAV converts a fund-level NAV query result into a FundNAVResponse.
// Decimal fields are serialised as strings so the wire value preserves the
// repository's precision. Optional pointer fields are emitted only when set.
func FromFundNAV(r *query.GetLatestFundNAVResult) FundNAVResponse {
	if r == nil || r.Fund == nil {
		return FundNAVResponse{}
	}
	out := FundNAVResponse{
		FundID:         r.Fund.ID.String(),
		BusinessDate:   FormatDate(r.BusinessDate),
		ValuationCcy:   r.ValuationCcy,
		MarketValue:    r.MarketValue.String(),
		AUM:            r.AUM.String(),
		CashBalance:    r.CashBalance.String(),
		UnrealisedPnL:  r.UnrealisedPnL.String(),
		RealisedPnL:    r.RealisedPnL.String(),
		HasStaleInputs: r.HasStaleInputs,
		IsIndicative:   r.IsIndicative,
		PortfolioCount: r.PortfolioCount,
	}
	if r.ROI != nil {
		out.ROI = r.ROI.String()
	}
	if r.TotalUnits != nil {
		out.TotalUnits = r.TotalUnits.String()
	}
	if r.NAVPerUnit != nil {
		out.NAVPerUnit = r.NAVPerUnit.String()
	}
	return out
}

// FromFundAllocation converts a fund-level allocation query result into a
// FundAllocationResponse. Buckets keep their largest-first ordering.
func FromFundAllocation(r *query.GetFundAllocationResult) FundAllocationResponse {
	if r == nil {
		return FundAllocationResponse{}
	}
	out := FundAllocationResponse{
		FundID:         r.FundID.String(),
		AsOf:           FormatDate(r.AsOf),
		ValuationCcy:   r.ValuationCcy,
		TotalNAV:       r.TotalNAV.String(),
		TotalCash:      r.TotalCash.String(),
		PortfolioCount: r.PortfolioCount,
		ByAssetClass:   mapBuckets(r.ByAssetClass),
		BySector:       mapBuckets(r.BySector),
		ByCountry:      mapBuckets(r.ByCountry),
		ByCurrency:     mapBuckets(r.ByCurrency),
	}
	return out
}

func mapBuckets(buckets []query.AllocationBucket) []AllocationBucketResponse {
	out := make([]AllocationBucketResponse, 0, len(buckets))
	for _, b := range buckets {
		out = append(out, AllocationBucketResponse{
			Key:         b.Key,
			Label:       b.Label,
			MarketValue: b.MarketValue.String(),
			PctOfNAV:    b.PctOfNAV.String(),
		})
	}
	return out
}

// FromFundNAVHistory converts a fund-level NAV history query result.
// Series order is ascending by business_date so the frontend can draw the
// chart left-to-right without re-sorting.
func FromFundNAVHistory(r *query.GetFundNAVHistoryResult) FundNAVHistoryResponse {
	if r == nil {
		return FundNAVHistoryResponse{}
	}
	out := FundNAVHistoryResponse{
		FundID:   r.FundID.String(),
		Range:    string(r.Range),
		HasUnits: r.HasUnits,
		From:     FormatDate(r.From),
		To:       FormatDate(r.To),
		High:     r.High.String(),
		Low:      r.Low.String(),
		Latest:   r.Latest.String(),
		DeltaPct: r.DeltaPct.String(),
		IsEmpty:  r.IsEmpty,
	}
	out.Series = make([]NAVHistoryPointResponse, 0, len(r.Series))
	for _, p := range r.Series {
		row := NAVHistoryPointResponse{
			BusinessDate: FormatDate(p.BusinessDate),
			AUM:          p.AUM.String(),
		}
		if p.NAVPerUnit != nil {
			row.NAVPerUnit = p.NAVPerUnit.String()
		}
		out.Series = append(out.Series, row)
	}
	return out
}

// FromAUM converts an entity.AUMSnapshot.
func FromAUM(a *entity.AUMSnapshot) AUMResponse {
	if a == nil {
		return AUMResponse{}
	}
	return AUMResponse{
		ID:           a.ID,
		ScopeType:    string(a.ScopeType),
		ScopeID:      a.ScopeID,
		BusinessDate: FormatDate(a.BusinessDate),
		AUM:          a.AUM.String(),
		ValuationCcy: a.ValuationCcy,
		Source:       string(a.Source),
		CreatedAt:    a.CreatedAt,
	}
}

// FromResearchReport converts an entity.ResearchReport into a
// ResearchReportResponse for HTTP transport. The DerivedReviewStage field is
// computed from the report's own review_status — callers that have a richer
// view (e.g., the approval engine's task state) can overwrite it with a
// stage-aware label such as "PENDING_LEVEL_2".
func FromResearchReport(r *entity.ResearchReport) ResearchReportResponse {
	if r == nil {
		return ResearchReportResponse{}
	}
	out := ResearchReportResponse{
		ID:                   r.ID,
		ReportNo:             r.ReportNo,
		ReportDate:           FormatDate(r.ReportDate),
		EffectiveDate:        FormatDatePtr(r.EffectiveDate),
		OwnerUserID:          r.OwnerUserID,
		AuthorUserID:         r.AuthorUserID,
		ApplicableContractID: r.ApplicableContractID,
		InstrumentType:       r.InstrumentType,
		InstrumentCode:       r.InstrumentCode,
		InstrumentName:       r.InstrumentName,
		Market:               r.Market,
		Currency:             r.Currency,
		Recommendation:       string(r.Recommendation),
		ReportTitle:          r.ReportTitle,
		CompanyOverview:      r.CompanyOverview,
		CompanyOutlook:       r.CompanyOutlook,
		ESGComment:           r.ESGComment,
		FinancialStatus:      r.FinancialStatus,
		InvestmentAnalysis:   r.InvestmentAnalysis,
		RejectionReason:      r.RejectionReason,
		PostSubmissionNote:   r.PostSubmissionNote,
		ReportStatus:         string(r.ReportStatus),
		ReviewStatus:         string(r.ReviewStatus),
		DerivedReviewStage:   deriveReviewStage(r),
		InvalidatedAt:        r.InvalidatedAt,
		InvalidatedBy:        r.InvalidatedBy,
		InvalidationReason:   r.InvalidationReason,
		CreatedAt:            r.CreatedAt,
		CreatedBy:            r.CreatedBy,
		UpdatedAt:            r.UpdatedAt,
		UpdatedBy:            r.UpdatedBy,
	}
	return out
}

// deriveReviewStage maps the report's own review/report status to a single
// display string. The approval engine is the source of truth for in-flight
// multi-stage approvals; callers that resolve the active approval task can
// overwrite DerivedReviewStage with a more precise label such as
// "PENDING_LEVEL_<N>" before returning the response. This default lets the
// frontend render a sensible badge even when the approval cross-module
// lookup is skipped.
func deriveReviewStage(r *entity.ResearchReport) string {
	if r == nil {
		return ""
	}
	if r.IsInvalidated() {
		return "INVALIDATED"
	}
	switch r.ReportStatus {
	case "REJECTED":
		return "REJECTED"
	}
	switch r.ReviewStatus {
	case "NOT_SUBMITTED":
		return "NOT_SUBMITTED"
	case "SUBMITTED":
		// Single-stage display fallback when the approval engine lookup is
		// not in play. Multi-level callers overwrite this.
		return "PENDING_LEVEL_1"
	case "REVIEW_COMPLETED":
		return "REVIEW_COMPLETED"
	}
	return ""
}

// FromPrice converts an entity.PriceSnapshot.
func FromPrice(p *entity.PriceSnapshot) PriceResponse {
	if p == nil {
		return PriceResponse{}
	}
	return PriceResponse{
		ID:           p.ID,
		InstrumentID: p.InstrumentID,
		BusinessDate: FormatDate(p.BusinessDate),
		Price:        p.Price.String(),
		Currency:     p.Currency,
		PriceSource:  p.PriceSource,
		ProviderRef:  p.ProviderRef,
		IsStale:      p.IsStale,
		CapturedAt:   p.CapturedAt,
	}
}

// FromDecision converts an entity.Decision to its API response.
func FromDecision(d *entity.Decision) DecisionResponse {
	if d == nil {
		return DecisionResponse{}
	}
	r := DecisionResponse{
		ID:                                 d.ID,
		DecisionNumber:                     d.DecisionNumber,
		FundID:                             d.FundID,
		PortfolioID:                        d.PortfolioID,
		InstrumentID:                       d.InstrumentID,
		InstrumentCode:                     d.InstrumentCode,
		BusinessDate:                       FormatDate(d.BusinessDate),
		Side:                               string(d.Side),
		Quantity:                           FormatDecimal(d.Quantity),
		Amount:                             FormatDecimal(d.Amount),
		LimitPrice:                         FormatDecimal(d.LimitPrice),
		Currency:                           d.Currency,
		Exchange:                           d.Exchange,
		DecisionType:                       string(d.DecisionType),
		ProcessType:                        string(d.ProcessType),
		ProductType:                        string(d.ProductType),
		StrategyCode:                       d.StrategyCode,
		AmendmentNo:                        d.AmendmentNo,
		ResearchReportID:                   d.ResearchReportID,
		ResearchReportNo:                   d.ResearchReportNo,
		Rationale:                          d.Rationale,
		Status:                             string(d.Status),
		ApprovalRequestID:                  d.ApprovalRequestID,
		ComplianceReleaseApprovalRequestID: d.ComplianceReleaseApprovalRequestID,
		ApprovalStatus:                     d.ApprovalStatus,
		ComplianceCheckGroupID:             d.ComplianceCheckGroupID,
		SubmitterUserID:                    d.SubmitterUserID,
		SubmittedAt:                        d.SubmittedAt,
		CancelledAt:                        d.CancelledAt,
		CancellationReason:                 d.CancellationReason,
		ReadyForExecutionAt:                d.ReadyForExecutionAt,
		CreatedAt:                          d.CreatedAt,
		UpdatedAt:                          d.UpdatedAt,
	}
	for _, l := range d.Lines {
		r.Lines = append(r.Lines, FromDecisionLine(l))
	}
	return r
}

// FromDecisionLine converts an entity.DecisionLine to its API response.
func FromDecisionLine(l *entity.DecisionLine) DecisionLineResponse {
	if l == nil {
		return DecisionLineResponse{}
	}
	return DecisionLineResponse{
		ID:             l.ID,
		LineNumber:     l.LineNumber,
		InstrumentID:   l.InstrumentID,
		InstrumentCode: l.InstrumentCode,
		ProductType:    string(l.ProductType),
		Side:           string(l.Side),
		Quantity:       FormatDecimal(l.Quantity),
		Amount:         FormatDecimal(l.Amount),
		TargetWeight:   FormatDecimal(l.TargetWeight),
		LimitPrice:     FormatDecimal(l.LimitPrice),
		Currency:       l.Currency,
		Notes:          l.Notes,
	}
}

// FromExecution converts an entity.Execution to its API response.
func FromExecution(e *entity.Execution) ExecutionResponse {
	if e == nil {
		return ExecutionResponse{}
	}
	return ExecutionResponse{
		ID:                 e.ID,
		DecisionID:         e.DecisionID,
		FundID:             e.FundID,
		PortfolioID:        e.PortfolioID,
		InstrumentID:       e.InstrumentID,
		InstrumentCode:     e.InstrumentCode,
		BusinessDate:       FormatDate(e.BusinessDate),
		Side:               string(e.Side),
		OrderedQuantity:    FormatDecimal(e.OrderedQuantity),
		OrderedAmount:      FormatDecimal(e.OrderedAmount),
		ExecutedQuantity:   FormatDecimal(e.ExecutedQuantity),
		ExecutedAmount:     FormatDecimal(e.ExecutedAmount),
		ExecutionPrice:     FormatDecimal(e.ExecutionPrice),
		Currency:           e.Currency,
		Status:             string(e.Status),
		TraderUserID:       e.TraderUserID,
		BrokerReference:    e.BrokerReference,
		ExecutedAt:         e.ExecutedAt,
		CancelledAt:        e.CancelledAt,
		CancellationReason: e.CancellationReason,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}

// FromTradeConfirmation converts an entity.TradeConfirmation to its API response.
func FromTradeConfirmation(c *entity.TradeConfirmation) TradeConfirmationResponse {
	if c == nil {
		return TradeConfirmationResponse{}
	}
	return TradeConfirmationResponse{
		ID:                c.ID,
		ExecutionID:       c.ExecutionID,
		DecisionID:        c.DecisionID,
		FundID:            c.FundID,
		PortfolioID:       c.PortfolioID,
		BusinessDate:      FormatDate(c.BusinessDate),
		ConfirmedQuantity: FormatDecimal(c.ConfirmedQuantity),
		ConfirmedAmount:   FormatDecimal(c.ConfirmedAmount),
		ConfirmedPrice:    FormatDecimal(c.ConfirmedPrice),
		Currency:          c.Currency,
		BrokerReference:   c.BrokerReference,
		ImportBatchID:     c.ImportBatchID,
		Status:            string(c.Status),
		DiscrepancyReason: c.DiscrepancyReason,
		ReviewedAt:        c.ReviewedAt,
		ReviewedBy:        c.ReviewedBy,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}
