# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Yammi is a Trello-like task board built with Go microservices, lightweight DDD, and clean architecture. The project language (README, comments) is primarily Russian.

## Build & Run Commands

**Start all services (Docker Compose):**
```bash
docker compose up --build
```

**Build a single service:**
```bash
cd services/<service-name>
CGO_ENABLED=0 GOOS=linux go build -o ./bin/<service-name> ./cmd/server
```

**Generate protobuf (per service, example for auth):**
```bash
cd services/auth
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/v1/auth.proto
```

**Run tests (per service):**
```bash
cd services/<service-name>
go test ./...
```

**Run a single test:**
```bash
cd services/<service-name>
go test ./internal/domain/ -run TestFunctionName -v
```

**Lint:**
```bash
cd services/<service-name>
go vet ./...
```

**Dependency management (per service — each has its own go.mod):**
```bash
cd services/<service-name>
go mod tidy
```

## Architecture

Seven microservices communicating via gRPC (sync) and NATS (async events):

```
Client → API Gateway (:8080) → gRPC → [Auth :50051 | User :50052 | Board :50053 | Comment :50054]
                                                         ↓
                                                  NATS Event Bus
                                                  ↓              ↓
                                        Notification :50055   WebSocket Gateway :8081 → Client
```

- **API Gateway** — HTTP entry point, JWT verification (local, via public key), rate limiting. No business logic.
- **Auth Service** — Registration, login, JWT (EdDSA asymmetric keys), refresh/revoke tokens.
- **User Service** — User profiles.
- **Board Service** — Core domain. Boards, columns, cards. Redis cache, NATS event publishing, optimistic locking.
- **Comment Service** — Card comments.
- **Notification Service** — Async event consumer only. No sync API.
- **WebSocket Gateway** (`services/gateway`) — Async event consumer, pushes real-time updates to clients. Never calls other services synchronously.

Each service has its own PostgreSQL database (see `scripts/init-databases.sql`). Cross-service data access is only via gRPC.

Infrastructure: PostgreSQL 16, Redis 7, NATS 2 (JetStream). All defined in `docker-compose.yml`.

## Clean Architecture (per service)

Every service follows the same internal structure:

```
cmd/server/main.go          — entry point, DI wiring
internal/
  domain/                   — entities, business rules, errors (zero external deps)
  usecase/                  — orchestration, interfaces for repositories
  delivery/grpc/            — gRPC handlers (or websocket/ for gateway)
  repository/postgres/      — repository implementations
  infrastructure/           — DB connection, JWT, migrations, cache, queue adapters
api/proto/v1/               — protobuf definitions
migrations/                 — SQL migration files (000001_init.up.sql format)
```

**Key rule:** Domain has zero dependencies. Usecases define repository interfaces; infrastructure implements them. Business logic lives in domain entities, not usecases.

## DDD: Board as Aggregate Root

Board is the sole aggregate root in Board Service. Column and Card are value objects within Board — there are no separate `CardRepository` or `ColumnRepository`. All persistence goes through `BoardRepository`.

Invariants enforced in domain: card belongs to exactly one column, unique ordering within column, version increment on every change (optimistic locking).

## Conventions

- Each service is an independent Go module (`services/<name>/go.mod`) — Go 1.24 for auth, Go 1.23 for others.
- `pkg/` contains only shared contracts (event definitions, shared proto). No utilities or middleware — those are `internal` per service.
- Environment variables configure services: `DATABASE_URL`, `<SERVICE>_GRPC_PORT`, `REDIS_URL`, `NATS_URL`.
- Domain errors use typed sentinels: `ErrCardNotFound`, `ErrAccessDenied`, etc.
- Authorization (role checks: Owner/Member) happens in usecase layer, not delivery.
- Events carry `event_id` (UUID for idempotency), `event_version`, and `occurred_at`.
