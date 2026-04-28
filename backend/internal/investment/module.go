// Package investment wires the Stock Investment Management module
// (analysis reports, decisions, execution, review).
//
// Cross-module integrations live behind pkg/contract interfaces — this module
// never imports internal/compliance or internal/workflow directly.
package investment

import (
	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

// Module owns investment write-side use cases and their HTTP transport.
type Module struct {
	submitDecision *command.SubmitDecisionForExecutionHandler
}

// NewModule constructs the investment module.
//
// Dependencies:
//   - compliance: pre-trade IRG gate (pkg/contract.ComplianceChecker).
//     MUST NOT be nil — pre-trade enforcement is a non-negotiable regulatory
//     control. A nil checker is treated as a deploy-time configuration error
//     by callers downstream.
//
// Persistence wiring (DecisionRepository) is intentionally deferred — this
// scaffolding exposes the command constructor so later batches can inject a
// real Postgres repository without a signature change.
func NewModule(_ contract.ComplianceChecker) *Module {
	// DecisionRepository is not yet wired — the submit command will be
	// constructed when the persistence layer lands. Returning a Module with a
	// nil handler keeps callers compile-compatible while making it obvious
	// that routes are not yet registered.
	return &Module{}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(_ chi.Router) {
	// Routes are registered in a later batch, once persistence + transport DTOs
	// are in place. Keeping this a no-op preserves the wire-up shape.
}

// SubmitDecisionHandler returns the command handler for direct use by callers
// that have already assembled its dependencies (e.g., integration tests).
func (m *Module) SubmitDecisionHandler() *command.SubmitDecisionForExecutionHandler {
	if m == nil {
		return nil
	}
	return m.submitDecision
}
