package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiter defines the interface for rate limiting.
// Implementations must be safe for concurrent use.
type RateLimiter interface {
	Allow(ctx context.Context, key string) (allowed bool, count int, retryAfter int, err error)
	GetAttempts(ctx context.Context, key string) (int, error)
	Reset(ctx context.Context, key string) error
}

// RateLimitPolicy describes a named rate limit tier.
type RateLimitPolicy struct {
	Name   string
	Max    int
	Window time.Duration
}

// InMemoryRateLimiter is a fixed-window, process-local rate limiter.
// Suitable for single-instance deployments and development.
// For multi-instance production, swap to a Redis-backed implementation.
type InMemoryRateLimiter struct {
	attempts map[string]*attemptRecord
	mu       sync.RWMutex
	window   time.Duration
	maxReqs  int
}

type attemptRecord struct {
	count     int
	firstSeen time.Time
}

// NewInMemoryRateLimiter creates an in-memory rate limiter.
func NewInMemoryRateLimiter(window time.Duration, maxReqs int) RateLimiter {
	limiter := &InMemoryRateLimiter{
		attempts: make(map[string]*attemptRecord),
		window:   window,
		maxReqs:  maxReqs,
	}
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()
	return limiter
}

// Allow checks if request is allowed under rate limit.
// Returns (allowed, currentCount, retryAfterSeconds, error).
func (r *InMemoryRateLimiter) Allow(_ context.Context, key string) (bool, int, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	record, exists := r.attempts[key]
	now := time.Now()
	if !exists || now.Sub(record.firstSeen) > r.window {
		r.attempts[key] = &attemptRecord{count: 1, firstSeen: now}
		retryAfter := int(r.window.Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		return true, 1, retryAfter, nil
	}
	retryAfter := int(record.firstSeen.Add(r.window).Sub(now).Seconds()) + 1
	if retryAfter < 1 {
		retryAfter = 1
	}
	record.count++
	if record.count > r.maxReqs {
		return false, record.count, retryAfter, nil
	}
	return true, record.count, retryAfter, nil
}

// GetAttempts returns the number of attempts for a key.
func (r *InMemoryRateLimiter) GetAttempts(_ context.Context, key string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	record, exists := r.attempts[key]
	if !exists {
		return 0, nil
	}
	if time.Since(record.firstSeen) > r.window {
		return 0, nil
	}
	return record.count, nil
}

// Reset clears the attempt counter for a key.
func (r *InMemoryRateLimiter) Reset(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.attempts, key)
	return nil
}

// WindowDuration returns the configured window for header calculations.
func (r *InMemoryRateLimiter) WindowDuration() time.Duration {
	return r.window
}

// MaxRequests returns the configured max for header calculations.
func (r *InMemoryRateLimiter) MaxRequests() int {
	return r.maxReqs
}

func (r *InMemoryRateLimiter) cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for key, record := range r.attempts {
		if now.Sub(record.firstSeen) > r.window {
			delete(r.attempts, key)
		}
	}
}

// RateLimit creates middleware that limits requests by a key extracted from the request.
// keyFunc determines the rate-limit key (e.g., client IP, user ID, compound key).
func RateLimitByKey(limiter RateLimiter, policy RateLimitPolicy, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			allowed, count, retryAfter, err := limiter.Allow(r.Context(), key)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "rate limit check failed"})
				return
			}

			remaining := policy.Max - count
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(policy.Max))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(retryAfter)*time.Second).Unix(), 10))
			w.Header().Set("X-RateLimit-Policy", policy.Name)

			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":       "rate limit exceeded",
					"retry_after": retryAfter,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit creates middleware that limits requests per client IP.
// This is the simple per-IP variant used for global and login rate limiting.
func RateLimit(limiter RateLimiter, policy RateLimitPolicy) func(http.Handler) http.Handler {
	return RateLimitByKey(limiter, policy, func(r *http.Request) string {
		return fmt.Sprintf("rl:%s:%s", policy.Name, GetClientIP(r))
	})
}

// RateLimitByUser creates middleware that limits requests per authenticated user.
// Falls back to client IP if no JWT claims are present.
func RateLimitByUser(limiter RateLimiter, policy RateLimitPolicy) func(http.Handler) http.Handler {
	return RateLimitByKey(limiter, policy, func(r *http.Request) string {
		claims := GetUserClaims(r.Context())
		if claims != nil {
			return fmt.Sprintf("rl:%s:user:%s", policy.Name, claims.Subject)
		}
		return fmt.Sprintf("rl:%s:ip:%s", policy.Name, GetClientIP(r))
	})
}

// TrustedProxyConfig holds trusted proxy CIDR ranges.
type TrustedProxyConfig struct {
	nets []*net.IPNet
}

// NewTrustedProxyConfig parses a list of CIDR strings.
func NewTrustedProxyConfig(cidrs []string) *TrustedProxyConfig {
	var nets []*net.IPNet
	for _, cidr := range cidrs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		nets = append(nets, ipNet)
	}
	return &TrustedProxyConfig{nets: nets}
}

// IsTrusted returns true if the given IP is in the trusted proxy list.
func (t *TrustedProxyConfig) IsTrusted(ip net.IP) bool {
	if ip == nil || len(t.nets) == 0 {
		return false
	}
	for _, ipNet := range t.nets {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// trustedProxies is the global trusted proxy config. Set via SetTrustedProxies.
var trustedProxies = &TrustedProxyConfig{}

// SetTrustedProxies configures the global trusted proxy list.
func SetTrustedProxies(config *TrustedProxyConfig) {
	if config != nil {
		trustedProxies = config
	}
}

// GetClientIP extracts the client IP from the request.
func GetClientIP(r *http.Request) string {
	directIP := remoteIP(r)

	if trustedProxies.IsTrusted(net.ParseIP(directIP)) {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ips := strings.Split(forwarded, ",")
			if len(ips) > 0 {
				clientIP := strings.TrimSpace(ips[0])
				if clientIP != "" {
					return clientIP
				}
			}
		}
		if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			return realIP
		}
	}

	return directIP
}

// remoteIP extracts the IP from RemoteAddr, stripping port.
func remoteIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
