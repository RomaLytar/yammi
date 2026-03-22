# 004. Lexorank для позиций карточек

**Статус:** ✅ Принято

**Дата:** 2024-03-20

---

## Контекст

Карточки в колонке должны иметь порядок. Пользователь может drag-and-drop карточку между другими.

**Проблема:**
Как хранить позицию карточки, чтобы reorder был **быстрым** и **без race conditions**?

**Требования:**
- 100k concurrent users
- Частые reorder операции (drag-and-drop)
- Нет массовых UPDATE при перемещении одной карточки
- Concurrent reorder не должен вызывать конфликты

---

## Решение

**Lexorank** — лексикографическое ранжирование (как в Jira, Trello).

**Позиция — строка, а не INT:**
```go
type Card struct {
    Position string  // "a", "am", "b", "c", ...
}
```

**Алгоритм:**
```
Карточки: [a] [b] [c]
Переместить новую между [a] и [b]:
  position = LexorankBetween("a", "b") = "am"

Результат: [a] [am] [b] [c]

UPDATE cards SET position = 'am' WHERE id = $1;  // только 1 UPDATE!
```

**Реализация:** `services/board/internal/domain/lexorank.go`

---

## Альтернативы

### ❌ Вариант 1: Position INT с промежутками

**Идея:** Позиции 10, 20, 30, ... (оставить промежутки для вставок)

```go
type Card struct {
    Position int  // 10, 20, 30, ...
}
```

**Проблемы:**
```
Карточки: [10] [20] [30]
Вставляем между 10 и 20 → position = 15 (ок)
Вставляем между 10 и 15 → position = 12 (ок)
Вставляем между 10 и 12 → position = 11 (ок)
Вставляем между 10 и 11 → ??? Промежуток закончился!

Решение: rebalance всех позиций (10, 20, 30, ...) → UPDATE N строк
```

**Минусы:**
- Периодически нужен rebalance (массовые UPDATE)
- Concurrent reorder → race conditions на rebalance

### ❌ Вариант 2: Position FLOAT

**Идея:** Позиции 1.0, 2.0, 3.0, вставка между → 1.5

```go
type Card struct {
    Position float64
}
```

**Проблемы:**
```
Вставляем между 1.0 и 2.0 → 1.5
Вставляем между 1.0 и 1.5 → 1.25
Вставляем между 1.0 и 1.25 → 1.125
...
После N вставок → точность float64 кончится
```

**Минусы:**
- Ограниченная точность (float64 не бесконечный)
- Нужен eventual rebalance

### ❌ Вариант 3: Position INT с автоинкрементом + сортировка в памяти

**Идея:** Position = sequence, сортировка на фронте

**Минусы:**
- Фронтенд не знает порядок до загрузки всех карточек
- PostgreSQL `ORDER BY position` не будет работать корректно

---

## Последствия

### ✅ Плюсы

1. **1 UPDATE вместо N** — при reorder обновляется только одна карточка
2. **Нет race conditions** — каждая карточка независимая
3. **Нет rebalance** — позиции никогда не "кончаются" (строки бесконечны)
4. **Concurrent reorder безопасен** — карточки не влияют друг на друга
5. **Простая сортировка** — `ORDER BY position ASC` (лексикографическая)

### ⚠️ Минусы

1. **Позиции могут "схлопнуться"** — после 1000+ reorder одной карточки строка станет очень длинной
   - **Решение:** Периодический rebalance (но не критично, можно делать раз в месяц)
2. **Сложнее в понимании** — не интуитивно для новых разработчиков
   - **Решение:** Документация + unit тесты

### 🔧 Компенсация минусов

**Схлопывание позиций:**
```go
// Если position стала слишком длинной (>100 символов) — rebalance колонки
func (uc *RebalanceColumnUseCase) Execute(columnID string) {
    cards := uc.repo.ListByColumnID(columnID)
    for i, card := range cards {
        card.Position = fmt.Sprintf("%c", 'a'+i)  // a, b, c, d, ...
    }
    // UPDATE N карточек (но только при необходимости, не каждый reorder)
}
```

**Документация:**
- Гайд: `guides/board-service.md`
- Unit тесты: `internal/domain/lexorank_test.go`

---

## Метрики

### Unit Tests

✅ **21 test cases в `lexorank_test.go`:**
- Between two positions
- Edge cases (empty, adjacent, long strings)
- Ordering verification
- 1000+ sequential inserts

### Performance Benchmarks

```bash
BenchmarkLexorankBetween-8   5000000   250 ns/op
```

**UPDATE latency (PostgreSQL):**
```sql
UPDATE cards SET position = 'am' WHERE id = $1;
-- Ожидаемая latency: < 5ms (indexed)
```

---

## Связанные решения

- [Board Service Architecture](../board-service.md)
- [Performance & Highload](../performance.md)

---

## Примеры использования

**CreateCard (в конец колонки):**
```go
lastCard := repo.GetLastCardInColumn(columnID)
newPosition, _ := LexorankBetween(lastCard.Position, "")  // append
card := NewCard(columnID, title, desc, newPosition, assigneeID)
```

**MoveCard (между двумя карточками):**
```go
prevCard := repo.GetCardAtPosition(columnID, targetPosition - 1)
nextCard := repo.GetCardAtPosition(columnID, targetPosition)
newPosition, _ := LexorankBetween(prevCard.Position, nextCard.Position)

card.Move(targetColumnID, newPosition)
repo.Update(card)  // UPDATE 1 строку
```

**Rebalance (если нужно):**
```go
if len(card.Position) > 100 {
    uc.RebalanceColumn(columnID)  // редкая операция
}
```
