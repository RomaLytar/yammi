package http

import (
	"net/http"

	"github.com/romanlovesweed/yammi/services/api-gateway/internal/infrastructure"
)

func NewRouter(clients *infrastructure.GRPCClients) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth routes
	auth := NewAuthHandler(clients.AuthClient)
	mux.HandleFunc("POST /api/v1/auth/register", auth.Register)
	mux.HandleFunc("POST /api/v1/auth/login", auth.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", auth.RefreshToken)
	mux.HandleFunc("POST /api/v1/auth/revoke", auth.RevokeToken)
	mux.HandleFunc("GET /api/v1/auth/public-key", auth.GetPublicKey)

	// User routes
	user := NewUserHandler(clients.UserClient, clients.AuthClient)
	mux.HandleFunc("GET /api/v1/users/{id}", user.GetProfile)
	mux.HandleFunc("PUT /api/v1/users/{id}", user.UpdateProfile)
	mux.HandleFunc("DELETE /api/v1/users/{id}", user.DeleteUser)

	return mux
}
