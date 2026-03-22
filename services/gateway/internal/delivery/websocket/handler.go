package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/romanlovesweed/yammi/services/gateway/internal/infrastructure/auth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // разрешаем все origins (для dev)
	},
}

// ServeWS обрабатывает HTTP-запрос на апгрейд до WebSocket.
// Аутентификация через query-параметр token: /ws?token=<jwt>.
func ServeWS(hub *Hub, verifier *auth.JWTVerifier, w http.ResponseWriter, r *http.Request) {
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

	client := NewClient(hub, conn, userID)
	hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
