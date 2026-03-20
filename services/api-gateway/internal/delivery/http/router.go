package http

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RomaLytar/yammi/services/api-gateway/internal/infrastructure"
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

	// Auth routes — refresh публичный (токен проверяется в Auth Service), revoke требует авторизацию
	mux.HandleFunc("POST /api/v1/auth/refresh", RateLimitHandlerFunc(refreshLimiter, auth.RefreshToken))
	mux.Handle("POST /api/v1/auth/revoke", rateLimit(requireAuth(http.HandlerFunc(auth.RevokeToken))))

	// User routes
	user := NewUserHandler(clients.UserClient, clients.AuthClient)
	mux.Handle("GET /api/v1/users/{id}", rateLimit(requireAuth(http.HandlerFunc(user.GetProfile))))
	mux.Handle("PUT /api/v1/users/{id}", rateLimit(requireAuth(http.HandlerFunc(OwnerOnly(user.UpdateProfile)))))
	mux.Handle("DELETE /api/v1/users/{id}", rateLimit(requireAuth(http.HandlerFunc(OwnerOnly(user.DeleteUser)))))

	// Board routes — все требуют auth
	board := NewBoardHandler(clients.BoardClient)
	mux.Handle("POST /api/v1/boards", rateLimit(requireAuth(http.HandlerFunc(board.CreateBoard))))
	mux.Handle("GET /api/v1/boards/{id}", rateLimit(requireAuth(http.HandlerFunc(board.GetBoard))))
	mux.Handle("GET /api/v1/boards", rateLimit(requireAuth(http.HandlerFunc(board.ListBoards))))
	mux.Handle("PUT /api/v1/boards/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateBoard))))
	mux.Handle("DELETE /api/v1/boards/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteBoard))))

	// Column routes
	mux.Handle("POST /api/v1/boards/{id}/columns", rateLimit(requireAuth(http.HandlerFunc(board.AddColumn))))
	mux.Handle("GET /api/v1/boards/{id}/columns", rateLimit(requireAuth(http.HandlerFunc(board.GetColumns))))
	mux.Handle("PUT /api/v1/boards/{id}/columns/reorder", rateLimit(requireAuth(http.HandlerFunc(board.ReorderColumns))))
	mux.Handle("PUT /api/v1/columns/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateColumn))))
	mux.Handle("DELETE /api/v1/columns/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteColumn))))

	// Card routes
	mux.Handle("POST /api/v1/columns/{id}/cards", rateLimit(requireAuth(http.HandlerFunc(board.CreateCard))))
	mux.Handle("GET /api/v1/cards/{id}", rateLimit(requireAuth(http.HandlerFunc(board.GetCard))))
	mux.Handle("GET /api/v1/columns/{id}/cards", rateLimit(requireAuth(http.HandlerFunc(board.GetCards))))
	mux.Handle("PUT /api/v1/cards/{id}", rateLimit(requireAuth(http.HandlerFunc(board.UpdateCard))))
	mux.Handle("PUT /api/v1/cards/{id}/move", rateLimit(requireAuth(http.HandlerFunc(board.MoveCard))))
	mux.Handle("DELETE /api/v1/cards/{id}", rateLimit(requireAuth(http.HandlerFunc(board.DeleteCard))))

	// Member routes
	mux.Handle("POST /api/v1/boards/{id}/members", rateLimit(requireAuth(http.HandlerFunc(board.AddMember))))
	mux.Handle("DELETE /api/v1/boards/{boardId}/members/{userId}", rateLimit(requireAuth(http.HandlerFunc(board.RemoveMember))))
	mux.Handle("GET /api/v1/boards/{id}/members", rateLimit(requireAuth(http.HandlerFunc(board.ListMembers))))

	shutdown := func() {
		registerLimiter.Stop()
		loginLimiter.Stop()
		refreshLimiter.Stop()
		defaultLimiter.Stop()
	}

	return mux, shutdown
}
