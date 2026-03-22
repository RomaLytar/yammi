# 006. Миграция с lib/pq на pgx

**Статус:** ✅ Принято
**Дата:** 2026-03-22

---

## Контекст

PgBouncer в `transaction` mode необходим для connection multiplexing (205 app connections → 20 реальных PG connections). Но `lib/pq` использует extended query protocol (Parse/Bind/Execute), который **несовместим** с PgBouncer transaction mode — prepared statements теряются при смене connection.

Нагрузочный тест (1000 VU) показал:
- PgBouncer transaction + lib/pq = **69% ошибок** (prepared statement mismatch)
- PgBouncer session + lib/pq = 0% ошибок, но **без мультиплексирования** (latency выше baseline)

## Решение

Замена `lib/pq` на `pgx/v5/stdlib` во всех сервисах:

```go
// Было
import _ "github.com/lib/pq"
db, err := sql.Open("postgres", url)

// Стало
import _ "github.com/jackc/pgx/v5/stdlib"
db, err := sql.Open("pgx", url)
```

`pgx` нативно поддерживает PgBouncer transaction mode — не использует extended query protocol для простых запросов.

## Альтернативы

- ❌ **lib/pq + session mode** — работает, но нет мультиплексирования
- ❌ **lib/pq + `?binary_parameters=yes`** — не решает prepared statement проблему
- ❌ **Увеличить max_connections** — не скалируется, PostgreSQL деградирует при >200 connections

## Последствия

- ✅ PgBouncer transaction mode работает с 0% ошибок
- ✅ Connection multiplexing: 105 app connections → 20 реальных
- ✅ `pgtype.FlatArray` вместо `pq.Array` — нативная поддержка PostgreSQL arrays
- ⚠️ Board Service сохраняет `lib/pq` для integration tests (testcontainers подключаются напрямую к PG)
- ⚠️ Миграции идут через `MIGRATION_DATABASE_URL` напрямую к PostgreSQL (advisory locks не работают через PgBouncer)

## Связанные решения

- [005-pgbouncer.md](./005-pgbouncer.md) — почему PgBouncer
