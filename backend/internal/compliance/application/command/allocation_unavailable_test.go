package command_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/engine"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"

	_ "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/rules/allocation"
)

type unavailablePositionPort struct{}

func (unavailablePositionPort) GetSnapshot(context.Context, uuid.UUID, time.Time) (*spi.PositionSnapshot, error) {
	return &spi.PositionSnapshot{
		Holdings: []spi.Holding{{
			Ticker:      "MISSING-CLASS",
			Quantity:    decimal.NewFromInt(100),
			MarketValue: decimal.NewFromInt(10_000),
		}},
	}, nil
}

type unavailableMarketDataPort struct{}

func (unavailableMarketDataPort) GetNAV(context.Context, uuid.UUID, time.Time) (*spi.NAVSnapshot, error) {
	return &spi.NAVSnapshot{NAV: decimal.NewFromInt(100_000)}, nil
}
func (unavailableMarketDataPort) GetPrices(context.Context, []string, time.Time) (*spi.MarketPriceSnapshot, error) {
	return &spi.MarketPriceSnapshot{Prices: map[string]decimal.Decimal{}}, nil
}
func (unavailableMarketDataPort) GetFXRates(context.Context, string, time.Time) (*spi.FXRateSnapshot, error) {
	return nil, nil
}

type partialClassificationPort struct{}

func (partialClassificationPort) GetClassifications(context.Context, []string) (*spi.ClassificationSnapshot, error) {
	return spi.NewClassificationSnapshot([]spi.InstrumentClassification{{
		Ticker: "PTT", AssetClass: "EQUITY",
	}}), nil
}

type livePortfolioMetadataPort struct{}

func (livePortfolioMetadataPort) GetMetadata(context.Context, uuid.UUID) (*spi.PortfolioMetadata, error) {
	return &spi.PortfolioMetadata{PortfolioType: "LIVE"}, nil
}

func TestPreTradeCheck_ClassificationUnavailable_PersistsEvidenceAndTypedStatus(t *testing.T) {
	t.Parallel()
	// MONITOR would normally cap a raw BLOCK to PASS. Unavailable control data
	// must bypass that cap so no false PASS is persisted or returned.
	binding := minimumTradeBinding(vo.SeverityMonitor)
	binding.RuleInstance.RuleTypeID = "allocation.asset_class_max"
	binding.CurrentVersion.Parameters = json.RawMessage(`{"asset_class":"EQUITY","max_percent_nav":60}`)

	bindingRepo := &pretradeBindingRepo{bindings: []domain.ResolvedBinding{binding}}
	checkRepo := &pretradeCheckRepo{}
	breachRepo := &pretradeBreachRepo{}
	fetcher := engine.NewFetcher(
		unavailablePositionPort{},
		unavailableMarketDataPort{},
		partialClassificationPort{},
		nil, nil, nil, nil, livePortfolioMetadataPort{},
	)
	pipeline := engine.NewPipeline(spi.GlobalRegistry(), bindingRepo, checkRepo, breachRepo, fetcher)
	handler := command.NewRunPreTradeCheckHandler(pipeline, spi.GlobalRegistry())

	resp, err := handler.Handle(context.Background(), basePreTradeReq(10, 60))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != vo.ComplianceStatusUnavailable {
		t.Fatalf("status = %q, want %q", resp.Status, vo.ComplianceStatusUnavailable)
	}
	if resp.Verdict != vo.VerdictBlock {
		t.Fatalf("final verdict = %q, want BLOCK with severity cap bypassed", resp.Verdict)
	}
	if len(checkRepo.batches) != 1 || len(checkRepo.batches[0]) != 1 {
		t.Fatalf("check record batches = %#v, want one auditable record", checkRepo.batches)
	}
	record := checkRepo.batches[0][0]
	if record.Verdict != vo.VerdictBlock || record.FinalVerdict != vo.VerdictBlock {
		t.Fatalf("record verdicts = raw %q final %q, want BLOCK/BLOCK", record.Verdict, record.FinalVerdict)
	}
	var evidence vo.Evidence
	if err := json.Unmarshal(record.Evidence, &evidence); err != nil {
		t.Fatalf("decode persisted evidence: %v", err)
	}
	if evidence.Metrics["reason"] != "MISSING_ASSET_CLASSIFICATION" {
		t.Fatalf("persisted reason = %q", evidence.Metrics["reason"])
	}
	if evidence.References["missing_instruments"] != "MISSING-CLASS" {
		t.Fatalf("persisted missing instrument = %q", evidence.References["missing_instruments"])
	}
	if len(breachRepo.breaches) != 1 || breachRepo.breaches[0].Verdict != vo.VerdictBlock {
		t.Fatalf("breaches = %#v, want one BLOCK breach", breachRepo.breaches)
	}
}
