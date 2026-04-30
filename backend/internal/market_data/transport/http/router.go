package httptransport

import "github.com/go-chi/chi/v5"

func Routes(r chi.Router, h *Handler) {
	r.Route("/market-data", func(r chi.Router) {
		r.Get("/quote", h.GetQuote)
		r.Get("/history", h.GetHistory)
		r.Post("/import", h.ImportMarketData)
		r.Get("/provider-health", h.ProviderHealth)
	})
}
