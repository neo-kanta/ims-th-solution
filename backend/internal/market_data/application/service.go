package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

const maxSymbolLength = 64

var (
	ErrInvalidMarketDataRequest       = errors.New("invalid market data request")
	ErrMarketDataProviderNotAvailable = errors.New("market data provider is not available")
)

type Config struct {
	PrimaryProvider    string
	FallbackProvider   string
	HTTPTimeout        time.Duration
	OperationTimeout   time.Duration
	CacheTTL           time.Duration
	ProviderConfigured map[string]bool
}

type Service struct {
	providers        map[string]domain.MarketDataProvider
	repo             domain.SnapshotRepository
	logger           domain.ProviderRequestLogger
	cache            domain.QuoteCache
	batchRepo        domain.ImportBatchRepository
	securityResolver SecurityResolver
	cfg              Config
	now              func() time.Time
}

type ImportMarketDataRequest struct {
	Symbol         string
	Provider       string
	IncludeQuote   bool
	IncludeHistory bool
	HistoryLimit   int
}

type ImportMarketDataResult struct {
	Symbol              string            `json:"symbol"`
	RequestedProvider   string            `json:"requested_provider"`
	QuoteProvider       string            `json:"quote_provider,omitempty"`
	HistoryProvider     string            `json:"history_provider,omitempty"`
	Quote               *domain.Quote     `json:"quote,omitempty"`
	DailyPrices         []domain.PriceBar `json:"daily_prices,omitempty"`
	DailyPricesImported int               `json:"daily_prices_imported"`
	ImportedAt          time.Time         `json:"imported_at"`
}

func NewService(
	cfg Config,
	providers []domain.MarketDataProvider,
	repo domain.SnapshotRepository,
	logger domain.ProviderRequestLogger,
	cache domain.QuoteCache,
) *Service {
	if cfg.PrimaryProvider == "" {
		cfg.PrimaryProvider = domain.ProviderAlphaVantage
	}
	if cfg.FallbackProvider == "" {
		cfg.FallbackProvider = domain.ProviderYahoo
	}
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = 10 * time.Second
	}
	if cfg.OperationTimeout <= 0 {
		cfg.OperationTimeout = cfg.HTTPTimeout
	}
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 15 * time.Minute
	}
	if cfg.ProviderConfigured == nil {
		cfg.ProviderConfigured = map[string]bool{}
	}

	byName := make(map[string]domain.MarketDataProvider, len(providers))
	for _, p := range providers {
		if p == nil {
			continue
		}
		byName[p.ProviderName()] = p
	}
	return &Service{
		providers: byName,
		repo:      repo,
		logger:    logger,
		cache:     cache,
		cfg:       cfg,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

// SetImportBatchRepository wires the import batch repository used by
// CreateImportBatch / RunImportBatch / GetImportBatch endpoints. It is
// kept separate from NewService to avoid breaking the existing test
// constructors that don't need the batch table.
func (s *Service) SetImportBatchRepository(repo domain.ImportBatchRepository) {
	if s == nil {
		return
	}
	s.batchRepo = repo
}

// SetSecurityResolver wires the reference_data resolver used by batch import
// to translate input symbols into canonical IMS securities and provider
// symbols. nil-safe; when unset, market_data falls back to the legacy
// market_symbols path.
func (s *Service) SetSecurityResolver(resolver SecurityResolver) {
	if s == nil {
		return
	}
	s.securityResolver = resolver
}

func (s *Service) GetQuote(ctx context.Context, symbol string) (*domain.Quote, error) {
	var err error
	symbol, err = normalizeSymbol(symbol)
	if err != nil {
		return nil, err
	}
	mapping := s.symbolMapping(ctx, symbol)

	if s.cache != nil {
		if q, ok, err := s.cache.GetQuote(ctx, symbol); err == nil && ok && q != nil {
			q.Cached = true
			q.Stale = false
			q.StaleReason = ""
			return q, nil
		}
	}

	q, failures, err := s.fetchQuoteWithOrder(ctx, symbol, s.providerOrder(), mapping)
	if err == nil {
		return q, nil
	}

	if s.repo != nil {
		if q, err := s.repo.GetLatestQuote(ctx, symbol); err == nil && q != nil {
			q.Symbol = symbol
			q.Stale = true
			q.Cached = true
			q.StaleReason = "all configured providers failed; returning latest cached snapshot"
			if len(failures) > 0 {
				q.StaleReason += ": " + strings.Join(failures, "; ")
			}
			return q, nil
		}
	}

	return nil, fmt.Errorf("market data quote unavailable for %s: %s", symbol, strings.Join(failures, "; "))
}

// GetQuoteAsOf resolves the latest normalized snapshot for symbol dated on
// or before businessDate directly from the snapshot repository. It never
// makes a live provider call, so it is safe to use for historical dates and
// never blocks on a provider outage. Returns (nil, nil) when no snapshot
// exists on or before businessDate.
func (s *Service) GetQuoteAsOf(ctx context.Context, symbol string, businessDate time.Time) (*domain.Quote, error) {
	symbol, err := normalizeSymbol(symbol)
	if err != nil {
		return nil, err
	}
	if s.repo == nil {
		return nil, nil
	}
	q, err := s.repo.GetSnapshotAsOf(ctx, symbol, businessDate)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (s *Service) GetQuoteFromProvider(ctx context.Context, symbol, providerName string) (*domain.Quote, error) {
	symbol, err := normalizeSymbol(symbol)
	if err != nil {
		return nil, err
	}
	order, requestedProvider, err := s.providerOrderForSelection(providerName)
	if err != nil {
		return nil, err
	}
	q, failures, err := s.fetchQuoteWithOrder(ctx, symbol, order, s.symbolMapping(ctx, symbol))
	if err != nil {
		return nil, fmt.Errorf("market data quote unavailable for %s via %s: %s", symbol, requestedProvider, strings.Join(failures, "; "))
	}
	return q, nil
}

func (s *Service) GetDailyPrices(ctx context.Context, symbol string, limit int) ([]domain.PriceBar, error) {
	var err error
	symbol, err = normalizeSymbol(symbol)
	if err != nil {
		return nil, err
	}
	mapping := s.symbolMapping(ctx, symbol)
	if limit <= 0 {
		limit = 250
	}

	bars, failures, err := s.fetchHistoryWithOrder(ctx, symbol, limit, s.providerOrder(), mapping)
	if err == nil {
		return bars, nil
	}

	if s.repo != nil {
		repoLimit := limit * len(s.providerOrder())
		if repoLimit < limit {
			repoLimit = limit
		}
		bars, err := s.repo.ListDailyPrices(ctx, symbol, repoLimit)
		if err == nil && len(bars) > 0 {
			bars = dedupeDailyBars(bars, s.providerOrder(), limit)
			reason := "all configured providers failed; returning latest cached daily prices"
			if len(failures) > 0 {
				reason += ": " + strings.Join(failures, "; ")
			}
			for i := range bars {
				bars[i].Symbol = symbol
				bars[i].Stale = true
				bars[i].StaleReason = reason
			}
			return bars, nil
		}
	}

	return nil, fmt.Errorf("market data history unavailable for %s: %s", symbol, strings.Join(failures, "; "))
}

func (s *Service) GetDailyPricesFromProvider(ctx context.Context, symbol string, limit int, providerName string) ([]domain.PriceBar, error) {
	symbol, err := normalizeSymbol(symbol)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 250
	}
	order, requestedProvider, err := s.providerOrderForSelection(providerName)
	if err != nil {
		return nil, err
	}
	bars, failures, err := s.fetchHistoryWithOrder(ctx, symbol, limit, order, s.symbolMapping(ctx, symbol))
	if err != nil {
		return nil, fmt.Errorf("market data history unavailable for %s via %s: %s", symbol, requestedProvider, strings.Join(failures, "; "))
	}
	return bars, nil
}

func (s *Service) ImportMarketData(ctx context.Context, req ImportMarketDataRequest) (*ImportMarketDataResult, error) {
	symbol, err := normalizeSymbol(req.Symbol)
	if err != nil {
		return nil, err
	}
	if !req.IncludeQuote && !req.IncludeHistory {
		return nil, fmt.Errorf("%w: at least one of include_quote or include_history must be true", ErrInvalidMarketDataRequest)
	}
	if req.HistoryLimit <= 0 {
		req.HistoryLimit = 250
	}
	if req.HistoryLimit > 1000 {
		req.HistoryLimit = 1000
	}

	order, requestedProvider, err := s.providerOrderForSelection(req.Provider)
	if err != nil {
		return nil, err
	}
	mapping := s.symbolMapping(ctx, symbol)
	result := &ImportMarketDataResult{
		Symbol:            symbol,
		RequestedProvider: requestedProvider,
		ImportedAt:        s.now(),
	}

	if req.IncludeQuote {
		q, failures, err := s.fetchQuoteWithOrder(ctx, symbol, order, mapping)
		if err != nil {
			return nil, fmt.Errorf("importing quote for %s via %s failed: %s", symbol, requestedProvider, strings.Join(failures, "; "))
		}
		result.Quote = q
		result.QuoteProvider = q.Provider
	}

	if req.IncludeHistory {
		bars, failures, err := s.fetchHistoryWithOrder(ctx, symbol, req.HistoryLimit, order, mapping)
		if err != nil {
			return nil, fmt.Errorf("importing daily prices for %s via %s failed: %s", symbol, requestedProvider, strings.Join(failures, "; "))
		}
		result.DailyPrices = bars
		result.DailyPricesImported = len(bars)
		if len(bars) > 0 {
			result.HistoryProvider = bars[0].Provider
		}
	}

	return result, nil
}

func (s *Service) ProviderHealth(ctx context.Context) ([]domain.ProviderHealth, error) {
	lastByProvider := map[string]domain.ProviderRequestLog{}
	if s.logger != nil {
		logs, err := s.logger.ListLatestProviderRequests(ctx, 50)
		if err != nil {
			return nil, err
		}
		for _, l := range logs {
			if _, exists := lastByProvider[l.ProviderName]; !exists {
				lastByProvider[l.ProviderName] = l
			}
		}
	}

	out := make([]domain.ProviderHealth, 0, 2)
	for _, providerName := range s.providerOrder() {
		if providerName == "" {
			continue
		}
		configured, ok := s.cfg.ProviderConfigured[providerName]
		if !ok {
			configured = s.providers[providerName] != nil
		}
		h := domain.ProviderHealth{
			ProviderName: providerName,
			Role:         s.providerRole(providerName),
			Configured:   configured,
			Official:     providerName == domain.ProviderAlphaVantage,
			Healthy:      configured && s.providers[providerName] != nil,
		}
		if l, ok := lastByProvider[providerName]; ok {
			h.LastStatus = l.Status
			h.LastError = l.ErrorMessage
			h.LastChecked = l.RequestedAt
			h.Healthy = h.Healthy && l.Status == "success"
		}
		out = append(out, h)
	}
	return out, nil
}

func (s *Service) fetchQuoteWithOrder(ctx context.Context, symbol string, order []string, mapping *domain.SymbolMapping) (*domain.Quote, []string, error) {
	var failures []string
	for _, providerName := range order {
		provider := s.providers[providerName]
		if provider == nil {
			failures = append(failures, providerName+": provider not wired")
			continue
		}

		providerSymbol := mappedProviderSymbol(mapping, providerName, symbol)
		q, err := s.withTimeoutQuote(ctx, provider, providerSymbol)
		if err == nil && q != nil {
			q.Symbol = symbol
			q.Provider = provider.ProviderName()
			if q.CapturedAt.IsZero() {
				q.CapturedAt = s.now()
			}
			applyMapping(q, mapping)
			q.Stale = false
			q.StaleReason = ""
			s.persistQuote(ctx, *q)
			if s.cache != nil {
				_ = s.cache.SetQuote(ctx, *q, s.cfg.CacheTTL)
			}
			return q, failures, nil
		}
		failures = append(failures, failureMessage(providerName, err))
	}
	return nil, failures, fmt.Errorf("all selected market data providers failed")
}

func (s *Service) fetchHistoryWithOrder(ctx context.Context, symbol string, limit int, order []string, mapping *domain.SymbolMapping) ([]domain.PriceBar, []string, error) {
	var failures []string
	for _, providerName := range order {
		provider := s.providers[providerName]
		if provider == nil {
			failures = append(failures, providerName+": provider not wired")
			continue
		}

		providerSymbol := mappedProviderSymbol(mapping, providerName, symbol)
		bars, err := s.withTimeoutHistory(ctx, provider, providerSymbol)
		if err == nil && len(bars) > 0 {
			for i := range bars {
				bars[i].Symbol = symbol
				bars[i].Provider = provider.ProviderName()
				if mapping != nil && bars[i].Currency == "" {
					bars[i].Currency = mapping.Currency
				}
				bars[i].Stale = false
				bars[i].StaleReason = ""
			}
			if len(bars) > limit {
				bars = bars[:limit]
			}
			s.persistHistory(ctx, symbol, provider.ProviderName(), bars)
			return bars, failures, nil
		}
		if err == nil {
			err = fmt.Errorf("provider returned no daily prices")
		}
		failures = append(failures, failureMessage(providerName, err))
	}
	return nil, failures, fmt.Errorf("all selected market data providers failed")
}

func (s *Service) withTimeoutQuote(ctx context.Context, provider domain.MarketDataProvider, symbol string) (*domain.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.OperationTimeout)
	defer cancel()

	start := s.now()
	q, err := provider.GetQuote(ctx, symbol)
	s.logRequest(ctx, provider.ProviderName(), "quote", symbol, err, time.Since(start))
	return q, err
}

func (s *Service) withTimeoutHistory(ctx context.Context, provider domain.MarketDataProvider, symbol string) ([]domain.PriceBar, error) {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.OperationTimeout)
	defer cancel()

	start := s.now()
	bars, err := provider.GetDailyPrices(ctx, symbol)
	s.logRequest(ctx, provider.ProviderName(), "history", symbol, err, time.Since(start))
	return bars, err
}

func (s *Service) logRequest(ctx context.Context, providerName, operation, symbol string, err error, duration time.Duration) {
	if s.logger == nil {
		return
	}
	log := domain.ProviderRequestLog{
		ProviderName: providerName,
		Symbol:       symbol,
		Operation:    operation,
		Status:       "success",
		Duration:     duration,
		RequestedAt:  s.now(),
	}
	if err != nil {
		log.Status = "error"
		log.ErrorMessage = err.Error()
		var perr *domain.ProviderError
		if errors.As(err, &perr) {
			log.StatusCode = perr.StatusCode
			log.ErrorCode = perr.Code
			log.ErrorMessage = perr.Message
		}
	}
	_ = s.logger.LogProviderRequest(ctx, log)
}

func (s *Service) persistQuote(ctx context.Context, quote domain.Quote) {
	if s.repo == nil {
		return
	}
	_ = s.repo.UpsertSymbol(ctx, domain.SymbolMapping{
		Symbol:             quote.Symbol,
		AssetType:          quote.AssetType,
		Currency:           quote.Currency,
		AlphaVantageSymbol: providerSymbol(quote.Provider, domain.ProviderAlphaVantage, quote.Symbol),
		YahooFinanceSymbol: providerSymbol(quote.Provider, domain.ProviderYahoo, quote.Symbol),
	})
	_ = s.repo.SaveQuote(ctx, quote)
}

func (s *Service) persistHistory(ctx context.Context, symbol, provider string, bars []domain.PriceBar) {
	if s.repo == nil {
		return
	}
	_ = s.repo.UpsertSymbol(ctx, domain.SymbolMapping{
		Symbol:             symbol,
		AssetType:          domain.AssetTypeUnknown,
		AlphaVantageSymbol: providerSymbol(provider, domain.ProviderAlphaVantage, symbol),
		YahooFinanceSymbol: providerSymbol(provider, domain.ProviderYahoo, symbol),
	})
	_ = s.repo.SaveDailyPrices(ctx, symbol, provider, bars)
}

func (s *Service) symbolMapping(ctx context.Context, symbol string) *domain.SymbolMapping {
	if s.repo == nil {
		return nil
	}
	mapping, err := s.repo.GetSymbolMapping(ctx, symbol)
	if err != nil {
		return nil
	}
	return mapping
}

func mappedProviderSymbol(mapping *domain.SymbolMapping, providerName, fallback string) string {
	if mapping == nil {
		return fallback
	}
	switch providerName {
	case domain.ProviderAlphaVantage:
		if mapping.AlphaVantageSymbol != "" {
			return mapping.AlphaVantageSymbol
		}
	case domain.ProviderYahoo:
		if mapping.YahooFinanceSymbol != "" {
			return mapping.YahooFinanceSymbol
		}
	}
	return fallback
}

func applyMapping(quote *domain.Quote, mapping *domain.SymbolMapping) {
	if quote == nil || mapping == nil {
		return
	}
	if quote.AssetType == "" || quote.AssetType == domain.AssetTypeUnknown {
		quote.AssetType = mapping.AssetType
	}
	if quote.Currency == "" {
		quote.Currency = mapping.Currency
	}
}

func dedupeDailyBars(bars []domain.PriceBar, providerOrder []string, limit int) []domain.PriceBar {
	if len(bars) == 0 {
		return nil
	}
	priority := make(map[string]int, len(providerOrder))
	for i, provider := range providerOrder {
		priority[provider] = i
	}
	const defaultPriority = 1000

	byDate := make(map[string]domain.PriceBar, len(bars))
	for _, bar := range bars {
		key := bar.Date.UTC().Format("2006-01-02")
		current, exists := byDate[key]
		if !exists || providerPriority(bar.Provider, priority, defaultPriority) < providerPriority(current.Provider, priority, defaultPriority) {
			byDate[key] = bar
		}
	}

	out := make([]domain.PriceBar, 0, len(byDate))
	for _, bar := range byDate {
		out = append(out, bar)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Date.After(out[j].Date)
	})
	if limit > 0 && len(out) > limit {
		return out[:limit]
	}
	return out
}

func providerPriority(provider string, priority map[string]int, fallback int) int {
	if p, ok := priority[provider]; ok {
		return p
	}
	return fallback
}

func (s *Service) providerOrder() []string {
	seen := map[string]bool{}
	out := []string{}
	for _, p := range []string{s.cfg.PrimaryProvider, s.cfg.FallbackProvider} {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

func (s *Service) providerOrderForSelection(providerName string) ([]string, string, error) {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	if providerName == "" || providerName == "default" {
		order := s.providerOrder()
		if len(order) == 0 {
			return nil, "default", fmt.Errorf("%w: no market data providers configured", ErrMarketDataProviderNotAvailable)
		}
		return order, "default", nil
	}
	if s.providers[providerName] == nil {
		return nil, providerName, fmt.Errorf("%w: %q", ErrMarketDataProviderNotAvailable, providerName)
	}
	return []string{providerName}, providerName, nil
}

func (s *Service) providerRole(providerName string) string {
	switch providerName {
	case s.cfg.PrimaryProvider:
		return "primary"
	case s.cfg.FallbackProvider:
		return "fallback"
	default:
		return "available"
	}
}

func normalizeSymbol(symbol string) (string, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return "", fmt.Errorf("%w: symbol is required", ErrInvalidMarketDataRequest)
	}
	if len(symbol) > maxSymbolLength {
		return "", fmt.Errorf("%w: symbol must be at most %d characters", ErrInvalidMarketDataRequest, maxSymbolLength)
	}
	if strings.IndexFunc(symbol, func(r rune) bool {
		return unicode.IsControl(r) || unicode.IsSpace(r)
	}) >= 0 {
		return "", fmt.Errorf("%w: symbol contains invalid characters", ErrInvalidMarketDataRequest)
	}
	return symbol, nil
}

func providerSymbol(actualProvider, providerName, symbol string) string {
	if actualProvider == providerName {
		return symbol
	}
	return ""
}

func failureMessage(providerName string, err error) string {
	if err == nil {
		return providerName + ": unknown failure"
	}
	var perr *domain.ProviderError
	if errors.As(err, &perr) && perr.Code != "" {
		return providerName + ": " + perr.Code
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return providerName + ": timeout"
	}
	if errors.Is(err, context.Canceled) {
		return providerName + ": cancelled"
	}
	return providerName + ": unavailable"
}
