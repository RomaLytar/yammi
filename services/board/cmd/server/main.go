package main

import (
	"context"
	"log"
	"net"
	"runtime/debug"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	delivery "github.com/RomaLytar/yammi/services/board/internal/delivery/grpc"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/database"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/metrics"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/nats"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/storage"
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

	minioURL := os.Getenv("MINIO_URL")
	if minioURL == "" {
		minioURL = "minio:9000"
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "yammi"
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "yammipass"
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

	// MinIO storage
	minioPublicURL := os.Getenv("MINIO_PUBLIC_URL")
	if minioPublicURL == "" {
		minioPublicURL = "localhost:9000"
	}
	fileStorage, err := storage.NewMinIOStorage(minioURL, minioPublicURL, minioAccessKey, minioSecretKey, false)
	if err != nil {
		log.Fatalf("failed to create minio storage: %v", err)
	}

	// Repositories
	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	attachmentRepo := postgres.NewAttachmentRepository(db)
	activityRepo := postgres.NewActivityRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	cardLinkRepo := postgres.NewCardLinkRepository(db)
	checklistRepo := postgres.NewChecklistRepository(db)
	customFieldRepo := postgres.NewCustomFieldRepository(db)
	automationRuleRepo := postgres.NewAutomationRuleRepository(db)

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

	createCardUC := usecase.NewCreateCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher)
	getCardUC := usecase.NewGetCardUseCase(cardRepo, memberRepo)
	getCardsUC := usecase.NewGetCardsUseCase(cardRepo, memberRepo)
	updateCardUC := usecase.NewUpdateCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher)
	moveCardUC := usecase.NewMoveCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher)
	deleteCardUC := usecase.NewDeleteCardUseCase(cardRepo, boardRepo, memberRepo, publisher)

	assignCardUC := usecase.NewAssignCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher)
	unassignCardUC := usecase.NewUnassignCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher)
	listCardActivityUC := usecase.NewListCardActivityUseCase(activityRepo, memberRepo)

	addMemberUC := usecase.NewAddMemberUseCase(boardRepo, memberRepo, publisher)
	removeMemberUC := usecase.NewRemoveMemberUseCase(boardRepo, cardRepo, memberRepo, publisher)
	listMembersUC := usecase.NewListMembersUseCase(boardRepo, memberRepo)

	uploadAttachmentUC := usecase.NewUploadAttachmentUseCase(attachmentRepo, activityRepo, memberRepo, fileStorage, publisher)
	confirmUploadUC := usecase.NewConfirmUploadUseCase(attachmentRepo, memberRepo, fileStorage)
	getDownloadURLUC := usecase.NewGetDownloadURLUseCase(attachmentRepo, memberRepo, fileStorage)
	listAttachmentsUC := usecase.NewListAttachmentsUseCase(attachmentRepo, memberRepo)
	deleteAttachmentUC := usecase.NewDeleteAttachmentUseCase(attachmentRepo, activityRepo, memberRepo, fileStorage, publisher)

	createLabelUC := usecase.NewCreateLabelUseCase(labelRepo, memberRepo, publisher)
	listLabelsUC := usecase.NewListLabelsUseCase(labelRepo, memberRepo)
	updateLabelUC := usecase.NewUpdateLabelUseCase(labelRepo, memberRepo, publisher)
	deleteLabelUC := usecase.NewDeleteLabelUseCase(labelRepo, memberRepo, publisher)
	addLabelToCardUC := usecase.NewAddLabelToCardUseCase(labelRepo, memberRepo, publisher)
	removeLabelFromCardUC := usecase.NewRemoveLabelFromCardUseCase(labelRepo, memberRepo, publisher)
	getCardLabelsUC := usecase.NewGetCardLabelsUseCase(labelRepo, memberRepo)

	linkCardsUC := usecase.NewLinkCardsUseCase(cardLinkRepo, cardRepo, memberRepo, publisher)
	unlinkCardsUC := usecase.NewUnlinkCardsUseCase(cardLinkRepo, memberRepo, publisher)
	getCardChildrenUC := usecase.NewGetCardChildrenUseCase(cardLinkRepo, memberRepo)
	getCardParentsUC := usecase.NewGetCardParentsUseCase(cardLinkRepo, memberRepo)

	createChecklistUC := usecase.NewCreateChecklistUseCase(checklistRepo, memberRepo, publisher)
	getChecklistsUC := usecase.NewGetChecklistsUseCase(checklistRepo, memberRepo)
	updateChecklistUC := usecase.NewUpdateChecklistUseCase(checklistRepo, memberRepo, publisher)
	deleteChecklistUC := usecase.NewDeleteChecklistUseCase(checklistRepo, memberRepo, publisher)
	createChecklistItemUC := usecase.NewCreateChecklistItemUseCase(checklistRepo, memberRepo)
	updateChecklistItemUC := usecase.NewUpdateChecklistItemUseCase(checklistRepo, memberRepo)
	deleteChecklistItemUC := usecase.NewDeleteChecklistItemUseCase(checklistRepo, memberRepo)
	toggleChecklistItemUC := usecase.NewToggleChecklistItemUseCase(checklistRepo, memberRepo, publisher)

	createCustomFieldUC := usecase.NewCreateCustomFieldUseCase(customFieldRepo, memberRepo, publisher)
	listCustomFieldsUC := usecase.NewListCustomFieldsUseCase(customFieldRepo, memberRepo)
	updateCustomFieldUC := usecase.NewUpdateCustomFieldUseCase(customFieldRepo, memberRepo, publisher)
	deleteCustomFieldUC := usecase.NewDeleteCustomFieldUseCase(customFieldRepo, memberRepo, publisher)
	setCustomFieldValueUC := usecase.NewSetCustomFieldValueUseCase(customFieldRepo, memberRepo, publisher)
	getCardCustomFieldsUC := usecase.NewGetCardCustomFieldsUseCase(customFieldRepo, memberRepo)

	createAutomationRuleUC := usecase.NewCreateAutomationRuleUseCase(automationRuleRepo, memberRepo, publisher)
	listAutomationRulesUC := usecase.NewListAutomationRulesUseCase(automationRuleRepo, memberRepo)
	updateAutomationRuleUC := usecase.NewUpdateAutomationRuleUseCase(automationRuleRepo, memberRepo, publisher)
	deleteAutomationRuleUC := usecase.NewDeleteAutomationRuleUseCase(automationRuleRepo, memberRepo, publisher)
	getAutomationHistoryUC := usecase.NewGetAutomationHistoryUseCase(automationRuleRepo, memberRepo)

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
		assignCardUC,
		unassignCardUC,
		listCardActivityUC,
		addMemberUC,
		removeMemberUC,
		listMembersUC,
		cardRepo,
		uploadAttachmentUC,
		confirmUploadUC,
		getDownloadURLUC,
		listAttachmentsUC,
		deleteAttachmentUC,
		createLabelUC,
		listLabelsUC,
		updateLabelUC,
		deleteLabelUC,
		addLabelToCardUC,
		removeLabelFromCardUC,
		getCardLabelsUC,
		linkCardsUC,
		unlinkCardsUC,
		getCardChildrenUC,
		getCardParentsUC,
		createChecklistUC,
		getChecklistsUC,
		updateChecklistUC,
		deleteChecklistUC,
		createChecklistItemUC,
		updateChecklistItemUC,
		deleteChecklistItemUC,
		toggleChecklistItemUC,
		createCustomFieldUC,
		listCustomFieldsUC,
		updateCustomFieldUC,
		deleteCustomFieldUC,
		setCustomFieldValueUC,
		getCardCustomFieldsUC,
		createAutomationRuleUC,
		listAutomationRulesUC,
		updateAutomationRuleUC,
		deleteAutomationRuleUC,
		getAutomationHistoryUC,
	)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryInterceptor(),
			metrics.UnaryServerInterceptor(),
		),
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
