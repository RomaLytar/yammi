# Схема базы данных Board Service

## Обзор

Board Service использует PostgreSQL 16 с продвинутыми возможностями:
- **HASH партиционирование** (cards по board_id)
- **Optimistic locking** (version field)
- **Cursor pagination индексы**
- **Foreign keys с CASCADE DELETE**

База данных: `yammi_board`

Миграции: `services/board/migrations/000001_init.up.sql`, `000002_board_search_sort.up.sql`, `000003_card_creator_id.up.sql`

## Архитектурное решение: Micro-Aggregates

**Ключевое отличие от традиционного DDD:** Board, Column и Card — отдельные aggregate roots, а не один большой Board aggregate.

**Почему так:**
1. **Производительность** — `GetBoard` не загружает 500 карточек в память
2. **Granular caching** — кеш карточки инвалидируется независимо от доски
3. **Concurrency** — 5 пользователей могут одновременно редактировать разные карточки без конфликтов version

**Trade-off:** Транзакционные границы слабее. Нельзя гарантировать что "board всегда имеет ровно 3 колонки" на уровне БД (проверка в usecase).

---

## Таблица: boards

Хранит метаданные досок (БЕЗ columns и cards — они в отдельных таблицах).

```sql
CREATE TABLE boards (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    owner_id UUID NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Колонки

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID | Primary key, генерируется в domain (`uuid.NewString()`) |
| `title` | VARCHAR(255) | Название доски (NOT NULL, валидация в domain) |
| `description` | TEXT | Описание доски (опциональное, default пустая строка) |
| `owner_id` | UUID | Создатель доски (ссылка на user из Auth Service) |
| `version` | INT | **Optimistic locking счетчик** (инкремент при каждом UPDATE) |
| `created_at` | TIMESTAMPTZ | Дата создания (автоматически при INSERT) |
| `updated_at` | TIMESTAMPTZ | Дата последнего изменения (обновляется в domain) |

### Индексы

```sql
CREATE INDEX idx_boards_owner_id ON boards(owner_id);
CREATE INDEX idx_boards_cursor ON boards(created_at DESC, id DESC);
CREATE INDEX idx_boards_title_trgm ON boards USING gin (title gin_trgm_ops);
CREATE INDEX idx_boards_updated_at ON boards(updated_at DESC);
```

- **`idx_boards_owner_id`** — поиск всех досок конкретного владельца (`ListBoards`)
- **`idx_boards_cursor`** — cursor pagination (`ORDER BY created_at DESC, id DESC LIMIT 20`)
- **`idx_boards_title_trgm`** — **pg_trgm GIN индекс** для нечеткого/частичного поиска по названию доски (`WHERE title ILIKE '%query%'` или trigram similarity)
- **`idx_boards_updated_at`** — сортировка досок по дате последнего изменения (`ORDER BY updated_at DESC`)

### Почему такая структура

**1. БЕЗ columns/cards в JSON-полях**

❌ Плохо (aggregate root с вложенными объектами):
```sql
CREATE TABLE boards (
    ...
    columns JSONB  -- [{id, title, cards: [...]}]
);
```

✅ Хорошо (нормализация, отдельные таблицы):
- Можно обновить карточку без загрузки всей доски (performance)
- PostgreSQL индексы работают с отдельными таблицами, не с JSONB
- Partitioning cards по board_id (горизонтальное масштабирование)

**2. `version` для optimistic locking**

**Проблема:**
```
User A: GET /boards/123  (version=5)
User B: GET /boards/123  (version=5)
User A: PUT /boards/123 {title="New"}  -> version=6
User B: PUT /boards/123 {title="Foo"}  -> перезаписал изменения User A (lost update!)
```

**Решение:**
```sql
UPDATE boards SET title = $1, version = version + 1
WHERE id = $2 AND version = $3
RETURNING *;
```

Если version изменился — UPDATE вернет 0 строк → usecase вернет `ErrOptimisticLockFailed`.

**3. `owner_id` вместо массива members**

Members хранятся в отдельной таблице `board_members` (см. ниже). `owner_id` — denormalization для быстрой проверки владения:
```sql
SELECT EXISTS (SELECT 1 FROM boards WHERE id = $1 AND owner_id = $2);
```

Owner автоматически добавляется в `board_members` при создании доски (usecase).

---

## Таблица: board_members

Many-to-many таблица для sharing досок. Реализует access control (owner vs member).

```sql
CREATE TABLE board_members (
    board_id UUID REFERENCES boards(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('owner', 'member')),
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (board_id, user_id)
);
```

### Колонки

| Колонка | Тип | Описание |
|---------|-----|----------|
| `board_id` | UUID | FK на boards (CASCADE DELETE) |
| `user_id` | UUID | ID пользователя (из Auth/User Service) |
| `role` | VARCHAR(20) | Роль: `'owner'` или `'member'` (CHECK constraint) |
| `joined_at` | TIMESTAMPTZ | Дата добавления в доску |

**Primary Key:** `(board_id, user_id)` — уникальная пара (один user на доске только один раз).

### Индексы

```sql
CREATE INDEX idx_board_members_user_id ON board_members(user_id);
```

- Быстрый поиск всех досок пользователя (`ListBoards WHERE user_id = $1`)

### Почему отдельная таблица

**1. Производительность при большом количестве участников**

Доска может иметь 100+ участников. Загружать всех при `GetBoard` — тяжело. Вместо этого:
- `GetBoard` возвращает только метаданные доски
- `ListMembers` — отдельный эндпоинт с пагинацией

**2. Проверка доступа — fast query**

```sql
SELECT EXISTS (
    SELECT 1 FROM board_members
    WHERE board_id = $1 AND user_id = $2
);
```

Индекс по Primary Key (board_id, user_id) — O(1) lookup.

**3. Pagination участников**

```sql
SELECT user_id, role, joined_at
FROM board_members
WHERE board_id = $1
ORDER BY joined_at
LIMIT 50 OFFSET 0;
```

**4. Roles для авторизации**

| Role | Права |
|------|-------|
| `owner` | Все (CRUD доски, управление участниками, удаление доски) |
| `member` | CRUD карточек, чтение доски (нельзя редактировать метаданные, добавлять участников) |

Проверка в usecase:
```go
member, err := membershipRepo.GetMembership(ctx, boardID, userID)
if !member.CanModifyBoard() {
    return ErrAccessDenied
}
```

### Бизнес-правила (enforced в usecase)

1. **При CREATE board** → автоматически `INSERT (board_id, owner_id, 'owner')`
2. **Нельзя удалить owner** — защита в `RemoveMemberUseCase`
3. **Только owner может добавлять/удалять members** — проверка в `AddMemberUseCase`
4. **CASCADE DELETE** — при удалении доски все members автоматически удаляются

---

## Таблица: columns

Колонки доски (To Do, In Progress, Done).

```sql
CREATE TABLE columns (
    id UUID PRIMARY KEY,
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    position INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Колонки

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID | Primary key (генерируется в domain) |
| `board_id` | UUID | FK на boards (CASCADE DELETE) |
| `title` | VARCHAR(255) | Название колонки (To Do, In Progress, Done) |
| `position` | INT | Порядок колонки (0, 1, 2, ...) |
| `created_at` | TIMESTAMPTZ | Дата создания |

### Индексы

```sql
CREATE INDEX idx_columns_board_id ON columns(board_id);
CREATE INDEX idx_columns_position ON columns(board_id, position);
```

- **`idx_columns_board_id`** — загрузка всех колонок доски
- **`idx_columns_position`** — сортировка по позиции (`ORDER BY position`)

### Почему INT position (не lexorank)

**Columns vs Cards:**
| Характеристика | Columns | Cards |
|----------------|---------|-------|
| Количество | 3-10 | 100-500 |
| Reorder frequency | Редко (раз в месяц) | Часто (десятки раз в день) |
| Position type | **INT** | **Lexorank (string)** |

**Почему INT для columns:**
- Колонок мало (обычно 3-5)
- Reorder колонок — редкая операция
- Массовый UPDATE 10 колонок — не критично (< 10ms)

**Пример reorder columns:**
```sql
-- Переместить колонку 2 на позицию 0
UPDATE columns SET position = CASE
    WHEN id = 'col2' THEN 0
    WHEN position < 2 THEN position + 1
    ELSE position
END
WHERE board_id = $1;
```

**Альтернатива (lexorank):** Усложняет код без значительного выигрыша в производительности.

### CASCADE DELETE

При удалении доски → все колонки автоматически удаляются (FK с `ON DELETE CASCADE`).

**Цепная реакция:**
```
DELETE boards WHERE id = $1
  ↓ CASCADE
DELETE columns WHERE board_id = $1
  ↓ CASCADE (FK в cards)
DELETE cards WHERE column_id IN (...)
```

---

## Таблица: cards (PARTITIONED!)

Карточки с **HASH partitioning по board_id**.

```sql
CREATE TABLE cards (
    id UUID NOT NULL,
    column_id UUID NOT NULL,
    board_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    position VARCHAR(100) NOT NULL,
    assignee_id UUID,
    creator_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (board_id, id)
) PARTITION BY HASH (board_id);

-- 4 партиции
CREATE TABLE cards_p0 PARTITION OF cards FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE cards_p1 PARTITION OF cards FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE cards_p2 PARTITION OF cards FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE cards_p3 PARTITION OF cards FOR VALUES WITH (MODULUS 4, REMAINDER 3);
```

### Колонки

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | UUID | ID карточки (НЕ primary key сам по себе!) |
| `column_id` | UUID | FK на columns (в какой колонке карточка) |
| `board_id` | UUID | **Partition key** (для распределения по партициям) |
| `title` | VARCHAR(255) | Название карточки (NOT NULL) |
| `description` | TEXT | Описание задачи (опциональное) |
| **`position`** | **VARCHAR(100)** | **Lexorank позиция (string!)**: "a", "am", "b", ... |
| `assignee_id` | UUID | Кому назначена карточка (nullable) |
| `creator_id` | UUID | **Создатель карточки** (NOT NULL, используется для permission checks) |
| `created_at` | TIMESTAMPTZ | Дата создания |
| `updated_at` | TIMESTAMPTZ | Дата изменения (обновляется в domain) |

**Primary Key:** `(board_id, id)` — composite key (board_id обязательно для partitioning).

### Индексы

```sql
CREATE INDEX idx_cards_column_id ON cards(column_id);
CREATE INDEX idx_cards_position ON cards(column_id, position);
CREATE INDEX idx_cards_assignee_id ON cards(assignee_id) WHERE assignee_id IS NOT NULL;
CREATE INDEX idx_cards_creator_id ON cards(creator_id);
```

- **`idx_cards_column_id`** — загрузка всех карточек колонки
- **`idx_cards_position`** — сортировка по lexorank (`ORDER BY position`)
- **`idx_cards_assignee_id`** — **partial index** (только NOT NULL, экономит место)
- **`idx_cards_creator_id`** — поиск всех карточек, созданных конкретным пользователем (используется для permission checks при удалении)

### Почему HASH partitioning

**1. Производительность — запросы к одной доске → только 1 партиция**

Без partitioning:
```sql
SELECT * FROM cards WHERE column_id = $1;
-- Сканирует ВСЕ карточки всех досок (millions of rows)
```

С partitioning:
```sql
SELECT * FROM cards WHERE column_id = $1;
-- PostgreSQL определяет board_id из column_id (JOIN)
-- → сканирует только 1 партицию (1/4 данных)
```

**2. Масштабируемость — можно добавить партиции**

Текущая конфигурация: 4 партиции. Если база разрастется:
```sql
-- Добавляем еще партиции (requires re-partitioning)
CREATE TABLE cards_p4 PARTITION OF cards FOR VALUES WITH (MODULUS 8, REMAINDER 4);
...
```

**3. Автоматическое распределение**

PostgreSQL сам определяет партицию по `HASH(board_id) % 4`:
- board_id = abc → HASH = 123 → 123 % 4 = 3 → cards_p3
- board_id = xyz → HASH = 456 → 456 % 4 = 0 → cards_p0

**Trade-off:** Нельзя легко запросить "все карточки пользователя X" (нужен `board_id`).

### Почему lexorank position (string)

**Проблема INT position:**
```
Cards: [1, 2, 3, 4, 5]
Переместить карточку 5 между 2 и 3:
  → UPDATE cards SET position = position + 1 WHERE position >= 3;  (затронуто 3 строки)
  → UPDATE cards SET position = 3 WHERE id = 5;  (еще 1 UPDATE)
  → Итого: 4 UPDATEs для одного перемещения!
```

При 500 карточках в колонке → можем обновить 250+ строк за один drag&drop.

**Решение: lexorank (лексикографическое позиционирование)**

```
Cards: ["a", "b", "c", "d", "e"]
Переместить "e" между "b" и "c":
  → newPosition = LexorankBetween("b", "c") = "bm"
  → UPDATE cards SET position = 'bm' WHERE id = e;  (1 UPDATE!)
  → Результат: ["a", "b", "bm", "c", "d", "e"]
```

**Сортировка — лексикографическая (алфавитная):**
```sql
SELECT * FROM cards WHERE column_id = $1 ORDER BY position ASC;
-- "a" < "am" < "b" < "bm" < "c" < "d" < "e"
```

**Подробнее:** См. `/home/roman/PetProject/yammi/guides/lexorank-explained.md`

### Partitioning прозрачен для приложения

**Приложение работает так, как будто partitioning нет:**

```go
// INSERT — PostgreSQL сам выбирает партицию
db.Exec("INSERT INTO cards (id, board_id, ...) VALUES ($1, $2, ...)")

// SELECT — PostgreSQL сканирует только нужную партицию
db.Query("SELECT * FROM cards WHERE board_id = $1")
```

**Важно:** `board_id` должен быть в WHERE-clause (иначе PostgreSQL сканирует все партиции).

---

## Foreign Keys и CASCADE DELETE

**Цепочка удалений:**

```
boards (id)
  ↓ FK: columns(board_id) ON DELETE CASCADE
columns (id)
  ↓ FK: cards(column_id) ON DELETE CASCADE
cards
```

**Пример:**
```sql
DELETE FROM boards WHERE id = 'board-123';

-- Автоматически:
-- 1. DELETE FROM columns WHERE board_id = 'board-123';
-- 2. DELETE FROM cards WHERE column_id IN (select id from columns where board_id = 'board-123');
-- 3. DELETE FROM board_members WHERE board_id = 'board-123';
```

**Почему CASCADE, а не soft delete:**
- Board Service — не система учета (не нужен audit trail)
- Soft delete усложняет queries (везде `WHERE deleted_at IS NULL`)
- Если нужна история — используем NATS events (`board.deleted`, `card.deleted`)

---

## Optimistic Locking (детально)

**Зачем:** Предотвратить lost updates при concurrent изменениях.

**Как работает:**

1. **Клиент A:** `GET /boards/123` → `{title: "Foo", version: 5}`
2. **Клиент B:** `GET /boards/123` → `{title: "Foo", version: 5}`
3. **Клиент A:** `PUT /boards/123 {title: "Bar"}` + `version: 5` в запросе
   - Usecase: `repo.Update(board) WHERE version = 5`
   - SQL: `UPDATE boards SET title='Bar', version=6 WHERE id=123 AND version=5`
   - Результат: 1 row affected → успех
4. **Клиент B:** `PUT /boards/123 {title: "Baz"}` + `version: 5` в запросе
   - Usecase: `repo.Update(board) WHERE version = 5`
   - SQL: `UPDATE boards SET title='Baz', version=6 WHERE id=123 AND version=5`
   - Результат: **0 rows affected** (version уже 6!) → `ErrOptimisticLockFailed`
   - Клиент B получает 409 Conflict → "Доска была изменена, перезагрузите страницу"

**Реализация в repository:**

```go
func (r *BoardRepo) Update(ctx context.Context, board *Board) error {
    query := `
        UPDATE boards
        SET title = $1, description = $2, version = version + 1, updated_at = NOW()
        WHERE id = $3 AND version = $4
        RETURNING version
    `

    var newVersion int
    err := r.db.QueryRowContext(ctx, query,
        board.Title, board.Description, board.ID, board.Version,
    ).Scan(&newVersion)

    if err == sql.ErrNoRows {
        return ErrOptimisticLockFailed  // version изменился
    }

    board.Version = newVersion  // обновляем version в объекте
    return err
}
```

**Почему только boards, а не cards/columns:**
- Cards/Columns изменяются независимо (concurrent edits разных карточек — OK)
- Board — shared state (2 юзера меняют title одновременно — conflict)

---

## Cursor Pagination (детально)

**Зачем:** OFFSET-based pagination не масштабируется (большие offset — slow).

**Как работает:**

**1. Первый запрос (без cursor):**
```sql
SELECT id, title, created_at
FROM boards
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

**2. Клиент получает последнюю запись:**
```json
{
  "id": "uuid-20",
  "title": "Board 20",
  "created_at": "2026-03-19T10:00:00Z"
}
```

**3. Следующий запрос (с cursor):**
```sql
SELECT id, title, created_at
FROM boards
WHERE user_id = $1
  AND (created_at, id) < ($2, $3)  -- cursor (timestamp, uuid)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

**Индекс:** `idx_boards_cursor (created_at DESC, id DESC)` — PostgreSQL использует index scan (fast).

**Почему (created_at, id), а не только created_at:**
- Две доски могут иметь одинаковый `created_at` (миллисекунды)
- `id` — уникальный tiebreaker

---

## Partial Index (детально)

```sql
CREATE INDEX idx_cards_assignee_id ON cards(assignee_id)
WHERE assignee_id IS NOT NULL;
```

**Зачем:**
- 80% карточек не имеют assignee (NULL)
- Индексировать NULL — бесполезно (никогда не ищем `WHERE assignee_id IS NULL`)
- Partial index — только NOT NULL → индекс в 5 раз меньше

**Пример запроса:**
```sql
SELECT * FROM cards WHERE assignee_id = 'user-123';
-- PostgreSQL использует idx_cards_assignee_id (только NOT NULL строки)
```

---

## Миграции

**Файлы:**
- `services/board/migrations/000001_init.up.sql` — начальная схема (boards, board_members, columns, cards с partitioning)
- `services/board/migrations/000002_board_search_sort.up.sql` — pg_trgm индекс на boards.title, индекс на boards.updated_at
- `services/board/migrations/000003_card_creator_id.up.sql` — добавление `creator_id` в cards, индекс `idx_cards_creator_id`

**Как накатываются:**
1. При старте сервиса → `infrastructure/migrator.go`
2. Читает `.up.sql` файлы из папки `migrations/`
3. Выполняет в транзакции (если упало — rollback)

**Именование:** `<version>_<description>.up.sql` (автоматическая сортировка по версии)

**Down миграции:** `000001_init.down.sql` (для rollback, пока не реализовано)

---

## Производительность и масштабирование

### Текущие bottlenecks

| Операция | Сложность | Узкое место |
|----------|-----------|-------------|
| GetBoard | O(1) | Primary key lookup |
| ListCards | O(n log n) | Sort по lexorank position (INDEX SCAN) |
| ReorderCard | O(1) | UPDATE одной строки (lexorank) |
| DeleteBoard | O(columns * cards) | CASCADE DELETE (но редкая операция) |

### Кеширование (Redis)

**Стратегия:** Read-through cache с TTL 5 минут.

**Cached entities:**
- `board:{id}` — метаданные доски
- `columns:{board_id}` — список колонок
- `cards:{column_id}` — список карточек колонки

**Invalidation:**
- При UPDATE board → `DEL board:{id}`
- При ADD/DELETE column → `DEL columns:{board_id}`
- При UPDATE card → `DEL cards:{column_id}`

**Trade-off:** Eventual consistency (кеш может быть stale до 5 минут).

### Horizontal Scaling

**Stateless сервисы:** Board Service можно масштабировать горизонтально (N реплик за load balancer).

**Shared state:**
- PostgreSQL — single master (для write), read replicas (для read)
- Redis — можно кластеризовать (Redis Cluster)
- NATS — built-in clustering

**Partitioning cards:** При росте данных можно увеличить количество партиций (4 → 8 → 16).

---

## Сравнение с традиционным DDD

| Аспект | Традиционный DDD | Yammi Board Service |
|--------|------------------|---------------------|
| Aggregate Root | Board содержит Columns и Cards | Board, Column, Card — отдельные aggregates |
| Persistence | `BoardRepository.Save(board)` сохраняет всё | Отдельные repos: BoardRepo, ColumnRepo, CardRepo |
| Транзакции | Одна транзакция на aggregate | Отдельные транзакции (trade-off consistency) |
| Performance | Загружает весь граф (board + columns + cards) | Granular loading (только нужные части) |
| Caching | Инвалидация всего aggregate | Инвалидация отдельных частей |

**Когда использовать каждый подход:**

- **Традиционный DDD:** Сложные инварианты (например, "сумма всех card.points == board.total_points")
- **Micro-aggregates (Yammi):** Простые инварианты, акцент на производительности

---

## Permissions Model

Авторизация действий зависит от роли пользователя (`board_members.role`) и поля `creator_id` в карточках.

### Удаление карточек

| Роль | Правило |
|------|---------|
| `owner` | Может удалить **любую** карточку на доске |
| `member` | Может удалить **только свои** карточки (где `cards.creator_id = user_id`) |

**Проверка в usecase:**
```go
if member.Role != "owner" && card.CreatorID != userID {
    return ErrAccessDenied // member может удалять только свои карточки
}
```

### Удаление колонок

| Роль | Правило |
|------|---------|
| `owner` | Может удалить любую колонку |
| `member` | **Не может** удалять колонки |

**Проверка в usecase:**
```go
if member.Role != "owner" {
    return ErrAccessDenied // только owner может удалять колонки
}
```

### Сводная таблица прав

| Операция | Owner | Member |
|----------|-------|--------|
| Создание карточки | Да | Да |
| Редактирование карточки | Да (любой) | Да (любой) |
| Удаление карточки | Да (любой) | Только свои (`creator_id`) |
| Создание колонки | Да | Да |
| Удаление колонки | Да | Нет |
| Редактирование доски | Да | Нет |
| Управление участниками | Да | Нет |

---

## TouchUpdatedAt: автообновление board.updated_at

При любом изменении карточек или колонок автоматически обновляется `boards.updated_at`. Это позволяет сортировать доски по дате последней активности (индекс `idx_boards_updated_at`).

**Механизм:** После операций с карточками/колонками usecase вызывает обновление `updated_at` у родительской доски:
```sql
UPDATE boards SET updated_at = NOW() WHERE id = $1;
```

**Какие операции обновляют `board.updated_at`:**
- Создание/удаление/перемещение карточки
- Редактирование карточки (title, description, assignee)
- Создание/удаление/переименование колонки
- Изменение порядка колонок

**Зачем:** Пользователь видит доски, отсортированные по последней активности ("недавние доски сверху"), а не только по дате создания. Индекс `idx_boards_updated_at` обеспечивает быструю сортировку.

---

## Итоговая схема

```
boards (id, title, owner_id, version)
  ├─ board_members (board_id, user_id, role) — many-to-many
  └─ columns (board_id, title, position)
       └─ cards (column_id, board_id, position lexorank, creator_id) — PARTITIONED!
```

**Ключевые особенности:**
1. **Partitioning cards** — масштабируемость
2. **Lexorank position** — O(1) reordering
3. **Optimistic locking** — concurrent safety
4. **Cursor pagination** — performance на больших датасетах
5. **Partial index** — экономия места
6. **CASCADE DELETE** — упрощение логики удаления
