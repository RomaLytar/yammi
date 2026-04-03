package websocket

import (
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

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
			return false // reject если origins не сконфигурированы
		}
		origin := r.Header.Get("Origin")
		return allowedOrigins[origin]
	},
}

// BoardAccessChecker проверяет членство пользователя в доске.
type BoardAccessChecker interface {
	IsMember(boardID, token string) bool
}

// ConnLimiter ограничивает количество новых WebSocket-подключений по IP.
type ConnLimiter struct {
	mu      sync.Mutex
	entries map[string]*connEntry
	max     int
	window  time.Duration
}

type connEntry struct {
	count   int
	resetAt time.Time
}

func NewConnLimiter(maxPerWindow int, window time.Duration) *ConnLimiter {
	return &ConnLimiter{
		entries: make(map[string]*connEntry),
		max:     maxPerWindow,
		window:  window,
	}
}

func (cl *ConnLimiter) Allow(ip string) bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	now := time.Now()
	e, ok := cl.entries[ip]
	if !ok || now.After(e.resetAt) {
		cl.entries[ip] = &connEntry{count: 1, resetAt: now.Add(cl.window)}
		return true
	}
	if e.count >= cl.max {
		return false
	}
	e.count++
	return true
}

// wsConnLimiter — глобальный лимитер: 20 подключений в минуту на IP.
var wsConnLimiter = NewConnLimiter(20, time.Minute)

// ServeWS обрабатывает HTTP-запрос на апгрейд до WebSocket.
// Аутентификация: Authorization header (предпочтительно) или query-параметр token (fallback).
func ServeWS(hub *Hub, verifier *auth.JWTVerifier, checker BoardAccessChecker, w http.ResponseWriter, r *http.Request) {
	// Rate limiting по IP
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if ip == "" {
		ip = r.RemoteAddr
	}
	if !wsConnLimiter.Allow(ip) {
		http.Error(w, "too many connections", http.StatusTooManyRequests)
		return
	}

	// JWT: Authorization header (предпочтительно) или query-параметр token
	// (fallback для браузерного WebSocket API, который не поддерживает кастомные headers)
	token := ""
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
		token = strings.TrimPrefix(header, "Bearer ")
	}
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if token == "" {
		http.Error(w, "missing or invalid authorization", http.StatusUnauthorized)
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

	client := NewClient(hub, conn, userID, token, checker, verifier)
	hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
