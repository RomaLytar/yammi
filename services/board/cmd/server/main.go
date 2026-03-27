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
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/cache"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/database"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/metrics"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/nats"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/storage"
	"github.com/RomaLytar/yammi/services/board/internal/repository/cached"
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

	// Redis membership cache (CQRS: event-driven, no TTL)
	redisURL := os.Getenv("REDIS_URL")
	var membershipCache *cache.MembershipCache
	if redisURL != "" {
		mc, err := cache.NewMembershipCache(redisURL)
		if err != nil {
			log.Printf("WARNING: Redis unavailable, running without cache: %v", err)
		} else {
			membershipCache = mc
			defer membershipCache.Close()

			// Cache consumer: синхронизирует Redis из NATS событий (DeliverAll replay)
			cacheConsumer, err := nats.NewCacheConsumer(natsURL, membershipCache)
			if err != nil {
				log.Printf("WARNING: cache consumer failed to start: %v", err)
			} else {
				defer cacheConsumer.Close()
				if err := cacheConsumer.Start(); err != nil {
					log.Printf("WARNING: cache consumer start failed: %v", err)
				}
			}
		}
	}

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
	pgMemberRepo := postgres.NewMembershipRepository(db)
	attachmentRepo := postgres.NewAttachmentRepository(db)
	activityRepo := postgres.NewActivityRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	cardLinkRepo := postgres.NewCardLinkRepository(db)
	checklistRepo := postgres.NewChecklistRepository(db)
	customFieldRepo := postgres.NewCustomFieldRepository(db)
	automationRuleRepo := postgres.NewAutomationRuleRepository(db)

	// Membership repository: Redis cache decorator over PostgreSQL.
	// Если Redis недоступен — используется чистый PostgreSQL.
	var memberRepo usecase.MembershipRepository
	if membershipCache != nil {
		memberRepo = cached.NewMembershipRepository(pgMemberRepo, membershipCache)
		log.Println("membership: using Redis cache + PostgreSQL fallback")
	} else {
		memberRepo = pgMemberRepo
		log.Println("membership: using PostgreSQL only (no Redis)")
	}

	// Sub-handlers (группируют use cases по доменным областям)
	boardsHandler := delivery.NewBoardCoreHandler(
		usecase.NewCreateBoardUseCase(boardRepo, memberRepo, publisher),
		usecase.NewGetBoardUseCase(boardRepo, memberRepo),
		usecase.NewListBoardsUseCase(boardRepo),
		usecase.NewUpdateBoardUseCase(boardRepo, memberRepo, publisher),
		usecase.NewDeleteBoardUseCase(boardRepo, memberRepo, publisher),
	)

	columnsHandler := delivery.NewColumnHandler(
		usecase.NewAddColumnUseCase(columnRepo, boardRepo, memberRepo, publisher),
		usecase.NewGetColumnsUseCase(columnRepo, memberRepo),
		usecase.NewUpdateColumnUseCase(columnRepo, boardRepo, memberRepo, publisher),
		usecase.NewDeleteColumnUseCase(columnRepo, boardRepo, memberRepo, publisher),
		usecase.NewReorderColumnsUseCase(columnRepo, boardRepo, memberRepo, publisher),
	)

	cardsHandler := delivery.NewCardHandler(
		usecase.NewCreateCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher),
		usecase.NewGetCardUseCase(cardRepo, memberRepo),
		usecase.NewGetCardsUseCase(cardRepo, memberRepo),
		usecase.NewUpdateCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher),
		usecase.NewMoveCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher),
		usecase.NewDeleteCardUseCase(cardRepo, boardRepo, memberRepo, publisher),
		usecase.NewAssignCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher),
		usecase.NewUnassignCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher),
		usecase.NewListCardActivityUseCase(activityRepo, memberRepo),
		cardRepo,
	)

	membersHandler := delivery.NewMemberHandler(
		usecase.NewAddMemberUseCase(boardRepo, memberRepo, publisher),
		usecase.NewRemoveMemberUseCase(boardRepo, cardRepo, memberRepo, publisher),
		usecase.NewListMembersUseCase(boardRepo, memberRepo),
	)

	attachmentsHandler := delivery.NewAttachmentHandler(
		usecase.NewUploadAttachmentUseCase(attachmentRepo, activityRepo, memberRepo, fileStorage, publisher),
		usecase.NewConfirmUploadUseCase(attachmentRepo, memberRepo, fileStorage),
		usecase.NewGetDownloadURLUseCase(attachmentRepo, memberRepo, fileStorage),
		usecase.NewListAttachmentsUseCase(attachmentRepo, memberRepo),
		usecase.NewDeleteAttachmentUseCase(attachmentRepo, activityRepo, memberRepo, fileStorage, publisher),
	)

	labelsHandler := delivery.NewLabelHandler(
		usecase.NewCreateLabelUseCase(labelRepo, memberRepo, publisher),
		usecase.NewListLabelsUseCase(labelRepo, memberRepo),
		usecase.NewUpdateLabelUseCase(labelRepo, memberRepo, publisher),
		usecase.NewDeleteLabelUseCase(labelRepo, memberRepo, publisher),
		usecase.NewAddLabelToCardUseCase(labelRepo, memberRepo, publisher),
		usecase.NewRemoveLabelFromCardUseCase(labelRepo, memberRepo, publisher),
		usecase.NewGetCardLabelsUseCase(labelRepo, memberRepo),
	)

	cardLinksHandler := delivery.NewCardLinkHandler(
		usecase.NewLinkCardsUseCase(cardLinkRepo, cardRepo, memberRepo, publisher),
		usecase.NewUnlinkCardsUseCase(cardLinkRepo, memberRepo, publisher),
		usecase.NewGetCardChildrenUseCase(cardLinkRepo, memberRepo),
		usecase.NewGetCardParentsUseCase(cardLinkRepo, memberRepo),
	)

	checklistHandler := delivery.NewChecklistHandler(
		usecase.NewCreateChecklistUseCase(checklistRepo, memberRepo, publisher),
		usecase.NewGetChecklistsUseCase(checklistRepo, memberRepo),
		usecase.NewUpdateChecklistUseCase(checklistRepo, memberRepo, publisher),
		usecase.NewDeleteChecklistUseCase(checklistRepo, memberRepo, publisher),
		usecase.NewCreateChecklistItemUseCase(checklistRepo, memberRepo),
		usecase.NewUpdateChecklistItemUseCase(checklistRepo, memberRepo),
		usecase.NewDeleteChecklistItemUseCase(checklistRepo, memberRepo),
		usecase.NewToggleChecklistItemUseCase(checklistRepo, memberRepo, publisher),
	)

	customFieldHandler := delivery.NewCustomFieldHandler(
		usecase.NewCreateCustomFieldUseCase(customFieldRepo, memberRepo, publisher),
		usecase.NewListCustomFieldsUseCase(customFieldRepo, memberRepo),
		usecase.NewUpdateCustomFieldUseCase(customFieldRepo, memberRepo, publisher),
		usecase.NewDeleteCustomFieldUseCase(customFieldRepo, memberRepo, publisher),
		usecase.NewSetCustomFieldValueUseCase(customFieldRepo, memberRepo, publisher),
		usecase.NewGetCardCustomFieldsUseCase(customFieldRepo, memberRepo),
	)

	automationHandler := delivery.NewAutomationHandler(
		usecase.NewCreateAutomationRuleUseCase(automationRuleRepo, memberRepo, publisher),
		usecase.NewListAutomationRulesUseCase(automationRuleRepo, memberRepo),
		usecase.NewUpdateAutomationRuleUseCase(automationRuleRepo, memberRepo, publisher),
		usecase.NewDeleteAutomationRuleUseCase(automationRuleRepo, memberRepo, publisher),
		usecase.NewGetAutomationHistoryUseCase(automationRuleRepo, memberRepo),
	)

	// gRPC server
	handler := delivery.NewBoardServiceServer(
		boardsHandler,
		columnsHandler,
		cardsHandler,
		membersHandler,
		attachmentsHandler,
		labelsHandler,
		cardLinksHandler,
		checklistHandler,
		customFieldHandler,
		automationHandler,
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
