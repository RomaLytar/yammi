# Yammi

Trello-like task board. Go microservices, PostgreSQL, NATS JetStream, gRPC, WebSocket.

## Стек

| | Технология |
|--|-----------|
| Backend | Go, gRPC, Protocol Buffers |
| Frontend | Vue 3, TypeScript |
| БД | PostgreSQL 16, PgBouncer (transaction pooling) |
| Кеш | Redis 7 |
| Events | NATS JetStream |
| Monitoring | Prometheus, Grafana |
| Load testing | k6 |
| Контейнеры | Docker Compose |

## Архитектура

```
Client → Frontend (:3000) → API Gateway (:8080) → gRPC → Microservices
                                                            │
                                                      NATS JetStream
                                                       │          │
                                              Notification    WebSocket Gateway (:8081) → Client
```

7 микросервисов, каждый со своей PostgreSQL базой, clean architecture, lightweight DDD.

## Makefile команды

```bash
# Поднять всё (бек, фронт, графана, сокеты)
make up

# Остановить всё
make down

# Перезапустить
make restart

# Нагрузочный тест (FILE обязателен)
make test FILE=realistic_1000_users.js

# Тест с повышенным rate limit (все эндпоинты)
make test FILE=realistic_1000_users.js LIMIT=300000

# Тест + очистка данных после
make test FILE=realistic_1000_users.js CLEAN=1

# Все параметры вместе
make test FILE=realistic_1000_users.js LIMIT=300000 CLEAN=1 REDIS_PASS=my_secret

# Очистка всех БД и Redis (без теста)
make clean

# Очистка с кастомным паролем Redis
make clean REDIS_PASS=my_secret
```

Доступные тест-файлы: `tests/load/` — `realistic_1000_users.js`, `realistic_2000_users.js`, `burst_3000_users.js`, `ceiling_test.js` и др.

## Ссылки

| | URL |
|--|-----|
| Frontend | http://localhost:3000 |
| API | http://localhost:8080 |
| Grafana | http://localhost:3033 |
| Prometheus | http://localhost:9090 |

## Нагрузочное тестирование (k6, 1000 VU)

Реалистичный сценарий: 70% workers / 20% readers / 10% heavy users, 3 минуты, 27 прогонов оптимизации.

### Лучший API — прогон #8 (1-3 members/board)

pgx + PgBouncer txn + async fire-and-forget + дедупликация IsMember (3→1)

| Метрика (p95) | Baseline (#1) | Best (#8) | Δ |
|---------------|---------------|-----------|---|
| Create board | 135ms | **74ms** | -45% |
| Add member | 378ms | **191ms** | -49% |
| Create column | 296ms | **103ms** | -65% |
| Create card | 354ms | **142ms** | -60% |
| Move card | 598ms | **223ms** | -63% |
| Notifications GET | 36ms | **24ms** | -33% |
| Notif delivery | 14.9s | **8.1s** | -46% |
| Duration | 324ms | **135ms** | -58% |
| Throughput | 157k | **170k** | +8% |
| Errors | 0% | **0%** | — |

### Лучший с большими досками — прогон #26 (20 members/board)

3 consumer instances + split PgBouncer (API 30 + Consumer 30) + in-memory name cache + Redis lazy cache

| Метрика (p95) | Baseline 20m (#11) | Best 20m (#26) | Δ |
|---------------|--------------------|----|---|
| Create board | 237ms | **238ms** | — |
| Add member | 592ms | **664ms** | — |
| Create column | 360ms | — | — |
| Create card | 491ms | **553ms** | — |
| Move card | 704ms | **817ms** | — |
| Notifications GET | 47ms | **66ms** | — |
| Notif delivery | 20.8s | **8.6s** | **-59%** |
| Duration | 557ms | **626ms** | — |
| Throughput | 198k | **194k** | — |
| Errors | 0% | **0%** | — |

API latency на уровне baseline — основной выигрыш в delivery: consumer groups распараллеливают обработку NATS событий.

### Текущий — прогон #27 (split API + Consumer)

notification-api (gRPC) и notification-consumer (NATS) разделены в отдельные процессы.

| Метрика (p95) | Best (#26) | Current (#27) | Δ |
|---------------|------------|---------------|---|
| Create board | 238ms | 282ms | +18% |
| Add member | 664ms | 728ms | +10% |
| Create card | 553ms | 600ms | +8% |
| Move card | 817ms | 884ms | +8% |
| Notifications GET | 66ms | 72ms | +9% |
| Notif delivery | 8.6s | 14.3s | +66% |
| Duration | 626ms | 688ms | +10% |
| Throughput | 194k | 184k | -5% |
| Errors | 0% | 0% | — |

> **Почему медленнее:** split на отдельные процессы — PgBouncer connection churn, отдельные NATS-соединения на каждый consumer instance, overhead отдельных GC/scheduler.
>
> **Зачем:** независимое масштабирование API и Consumer (разный CPU/memory профиль), изоляция отказов (падение consumer не ломает gRPC), независимый деплой. Tradeoff оправдан в production при 5+ consumer instances.

### Сводная таблица

| Метрика (p95) | #1 Baseline | #8 Best API | #26 Best 20m | #27 Current |
|---------------|-------------|-------------|-------------|-------------|
| | *1-3 members* | *1-3 members* | *20 members* | *20 members* |
| Create board | 135ms | **74ms** | 238ms | 282ms |
| Add member | 378ms | **191ms** | 664ms | 728ms |
| Create column | 296ms | **103ms** | 360ms | — |
| Create card | 354ms | **142ms** | 553ms | 600ms |
| Move card | 598ms | **223ms** | 817ms | 884ms |
| Notifications GET | 36ms | **24ms** | 66ms | 72ms |
| Notif delivery | 14.9s | 8.1s | **8.6s** | 14.3s |
| Duration | 324ms | **135ms** | 626ms | 688ms |
| Throughput | 157k | **170k** | 194k | 184k |
| Errors | 0% | 0% | 0% | 0% |
