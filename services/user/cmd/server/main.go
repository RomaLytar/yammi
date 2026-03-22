package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userpb "github.com/RomaLytar/yammi/services/user/api/proto/v1"
	delivery "github.com/RomaLytar/yammi/services/user/internal/delivery/grpc"
	"github.com/RomaLytar/yammi/services/user/internal/infrastructure"
	"github.com/RomaLytar/yammi/services/user/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/user/internal/usecase"
)

func main() {
	port := os.Getenv("USER_GRPC_PORT")
	if port == "" {
		port = "50052"
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
	db, err := infrastructure.NewPostgresDB(databaseURL)
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
	migrationDB, err := infrastructure.NewPostgresDB(migrationURL)
	if err != nil {
		log.Fatalf("failed to connect to migration database: %v", err)
	}
	if err := infrastructure.RunMigrations(migrationDB, migrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	migrationDB.Close()

	// Repository
	userRepo := postgres.NewUserRepo(db)

	// Usecase
	userUC := usecase.NewUserUseCase(userRepo)

	// NATS consumer
	consumer, err := infrastructure.NewNATSConsumer(natsURL, userUC)
	if err != nil {
		log.Fatalf("failed to create nats consumer: %v", err)
	}
	defer consumer.Close()

	if err := consumer.Start(); err != nil {
		log.Fatalf("failed to start nats consumer: %v", err)
	}

	// DLQ monitor
	dlqMonitor := infrastructure.NewDLQMonitor(consumer.JetStream())
	if err := dlqMonitor.Start(); err != nil {
		log.Fatalf("failed to start dlq monitor: %v", err)
	}
	defer dlqMonitor.Close()

	// gRPC server
	handler := delivery.NewUserHandler(userUC)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(recoveryInterceptor()),
	)
	userpb.RegisterUserServiceServer(grpcServer, handler)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("user-service started on :%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("user-service shutting down...")
	grpcServer.GracefulStop()
}

func recoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC recovered in %s: %v\n%s", info.FullMethod, r, debug.Stack())
				err = status.Errorf(codes.Internal, fmt.Sprintf("internal error: %v", r))
			}
		}()
		return handler(ctx, req)
	}
}
