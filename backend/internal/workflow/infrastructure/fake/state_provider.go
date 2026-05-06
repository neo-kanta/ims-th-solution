// Package fake holds development-only fakes of cross-module contracts that
// the workflow module exposes. Production wiring MUST resolve these contracts
// to a real implementation; the contract-check binary asserts that.
//
// Each fake type implements contract.FakeMarker so the contract-check can
// detect it and refuse production startup unless the matching env-gate is
// set (e.g., IMS_ALLOW_FAKE_WORKFLOW_STATE=true).
//
// SCOPE Phase 0:
//   - WorkflowStateProvider exists only as a scaffolding example. The
//     workflow module already implements the real contract on *workflow.Module,
//     so this fake is not wired into cmd/server today. It establishes the
//     pattern for future contracts whose real implementation is not yet ready.
package fake

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// FakeNameWorkflowState is the contract-check identifier for this fake.
// Override gate: IMS_ALLOW_FAKE_WORKFLOW_STATE=true.
const FakeNameWorkflowState = "WORKFLOW_STATE"

// WorkflowStateProvider is a no-op fake of contract.WorkflowStateProvider.
// All calls log at WARN level so a misconfigured environment is loud.
//
// Default behaviour:
//   - IsTradeAllowed returns false (fail-closed: the safest default for a
//     pre-trade gate).
//   - IsTransactionLocked returns false.
type WorkflowStateProvider struct {
	logger *slog.Logger
}

// Compile-time assertions.
var (
	_ contract.WorkflowStateProvider = (*WorkflowStateProvider)(nil)
	_ contract.FakeMarker            = (*WorkflowStateProvider)(nil)
)

// NewWorkflowStateProvider returns a fake instance with an attached logger.
// When logger is nil, slog.Default is used.
func NewWorkflowStateProvider(logger *slog.Logger) *WorkflowStateProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &WorkflowStateProvider{logger: logger}
}

// FakeName implements contract.FakeMarker.
func (*WorkflowStateProvider) FakeName() string { return FakeNameWorkflowState }

// IsTradeAllowed implements contract.WorkflowStateProvider. Always returns
// false so that any consumer wired to the fake fails closed.
func (f *WorkflowStateProvider) IsTradeAllowed(
	_ context.Context,
	contractID uuid.UUID,
	businessDate time.Time,
) (bool, error) {
	f.logger.Warn(
		"fake WorkflowStateProvider invoked: IsTradeAllowed -> false",
		"fake", FakeNameWorkflowState,
		"contract_id", contractID,
		"business_date", businessDate.Format("2006-01-02"),
	)
	return false, nil
}

// IsTransactionLocked implements contract.WorkflowStateProvider. Always
// returns false to surface "not locked" so callers do not silently behave
// as if a real workflow gate were in place.
func (f *WorkflowStateProvider) IsTransactionLocked(
	_ context.Context,
	contractID uuid.UUID,
	businessDate time.Time,
) (bool, error) {
	f.logger.Warn(
		"fake WorkflowStateProvider invoked: IsTransactionLocked -> false",
		"fake", FakeNameWorkflowState,
		"contract_id", contractID,
		"business_date", businessDate.Format("2006-01-02"),
	)
	return false, nil
}
