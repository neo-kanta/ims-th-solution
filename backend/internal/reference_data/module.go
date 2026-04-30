package referencedata

import "github.com/go-chi/chi/v5"

// Module Reference Data — currencies, markets, instruments, Thai holidays.
type Module struct {
	// Dependencies will be injected here during wire-up.
}

// NewModule creates a new reference_data module with its dependencies.
func NewModule() *Module {
	return &Module{}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(r chi.Router) {
	// TODO: register routes for reference_data module
}
