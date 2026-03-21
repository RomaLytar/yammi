# Yammi — Trello-like Task Board (Microservices, Go)

Highload Trello-клон на Go с микросервисной архитектурой, lightweight DDD и clean architecture.

## Архитектура

```
[Client]
    |
[API Gateway]  ← routing / auth (JWT verify) / rate limiting
    |
----------------------------------------------
| Auth | User | Board | Comment |
----------------------------------------------
                 |
          [Event Bus (NATS/Kafka)]
                 |
         ┌───────┴────────┐
   [Notification]   [Realtime Gateway]
                          |
                    [WebSocket → Client]
```

**API Gateway** — единая точка входа. Роутинг, проверка JWT (локально, по публичному ключу), rate limiting.

**Realtime Gateway** — НЕ ходит в сервисы синхронно. Только слушает события из Event Bus и пушит клиентам через WebSocket.

### Коммуникация между сервисами

- **Синхронно** — gRPC (между сервисами, через API Gateway)
- **Асинхронно** — NATS / Kafka (event-driven, для Notification и Realtime)

### Data Ownership

Каждый сервис владеет своей моделью данных. Прямой доступ к БД другого сервиса **запрещён**. Данные получаются только через gRPC API сервиса-владельца.

| Сервис       | БД         | Кеш   |
|-------------|------------|-------|
| Auth        | PostgreSQL | —     |
| User        | PostgreSQL | —     |
| Board       | PostgreSQL | Redis |
| Comment     | PostgreSQL | —     |
| Notification| PostgreSQL | —     |

### Consistency

- **Eventual consistency** между сервисами
- Синхронизация через события в Event Bus
- Внутри одного сервиса — strong consistency (транзакции PostgreSQL)

## Микросервисы

### API Gateway (`services/api-gateway`)
- Единая точка входа для клиента
- Routing запросов к нужному сервису
- JWT валидация (локально, по публичному ключу Auth Service)
- Rate limiting
- НЕ содержит бизнес-логики

### Auth Service (`services/auth`)
- Регистрация / логин
- Выпуск JWT токенов (подписывает приватным ключом)
- Refresh / revoke токенов
- Раздача публичного ключа другим сервисам (для локальной валидации)

> **Важно:** Другие сервисы НЕ ходят в Auth на каждый запрос. JWT валидируется локально по публичному ключу. Auth дёргается только для refresh/revoke.

### User Service (`services/user`)
- Профиль пользователя
- Данные (имя, аватар)
- Поиск пользователя по email (для добавления участников в доску)

### Board Service (`services/board`) — ядро системы
- Доски, колонки, карточки
- Управление участниками доски (Owner/Member роли)
- Фильтрация (только свои / все), поиск по названию (ILIKE + pg_trgm), сортировка (по активности / дате / алфавиту)
- Перемещение карточек (optimistic locking)
- Batch delete: доски (`POST /boards/delete`) и карточки (`POST /cards/delete`) — одна транзакция
- `creator_id` на карточках — отслеживание автора, права на удаление
- Автообновление `updated_at` доски при изменении карточек/колонок
- Кеширование через Redis
- Публикация событий в очередь

### Comment Service (`services/comment`)
- Комментарии к карточкам

### Notification Service (`services/notification`)
- Consumer: подписка на события из Event Bus
- Email / push уведомления
- НЕ принимает синхронные запросы от других сервисов

### Realtime Gateway (`services/gateway`)
- WebSocket соединения с клиентами
- Consumer: подписка на события из Event Bus
- Пуш обновлений в реальном времени
- **НЕ ходит в сервисы синхронно** — только слушает события

## Структура проекта

```
yammi/
├── services/
│   ├── api-gateway/                # API Gateway
│   │   ├── cmd/server/             # точка входа
│   │   ├── internal/
│   │   │   ├── delivery/grpc/      # proxy к сервисам
│   │   │   └── infrastructure/     # rate limiter, JWT verify
│   │   ├── configs/
│   │   └── tests/
│   │
│   ├── auth/                       # Auth Service
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── domain/             # User credentials, Token
│   │   │   ├── usecase/            # Login, Register, RefreshToken
│   │   │   ├── delivery/grpc/
│   │   │   ├── repository/postgres/
│   │   │   └── infrastructure/
│   │   ├── api/proto/v1/           # versioned proto
│   │   ├── migrations/
│   │   ├── configs/
│   │   └── tests/
│   │
│   ├── user/                       # User Service
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   ├── usecase/
│   │   │   ├── delivery/grpc/
│   │   │   ├── repository/postgres/
│   │   │   └── infrastructure/
│   │   ├── api/proto/v1/
│   │   ├── migrations/
│   │   ├── configs/
│   │   └── tests/
│   │
│   ├── board/                      # Board Service (ядро)
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── domain/             # Board (aggregate root), Column, Card
│   │   │   ├── usecase/            # MoveCard, CreateBoard, AddColumn
│   │   │   ├── delivery/grpc/
│   │   │   ├── repository/postgres/
│   │   │   └── infrastructure/
│   │   │       ├── cache/          # Redis
│   │   │       └── queue/          # NATS/Kafka publisher
│   │   ├── api/proto/v1/
│   │   ├── migrations/
│   │   ├── configs/
│   │   └── tests/
│   │
│   ├── comment/                    # Comment Service
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   ├── usecase/
│   │   │   ├── delivery/grpc/
│   │   │   ├── repository/postgres/
│   │   │   └── infrastructure/
│   │   ├── api/proto/v1/
│   │   ├── migrations/
│   │   ├── configs/
│   │   └── tests/
│   │
│   ├── notification/               # Notification Service
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   ├── usecase/
│   │   │   └── infrastructure/
│   │   │       └── queue/          # NATS/Kafka consumer
│   │   ├── configs/
│   │   └── tests/
│   │
│   └── gateway/                    # Realtime Gateway
│       ├── cmd/server/
│       ├── internal/
│       │   ├── delivery/
│       │   │   └── websocket/      # WebSocket handlers
│       │   └── infrastructure/
│       │       └── queue/          # NATS/Kafka consumer
│       ├── configs/
│       └── tests/
│
├── pkg/                            # Shared contracts (минимум!)
│   ├── events/                     # Определения событий
│   └── proto/v1/                   # Shared proto definitions
│
├── deployments/                    # docker-compose, k8s manifests
├── scripts/                        # скрипты для разработки
└── README.md
```

> **pkg/ — минимум кода.** Только shared contracts: события и proto. Всё остальное (middleware, logger) — internal в каждом сервисе. Иначе pkg превращается в помойку и ломает границы сервисов.

## Clean Architecture (в каждом сервисе)

```
delivery (gRPC/WS) → usecase → domain
                       ↓
                    repository (interface)
                       ↓
                    infrastructure (postgres, redis, nats)
```

- **domain** — сущности и бизнес-правила, 0 зависимостей
- **usecase** — сценарии, оркестрирует domain и repository
- **repository** — интерфейсы в usecase, реализация в infrastructure
- **delivery** — gRPC/WebSocket handlers, вызывают usecase
- **infrastructure** — БД, кеш, очереди

## DDD подход (lightweight)

### Board — Aggregate Root

**Board** — единственный aggregate root в Board Service. Column и Card — value objects внутри Board. Нет отдельных `CardRepository`, `ColumnRepository` — всё через `BoardRepository`.

### Инварианты

- Карточка принадлежит **только одной** колонке
- Порядок карточек **уникален** в пределах колонки
- Нельзя переместить карточку в несуществующую колонку
- Нельзя добавить карточку в чужую доску (проверка membership)
- Version increment при каждом изменении (optimistic locking)

### Пример usecase

```go
func (uc *MoveCardUseCase) Execute(cmd MoveCardCommand) error {
    board, err := uc.repo.GetByID(cmd.BoardID)
    if err != nil {
        return err
    }

    // domain содержит бизнес-логику, usecase — оркестрацию
    err = board.MoveCard(cmd.CardID, cmd.TargetColumnID, cmd.Position)
    if err != nil {
        return err
    }

    if err := uc.repo.Save(board); err != nil {
        return err // optimistic lock failure → retry на уровне delivery
    }

    // публикуем событие асинхронно
    uc.publisher.Publish(events.CardMoved{
        EventID:      uuid.New().String(),
        OccurredAt:   time.Now(),
        BoardID:      cmd.BoardID,
        CardID:       cmd.CardID,
        FromColumnID: cmd.FromColumnID,
        ToColumnID:   cmd.TargetColumnID,
        Position:     cmd.Position,
    })

    return nil
}
```

## Authorization (не путать с Auth)

Auth Service отвечает за **аутентификацию** (кто ты?).
**Авторизация** (что тебе можно?) живёт в Board Service:

### Роли

| Роль     | Права                                    |
|---------|------------------------------------------|
| Owner   | всё: удаление доски/колонок, удаление любых карточек, управление members  |
| Member  | создание карточек, перемещение, удаление только своих карточек (по creator_id), комментарии  |

### Проверка прав

```go
// внутри usecase, перед выполнением действия
func (uc *MoveCardUseCase) Execute(userID string, cmd MoveCardCommand) error {
    board, _ := uc.repo.GetByID(cmd.BoardID)

    if !board.IsMember(userID) {
        return domain.ErrAccessDenied
    }
    // ...
}
```

## Events (структура событий)

Все события формализованы, с idempotency и versioning:

```go
// pkg/events/board_events.go

type CardMoved struct {
    EventID      string    `json:"event_id"`      // UUID, для idempotency
    EventVersion int       `json:"event_version"`  // версия схемы события
    OccurredAt   time.Time `json:"occurred_at"`
    BoardID      string    `json:"board_id"`
    CardID       string    `json:"card_id"`
    FromColumnID string    `json:"from_column_id"`
    ToColumnID   string    `json:"to_column_id"`
    Position     int       `json:"position"`
}

type CardCreated struct {
    EventID      string    `json:"event_id"`
    EventVersion int       `json:"event_version"`
    OccurredAt   time.Time `json:"occurred_at"`
    BoardID      string    `json:"board_id"`
    ColumnID     string    `json:"column_id"`
    CardID       string    `json:"card_id"`
    Title        string    `json:"title"`
}

type BoardUpdated struct {
    EventID      string    `json:"event_id"`
    EventVersion int       `json:"event_version"`
    OccurredAt   time.Time `json:"occurred_at"`
    BoardID      string    `json:"board_id"`
}
```

### Idempotency

Consumer хранит `event_id` обработанных событий в **Redis (SET с TTL)** или **отдельной таблице `processed_events`** в PostgreSQL. Повторное получение того же события — skip.

```
Redis:  SETNX processed_event:{event_id} 1 EX 86400   (TTL 24h)
PG:     INSERT INTO processed_events (event_id, processed_at)
        ON CONFLICT DO NOTHING
```

Выбор хранилища:
- **Redis** — быстрее, подходит если потеря при рестарте допустима (TTL)
- **PostgreSQL** — надёжнее, в одной транзакции с обработкой события (exactly-once semantics)

### Retry Policy и Dead-Letter Queue

```
Retry:
  max_attempts: 3
  backoff:      exponential (1s → 2s → 4s)

Dead-Letter Queue (DLQ):
  после max_attempts → событие уходит в DLQ
  DLQ мониторится через алерты
  ручной replay из DLQ после фикса
```

- Каждый consumer реализует retry с exponential backoff
- Если событие не обработано после N попыток — отправляется в Dead-Letter Queue
- DLQ события логируются и мониторятся (алерт в Grafana)
- После исправления бага — ручной replay событий из DLQ обратно в основную очередь

### Event Versioning

Поле `event_version` позволяет менять схему событий без breaking changes. Consumer проверяет версию и обрабатывает соответственно.

## Highload фишки

### Optimistic Locking

Поле `version` в Board. При `Save()` проверяется `WHERE version = $expected`. Конфликт → retry на уровне delivery.

### Redis Cache (Board Service)

```
Strategy: read-through cache
Key:      board:{id}
TTL:      5 min
Invalidation: на событие BoardUpdated (через Event Bus)
```

- GET board → проверяем Redis → miss → PostgreSQL → кладём в Redis
- Board изменён → publish BoardUpdated → consumer инвалидирует кеш

### Event-driven

Изменение → событие в очередь → Notification Service + Realtime Gateway

### Rate Limiting

Реализуется в API Gateway (token bucket / sliding window).

### Горутины и каналы

Конкурентная обработка событий в consumer'ах (fan-out pattern).

### Monitoring / Metrics

```
Stack: Prometheus + Grafana

Метрики (каждый сервис экспортирует через /metrics):

  Latency:
    - grpc_request_duration_seconds (histogram, по method)
    - db_query_duration_seconds
    - cache_hit / cache_miss ratio

  Throughput:
    - grpc_requests_total (counter, по method + status)
    - events_published_total
    - events_consumed_total

  Errors:
    - grpc_errors_total (по code)
    - events_failed_total
    - dlq_events_total (Dead-Letter Queue)

Алерты:
    - p99 latency > 500ms
    - error rate > 1%
    - DLQ не пустой
    - consumer lag растёт
```

Каждый сервис экспортирует метрики через Prometheus endpoint. Grafana dashboards для визуализации. Алерты через Alertmanager.

## Тестирование

| Тип             | Что тестируем                              | Инструменты          |
|----------------|-------------------------------------------|---------------------|
| Unit           | usecase, domain логика                    | go test, testify    |
| Feature        | бизнес-сценарии, права доступа, batch ops | go test, реальная БД |
| Integration    | repository + реальная БД                  | testcontainers-go   |
| Contract       | gRPC контракты между сервисами            | grpc-testing        |
| Load           | точки нагрузки (GetBoard, MoveCard)       | k6                  |

### Покрытие (Board Service)

| Слой | Coverage | Тестов | Время |
|------|----------|--------|-------|
| domain | **95.9%** | 111 | ~0.2s |
| usecase (unit) | **46.9%** | 47 | ~0.01s |
| internal/* (feature + integration) | **63.8%** | 65 | ~1.8s |

**Всего: 223 теста, параллельно на 8 потоках**

### Unit тесты (`internal/usecase/*_test.go`, `internal/domain/*_test.go`)
- Domain: lexorank, карточки (CRUD + валидация), колонки, доски, members
- Usecase: create/delete board, create/delete card (batch, permissions), add column, add member, list boards, move card, get board

### Feature тесты (`tests/integration/feature_test.go`, `delete_test.go`)
Реальный PostgreSQL, все миграции, изолированные данные (UUID), параллельный запуск:

**Доски:**
- Создание → owner автоматически member
- Список → видно только свои и shared доски
- Фильтр owner_only, поиск по названию (ILIKE)
- Обновление → non-member denied
- Batch delete → owner only

**Колонки:**
- Member может создавать
- Non-member denied
- Удаление → owner only

**Карточки:**
- Создание → creator_id сохраняется
- Non-member denied
- Перемещение → member может, non-member нет
- Удаление: creator может свои, owner любые, member чужие — denied
- Batch delete
- Каскадное удаление при удалении доски

**Участники:**
- Добавление → только owner
- Удаление → только owner, нельзя удалить owner'а
- После удаления → доступ к доске закрыт

### Запуск тестов

```bash
# Unit тесты
cd services/board
go test ./internal/domain/ ./internal/usecase/ -parallel 8

# Feature + Integration тесты (нужен PostgreSQL)
TEST_DATABASE_URL="postgres://yammi:yammi@localhost:5432/board_test?sslmode=disable" \
  go test ./tests/integration/ -parallel 8 -timeout 180s
```

### Приоритет
1. Unit — domain и usecase (самое важное)
2. Feature — бизнес-сценарии с реальной БД
3. Integration — repository с PostgreSQL
4. Contract — гарантия совместимости между сервисами
5. Load — k6 сценарии для точек нагрузки

## Стек

| Компонент     | Технология        |
|--------------|------------------|
| Язык         | Go               |
| API          | gRPC + WebSocket |
| Proto        | protobuf v1      |
| БД           | PostgreSQL       |
| Кеш          | Redis            |
| Очередь      | NATS / Kafka     |
| Monitoring   | Prometheus + Grafana |
| Alerting     | Alertmanager     |
| Контейнеры   | Docker Compose   |
| CI           | GitHub Actions   |
| Load testing | k6               |

## Локальная инфраструктура

После запуска `docker compose up --build` доступны:

### Приложение

| Компонент          | URL                                      | Описание                          |
|--------------------|------------------------------------------|-----------------------------------|
| API Gateway        | http://localhost:8080                     | Основная точка входа (REST API)   |
| API Health Check   | http://localhost:8080/health              | Проверка статуса API Gateway      |
| WebSocket Gateway  | ws://localhost:8081/ws                    | Real-time обновления (WIP)        |
| WS Health Check    | http://localhost:8081/health              | Проверка статуса WS Gateway       |

### Мониторинг

| Компонент   | URL                          | Логин / Пароль    | Описание                              |
|-------------|------------------------------|-------------------|---------------------------------------|
| Grafana     | http://localhost:3033        | `admin` / `admin` | Дашборды: NATS JetStream, User Events |
| Prometheus  | http://localhost:9090        | —                 | Метрики, запросы PromQL               |
| NATS Monitor| http://localhost:8222        | —                 | Встроенный мониторинг NATS            |

### Базы данных и инфраструктура

| Компонент   | Адрес                        | Креды                | Описание                     |
|-------------|------------------------------|----------------------|------------------------------|
| PostgreSQL  | `localhost:5432`             | `yammi` / `yammi`    | 5 БД: yammi_auth, yammi_user, yammi_board, yammi_comment, yammi_notification |
| Redis       | `localhost:6380`             | —                    | Кеш (Board Service)          |
| NATS        | `localhost:4222`             | —                    | Event Bus (JetStream)        |

### Grafana дашборды

После первого запуска дашборды подгружаются автоматически через provisioning:

- **NATS JetStream** — состояние стримов, consumer lag, сообщения в секунду
- **NATS User Deleted** — мониторинг обработки событий UserDeleted

### Полезные ссылки Prometheus

- Все метрики: http://localhost:9090/targets
- NATS метрики: http://localhost:9090/graph?g0.expr=nats_varz_connections

### DLQ (Dead-Letter Queue) утилита

```bash
# Просмотр сообщений в DLQ
docker compose run --rm dlq list

# Переотправка событий из DLQ
docker compose run --rm dlq replay

# Очистка DLQ
docker compose run --rm dlq purge
```

### Postman

Коллекция API Gateway находится в `postman/Yammi_API_Gateway.postman_collection.json`. Импортируйте в Postman для тестирования всех эндпоинтов.

## Статус реализации

| Сервис              | Статус          | Описание                                           |
|---------------------|-----------------|-----------------------------------------------------|
| Auth Service        | ✅ Реализован    | JWT (EdDSA), регистрация, логин, refresh/revoke      |
| User Service        | ✅ Реализован    | Профили, поиск по email, NATS consumer, DLQ monitor  |
| API Gateway         | ✅ Реализован    | REST API, JWT verify, rate limiting                  |
| Board Service       | ✅ Реализован    | Domain, Repository, Use Cases, gRPC, API Gateway, Members, creator_id, batch delete, 39+ unit тестов, Grafana dashboard |
| Comment Service     | ⬜ Заглушка      | Только main.go, нет бизнес-логики                    |
| Notification Service| ⬜ Заглушка      | Только main.go, нет бизнес-логики                    |
| WebSocket Gateway   | ⬜ Заглушка      | /health работает, /ws возвращает 501                 |

## План разработки

1. ~~**Domain + Usecase** — Board aggregate, инварианты, бизнес-логика~~
2. ~~**Clean Architecture** — DI, слои, repository interfaces~~
3. ~~**gRPC** — proto/v1 файлы, handlers, базовые методы~~
4. ~~**Auth Service** — регистрация, логин, JWT (asymmetric keys)~~
5. ~~**API Gateway** — routing, JWT verify, rate limiting~~
6. ~~**PostgreSQL** — миграции, repository реализация~~
7. ~~**Authorization** — roles (owner/member), проверка в usecase, управление участниками~~
8. ~~**Event Bus** — NATS, формализованные события~~
9. **Realtime Gateway** — WebSocket, consumer событий
10. **Redis** — read-through cache для досок
11. ~~**Тесты** — unit, integration, load (k6)~~
12. ~~**Docker + CI** — docker-compose, GitHub Actions, k6~~
