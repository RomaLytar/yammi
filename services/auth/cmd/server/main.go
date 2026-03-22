package main

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "github.com/romanlovesweed/yammi/services/auth/api/proto/v1"
	delivery "github.com/romanlovesweed/yammi/services/auth/internal/delivery/grpc"
	"github.com/romanlovesweed/yammi/services/auth/internal/infrastructure"
	"github.com/romanlovesweed/yammi/services/auth/internal/repository/postgres"
	"github.com/romanlovesweed/yammi/services/auth/internal/usecase"
)

func main() {
	port := os.Getenv("AUTH_GRPC_PORT")
	if port == "" {
		port = "50051"
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

	// JWT keys — если задан JWT_SEED, все реплики получат одинаковую пару ключей
	var privateKey ed25519.PrivateKey
	var publicKey ed25519.PublicKey
	if seed := os.Getenv("JWT_SEED"); seed != "" {
		priv, pub, err := infrastructure.KeyPairFromSeed(seed)
		if err != nil {
			log.Fatalf("failed to load key from JWT_SEED: %v", err)
		}
		privateKey, publicKey = priv, pub
		log.Println("jwt: using shared key from JWT_SEED")
	} else {
		priv, pub, err := infrastructure.GenerateKeyPair()
		if err != nil {
			log.Fatalf("failed to generate key pair: %v", err)
		}
		privateKey, publicKey = priv, pub
		log.Println("jwt: WARNING — generated ephemeral key pair (not suitable for multiple replicas)")
	}

	tokenGenerator := infrastructure.NewJWTGenerator(privateKey, publicKey, "yammi-auth", 15*time.Minute)

	// NATS publisher
	publisher, err := infrastructure.NewNATSPublisher(natsURL)
	if err != nil {
		log.Fatalf("failed to create nats publisher: %v", err)
	}
	defer publisher.Close()

	// Bcrypt worker pool
	bcryptCost := 10 // default
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		if c, err := strconv.Atoi(v); err == nil {
			bcryptCost = c
		}
	}
	hasher := infrastructure.NewBcryptPool(0, bcryptCost)

	// Repositories
	userRepo := postgres.NewUserRepo(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepo(db)

	// Usecase
	authUC := usecase.NewAuthUseCase(userRepo, refreshTokenRepo, tokenGenerator, publisher, hasher, 7*24*time.Hour)

	// gRPC server
	handler := delivery.NewAuthHandler(authUC)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(recoveryInterceptor()),
	)
	authpb.RegisterAuthServiceServer(grpcServer, handler)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("auth-service started on :%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("auth-service shutting down...")
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
