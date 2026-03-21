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

## Testing

**223 теста, 8 параллельных потоков, ~2s total.**

### Unit тесты (158 тестов)

**Domain** (`services/board/internal/domain/*_test.go`) — **95.9% coverage**, ~0.2s:
- `card_test.go` — NewCard валидация, Update, Move, Reorder, lexorank validation, creator_id
- `board_test.go` — NewBoard, Update, IsOwner, IncrementVersion
- `column_test.go` — NewColumn, Update, UpdatePosition
- `member_test.go` — NewMember, CanModifyBoard, CanModifyCards, роли
- `lexorank_test.go` — LexorankBetween (инварианты: prev < result < next), ValidateLexorank, ordering, sequences, edge cases

**Usecase** (`services/board/internal/usecase/*_test.go`) — **46.9% coverage**, ~0.01s:
- `create_board_test.go` — создание, пустой title, пустой ownerID, ошибка repo
- `create_card_test.go` — с позицией, без позиции (пустая колонка / конец), non-member denied, пустой title, невалидный lexorank, ошибка GetLastInColumn
- `delete_board_test.go` — single/batch delete owner, not-owner denied, not-member denied, partial ownership batch denied, IsMember error, BatchDelete error
- `delete_card_test.go` — owner удаляет чужую, creator удаляет свою, non-member denied, member чужую denied, batch owner, card not found, empty IDs, IsMember error, BatchDelete error
- `add_column_test.go` — owner/member add, non-member denied, пустой title, отрицательная позиция, ошибка save
- `add_member_test.go` — owner добавляет member/owner, not-owner denied, невалидная роль, board not found, duplicate, ошибка add
- `get_board_test.go` — owner/member get, not found, non-member denied, IsMember error
- `list_boards_test.go` — дефолтный/кастомный лимит, лимит >100, отрицательный, ошибка repo
- `move_card_test.go` — перемещение в начало/конец/между/пустую колонку

### Feature тесты (65 тестов, реальный PostgreSQL)

**Integration repos** (`services/board/tests/integration/*_repository_test.go`):
- `board_repository_test.go` — Create, GetByID_NotFound, Update, OptimisticLocking, Delete, CursorPagination, ListByUserID_EmptyResult
- `card_repository_test.go` — Create, CreateWithoutAssignee, GetByID_NotFound, ListByColumnID, LexorankPositioning, Update, Move, Delete, Partitioning
- `column_repository_test.go` — Create, GetByID_NotFound, ListByBoardID, Update, Delete, CascadeDelete
- `membership_repository_test.go` — AddMember, AddMember_Duplicate, AddMember_InvalidRole, RemoveMember, RemoveMember_NotFound, RemoveMember_CannotRemoveOwner, IsMember, ListMembers, ListMembers_Pagination

**Feature сценарии** (`services/board/tests/integration/feature_test.go`):
- Board: CreateBoard_OwnerAutoMember, ListBoards_OnlyMemberBoards, ListBoards_MemberSeesSharedBoards, ListBoards_OwnerOnlyFilter, ListBoards_SearchByTitle, UpdateBoard_OnlyOwner, GetBoard_NonMemberDenied
- Column: AddColumn_MemberCanAdd, AddColumn_NonMemberDenied, DeleteColumn_OnlyOwnerCanDelete
- Card: CreateCard_SetsCreatorID, CreateCard_NonMemberDenied, MoveCard_MemberCanMove, MoveCard_NonMemberDenied, UpdateCard_MemberCanUpdate
- Member: AddMember_OnlyOwnerCanAdd, RemoveMember_OnlyOwnerCanRemove, RemoveMember_CannotRemoveOwner, AfterRemoval_NoAccess

**Delete сценарии** (`services/board/tests/integration/delete_test.go`):
- Board: OwnerCanDelete, MemberCannotDelete, BatchDelete, CascadeDeletesCards
- Card: CreatorCanDelete, OwnerCanDeleteAnyCard, MemberCannotDeleteOthersCard, BatchDelete, NonMemberCannotDelete

### Запуск

```bash
# Unit
cd services/board && go test ./internal/... -parallel 8

# Feature (нужен PostgreSQL)
TEST_DATABASE_URL="postgres://yammi:yammi@localhost:5432/board_test?sslmode=disable" \
  go test ./tests/integration/ -parallel 8 -timeout 180s
```
