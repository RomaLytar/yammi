# Архитектура Yammi

## Общая концепция

Yammi — микросервисная система для управления досками задач (Trello-like). Построена на принципах:
- **Микросервисы** — 7 независимых сервисов, каждый со своей областью ответственности
- **Event-Driven Architecture** — асинхронная коммуникация через NATS
- **Clean Architecture** — separation of concerns (domain, usecase, delivery, infrastructure)
- **Domain-Driven Design** — с адаптацией под микросервисы (micro-aggregates)

---

## Схема взаимодействия сервисов

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client (Browser)                         │
└────────────┬────────────────────────────┬───────────────────────┘
             │ HTTP                       │ WebSocket
             ▼                            ▼
┌─────────────────────┐         ┌──────────────────────┐
│   API Gateway       │         │  WebSocket Gateway   │
│   :8080             │         │  :8081               │
└──────────┬──────────┘         └──────────┬───────────┘
           │ gRPC                          │ NATS (sub)
           │                               │
    ┌──────┴──────┬────────────────┬───────┴──────┐
    │             │                │              │
    ▼             ▼                ▼              ▼
┌────────┐  ┌─────────┐     ┌──────────┐   ┌──────────────┐
│ Auth   │  │  User   │     │  Board   │   │ Notification │
│ :50051 │  │ :50052  │     │ :50053   │   │ :50055       │
│ (x5)   │  │         │     │          │   │              │
└────┬───┘  └────┬────┘     └────┬─────┘   └──────┬───────┘
     │           │               │                 │
     └───────────┴───────────────┴─────────────────┘
                        │
                        ▼
               ┌─────────────────┐
               │   NATS (Event   │
               │   Bus/JetStream)│
               └─────────────────┘
                        │
         ┌──────────────┼──────────────┐
         │              │              │
         ▼              ▼              ▼
  ┌──────────┐  ┌──────────┐  ┌──────────┐
  │PostgreSQL│  │  Redis   │  │ Grafana  │
  │(5 баз)   │  │  Cache   │  │+ Prom.   │
  └──────────┘  └──────────┘  └──────────┘
```

### Синхронная коммуникация (gRPC)

**Client → API Gateway → Service:**
- Register/Login → Auth Service
- GetProfile/UpdateProfile → User Service
- Board/Card CRUD → Board Service

**Характеристики:**
- Request-Response (клиент ждет ответ)
- Строгий контракт (protobuf)
- Load balancing (round-robin для Auth Service x5)
- Timeout (5 секунд на запрос)

### Асинхронная коммуникация (NATS)

**Service → NATS → Service (event-driven):**
- Auth publishes `user.created` → User Service создает профиль
- Auth publishes `user.deleted` → User Service удаляет профиль
- Board publishes `card.updated` → WebSocket Gateway пушит обновление клиентам
- Board publishes `card.updated` → Notification Service создает уведомление

**Характеристики:**
- Fire-and-forget (publisher не ждет обработки)
- At-least-once delivery (JetStream гарантирует доставку)
- Retry с exponential backoff
- Dead Letter Queue при failures

---

## Сервисы (детально)

### 1. API Gateway (:8080)

**Назначение:** Единая HTTP-точка входа для клиентов.

**Ответственность:**
- HTTP → gRPC маппинг
- JWT verification (локальная, без запросов в Auth)
- Rate limiting (per IP, token bucket)
- CORS
- Request/Response logging

**НЕ делает:**
- Бизнес-логику
- Прямые запросы в БД
- Трансформацию данных (кроме proto ↔ JSON)

**Технологии:**
- `net/http` (стандартная библиотека Go)
- gRPC clients (Auth, User, Board)
- In-memory rate limiter (без Redis)

**Особенности:**
- Stateless (можно горизонтально масштабировать)
- JWT public key cache (обновляется при fail + cooldown 30s)
- Middleware chain: Logger → RateLimiter → JWTVerifier → OwnerOnly

**Роуты:**
```
POST   /api/v1/auth/register      (public, rate limit 50/min)
POST   /api/v1/auth/login          (public, rate limit 50/min)
GET    /api/v1/auth/public-key     (public, no rate limit)
POST   /api/v1/auth/refresh        (auth required)
POST   /api/v1/auth/revoke         (auth required)
GET    /api/v1/users/{id}          (auth required)
PUT    /api/v1/users/{id}          (auth required + owner only)
DELETE /api/v1/users/{id}          (auth required + owner only)
GET    /health                     (public)
```

### 2. Auth Service (:50051, 5 реплик)

**Назначение:** Аутентификация и управление токенами.

**Ответственность:**
- Регистрация (bcrypt hashing, UUID generation)
- Логин (bcrypt compare, JWT generation)
- Refresh token rotation
- Revoke tokens
- DeleteUser (каскадное удаление)
- Публикация событий (`user.created`, `user.deleted`)

**База данных:** `yammi_auth`
```sql
users (id, email UNIQUE, name, password_hash, created_at, updated_at)
refresh_tokens (id, user_id FK, token UNIQUE, expires_at, revoked, created_at)
```

**JWT:**
- Алгоритм: EdDSA (Ed25519) — быстрее RSA, короче ключи
- Access token TTL: 15 минут
- Refresh token TTL: 7 дней
- Seed-based key generation (детерминированный из `JWT_SEED` env var)
- Все 5 реплик используют одинаковый seed → одинаковые ключи

**Bcrypt Pool:**
- Ограничение параллелизма (semaphore на `runtime.NumCPU()`)
- Защита от CPU exhaustion при burst регистрациях

**NATS Publishing:**
- Non-blocking (если NATS недоступен — warning в лог, регистрация успешна)
- Event ID (UUID для idempotency)
- Event version (для schema evolution)

**Почему 5 реплик:**
- Auth — самая нагруженная точка (каждый защищенный запрос → JWT verify)
- Bcrypt hashing — CPU-intensive
- Load balancing через DNS round-robin (`dns:///auth:50051`)

### 3. User Service (:50052)

**Назначение:** Управление профилями пользователей.

**Ответственность:**
- Хранение профилей (name, avatar_url, bio)
- GetProfile, UpdateProfile
- Автоматическое создание/удаление профилей (через NATS events)

**База данных:** `yammi_user`
```sql
profiles (id UUID, email UNIQUE, name, avatar_url, bio, created_at, updated_at)
```

**NATS Consumers:**
- `user-service-user-created-v4` — подписка на `user.created`
- `user-service-user-deleted-v1` — подписка на `user.deleted`

**Consumer config:**
```go
MaxDeliver: 7         // макс 7 попыток
AckWait: 30s          // если не Ack за 30s → retry
MaxAckPending: 500    // не более 500 необработанных сообщений
```

**Retry logic:**
- Exponential backoff: 2s → 4s → 8s → 16s → 30s (cap)
- Jitter: ±20% (избегаем thundering herd)
- После 7 неудач → событие отправляется в DLQ

**Idempotency:**
- Если профиль с таким email уже существует → Ack без ошибки
- Event ID не сохраняется (упрощение, для production нужна таблица processed_events)

**Ключевая особенность:**
- НЕТ синхронного CreateProfile эндпоинта
- Профиль создается только через event
- Eventual consistency (может быть задержка 100-500ms между регистрацией и созданием профиля)

### 4. Board Service (:50053)

**Назначение:** Core domain. Управление досками, колонками, карточками.

**Ответственность:**
- CRUD boards
- CRUD columns
- CRUD cards
- Board sharing (добавление/удаление members)
- Authorization (owner vs member)
- Optimistic locking (для boards)
- Lexorank positioning (для cards)
- Event publishing (`board.*`, `card.*`, `column.*`)

**База данных:** `yammi_board`
```sql
boards (id, title, owner_id, version, ...)
board_members (board_id, user_id, role)
columns (id, board_id, title, position INT)
cards (id, column_id, board_id, position VARCHAR, ...) PARTITIONED!
```

**Партиционирование:**
- Cards — HASH partitioned по board_id (4 партиции)
- Запросы к одной доске → только 1 партиция (performance)

**Redis Cache:**
- Read-through cache с TTL 5 минут
- Cached keys: `board:{id}`, `columns:{board_id}`, `cards:{column_id}`
- Invalidation при UPDATE/DELETE

**Архитектурное решение: Micro-Aggregates**

**Традиционный DDD:**
```go
type Board struct {
    ID       string
    Title    string
    Columns  []Column  // вложенные
    Members  []Member  // вложенные
}
```

**Yammi (micro-aggregates):**
```go
type Board struct {
    ID          string
    Title       string
    // БЕЗ Columns и Members!
}

type Column struct {
    ID      string
    BoardID string  // ссылка
}

type Card struct {
    ID       string
    ColumnID string  // ссылка
}
```

**Почему:**
- `GetBoard` не загружает 500 карточек в память (performance)
- Granular caching (инвалидация только измененной части)
- Concurrency (5 юзеров редактируют разные карточки без locks)
- Отдельные repositories (BoardRepo, ColumnRepo, CardRepo)

**Trade-off:**
- Слабее транзакционные гарантии (нельзя в одной транзакции изменить board + 10 cards)
- Инварианты проверяются в usecase, не в domain

**Authorization:**

| Role | Права |
|------|-------|
| Owner | Все (CRUD board, manage members, delete board) |
| Member | CRUD cards, read board (нельзя edit metadata, add members) |

**Проверка в usecase:**
```go
member, _ := membershipRepo.GetMembership(ctx, boardID, userID)
if !member.CanModifyBoard() {
    return ErrAccessDenied
}
```

**Lexorank:**
- Cards используют lexorank position (string вместо INT)
- O(1) reordering (UPDATE только одной карточки)
- Columns используют INT position (reorder редкий, колонок мало)

**Подробнее:**
- Database schema: `/home/roman/PetProject/yammi/guides/database-schema.md`
- Lexorank: `/home/roman/PetProject/yammi/guides/lexorank-explained.md`

### 5. Comment Service (:50054) — WIP

**Назначение:** Комментарии к карточкам.

**Планируемая структура:**
```sql
comments (id, card_id, user_id, content, created_at, updated_at)
```

**Особенности:**
- Полиморфная привязка (card_id может ссылаться на Board Service cards)
- Pagination (LIMIT/OFFSET или cursor)
- Нет nested comments (только flat список)

### 6. Notification Service (:50055) — WIP

**Назначение:** Асинхронный consumer. Создание уведомлений из событий.

**Ключевая особенность:**
- НЕТ синхронного gRPC API (только NATS consumer)
- Не вызывает другие сервисы синхронно
- Fire-and-forget (publish в NATS или email)

**Планируемые события:**
- `card.assigned` → "Вам назначена карточка X"
- `board.member_added` → "Вас добавили в доску Y"
- `comment.created` → "@user упомянул вас в комментарии"

### 7. WebSocket Gateway (:8081) — WIP

**Назначение:** Real-time push обновлений клиентам.

**Архитектура:**
- WebSocket connections (1000+ concurrent)
- NATS consumer (подписка на `board.*`, `card.*`)
- In-memory map: `boardID → []WebSocket`
- При получении события → push всем подключенным клиентам

**НЕ делает:**
- Синхронные вызовы других сервисов
- Бизнес-логику
- Авторизацию (клиент должен сначала авторизоваться в API Gateway)

**Flow:**
```
User A: drag card → API Gateway → Board Service → NATS event "card.moved"
                                                          ↓
WebSocket Gateway (consumer) получает событие
                                                          ↓
Push всем WebSocket клиентам доски X
                                                          ↓
User B (подключен к WebSocket) → получает обновление в реальном времени
```

---

## Clean Architecture (детально)

Каждый сервис следует одной структуре (пример: Auth Service):

```
services/auth/
├── cmd/server/main.go           # DI wiring
├── api/proto/v1/auth.proto      # gRPC контракт
├── internal/
│   ├── domain/                  # Бизнес-сущности + валидация
│   │   ├── user.go              # type User + ValidateRegistration
│   │   ├── token.go             # type RefreshToken + IsValid/Revoke
│   │   └── errors.go            # Typed errors (ErrEmailExists, ...)
│   ├── usecase/                 # Оркестрация domain + repo
│   │   ├── interfaces.go        # Интерфейсы (UserRepository, TokenGenerator)
│   │   ├── register.go          # func Register(ctx, email, password, name)
│   │   ├── login.go             # func Login(ctx, email, password)
│   │   └── token.go             # func RefreshToken, RevokeToken
│   ├── delivery/grpc/           # gRPC handlers
│   │   └── handler.go           # AuthHandler (proto → usecase → proto)
│   ├── repository/postgres/     # Реализация интерфейсов
│   │   ├── user_repo.go         # type UserRepo implements UserRepository
│   │   └── refresh_token_repo.go
│   └── infrastructure/          # External dependencies
│       ├── database.go          # PostgreSQL connection
│       ├── migrator.go          # SQL migrations runner
│       ├── jwt.go               # JWT generation (EdDSA)
│       ├── nats.go              # NATS publisher
│       └── hasher.go            # Bcrypt pool
└── migrations/
    └── 000001_init.up.sql       # SQL schema
```

### Dependency Rule

```
Внешний мир (HTTP/gRPC/NATS)
    ↓
Delivery Layer (handlers)
    ↓
Usecase Layer (бизнес-сценарии)
    ↓
Domain Layer (entities + validation)
    ↑
Repository Interfaces (определены в usecase)
    ↑
Repository Implementation (в infrastructure)
```

**Правила:**
1. **Domain** не импортирует ничего кроме stdlib и UUID
2. **Usecase** определяет интерфейсы, не знает про PostgreSQL/NATS
3. **Infrastructure** реализует интерфейсы конкретными технологиями
4. **Delivery** знает только про usecase, не про repository

**Пример (Register usecase):**

```go
// domain/user.go (validation)
func ValidateRegistration(email, password, name string) error {
    if !strings.Contains(email, "@") {
        return ErrInvalidEmail
    }
    if len(password) < 8 {
        return ErrWeakPassword
    }
    // ...
}

// usecase/interfaces.go (абстракция)
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type EventPublisher interface {
    PublishUserCreated(ctx context.Context, event events.UserCreated) error
}

// usecase/register.go (оркестрация)
func (uc *AuthUseCase) Register(ctx, email, password, name string) (*domain.User, string, string, error) {
    // 1. Валидация (domain)
    if err := domain.ValidateRegistration(email, password, name); err != nil {
        return nil, "", "", err
    }

    // 2. Проверка уникальности (repo)
    if _, err := uc.userRepo.GetByEmail(ctx, email); err == nil {
        return nil, "", "", domain.ErrEmailExists
    }

    // 3. Хеширование (infrastructure abstraction)
    passwordHash, _ := uc.hasher.Hash(password)

    // 4. Создание user (domain)
    user := &domain.User{
        ID:           uuid.NewString(),
        Email:        email,
        Name:         name,
        PasswordHash: passwordHash,
        CreatedAt:    time.Now(),
    }

    // 5. Сохранение (repo)
    uc.userRepo.Create(ctx, user)

    // 6. Генерация токенов (infrastructure abstraction)
    accessToken, _ := uc.tokenGenerator.Generate(user.ID)
    refreshToken := &domain.RefreshToken{...}
    uc.refreshTokenRepo.Create(ctx, refreshToken)

    // 7. Публикация события (async, best-effort)
    uc.eventPublisher.PublishUserCreated(ctx, events.UserCreated{
        EventID:    uuid.NewString(),
        UserID:     user.ID,
        Email:      user.Email,
        Name:       user.Name,
        OccurredAt: time.Now(),
    })

    return user, accessToken, refreshToken.Token, nil
}

// repository/postgres/user_repo.go (реализация)
type UserRepo struct { db *sql.DB }

func (r *UserRepo) Create(ctx, user *domain.User) error {
    query := `INSERT INTO users (id, email, name, password_hash, created_at) VALUES ($1, $2, $3, $4, $5)`
    _, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Name, user.PasswordHash, user.CreatedAt)
    return err
}

// infrastructure/nats.go (реализация)
type NATSPublisher struct { js nats.JetStreamContext }

func (p *NATSPublisher) PublishUserCreated(ctx, event events.UserCreated) error {
    data, _ := json.Marshal(event)
    _, err := p.js.Publish("user.created", data)
    return err
}
```

### Dependency Injection (ручная, в main.go)

```go
// cmd/server/main.go
func main() {
    // 1. Infrastructure
    db := infrastructure.NewPostgresDB(os.Getenv("DATABASE_URL"))
    publisher := infrastructure.NewNATSPublisher(os.Getenv("NATS_URL"))
    hasher := infrastructure.NewBcryptPool(0, bcrypt.DefaultCost)
    tokenGenerator := infrastructure.NewJWTGenerator(privateKey, publicKey, ...)

    // 2. Repositories
    userRepo := postgres.NewUserRepo(db)
    refreshTokenRepo := postgres.NewRefreshTokenRepo(db)

    // 3. Usecases (внедряем зависимости)
    authUC := usecase.NewAuthUseCase(
        userRepo,           // интерфейс UserRepository
        refreshTokenRepo,   // интерфейс RefreshTokenRepository
        tokenGenerator,     // интерфейс TokenGenerator
        publisher,          // интерфейс EventPublisher
        hasher,             // интерфейс Hasher
        7*24*time.Hour,     // refresh token TTL
    )

    // 4. Delivery
    handler := grpc.NewAuthHandler(authUC)  // внедряем usecase

    // 5. gRPC Server
    grpcServer := grpc.NewServer()
    authpb.RegisterAuthServiceServer(grpcServer, handler)
    grpcServer.Serve(listener)
}
```

**Почему без framework (Wire, Dig):**
- Простота (видно все зависимости явно)
- Малое количество сервисов (не нужна автоматизация)
- Понятно для junior-разработчиков

---

## Event-Driven Architecture (детально)

### NATS JetStream

**Что это:**
- Broker сообщений с гарантией доставки (at-least-once)
- Streams — append-only log событий
- Consumers — подписчики на stream с tracking прогресса
- Acknowledgments — consumer подтверждает обработку

**Конфигурация Yammi:**

**Stream: USERS**
```go
{
    Name:     "USERS",
    Subjects: []string{"user.created", "user.deleted"},
    Storage:  FileStorage,  // сохраняется на диск
    MaxAge:   24 * time.Hour,  // события хранятся 24 часа
}
```

**Consumer: user-service-user-created-v4**
```go
{
    Durable:       "user-service-user-created-v4",
    FilterSubject: "user.created",
    AckWait:       30 * time.Second,  // если не Ack → retry
    MaxDeliver:    7,  // макс 7 попыток
    MaxAckPending: 500,  // не более 500 необработанных
    AckPolicy:     AckExplicit,  // manual Ack (не auto)
}
```

### Event Flow (детально)

**1. Auth Service публикует событие:**

```go
// internal/usecase/register.go
event := events.UserCreated{
    EventID:      uuid.NewString(),      // для idempotency
    EventVersion: 1,                      // schema version
    OccurredAt:   time.Now(),
    UserID:       user.ID,
    Email:        user.Email,
    Name:         user.Name,
}

data, _ := json.Marshal(event)
uc.eventPublisher.Publish("user.created", data)  // async, no wait
```

**2. NATS принимает событие:**
- Добавляет в stream USERS (append-only log)
- Уведомляет всех подписчиков (consumers)

**3. User Service получает событие:**

```go
// internal/infrastructure/nats_consumer.go
msg, _ := subscription.NextMsg(ctx)

var event events.UserCreated
json.Unmarshal(msg.Data, &event)

// Обработка
err := uc.HandleUserCreated(ctx, event)

if err != nil {
    msg.Nak()  // negative acknowledgment → retry
    return
}

msg.Ack()  // positive acknowledgment → NATS удаляет из pending
```

**4. Retry (если Nak):**
- NATS ждет `AckWait` (30 секунд)
- Если за 30s не получил Ack → считает что обработка failed
- Exponential backoff в usecase: 2s → 4s → 8s → 16s → 30s
- После 7 попыток → отправка в DLQ

**5. Dead Letter Queue (DLQ):**

```go
// После MaxDeliver попыток (7)
dlqEvent := events.DLQEnvelope{
    OriginalSubject: "user.created",
    OriginalData:    msg.Data,
    Error:           err.Error(),
    Attempts:        msg.Metadata.NumDelivered,
    FailedAt:        time.Now(),
}

data, _ := json.Marshal(dlqEvent)
js.Publish("dlq.user.created", data)
```

**DLQ Monitor:**
- Отдельный consumer подписан на `dlq.user.>`
- Логирует каждое событие в DLQ (алерт для DevOps)
- Не обрабатывает (требует ручного вмешательства)

### Idempotency

**Проблема:** At-least-once delivery → событие может прийти дважды.

**Решение 1: Idempotent operations (User Service):**
```go
func (uc *UserUseCase) HandleUserCreated(ctx, event events.UserCreated) error {
    // Проверяем существует ли профиль
    _, err := uc.repo.GetByEmail(ctx, event.Email)
    if err == nil {
        return nil  // профиль уже создан → Ack без ошибки
    }

    // Создаем профиль
    profile := domain.NewUserFromEvent(event)
    return uc.repo.Create(ctx, profile)
}
```

Если событие придет дважды → `GetByEmail` найдет профиль → early return → Ack.

**Решение 2: Deduplication table (для non-idempotent operations):**
```sql
CREATE TABLE processed_events (
    event_id UUID PRIMARY KEY,
    processed_at TIMESTAMPTZ
);
```

```go
func (uc *SomeUseCase) HandleEvent(ctx, event Event) error {
    // Проверяем обрабатывали ли раньше
    exists, _ := uc.repo.EventProcessed(ctx, event.EventID)
    if exists {
        return nil  // уже обрабатывали → skip
    }

    // Обрабатываем + сохраняем event_id в одной транзакции
    tx, _ := uc.db.Begin()
    defer tx.Rollback()

    uc.doSomething(ctx, tx, event)
    tx.Exec("INSERT INTO processed_events (event_id) VALUES ($1)", event.EventID)

    tx.Commit()
    return nil
}
```

**Yammi:** Использует Решение 1 (idempotent operations). Решение 2 — для future features.

### Event Versioning

**Проблема:** Схема события меняется (добавляем новое поле).

**Решение: Event Version field:**

```go
type UserCreated struct {
    EventVersion int  `json:"event_version"`  // 1 → 2
    UserID       string
    Email        string
    Name         string
    AvatarURL    string  `json:"avatar_url,omitempty"`  // добавлено в v2
}
```

**Consumer обрабатывает обе версии:**
```go
func (uc *UserUseCase) HandleUserCreated(ctx, event events.UserCreated) error {
    switch event.EventVersion {
    case 1:
        // Старая схема (без AvatarURL)
        profile := &domain.User{
            ID:    event.UserID,
            Email: event.Email,
            Name:  event.Name,
        }
        return uc.repo.Create(ctx, profile)

    case 2:
        // Новая схема (с AvatarURL)
        profile := &domain.User{
            ID:        event.UserID,
            Email:     event.Email,
            Name:      event.Name,
            AvatarURL: event.AvatarURL,
        }
        return uc.repo.Create(ctx, profile)

    default:
        return fmt.Errorf("unknown event version: %d", event.EventVersion)
    }
}
```

**Альтернатива:** Создать новый subject `user.created.v2` и новый consumer.

---

## Инфраструктура

### PostgreSQL (5 баз)

```
yammi_auth         — users, refresh_tokens
yammi_user         — profiles
yammi_board        — boards, board_members, columns, cards (partitioned)
yammi_comment      — comments (WIP)
yammi_notification — notifications (WIP)
```

**Почему отдельные базы:**
- Database per service pattern (фундаментальное правило микросервисов)
- Нельзя делать JOIN между сервисами (только через gRPC)
- Независимое масштабирование (разные БД на разных серверах)

**Connection string (пример):**
```
DATABASE_URL=postgres://user:pass@postgres:5432/yammi_auth?sslmode=disable
```

### Redis (cache)

**Используется:** Board Service (read-through cache)

**Стратегия:**
- Cache keys: `board:{id}`, `columns:{board_id}`, `cards:{column_id}`
- TTL: 5 минут
- Invalidation: при UPDATE/DELETE соответствующей entity

**Не используется:** Session storage (JWT stateless), rate limiting (in-memory).

### NATS (event bus)

**Порты:**
- `4222` — client connections
- `8222` — HTTP monitoring

**Streams:**
- `USERS` (subjects: `user.created`, `user.deleted`)
- `BOARDS` (subjects: `board.*`, `card.*`, `column.*`) — WIP
- `DLQ` (subjects: `dlq.user.>`, `dlq.board.>`)

**Consumers:**
- `user-service-user-created-v4` (User Service)
- `user-service-user-deleted-v1` (User Service)
- `dlq-monitor` (DLQ Monitor)

### Monitoring (Prometheus + Grafana)

**Prometheus:**
- Scraping интервал: 5 секунд
- Targets: NATS exporter (пока только NATS метрики)

**Grafana:**
- Dashboards: NATS JetStream, User Deleted Events
- Datasource: Prometheus (auto-provisioned)
- URL: `http://localhost:3033`

**NATS Exporter:**
- Конвертирует NATS metrics → Prometheus format
- Metrics: stream message count, consumer lag, deliver count

---

## Масштабирование

### Horizontal Scaling

**Stateless сервисы (можно масштабировать):**
- API Gateway (x N реплик за load balancer)
- Auth Service (уже x5 реплик)
- User Service (x N)
- Board Service (x N)

**Stateful компоненты:**
- PostgreSQL — master-replica (write → master, read → replicas)
- Redis — Redis Cluster (sharding)
- NATS — built-in clustering (3+ nodes)

### Vertical Scaling

**CPU-bound:**
- Auth Service (bcrypt hashing) → больше CPU
- Board Service (много запросов) → больше CPU

**Memory-bound:**
- Redis Cache → больше RAM
- PostgreSQL (indexes) → больше RAM

### Database Scaling

**Partitioning:**
- Cards — уже HASH partitioned по board_id (4 партиции)
- При росте → увеличить до 8, 16 партиций

**Sharding (будущее):**
- Разные boards на разных серверах PostgreSQL
- Shard key: board_id

**Read Replicas:**
- Master (write) + 3 replicas (read)
- API Gateway → read from replicas

---

## Security

### Authentication

**JWT (EdDSA):**
- Access token в Authorization header: `Bearer <token>`
- Refresh token в cookie (HttpOnly, Secure)
- Access token short-lived (15 min) → minimize impact при leak
- Refresh token rotation (старый revoke при refresh)

### Authorization

**Board Service:**
- Owner → full access
- Member → CRUD cards, read board
- Проверка в usecase (не в delivery)

**API Gateway:**
- OwnerOnly middleware → блокирует чужие PUT/DELETE `/users/{id}`

### Rate Limiting

**API Gateway:**
- Token bucket per IP
- 50 req/min для register/login
- 50 req/min default для остальных

**Защита от:**
- Brute-force password guessing
- DDoS

### SQL Injection

**Защита:** Prepared statements (PostgreSQL placeholders `$1, $2, ...`)

```go
// ✅ Safe
db.Exec("SELECT * FROM users WHERE email = $1", email)

// ❌ Unsafe (не используется в Yammi)
db.Exec("SELECT * FROM users WHERE email = '" + email + "'")
```

### CORS

**API Gateway:**
- Разрешены origins: `http://localhost:3000` (dev), `https://yammi.com` (prod)
- Credentials: allowed (для cookies)

---

## Trade-offs и ограничения

### Eventual Consistency

**Проблема:**
```
User регистрируется → Auth создает user → профиль создается асинхронно через NATS
→ Между регистрацией и созданием профиля может пройти 100-500ms
→ Клиент делает GET /users/{id} → 404 (профиль еще не создан)
```

**Решение:**
- Frontend polling (опрашивает GET /users/{id} каждые 200ms до успеха)
- Timeout 5 секунд (если не создался → ошибка)

**Trade-off:** Сложнее UX vs высокая доступность (Auth Service не зависит от User Service).

### No Distributed Transactions

**Проблема:**
```
Board Service: CREATE board + ADD owner to board_members
→ Если CREATE успешно, но ADD упало → board без owner (inconsistency)
```

**Решение в Yammi:**
- Обе операции в одной транзакции (в одном сервисе → можно ACID)

**Проблема 2:**
```
Auth Service: DELETE user + NATS publish user.deleted
→ Если DELETE успешно, но NATS publish упал → User Service не удалит профиль
```

**Решение:**
- Retry NATS publish (best-effort)
- Если упало → warning в лог (не fail транзакции)
- Eventual consistency (профиль удалится позже через manual trigger или cronjob)

### No SAGA Pattern

**Что это:** Distributed transaction через компенсирующие транзакции.

**Пример:**
```
1. Auth Service: CREATE user
2. User Service: CREATE profile
3. Если (2) упало → Auth Service: DELETE user (компенсация)
```

**Yammi не использует:** Сложность не оправдана для текущего масштаба. При росте — добавим.

---

## Итоговая таблица: что за что отвечает

| Сервис | Порт | Протокол | Ответственность | Статус |
|--------|------|----------|-----------------|--------|
| API Gateway | 8080 | HTTP | HTTP ↔ gRPC, JWT verify, rate limit | ✅ Done |
| Auth | 50051 | gRPC | Register, Login, Tokens, Events | ✅ Done |
| User | 50052 | gRPC | Profiles, NATS consumer | ✅ Done |
| Board | 50053 | gRPC | Boards, Columns, Cards, Sharing | 🚧 WIP |
| Comment | 50054 | gRPC | Comments | ⏸️ Planned |
| Notification | 50055 | — | NATS consumer, Notifications | ⏸️ Planned |
| WebSocket Gateway | 8081 | WS | Real-time push, NATS consumer | ⏸️ Planned |

**Infrastructure:**
| Компонент | Порт | Назначение |
|-----------|------|------------|
| PostgreSQL | 5432 | 5 баз данных (yammi_auth, yammi_user, ...) |
| Redis | 6380 | Cache (Board Service) |
| NATS | 4222, 8222 | Event bus (JetStream) |
| Prometheus | 9090 | Metrics storage |
| Grafana | 3033 | Dashboards |
