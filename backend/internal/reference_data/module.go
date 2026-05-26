// Package referencedata wires the reference_data module — canonical IMS
// security identity, provider-symbol resolution, and the unmapped-candidate
// review queue. This module is the single source of truth for the
// SecurityResolver contract consumed by market_data and (later) investment.
package referencedata

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/infrastructure/persistence"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/transport"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/transport/handler"
)

// Module Reference Data — canonical IMS security identity.
type Module struct {
	service *application.Service
	handler *handler.SecuritiesHandler
}

// NewModule wires the reference_data module against a pgx pool.
func NewModule(pool *pgxpool.Pool) *Module {
	repo := persistence.NewPostgresRepository(pool)
	service := application.NewService(repo)
	h := handler.NewSecuritiesHandler(service)
	return &Module{service: service, handler: h}
}

// RegisterRoutes mounts the module's HTTP routes onto r.
func (m *Module) RegisterRoutes(r chi.Router) {
	if m == nil || m.handler == nil {
		return
	}
	transport.RegisterRoutes(r, m.handler)
}

// Resolver exposes the SecurityResolver contract for cross-module consumers.
// market_data wires this so it can resolve provider symbols without taking a
// hard dependency on internal/* of reference_data.
func (m *Module) Resolver() domain.SecurityResolver {
	if m == nil {
		return nil
	}
	return m.service
}

// Service returns the underlying application service (full surface) for
// internal callers that need write methods. Cross-module consumers should
// prefer Resolver().
func (m *Module) Service() *application.Service {
	if m == nil {
		return nil
	}
	return m.service
}
