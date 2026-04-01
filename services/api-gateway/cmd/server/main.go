package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	delivery "github.com/RomaLytar/yammi/services/api-gateway/internal/delivery/http"
	"github.com/RomaLytar/yammi/services/api-gateway/internal/infrastructure"
)

func main() {
	port := os.Getenv("GATEWAY_HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	authAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authAddr == "" {
		authAddr = "localhost:50051"
	}

	userAddr := os.Getenv("USER_GRPC_ADDR")
	if userAddr == "" {
		userAddr = "localhost:50052"
	}

	boardAddr := os.Getenv("BOARD_GRPC_ADDR")
	if boardAddr == "" {
		boardAddr = "localhost:50053"
	}

	commentAddr := os.Getenv("COMMENT_GRPC_ADDR")
	if commentAddr == "" {
		commentAddr = "localhost:50054"
	}

	notificationAddr := os.Getenv("NOTIFICATION_GRPC_ADDR")
	if notificationAddr == "" {
		notificationAddr = "localhost:50055"
	}

	grpcSharedSecret := os.Getenv("GRPC_SHARED_SECRET")
	if grpcSharedSecret == "" {
		log.Fatal("GRPC_SHARED_SECRET is required")
	}

	// gRPC clients
	clients, err := infrastructure.NewGRPCClients(authAddr, userAddr, boardAddr, commentAddr, notificationAddr, grpcSharedSecret)
	if err != nil {
		log.Fatalf("failed to create grpc clients: %v", err)
	}
	defer clients.Close()

	// JWT verifier (загружает публичный ключ из Auth Service)
	verifier := infrastructure.NewJWTVerifier(clients.AuthClient)

	// HTTP router
	router, shutdownLimiters := delivery.NewRouter(clients, verifier)

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("api-gateway started on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("api-gateway: failed to listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("api-gateway shutting down")
	shutdownLimiters()
	server.Close()
}
