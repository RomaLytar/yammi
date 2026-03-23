package http

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// rateLimitEntry хранит состояние rate limiter для одного IP (token bucket).
type rateLimitEntry struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter — in-memory rate limiter на основе token bucket.
// Ключ — IP-адрес клиента.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	rate    float64 // токенов в секунду
	burst   int     // максимум токенов (размер bucket)
	stop    chan struct{}
}

// NewRateLimiter создаёт rate limiter с указанным лимитом запросов за period.
// Запускает фоновую горутину для очистки устаревших записей.
func NewRateLimiter(maxRequests int, period time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		rate:    float64(maxRequests) / period.Seconds(),
		burst:   maxRequests,
		stop:    make(chan struct{}),
	}

	go rl.cleanup()

	return rl
}

// Allow проверяет, разрешён ли запрос для данного IP.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[ip]

	if !exists {
		rl.entries[ip] = &rateLimitEntry{
			tokens:     float64(rl.burst) - 1, // один токен сразу тратим
			lastRefill: now,
		}
		return true
	}

	// Пополняем токены на основе прошедшего времени
	elapsed := now.Sub(entry.lastRefill).Seconds()
	entry.tokens += elapsed * rl.rate
	if entry.tokens > float64(rl.burst) {
		entry.tokens = float64(rl.burst)
	}
	entry.lastRefill = now

	if entry.tokens < 1 {
		return false
	}

	entry.tokens--
	return true
}

// cleanup периодически удаляет записи, у которых bucket полностью восстановлен
// (клиент давно не отправлял запросы). Запускается как горутина.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, entry := range rl.entries {
				elapsed := now.Sub(entry.lastRefill).Seconds()
				refilled := entry.tokens + elapsed*rl.rate
				if refilled >= float64(rl.burst) {
					delete(rl.entries, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stop:
			return
		}
	}
}

// Stop останавливает фоновую горутину очистки.
func (rl *RateLimiter) Stop() {
	close(rl.stop)
}

// RateLimitMiddleware оборачивает http.Handler, ограничивая количество запросов по IP.
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)

			if !limiter.Allow(ip) {
				writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitHandlerFunc оборачивает http.HandlerFunc напрямую (удобно для публичных роутов).
func RateLimitHandlerFunc(limiter *RateLimiter, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)

		if !limiter.Allow(ip) {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		handler(w, r)
	}
}

// extractIP извлекает IP-адрес клиента из RemoteAddr, отбрасывая порт.
// X-Forwarded-For и X-Real-IP не используются — без доверенного прокси они легко подделываются.
func extractIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
