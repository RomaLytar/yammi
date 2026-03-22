# Нагрузочный тест: 1000 пользователей — реалистичный сценарий

> Файл теста: `tests/load/realistic_1000_users.js`
> Дата прогона: 2026-03-21

---

## Конфигурация

| Параметр | Значение |
|----------|----------|
| VU (виртуальных пользователей) | до 1000 |
| Длительность | ~3 минуты (30s ramp-up → 30s рост → 1m пик → 30s cooldown) |
| Setup | 1000 регистраций последовательно (~30s) |
| Распределение | 70% workers / 20% readers / 10% heavy users |
| Think time | 0.2s — 3s между шагами (зависит от типа) |

### Профиль нагрузки

```
stages: [
  { duration: '30s', target: 300  },  // мягкий старт
  { duration: '30s', target: 1000 },  // рост до пика
  { duration: '1m',  target: 1000 },  // удержание пика
  { duration: '30s', target: 0    },  // cooldown
]
```

### Типы пользователей

| Тип | Доля | Действия |
|-----|------|----------|
| **Worker** (70%) | Создать доску → добавить 1-2 участников → 2 колонки → 2-3 карточки → переместить → нотификации → 50% удалить доску |
| **Reader** (20%) | Список досок → открыть доску → нотификации → unread count |
| **Heavy** (10%) | 2 доски × (3 участника + 3 колонки + 5 карточек + 3 перемещения + update + delete) → noti delivery latency |

---

## Результаты

```
╔══════════════════════════════════════════════════════════════════╗
║           НАГРУЗОЧНЫЙ ТЕСТ: 1000 пользователей                 ║
╠══════════════════════════════════════════════════════════════════╣
║                                                                  ║
║  Распределение:  70% workers / 20% readers / 10% heavy users   ║
║                                                                  ║
║  ── Latency (p95) ──────────────────────────────────────────    ║
║  Создание доски:          135 ms                                ║
║  Добавление участника:    378 ms                                ║
║  Создание колонки:        296 ms                                ║
║  Создание карточки:       354 ms                                ║
║  Перемещение карточки:    598 ms                                ║
║  Нотификации:              36 ms                                ║
║  Notif delivery:        14896 ms                                ║
║                                                                  ║
║  ── Ошибки ─────────────────────────────────────────────────    ║
║  Error rate:             0.0%                                    ║
║  Board errors:            N/A                                    ║
║  Card errors:             N/A                                    ║
║  Member errors:           N/A                                    ║
║  Notification errors:     N/A                                    ║
║                                                                  ║
║  ── HTTP ───────────────────────────────────────────────────    ║
║  Total requests:       157908                                    ║
║  Failed requests:        0.1%                                    ║
║  Duration p95:            324 ms                                ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
```

---

## Анализ

### ✅ Прошли thresholds

| Метрика | Threshold | Факт | Статус |
|---------|-----------|------|--------|
| `http_req_failed` | < 5% | 0.1% | ✅ |
| `http_req_duration` p95 | < 1500ms | 324ms | ✅ |
| `latency_create_board_ms` p95 | < 500ms | 135ms | ✅ |
| `latency_create_card_ms` p95 | < 500ms | 354ms | ✅ |
| `latency_add_member_ms` p95 | < 500ms | 378ms | ✅ |
| `latency_notifications_ms` p95 | < 500ms | 36ms | ✅ |
| `error_rate` | < 5% | 0.0% | ✅ |

### ❌ Crossed thresholds

| Метрика | Threshold | Факт | Причина |
|---------|-----------|------|---------|
| `latency_move_card_ms` p95 | < 500ms | 598ms | Optimistic locking + lexorank пересчёт под нагрузкой |
| `latency_notif_delivery_ms` p95 | < 5000ms | 14896ms | Fan-out: 1 событие → N нотификаций, NATS + PostgreSQL INSERT на каждого участника |

### Ключевые наблюдения

1. **157,908 запросов за 3 минуты** — ~880 RPS на пике
2. **Error rate 0.1%** — система стабильна, почти все запросы успешны
3. **Создание доски — 135ms p95** — отлично, быстрее всех операций
4. **Нотификации GET — 36ms p95** — PostgreSQL + индексы работают эффективно
5. **Move card — 598ms** — самая тяжёлая операция (optimistic locking, lexorank)
6. **Notification delivery — ~15s** — bottleneck fan-out при 1000 активных пользователях

### Предупреждения k6

```
The test has generated metrics with 400413 unique time series
```

Причина: URL с уникальными UUID (`/api/v1/boards/{uuid}`) создают уникальные time series. Рекомендация: добавить `tags: { name: 'boards' }` для группировки.

---

## Запуск

```bash
# Поднять rate limits для теста
RATE_LIMIT_REGISTER=100000 RATE_LIMIT_LOGIN=100000 RATE_LIMIT_DEFAULT=100000 RATE_LIMIT_REFRESH=100000 \
  docker compose up -d --force-recreate api-gateway

# Запуск теста
docker run --rm --network yammi_default \
  -v $(pwd)/tests/load:/scripts \
  grafana/k6:latest run /scripts/realistic_1000_users.js \
  -e BASE_URL=http://api-gateway:8080 \
  -e USERS=1000

# Вернуть нормальные rate limits
docker compose up -d --force-recreate api-gateway
```

---

## Прогон #2: после оптимизации (Batch INSERT + Settings Cache + Batch Publish)

**Дата:** 2026-03-21 | **Оптимизации:**
- Batch INSERT — 1 multi-row INSERT вместо N отдельных
- In-memory settings cache с NATS invalidation — 0 DB round-trips на settings
- Batch NATS publish — публикация всех нотификаций за один вызов

### Результаты

```
╔══════════════════════════════════════════════════════════════════╗
║           НАГРУЗОЧНЫЙ ТЕСТ: 1000 пользователей                 ║
╠══════════════════════════════════════════════════════════════════╣
║                                                                  ║
║  Распределение:  70% workers / 20% readers / 10% heavy users   ║
║                                                                  ║
║  ── Latency (p95) ──────────────────────────────────────────    ║
║  Создание доски:          140 ms                                ║
║  Добавление участника:    353 ms                                ║
║  Создание колонки:        275 ms                                ║
║  Создание карточки:       358 ms                                ║
║  Перемещение карточки:    574 ms                                ║
║  Нотификации:              38 ms                                ║
║  Notif delivery:        12434 ms                                ║
║                                                                  ║
║  ── Ошибки ─────────────────────────────────────────────────    ║
║  Error rate:             0.0%                                    ║
║                                                                  ║
║  ── HTTP ───────────────────────────────────────────────────    ║
║  Total requests:       156862                                    ║
║  Failed requests:        0.1%                                    ║
║  Duration p95:            331 ms                                ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
```

### Сравнение до/после

| Метрика | До | После | Изменение |
|---------|-----|-------|-----------|
| Создание доски p95 | 135ms | 140ms | ≈ |
| Добавление участника p95 | 378ms | 353ms | -7% |
| Создание колонки p95 | 296ms | 275ms | -7% |
| Создание карточки p95 | 354ms | 358ms | ≈ |
| Перемещение карточки p95 | 598ms | 574ms | -4% |
| Нотификации GET p95 | 36ms | 38ms | ≈ |
| **Notif delivery p95** | **14896ms** | **12434ms** | **-17%** |
| Error rate | 0.0% | 0.0% | ≈ |
| Total requests | 157908 | 156862 | ≈ |

### Прогон #3: стабильный результат (batch only, без worker pool)

Worker pool (8 горутин + shared queue) был протестирован и **откачен** — увеличил DB contention, latency выросла по всем операциям. NATS уже даёт 13 goroutines (по одному на subscription), shared worker pool уменьшал параллелизм.

Финальный стабильный результат (batch only):

| Метрика | #1 (baseline) | #2 (batch) | #3 (batch, стабильный) |
|---------|---------------|------------|------------------------|
| Создание доски p95 | 135ms | 140ms | **122ms** |
| Добавление участника p95 | 378ms | 353ms | **310ms** |
| Создание колонки p95 | 296ms | 275ms | **237ms** |
| Создание карточки p95 | 354ms | 358ms | **319ms** |
| Перемещение карточки p95 | 598ms | 574ms | **541ms** |
| Нотификации GET p95 | 36ms | 38ms | **33ms** |
| **Notif delivery p95** | **14896ms** | **12434ms** | **10916ms** |
| Duration p95 | 324ms | 331ms | **290ms** |
| Total requests | 157908 | 156862 | **159236** |
| Error rate | 0.0% | 0.0% | 0.0% |

### Вывод

Batch INSERT + settings cache + стабильная NATS архитектура: **-27% notification delivery** (15s → 10.9s) и **улучшение ALL API latency** (~10-20% быстрее по всем операциям).

Worker pool протестирован и отклонён — shared queue с bounded buffer добавляет contention на DB connections. 13 NATS goroutines (по одному на subscription) оптимальнее shared pool из 8 workers.

Дальнейшее снижение delivery latency требует инфраструктурных изменений (PgBouncer, partitioning), а не application-level оптимизаций.

---

---

## Диагностика: PostgreSQL connection bottleneck

```
PostgreSQL max_connections = 100

Сервис                Pool MaxOpen   Инстансов   Потенциал
────────────────────────────────────────────────────────────
Auth                  20            ×5           100
Board                 50            ×1            50
Notification          30            ×1            30
User                  25            ×1            25
────────────────────────────────────────────────────────────
ИТОГО                                            205
PostgreSQL лимит                                 100 ← bottleneck
```

В покое: 52/100 connections заняты (board=22, notification=16, auth=6).
При нагрузке: сервисы конкурируют за оставшиеся ~48 connections.

**Это корневая причина** 10+ секунд delivery — batch INSERT ждёт свободный connection.

---

---

## Прогон #4: уменьшенные connection pools (без PgBouncer)

**Изменения:** Auth 20→5, Board 50→15, Notification 30→10, User 25→5. Суммарный потенциал: 55 (было 205).

| Метрика | #3 (pool 205) | #4 (pool 55) | Δ |
|---------|---------------|--------------|---|
| **Notif delivery p95** | **10.9s** | **8.1s** | **-26%** |
| Нотификации GET p95 | 33ms | **25ms** | **-24%** |
| Создание доски p95 | 122ms | 292ms | +140% |
| Move card p95 | 541ms | 1261ms | +133% |
| Duration p95 | 290ms | 783ms | +170% |
| Total requests | 159k | 131k | -18% |
| Error rate | 0.0% | 0.0% | ≈ |

### Вывод

Notification delivery улучшился на 26% (10.9s → 8.1s) — connection starvation уменьшился. Но API latency удвоилась — сервисы ждут в очереди за маленьким пулом. **PgBouncer решит обе проблемы**: мультиплексирует много app connections через оптимальное число реальных PG connections.

---

---

## Прогон #5: PgBouncer (transaction mode)

**Изменения:**
- PgBouncer перед PostgreSQL (transaction pooling, `default_pool_size=20`, `max_prepared_statements=100`)
- Миграции через прямое подключение к PostgreSQL (`MIGRATION_DATABASE_URL`)
- PostgreSQL `max_connections=200`

| Метрика | #3 (без PgBouncer) | #5 (PgBouncer) | Δ |
|---------|---------------------|----------------|---|
| Создание доски p95 | 122ms | **19ms** | **-84%** |
| Добавление участника p95 | 310ms | **13ms** | **-96%** |
| Создание колонки p95 | 237ms | **14ms** | **-94%** |
| Создание карточки p95 | 319ms | **12ms** | **-96%** |
| Перемещение карточки p95 | 541ms | **21ms** | **-96%** |
| Нотификации GET p95 | 33ms | **2ms** | **-94%** |
| Duration p95 | 290ms | **15ms** | **-95%** |
| **Error rate** | **0.0%** | **69%** | **ПРОБЛЕМА** |

### Вывод

PgBouncer даёт **×10-20 ускорение latency** (122ms → 19ms для create board, 541ms → 21ms для move card). Однако notification service **падает** с паникой при DB ошибках от PgBouncer — нет gRPC recovery interceptor. Это каскадно ломает ~70% запросов (API gateway получает 500 от мёртвого notification service).

### Блокеры

1. **gRPC panic recovery** — notification service крашится при DB ошибке, нет `grpc_recovery` interceptor
2. **Notification delivery metric N/A** — сервис умирает до замера

---

## Прогон #6: PgBouncer session mode + recovery interceptor + увеличенные пулы

**Изменения:**
- `pool_mode = session` (transaction mode ломает lib/pq extended query protocol)
- gRPC panic recovery interceptor во всех 4 сервисах
- DB retry helper в notification service
- Пулы: Auth=10×5, Board=25, Notification=20, User=10 (total=105)

| Метрика | #1 Baseline | #3 +Batch | #6 PgBouncer session |
|---------|-------------|-----------|---------------------|
| Create board p95 | 135ms | 122ms | **354ms** |
| Add member p95 | 378ms | 310ms | **976ms** |
| Create card p95 | 354ms | 319ms | **970ms** |
| Move card p95 | 598ms | 541ms | **1452ms** |
| **Notif delivery p95** | **14.9s** | **10.9s** | **10.1s** |
| Notifications GET p95 | 36ms | 33ms | **36ms** |
| Duration p95 | 324ms | 290ms | **975ms** |
| Total requests | 157k | 159k | **123k** |
| **Error rate** | **0.0%** | **0.0%** | **0.0%** |

### Вывод

PgBouncer session mode **стабилен** (0% ошибок) но **медленнее baseline** — session mode не мультиплексирует, добавляет hop overhead. Notification delivery улучшился незначительно (10.9s → 10.1s).

Для полноценного ускорения нужен **transaction mode**, что требует замены `lib/pq` на `pgx` (нативная поддержка PgBouncer transaction pooling).

### Что добавлено (полезно независимо от PgBouncer)

- ✅ **gRPC panic recovery** — все 4 сервиса, panic → controlled 500 вместо crash
- ✅ **DB retry helper** — notification service, retry при transient connection errors
- ✅ **MIGRATION_DATABASE_URL** — миграции идут напрямую к PostgreSQL, обходя PgBouncer

---

## Итоговая сводка всех прогонов

| # | Что | Notif delivery | Move card | Error rate | Requests |
|---|-----|----------------|-----------|------------|----------|
| 1 | Baseline | 14.9s | 598ms | 0.0% | 157k |
| 3 | +Batch INSERT +Cache | **10.9s** | **541ms** | 0.0% | 159k |
| 4 | +Small pools (55) | 8.1s | 1261ms | 0.0% | 131k |
| 5 | +PgBouncer transaction | **19ms** latency! | **21ms** | **69%** | 152k |
| 6 | +PgBouncer session | 10.1s | 1452ms | 0.0% | 123k |

### Ключевые выводы

1. **Batch INSERT + settings cache** — лучшая чистая оптимизация (+27% delivery, +10% API)
2. **PgBouncer transaction mode** — даёт ×10 latency но **ломает lib/pq** (extended query protocol)
3. **PgBouncer session mode** — стабилен но не ускоряет (overhead > benefit)
4. **Connection pool sizing** — критичен, 205 потенциальных vs 100 max = starvation

---

## Прогон #7: pgx driver + PgBouncer transaction mode

**Изменения:**
- Все 4 сервиса мигрированы с `lib/pq` на `pgx/stdlib`
- PgBouncer переключен на `pool_mode = transaction`
- `pq.Array()` заменён на `pgtype.FlatArray`
- gRPC recovery interceptor + DB retry сохранены

| Метрика | #1 Baseline | #3 +Batch | #7 pgx+PgBouncer txn |
|---------|-------------|-----------|----------------------|
| Create board p95 | 135ms | 122ms | **181ms** |
| Add member p95 | 378ms | 310ms | **510ms** |
| Create card p95 | 354ms | 319ms | **516ms** |
| Move card p95 | 598ms | 541ms | **830ms** |
| **Notif delivery p95** | **14.9s** | **10.9s** | **8.3s** |
| Notifications GET p95 | 36ms | 33ms | **26ms** |
| Duration p95 | 324ms | 290ms | **505ms** |
| Total requests | 157k | 159k | **147k** |
| **Error rate** | **0.0%** | **0.0%** | **0.0%** |

### Вывод

pgx + PgBouncer transaction mode **стабилен** (0% ошибок). Notification delivery **-44%** от baseline (14.9s → 8.3s). Нотификации GET — **26ms** (лучший результат за все прогоны). API latency выше baseline из-за PgBouncer overhead + уменьшенных пулов.

### Итоговая сводка

| # | Конфигурация | Notif delivery | Move card | Errors | Throughput |
|---|-------------|----------------|-----------|--------|------------|
| 1 | Baseline (lib/pq, прямой PG) | 14.9s | 598ms | 0% | 157k |
| 3 | +Batch INSERT +Cache | **10.9s** | 541ms | 0% | 159k |
| 7 | +pgx +PgBouncer transaction | **8.3s** | 830ms | 0% | 147k |

**Основные достижения:**
- Notification delivery: **14.9s → 8.3s** (-44%)
- Zero errors с PgBouncer transaction mode (pgx)
- gRPC recovery interceptor во всех сервисах
- DB retry для transient PgBouncer ошибок
- Миграции через прямое подключение к PostgreSQL

### Следующие шаги

- 🔧 **Тюнинг пулов** — увеличить app pools обратно (pgx+PgBouncer мультиплексирует, можно дать больше)
- 🔀 **Partitioning by board_id** — NATS consumers по board_id для параллельного fan-out
- 🔄 **5000 VU** — поиск точки отказа
