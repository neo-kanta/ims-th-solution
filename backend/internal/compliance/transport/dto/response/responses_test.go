package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/compliance/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

func TestListRuleInstancesWireJSONMatchesOpenAPIContract(t *testing.T) {
	ruleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	now := time.Date(2026, time.July, 20, 3, 4, 5, 0, time.UTC)
	metadata := spi.RuleMetadata{
		TypeID:           "concentration.single_issuer",
		Version:          "1.0.0",
		Category:         spi.CategoryMandate,
		DefaultSeverity:  vo.SeverityBlock,
		SupportedTimings: []vo.CheckTiming{vo.TimingPreTrade},
		SupportedScopes:  []vo.ScopeType{vo.ScopePortfolio},
		Overridable:      true,
		Description:      "Single issuer limit",
	}

	payload := FromListRuleInstances(&query.ListRuleInstancesResult{
		Instances: []query.RuleInstanceDetail{{
			RuleInstance: entity.RuleInstance{
				ID:             ruleID,
				RuleTypeID:     metadata.TypeID,
				Name:           "Issuer limit",
				Description:    "Limit one issuer",
				CurrentVersion: 1,
				IsActive:       true,
				EffectiveWindow: vo.EffectiveWindow{
					ValidFrom: now,
				},
				CreatedBy: creatorID,
				CreatedAt: now,
				UpdatedAt: now,
			},
			TypeMetadata: &metadata,
		}},
		Total:  1,
		Offset: 0,
		Limit:  200,
	})

	data := encodeSuccessEnvelope(t, http.StatusOK, payload)
	instances := objectSlice(t, data, "instances")
	item := instances[0]

	assertJSONValue(t, item, "id", ruleID.String())
	assertJSONValue(t, item, "ruleTypeID", metadata.TypeID)
	assertJSONValue(t, item, "currentVersion", float64(1))
	assertJSONValue(t, item, "isActive", true)
	assertMissingJSONKey(t, item, "ID")
	assertMissingJSONKey(t, item, "RuleTypeID")

	window := objectValue(t, item, "effectiveWindow")
	assertJSONValue(t, window, "valid_from", now.Format(time.RFC3339))
	meta := objectValue(t, item, "type_metadata")
	assertJSONValue(t, meta, "type_id", metadata.TypeID)
}

func TestListBreachesWireJSONMatchesOpenAPIContract(t *testing.T) {
	breachID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	recordID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	groupID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	portfolioID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	ruleID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	now := time.Date(2026, time.July, 20, 6, 7, 8, 0, time.UTC)

	payload := FromListBreaches(&query.ListBreachesResult{
		Breaches: []entity.Breach{{
			ID:             breachID,
			CheckRecordID:  recordID,
			CheckGroupID:   groupID,
			PortfolioID:    portfolioID,
			RuleTypeID:     "cash.availability",
			RuleInstanceID: ruleID,
			Severity:       vo.SeverityBlock,
			Verdict:        vo.VerdictBlock,
			Status:         entity.BreachStatusOpen,
			Evidence: vo.Evidence{
				Metrics: map[string]string{"available_cash": "100.00"},
			},
			Message:      "Insufficient cash",
			BusinessDate: now,
			CreatedAt:    now,
		}},
		Total:  1,
		Offset: 0,
		Limit:  50,
	})

	data := encodeSuccessEnvelope(t, http.StatusOK, payload)
	breaches := objectSlice(t, data, "breaches")
	item := breaches[0]

	assertJSONValue(t, item, "id", breachID.String())
	assertJSONValue(t, item, "checkRecordID", recordID.String())
	assertJSONValue(t, item, "checkGroupID", groupID.String())
	assertJSONValue(t, item, "portfolioID", portfolioID.String())
	assertJSONValue(t, item, "ruleInstanceID", ruleID.String())
	assertMissingJSONKey(t, item, "ID")
	assertMissingJSONKey(t, item, "CheckRecordID")
	assertMissingJSONKey(t, item, "ContractID")

	evidence := objectValue(t, item, "evidence")
	metrics := objectValue(t, evidence, "metrics")
	assertJSONValue(t, metrics, "available_cash", "100.00")
}

func TestPreTradeCheckWireJSONMatchesOpenAPIContract(t *testing.T) {
	groupID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	breachID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	payload := FromPreTradeCheck(&command.PreTradeCheckResponse{
		CheckGroupID:    groupID,
		Status:          vo.ComplianceStatusEvaluated,
		Verdict:         vo.VerdictBlock,
		RulesEvaluated:  2,
		TotalDurationMs: 7,
		Breaches: []command.BreachSummary{{
			BreachID:    breachID,
			RuleTypeID:  "cash.availability",
			Severity:    vo.SeverityBlock,
			Verdict:     vo.VerdictBlock,
			Message:     "Insufficient cash",
			Overridable: true,
		}},
	})

	data := encodeSuccessEnvelope(t, http.StatusCreated, payload)
	assertJSONValue(t, data, "check_group_id", groupID.String())
	assertJSONValue(t, data, "status", string(vo.ComplianceStatusEvaluated))
	assertJSONValue(t, data, "verdict", string(vo.VerdictBlock))
	assertJSONValue(t, data, "rules_evaluated", float64(2))
	assertJSONValue(t, data, "total_duration_ms", float64(7))
	assertMissingJSONKey(t, data, "CheckGroupID")

	breaches := objectSlice(t, data, "breaches")
	item := breaches[0]
	assertJSONValue(t, item, "breach_id", breachID.String())
	assertJSONValue(t, item, "rule_type_id", "cash.availability")
	assertJSONValue(t, item, "severity", string(vo.SeverityBlock))
	assertJSONValue(t, item, "overridable", true)
	assertMissingJSONKey(t, item, "BreachID")
}

func TestPostTradeCheckWireJSONOmitsEmptyBreaches(t *testing.T) {
	groupID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	payload := FromPostTradeCheck(&command.PostTradeCheckResponse{
		CheckGroupID:    groupID,
		Verdict:         vo.VerdictPass,
		RulesEvaluated:  3,
		TotalDurationMs: 4,
	})

	data := encodeSuccessEnvelope(t, http.StatusCreated, payload)
	assertJSONValue(t, data, "check_group_id", groupID.String())
	assertJSONValue(t, data, "verdict", string(vo.VerdictPass))
	assertJSONValue(t, data, "rules_evaluated", float64(3))
	assertMissingJSONKey(t, data, "breaches")
	assertMissingJSONKey(t, data, "status")
}

func encodeSuccessEnvelope(t *testing.T, status int, payload any) map[string]any {
	t.Helper()
	recorder := httptest.NewRecorder()
	if status == http.StatusCreated {
		httputil.Created(recorder, payload)
	} else {
		httputil.OK(recorder, payload)
	}
	if recorder.Code != status {
		t.Fatalf("status: got %d, want %d", recorder.Code, status)
	}

	var envelope map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return objectValue(t, envelope, "data")
}

func objectSlice(t *testing.T, object map[string]any, key string) []map[string]any {
	t.Helper()
	raw, ok := object[key].([]any)
	if !ok || len(raw) == 0 {
		t.Fatalf("%s: got %#v, want non-empty array", key, object[key])
	}
	items := make([]map[string]any, 0, len(raw))
	for _, value := range raw {
		item, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("%s item: got %T, want object", key, value)
		}
		items = append(items, item)
	}
	return items
}

func objectValue(t *testing.T, object map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := object[key].(map[string]any)
	if !ok {
		t.Fatalf("%s: got %T, want object", key, object[key])
	}
	return value
}

func assertJSONValue(t *testing.T, object map[string]any, key string, want any) {
	t.Helper()
	if got := object[key]; got != want {
		t.Fatalf("%s: got %#v, want %#v", key, got, want)
	}
}

func assertMissingJSONKey(t *testing.T, object map[string]any, key string) {
	t.Helper()
	if _, exists := object[key]; exists {
		t.Fatalf("unexpected legacy JSON key %q in %#v", key, object)
	}
}
