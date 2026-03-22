package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RomaLytar/yammi/services/gateway/internal/delivery/websocket"
	"github.com/RomaLytar/yammi/services/gateway/internal/infrastructure/auth"
	"github.com/RomaLytar/yammi/services/gateway/internal/infrastructure/queue"
)

func main() {
	port := os.Getenv("WS_GATEWAY_PORT")
	if port == "" {
		port = "8081"
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	// JWT верификатор — загружает публичный ключ из env или через HTTP.
	verifier, err := auth.NewJWTVerifier()
	if err != nil {
		log.Fatalf("ws-gateway: failed to create JWT verifier: %v", err)
	}

	// Hub — управление соединениями и маршрутизация.
	hub := websocket.NewHub()
	go hub.Run()

	// NATS consumer — подписка на события.
	consumer, err := queue.NewConsumer(natsURL, hub)
	if err != nil {
		log.Fatalf("ws-gateway: failed to create NATS consumer: %v", err)
	}

	if err := consumer.Start(); err != nil {
		log.Fatalf("ws-gateway: failed to start NATS consumer: %v", err)
	}

	// HTTP маршруты.
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWS(hub, verifier, w, r)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("ws-gateway started on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ws-gateway: failed to listen: %v", err)
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("ws-gateway shutting down...")

	hub.Stop()
	consumer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("ws-gateway: shutdown error: %v", err)
	}

	log.Println("ws-gateway stopped")
}
