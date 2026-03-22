package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notificationpb "github.com/RomaLytar/yammi/services/notification/api/proto/v1"
	delivery "github.com/RomaLytar/yammi/services/notification/internal/delivery/grpc"
	"github.com/RomaLytar/yammi/services/notification/internal/infrastructure/cache"
	"github.com/RomaLytar/yammi/services/notification/internal/infrastructure/database"
	"github.com/RomaLytar/yammi/services/notification/internal/infrastructure/metrics"
	natspkg "github.com/RomaLytar/yammi/services/notification/internal/infrastructure/nats"
	redispkg "github.com/RomaLytar/yammi/services/notification/internal/infrastructure/redis"
	"github.com/RomaLytar/yammi/services/notification/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/notification/internal/usecase"
)

func main() {
	port := os.Getenv("NOTIFICATION_GRPC_PORT")
	if port == "" {
		port = "50055"
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "2112"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL is required")
	}

	// Goroutine count reporter
	go func() {
		for {
			metrics.Goroutines.Set(float64(runtime.NumGoroutine()))
			time.Sleep(5 * time.Second)
		}
	}()

	// Metrics HTTP server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Printf("metrics server started on :%s", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
			log.Fatalf("metrics server failed: %v", err)
		}
	}()

	// Database
	db, err := database.NewPostgresDB(databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Migrations (напрямую к PostgreSQL, не через PgBouncer)
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "/app/migrations"
	}
	migrationURL := os.Getenv("MIGRATION_DATABASE_URL")
	if migrationURL == "" {
		migrationURL = databaseURL
	}
	migrationDB, err := database.NewPostgresDB(migrationURL)
	if err != nil {
		log.Fatalf("failed to connect to migration database: %v", err)
	}
	if err := database.RunMigrations(migrationDB, migrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	migrationDB.Close()

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL is required")
	}

	// Repositories
	notificationRepo := postgres.NewNotificationRepo(db)
	settingsRepo := postgres.NewSettingsRepo(db)
	boardMemberRepo := postgres.NewBoardMemberRepo(db)
	boardEventRepo := postgres.NewBoardEventRepo(db)
	nameCacheRepo := postgres.NewNameCacheRepo(db)

	// Redis unread counter
	unreadCounter, err := redispkg.NewUnreadCounter(redisURL)
	if err != nil {
		log.Fatalf("failed to create unread counter: %v", err)
	}
	defer unreadCounter.Close()

	// Settings cache (decorator над settingsRepo)
	settingsCache := cache.NewSettingsCache(settingsRepo)

	// NATS consumer (создаём до publisher, чтобы получить JetStream context)
	createUC := usecase.NewCreateNotificationUseCase(notificationRepo, settingsCache, nil, boardEventRepo, unreadCounter, boardMemberRepo)

	consumer, err := natspkg.NewConsumer(natsURL, createUC, boardMemberRepo, nameCacheRepo, settingsCache)
	if err != nil {
		log.Fatalf("failed to create nats consumer: %v", err)
	}
	defer consumer.Close()

	// Создаём publisher из JetStream context consumer-а
	publisher := natspkg.NewPublisher(consumer.JetStream())

	// Обновляем createUC с publisher (без пересоздания NATS-соединения)
	createUC = usecase.NewCreateNotificationUseCase(notificationRepo, settingsCache, publisher, boardEventRepo, unreadCounter, boardMemberRepo)
	consumer.SetCreateUC(createUC)

	if err := consumer.Start(); err != nil {
		log.Fatalf("failed to start nats consumer: %v", err)
	}

	// Usecases
	listUC := usecase.NewListNotificationsUseCase(notificationRepo, boardEventRepo, boardMemberRepo)
	markReadUC := usecase.NewMarkReadUseCase(notificationRepo, boardEventRepo)
	markAllUC := usecase.NewMarkAllReadUseCase(notificationRepo, boardEventRepo, boardMemberRepo)
	unreadUC := usecase.NewGetUnreadCountUseCase(boardEventRepo, boardMemberRepo, notificationRepo)
	settingsUC := usecase.NewSettingsUseCase(settingsCache, publisher)

	// gRPC server
	handler := delivery.NewHandler(listUC, markReadUC, markAllUC, unreadUC, settingsUC)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryInterceptor(),
			metrics.UnaryServerInterceptor(),
		),
	)
	notificationpb.RegisterNotificationServiceServer(grpcServer, handler)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("notification-service started on :%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("notification-service shutting down...")
	grpcServer.GracefulStop()
}

func recoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC recovered in %s: %v\n%s", info.FullMethod, r, debug.Stack())
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
