package websocket

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/RomaLytar/yammi/services/gateway/internal/infrastructure/auth"
)

// allowedOrigins загружается из env ALLOWED_ORIGINS (через запятую).
var allowedOrigins = func() map[string]bool {
	origins := make(map[string]bool)
	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		for _, o := range strings.Split(v, ",") {
			origins[strings.TrimSpace(o)] = true
		}
	}
	return origins
}()

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		if len(allowedOrigins) == 0 {
			return true // fallback для dev без конфигурации
		}
		origin := r.Header.Get("Origin")
		return allowedOrigins[origin]
	},
}

// BoardAccessChecker проверяет членство пользователя в доске.
type BoardAccessChecker interface {
	IsMember(boardID, token string) bool
}

// ServeWS обрабатывает HTTP-запрос на апгрейд до WebSocket.
// Аутентификация через query-параметр token: /ws?token=<jwt>.
func ServeWS(hub *Hub, verifier *auth.JWTVerifier, checker BoardAccessChecker, w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	userID, err := verifier.VerifyToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws-handler: upgrade failed: %v", err)
		return
	}

	client := NewClient(hub, conn, userID, token, checker)
	hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
