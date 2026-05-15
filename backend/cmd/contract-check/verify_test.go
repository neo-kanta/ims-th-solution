package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// stubChecker is a real (non-fake) implementation of ComplianceChecker used
// to assert the "ok" path.
type stubChecker struct{}

func (stubChecker) CheckProposedOrder(_ context.Context, _ contract.ProposedOrderCheck) (*contract.ProposedOrderResult, error) {
	return nil, nil
}

// stubFakeProvider is a fake implementation that exposes itself via FakeMarker.
type stubFakeProvider struct {
	name string
}

func (s stubFakeProvider) FakeName() string { return s.name }
func (stubFakeProvider) IsTradeAllowed(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return false, nil
}
func (stubFakeProvider) IsTransactionLocked(_ context.Context, _ uuid.UUID, _ time.Time) (bool, error) {
	return false, nil
}

func envFrom(m map[string]string) envLookup {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func TestVerifyBindings_AllOK(t *testing.T) {
	t.Parallel()
	bindings := []binding{
		{Name: "compliance", Resolved: stubChecker{}},
	}
	outcomes, err := verifyBindings(bindings, envFrom(nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outcomes) != 1 || outcomes[0].Status != "ok" {
		t.Errorf("expected single ok outcome, got %+v", outcomes)
	}
}

func TestVerifyBindings_NilIsRejected(t *testing.T) {
	t.Parallel()
	var nilChecker contract.ComplianceChecker
	bindings := []binding{
		{Name: "compliance", Resolved: nilChecker},
	}
	outcomes, err := verifyBindings(bindings, envFrom(nil))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if outcomes[0].Status != "nil" {
		t.Errorf("expected status nil, got %q", outcomes[0].Status)
	}
	if !strings.Contains(err.Error(), "compliance") {
		t.Errorf("error must cite binding name: %v", err)
	}
}

func TestVerifyBindings_NilTypedPointerIsRejected(t *testing.T) {
	t.Parallel()
	var typed *stubChecker
	bindings := []binding{
		{Name: "compliance", Resolved: typed},
	}
	outcomes, err := verifyBindings(bindings, envFrom(nil))
	if err == nil {
		t.Fatalf("expected error for typed nil pointer, got nil")
	}
	if outcomes[0].Status != "nil" {
		t.Errorf("expected status nil for typed-nil-pointer trap, got %q", outcomes[0].Status)
	}
}

func TestVerifyBindings_FakeRejectedByDefault(t *testing.T) {
	t.Parallel()
	bindings := []binding{
		{Name: "workflow", Resolved: stubFakeProvider{name: "WORKFLOW_STATE"}},
	}
	outcomes, err := verifyBindings(bindings, envFrom(nil))
	if err == nil {
		t.Fatalf("expected error for fake binding, got nil")
	}
	if outcomes[0].Status != "fake-rejected" {
		t.Errorf("expected fake-rejected, got %q", outcomes[0].Status)
	}
	if !strings.Contains(outcomes[0].Message, "IMS_ALLOW_FAKE_WORKFLOW_STATE") {
		t.Errorf("message should cite env-gate variable: %q", outcomes[0].Message)
	}
}

func TestVerifyBindings_FakeAllowedByEnv(t *testing.T) {
	t.Parallel()
	bindings := []binding{
		{Name: "workflow", Resolved: stubFakeProvider{name: "WORKFLOW_STATE"}},
	}
	env := envFrom(map[string]string{"IMS_ALLOW_FAKE_WORKFLOW_STATE": "true"})
	outcomes, err := verifyBindings(bindings, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcomes[0].Status != "fake-allowed" {
		t.Errorf("expected fake-allowed, got %q", outcomes[0].Status)
	}
}

func TestVerifyBindings_CustomAllowFakeEnvHonoured(t *testing.T) {
	t.Parallel()
	bindings := []binding{
		{Name: "workflow", Resolved: stubFakeProvider{name: "X"}, AllowFakeEnv: "ALLOW_X"},
	}
	env := envFrom(map[string]string{"ALLOW_X": "true"})
	outcomes, err := verifyBindings(bindings, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outcomes[0].Status != "fake-allowed" {
		t.Errorf("expected fake-allowed, got %q", outcomes[0].Status)
	}
}

func TestVerifyBindings_MultipleFailuresAreCollected(t *testing.T) {
	t.Parallel()
	var nilChecker contract.ComplianceChecker
	bindings := []binding{
		{Name: "a", Resolved: nilChecker},
		{Name: "b", Resolved: stubFakeProvider{name: "B"}},
		{Name: "c", Resolved: stubChecker{}},
	}
	outcomes, err := verifyBindings(bindings, envFrom(nil))
	if err == nil {
		t.Fatalf("expected combined error")
	}
	if len(outcomes) != 3 {
		t.Fatalf("expected 3 outcomes, got %d", len(outcomes))
	}
	wantStatuses := []string{"nil", "fake-rejected", "ok"}
	for i, want := range wantStatuses {
		if outcomes[i].Status != want {
			t.Errorf("outcome[%d] status: got %q, want %q", i, outcomes[i].Status, want)
		}
	}
}
