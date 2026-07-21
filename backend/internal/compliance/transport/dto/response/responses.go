// Package response defines the compliance module's outbound HTTP DTOs.
// Domain entities deliberately remain inside the module; these types are the
// single JSON/OpenAPI contract consumed by the frontend.
package response

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// RuleEffectiveWindow is the documented effective-date range for a rule.
type RuleEffectiveWindow struct {
	ValidFrom time.Time  `json:"valid_from"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
}

// RuleMetadata is the stable rule-registry metadata exposed with list results.
type RuleMetadata struct {
	TypeID           string   `json:"type_id"`
	Version          string   `json:"version"`
	Category         string   `json:"category"`
	DefaultSeverity  string   `json:"default_severity"`
	SupportedTimings []string `json:"supported_timings"`
	SupportedScopes  []string `json:"supported_scopes"`
	Overridable      bool     `json:"overridable"`
	Description      string   `json:"description"`
}

// RuleInstance is the HTTP representation returned after rule creation.
// Camel-case fields are retained for compatibility with the published V1
// OpenAPI contract and its generated frontend client.
type RuleInstance struct {
	ID              uuid.UUID           `json:"id"`
	RuleTypeID      string              `json:"ruleTypeID"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	CurrentVersion  int                 `json:"currentVersion"`
	IsActive        bool                `json:"isActive"`
	EffectiveWindow RuleEffectiveWindow `json:"effectiveWindow"`
	CreatedBy       uuid.UUID           `json:"createdBy"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
}

// RuleInstanceDetail augments a rule instance with live registry metadata.
type RuleInstanceDetail struct {
	RuleInstance
	TypeMetadata *RuleMetadata `json:"type_metadata,omitempty"`
}

// RuleInstanceVersion is the initial immutable parameter version returned by
// the create endpoint.
type RuleInstanceVersion struct {
	ID             uuid.UUID       `json:"id"`
	RuleInstanceID uuid.UUID       `json:"ruleInstanceID"`
	VersionNumber  int             `json:"versionNumber"`
	Parameters     json.RawMessage `json:"parameters" swaggertype:"object"`
	ChangeReason   string          `json:"changeReason"`
	ApprovedBy     *uuid.UUID      `json:"approvedBy,omitempty"`
	CreatedBy      uuid.UUID       `json:"createdBy"`
	CreatedAt      time.Time       `json:"createdAt"`
}

// CreateRuleInstanceResult is the documented create-rule response data.
type CreateRuleInstanceResult struct {
	Instance RuleInstance        `json:"instance"`
	Version  RuleInstanceVersion `json:"version"`
}

// ListRuleInstancesResult is the documented paginated rule response data.
type ListRuleInstancesResult struct {
	Instances []RuleInstanceDetail `json:"instances"`
	Total     int64                `json:"total"`
	Offset    int                  `json:"offset"`
	Limit     int                  `json:"limit"`
}

// ThresholdBreach describes the concrete threshold comparison in evidence.
type ThresholdBreach struct {
	MetricName string `json:"metric_name"`
	Actual     string `json:"actual"`
	Limit      string `json:"limit"`
	Operator   string `json:"operator"`
	Unit       string `json:"unit"`
}

// Evidence is the structured proof attached to a breach.
type Evidence struct {
	Metrics           map[string]string `json:"metrics,omitempty"`
	ThresholdBreached *ThresholdBreach  `json:"threshold_breached,omitempty"`
	References        map[string]string `json:"references,omitempty"`
}

// Breach is the documented HTTP representation of a persisted breach.
type Breach struct {
	ID             uuid.UUID  `json:"id"`
	CheckRecordID  uuid.UUID  `json:"checkRecordID"`
	CheckGroupID   uuid.UUID  `json:"checkGroupID"`
	PortfolioID    uuid.UUID  `json:"portfolioID"`
	ContractID     *uuid.UUID `json:"contractID,omitempty"`
	RuleTypeID     string     `json:"ruleTypeID"`
	RuleInstanceID uuid.UUID  `json:"ruleInstanceID"`
	Severity       string     `json:"severity"`
	Verdict        string     `json:"verdict"`
	Status         string     `json:"status"`
	Evidence       Evidence   `json:"evidence"`
	Message        string     `json:"message"`
	BusinessDate   time.Time  `json:"businessDate"`
	CreatedAt      time.Time  `json:"createdAt"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	ResolvedBy     *uuid.UUID `json:"resolvedBy,omitempty"`
}

// ListBreachesResult is the documented paginated breach response data.
type ListBreachesResult struct {
	Breaches []Breach `json:"breaches"`
	Total    int64    `json:"total"`
	Offset   int      `json:"offset"`
	Limit    int      `json:"limit"`
}

// BreachSummary is the compact breach representation returned by the
// pre-trade and post-trade check endpoints.
type BreachSummary struct {
	BreachID    uuid.UUID `json:"breach_id"`
	RuleTypeID  string    `json:"rule_type_id"`
	Severity    string    `json:"severity"`
	Verdict     string    `json:"verdict"`
	Message     string    `json:"message"`
	Overridable bool      `json:"overridable"`
}

// PreTradeCheckResponse is the documented pre-trade check response data. A
// BLOCK verdict is a business result carried in the body, not an HTTP error.
type PreTradeCheckResponse struct {
	CheckGroupID    uuid.UUID       `json:"check_group_id"`
	Status          string          `json:"status"`
	Verdict         string          `json:"verdict"`
	RulesEvaluated  int             `json:"rules_evaluated"`
	TotalDurationMs int64           `json:"total_duration_ms"`
	Breaches        []BreachSummary `json:"breaches,omitempty"`
}

// PostTradeCheckResponse is the documented post-trade check response data.
type PostTradeCheckResponse struct {
	CheckGroupID    uuid.UUID       `json:"check_group_id"`
	Verdict         string          `json:"verdict"`
	RulesEvaluated  int             `json:"rules_evaluated"`
	TotalDurationMs int64           `json:"total_duration_ms"`
	Breaches        []BreachSummary `json:"breaches,omitempty"`
}

// CheckRecord is the documented HTTP representation of one immutable rule
// evaluation record.
type CheckRecord struct {
	ID                  uuid.UUID       `json:"id"`
	CheckGroupID        uuid.UUID       `json:"checkGroupID"`
	Timing              string          `json:"timing"`
	OrderID             *uuid.UUID      `json:"orderID,omitempty"`
	PortfolioID         uuid.UUID       `json:"portfolioID"`
	ContractID          *uuid.UUID      `json:"contractID,omitempty"`
	Ticker              string          `json:"ticker,omitempty"`
	RuleTypeID          string          `json:"ruleTypeID"`
	RuleInstanceID      uuid.UUID       `json:"ruleInstanceID"`
	RuleInstanceVersion int             `json:"ruleInstanceVersion"`
	ParameterSnapshot   json.RawMessage `json:"parameterSnapshot" swaggertype:"object"`
	Verdict             string          `json:"verdict"`
	EffectiveSeverity   string          `json:"effectiveSeverity"`
	FinalVerdict        string          `json:"finalVerdict"`
	Evidence            json.RawMessage `json:"evidence" swaggertype:"object"`
	Message             string          `json:"message"`
	DataSnapshotHash    string          `json:"dataSnapshotHash,omitempty"`
	EvalDurationMs      int64           `json:"evalDurationMs"`
	CheckedBy           string          `json:"checkedBy"`
	BusinessDate        time.Time       `json:"businessDate"`
	CheckedAt           time.Time       `json:"checkedAt"`
	CreatedAt           time.Time       `json:"createdAt"`
}

// CheckGroupResult is the documented check-group response data.
type CheckGroupResult struct {
	CheckGroupID uuid.UUID     `json:"check_group_id"`
	Records      []CheckRecord `json:"records"`
	Breaches     []Breach      `json:"breaches"`
}

// Override is the documented HTTP representation of an audited override.
type Override struct {
	ID            uuid.UUID  `json:"id"`
	BreachID      uuid.UUID  `json:"breachID"`
	Reason        string     `json:"reason"`
	OverriddenBy  uuid.UUID  `json:"overriddenBy"`
	DelegatedFrom *uuid.UUID `json:"delegatedFrom,omitempty"`
	ApprovedBy    *uuid.UUID `json:"approvedBy,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// FromListRuleInstances maps application query output to its HTTP contract.
func FromListRuleInstances(result *query.ListRuleInstancesResult) ListRuleInstancesResult {
	instances := make([]RuleInstanceDetail, 0, len(result.Instances))
	for _, item := range result.Instances {
		instances = append(instances, RuleInstanceDetail{
			RuleInstance: fromRuleInstance(item.RuleInstance),
			TypeMetadata: fromRuleMetadata(item.TypeMetadata),
		})
	}
	return ListRuleInstancesResult{
		Instances: instances,
		Total:     result.Total,
		Offset:    result.Offset,
		Limit:     result.Limit,
	}
}

// FromCreateRuleInstance maps application command output to its HTTP contract.
func FromCreateRuleInstance(result *command.CreateRuleInstanceResult) CreateRuleInstanceResult {
	return CreateRuleInstanceResult{
		Instance: fromRuleInstance(result.Instance),
		Version:  fromRuleInstanceVersion(result.Version),
	}
}

// FromListBreaches maps application query output to its HTTP contract.
func FromListBreaches(result *query.ListBreachesResult) ListBreachesResult {
	breaches := make([]Breach, 0, len(result.Breaches))
	for _, item := range result.Breaches {
		breaches = append(breaches, fromBreach(item))
	}
	return ListBreachesResult{
		Breaches: breaches,
		Total:    result.Total,
		Offset:   result.Offset,
		Limit:    result.Limit,
	}
}

// FromPreTradeCheck maps application command output to its HTTP contract.
func FromPreTradeCheck(result *command.PreTradeCheckResponse) PreTradeCheckResponse {
	return PreTradeCheckResponse{
		CheckGroupID:    result.CheckGroupID,
		Status:          string(result.Status),
		Verdict:         string(result.Verdict),
		RulesEvaluated:  result.RulesEvaluated,
		TotalDurationMs: result.TotalDurationMs,
		Breaches:        fromBreachSummaries(result.Breaches),
	}
}

// FromPostTradeCheck maps application command output to its HTTP contract.
func FromPostTradeCheck(result *command.PostTradeCheckResponse) PostTradeCheckResponse {
	return PostTradeCheckResponse{
		CheckGroupID:    result.CheckGroupID,
		Verdict:         string(result.Verdict),
		RulesEvaluated:  result.RulesEvaluated,
		TotalDurationMs: result.TotalDurationMs,
		Breaches:        fromBreachSummaries(result.Breaches),
	}
}

// FromCheckGroup maps application query output to its HTTP contract.
func FromCheckGroup(result *query.CheckGroupResult) CheckGroupResult {
	records := make([]CheckRecord, 0, len(result.Records))
	for _, item := range result.Records {
		records = append(records, fromCheckRecord(item))
	}
	breaches := make([]Breach, 0, len(result.Breaches))
	for _, item := range result.Breaches {
		breaches = append(breaches, fromBreach(item))
	}
	return CheckGroupResult{
		CheckGroupID: result.CheckGroupID,
		Records:      records,
		Breaches:     breaches,
	}
}

// FromOverride maps a domain override to its HTTP contract.
func FromOverride(item *entity.Override) Override {
	return Override{
		ID:            item.ID,
		BreachID:      item.BreachID,
		Reason:        item.Reason,
		OverriddenBy:  item.OverriddenBy,
		DelegatedFrom: item.DelegatedFrom,
		ApprovedBy:    item.ApprovedBy,
		CreatedAt:     item.CreatedAt,
	}
}

func fromRuleInstance(item entity.RuleInstance) RuleInstance {
	return RuleInstance{
		ID:             item.ID,
		RuleTypeID:     item.RuleTypeID,
		Name:           item.Name,
		Description:    item.Description,
		CurrentVersion: item.CurrentVersion,
		IsActive:       item.IsActive,
		EffectiveWindow: RuleEffectiveWindow{
			ValidFrom: item.EffectiveWindow.ValidFrom,
			ValidTo:   item.EffectiveWindow.ValidTo,
		},
		CreatedBy: item.CreatedBy,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func fromRuleInstanceVersion(item entity.RuleInstanceVersion) RuleInstanceVersion {
	return RuleInstanceVersion{
		ID:             item.ID,
		RuleInstanceID: item.RuleInstanceID,
		VersionNumber:  item.VersionNumber,
		Parameters:     append(json.RawMessage(nil), item.Parameters...),
		ChangeReason:   item.ChangeReason,
		ApprovedBy:     item.ApprovedBy,
		CreatedBy:      item.CreatedBy,
		CreatedAt:      item.CreatedAt,
	}
}

func fromRuleMetadata(item *spi.RuleMetadata) *RuleMetadata {
	if item == nil {
		return nil
	}
	timings := make([]string, 0, len(item.SupportedTimings))
	for _, timing := range item.SupportedTimings {
		timings = append(timings, string(timing))
	}
	scopes := make([]string, 0, len(item.SupportedScopes))
	for _, scope := range item.SupportedScopes {
		scopes = append(scopes, string(scope))
	}
	return &RuleMetadata{
		TypeID:           item.TypeID,
		Version:          item.Version,
		Category:         string(item.Category),
		DefaultSeverity:  string(item.DefaultSeverity),
		SupportedTimings: timings,
		SupportedScopes:  scopes,
		Overridable:      item.Overridable,
		Description:      item.Description,
	}
}

// fromBreachSummaries keeps a nil slice nil so `omitempty` continues to omit
// the breaches key exactly as the application-layer response did.
func fromBreachSummaries(items []command.BreachSummary) []BreachSummary {
	if len(items) == 0 {
		return nil
	}
	out := make([]BreachSummary, 0, len(items))
	for _, item := range items {
		out = append(out, BreachSummary{
			BreachID:    item.BreachID,
			RuleTypeID:  item.RuleTypeID,
			Severity:    string(item.Severity),
			Verdict:     string(item.Verdict),
			Message:     item.Message,
			Overridable: item.Overridable,
		})
	}
	return out
}

func fromBreach(item entity.Breach) Breach {
	return Breach{
		ID:             item.ID,
		CheckRecordID:  item.CheckRecordID,
		CheckGroupID:   item.CheckGroupID,
		PortfolioID:    item.PortfolioID,
		ContractID:     item.ContractID,
		RuleTypeID:     item.RuleTypeID,
		RuleInstanceID: item.RuleInstanceID,
		Severity:       string(item.Severity),
		Verdict:        string(item.Verdict),
		Status:         string(item.Status),
		Evidence:       fromEvidence(item.Evidence),
		Message:        item.Message,
		BusinessDate:   item.BusinessDate,
		CreatedAt:      item.CreatedAt,
		ResolvedAt:     item.ResolvedAt,
		ResolvedBy:     item.ResolvedBy,
	}
}

func fromEvidence(item vo.Evidence) Evidence {
	var threshold *ThresholdBreach
	if item.ThresholdBreached != nil {
		threshold = &ThresholdBreach{
			MetricName: item.ThresholdBreached.MetricName,
			Actual:     item.ThresholdBreached.Actual,
			Limit:      item.ThresholdBreached.Limit,
			Operator:   item.ThresholdBreached.Operator,
			Unit:       item.ThresholdBreached.Unit,
		}
	}
	return Evidence{
		Metrics:           item.Metrics,
		ThresholdBreached: threshold,
		References:        item.References,
	}
}

func fromCheckRecord(item entity.CheckRecord) CheckRecord {
	return CheckRecord{
		ID:                  item.ID,
		CheckGroupID:        item.CheckGroupID,
		Timing:              string(item.Timing),
		OrderID:             item.OrderID,
		PortfolioID:         item.PortfolioID,
		ContractID:          item.ContractID,
		Ticker:              item.Ticker,
		RuleTypeID:          item.RuleTypeID,
		RuleInstanceID:      item.RuleInstanceID,
		RuleInstanceVersion: item.RuleInstanceVersion,
		ParameterSnapshot:   append(json.RawMessage(nil), item.ParameterSnapshot...),
		Verdict:             string(item.Verdict),
		EffectiveSeverity:   string(item.EffectiveSeverity),
		FinalVerdict:        string(item.FinalVerdict),
		Evidence:            append(json.RawMessage(nil), item.Evidence...),
		Message:             item.Message,
		DataSnapshotHash:    item.DataSnapshotHash,
		EvalDurationMs:      item.EvalDurationMs,
		CheckedBy:           item.CheckedBy,
		BusinessDate:        item.BusinessDate,
		CheckedAt:           item.CheckedAt,
		CreatedAt:           item.CreatedAt,
	}
}
