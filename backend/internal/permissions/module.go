package permissions

import "github.com/go-chi/chi/v5"

// Module Permissions Management — accounts, groups, function permissions, data permissions.
type Module struct {
// Dependencies will be injected here during wire-up.
}

// NewModule creates a new permissions module with its dependencies.
func NewModule() *Module {
return &Module{}
}

// RegisterRoutes mounts this module's HTTP routes onto the given router.
func (m *Module) RegisterRoutes(r chi.Router) {
// TODO: register routes for permissions module
}
