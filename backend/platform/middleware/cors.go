package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

var defaultDevOrigins = []string{
	"http://localhost:3000",
	"http://localhost:8080",
	"http://localhost:3003",
	"http://localhost:3004",
}

// NewCORS returns a configured CORS middleware handler.
func NewCORS(origins []string) func(http.Handler) http.Handler {
	if len(origins) == 0 {
		origins = defaultDevOrigins
	}
	return cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
