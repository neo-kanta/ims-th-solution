package leavedelegation

import "github.com/go-chi/chi/v5"

// Module Leave and Delegation Management — leave requests, agent/delegation, priority-based assignment.
type Module struct {
// Dependencies will be injected here during wire-up.
}

// NewModule creates a new leave_delegation module with its dependencies.
func NewModule() *Module {
return &Module{}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(r chi.Router) {
// TODO: register routes for leave_delegation module
}
