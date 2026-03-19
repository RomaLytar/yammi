package http

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/romanlovesweed/yammi/services/api-gateway/internal/infrastructure"
)

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// NewRouter создаёт HTTP роутер. Возвращает handler и функцию shutdown для остановки rate limiter горутин.
func NewRouter(clients *infrastructure.GRPCClients, verifier *infrastructure.JWTVerifier) (http.Handler, func()) {
	mux := http.NewServeMux()
	requireAuth := AuthMiddleware(verifier)

	// Rate limiters — лимиты настраиваются через env, дефолт 50 req/min
	registerLimiter := NewRateLimiter(envInt("RATE_LIMIT_REGISTER", 50), time.Minute)
	loginLimiter := NewRateLimiter(envInt("RATE_LIMIT_LOGIN", 50), time.Minute)
	refreshLimiter := NewRateLimiter(envInt("RATE_LIMIT_REFRESH", 50), time.Minute)
	defaultLimiter := NewRateLimiter(envInt("RATE_LIMIT_DEFAULT", 50), time.Minute)
	rateLimit := RateLimitMiddleware(defaultLimiter)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth routes — публичные (с жёсткими лимитами)
	auth := NewAuthHandler(clients.AuthClient)
	mux.HandleFunc("POST /api/v1/auth/register", RateLimitHandlerFunc(registerLimiter, auth.Register))
	mux.HandleFunc("POST /api/v1/auth/login", RateLimitHandlerFunc(loginLimiter, auth.Login))
	mux.HandleFunc("GET /api/v1/auth/public-key", auth.GetPublicKey)

	// Auth routes — требуют авторизацию
	mux.Handle("POST /api/v1/auth/refresh", RateLimitMiddleware(refreshLimiter)(requireAuth(http.HandlerFunc(auth.RefreshToken))))
	mux.Handle("POST /api/v1/auth/revoke", rateLimit(requireAuth(http.HandlerFunc(auth.RevokeToken))))

	// User routes
	user := NewUserHandler(clients.UserClient, clients.AuthClient)
	mux.Handle("GET /api/v1/users/{id}", rateLimit(requireAuth(http.HandlerFunc(user.GetProfile))))
	mux.Handle("PUT /api/v1/users/{id}", rateLimit(requireAuth(http.HandlerFunc(OwnerOnly(user.UpdateProfile)))))
	mux.Handle("DELETE /api/v1/users/{id}", rateLimit(requireAuth(http.HandlerFunc(OwnerOnly(user.DeleteUser)))))

	shutdown := func() {
		registerLimiter.Stop()
		loginLimiter.Stop()
		refreshLimiter.Stop()
		defaultLimiter.Stop()
	}

	return mux, shutdown
}
