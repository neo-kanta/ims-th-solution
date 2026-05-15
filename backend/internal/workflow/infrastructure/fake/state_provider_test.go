package fake

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

func TestWorkflowStateProvider_IsFake(t *testing.T) {
	t.Parallel()
	var sp contract.WorkflowStateProvider = NewWorkflowStateProvider(nil)

	marker, ok := sp.(contract.FakeMarker)
	if !ok {
		t.Fatalf("fake.WorkflowStateProvider must implement contract.FakeMarker")
	}
	if got := marker.FakeName(); got != FakeNameWorkflowState {
		t.Errorf("FakeName() = %q, want %q", got, FakeNameWorkflowState)
	}
}

func TestWorkflowStateProvider_IsTradeAllowed_FailClosed(t *testing.T) {
	t.Parallel()
	sp := NewWorkflowStateProvider(nil)
	allowed, err := sp.IsTradeAllowed(context.Background(), uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Errorf("fake must fail closed for IsTradeAllowed; got allowed=true")
	}
}

func TestWorkflowStateProvider_IsTransactionLocked_NotLocked(t *testing.T) {
	t.Parallel()
	sp := NewWorkflowStateProvider(nil)
	locked, err := sp.IsTransactionLocked(context.Background(), uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if locked {
		t.Errorf("fake must report not-locked for IsTransactionLocked; got locked=true")
	}
}
