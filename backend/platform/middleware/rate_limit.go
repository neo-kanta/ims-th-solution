package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// RateLimiter provides in-memory sliding window rate limiting for login endpoints.
// For production at scale, replace with Redis-backed implementation.
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time // key -> list of attempt timestamps
	maxPerIP int
	window   time.Duration
}

// NewRateLimiter creates a new in-memory rate limiter.
func NewRateLimiter(maxPerIP int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		attempts: make(map[string][]time.Time),
		maxPerIP: maxPerIP,
		window:   window,
	}
	// Background cleanup every minute
	go rl.cleanup()
	return rl
}

// RecordAndCheck records an attempt and returns true if the request should be blocked.
func (rl *RateLimiter) RecordAndCheck(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Filter out old entries
	existing := rl.attempts[key]
	var recent []time.Time
	for _, t := range existing {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= rl.maxPerIP {
		rl.attempts[key] = recent
		return true // blocked
	}

	recent = append(recent, now)
	rl.attempts[key] = recent
	return false
}

// Reset clears rate limit state for a key (e.g., after successful login).
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)
		for key, attempts := range rl.attempts {
			var recent []time.Time
			for _, t := range attempts {
				if t.After(cutoff) {
					recent = append(recent, t)
				}
			}
			if len(recent) == 0 {
				delete(rl.attempts, key)
			} else {
				rl.attempts[key] = recent
			}
		}
		rl.mu.Unlock()
	}
}

// LoginRateLimit returns middleware that rate-limits by client IP.
func LoginRateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractClientIP(r)
			if limiter.RecordAndCheck(ip) {
				w.Header().Set("Retry-After", "60")
				httputil.JSON(w, http.StatusTooManyRequests, map[string]interface{}{
					"error":   "too many login attempts",
					"message": "Please wait before trying again",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractClientIP(r *http.Request) string {
	var ip string
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		ip = strings.TrimSpace(ips[0])
	} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		ip = realIP
	} else {
		ip = r.RemoteAddr
	}

	if strings.Contains(ip, ":") {
		host, _, err := net.SplitHostPort(ip)
		if err == nil {
			return host
		}
	}
	return ip
}
