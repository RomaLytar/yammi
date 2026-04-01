package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/RomaLytar/yammi/services/notification/internal/infrastructure/cache"
	"github.com/RomaLytar/yammi/services/notification/internal/infrastructure/metrics"
	natspkg "github.com/RomaLytar/yammi/services/notification/internal/infrastructure/nats"
	redispkg "github.com/RomaLytar/yammi/services/notification/internal/infrastructure/redis"
	"github.com/RomaLytar/yammi/services/notification/internal/infrastructure/database"
	"github.com/RomaLytar/yammi/services/notification/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/notification/internal/usecase"
)

func main() {
	log.SetPrefix("[notification-consumer] ")

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "2113"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL is required")
	}

	// Goroutine reporter
	go func() {
		for {
			metrics.Goroutines.Set(float64(runtime.NumGoroutine()))
			time.Sleep(5 * time.Second)
		}
	}()

	// Metrics + health
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})
		srv := &http.Server{
			Addr:         ":" + metricsPort,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		log.Printf("metrics server started on :%s", metricsPort)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("metrics server failed: %v", err)
		}
	}()

	// Database (consumer pool — через pgbouncer-consumer)
	db, err := database.NewPostgresDB(databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Redis
	unreadCounter, err := redispkg.NewUnreadCounter(redisURL)
	if err != nil {
		log.Fatalf("failed to create unread counter: %v", err)
	}
	defer unreadCounter.Close()

	// Repositories
	notificationRepo := postgres.NewNotificationRepo(db)
	settingsRepo := postgres.NewSettingsRepo(db)
	boardMemberRepo := postgres.NewBoardMemberRepo(db)
	boardEventRepo := postgres.NewBoardEventRepo(db)
	nameCacheRepo := postgres.NewNameCacheRepo(db)

	// Caches
	settingsCache := cache.NewSettingsCache(settingsRepo)
	nameCache := cache.NewInMemoryNameCache(nameCacheRepo)

	// NATS consumer (first without publisher)
	createUC := usecase.NewCreateNotificationUseCase(notificationRepo, settingsCache, nil, boardEventRepo, unreadCounter, boardMemberRepo)

	consumer, err := natspkg.NewConsumer(natsURL, createUC, boardMemberRepo, nameCache, settingsCache)
	if err != nil {
		log.Fatalf("failed to create nats consumer: %v", err)
	}

	// Publisher from JetStream context
	publisher := natspkg.NewPublisher(consumer.JetStream())

	// Recreate usecase with publisher
	createUC = usecase.NewCreateNotificationUseCase(notificationRepo, settingsCache, publisher, boardEventRepo, unreadCounter, boardMemberRepo)
	consumer.SetCreateUC(createUC)

	if err := consumer.Start(); err != nil {
		log.Fatalf("failed to start nats consumer: %v", err)
	}

	log.Println("consumer started, processing events...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	// Drain гарантирует обработку уже полученных сообщений
	consumer.Close()
}
