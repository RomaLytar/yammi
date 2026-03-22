package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commentpb "github.com/RomaLytar/yammi/services/comment/api/proto/v1"
	delivery "github.com/RomaLytar/yammi/services/comment/internal/delivery/grpc"
	"github.com/RomaLytar/yammi/services/comment/internal/infrastructure/database"
	boardclient "github.com/RomaLytar/yammi/services/comment/internal/infrastructure/grpc"
	"github.com/RomaLytar/yammi/services/comment/internal/infrastructure/metrics"
	"github.com/RomaLytar/yammi/services/comment/internal/infrastructure/nats"
	"github.com/RomaLytar/yammi/services/comment/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/comment/internal/usecase"
)

func main() {
	port := os.Getenv("COMMENT_GRPC_PORT")
	if port == "" {
		port = "50054"
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

	boardGRPCAddr := os.Getenv("BOARD_GRPC_ADDR")
	if boardGRPCAddr == "" {
		boardGRPCAddr = "board:50053"
	}

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

	// Migrations (напрямую к PostgreSQL, не через PgBouncer — advisory locks не работают в transaction mode)
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

	// NATS publisher
	natsPublisher, err := nats.NewPublisher(natsURL)
	if err != nil {
		log.Fatalf("failed to create nats publisher: %v", err)
	}
	defer natsPublisher.Close()

	// Event publisher (wraps NATS publisher)
	publisher := nats.NewEventPublisher(natsPublisher)

	// Board membership checker (gRPC client)
	membershipChecker, err := boardclient.NewBoardMembershipChecker(boardGRPCAddr)
	if err != nil {
		log.Fatalf("failed to create board membership checker: %v", err)
	}
	defer membershipChecker.Close()

	// Repository
	commentRepo := postgres.NewCommentRepository(db)

	// Use Cases
	createCommentUC := usecase.NewCreateCommentUseCase(commentRepo, membershipChecker, publisher)
	listCommentsUC := usecase.NewListCommentsUseCase(commentRepo, membershipChecker)
	updateCommentUC := usecase.NewUpdateCommentUseCase(commentRepo, membershipChecker, publisher)
	deleteCommentUC := usecase.NewDeleteCommentUseCase(commentRepo, membershipChecker, publisher)
	getCommentCountUC := usecase.NewGetCommentCountUseCase(commentRepo, membershipChecker)

	// gRPC server
	handler := delivery.NewCommentServiceServer(
		createCommentUC,
		listCommentsUC,
		updateCommentUC,
		deleteCommentUC,
		getCommentCountUC,
	)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryInterceptor(),
			metrics.UnaryServerInterceptor(),
		),
	)
	commentpb.RegisterCommentServiceServer(grpcServer, handler)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("comment-service started on :%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("comment-service shutting down...")
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
