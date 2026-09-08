package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	limit   rate.Limit
	burst   int
	entries map[string]*limiterEntry
	ttl     time.Duration
	keyFunc func(r *http.Request) string
}

func NewRateLimiter(rps, burst int, keyFunc func(r *http.Request) string) *RateLimiter {
	rl := &RateLimiter{
		limit:   rate.Limit(rps),
		burst:   burst,
		entries: make(map[string]*limiterEntry),
		ttl:     10 * time.Minute,
		keyFunc: keyFunc,
	}
	go rl.cleanup()
	return rl
}

func IPKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := rl.keyFunc(r)
			if !rl.allow(key) {
				w.Header().Set("Retry-After", "1")
				httpx.RespondError(w, r.Context(), httpx.APIError{
					Status: http.StatusTooManyRequests, Code: "rate_limited", Message: "Rate limit exceeded",
				}, nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	entry, ok := rl.entries[key]
	if !ok {
		entry = &limiterEntry{limiter: rate.NewLimiter(rl.limit, rl.burst), lastSeen: time.Now()}
		rl.entries[key] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for key, entry := range rl.entries {
			if time.Since(entry.lastSeen) > rl.ttl {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

func PartnerKey(prefix string) func(r *http.Request) string {
	return func(r *http.Request) string {
		if prefix != "" {
			return prefix
		}
		return IPKey(r)
	}
}

func RetryAfterSeconds(seconds int) string {
	return strconv.Itoa(seconds)
}
