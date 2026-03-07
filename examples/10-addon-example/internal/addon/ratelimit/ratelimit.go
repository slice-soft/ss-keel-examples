// Package ratelimit provides a simple in-memory rate limiter addon for Keel.
// In a real project this package would live in its own repository and be
// installed via: keel add github.com/example/keel-addon-ratelimit
package ratelimit

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/core"
)

// Config holds rate limiter settings.
type Config struct {
	// Max is the maximum number of requests allowed per window.
	Max int
	// Window is the sliding window duration.
	Window time.Duration
	// KeyFunc extracts the rate-limit key from a request (defaults to IP).
	KeyFunc func(c *fiber.Ctx) string
}

// RateLimiter is the addon entry point.
type RateLimiter struct {
	cfg     Config
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	count    int
	resetsAt time.Time
}

// New creates a new RateLimiter with the provided config.
func New(cfg Config) *RateLimiter {
	if cfg.Max <= 0 {
		cfg.Max = 100
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c *fiber.Ctx) string { return c.IP() }
	}
	return &RateLimiter{
		cfg:     cfg,
		buckets: make(map[string]*bucket),
	}
}

// Middleware returns a Fiber handler that enforces the rate limit.
// Returns 429 Too Many Requests when the limit is exceeded.
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := rl.cfg.KeyFunc(c)
		now := time.Now()

		rl.mu.Lock()
		b, ok := rl.buckets[key]
		if !ok || now.After(b.resetsAt) {
			b = &bucket{count: 0, resetsAt: now.Add(rl.cfg.Window)}
			rl.buckets[key] = b
		}
		b.count++
		count := b.count
		resetsAt := b.resetsAt
		rl.mu.Unlock()

		c.Set("X-RateLimit-Limit", itoa(rl.cfg.Max))
		remaining := rl.cfg.Max - count
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-RateLimit-Remaining", itoa(remaining))
		c.Set("X-RateLimit-Reset", itoa(int(resetsAt.Unix())))

		if count > rl.cfg.Max {
			return &core.KError{
				Code:       "RATE_LIMITED",
				StatusCode: fiber.StatusTooManyRequests,
				Message:    "too many requests — slow down",
			}
		}
		return c.Next()
	}
}

// Stats returns the current request counts per key (useful for monitoring).
func (rl *RateLimiter) Stats() map[string]int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	out := make(map[string]int, len(rl.buckets))
	for k, b := range rl.buckets {
		out[k] = b.count
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n >= 10 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	pos--
	buf[pos] = byte('0' + n)
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
