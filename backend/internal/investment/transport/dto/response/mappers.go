package response

import (
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
)

// FromFund converts an entity.Fund into a FundResponse.
func FromFund(f *entity.Fund) FundResponse {
	if f == nil {
		return FundResponse{}
	}
	return FundResponse{
		ID:             f.ID,
		Code:           f.Code,
		Name:           f.Name,
		ShortName:      f.ShortName,
		FundCategoryID: f.FundCategoryID,
		BaseCurrency:   f.BaseCurrency,
		InceptionDate:  FormatDate(f.InceptionDate),
		ManagerUserID:  f.ManagerUserID,
		Benchmark:      f.Benchmark,
		RiskProfile:    string(f.RiskProfile),
		HasUnits:       f.HasUnits,
		ExternalPAMRef: f.ExternalPAMRef,
		Status:         string(f.Status),
		Version:        f.Version,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
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
