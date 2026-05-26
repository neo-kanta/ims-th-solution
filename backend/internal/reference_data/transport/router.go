// Package transport mounts reference_data HTTP routes onto the parent router.
package transport

import (
	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/transport/handler"
)

// RegisterRoutes mounts /reference-data/* on r. The parent router is expected
// to apply auth middleware.
func RegisterRoutes(r chi.Router, h *handler.SecuritiesHandler) {
	r.Route("/reference-data", func(r chi.Router) {
		r.Route("/securities", func(r chi.Router) {
			r.Get("/search", h.SearchSecurities)
			r.Post("/", h.CreateSecurity)
			r.Get("/{security_id}", h.GetSecurity)
			r.Patch("/{security_id}", h.UpdateSecurity)
			r.Get("/{security_id}/mappings", h.ListSecurityMappings)
			r.Post("/{security_id}/mappings", h.AddSecurityMapping)
			r.Delete("/{security_id}/mappings/{mapping_id}", h.DeleteSecurityMapping)
		})
		r.Route("/unmapped-candidates", func(r chi.Router) {
			r.Get("/", h.ListUnmappedCandidates)
			r.Post("/{candidate_id}/map", h.MapUnmappedCandidate)
			r.Post("/{candidate_id}/reject", h.RejectUnmappedCandidate)
		})
	})
}
