# Notification Service

> Асинхронный сервис уведомлений. Потребляет события из NATS, создаёт уведомления для пользователей, отдаёт их по gRPC через API Gateway.

---

## Обзор

Notification Service — **event-driven** сервис, который:

1. **Слушает NATS JetStream** — 13 типов событий (board/column/card/member/user)
2. **Создаёт уведомления** — каждому участнику доски (кроме автора действия)
3. **Отдаёт по gRPC** — список, пометка прочитанным, счётчик непрочитанных, настройки
4. **Публикует в NATS** — событие `notification.created` для WebSocket Gateway (real-time push)

```
NATS JetStream                                    API Gateway
    │                                                  │
    ▼                                                  ▼
┌──────────────────────────┐              ┌──────────────────┐
│  NATS Consumer           │              │  gRPC Handler    │
│  (13 event subscribers)  │              │  (6 RPC methods) │
│           │               │              │        │         │
│           ▼               │              │        ▼         │
│  CreateNotification UC   │              │  List/MarkRead/  │
│           │               │              │  UnreadCount/    │
│           ▼               │              │  Settings        │
│  PostgreSQL + NATS pub   │              │        │         │
└──────────────────────────┘              └────────┼─────────┘
                                                   ▼
                                              PostgreSQL
```

---

## Технологии

| Компонент | Технология |
|-----------|------------|
| Язык | Go 1.24 |
| DB Driver | pgx/v5 (нативная поддержка PgBouncer transaction mode) |
| Sync API | gRPC + Protocol Buffers |
| Async | NATS JetStream (consumer + publisher) |
| БД | PostgreSQL 16 через PgBouncer (`yammi_notification`) |
| Поиск | PostgreSQL GIN trigram (`pg_trgm`) |
| Пагинация | Cursor-based (`created_at` timestamp) |
| Метрики | Prometheus (`/metrics` на порту 2112) |

### Оптимизации производительности

- **Batch INSERT** — `BatchCreate` для fan-out: 1 multi-row INSERT вместо N отдельных
- **In-memory settings cache** — `sync.RWMutex` + `map`, мгновенная инвалидация через NATS event `notification.settings.updated`
- **Batch settings check** — `BatchGet` с `WHERE user_id = ANY($1)` вместо N SELECT
- **gRPC panic recovery** — interceptor ловит panic от transient DB ошибок, возвращает controlled 500
- **DB retry** — автоматический retry при connection reset/broken pipe от PgBouncer

---

## Domain Model

### Notification

```go
type Notification struct {
    ID        string              // UUID, генерируется при создании
    UserID    string              // Кому адресовано
    Type      NotificationType    // Тип (13 вариантов)
    Title     string              // Заголовок, max 250 символов
    Message   string              // Подробности (может быть пустым)
    Metadata  map[string]string   // Дополнительные данные (board_id, card_title и т.д.)
    IsRead    bool                // Прочитано ли
    CreatedAt time.Time           // Время создания
}
```

### 13 типов уведомлений

| Тип | Когда создаётся | Кому |
|-----|-----------------|------|
| `welcome` | Новый пользователь зарегистрировался | Самому пользователю |
| `board_created` | Создана доска | Владельцу |
| `board_updated` | Обновлена доска | Всем участникам (кроме автора) |
| `board_deleted` | Удалена доска | Всем участникам (кроме автора) |
| `column_created` | Создана колонка | Всем участникам (кроме автора) |
| `column_updated` | Обновлена колонка | Всем участникам (кроме автора) |
| `column_deleted` | Удалена колонка | Всем участникам (кроме автора) |
| `card_created` | Создана карточка | Всем участникам (кроме автора) |
| `card_updated` | Обновлена карточка | Всем участникам (кроме автора) |
| `card_moved` | Перемещена карточка | Всем участникам (кроме автора) |
| `card_deleted` | Удалена карточка | Всем участникам (кроме автора) |
| `member_added` | Добавлен участник | Добавленному пользователю |
| `member_removed` | Удалён участник | Удалённому пользователю |

### NotificationSettings

```go
type NotificationSettings struct {
    UserID          string  // UUID пользователя
    Enabled         bool    // Глобальный переключатель (default: true)
    RealtimeEnabled bool    // Push через WebSocket (default: true)
}
```

Если `Enabled = false` — уведомления не создаются вообще. Если `RealtimeEnabled = false` — уведомления создаются в БД, но не пушатся в WebSocket.

### Domain Errors

| Ошибка | Когда |
|--------|-------|
| `ErrNotificationNotFound` | Не найдено по ID |
| `ErrEmptyUserID` | Пустой user_id |
| `ErrEmptyTitle` | Пустой заголовок |
| `ErrEmptyType` | Пустой тип |

---

## NATS Event Consumers

### Как работает обработка событий

```
NATS JetStream → Subscribe (Durable, DeliverNew)
                      │
                      ▼
              Unmarshal JSON event
                      │
            ┌─────────┼─────────┐
            │ Ошибка парсинга   │ Успешно
            │         │         │
            ▼         │         ▼
         DLQ          │  handleWithRetry()
                      │         │
                      │    ┌────┼────┐
                      │    │ Ошибка  │ Успех
                      │    │    │    │
                      │    ▼    │    ▼
                      │ Retry   │   Ack
                      │ (NakWithDelay)
                      │    │
                      │    ▼ (после 7 попыток)
                      │   DLQ
```

### Конфигурация consumers

| Параметр | Значение | Зачем |
|----------|----------|-------|
| `MaxDeliver` | 7 | Максимум ретраев до DLQ |
| `MaxAckPending` | 500 | Максимум неподтверждённых сообщений |
| `AckWait` | 30s | Таймаут до реdelivery |
| Backoff | Exponential (1s → 30s) | С jitter ±20% |

### Consumer versioning

Имена consumers: `notification-service-<event>-v2`. Суффикс `v2` позволяет пересоздать consumer с новой логикой — NATS создаст новый consumer с `DeliverNew`, не переигрывая старые события.

### Кеширование имён

Для формирования человекочитаемых заголовков (например, `"Карточка 'Задача' перемещена в доске 'Спринт 1'"`) сервис поддерживает **локальный кеш имён** в PostgreSQL:

| Таблица | Что кешируется | Как заполняется |
|---------|----------------|-----------------|
| `board_names` | Названия досок | `board.created`, `board.updated` |
| `user_names` | Имена пользователей | `user.created` |
| `card_names` | Названия карточек | `card.created`, `card.updated` |
| `column_names` | Названия колонок | `column.created`, `column.updated` |

Кеш заполняется **при старте** сервиса через отдельные subscribers с `DeliverAll` (проигрывают все старые события).

### Кеш участников досок

Таблица `board_members` — локальная копия участников досок для маршрутизации уведомлений. Обновляется из событий `board.created`, `member.added`, `member.removed`, `board.deleted`. Позволяет узнать кому отправить уведомление **без gRPC-вызова** в Board Service.

---

## gRPC API

Proto: `services/notification/api/proto/v1/notification.proto`

### ListNotifications

Список уведомлений пользователя с cursor-пагинацией.

```protobuf
rpc ListNotifications(ListNotificationsRequest) returns (ListNotificationsResponse)
```

| Параметр | Тип | Описание |
|----------|-----|----------|
| `user_id` | string | UUID пользователя (из JWT) |
| `limit` | int32 | Количество (default 20, max 100) |
| `cursor` | string | Курсор для пагинации (RFC3339Nano) |
| `type_filter` | string | Фильтр по типу (префиксный: `"card"` → все card_*) |
| `search` | string | Поиск по заголовку (ILIKE) |

Ответ содержит `next_cursor` (пустой если конец) и `total_unread`.

### MarkAsRead / MarkAllAsRead

```protobuf
rpc MarkAsRead(MarkAsReadRequest) returns (MarkAsReadResponse)
rpc MarkAllAsRead(MarkAllAsReadRequest) returns (MarkAllAsReadResponse)
```

### GetUnreadCount

```protobuf
rpc GetUnreadCount(GetUnreadCountRequest) returns (GetUnreadCountResponse)
```

### GetSettings / UpdateSettings

```protobuf
rpc GetSettings(GetSettingsRequest) returns (GetSettingsResponse)
rpc UpdateSettings(UpdateSettingsRequest) returns (UpdateSettingsResponse)
```

---

## Схема базы данных

БД: `yammi_notification`

### Таблица notifications

```sql
CREATE TABLE notifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL,
    type       VARCHAR(50) NOT NULL,
    title      VARCHAR(500) NOT NULL,
    message    TEXT DEFAULT '',
    metadata   JSONB DEFAULT '{}',
    is_read    BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Индексы:**

| Индекс | Колонки | Назначение |
|--------|---------|------------|
| `idx_notifications_user_created` | `(user_id, created_at DESC)` | Основной запрос — список уведомлений |
| `idx_notifications_user_unread` | `(user_id) WHERE is_read = FALSE` | Partial index для подсчёта непрочитанных |
| `idx_notifications_user_type` | `(user_id, type, created_at DESC)` | Фильтрация по типу |
| `idx_notifications_search` | `GIN(title gin_trgm_ops)` | Полнотекстовый поиск (pg_trgm) |

### Таблица notification_settings

```sql
CREATE TABLE notification_settings (
    user_id          UUID PRIMARY KEY,
    enabled          BOOLEAN DEFAULT true,
    realtime_enabled BOOLEAN DEFAULT true,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);
```

### Таблицы кеша

```sql
-- Локальный кеш участников досок
CREATE TABLE board_members (
    board_id UUID NOT NULL,
    user_id  UUID NOT NULL,
    PRIMARY KEY (board_id, user_id)
);

-- Кеш имён
CREATE TABLE board_names  (board_id  UUID PRIMARY KEY, title VARCHAR(500));
CREATE TABLE user_names   (user_id   UUID PRIMARY KEY, name  VARCHAR(255));
CREATE TABLE card_names   (card_id   UUID PRIMARY KEY, title VARCHAR(500));
CREATE TABLE column_names (column_id UUID PRIMARY KEY, title VARCHAR(500));
```

---

## Сборка и запуск

### Docker Compose (рекомендуемый)

```bash
docker compose up --build notification
```

Порты:
- `50055` — gRPC
- `2112` — Prometheus metrics (`/metrics`)

### Переменные окружения

| Переменная | Значение | Описание |
|------------|----------|----------|
| `NOTIFICATION_GRPC_PORT` | `50055` | Порт gRPC сервера |
| `METRICS_PORT` | `2112` | Порт HTTP сервера для метрик |
| `DATABASE_URL` | `postgres://...` | Подключение к PostgreSQL |
| `NATS_URL` | `nats://nats:4222` | Подключение к NATS |
| `MIGRATIONS_DIR` | `/app/migrations` | Путь к SQL миграциям |

### Генерация protobuf

```bash
cd services/notification
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/v1/notification.proto
```

---

## Связанные документы

- [notification-events.md](./notification-events.md) — детальная карта событий NATS
- [monitoring.md](../infrastructure/monitoring.md) — Prometheus метрики и Grafana дашборд
- [NOTIFICATION_ROUTES.md](../api-gateway/NOTIFICATION_ROUTES.md) — HTTP endpoints
- [architecture.md](../architecture.md) — общая архитектура системы
