// Package transport wires the compliance module's HTTP routes.
package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/transport/handler"
)

// RegisterRoutes mounts all compliance routes under the given router prefix.
// Caller is responsible for applying auth + permission middleware before calling this.
//
// Route layout:
//
//	POST   /compliance/checks/pre-trade              → RunPreTradeCheck
//	POST   /compliance/checks/post-trade             → RunPostTradeCheck
//	GET    /compliance/checks/{groupID}              → GetCheckGroup
//	GET    /compliance/breaches                      → ListBreaches
//	POST   /compliance/breaches/{breachID}/override  → OverrideBreach
//	GET    /compliance/rules                         → ListRuleInstances
func RegisterRoutes(r chi.Router, h *handler.ComplianceHandler) {
	r.Route("/compliance", func(r chi.Router) {
		// Check execution (OMS calls these)
		r.Post("/checks/pre-trade", h.RunPreTradeCheck)
		r.Post("/checks/post-trade", h.RunPostTradeCheck)
		r.Get("/checks/{groupID}", h.GetCheckGroup)

		// Breach management
		r.Get("/breaches", h.ListBreaches)
		r.Post("/breaches/{breachID}/override", h.OverrideBreach)

		// Rule administration
		r.Get("/rules", h.ListRuleInstances)
	})
}

// HealthHandler returns a simple 200 OK for liveness checks.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","module":"compliance"}`))
	}
}
