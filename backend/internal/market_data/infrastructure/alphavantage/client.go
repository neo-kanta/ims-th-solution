package alphavantage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

const (
	defaultBaseURL                   = "https://www.alphavantage.co/query"
	defaultRetryAttempts             = 3
	defaultInitialBackoff            = 250 * time.Millisecond
	DefaultFreeTierRateLimitInterval = 12 * time.Second
)

type Provider struct {
	apiKey         string
	baseURL        string
	httpClient     *http.Client
	retryAttempts  int
	initialBackoff time.Duration
	limiter        *intervalLimiter
}

type Option func(*Provider)

func WithBaseURL(baseURL string) Option {
	return func(p *Provider) {
		if strings.TrimSpace(baseURL) != "" {
			p.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(p *Provider) {
		if client != nil {
			p.httpClient = client
		}
	}
}

func WithRetryPolicy(attempts int, initialBackoff time.Duration) Option {
	return func(p *Provider) {
		if attempts > 0 {
			p.retryAttempts = attempts
		}
		if initialBackoff >= 0 {
			p.initialBackoff = initialBackoff
		}
	}
}

func WithRateLimitInterval(interval time.Duration) Option {
	return func(p *Provider) {
		if interval <= 0 {
			p.limiter = nil
			return
		}
		p.limiter = &intervalLimiter{interval: interval}
	}
}

func NewProvider(apiKey string, timeout time.Duration, opts ...Option) *Provider {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	p := &Provider{
		apiKey:         strings.TrimSpace(apiKey),
		baseURL:        defaultBaseURL,
		httpClient:     &http.Client{Timeout: timeout},
		retryAttempts:  defaultRetryAttempts,
		initialBackoff: defaultInitialBackoff,
		// Alpha Vantage's commonly used free tier allows roughly five calls
		// per minute. Keeping this at a provider boundary prevents business
		// logic from depending on vendor-specific throttling.
		limiter: &intervalLimiter{interval: DefaultFreeTierRateLimitInterval},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func DefaultRateLimitInterval() time.Duration {
	return DefaultFreeTierRateLimitInterval
}

func (p *Provider) ProviderName() string {
	return domain.ProviderAlphaVantage
}

func (p *Provider) GetQuote(ctx context.Context, symbol string) (*domain.Quote, error) {
	if p.apiKey == "" {
		return nil, p.providerError("quote", "not_configured", "ALPHA_VANTAGE_API_KEY is not configured", 0, false, domain.ErrNotConfigured)
	}

	var payload globalQuoteResponse
	if err := p.fetch(ctx, "quote", map[string]string{
		"function": "GLOBAL_QUOTE",
		"symbol":   symbol,
	}, &payload); err != nil {
		return nil, err
	}

	gq := payload.GlobalQuote
	if len(gq) == 0 {
		return nil, p.providerError("quote", "no_data", "Alpha Vantage returned an empty Global Quote payload", 0, false, nil)
	}

	price, err := decimalFromString(gq["05. price"])
	if err != nil || !price.IsPositive() {
		return nil, p.providerError("quote", "invalid_payload", "Alpha Vantage quote price is missing or invalid", 0, false, err)
	}

	asOf := time.Now().UTC()
	if latest := strings.TrimSpace(gq["07. latest trading day"]); latest != "" {
		if d, err := time.Parse("2006-01-02", latest); err == nil {
			asOf = d.UTC()
		}
	}

	open, _ := decimalFromString(gq["02. open"])
	high, _ := decimalFromString(gq["03. high"])
	low, _ := decimalFromString(gq["04. low"])
	prevClose, _ := decimalFromString(gq["08. previous close"])
	change, _ := decimalFromString(gq["09. change"])
	changePct, _ := decimalFromPercent(gq["10. change percent"])
	volume, _ := strconv.ParseInt(strings.TrimSpace(gq["06. volume"]), 10, 64)

	return &domain.Quote{
		Symbol:        strings.ToUpper(strings.TrimSpace(gq["01. symbol"])),
		Provider:      p.ProviderName(),
		AssetType:     domain.AssetTypeUnknown,
		Price:         price,
		Open:          open,
		High:          high,
		Low:           low,
		PreviousClose: prevClose,
		Change:        change,
		ChangePercent: changePct,
		Volume:        volume,
		AsOf:          asOf,
		CapturedAt:    time.Now().UTC(),
	}, nil
}

func (p *Provider) GetDailyPrices(ctx context.Context, symbol string) ([]domain.PriceBar, error) {
	if p.apiKey == "" {
		return nil, p.providerError("history", "not_configured", "ALPHA_VANTAGE_API_KEY is not configured", 0, false, domain.ErrNotConfigured)
	}

	var payload dailyResponse
	if err := p.fetch(ctx, "history", map[string]string{
		"function":   "TIME_SERIES_DAILY",
		"symbol":     symbol,
		"outputsize": "compact",
	}, &payload); err != nil {
		return nil, err
	}

	if len(payload.TimeSeriesDaily) == 0 {
		return nil, p.providerError("history", "no_data", "Alpha Vantage returned an empty daily time-series payload", 0, false, nil)
	}

	out := make([]domain.PriceBar, 0, len(payload.TimeSeriesDaily))
	for dateStr, row := range payload.TimeSeriesDaily {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		open, err := decimalFromString(row.Open)
		if err != nil {
			continue
		}
		high, err := decimalFromString(row.High)
		if err != nil {
			continue
		}
		low, err := decimalFromString(row.Low)
		if err != nil {
			continue
		}
		closePrice, err := decimalFromString(row.Close)
		if err != nil {
			continue
		}
		volume, _ := strconv.ParseInt(strings.TrimSpace(row.Volume), 10, 64)
		out = append(out, domain.PriceBar{
			Symbol:   strings.ToUpper(strings.TrimSpace(symbol)),
			Provider: p.ProviderName(),
			Date:     date.UTC(),
			Open:     open,
			High:     high,
			Low:      low,
			Close:    closePrice,
			Volume:   volume,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Date.After(out[j].Date)
	})
	if len(out) == 0 {
		return nil, p.providerError("history", "invalid_payload", "Alpha Vantage daily payload had no parseable bars", 0, false, nil)
	}
	return out, nil
}

func (p *Provider) fetch(ctx context.Context, operation string, params map[string]string, target any) error {
	attempts := p.retryAttempts
	if attempts <= 0 {
		attempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if p.limiter != nil {
			if err := p.limiter.Wait(ctx); err != nil {
				return err
			}
		}

		err := p.fetchOnce(ctx, operation, params, target)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetriable(err) || attempt == attempts {
			break
		}
		if err := sleepBackoff(ctx, p.initialBackoff, attempt); err != nil {
			return err
		}
	}
	return lastErr
}

func (p *Provider) fetchOnce(ctx context.Context, operation string, params map[string]string, target any) error {
	endpoint, err := url.Parse(p.baseURL)
	if err != nil {
		return p.providerError(operation, "invalid_config", "invalid Alpha Vantage base URL", 0, false, err)
	}
	q := endpoint.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	q.Set("apikey", p.apiKey)
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return p.providerError(operation, "request_build_failed", "failed to build Alpha Vantage request", 0, false, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ims-th-solution/marketdata")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return p.providerError(operation, "request_failed", "Alpha Vantage request failed", 0, true, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return p.providerError(operation, "read_failed", "failed to read Alpha Vantage response", resp.StatusCode, true, err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return p.providerError(operation, "http_error", http.StatusText(resp.StatusCode), resp.StatusCode, resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500, nil)
	}

	if err := detectProviderMessage(body); err != nil {
		var perr *domain.ProviderError
		if ok := errorAs(err, &perr); ok {
			perr.Operation = operation
			return perr
		}
		return err
	}

	if err := json.Unmarshal(body, target); err != nil {
		return p.providerError(operation, "decode_failed", "failed to decode Alpha Vantage JSON", resp.StatusCode, false, err)
	}
	return nil
}

func detectProviderMessage(body []byte) error {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(body, &top); err != nil {
		return nil
	}
	for _, key := range []string{"Error Message", "Note", "Information"} {
		raw, ok := top[key]
		if !ok {
			continue
		}
		var msg string
		_ = json.Unmarshal(raw, &msg)
		code := "provider_message"
		retriable := false
		switch key {
		case "Error Message":
			code = "provider_error"
		case "Note":
			code = "rate_limited"
			retriable = false
		case "Information":
			code = "provider_information"
		}
		return &domain.ProviderError{
			Provider:  domain.ProviderAlphaVantage,
			Code:      code,
			Message:   strings.TrimSpace(msg),
			Retriable: retriable,
		}
	}
	return nil
}

func (p *Provider) providerError(operation, code, message string, statusCode int, retriable bool, cause error) *domain.ProviderError {
	return &domain.ProviderError{
		Provider:   p.ProviderName(),
		Operation:  operation,
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Retriable:  retriable,
		Cause:      cause,
	}
}

func isRetriable(err error) bool {
	var perr *domain.ProviderError
	if errorAs(err, &perr) {
		return perr.Retriable
	}
	return false
}

func errorAs(err error, target **domain.ProviderError) bool {
	if err == nil {
		return false
	}
	if perr, ok := err.(*domain.ProviderError); ok {
		*target = perr
		return true
	}
	return false
}

func sleepBackoff(ctx context.Context, initial time.Duration, attempt int) error {
	if initial <= 0 {
		return nil
	}
	delay := initial
	for i := 1; i < attempt; i++ {
		delay *= 2
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func decimalFromString(raw string) (decimal.Decimal, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return decimal.Zero, fmt.Errorf("empty decimal")
	}
	return decimal.NewFromString(raw)
}

func decimalFromPercent(raw string) (decimal.Decimal, error) {
	raw = strings.TrimSpace(strings.TrimSuffix(raw, "%"))
	if raw == "" {
		return decimal.Zero, fmt.Errorf("empty decimal percent")
	}
	return decimal.NewFromString(raw)
}

type intervalLimiter struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
}

func (l *intervalLimiter) Wait(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.interval <= 0 {
		return nil
	}
	now := time.Now()
	if l.last.IsZero() {
		l.last = now
		return nil
	}
	wait := l.interval - now.Sub(l.last)
	if wait <= 0 {
		l.last = now
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		l.last = time.Now()
		return nil
	}
}

type globalQuoteResponse struct {
	GlobalQuote map[string]string `json:"Global Quote"`
}

type dailyResponse struct {
	MetaData        map[string]string       `json:"Meta Data"`
	TimeSeriesDaily map[string]dailyDataRow `json:"Time Series (Daily)"`
}

type dailyDataRow struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}

var _ domain.MarketDataProvider = (*Provider)(nil)
