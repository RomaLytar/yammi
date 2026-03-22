# 📚 Yammi — Документация

> **Главная точка входа** для навигации по всей документации проекта

---

## 🚀 Быстрый старт

### Для новых разработчиков

**Шаг 1: Запустите проект**
```bash
docker compose up --build
```
Подробнее: [CLAUDE.md](/home/roman/PetProject/yammi/CLAUDE.md) — команды сборки и правила

**Шаг 2: Изучите архитектуру**
1. [getting-started.md](./getting-started.md) — детальный гайд по проекту
2. [architecture.md](./architecture.md) — архитектура системы (микросервисы, clean architecture, DDD)

**Шаг 3: Выберите сервис**
- [Board Service](#-board-service) — доски, колонки, карточки (Trello-like)
- [Notification Service](#-notification-service) — уведомления (event-driven)
- [Auth Service](#-auth-service) — регистрация, авторизация, JWT
- [API Gateway](#-api-gateway) — HTTP entry point

---

## 🗂 Структура документации

```
guides/
│
├── INDEX.md                         # ← Вы здесь (главная точка входа)
├── getting-started.md               # Полный гайд по проекту
├── architecture.md                  # Общая архитектура системы
│
├── board/                           # 📋 Board Service
│   ├── board-service-build.md       # Сборка, тестирование, запуск
│   ├── board-service-proto.md       # Генерация gRPC кода
│   ├── lexorank-explained.md        # Алгоритм позиционирования карточек
│   ├── integration-tests.md         # Интеграционные тесты
│   └── tests-summary.md             # Сводка по тестам
│
├── notification/                     # 🔔 Notification Service
│   ├── notification-service.md      # Обзор, domain, NATS, gRPC, БД, сборка
│   └── notification-events.md       # Детальная карта NATS событий
│
├── api-gateway/                     # 🌐 API Gateway
│   ├── BOARD_ROUTES.md              # HTTP endpoints для Board Service
│   └── NOTIFICATION_ROUTES.md       # HTTP endpoints для Notification Service
│
├── infrastructure/                  # 🛠 Инфраструктура
│   ├── database-schema.md           # Схема БД (PostgreSQL)
│   └── monitoring.md                # Prometheus + Grafana (метрики, дашборды)
│
├── testing/                         # 🧪 Тестирование
│   ├── README.md                    # Unit, feature, integration тесты
│   └── load-test-realistic-1000-users.md  # Нагрузочный тест: 7 прогонов, оптимизации
│
├── decisions/                       # 📝 Архитектурные решения (ADR)
│   ├── 002-members-separate-table.md
│   ├── 004-lexorank-positions.md
│   └── 005-pgbouncer.md
│
└── auth/                            # 🔐 Auth Service (TODO)
    └── (будущая документация)
```

---

## 📋 Board Service

**Описание:** Основной сервис для управления досками, колонками и карточками (как Trello)

### Документация

| Документ | Описание |
|----------|----------|
| [board-service-build.md](./board/board-service-build.md) | Как собрать, протестировать и запустить Board Service |
| [board-service-proto.md](./board/board-service-proto.md) | Генерация gRPC кода из proto файлов |
| [lexorank-explained.md](./board/lexorank-explained.md) | Алгоритм lexorank для O(1) reordering карточек |
| [integration-tests.md](./board/integration-tests.md) | Как запустить интеграционные тесты (testcontainers) |
| [tests-summary.md](./board/tests-summary.md) | Сводка по всем тестам (223 тестов: 111 domain + 47 usecase + 65 integration) |
| [testing/README.md](./testing/README.md) | Полное руководство по тестированию Board Service |

### Правила кодирования

[services/board/README.md](/home/roman/PetProject/yammi/services/board/README.md) — соглашения, архитектурные правила, что НЕЛЬЗЯ делать

### Ключевые технологии

- **PostgreSQL** — HASH partitioning для cards (4 партиции)
- **Lexorank** — string-based позиционирование ("a", "am", "b")
- **Cursor pagination** — вместо OFFSET (для highload)
- **Optimistic locking** — version field для конкурентных обновлений
- **Clean Architecture** — domain → usecase → delivery → repository

### Highload решения

- ✅ Micro-aggregates (Board, Column, Card — отдельно, не загружаем всё сразу)
- ✅ IsMember() query вместо загрузки всех 100+ members
- ✅ HASH partitioning для cards (миллионы записей)
- ✅ Lexorank вместо INT position (1 UPDATE вместо N)
- ✅ Cursor pagination (стабильная latency на больших offset)

---

## 🔔 Notification Service

**Описание:** Асинхронный event-driven сервис уведомлений. Слушает NATS, создаёт уведомления участникам досок, отдаёт по gRPC.

### Документация

| Документ | Описание |
|----------|----------|
| [notification-service.md](./notification/notification-service.md) | Обзор: domain model, NATS consumers, gRPC API, схема БД, сборка |
| [notification-events.md](./notification/notification-events.md) | Детальная карта всех 13 NATS событий + DLQ + retry |
| [NOTIFICATION_ROUTES.md](./api-gateway/NOTIFICATION_ROUTES.md) | 6 HTTP endpoints (список, прочитать, настройки) |
| [monitoring.md](./infrastructure/monitoring.md) | Prometheus метрики + Grafana дашборд |

### Ключевые технологии

- **NATS JetStream** — 13 event consumers (board/column/card/member/user)
- **PostgreSQL** — уведомления + кеш имён + кеш участников
- **Prometheus** — 9 метрик (counters, histograms)
- **Cursor pagination** — по `created_at` timestamp
- **GIN trigram** — полнотекстовый поиск по заголовкам (`pg_trgm`)

### Ключевые паттерны

- ✅ Event-driven: уведомления создаются только из NATS событий
- ✅ Локальный кеш: имена досок/карточек/пользователей без gRPC-вызовов
- ✅ Retry + DLQ: 7 попыток с exponential backoff, затем dead letter queue
- ✅ Settings per user: глобальный переключатель + realtime toggle

---

## 🌐 API Gateway

**Описание:** HTTP entry point, JWT верификация, rate limiting

### Документация

| Документ | Описание |
|----------|----------|
| [BOARD_ROUTES.md](./api-gateway/BOARD_ROUTES.md) | 23 HTTP endpoints для Board Service (REST API) |
| [NOTIFICATION_ROUTES.md](./api-gateway/NOTIFICATION_ROUTES.md) | 6 HTTP endpoints для Notification Service |

### Правила кодирования

[services/api-gateway/README.md](/home/roman/PetProject/yammi/services/api-gateway/README.md)

### Ключевые фичи

- ✅ JWT verification (public key от Auth Service)
- ✅ Rate limiting (защита от перегрузки)
- ✅ gRPC → REST маппинг
- ✅ Centralized error handling

---

## 🔐 Auth Service

**Описание:** Регистрация, логин, JWT (EdDSA asymmetric keys), refresh/revoke tokens

### Документация

*(В разработке: создайте `guides/auth/` для документации Auth Service)*

### Правила кодирования

[services/auth/README.md](/home/roman/PetProject/yammi/services/auth/README.md)

---

## 🧪 Тестирование

### Документация

| Документ | Описание |
|----------|----------|
| [testing/README.md](./testing/README.md) | Unit, feature, integration тесты (223 теста, 95.9% domain coverage) |
| [load-test-realistic-1000-users.md](./testing/load-test-realistic-1000-users.md) | Нагрузочный тест: 1000 VU, 7 прогонов оптимизации, сводная таблица |

### Нагрузочные тесты (k6)

| Тест | Что делает |
|------|-----------|
| `tests/load/realistic_1000_users.js` | 1000 VU: регистрация → доски → колонки → карточки → нотификации (3 мин) |
| `tests/load/register_3000_users.js` | Burst регистрация 3000 юзеров, проверка async profile creation |
| `tests/load/burst_3000_users.js` | 3000 concurrent VU, stress test |
| `tests/load/delete_1000_users.js` | Batch удаление 1000 юзеров, проверка cascade |

### Ключевые результаты (1000 VU)

| Метрика | Baseline | После оптимизации |
|---------|----------|-------------------|
| Notification delivery p95 | 14.9s | **8.3s** (-44%) |
| Create board p95 | 135ms | 181ms |
| Error rate | 0.0% | 0.0% |

---

## 🛠 Инфраструктура

### Документация

| Документ | Описание |
|----------|----------|
| [database-schema.md](./infrastructure/database-schema.md) | Схема БД Board Service с объяснением ПОЧЕМУ для каждого решения |
| [monitoring.md](./infrastructure/monitoring.md) | Prometheus + Grafana: метрики, scrape targets, дашборды, правила создания |

### Компоненты

- **PostgreSQL 16** — основная БД (каждый сервис — своя база, `max_connections=200`)
- **PgBouncer** — connection pooling (transaction mode, `default_pool_size=20`)
- **Redis 7** — кеш (Board Service)
- **NATS 2** — event bus (JetStream, 4 streams: USERS, BOARDS, NOTIFICATIONS, DLQ)
- **Prometheus** — метрики (3 scrape targets: nats, board, notification)
- **Grafana** — 4 дашборда (Notification Service, Board Service, NATS JetStream, NATS User Deleted)

### Ключевые оптимизации

- ✅ **pgx driver** — нативная поддержка PgBouncer transaction mode (вместо lib/pq)
- ✅ **Batch INSERT** — multi-row INSERT для notification fan-out (1 запрос вместо N)
- ✅ **In-memory settings cache** — с мгновенной NATS инвалидацией
- ✅ **gRPC recovery interceptor** — panic → controlled error вместо crash
- ✅ **DB retry helper** — retry при transient PgBouncer ошибках
- ✅ **MIGRATION_DATABASE_URL** — миграции напрямую к PostgreSQL, обходя PgBouncer

---

## 📝 Архитектурные решения (ADR)

**Описание:** Документированные архитектурные решения с обоснованием ПОЧЕМУ

| ADR | Решение | ПОЧЕМУ |
|-----|---------|--------|
| [002-members-separate-table.md](./decisions/002-members-separate-table.md) | Members в отдельной таблице | 100+ members в Board aggregate → тяжелый payload, медленный GetBoard |
| [004-lexorank-positions.md](./decisions/004-lexorank-positions.md) | Lexorank вместо INT position | INT → массовые UPDATE при reorder, race conditions |
| [005-pgbouncer.md](./decisions/005-pgbouncer.md) | PgBouncer для connection pooling | 5 replicas × 50 connections = 250 > max_connections (200) |
| [006-pgx-driver.md](./decisions/006-pgx-driver.md) | Миграция lib/pq → pgx | lib/pq несовместим с PgBouncer transaction mode (prepared statements) |

---

## 📖 Общие концепции

### Clean Architecture

Все сервисы следуют одной структуре:

```
internal/
├── domain/           # Entities, business rules, errors (ZERO external deps)
├── usecase/          # Orchestration, authorization, interfaces
├── delivery/         # gRPC/HTTP handlers
├── repository/       # Реализация интерфейсов (PostgreSQL)
└── infrastructure/   # DB connection, JWT, migrations, cache, queue
```

**Правило:** Domain не зависит ни от чего. Usecase определяет интерфейсы. Infrastructure реализует.

### DDD: Micro-Aggregates

**Board Service использует НЕ традиционный DDD:**
- Board, Column, Card — **отдельные aggregates** (не загружаем всё сразу)
- IsMember() query вместо загрузки массива members
- Инварианты: optimistic locking (version), lexorank ordering

**Почему не традиционный DDD:** см. [decisions/002-members-separate-table.md](./decisions/002-members-separate-table.md)

### Event-Driven Architecture

- **Sync:** gRPC (Request/Response)
- **Async:** NATS JetStream (Events)
- **4 streams:** USERS, BOARDS, NOTIFICATIONS, DLQ
- **13 типов событий:** board/column/card/member CRUD + user.created
- **Consumers:** Notification Service (13 consumers), WebSocket Gateway
- **Retry:** exponential backoff (7 попыток), затем DLQ
- **Инвалидация кеша:** notification.settings.updated через NATS

### Database

- **pgx driver** — все сервисы используют `pgx/v5/stdlib` (не `lib/pq`)
- **PgBouncer** — transaction pooling между сервисами и PostgreSQL
- **Миграции** — напрямую к PostgreSQL через `MIGRATION_DATABASE_URL` (обход PgBouncer)

---

## 🎯 Как читать документацию

### День 1 (новичок в проекте)

1. [CLAUDE.md](/home/roman/PetProject/yammi/CLAUDE.md) — запустите проект
2. [architecture.md](./architecture.md) — общая картина
3. [getting-started.md](./getting-started.md) — детальный разбор

### День 2-3 (изучаем Board Service)

1. [board/board-service-build.md](./board/board-service-build.md) — сборка и запуск
2. [infrastructure/database-schema.md](./infrastructure/database-schema.md) — схема БД
3. [board/lexorank-explained.md](./board/lexorank-explained.md) — алгоритм позиционирования
4. [services/board/README.md](/home/roman/PetProject/yammi/services/board/README.md) — правила кодирования

### Работаете с Board Service

**Обязательно прочитать:**
1. [services/board/README.md](/home/roman/PetProject/yammi/services/board/README.md) — правила кодирования
2. [infrastructure/database-schema.md](./infrastructure/database-schema.md) — схема БД
3. [board/lexorank-explained.md](./board/lexorank-explained.md) — алгоритм

**Для разработки:**
1. [board/board-service-build.md](./board/board-service-build.md) — команды сборки
2. [board/board-service-proto.md](./board/board-service-proto.md) — генерация proto

---

## 🔍 Соглашения

### Формат документации

- **Markdown** — вся документация в `.md` файлах
- **Русский язык** — guides/, README.md (кроме CLAUDE.md)
- **Английский язык** — CLAUDE.md (для Claude Code)
- **Абсолютные пути** — ссылки на файлы вне guides/ (например, `/home/roman/PetProject/yammi/services/board/README.md`)
- **Относительные пути** — ссылки на файлы внутри guides/ (например, `./board/lexorank-explained.md`)

### Правила размещения

| Тип документа | Где хранить |
|---------------|-------------|
| Правила кодирования сервиса | `services/<service>/README.md` |
| Общая документация проекта | `guides/` |
| Документация конкретного сервиса | `guides/<service>/` |
| Архитектурные решения | `guides/decisions/` |
| Инфраструктура (БД, Redis, NATS) | `guides/infrastructure/` |

---

## ❓ Помощь

**Если вы не нашли нужную информацию:**

1. Проверьте [getting-started.md](./getting-started.md) — самый детальный документ
2. Проверьте README.md конкретного сервиса (`services/<service>/README.md`)
3. Ищите по ключевым словам в этом файле (INDEX.md)
4. Проверьте [architecture.md](./architecture.md) — общая архитектура
5. Проверьте ADR в [decisions/](./decisions/) — архитектурные решения

---

## 🚦 Статус сервисов

| Сервис | Статус | Документация |
|--------|--------|--------------|
| Auth Service | ✅ Реализован | [services/auth/README.md](/home/roman/PetProject/yammi/services/auth/README.md) |
| User Service | ✅ Реализован | [services/user/README.md](/home/roman/PetProject/yammi/services/user/README.md) |
| Board Service | ✅ Реализован | [guides/board/](#-board-service) |
| Comment Service | 🚧 В разработке | - |
| Notification Service | ✅ Реализован | [guides/notification/](#-notification-service) |
| WebSocket Gateway | ✅ Реализован | - |
| API Gateway | ✅ Реализован | [guides/api-gateway/](#-api-gateway) |
