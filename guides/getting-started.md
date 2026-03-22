# Yammi — Полный гайд по проекту

Этот документ объясняет как устроен проект изнутри: зачем каждый компонент, как они связаны, что откуда берётся и куда идёт.

## Общая идея

Yammi — это Trello-подобная доска задач. Пользователь регистрируется, создаёт доски с колонками, добавляет карточки, перетаскивает их между колонками. Другие участники доски видят изменения в реальном времени через WebSocket.

Проект построен на **микросервисной архитектуре**: вместо одного большого приложения — 7 отдельных сервисов, каждый отвечает за свою область. Они общаются друг с другом двумя способами:
- **gRPC** — синхронные запросы (клиент ждёт ответ)
- **NATS** — асинхронные события (отправил и забыл, получатель обработает когда сможет)

## Как выглядит путь запроса

```
Клиент (браузер / Postman)
    │
    │  HTTP POST /api/v1/auth/register
    ▼
┌─────────────────┐
│  API Gateway    │  ← проверяет rate limit, для защищённых роутов проверяет JWT
│  :8080          │
└────────┬────────┘
         │  gRPC (protobuf)
         ▼
┌─────────────────┐
│  Auth Service   │  ← хеширует пароль, создаёт пользователя в БД, генерирует JWT
│  :50051         │
└────────┬────────┘
         │  NATS JetStream (событие user.created)
         ▼
┌─────────────────┐
│  User Service   │  ← ловит событие, создаёт профиль (имя, аватар, bio)
│  :50052         │
└─────────────────┘
```

**Ключевой момент**: Auth Service не вызывает User Service напрямую. Вместо этого он публикует событие `user.created` в NATS, а User Service подписан на эти события и сам создаёт профиль. Это и есть **event-driven архитектура** — сервисы не знают друг о друге, они общаются через события.

## Зачем столько инфраструктуры?

### PostgreSQL — основная база данных

У каждого сервиса **своя отдельная база данных**. Это фундаментальное правило микросервисов: сервис не лезет в чужую базу. Если User Service нужны данные из Auth — он идёт через gRPC, а не делает SQL-запрос в `yammi_auth`.

При первом запуске скрипт `scripts/init-databases.sql` создаёт 5 баз:
```
yammi_auth         — пользователи (email, password_hash) + refresh-токены
yammi_user         — профили (имя, аватар, bio)
yammi_board        — доски, колонки, карточки (пока не реализовано)
yammi_comment      — комментарии (пока не реализовано)
yammi_notification — уведомления (пока не реализовано)
```

Каждый сервис при старте сам накатывает свои **миграции** — SQL-файлы из папки `migrations/`. Например, Auth Service создаёт таблицы `users` и `refresh_tokens`.

### NATS — шина событий (Event Bus)

NATS — это брокер сообщений. Когда что-то происходит (зарегистрировался пользователь, удалили карточку), сервис публикует **событие** в NATS. Другие сервисы, которым это интересно, подписаны на эти события и реагируют.

Мы используем **NATS JetStream** — это надстройка над обычным NATS, которая гарантирует доставку: если сервис-получатель упал, сообщение не потеряется и будет доставлено когда сервис поднимется.

Сейчас реализованы два потока событий:
```
Auth Service  ──publish──▶  NATS стрим "USERS"  ──subscribe──▶  User Service
                            (субъекты: user.created, user.deleted)
```

**Как работает на практике:**
1. Пользователь регистрируется через API Gateway → Auth Service
2. Auth Service сохраняет `email + password_hash` в `yammi_auth.users`
3. Auth Service публикует событие `user.created` с данными `{user_id, email, name}` в NATS
4. User Service получает это событие и создаёт запись в `yammi_user.profiles`
5. Теперь у пользователя есть и учётные данные (auth), и профиль (user)

**Зачем это нужно вместо простого HTTP-вызова?**
- Auth Service не зависит от User Service. Если User Service упал — регистрация всё равно пройдёт, а профиль создастся позже
- Можно добавить новых подписчиков без изменения Auth Service (например, Notification Service тоже подпишется на `user.created` и отправит welcome-email)

### Redis — кеш

Redis используется Board Service для кеширования досок (read-through cache). При запросе доски:
1. Сначала проверяем Redis
2. Если нет (cache miss) — идём в PostgreSQL
3. Кладём результат в Redis с TTL 5 минут

Пока Board Service не реализован, Redis поднят, но не используется.

### Prometheus — сбор метрик

Prometheus — система мониторинга. Она **сама ходит** на все сервисы по расписанию (каждые 5 секунд) и забирает метрики. Это называется **pull-model** (Prometheus тянет, а не сервисы пушат).

Сейчас настроен сбор метрик с **NATS** через отдельный контейнер `nats-exporter`:
```
Prometheus  ──scrape каждые 5s──▶  nats-exporter:7777  ──читает──▶  NATS :8222
```

`nats-exporter` превращает внутреннюю статистику NATS (количество соединений, сообщений, consumer lag) в формат, который понимает Prometheus.

Конфиг: `deployments/monitoring/prometheus/prometheus.yml`

### Grafana — визуализация метрик

Grafana — это веб-интерфейс для построения дашбордов из метрик Prometheus. Вместо того чтобы писать PromQL-запросы руками, ты смотришь на графики.

Настроено два дашборда:
- **NATS JetStream** — общее состояние стримов, сколько сообщений в секунду, consumer lag
- **NATS User Deleted** — мониторинг обработки событий `UserDeleted` (отслеживает, не застряли ли события)

Дашборды подгружаются автоматически через **provisioning** (Grafana при старте читает JSON-файлы из `deployments/monitoring/grafana/dashboards/` и создаёт дашборды).

Grafana подключена к Prometheus как datasource — тоже через provisioning (`deployments/monitoring/grafana/provisioning/datasources/`).

### NATS Monitoring UI

У NATS есть встроенный HTTP-мониторинг на порту 8222. Это не красивый дашборд, а raw JSON с информацией: сколько подключений, сколько сообщений, состояние стримов. Полезно для быстрой диагностики.

## Docker-контейнеры

`docker compose up --build` поднимает 14 контейнеров:

### Инфраструктура (3 контейнера)

| Контейнер   | Образ              | Порт        | Зачем                                      |
|-------------|--------------------|-------------|---------------------------------------------|
| `postgres`  | postgres:16-alpine | 5432        | Единый сервер PostgreSQL, 5 баз внутри      |
| `redis`     | redis:7-alpine     | 6380→6379   | Кеш для Board Service                      |
| `nats`      | nats:2-alpine      | 4222, 8222  | Event Bus (JetStream) + мониторинг         |

**Почему один PostgreSQL, а не 5?** В продакшене каждый сервис имел бы свою инстансу БД. Для локальной разработки это избыточно — проще один сервер с 5 базами. Изоляция данных всё равно соблюдается: сервисы используют разные connection strings (`yammi_auth`, `yammi_user`, ...).

**Почему Redis на порту 6380, а не 6379?** Чтобы не конфликтовать с Redis, который может быть установлен локально на хосте. Внутри Docker-сети сервисы всё равно обращаются по стандартному порту `redis:6379`.

### Мониторинг (3 контейнера)

| Контейнер       | Образ                                 | Порт  | Зачем                                        |
|-----------------|---------------------------------------|-------|-----------------------------------------------|
| `nats-exporter` | natsio/prometheus-nats-exporter       | —     | Конвертирует метрики NATS в формат Prometheus |
| `prometheus`    | prom/prometheus                       | 9090  | Собирает и хранит метрики                     |
| `grafana`       | grafana/grafana                       | 3033  | Веб-дашборды для визуализации метрик          |

**Цепочка**: NATS → nats-exporter → Prometheus (scrape) → Grafana (query).

### Сервисы (7 контейнеров)

| Контейнер      | Порт  | Реплик | Протокол | Статус         |
|----------------|-------|--------|----------|----------------|
| `auth`         | —     | **5**  | gRPC     | Реализован     |
| `user`         | 50052 | 1      | gRPC     | Реализован     |
| `api-gateway`  | 8080  | 1      | HTTP     | Реализован     |
| `board`        | 50053 | 1      | gRPC     | Заглушка       |
| `comment`      | 50054 | 1      | gRPC     | Заглушка       |
| `notification` | 50055 | 1      | gRPC     | Заглушка       |
| `gateway`      | 8081  | 1      | HTTP/WS  | Заглушка       |

**Почему Auth Service запущен в 5 репликах?** Аутентификация — самая нагруженная точка. Каждый запрос к защищённому эндпоинту начинается с проверки токена. Bcrypt-хеширование паролей — CPU-интенсивная операция. 5 реплик позволяют распределить нагрузку.

**Как 5 реплик работают с одним JWT-ключом?** Все реплики получают одинаковый `JWT_SEED` через environment variable. Из seed детерминированно генерируется пара ключей EdDSA. Результат: все реплики подписывают токены одним ключом, и любая реплика может отдать правильный публичный ключ.

**Почему у Auth нет проброшенного порта, а у остальных сервисов есть?** Auth Service доступен через Docker DNS-имя `auth:50051`. API Gateway обращается к нему по `dns:///auth:50051` с round-robin балансировкой. Прямой доступ извне не нужен — всё идёт через API Gateway. У остальных порты проброшены для удобства отладки.

### Инструменты (1 контейнер, по запросу)

| Контейнер | Профиль | Зачем                                      |
|-----------|---------|---------------------------------------------|
| `dlq`     | tools   | CLI для работы с Dead Letter Queue          |

Контейнер `dlq` не запускается по умолчанию (профиль `tools`). Запуск вручную:
```bash
docker compose run --rm dlq list
```

## Структура сервиса (Clean Architecture)

Каждый реализованный сервис следует одной и той же структуре. Разберём на примере Auth Service:

```
services/auth/
├── cmd/server/main.go           ← точка входа, собирает всё вместе (DI)
├── api/proto/v1/auth.proto      ← контракт gRPC API (protobuf)
├── internal/
│   ├── domain/                  ← бизнес-сущности, 0 зависимостей от фреймворков
│   │   ├── user.go              ← сущность User + валидация регистрации
│   │   ├── token.go             ← сущность RefreshToken + IsValid/Revoke
│   │   └── errors.go            ← типизированные ошибки (ErrEmailExists, ...)
│   ├── usecase/                 ← бизнес-сценарии, оркестрирует domain
│   │   ├── auth.go              ← структура AuthUseCase, конструктор
│   │   ├── interfaces.go        ← интерфейсы (UserRepository, TokenGenerator, ...)
│   │   ├── register.go          ← Register + DeleteUser
│   │   ├── login.go             ← Login
│   │   └── token.go             ← RefreshToken + RevokeToken
│   ├── delivery/grpc/           ← входная точка (обработка gRPC-запросов)
│   │   └── handler.go           ← AuthHandler, маппинг proto → usecase → proto
│   ├── repository/postgres/     ← реализация интерфейсов (SQL-запросы)
│   │   ├── user_repo.go
│   │   └── refresh_token_repo.go
│   └── infrastructure/          ← внешние зависимости (БД, JWT, NATS, ...)
│       ├── database.go          ← подключение к PostgreSQL
│       ├── migrator.go          ← накат миграций из SQL-файлов
│       ├── jwt.go               ← генерация JWT (EdDSA) + KeyPairFromSeed
│       ├── nats.go              ← NATS publisher (UserCreated, UserDeleted)
│       └── hasher.go            ← bcrypt pool (ограниченный параллелизм)
├── migrations/
│   └── 000001_init.up.sql       ← SQL: таблицы users, refresh_tokens
└── Dockerfile                   ← multi-stage build с protoc
```

### Правило зависимостей

```
delivery → usecase → domain
              ↓
           repository (интерфейс определён в usecase, реализация в infrastructure)
```

- **domain** не импортирует ничего кроме стандартной библиотеки Go и uuid. Это чистые бизнес-правила
- **usecase** определяет **интерфейсы** (UserRepository, TokenGenerator, EventPublisher) и работает с ними. Не знает про PostgreSQL или NATS
- **infrastructure** реализует эти интерфейсы конкретными технологиями
- **delivery** принимает внешние запросы (gRPC/HTTP) и вызывает usecase

### Как это собирается вместе

`cmd/server/main.go` — это место, где происходит **Dependency Injection** (ручная, без фреймворков):

```go
// 1. Создаём инфраструктуру
db := infrastructure.NewPostgresDB(databaseURL)
publisher := infrastructure.NewNATSPublisher(natsURL)
hasher := infrastructure.NewBcryptPool(0, bcryptCost)
tokenGenerator := infrastructure.NewJWTGenerator(privateKey, publicKey, "yammi-auth", 15*time.Minute)

// 2. Создаём репозитории (реализация интерфейсов)
userRepo := postgres.NewUserRepo(db)
refreshTokenRepo := postgres.NewRefreshTokenRepo(db)

// 3. Создаём usecase, передаём зависимости
authUC := usecase.NewAuthUseCase(userRepo, refreshTokenRepo, tokenGenerator, publisher, hasher, 7*24*time.Hour)

// 4. Создаём delivery handler, передаём usecase
handler := delivery.NewAuthHandler(authUC)

// 5. Регистрируем handler на gRPC сервере
grpcServer := grpc.NewServer()
authpb.RegisterAuthServiceServer(grpcServer, handler)
```

## Детально: как работает каждый сервис

### Auth Service — аутентификация

**Задача**: регистрация, логин, управление JWT-токенами.

**Таблицы в `yammi_auth`:**

```sql
users (id UUID, email UNIQUE, name, password_hash, created_at, updated_at)
refresh_tokens (id UUID, user_id FK→users, token UNIQUE, expires_at, revoked, created_at)
```

**Сценарий: Register**
1. Валидация в domain: email не пустой, содержит `@`, пароль >= 8 символов, имя не пустое
2. Хеширование пароля через bcrypt pool (ограниченный параллелизм через семафор)
3. Создание User в БД (UUID генерируется в domain)
4. Генерация JWT access token (EdDSA, TTL 15 минут)
5. Создание RefreshToken в БД (UUID, TTL 7 дней)
6. Публикация события `user.created` в NATS (async, если не удалось — warning в лог, регистрация не отменяется)
7. Возврат `{user_id, access_token, refresh_token}`

**Сценарий: Login**
1. Поиск пользователя по email
2. Проверка пароля через bcrypt.CompareHashAndPassword
3. Генерация нового access token + refresh token
4. Возврат `{user_id, access_token, refresh_token}`

**Сценарий: RefreshToken**
1. Поиск refresh token в БД
2. Проверка: не revoked? не expired?
3. Revoke старый refresh token
4. Создание нового access token + refresh token (rotation)
5. Возврат новой пары токенов

**Сценарий: DeleteUser**
1. Удаление пользователя из БД (CASCADE удалит refresh_tokens)
2. Публикация события `user.deleted` в NATS

**JWT-ключи (EdDSA)**:
- Ed25519 — алгоритм цифровой подписи (быстрее RSA, короче ключи)
- Приватный ключ — только в Auth Service (подписывает токены)
- Публичный ключ — раздаётся через gRPC метод `GetPublicKey` (API Gateway забирает для локальной верификации)
- `JWT_SEED` — base64-encoded 32 байта, из которых **детерминированно** генерируется пара ключей. Все 5 реплик получают одинаковый seed → одинаковые ключи

**Bcrypt Pool**:
Хеширование bcrypt — CPU-тяжёлая операция. Если 100 запросов на регистрацию придут одновременно, они все начнут хешировать пароли и загрузят CPU. `BcryptPool` ограничивает параллелизм через буферизированный канал (семафор). По умолчанию — `runtime.NumCPU()` одновременных операций.

### User Service — профили пользователей

**Задача**: хранение и редактирование профилей (имя, аватар, bio).

**Таблица в `yammi_user`:**

```sql
profiles (id UUID, email UNIQUE, name, avatar_url, bio, created_at, updated_at)
```

**Ключевая особенность**: User Service **не имеет эндпоинта для создания профиля**. Профиль создаётся **автоматически** при получении события `user.created` из NATS.

**NATS Consumer**:
- Durable consumer `user-service-user-created-v4` — подписка на `user.created`
- Durable consumer `user-service-user-deleted-v1` — подписка на `user.deleted`
- `MaxDeliver: 7` — максимум 7 попыток доставки
- `AckWait: 30s` — если за 30 секунд не подтвердил обработку — NATS повторит
- `MaxAckPending: 500` — не более 500 необработанных сообщений одновременно

**Retry с exponential backoff**:
Если обработка события упала — NATS повторит доставку. Задержка между попытками растёт экспоненциально: 2s → 4s → 8s → 16s → 30s (максимум). Добавляется jitter ±20% чтобы повторные попытки от разных событий не приходили одновременно.

**Dead Letter Queue (DLQ)**:
Если после 7 попыток событие всё ещё не обработано — оно отправляется в DLQ-стрим. DLQ Monitor подписан на `dlq.user.>` и логирует каждое такое событие (subject, ошибка, количество попыток, payload). Это **алерт** — DLQ не должен быть пустым в нормальном режиме.

**Versioning consumers**: Имя consumer включает версию (`-v4`). При изменении логики обработки событий создаётся новый consumer с новой версией. Старый consumer остаётся в JetStream и должен быть удалён вручную.

**Idempotency**: Если событие `user.created` пришло повторно и профиль уже существует (email уже занят) — consumer просто делает `Ack()` без ошибки.

### API Gateway — точка входа

**Задача**: единая HTTP-точка для клиента. Проксирует запросы в gRPC-сервисы, проверяет JWT, ограничивает частоту запросов.

**Роуты:**

```
POST   /api/v1/auth/register     ← публичный, rate limit 50/мин
POST   /api/v1/auth/login         ← публичный, rate limit 50/мин
GET    /api/v1/auth/public-key    ← публичный, без rate limit
POST   /api/v1/auth/refresh       ← требует JWT, rate limit 50/мин
POST   /api/v1/auth/revoke        ← требует JWT, rate limit 50/мин
GET    /api/v1/users/{id}         ← требует JWT, rate limit 50/мин
PUT    /api/v1/users/{id}         ← требует JWT + OwnerOnly
DELETE /api/v1/users/{id}         ← требует JWT + OwnerOnly
GET    /health                    ← проверка что gateway жив
```

**JWT Verification**:
API Gateway **не ходит в Auth Service на каждый запрос**. При старте он один раз забирает публичный ключ через gRPC `GetPublicKey` и кеширует его. Дальше каждый JWT проверяется локально.

Если ключ сменился (Auth Service перезапустился без `JWT_SEED`), JWTVerifier пробует перезагрузить ключ, но не чаще чем раз в 30 секунд (cooldown защищает от DoS — злоумышленник не может заспамить Auth Service невалидными токенами).

**Rate Limiting (Token Bucket)**:
Каждый IP получает "ведро" с токенами. Каждый запрос тратит 1 токен. Токены пополняются со скоростью `maxRequests / period`. Если токенов нет — ответ `429 Too Many Requests`.

Реализация in-memory (без Redis): `map[IP]*rateLimitEntry` под мьютексом. Фоновая горутина каждые 5 минут чистит записи для IP, чей bucket полностью восстановился. Несколько лимитеров: отдельные для register, login, refresh, и общий default.

**gRPC Clients**:
Подключение к Auth и User сервисам с настройками:
- `round_robin` load balancing (для Auth с 5 репликами)
- `dns:///auth:50051` — Docker DNS резолвит имя `auth` во все 5 IP-адресов реплик
- Keepalive: ping каждые 10 секунд, таймаут 3 секунды
- Автоматический reconnect с backoff

**OwnerOnly middleware**:
Для PUT/DELETE `/users/{id}` проверяет что `user_id` из JWT совпадает с `{id}` из URL. Нельзя редактировать чужой профиль.

### WebSocket Gateway — real-time обновления (WIP)

Пока только заглушка. `/health` возвращает `{"status": "ok"}`, `/ws` возвращает `501 Not Implemented`.

Планируемая работа: подписка на события Board Service через NATS, push обновлений клиентам через WebSocket. НЕ ходит в сервисы синхронно — только слушает события.

### Board, Comment, Notification — заглушки

Каждый имеет `main.go` который просто слушает порт, без бизнес-логики. Структура папок создана (domain, usecase, delivery, repository, infrastructure) с `.gitkeep` файлами.

## Протоколы и контракты

### gRPC + Protobuf

Сервисы общаются через **gRPC** — бинарный протокол поверх HTTP/2. Контракты описываются в `.proto` файлах (Protocol Buffers):

```protobuf
// services/auth/api/proto/v1/auth.proto
service AuthService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc RevokeToken(RevokeTokenRequest) returns (RevokeTokenResponse);
  rpc GetPublicKey(GetPublicKeyRequest) returns (GetPublicKeyResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
}
```

Из `.proto` файла генерируется Go-код (`protoc --go_out=... --go-grpc_out=...`). Генерация происходит при сборке Docker-образа (в Dockerfile). Сгенерированные файлы (`*.pb.go`, `*_grpc.pb.go`) **не** хранятся в Git.

**Почему gRPC, а не REST между сервисами?**
- Строгий контракт (proto файл) — компилятор ловит несоответствия
- Бинарная сериализация (protobuf) — быстрее и компактнее JSON
- HTTP/2 — мультиплексирование, стриминг
- Кодогенерация клиентов и серверов

### NATS события

Структуры событий определены в `pkg/events/user.go` — shared пакет, который импортируют и publisher (Auth), и consumer (User):

```go
type UserCreated struct {
    EventID      string    `json:"event_id"`       // UUID, для idempotency
    EventVersion int       `json:"event_version"`  // версия схемы события
    OccurredAt   time.Time `json:"occurred_at"`
    UserID       string    `json:"user_id"`
    Email        string    `json:"email"`
    Name         string    `json:"name"`
}
```

Поля:
- `event_id` — уникальный ID события. Если consumer получит одно событие дважды, он может проверить event_id и пропустить повторное
- `event_version` — версия схемы. При добавлении новых полей можно инкрементировать версию, а consumer будет обрабатывать разные версии по-разному
- `occurred_at` — когда произошло событие (для аудита и отладки)

## Сборка Docker-образов

Каждый сервис имеет **multi-stage Dockerfile**:

```dockerfile
# Стадия 1: сборка
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache protobuf protobuf-dev git
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /build
COPY pkg/ pkg/                          # shared пакет с событиями
COPY services/auth/ services/auth/      # код сервиса

WORKDIR /build/services/auth
RUN protoc ... api/proto/v1/auth.proto  # генерация Go-кода из proto
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/auth ./cmd/server

# Стадия 2: минимальный образ
FROM alpine:3.19
COPY --from=builder /app/auth /app/auth
COPY --from=builder .../migrations /app/migrations
CMD ["/app/auth"]
```

**Почему multi-stage?** Финальный образ содержит только скомпилированный бинарник (~15 MB) вместо всего Go SDK (~500 MB). В продакшене это экономит место и уменьшает поверхность атаки.

**Почему `CGO_ENABLED=0`?** Go по умолчанию может использовать C-библиотеки (CGO). Отключение CGO даёт полностью статический бинарник, который работает на минимальном Alpine Linux без дополнительных зависимостей.

## Тестирование

В проекте три уровня тестов: unit, интеграционные (feature) и нагрузочные. Никаких моков — все тесты работают с реальной инфраструктурой.

### Unit-тесты

Тестируют domain-логику и вспомогательные функции изолированно. Находятся рядом с кодом (`*_test.go`). Без внешних зависимостей — чистый `testing` из стандартной библиотеки Go.

**Какие файлы:**

```
services/auth/internal/domain/user_test.go        ← ValidateRegistration: email, пароль >= 8, имя
services/auth/internal/domain/token_test.go        ← RefreshToken: IsValid, Revoke, expiration
services/user/internal/domain/user_test.go         ← User.Update: валидация, UpdatedAt
services/user/internal/infrastructure/nats_helpers_test.go  ← backoffDelay: экспоненциальный рост, jitter ±20%
```

**Как запускать:**

```bash
# Все unit-тесты одного сервиса
cd services/auth
go test ./... -v

# Только domain-тесты
cd services/auth
go test ./internal/domain/ -v

# Конкретный тест
cd services/auth
go test ./internal/domain/ -run TestValidateRegistration -v

# Все unit-тесты другого сервиса
cd services/user
go test ./... -v
```

**Что тестируется:**

| Файл | Что проверяет |
|------|---------------|
| `auth/domain/user_test.go` | Пустой email → `ErrEmptyEmail`, нет `@` → `ErrInvalidEmail`, пароль < 8 символов → `ErrWeakPassword`, пустое имя → `ErrEmptyName`, валидные данные → `nil` |
| `auth/domain/token_test.go` | Новый токен валиден, expired → `ErrTokenExpired`, revoked → `ErrTokenRevoked`, revoked + expired → приоритет `ErrTokenRevoked`, UUID уникальны |
| `user/domain/user_test.go` | Update обновляет name/avatar/bio и UpdatedAt, пустое имя → `ErrEmptyName`, NewUserFromEvent сохраняет все поля |
| `user/infrastructure/nats_helpers_test.go` | backoff: attempt 1 → ~2s, attempt 2 → ~4s, attempt 3 → ~8s, attempt 100 → cap 30s, jitter в пределах ±20% (проверяется на 1000 итераций) |

### Интеграционные (feature) тесты

Полный HTTP lifecycle через реальные Docker-контейнеры. Тесты отправляют HTTP-запросы в API Gateway и проверяют всю цепочку: gateway → gRPC → сервис → БД → NATS → другой сервис.

**Структура:**

```
tests/integration/
├── main_test.go              ← точка входа: TestMain + TestUserLifecycle (36 подтестов)
├── register_test.go          ← 6 сценариев регистрации
├── login_test.go             ← 4 сценария логина
├── profile_test.go           ← 6 сценариев профиля (включая ожидание NATS-события)
├── auth_middleware_test.go   ← 9 сценариев JWT-проверки
├── token_test.go             ← 7 сценариев refresh/revoke/rotation
├── lifecycle_test.go         ← 4 сценария полного жизненного цикла
├── api_client.go             ← HTTP-клиент (get/post/put/delete, Bearer-токен)
├── api_client_auth.go        ← методы: Register, Login, RefreshToken, RevokeToken
├── api_client_user.go        ← методы: GetProfile, UpdateProfile, DeleteUser, WaitForProfile
├── response_types.go         ← структуры ответов (AuthResponse, ProfileResponse, ...)
├── assertions.go             ← хелперы: requireStatus, requireNotEmpty, requireEqual
└── go.mod                    ← отдельный Go module (без внешних зависимостей)
```

**Как запускать:**

```bash
# 1. Поднимаем всю инфраструктуру
docker compose up --build -d

# 2. Ждём пока всё стартует (health checks)

# 3. Запускаем тесты
cd tests/integration
go test -v

# С кастомным URL (по умолчанию http://localhost:8080)
API_GATEWAY_URL=http://192.168.1.100:8080 go test -v

# Конкретный подтест (имя из t.Run)
go test -v -run "TestUserLifecycle/01_register_success"
```

**Как устроены тесты изнутри:**

Все 36 подтестов запускаются **последовательно** внутри одного `TestUserLifecycle`. Это сделано намеренно — тесты образуют **цепочку жизненного цикла**: регистрация создаёт пользователя, логин получает токен, профиль использует этот токен, и так далее.

Глобальный `testState` передаёт данные между подтестами:
```go
type testState struct {
    userID       string  // создаётся в register, используется везде дальше
    accessToken  string  // создаётся в login, подставляется в Authorization header
    refreshToken string  // для тестов refresh/revoke
}
```

Если ранний тест упал (например, register), зависимые тесты пропускаются через `t.Skip()`.

**36 подтестов по группам:**

| # | Группа | Что проверяется |
|---|--------|-----------------|
| 01-06 | Registration | Успешная регистрация, дубликат email → 409, пустые поля → 400, слабый пароль → 400 |
| 07-10 | Login | Успешный логин, неверный пароль → 401, несуществующий email → 404, пустые поля → 400 |
| 11-16 | Profile | Ожидание NATS-события (polling 5 секунд), получение профиля, обновление name/avatar/bio, проверка что данные сохранились |
| 17-25 | Auth middleware | Запрос без токена → 401, невалидный токен → 401, чужой профиль → 403, просмотр чужого профиля → 200 (можно смотреть, нельзя менять) |
| 26-32 | Tokens | Revoke refresh token, использование revoked → 401, получение нового через refresh, rotation (старый refresh недействителен), replay protection |
| 33-36 | Lifecycle | Удаление пользователя, логин после удаления → 404, профиль удалён (NATS-событие), повторная регистрация с тем же email |

**Тестирование NATS (async):**

Профиль создаётся **асинхронно** через NATS-событие. Тесты используют **polling** вместо sleep:
```
WaitForProfile(userID, timeout=5s):
  цикл каждые 200ms → GET /api/v1/users/{id}
  если 200 → профиль создан, возвращаем
  если 404 → ещё не дошло, повторяем
  если timeout → test fail
```

Аналогично `WaitForProfileDeletion` — ждёт пока GET вернёт 404 (User Service обработал `user.deleted`).

### Нагрузочные тесты (k6)

JavaScript-сценарии для стресс-тестирования. Используют [k6](https://k6.io/) — инструмент для нагрузочного тестирования.

**Установка k6:**
```bash
# Ubuntu/Debian
sudo apt install k6

# macOS
brew install k6

# Docker (без установки)
docker run --rm -i --net=host grafana/k6 run - < tests/load/register_3000_users.js
```

**Какие сценарии есть:**

```
tests/load/
├── register_3000_users.js   ← плавная нагрузка: 5→75 req/s, ~3000 регистраций
├── burst_3000_users.js      ← взрывная нагрузка: 0→3000 VUs, пик за 50 секунд
└── delete_1000_users.js     ← удаление: setup создаёт 1000 юзеров, потом удаляет под нагрузкой
```

**Как запускать:**

```bash
# 1. Поднимаем инфраструктуру
docker compose up --build -d

# 2. Запускаем нужный сценарий
k6 run tests/load/register_3000_users.js

# С кастомным URL
BASE_URL=http://192.168.1.100:8080 k6 run tests/load/register_3000_users.js

# Через Docker (если k6 не установлен)
docker run --rm -i --net=host grafana/k6 run - < tests/load/register_3000_users.js
```

**Сценарий 1: `register_3000_users.js` — плавная регистрация**

Нагрузка нарастает плавно (ramping arrival rate):
```
Стадия 1 (10s):  5 → 35 req/s    ← прогрев
Стадия 2 (15s): 35 → 75 req/s    ← разгон
Стадия 3 (25s): 75 req/s         ← пиковая нагрузка
Стадия 4 (10s): 75 → 5 req/s     ← остывание
```

Что делает каждый виртуальный пользователь:
1. `POST /api/v1/auth/register` — регистрация с уникальным email
2. Polling `GET /api/v1/users/{id}` — ждёт создания профиля через NATS (до 10 попыток, 300ms между ними)
3. Замеряет latency: от регистрации до появления профиля

Пороги (тест **fail** если превышены):
- Ошибки HTTP: < 1%
- p95 latency: < 600ms
- Ошибки регистрации: < 30 из ~3000

Кастомные метрики:
- `profile_creation_latency_ms` — время от регистрации до создания профиля (NATS delivery latency)
- `profile_created` / `profile_not_ready_yet` — счётчики доставки событий

**Сценарий 2: `burst_3000_users.js` — взрывная нагрузка**

VU (Virtual Users) нарастают быстро (ramping VUs):
```
Стадия 1 (10s):     0 → 500 VUs     ← быстрый старт
Стадия 2 (20s):   500 → 1500 VUs    ← разгон
Стадия 3 (20s):  1500 → 3000 VUs    ← пик
Стадия 4 (10s):  3000 → 0 VUs       ← остановка
Graceful shutdown: 10s               ← дожидаемся in-flight запросов
```

Что делает каждый VU:
1. Регистрация (таймаут 30 секунд — система может быть перегружена)
2. `sleep(3s)` — даём NATS доставить событие
3. `GET /api/v1/users/{id}` — проверяем профиль
4. Если 404 → ещё `sleep(5s)` и повторная попытка
5. `sleep(random 0-1s)` — имитация реального пользователя

Пороги мягче (burst — экстремальная нагрузка):
- Ошибки: < 5%
- p95 register latency: < 5000ms

**Сценарий 3: `delete_1000_users.js` — удаление с каскадной проверкой**

Двухфазный тест:

**Фаза setup (до нагрузки):**
1. Регистрирует 1000 пользователей последовательно (батчами по 100, пауза 300ms между батчами)
2. Ждёт 10 секунд — NATS доставляет события, User Service создаёт профили
3. Сэмплит 10 случайных профилей — проверяет что они созданы

**Фаза нагрузки:**
```
Стадия 1 (10s):  5 → 10 req/s
Стадия 2 (20s): 10 → 25 req/s
Стадия 3 (20s): 25 req/s
Стадия 4 (10s): 25 → 5 req/s
```

Что делает каждый VU:
1. `DELETE /api/v1/users/{id}` — удаление из Auth Service
2. Повторный `DELETE` → ожидаем 404 (подтверждаем что Auth удалил)
3. `sleep(3s)` — ждём NATS-событие `user.deleted`
4. `GET /api/v1/users/{id}` → ожидаем 404 (User Service удалил профиль)
5. Если профиль ещё есть → `sleep(5s)` и повторная проверка

Кастомные метрики:
- `del_latency_ms` — время удаления
- `profile_del_latency_ms` — время каскадного удаления профиля (NATS latency)
- `auth_confirmed_gone` — Auth подтвердил удаление (повторный DELETE → 404)
- `profile_confirmed_gone` / `profile_still_exists` — проверка каскада

**Вывод результатов:**

Все три сценария используют `handleSummary()` для красивого табличного вывода:
```
┌─────────────────────────────────────────┐
│         REGISTRATION LOAD TEST          │
├─────────────────────────────────────────┤
│ Registrations                           │
│   Success:  2987                        │
│   Errors:   13                          │
│   Latency p95: 423ms                    │
├─────────────────────────────────────────┤
│ NATS Profile Creation                   │
│   Created:  2980                        │
│   Not ready: 7                          │
│   Latency p95: 1200ms                   │
└─────────────────────────────────────────┘
```

### Общая таблица: что, где, как

| Тип | Что тестирует | Требует инфраструктуру? | Как запускать |
|-----|---------------|------------------------|---------------|
| Unit | Domain-логику, валидацию, backoff | Нет | `cd services/<name> && go test ./... -v` |
| Интеграционные | Полный HTTP lifecycle (36 сценариев) | Да, `docker compose up` | `cd tests/integration && go test -v` |
| Нагрузочные | Throughput, latency, NATS delivery под нагрузкой | Да, `docker compose up` + k6 | `k6 run tests/load/<script>.js` |

## DLQ-утилита

CLI-инструмент для работы с Dead Letter Queue. Отдельный Go module (`tools/dlq/`).

```bash
# Показать все "мёртвые" сообщения
docker compose run --rm dlq list

# Переотправить их обратно в оригинальные стримы (после исправления бага)
docker compose run --rm dlq replay

# Очистить DLQ
docker compose run --rm dlq purge
```

**Workflow при инциденте**:
1. Grafana дашборд показывает ненулевой DLQ → алерт
2. `dlq list` — смотрим что за сообщения, какая ошибка
3. Исправляем баг, деплоим фикс
4. `dlq replay` — переотправляем сообщения, consumer обработает заново
5. `dlq purge` — чистим DLQ

## Shared пакет `pkg/`

Минимальный пакет с тем, что должны знать несколько сервисов:

```
pkg/
├── events/
│   └── user.go     — UserCreated, UserDeleted, DLQEnvelope, субъекты, имена стримов
└── postgres/
    └── helpers.go  — утилиты для PostgreSQL
```

**Правило: `pkg/` должен быть минимальным.** Только shared contracts (события, proto). Middleware, логгеры, утилиты — internal в каждом сервисе. Иначе pkg превращается в "помойку" и ломает границы сервисов.

## Postman

Коллекция `postman/Yammi_API_Gateway.postman_collection.json` содержит все HTTP-эндпоинты с примерами запросов и тестовыми скриптами для автоматического сохранения токенов в переменные.
