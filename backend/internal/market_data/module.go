package marketdata

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/infrastructure/alphavantage"
	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/infrastructure/persistence"
	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/infrastructure/yahoo"
	httptransport "github.com/neo-kanta/ims-th-solution/backend/internal/market_data/transport/http"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
)

type Module struct {
	service *application.Service
	handler *httptransport.Handler
}

func NewModule(pool *pgxpool.Pool, cfg *config.AppConfig, redisClient *redis.Client) *Module {
	if cfg == nil {
		cfg = &config.AppConfig{}
	}

	repo := persistence.NewPostgresRepository(pool)
	cache := persistence.NewRedisQuoteCache(redisClient)
	alpha := alphavantage.NewProvider(cfg.AlphaVantageAPIKey, cfg.MarketDataHTTPTimeout)
	yahooProvider := yahoo.NewProvider(cfg.MarketDataHTTPTimeout)
	operationTimeout := marketDataOperationTimeout(cfg.MarketDataHTTPTimeout, cfg.MarketDataPrimaryProvider, cfg.MarketDataFallbackProvider)

	service := application.NewService(application.Config{
		PrimaryProvider:  cfg.MarketDataPrimaryProvider,
		FallbackProvider: cfg.MarketDataFallbackProvider,
		HTTPTimeout:      cfg.MarketDataHTTPTimeout,
		OperationTimeout: operationTimeout,
		CacheTTL:         cfg.MarketDataCacheTTL,
		ProviderConfigured: map[string]bool{
			domain.ProviderAlphaVantage: cfg.AlphaVantageAPIKey != "",
			domain.ProviderYahoo:        true,
		},
	}, []domain.MarketDataProvider{alpha, yahooProvider}, repo, repo, cache)

	return &Module{
		service: service,
		handler: httptransport.NewHandler(service),
	}
}

func (m *Module) RegisterRoutes(r chi.Router) {
	if m == nil || m.handler == nil {
		return
	}
	httptransport.Routes(r, m.handler)
}

func (m *Module) Service() *application.Service {
	if m == nil {
		return nil
	}
	return m.service
}

func marketDataOperationTimeout(httpTimeout time.Duration, providers ...string) time.Duration {
	if httpTimeout <= 0 {
		httpTimeout = 15 * time.Second
	}
	if len(providers) == 0 || (len(providers) == 2 && providers[0] == "" && providers[1] == "") {
		providers = []string{domain.ProviderAlphaVantage, domain.ProviderYahoo}
	}
	for _, provider := range providers {
		if provider == domain.ProviderAlphaVantage {
			return httpTimeout + alphavantage.DefaultRateLimitInterval()
		}
	}
	return httpTimeout
}
