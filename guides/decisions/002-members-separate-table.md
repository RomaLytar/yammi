# 002. Members в отдельной таблице (НЕ в Board aggregate)

**Статус:** ✅ Принято

**Дата:** 2024-03-20

---

## Контекст

При разработке Board Service возник вопрос: **где хранить members (участников доски)?**

**Варианты:**
1. В Board aggregate как `[]Member` (загружать вместе с доской)
2. В отдельной таблице `board_members` (query при необходимости)

**Проблема:**
- Доска может иметь 100+ участников
- При каждом `GetBoard()` загружать 100 members = тяжёлый payload
- Redis cache будет хранить лишние данные
- При `AddMember` / `RemoveMember` — нужно обновлять всю доску

---

## Решение

**Members хранятся отдельно** — таблица `board_members` (many-to-many).

**Board aggregate содержит только `OwnerID`:**
```go
type Board struct {
    ID          string
    Title       string
    Description string
    OwnerID     string  // НЕ загружаем members в aggregate
    Version     int
}
```

**Проверка доступа — через MembershipRepository:**
```go
// Вместо board.HasMember(userID) в памяти
isMember, role, err := memberRepo.IsMember(boardID, userID)  // SELECT EXISTS

if !isMember {
    return ErrAccessDenied
}
```

**Отдельный API для управления members:**
```
GET /boards/{id}/members?limit=20&offset=0  — пагинация
POST /boards/{id}/members                    — добавить участника
DELETE /boards/{id}/members/{userId}         — удалить участника
```

---

## Альтернативы

### ❌ Вариант 1: Members в Board aggregate

**Код:**
```go
type Board struct {
    Members []Member  // загружаем всех при GetBoard()
}
```

**Минусы:**
- Загрузка 100 members при каждом GetBoard() → тяжёлый payload
- Redis cache хранит лишние данные
- При AddMember — UPDATE всей доски (optimistic lock conflict риск)
- Нельзя сделать pagination members

### ❌ Вариант 2: Members как отдельный aggregate с кешем

**Минусы:**
- Дублирование данных в Redis (`board_members:{boardID}`)
- Сложная синхронизация cache (при каждом AddMember/RemoveMember инвалидируем)
- Для простой проверки IsMember не нужен кеш — PostgreSQL справится

---

## Последствия

### ✅ Плюсы

1. **Лёгкий GetBoard()** — только метаданные (без members)
2. **Быстрая проверка доступа** — `SELECT EXISTS` (миллисекунды)
3. **Pagination для members** — легко добавить `LIMIT/OFFSET`
4. **Redis cache эффективнее** — кешируем только board метаданные
5. **Меньше optimistic lock conflicts** — изменение members не трогает board.version

### ⚠️ Минусы

1. **Дополнительный запрос** — `IsMember()` перед операциями с доской
2. **Два источника истины** — Board в одной таблице, members в другой

### 🔧 Компенсация минусов

- `IsMember()` — очень быстрый запрос (индекс на `board_id, user_id`)
- Foreign key `board_members.board_id → boards.id ON DELETE CASCADE` — при удалении доски members автоматически удаляются

---

## Метрики

(будут после load testing)

**Ожидаемые результаты:**
- GetBoard latency: < 50ms (без members)
- IsMember latency: < 5ms (индексированный query)
- AddMember latency: < 10ms (INSERT в board_members)

---

## Связанные решения

- [003: Granular Redis cache](003-redis-cache-strategy.md)
- [Board Service Architecture](../board-service.md)
