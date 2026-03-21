package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	notificationpb "github.com/romanlovesweed/yammi/services/notification/api/proto/v1"
	delivery "github.com/romanlovesweed/yammi/services/notification/internal/delivery/grpc"
	"github.com/romanlovesweed/yammi/services/notification/internal/infrastructure/database"
	natspkg "github.com/romanlovesweed/yammi/services/notification/internal/infrastructure/nats"
	"github.com/romanlovesweed/yammi/services/notification/internal/repository/postgres"
	"github.com/romanlovesweed/yammi/services/notification/internal/usecase"
)

func main() {
	port := os.Getenv("NOTIFICATION_GRPC_PORT")
	if port == "" {
		port = "50055"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL is required")
	}

	// Database
	db, err := database.NewPostgresDB(databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Migrations
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "/app/migrations"
	}
	if err := database.RunMigrations(db, migrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Repositories
	notificationRepo := postgres.NewNotificationRepo(db)
	settingsRepo := postgres.NewSettingsRepo(db)
	boardMemberRepo := postgres.NewBoardMemberRepo(db)
	nameCacheRepo := postgres.NewNameCacheRepo(db)

	// NATS consumer (создаём до publisher, чтобы получить JetStream context)
	// Сначала создаём usecase с nil publisher, потом обновим
	createUC := usecase.NewCreateNotificationUseCase(notificationRepo, settingsRepo, nil)

	consumer, err := natspkg.NewConsumer(natsURL, createUC, boardMemberRepo, nameCacheRepo)
	if err != nil {
		log.Fatalf("failed to create nats consumer: %v", err)
	}
	defer consumer.Close()

	// Создаём publisher из JetStream context consumer-а
	publisher := natspkg.NewPublisher(consumer.JetStream())

	// Пересоздаём createUC с publisher
	createUC = usecase.NewCreateNotificationUseCase(notificationRepo, settingsRepo, publisher)

	// Обновляем consumer с новым createUC
	consumer, err = natspkg.NewConsumer(natsURL, createUC, boardMemberRepo, nameCacheRepo)
	if err != nil {
		log.Fatalf("failed to recreate nats consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Start(); err != nil {
		log.Fatalf("failed to start nats consumer: %v", err)
	}

	// Usecases
	listUC := usecase.NewListNotificationsUseCase(notificationRepo)
	markReadUC := usecase.NewMarkReadUseCase(notificationRepo)
	markAllUC := usecase.NewMarkAllReadUseCase(notificationRepo)
	unreadUC := usecase.NewGetUnreadCountUseCase(notificationRepo)
	settingsUC := usecase.NewSettingsUseCase(settingsRepo)

	// gRPC server
	handler := delivery.NewHandler(listUC, markReadUC, markAllUC, unreadUC, settingsUC)
	grpcServer := grpc.NewServer()
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
