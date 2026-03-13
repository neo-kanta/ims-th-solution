package marketdata

import "github.com/go-chi/chi/v5"

// Module Market Data — market data integration adapter (stub for PoC).
type Module struct {
// Dependencies will be injected here during wire-up.
}

// NewModule creates a new market_data module with its dependencies.
func NewModule() *Module {
return &Module{}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(r chi.Router) {
// TODO: register routes for market_data module
}
