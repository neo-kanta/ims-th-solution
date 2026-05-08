package persistence

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
)

type RedisQuoteCache struct {
	client *redis.Client
	prefix string
}

func NewRedisQuoteCache(client *redis.Client) *RedisQuoteCache {
	if client == nil {
		return nil
	}
	return &RedisQuoteCache{
		client: client,
		prefix: "marketdata:quote:",
	}
}

func (c *RedisQuoteCache) GetQuote(ctx context.Context, symbol string) (*domain.Quote, bool, error) {
	if c == nil || c.client == nil {
		return nil, false, nil
	}
	raw, err := c.client.Get(ctx, c.key(symbol)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}
	var q domain.Quote
	if err := json.Unmarshal(raw, &q); err != nil {
		return nil, false, err
	}
	return &q, true, nil
}

func (c *RedisQuoteCache) SetQuote(ctx context.Context, quote domain.Quote, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	quote.Cached = false
	raw, err := json.Marshal(quote)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(quote.Symbol), raw, ttl).Err()
}

func (c *RedisQuoteCache) key(symbol string) string {
	return c.prefix + strings.ToUpper(strings.TrimSpace(symbol))
}

var _ domain.QuoteCache = (*RedisQuoteCache)(nil)
