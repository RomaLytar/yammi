# Lexorank — Позиционирование без массовых UPDATE

## Что это

**Lexorank** — алгоритм для позиционирования элементов в списке с помощью **строковых** позиций вместо целочисленных.

Ключевая идея: позиция элемента — это **строка** (например, `"a"`, `"am"`, `"b"`), которая сортируется **лексикографически** (как в словаре).

**Где используется:**
- Jira (реализовал алгоритм и open-source библиотеку)
- Trello
- Linear
- Yammi Board Service (cards.position)

---

## Проблема INT position

### Сценарий: перемещение карточки в середину списка

**Начальная ситуация:**
```
Cards в колонке:
  Card A: position = 1
  Card B: position = 2
  Card C: position = 3
  Card D: position = 4
  Card E: position = 5
```

**Действие:** Переместить Card E между Card B и Card C.

**SQL (с INT position):**
```sql
-- Шаг 1: Освободить место (сдвинуть все после позиции 2)
UPDATE cards SET position = position + 1
WHERE column_id = $1 AND position >= 3;
-- Затронуто 3 строки (C, D, E)

-- Шаг 2: Поставить Card E на позицию 3
UPDATE cards SET position = 3 WHERE id = 'card-e';
-- Затронута 1 строка

-- Итого: 4 UPDATEs
```

**Результат:**
```
  Card A: position = 1
  Card B: position = 2
  Card E: position = 3  ← переместили сюда
  Card C: position = 4  ← изменилось!
  Card D: position = 5  ← изменилось!
```

### Проблемы этого подхода

1. **Множественные UPDATE при одном drag&drop**
   - При 500 карточках в колонке → можем обновить 250+ строк за одно перемещение
   - Высокая нагрузка на БД

2. **Locks и contention**
   - PostgreSQL блокирует строки при UPDATE
   - Concurrent reorder разных карточек → конфликты

3. **Race conditions**
   ```
   User A: drag Card X to position 5 (старт UPDATE position >= 5)
   User B: drag Card Y to position 4 (старт UPDATE position >= 4)
   → Один из UPDATE может затереть результаты другого
   ```

4. **Version increment на несвязанных записях**
   - Если бы мы использовали optimistic locking для cards
   - Сдвиг Card C (не затронутой пользователем) → version++
   - Concurrent edit Card C → ложный conflict

---

## Решение: Lexorank

### Основная идея

Позиция — это **строка**, которая сортируется **лексикографически** (алфавитно).

**Алфавит:** `0123456789abcdefghijklmnopqrstuvwxyz` (36 символов, base-36)

**Начальные позиции:**
- Первая карточка: `"n"` (середина алфавита)
- Вторая карточка: `"z"` (в конец)

**Вставка между двумя элементами:**
- Позиция = середина между соседями
- `Between("a", "b")` = `"am"`
- `Between("a", "am")` = `"ag"`

### Тот же сценарий с lexorank

**Начальная ситуация:**
```
Cards в колонке:
  Card A: position = "a"
  Card B: position = "b"
  Card C: position = "c"
  Card D: position = "d"
  Card E: position = "e"
```

**Действие:** Переместить Card E между Card B и Card C.

**SQL (с lexorank position):**
```sql
-- Шаг 1: Вычислить новую позицию (в приложении)
newPosition = LexorankBetween("b", "c")  // = "bm"

-- Шаг 2: UPDATE только одной карточки
UPDATE cards SET position = 'bm', updated_at = NOW()
WHERE id = 'card-e';
-- Затронута 1 строка!

-- Итого: 1 UPDATE
```

**Результат:**
```
  Card A: position = "a"
  Card B: position = "b"
  Card E: position = "bm"  ← переместили сюда
  Card C: position = "c"   ← НЕ изменилась!
  Card D: position = "d"   ← НЕ изменилась!
```

**Сортировка (лексикографическая):**
```sql
SELECT * FROM cards WHERE column_id = $1 ORDER BY position ASC;
-- "a" < "b" < "bm" < "c" < "d" < "e"
```

### Преимущества

✅ **O(1) UPDATE** — изменяется только одна строка
✅ **No locks на других карточках** — concurrent reorder безопасен
✅ **No race conditions** — каждая карточка обновляется независимо
✅ **Deterministic** — результат не зависит от порядка выполнения операций

---

## Как работает алгоритм

### Функция `LexorankBetween(prev, next)`

Вычисляет позицию между двумя строками `prev` и `next`.

**Реализация:** `/home/roman/PetProject/yammi/services/board/internal/domain/lexorank.go`

### Случай 1: Вставка между двумя элементами

```go
LexorankBetween("a", "c")
// Середина между 'a' и 'c' = 'b'
```

**Алгоритм:**
1. Выравнивает длины строк (добавляет '0' справа короткой)
   - `"a"` → `"a0"`
   - `"c"` → `"c0"`
2. Вычисляет среднее каждого символа
   - `'a'` = 10, `'c'` = 12 (в base-36)
   - Среднее = (10 + 12) / 2 = 11 = `'b'`
3. Результат: `"b"`

### Случай 2: Нет целого среднего (нужна дробь)

```go
LexorankBetween("a", "b")
// Между 'a' и 'b' нет целого символа!
// Решение: добавить разряд
```

**Алгоритм:**
1. `'a'` = 10, `'b'` = 11
2. Среднее = (10 + 11) / 2 = 10.5 (нецелое!)
3. Carry в следующий разряд → `"am"` (m = середина алфавита)

**Почему `"am"`:**
- `"a"` + `"m"` (m — 13-й символ, середина 0-36)
- `"a" < "am" < "b"` (лексикографически)

### Случай 3: Вставка в начало (prev пустая)

```go
LexorankBetween("", "b")
// Нужна позиция ПЕРЕД 'b'
```

**Алгоритм:**
1. Берем символ меньше `'b'`
2. `'b'` = 11 → 11 - 1 = 10 = `'a'`
3. Результат: `"a"`

Если `next = "0"` (минимальный символ):
```go
LexorankBetween("", "0")
// Нельзя взять меньше '0' → идем глубже
// Результат: "00n"  (добавляем разряд)
```

### Случай 4: Вставка в конец (next пустая)

```go
LexorankBetween("e", "")
// Нужна позиция ПОСЛЕ 'e'
```

**Алгоритм:**
1. Добавляем символ в конец
2. Результат: `"en"` (n — середина алфавита)

**Почему "en", а не "f":**
- `"f"` тоже работает, но оставляет меньше места для будущих вставок
- `"en"` — консистентная стратегия (всегда середина)

### Случай 5: Первая карточка в пустой колонке

```go
LexorankBetween("", "")
// Начальная позиция
```

**Результат:** `"n"` (константа `LexorankFirst`)

**Почему "n":**
- Середина алфавита (0-9a-z → n = 14-й символ)
- Максимум места для вставок в начало и конец

---

## Примеры работы

### Пример 1: Последовательное добавление в конец

```go
// Карточка 1
pos1 := LexorankBetween("", "")  // "n"

// Карточка 2 (в конец)
pos2 := LexorankBetween("n", "")  // "nn"

// Карточка 3 (в конец)
pos3 := LexorankBetween("nn", "")  // "nnn"

// Сортировка: "n" < "nn" < "nnn"
```

**Визуализация:**
```
┌─────┬─────┬─────┐
│  n  │ nn  │ nnn │
└─────┴─────┴─────┘
```

### Пример 2: Вставка в середину

```go
// Исходный список: "a", "c"
// Вставить между ними:
pos := LexorankBetween("a", "c")  // "b"

// Список: "a" < "b" < "c"
```

**Визуализация:**
```
Before:  ┌─────┬─────┐
         │  a  │  c  │
         └─────┴─────┘

After:   ┌─────┬─────┬─────┐
         │  a  │  b  │  c  │
         └─────┴─────┴─────┘
```

### Пример 3: Многократная вставка в одно место

```go
// Исходный список: "a", "b"
// Вставляем между ними 3 раза:

pos1 := LexorankBetween("a", "b")    // "am"
pos2 := LexorankBetween("a", "am")   // "ag"
pos3 := LexorankBetween("ag", "am")  // "ai"

// Список: "a" < "ag" < "ai" < "am" < "b"
```

**Визуализация:**
```
Step 1:  a  ───  b
         └─ am ─┘

Step 2:  a  ──────  am  ───  b
         └─ ag ───┘

Step 3:  a  ───  ag  ───  am  ───  b
                  └─ ai ─┘
```

### Пример 4: Real-world Trello-like сценарий

```
Колонка "To Do" (пустая):

1. Add Card A
   position = "n"

2. Add Card B (в конец)
   position = "nn"

3. Add Card C (в конец)
   position = "nnn"

4. Drag Card C между A и B
   position = LexorankBetween("n", "nn") = "nm"

Результат: "n" (A) < "nm" (C) < "nn" (B)
```

**SQL:**
```sql
-- Шаг 4: только один UPDATE!
UPDATE cards SET position = 'nm' WHERE id = 'card-c';
```

---

## Edge Cases и ограничения

### 1. Схлопывание (Collapse)

**Проблема:** При многократных вставках в одно место строка растет:
```
"a" → "am" → "ag" → "agm" → "agmm" → "agmmm" → ...
```

Теоретически может вырасти до очень длинной строки.

**Решение в Jira:**
- Периодический rebalancing (пересчет всех позиций)
- Триггерится когда длина position > 100 символов

**Решение в Yammi:**
- Пока не реализовано (TODO)
- VARCHAR(100) — достаточно для ~1000 вставок в одно место
- Если position > 90 символов → rebalance колонки (переназначить позиции равномерно)

**Частота проблемы:** Редкая (нужно 50+ вставок в одно место).

### 2. Concurrent вставка в одно место

**Сценарий:**
```
User A: Drag Card X между "a" и "b" → position = "am"
User B: Drag Card Y между "a" и "b" → position = "am"  (коллизия!)
```

**Проблема:** Обе карточки получат одинаковую позицию → порядок undefined.

**Решение:**
- После INSERT проверить нет ли коллизии:
  ```sql
  SELECT COUNT(*) FROM cards WHERE column_id = $1 AND position = $2;
  ```
- Если > 1 → пересчитать позицию (добавить микро-offset на основе ID):
  ```go
  newPosition = position + hash(cardID)[:2]  // "am" → "amx3"
  ```

**Частота проблемы:** Очень редкая (миллисекундное окно).

### 3. Миграция с INT на Lexorank

**Проблема:** У вас есть таблица с `position INT`, нужно перейти на lexorank.

**Решение:**
```sql
-- Добавляем новую колонку
ALTER TABLE cards ADD COLUMN position_lexorank VARCHAR(100);

-- Генерируем lexorank позиции на основе INT
UPDATE cards SET position_lexorank = (
    SELECT CASE
        WHEN position = 1 THEN 'a'
        WHEN position = 2 THEN 'b'
        WHEN position = 3 THEN 'c'
        -- ... до 26
        WHEN position > 26 THEN
            chr(96 + (position / 26)) || chr(96 + (position % 26))
    END
)
WHERE column_id = $1;

-- Переименовываем колонки
ALTER TABLE cards DROP COLUMN position;
ALTER TABLE cards RENAME COLUMN position_lexorank TO position;
```

**Yammi:** Изначально использует lexorank (миграция не нужна).

---

## Реализация в Yammi

### Код

**Файл:** `/home/roman/PetProject/yammi/services/board/internal/domain/lexorank.go`

**Основные функции:**

```go
// Вычислить позицию между prev и next
func LexorankBetween(prev, next string) (string, error)

// Валидация lexorank строки
func ValidateLexorank(s string) error

// Константы
const (
    LexorankFirst  = "n"   // Начальная позиция
    LexorankMiddle = "n"
    lexorankAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
)
```

### Тесты

**Файл:** `/home/roman/PetProject/yammi/services/board/internal/domain/lexorank_test.go`

**Покрытие:**
- Вставка в пустой список
- Вставка между элементами
- Вставка в начало
- Вставка в конец
- Граничные случаи (prev >= next → ошибка)
- Валидация (невалидные символы → ошибка)

### Использование в domain

**Создание карточки:**
```go
// internal/domain/card.go
func NewCard(columnID, title, description, position string, assigneeID *string) (*Card, error) {
    if err := ValidateLexorank(position); err != nil {
        return nil, err
    }
    // ...
}
```

**Перемещение карточки:**
```go
// internal/domain/card.go
func (c *Card) Move(targetColumnID, newPosition string) error {
    if err := ValidateLexorank(newPosition); err != nil {
        return err
    }
    c.Position = newPosition
    // ...
}
```

### Использование в usecase

**MoveCardUseCase:**
```go
// internal/usecase/card.go
func (uc *CardUseCase) MoveCard(ctx, cardID, targetColumnID, afterCardID string) error {
    // 1. Найти карточку после которой вставляем
    afterCard, _ := uc.cardRepo.GetByID(ctx, afterCardID)

    // 2. Найти карточку до которой вставляем (следующая в списке)
    nextCard, _ := uc.cardRepo.GetNextCard(ctx, targetColumnID, afterCard.Position)

    // 3. Вычислить новую позицию
    newPosition, err := domain.LexorankBetween(afterCard.Position, nextCard.Position)

    // 4. Переместить карточку
    card.Move(targetColumnID, newPosition)

    // 5. Сохранить (только 1 UPDATE!)
    return uc.cardRepo.Update(ctx, card)
}
```

---

## Сравнение: INT vs Lexorank

| Аспект | INT position | Lexorank position |
|--------|--------------|-------------------|
| **UPDATE при reorder** | O(n) строк | O(1) строк |
| **Locks** | Блокирует много строк | Блокирует только 1 строку |
| **Concurrent safety** | Race conditions | Safe (каждая карточка независима) |
| **Сложность алгоритма** | Простой (position + 1) | Средняя (midpoint calculation) |
| **Storage** | 4 bytes (INT) | ~5-20 bytes (VARCHAR) |
| **Миграция** | Простая | Нужна конвертация |
| **Collapse problem** | Нет | Редкая (нужен rebalance) |

**Вывод:**
- **INT:** Использовать для редко изменяемых списков (например, Columns)
- **Lexorank:** Использовать для часто изменяемых списков (например, Cards)

---

## FAQ

**Q: Почему не использовать FLOAT position?**

A: FLOAT имеет ограниченную точность (53 бита). При многократных вставках между элементами теряется точность:
```
1.0, 2.0 → insert → 1.5
1.0, 1.5 → insert → 1.25
1.0, 1.25 → insert → 1.125
...
1.0000000001, 1.0000000002 → insert → precision loss!
```

Lexorank не имеет этой проблемы (строка может расти бесконечно).

**Q: Почему не использовать DECIMAL?**

A: DECIMAL работает, но:
- Менее читаем (`"0.123456789"` vs `"am"`)
- Больше storage (15 bytes vs 5 bytes)
- Сложнее вычислять midpoint с произвольной точностью

**Q: Можно ли использовать UUID для position?**

A: Нет. UUID не имеет порядка (не сортируется осмысленно).

**Q: Зачем алфавит включает цифры (0-9)?**

A: Больше base (36 вместо 26) → меньше длина строк:
- Base-26: `"aa"` = 26 позиций
- Base-36: `"aa"` = 36 позиций (на 38% эффективнее)

**Q: Что если пользователь перетаскивает 100 карточек за раз (bulk move)?**

A: Два подхода:
1. **Naive:** 100 вызовов `LexorankBetween` + 100 UPDATEs
2. **Optimized:** Rebalance целевой колонки (равномерно распределить все позиции):
   ```go
   positions := GenerateEvenlySpaced(numCards)  // ["a", "b", "c", ...]
   for i, card := range cards {
       card.Position = positions[i]
   }
   ```

---

## Ссылки

- [Jira Lexorank алгоритм (Atlassian)](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-rest-api-3-issue-issueidorkey-put)
- [Fractional indexing (Figma engineering blog)](https://www.figma.com/blog/realtime-editing-of-ordered-sequences/)
- [Implementing lexorank in Go](https://github.com/kvl-ballista/lexorank)

---

## Итого

**Lexorank — это:**
- ✅ O(1) reordering (вместо O(n))
- ✅ Concurrent-safe
- ✅ No massive UPDATEs
- ⚠️ Требует периодического rebalancing (при длинных строках)
- ⚠️ Чуть сложнее в реализации

**Используется в Yammi для:**
- `cards.position` (VARCHAR(100))

**НЕ используется для:**
- `columns.position` (INT) — reorder редкий, колонок мало
