package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	delivery "github.com/RomaLytar/yammi/services/board/internal/delivery/grpc"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/database"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/metrics"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/nats"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

func main() {
	port := os.Getenv("BOARD_GRPC_PORT")
	if port == "" {
		port = "50053"
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

	// Repositories
	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)

	// Use Cases
	createBoardUC := usecase.NewCreateBoardUseCase(boardRepo, memberRepo, publisher)
	getBoardUC := usecase.NewGetBoardUseCase(boardRepo, memberRepo)
	listBoardsUC := usecase.NewListBoardsUseCase(boardRepo)
	updateBoardUC := usecase.NewUpdateBoardUseCase(boardRepo, memberRepo, publisher)
	deleteBoardUC := usecase.NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)

	addColumnUC := usecase.NewAddColumnUseCase(columnRepo, boardRepo, memberRepo, publisher)
	getColumnsUC := usecase.NewGetColumnsUseCase(columnRepo, memberRepo)
	updateColumnUC := usecase.NewUpdateColumnUseCase(columnRepo, boardRepo, memberRepo, publisher)
	deleteColumnUC := usecase.NewDeleteColumnUseCase(columnRepo, boardRepo, memberRepo, publisher)
	reorderColumnsUC := usecase.NewReorderColumnsUseCase(columnRepo, boardRepo, memberRepo, publisher)

	createCardUC := usecase.NewCreateCardUseCase(cardRepo, boardRepo, memberRepo, publisher)
	getCardUC := usecase.NewGetCardUseCase(cardRepo, memberRepo)
	getCardsUC := usecase.NewGetCardsUseCase(cardRepo, memberRepo)
	updateCardUC := usecase.NewUpdateCardUseCase(cardRepo, boardRepo, memberRepo, publisher)
	moveCardUC := usecase.NewMoveCardUseCase(cardRepo, boardRepo, memberRepo, publisher)
	deleteCardUC := usecase.NewDeleteCardUseCase(cardRepo, boardRepo, memberRepo, publisher)

	addMemberUC := usecase.NewAddMemberUseCase(boardRepo, memberRepo, publisher)
	removeMemberUC := usecase.NewRemoveMemberUseCase(boardRepo, memberRepo, publisher)
	listMembersUC := usecase.NewListMembersUseCase(boardRepo, memberRepo)

	// gRPC server
	handler := delivery.NewBoardServiceServer(
		createBoardUC,
		getBoardUC,
		listBoardsUC,
		updateBoardUC,
		deleteBoardUC,
		addColumnUC,
		getColumnsUC,
		updateColumnUC,
		deleteColumnUC,
		reorderColumnsUC,
		createCardUC,
		getCardUC,
		getCardsUC,
		updateCardUC,
		moveCardUC,
		deleteCardUC,
		addMemberUC,
		removeMemberUC,
		listMembersUC,
	)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(metrics.UnaryServerInterceptor()),
	)
	boardpb.RegisterBoardServiceServer(grpcServer, handler)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("board-service started on :%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("board-service shutting down...")
	grpcServer.GracefulStop()
}
