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

## Запуск

```bash
docker compose up --build
```

## Ссылки

| | URL |
|--|-----|
| Frontend | http://localhost:3000 |
| API | http://localhost:8080 |
| Grafana | http://localhost:3033 |
| Prometheus | http://localhost:9090 |

## Документация

[guides/INDEX.md](guides/INDEX.md) — полная документация проекта.
