package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// CheckOutput is the result of running the evaluation pipeline.
type CheckOutput struct {
	CheckGroupID    uuid.UUID            `json:"check_group_id"`
	FinalVerdict    vo.Verdict           `json:"final_verdict"`
	Records         []entity.CheckRecord `json:"records"`
	Breaches        []entity.Breach      `json:"breaches"`
	TotalDurationMs int64                `json:"total_duration_ms"`
	RulesEvaluated  int                  `json:"rules_evaluated"`
}

// Pipeline is the core IRG evaluation orchestrator.
type Pipeline struct {
	registry    *spi.RuleRegistry
	bindingRepo domain.RuleBindingRepository
	checkRepo   domain.CheckRecordRepository
	breachRepo  domain.BreachRepository
	fetcher     *Fetcher
}

// NewPipeline creates a new evaluation pipeline.
func NewPipeline(
	registry *spi.RuleRegistry,
	bindingRepo domain.RuleBindingRepository,
	checkRepo domain.CheckRecordRepository,
	breachRepo domain.BreachRepository,
	fetcher *Fetcher,
) *Pipeline {
	return &Pipeline{
		registry:    registry,
		bindingRepo: bindingRepo,
		checkRepo:   checkRepo,
		breachRepo:  breachRepo,
		fetcher:     fetcher,
	}
}

// RunCheck executes the full evaluation pipeline.
func (p *Pipeline) RunCheck(ctx context.Context, input spi.CheckInput, scopes []vo.Scope) (*CheckOutput, error) {
	return p.runCheck(ctx, input, scopes, true)
}

// RunCheckDryRun evaluates the same rules as RunCheck but skips persistence of
// check records and breaches. It is used by order simulation endpoints where a
// caller needs a faithful verdict without audit-table side effects.
func (p *Pipeline) RunCheckDryRun(ctx context.Context, input spi.CheckInput, scopes []vo.Scope) (*CheckOutput, error) {
	return p.runCheck(ctx, input, scopes, false)
}

func (p *Pipeline) runCheck(ctx context.Context, input spi.CheckInput, scopes []vo.Scope, persist bool) (*CheckOutput, error) {
	pipelineStart := time.Now()

	slog.Info("irg: starting check",
		"check_group_id", input.CheckGroupID,
		"timing", input.Timing,
		"portfolio_id", input.PortfolioID,
		"contract_id", input.ContractID,
		"persist", persist,
	)

	// 1. Resolve applicable rules from binding matrix
	resolved, err := p.bindingRepo.ResolveApplicable(ctx, scopes, input.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("resolving applicable rules: %w", err)
	}

	if len(resolved) == 0 {
		slog.Info("irg: no applicable rules", "check_group_id", input.CheckGroupID)
		return &CheckOutput{
			CheckGroupID:    input.CheckGroupID,
			FinalVerdict:    vo.VerdictPass,
			TotalDurationMs: time.Since(pipelineStart).Milliseconds(),
		}, nil
	}

	// Sort by scope specificity (most specific first), then priority
	sort.Slice(resolved, func(i, j int) bool {
		si := resolved[i].Binding.Scope.Type.Specificity()
		sj := resolved[j].Binding.Scope.Type.Specificity()
		if si != sj {
			return si > sj
		}
		return resolved[i].Binding.Priority < resolved[j].Binding.Priority
	})

	// 2. Union data dependencies from all applicable rules
	var deps spi.DataDependencies
	for _, rb := range resolved {
		evaluator, ok := p.registry.Get(rb.RuleInstance.RuleTypeID)
		if !ok {
			continue
		}
		deps = deps.Union(evaluator.DataDependencies())
	}

	// 3. Batch-fetch data through ports (fail-closed on error)
	bundle, err := p.fetcher.Fetch(ctx, deps, input)
	if err != nil {
		return nil, fmt.Errorf("fetching data for IRG check: %w", err)
	}
	dataHash := bundle.ComputeHash()

	// 4. Evaluate each rule — ALL rules evaluated, no short-circuit
	var records []entity.CheckRecord
	var breaches []entity.Breach

	for _, rb := range resolved {
		ruleStart := time.Now()
		record := p.evaluateOne(ctx, input, rb, bundle, dataHash, ruleStart)
		records = append(records, record)

		// Create breach for BLOCK or WARN final verdicts
		if record.FinalVerdict == vo.VerdictBlock || record.FinalVerdict == vo.VerdictWarn {
			breach := entity.Breach{
				ID:             uuid.New(),
				CheckRecordID:  record.ID,
				CheckGroupID:   input.CheckGroupID,
				PortfolioID:    input.PortfolioID,
				ContractID:     input.ContractID,
				RuleTypeID:     record.RuleTypeID,
				RuleInstanceID: record.RuleInstanceID,
				Severity:       record.EffectiveSeverity,
				Verdict:        record.FinalVerdict,
				Status:         entity.BreachStatusOpen,
				Evidence: vo.Evidence{
					Metrics:    parseEvidenceMetrics(record.Evidence),
					References: map[string]string{"check_record_id": record.ID.String()},
				},
				Message:      record.Message,
				BusinessDate: input.BusinessDate,
				CreatedAt:    time.Now().UTC(),
			}
			breaches = append(breaches, breach)
		}

		slog.Info("irg: rule evaluated",
			"check_group_id", input.CheckGroupID,
			"rule_type_id", record.RuleTypeID,
			"rule_instance_id", record.RuleInstanceID,
			"verdict", record.Verdict,
			"final_verdict", record.FinalVerdict,
			"severity", record.EffectiveSeverity,
			"duration_ms", record.EvalDurationMs,
		)
	}

	// 5. Persist all records (append-only), unless this is a dry run.
	if persist && len(records) > 0 {
		if err := p.checkRepo.CreateBatch(ctx, records); err != nil {
			return nil, fmt.Errorf("persisting check records: %w", err)
		}
	}

	// 6. Persist breaches, unless this is a dry run.
	if persist {
		for i := range breaches {
			if err := p.breachRepo.Create(ctx, &breaches[i]); err != nil {
				slog.Error("irg: failed to persist breach", "error", err, "breach_id", breaches[i].ID)
			}
		}
	}

	// 7. Aggregate final verdict
	finalVerdict := AggregateVerdict(records)

	output := &CheckOutput{
		CheckGroupID:    input.CheckGroupID,
		FinalVerdict:    finalVerdict,
		Records:         records,
		Breaches:        breaches,
		TotalDurationMs: time.Since(pipelineStart).Milliseconds(),
		RulesEvaluated:  len(records),
	}

	slog.Info("irg: check completed",
		"check_group_id", input.CheckGroupID,
		"final_verdict", finalVerdict,
		"rules_evaluated", len(records),
		"breaches", len(breaches),
		"persist", persist,
		"duration_ms", output.TotalDurationMs,
	)

	return output, nil
}

func (p *Pipeline) evaluateOne(
	ctx context.Context,
	input spi.CheckInput,
	rb domain.ResolvedBinding,
	bundle *spi.DataBundle,
	dataHash string,
	ruleStart time.Time,
) entity.CheckRecord {
	now := time.Now().UTC()
	recordID := uuid.New()

	base := entity.CheckRecord{
		ID:                  recordID,
		CheckGroupID:        input.CheckGroupID,
		Timing:              input.Timing,
		PortfolioID:         input.PortfolioID,
		ContractID:          input.ContractID,
		RuleTypeID:          rb.RuleInstance.RuleTypeID,
		RuleInstanceID:      rb.RuleInstance.ID,
		RuleInstanceVersion: rb.CurrentVersion.VersionNumber,
		ParameterSnapshot:   rb.CurrentVersion.Parameters,
		EffectiveSeverity:   rb.Binding.Severity,
		DataSnapshotHash:    dataHash,
		CheckedBy:           input.Actor,
		BusinessDate:        input.BusinessDate,
		CheckedAt:           now,
		CreatedAt:           now,
	}

	if input.ProposedOrder != nil {
		base.OrderID = &input.ProposedOrder.OrderID
		base.Ticker = input.ProposedOrder.Ticker
	}

	// Look up evaluator
	evaluator, ok := p.registry.Get(rb.RuleInstance.RuleTypeID)
	if !ok {
		// Fail-closed: unregistered rule type = BLOCK
		base.Verdict = vo.VerdictBlock
		base.FinalVerdict = vo.VerdictBlock
		base.Message = fmt.Sprintf("rule type not registered: %s", rb.RuleInstance.RuleTypeID)
		base.Evidence = mustJSON(vo.Evidence{Metrics: map[string]string{"error": "rule_type_not_registered"}})
		base.EvalDurationMs = time.Since(ruleStart).Milliseconds()
		return base
	}

	// Call Evaluate
	params := spi.NewParameterSet(rb.CurrentVersion.Parameters)
	result, err := evaluator.Evaluate(ctx, input, *bundle, params)
	if err != nil {
		// Fail-closed: evaluation error = BLOCK
		base.Verdict = vo.VerdictBlock
		base.FinalVerdict = vo.VerdictBlock
		base.Message = fmt.Sprintf("evaluation error: %v", err)
		base.Evidence = mustJSON(vo.Evidence{Metrics: map[string]string{"error": err.Error()}})
		base.EvalDurationMs = time.Since(ruleStart).Milliseconds()
		return base
	}

	// Apply severity cap
	base.Verdict = result.Verdict
	base.FinalVerdict = rb.Binding.Severity.CapVerdict(result.Verdict)
	base.Message = result.Message
	base.Evidence = mustJSON(result.Evidence)
	base.EvalDurationMs = time.Since(ruleStart).Milliseconds()

	return base
}

func mustJSON(v interface{}) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return raw
}

func parseEvidenceMetrics(raw json.RawMessage) map[string]string {
	var ev vo.Evidence
	if err := json.Unmarshal(raw, &ev); err != nil {
		return nil
	}
	return ev.Metrics
}
