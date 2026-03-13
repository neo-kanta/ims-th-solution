package compliance

import "github.com/go-chi/chi/v5"

// Module Compliance/IRG — blacklist/whitelist, investment ratio, instrument restriction hooks.
type Module struct {
// Dependencies will be injected here during wire-up.
}

// NewModule creates a new compliance module with its dependencies.
func NewModule() *Module {
return &Module{}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(r chi.Router) {
// TODO: register routes for compliance module
}
