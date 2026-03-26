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

- **API Gateway** — HTTP entry point, JWT verification (local, via public key), rate limiting, input validation (max lengths), 10s gRPC timeout interceptor, 1MB body limit. No business logic.
- **Auth Service** — Registration, login, JWT (EdDSA asymmetric keys), refresh/revoke tokens. Login protected against timing attacks (constant-time bcrypt). Auth events logged via slog.
- **User Service** — User profiles.
- **Board Service** — Core domain. Boards, columns, cards. Redis cache, NATS event publishing, optimistic locking.
- **Comment Service** — Card comments.
- **Notification Service** — Async event consumer only. No sync API.
- **WebSocket Gateway** (`services/gateway`) — Async event consumer, pushes real-time updates to clients. Never calls other services synchronously. JWT auth via Authorization header (preferred) or query param (fallback). CheckOrigin rejects when ALLOWED_ORIGINS not configured.

Each service has its own PostgreSQL database (see `scripts/init-databases.sql`). Cross-service data access is only via gRPC.

Infrastructure: PostgreSQL 16, Redis 7, NATS 2 (JetStream). All defined in `docker-compose.yml`.

## New Features

### Card Metadata
Cards now support `due_date` (ISO 8601 timestamp), `priority` (low/medium/high/critical), and `task_type` (bug/feature/task/improvement). All three are optional fields set during creation or update.

### Labels
Board-scoped color tags. Each label has a name and hex color. Max 50 labels per board. Labels can be attached/detached to/from cards (many-to-many). Stored in `labels` and `card_labels` tables.

### Checklists
Multiple checklists per card, each with ordered items. Items can be toggled done/undone. Progress tracking (done/total) is derived from items. Stored in `checklists` and `checklist_items` tables.

### Card Dependencies (Parent-Child Links)
Cards can be linked in parent-child relationships for subtask tracking. A card can have multiple parents and multiple children. Cycle detection prevents circular dependencies. Stored in `card_links` table.

### Custom Fields
Board-scoped custom field definitions with types: text, number, date, dropdown. Dropdown fields have configurable options. Max 30 custom fields per board. Values are stored per card. Stored in `custom_fields` and `card_custom_field_values` tables.

### Automation Rules
Board-scoped trigger-action rules. Trigger types: `card_moved_to_column`, `card_created`, `label_added`, `due_date_approaching`, `checklist_completed`. Action types: `set_priority`, `move_card`, `add_label`, `assign_member`, `send_notification`. Max 25 rules per board. Execution history is tracked. Stored in `automation_rules` and `automation_history` tables.

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

## Board Service Handler Decomposition

The gRPC handler in Board Service is decomposed into 10 domain-specific sub-handler structs to avoid a God Object. `BoardServiceServer` delegates to:

| Sub-handler | Domain | Methods |
|-------------|--------|---------|
| `BoardCoreHandler` | Boards CRUD | 5 |
| `ColumnHandler` | Columns | 5 |
| `CardHandler` | Cards + activity | 10 |
| `MemberHandler` | Membership | 4 |
| `AttachmentHandler` | File attachments | 5 |
| `LabelHandler` | Labels | 7 |
| `CardLinkHandler` | Parent-child links | 4 |
| `ChecklistHandler` | Checklists | 8 |
| `CustomFieldHandler` | Custom fields | 6 |
| `AutomationHandler` | Automation rules | 5 |

Each sub-handler has its own struct + constructor in the corresponding `*_handler.go` file. The main `NewBoardServiceServer()` takes 10 sub-handler params instead of 52 positional args. Methods remain on `*BoardServiceServer` (required by gRPC interface) but access deps through sub-handlers (e.g., `s.cards.create.Execute(...)`).

## DDD: Board as Aggregate Root

Board is the sole aggregate root in Board Service. Column and Card are value objects within Board — there are no separate `CardRepository` or `ColumnRepository`. All persistence goes through `BoardRepository`.

Invariants enforced in domain: card belongs to exactly one column, unique ordering within column, version increment on every change (optimistic locking).

## Security & Resilience

- **Async event publishing** — all `go func()` blocks use `context.WithTimeout(5s)` + `slog.Error` logging. No fire-and-forget.
- **gRPC timeouts** — API Gateway has 10s default timeout interceptor on all outgoing gRPC calls (`timeoutInterceptor` in `grpc_clients.go`).
- **Input validation** — API Gateway validates max lengths: title (500), description (5000), name (255), content (10000), search (200), color (7). Defined in `dto.go`.
- **SQL LIKE injection** — all ILIKE queries use `ESCAPE '\'` clause + `escapeLikePattern()` function.
- **Timing attack protection** — login always runs bcrypt even for non-existent users.
- **CORS** — requires explicit `ALLOWED_ORIGINS` env var; rejects all origins when not set.
- **WebSocket** — CheckOrigin rejects when origins not configured; JWT via Authorization header.
- **Request body limit** — 1MB via `MaxBodyMiddleware`.
- **Structured logging** — `log/slog` for auth events and event publishing errors.

## Conventions

- Each service is an independent Go module (`services/<name>/go.mod`) — Go 1.24 for auth, Go 1.23 for others.
- `pkg/` contains only shared contracts (event definitions, shared proto). No utilities or middleware — those are `internal` per service.
- Environment variables configure services: `DATABASE_URL`, `<SERVICE>_GRPC_PORT`, `REDIS_URL`, `NATS_URL`.
- Domain errors use typed sentinels: `ErrCardNotFound`, `ErrAccessDenied`, etc.
- Authorization (role checks: Owner/Member) happens in usecase layer, not delivery.
- Events carry `event_id` (UUID for idempotency), `event_version`, and `occurred_at`.

## Testing

**300+ тестов, 8 параллельных потоков, ~2s total.**

### Unit тесты (200+ тестов)

**Domain** (`services/board/internal/domain/*_test.go`) — **95.9% coverage**, ~0.2s:
- `card_test.go` — NewCard валидация, Update, Move, Reorder, lexorank validation, creator_id
- `board_test.go` — NewBoard, Update, IsOwner, IncrementVersion
- `column_test.go` — NewColumn, Update, UpdatePosition
- `member_test.go` — NewMember, CanModifyBoard, CanModifyCards, роли
- `lexorank_test.go` — LexorankBetween (инварианты: prev < result < next), ValidateLexorank, ordering, sequences, edge cases

**New domain tests:**
- `label_test.go` — NewLabel валидация, color format, board scoping, max 50 limit
- `checklist_test.go` — NewChecklist, AddItem, ToggleItem, progress tracking
- `card_link_test.go` — NewCardLink, cycle detection, parent-child invariants
- `custom_field_test.go` — NewCustomField, type validation, dropdown options, max 30 per board
- `automation_rule_test.go` — NewAutomationRule, trigger/action validation, enable/disable, max 25 per board

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
- New: label, checklist, card_link, custom_field, automation_rule usecase tests

### Feature тесты (100+ тестов, реальный PostgreSQL)

**Integration repos** (`services/board/tests/integration/*_repository_test.go`):
- `board_repository_test.go` — Create, GetByID_NotFound, Update, OptimisticLocking, Delete, CursorPagination, ListByUserID_EmptyResult
- `card_repository_test.go` — Create, CreateWithoutAssignee, GetByID_NotFound, ListByColumnID, LexorankPositioning, Update, Move, Delete, Partitioning
- `column_repository_test.go` — Create, GetByID_NotFound, ListByBoardID, Update, Delete, CascadeDelete
- `membership_repository_test.go` — AddMember, AddMember_Duplicate, AddMember_InvalidRole, RemoveMember, RemoveMember_NotFound, RemoveMember_CannotRemoveOwner, IsMember, ListMembers, ListMembers_Pagination
- `label_repository_test.go` — Create, ListByBoard, Update, Delete, AttachToCard, DetachFromCard, GetCardLabels
- `checklist_repository_test.go` — Create, ListByCard, Update, Delete, CreateItem, UpdateItem, DeleteItem, ToggleItem
- `card_link_repository_test.go` — Create, Delete, GetChildren, GetParents, CycleDetection
- `custom_field_repository_test.go` — Create, ListByBoard, Update, Delete, SetValue, GetCardValues
- `automation_rule_repository_test.go` — Create, ListByBoard, Update, Delete, GetHistory

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

## New API Endpoints

30 new routes added to API Gateway, grouped by feature:

### Labels (7 routes)
- `POST   /api/v1/boards/:boardId/labels` — create label
- `GET    /api/v1/boards/:boardId/labels` — list board labels
- `PUT    /api/v1/boards/:boardId/labels/:labelId` — update label
- `DELETE /api/v1/boards/:boardId/labels/:labelId` — delete label
- `POST   /api/v1/boards/:boardId/cards/:cardId/labels` — attach label to card
- `DELETE /api/v1/boards/:boardId/cards/:cardId/labels/:labelId` — detach label from card
- `GET    /api/v1/boards/:boardId/cards/:cardId/labels` — get card labels

### Checklists (8 routes)
- `POST   /api/v1/boards/:boardId/cards/:cardId/checklists` — create checklist
- `GET    /api/v1/boards/:boardId/cards/:cardId/checklists` — list card checklists
- `PUT    /api/v1/boards/:boardId/checklists/:checklistId` — update checklist
- `DELETE /api/v1/boards/:boardId/checklists/:checklistId` — delete checklist
- `POST   /api/v1/boards/:boardId/checklists/:checklistId/items` — create checklist item
- `PUT    /api/v1/boards/:boardId/checklist-items/:itemId` — update checklist item
- `DELETE /api/v1/boards/:boardId/checklist-items/:itemId` — delete checklist item
- `PUT    /api/v1/boards/:boardId/checklist-items/:itemId/toggle` — toggle item done/undone

### Card Links / Subtasks (4 routes)
- `POST   /api/v1/boards/:boardId/cards/:cardId/links` — create parent-child link
- `DELETE /api/v1/boards/:boardId/card-links/:linkId` — delete link
- `GET    /api/v1/boards/:boardId/cards/:cardId/children` — get child cards
- `GET    /api/v1/boards/:boardId/cards/:cardId/parents` — get parent cards

### Custom Fields (6 routes)
- `POST   /api/v1/boards/:boardId/custom-fields` — create field definition
- `GET    /api/v1/boards/:boardId/custom-fields` — list board fields
- `PUT    /api/v1/boards/:boardId/custom-fields/:fieldId` — update field definition
- `DELETE /api/v1/boards/:boardId/custom-fields/:fieldId` — delete field definition
- `PUT    /api/v1/boards/:boardId/cards/:cardId/custom-fields/:fieldId` — set card field value
- `GET    /api/v1/boards/:boardId/cards/:cardId/custom-fields` — get card field values

### Automation Rules (5 routes)
- `POST   /api/v1/boards/:boardId/automations` — create rule
- `GET    /api/v1/boards/:boardId/automations` — list board rules
- `PUT    /api/v1/boards/:boardId/automations/:ruleId` — update rule
- `DELETE /api/v1/boards/:boardId/automations/:ruleId` — delete rule
- `GET    /api/v1/boards/:boardId/automations/:ruleId/history` — get rule execution history

## Migrations

Board Service migrations: `services/board/migrations/` — files `000001_init.up.sql` through `000012_*.up.sql`.

Key migrations for new features:
- `000007` — Labels table and card_labels junction table
- `000008` — Checklists and checklist_items tables
- `000009` — Card links (parent-child) table with cycle detection constraints
- `000010` — Custom fields definitions and card_custom_field_values tables
- `000011` — Automation rules and automation_history tables
- `000012` — Card metadata columns (due_date, priority, task_type)

## Access Control

Feature-specific authorization rules (enforced in usecase layer):

- **Labels**: member can create/update/list; owner-only for delete
- **Checklists**: member for all operations (create, update, delete, toggle items)
- **Card Links**: member for link/unlink/view
- **Custom Fields**: owner-only for field definitions (create/update/delete); member for setting values and viewing
- **Automation Rules**: owner-only for CRUD (create/update/delete); member for list and history viewing
