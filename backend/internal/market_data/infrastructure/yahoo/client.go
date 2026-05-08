package yahoo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

const (
	defaultBaseURL        = "https://query1.finance.yahoo.com"
	defaultRetryAttempts  = 2
	defaultInitialBackoff = 300 * time.Millisecond
)

// Provider uses Yahoo Finance's chart endpoint as a PoC/fallback source only.
// Yahoo Finance does not provide a stable official public finance API contract,
// so callers must keep Alpha Vantage or another official vendor as primary.
type Provider struct {
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

func NewProvider(timeout time.Duration, opts ...Option) *Provider {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	p := &Provider{
		baseURL:        defaultBaseURL,
		httpClient:     &http.Client{Timeout: timeout},
		retryAttempts:  defaultRetryAttempts,
		initialBackoff: defaultInitialBackoff,
		limiter:        &intervalLimiter{interval: time.Second},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Provider) ProviderName() string {
	return domain.ProviderYahoo
}

func (p *Provider) GetQuote(ctx context.Context, symbol string) (*domain.Quote, error) {
	result, err := p.fetchChart(ctx, "quote", symbol, "5d")
	if err != nil {
		return nil, err
	}

	price, ok := numberToDecimal(result.Meta.RegularMarketPrice)
	asOf := time.Now().UTC()
	if result.Meta.RegularMarketTime > 0 {
		asOf = time.Unix(result.Meta.RegularMarketTime, 0).UTC()
	}

	var latestIndex = -1
	if !ok {
		latestIndex, price, ok = latestClose(result)
		if ok && latestIndex >= 0 && latestIndex < len(result.Timestamp) {
			asOf = time.Unix(result.Timestamp[latestIndex], 0).UTC()
		}
	} else {
		latestIndex, _, _ = latestClose(result)
	}
	if !ok || !price.IsPositive() {
		return nil, p.providerError("quote", "no_data", "Yahoo chart response had no latest close price", 0, false, nil)
	}

	open, high, low := decimal.Zero, decimal.Zero, decimal.Zero
	if latestIndex >= 0 && len(result.Indicators.Quote) > 0 {
		q := result.Indicators.Quote[0]
		open, _ = numberAt(q.Open, latestIndex)
		high, _ = numberAt(q.High, latestIndex)
		low, _ = numberAt(q.Low, latestIndex)
	}
	prevClose, _ := numberToDecimal(result.Meta.ChartPreviousClose)

	return &domain.Quote{
		Symbol:        strings.ToUpper(strings.TrimSpace(result.Meta.Symbol)),
		Provider:      p.ProviderName(),
		AssetType:     domain.AssetTypeUnknown,
		Price:         price,
		Open:          open,
		High:          high,
		Low:           low,
		PreviousClose: prevClose,
		Currency:      strings.ToUpper(strings.TrimSpace(result.Meta.Currency)),
		AsOf:          asOf,
		CapturedAt:    time.Now().UTC(),
	}, nil
}

func (p *Provider) GetDailyPrices(ctx context.Context, symbol string) ([]domain.PriceBar, error) {
	result, err := p.fetchChart(ctx, "history", symbol, "1y")
	if err != nil {
		return nil, err
	}
	if len(result.Timestamp) == 0 || len(result.Indicators.Quote) == 0 {
		return nil, p.providerError("history", "no_data", "Yahoo chart response had no timestamps or quote bars", 0, false, nil)
	}

	quote := result.Indicators.Quote[0]
	var adj []maybeNumber
	if len(result.Indicators.AdjClose) > 0 {
		adj = result.Indicators.AdjClose[0].AdjClose
	}

	out := make([]domain.PriceBar, 0, len(result.Timestamp))
	for i, ts := range result.Timestamp {
		closePrice, ok := numberAt(quote.Close, i)
		if !ok || !closePrice.IsPositive() {
			continue
		}
		open, ok := numberAt(quote.Open, i)
		if !ok {
			open = closePrice
		}
		high, ok := numberAt(quote.High, i)
		if !ok {
			high = closePrice
		}
		low, ok := numberAt(quote.Low, i)
		if !ok {
			low = closePrice
		}
		volume, _ := intAt(quote.Volume, i)
		var adjusted *decimal.Decimal
		if v, ok := numberAt(adj, i); ok {
			adjusted = &v
		}
		out = append(out, domain.PriceBar{
			Symbol:        strings.ToUpper(strings.TrimSpace(result.Meta.Symbol)),
			Provider:      p.ProviderName(),
			Date:          time.Unix(ts, 0).UTC(),
			Open:          open,
			High:          high,
			Low:           low,
			Close:         closePrice,
			AdjustedClose: adjusted,
			Volume:        volume,
			Currency:      strings.ToUpper(strings.TrimSpace(result.Meta.Currency)),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Date.After(out[j].Date)
	})
	if len(out) == 0 {
		return nil, p.providerError("history", "invalid_payload", "Yahoo chart response had no parseable bars", 0, false, nil)
	}
	return out, nil
}

func (p *Provider) fetchChart(ctx context.Context, operation, symbol, chartRange string) (*chartResult, error) {
	attempts := p.retryAttempts
	if attempts <= 0 {
		attempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if p.limiter != nil {
			if err := p.limiter.Wait(ctx); err != nil {
				return nil, err
			}
		}
		result, err := p.fetchChartOnce(ctx, operation, symbol, chartRange)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !isRetriable(err) || attempt == attempts {
			break
		}
		if err := sleepBackoff(ctx, p.initialBackoff, attempt); err != nil {
			return nil, err
		}
	}
	return nil, lastErr
}

func (p *Provider) fetchChartOnce(ctx context.Context, operation, symbol, chartRange string) (*chartResult, error) {
	endpoint, err := url.Parse(p.baseURL + "/v8/finance/chart/" + url.PathEscape(symbol))
	if err != nil {
		return nil, p.providerError(operation, "invalid_config", "invalid Yahoo Finance base URL", 0, false, err)
	}
	q := endpoint.Query()
	q.Set("range", chartRange)
	q.Set("interval", "1d")
	q.Set("events", "history")
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, p.providerError(operation, "request_build_failed", "failed to build Yahoo Finance request", 0, false, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ims-th-solution/marketdata")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, p.providerError(operation, "request_failed", "Yahoo Finance chart request failed", 0, true, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, p.providerError(operation, "read_failed", "failed to read Yahoo Finance response", resp.StatusCode, true, err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, p.providerError(operation, "http_error", http.StatusText(resp.StatusCode), resp.StatusCode, resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500, nil)
	}

	var envelope chartEnvelope
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&envelope); err != nil {
		return nil, p.providerError(operation, "decode_failed", "failed to decode Yahoo Finance chart JSON", resp.StatusCode, false, err)
	}
	if envelope.Chart.Error != nil {
		return nil, p.providerError(operation, envelope.Chart.Error.Code, envelope.Chart.Error.Description, resp.StatusCode, false, nil)
	}
	if len(envelope.Chart.Result) == 0 {
		return nil, p.providerError(operation, "no_data", "Yahoo Finance chart response had no result entries", resp.StatusCode, false, nil)
	}
	return &envelope.Chart.Result[0], nil
}

func latestClose(result *chartResult) (int, decimal.Decimal, bool) {
	if result == nil || len(result.Indicators.Quote) == 0 {
		return -1, decimal.Zero, false
	}
	closeValues := result.Indicators.Quote[0].Close
	for i := len(closeValues) - 1; i >= 0; i-- {
		if price, ok := numberAt(closeValues, i); ok && price.IsPositive() {
			return i, price, true
		}
	}
	return -1, decimal.Zero, false
}

func numberAt(values []maybeNumber, idx int) (decimal.Decimal, bool) {
	if idx < 0 || idx >= len(values) {
		return decimal.Zero, false
	}
	return numberToDecimal(values[idx])
}

func intAt(values []maybeNumber, idx int) (int64, bool) {
	if idx < 0 || idx >= len(values) || values[idx] == nil {
		return 0, false
	}
	n, err := values[idx].Int64()
	if err != nil {
		return 0, false
	}
	return n, true
}

func numberToDecimal(n maybeNumber) (decimal.Decimal, bool) {
	if n == nil {
		return decimal.Zero, false
	}
	s := strings.TrimSpace(n.String())
	if s == "" || s == "null" {
		return decimal.Zero, false
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, false
	}
	return d, true
}

func (p *Provider) providerError(operation, code, message string, statusCode int, retriable bool, cause error) *domain.ProviderError {
	return &domain.ProviderError{
		Provider:   p.ProviderName(),
		Operation:  operation,
		Code:       strings.TrimSpace(code),
		Message:    strings.TrimSpace(message),
		StatusCode: statusCode,
		Retriable:  retriable,
		Cause:      cause,
	}
}

func isRetriable(err error) bool {
	perr, ok := err.(*domain.ProviderError)
	return ok && perr.Retriable
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

type maybeNumber = *json.Number

type chartEnvelope struct {
	Chart struct {
		Result []chartResult `json:"result"`
		Error  *chartError   `json:"error"`
	} `json:"chart"`
}

type chartError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type chartResult struct {
	Meta       chartMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators chartIndicators `json:"indicators"`
}

type chartMeta struct {
	Symbol             string      `json:"symbol"`
	Currency           string      `json:"currency"`
	RegularMarketTime  int64       `json:"regularMarketTime"`
	RegularMarketPrice maybeNumber `json:"regularMarketPrice"`
	ChartPreviousClose maybeNumber `json:"chartPreviousClose"`
}

type chartIndicators struct {
	Quote    []quoteSeries    `json:"quote"`
	AdjClose []adjCloseSeries `json:"adjclose"`
}

type quoteSeries struct {
	Open   []maybeNumber `json:"open"`
	High   []maybeNumber `json:"high"`
	Low    []maybeNumber `json:"low"`
	Close  []maybeNumber `json:"close"`
	Volume []maybeNumber `json:"volume"`
}

type adjCloseSeries struct {
	AdjClose []maybeNumber `json:"adjclose"`
}

var _ domain.MarketDataProvider = (*Provider)(nil)
