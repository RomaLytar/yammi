# 005. PgBouncer для connection pooling

**Статус:** ✅ Принято (запланировано к реализации)

**Дата:** 2024-03-20

---

## Контекст

**Требования:** 100k concurrent users, 10k создают доски одновременно.

**Проблема:** PostgreSQL connection pool.

**Текущая конфигурация:**
```
Board Service (5 replicas) × 50 connections = 250 connections к PostgreSQL
```

**Ограничение PostgreSQL:**
```
max_connections = 200 (default)
```

**Проблема:** 250 > 200 → ERROR: too many connections

---

## Решение

**PgBouncer** — connection pooler между приложением и PostgreSQL.

**Архитектура:**
```
Board Service (5 replicas)
  ├─ Replica 1: 50 connections → PgBouncer
  ├─ Replica 2: 50 connections → PgBouncer
  ├─ Replica 3: 50 connections → PgBouncer
  ├─ Replica 4: 50 connections → PgBouncer
  └─ Replica 5: 50 connections → PgBouncer
               ↓
          PgBouncer
    max_client_conn: 10000   (принимает 10k connections)
    default_pool_size: 100   (держит 100 к PostgreSQL)
    pool_mode: transaction   (переиспользует connection после транзакции)
               ↓
          PostgreSQL
    max_connections: 200
```

**Конфигурация PgBouncer:**
```ini
[databases]
yammi_board = host=postgres port=5432 dbname=yammi_board user=yammi password=yammi

[pgbouncer]
pool_mode = transaction
max_client_conn = 10000
default_pool_size = 100
min_pool_size = 10
reserve_pool_size = 10
reserve_pool_timeout = 5

listen_addr = *
listen_port = 6432
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt

server_lifetime = 3600
server_idle_timeout = 600
```

**docker-compose.yml:**
```yaml
pgbouncer:
  image: pgbouncer/pgbouncer:latest
  environment:
    DATABASES_HOST: postgres
    DATABASES_PORT: 5432
    DATABASES_DBNAME: yammi_board
    DATABASES_USER: yammi
    DATABASES_PASSWORD: yammi
    PGBOUNCER_POOL_MODE: transaction
    PGBOUNCER_MAX_CLIENT_CONN: 10000
    PGBOUNCER_DEFAULT_POOL_SIZE: 100
  ports:
    - "6432:6432"
  depends_on:
    - postgres

board:
  environment:
    DATABASE_URL: postgres://yammi:yammi@pgbouncer:6432/yammi_board  # через PgBouncer!
```

---

## Альтернативы

### ❌ Вариант 1: Увеличить max_connections в PostgreSQL

**Идея:** `max_connections = 1000`

**Минусы:**
- Каждое connection потребляет ~10MB RAM
- 1000 connections × 10MB = 10GB RAM только на connections
- PostgreSQL не рассчитан на тысячи connections (context switching overhead)
- Performance деградирует при > 200 connections

### ❌ Вариант 2: Уменьшить connection pool в приложении

**Идея:** Board Service: `MaxOpenConns = 20` (вместо 50)

**Минусы:**
- При highload — очередь на получение connection
- Latency растёт (ожидание свободного connection)
- Throughput падает (меньше параллельных запросов)

### ❌ Вариант 3: Каждый сервис — свой PostgreSQL instance

**Минусы:**
- Expensive (5 PostgreSQL instances вместо одного)
- Сложнее управление (backups, replication)
- Оверкилл для текущей нагрузки

---

## Последствия

### ✅ Плюсы

1. **Масштабирование приложения** — можно добавлять replicas без лимита connections
2. **Эффективное использование PostgreSQL** — 100 реальных connections вместо 250 idle
3. **Connection reuse** — transaction mode переиспользует connection после COMMIT/ROLLBACK
4. **Failover** — PgBouncer может переключаться между PostgreSQL replicas
5. **Мониторинг** — PgBouncer экспортирует метрики (active connections, wait queue)

### ⚠️ Минусы

1. **Дополнительный компонент** — еще один сервис в инфраструктуре
2. **Transaction pool mode ограничения:**
   - Нельзя использовать prepared statements (PgBouncer их не поддерживает)
   - Нельзя использовать session-level переменные
   - **Решение:** Session pool mode (но меньше reuse)

### 🔧 Компенсация

- Prepared statements не критичны (Go database/sql работает без них)
- Session variables не используются в приложении

---

## Метрики

**До PgBouncer (прогноз):**
- PostgreSQL max_connections: 200
- Board Service: 5 replicas × 50 = 250 connections
- **Результат:** ERROR: too many connections

**После PgBouncer (прогноз):**
- Board Service: 5 replicas × 50 = 250 client connections к PgBouncer
- PgBouncer → PostgreSQL: 100 реальных connections (reuse)
- PostgreSQL usage: 100 / 200 = 50% (ок)

**При 10 replicas:**
- Board Service: 10 × 50 = 500 client connections
- PgBouncer → PostgreSQL: 100 connections (всё ещё reuse)

---

## Связанные решения

- [Performance & Highload](../performance.md)
- [Board Service Architecture](../board-service.md)

---

## Реализация

**Фаза:** Запланировано

**Файлы:**
- `deployments/pgbouncer.ini` — конфигурация PgBouncer
- `docker-compose.yml` — добавить pgbouncer сервис
- `services/board/internal/infrastructure/database.go` — изменить DATABASE_URL

**Тестирование:**
```bash
# Проверка подключения через PgBouncer
psql -h localhost -p 6432 -U yammi yammi_board

# Мониторинг connections
SHOW POOLS;
SHOW CLIENTS;
SHOW SERVERS;
```
