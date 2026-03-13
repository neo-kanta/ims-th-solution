package workflow

import "github.com/go-chi/chi/v5"

// Module Workflow Management — day-start, manager-approval, transaction-closing, accounting-closing.
type Module struct {
// Dependencies will be injected here during wire-up.
}

// NewModule creates a new workflow module with its dependencies.
func NewModule() *Module {
return &Module{}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(r chi.Router) {
// TODO: register routes for workflow module
}
