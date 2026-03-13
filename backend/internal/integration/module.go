package integration

import "github.com/go-chi/chi/v5"

// Module ETL/Integration — import/export adapters for market data, OMS, PAM.
type Module struct {
// Dependencies will be injected here during wire-up.
}

// NewModule creates a new integration module with its dependencies.
func NewModule() *Module {
return &Module{}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(r chi.Router) {
// TODO: register routes for integration module
}
