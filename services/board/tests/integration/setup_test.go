package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// sharedDB — единственное подключение к PostgreSQL для ВСЕХ интеграционных тестов.
// Поднимается один раз в TestMain, убивается после завершения всех тестов.
var sharedDB *sql.DB

// containerCleanup хранит функцию для остановки Docker контейнера
var containerCleanup func()

// TestMain — единая точка входа для всех интеграционных тестов.
// Один контейнер PostgreSQL на весь пакет.
func TestMain(m *testing.M) {
	dsn, cleanup := setupContainer()
	containerCleanup = cleanup

	db, err := waitForDB(dsn, 15)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	sharedDB = db

	runMigrationsRaw(db)

	code := m.Run()

	sharedDB.Close()
	if containerCleanup != nil {
		containerCleanup()
	}
	os.Exit(code)
}

// getSharedDB возвращает единственное подключение к тестовой БД.
// Используется всеми тестами вместо per-test контейнеров.
func getSharedDB(t *testing.T) *sql.DB {
	t.Helper()
	if sharedDB == nil {
		t.Fatal("sharedDB is nil — TestMain did not initialize properly")
	}
	return sharedDB
}

// setupContainer поднимает PostgreSQL контейнер (или использует TEST_DATABASE_URL).
func setupContainer() (string, func()) {
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		log.Printf("Using external database: %s", dsn)
		return dsn, func() {}
	}

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "board_test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("Failed to get container port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/board_test?sslmode=disable", host, port.Port())
	log.Printf("PostgreSQL container ready at %s", dsn)

	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}

	return dsn, cleanup
}

// runMigrationsRaw применяет все миграции (вызывается из TestMain, не из тестов).
func runMigrationsRaw(db *sql.DB) {
	migrationFiles := []string{
		"../../migrations/000001_init.up.sql",
		"../../migrations/000002_board_search_sort.up.sql",
		"../../migrations/000003_card_creator_id.up.sql",
		"../../migrations/000004_column_updated_at.up.sql",
		"../../migrations/000005_activity_log.up.sql",
		"../../migrations/000006_attachments.up.sql",
		"../../migrations/000007_card_metadata.up.sql",
		"../../migrations/000008_labels.up.sql",
		"../../migrations/000010_card_links.up.sql",
		"../../migrations/000011_custom_fields.up.sql",
		"../../migrations/000012_automation_rules.up.sql",
		"../../migrations/000013_optimize_indexes.up.sql",
		"../../migrations/000014_board_settings_user_labels.up.sql",
	}

	for _, path := range migrationFiles {
		sql, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read migration %s: %v", path, err)
		}
		if _, err := db.Exec(string(sql)); err != nil {
			// Ignore "already exists" — migrations are idempotent
			_ = err
		}
	}
	log.Printf("Migrations applied (%d files)", len(migrationFiles))
}

// waitForDB ожидает доступности базы данных с retry.
func waitForDB(dsn string, maxRetries int) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if err = db.Ping(); err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return db, nil
	}
	return nil, fmt.Errorf("failed to connect after %d retries: %w", maxRetries, err)
}
