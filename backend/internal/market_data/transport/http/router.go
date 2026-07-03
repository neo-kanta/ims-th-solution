package httptransport

import "github.com/go-chi/chi/v5"

func Routes(r chi.Router, h *Handler) {
	r.Route("/market-data", func(r chi.Router) {
		r.Get("/quote", h.GetQuote)
		r.Get("/history", h.GetHistory)
		r.Post("/import", h.ImportMarketData)
		r.Get("/provider-health", h.ProviderHealth)

		r.Post("/import-batches", h.CreateImportBatch)
		r.Post("/import-batches/{batch_id}/run", h.RunImportBatch)
		r.Get("/import-batches/{batch_id}", h.GetImportBatch)
		r.Get("/import-batches/{batch_id}/errors", h.GetImportBatchErrors)

		r.Get("/screen/watchlist", h.GetScreenWatchlist)
		r.Get("/screen/search", h.GetScreenSearch)
	})
}
