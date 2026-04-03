# Поиск и фильтрация карточек на доске

> Поиск по названию, фильтрация по исполнителю, типу задачи и приоритету. Работает мгновенно на клиенте (все карточки уже загружены) + доступен серверный API для внешних потребителей.

---

## Обзор

На доске над колонками отображается панель фильтров:

```
[Поиск по названию...] [Avatar1 Avatar2 ... +N] [Тип задачи v] [Приоритет v] [Сбросить]
```

**Элементы панели фильтров:**

| Элемент | Описание |
|---------|----------|
| Поиск | Текстовый input с debounce 250ms. Фильтрует по `card.title` (case-insensitive) |
| Аватарки исполнителей | Максимум 5 видимых аватарок участников доски. При наведении — имя пользователя (tooltip). Клик = фильтр по `assignee_id`. Повторный клик снимает фильтр |
| Кнопка +N | Если участников > 5 — кнопка с количеством скрытых. Клик открывает dropdown со списком остальных участников (аватарка + имя) |
| Тип задачи | Выпадающий список: Все типы / Баг / Фича / Задача / Улучшение. По умолчанию — "Все типы" |
| Приоритет | Выпадающий список: Все приоритеты / Критический / Высокий / Средний / Низкий. По умолчанию — "Все приоритеты" |
| Сбросить | Появляется только при наличии активных фильтров. Сбрасывает все фильтры |

**Все фильтры применяются одновременно (AND-логика).**

---

## Frontend: клиентская фильтрация

Карточки уже загружены в store при открытии доски (`fetchBoard` → `getCards` для каждой колонки). Фильтрация происходит на клиенте без сетевых запросов — мгновенный отклик.

### Компоненты

- **`BoardFilterBar.vue`** — панель фильтров. Принимает `members` (список участников доски), эмитит `update:filters` с объектом `BoardFilters`
- **`BoardPage.vue`** — получает фильтры, вычисляет `hiddenCardIds` (Set) и `filteredCardCounts` (Map), передаёт в `BoardColumn`
- **`BoardColumn.vue`** — скрывает карточки через `v-show` (не удаляет из DOM — drag-and-drop продолжает работать)

### Логика фильтрации

```typescript
interface BoardFilters {
  search: string      // подстрока в title (case-insensitive)
  assigneeId: string  // UUID исполнителя или '' для всех
  priority: string    // low|medium|high|critical или '' для всех
  taskType: string    // bug|feature|task|improvement или '' для всех
}

function cardMatchesFilter(card: Card): boolean {
  if (f.search && !card.title.toLowerCase().includes(f.search.toLowerCase())) return false
  if (f.assigneeId && card.assigneeId !== f.assigneeId) return false
  if (f.priority && card.priority !== f.priority) return false
  if (f.taskType && card.taskType !== f.taskType) return false
  return true
}
```

### Почему v-show, а не v-if / фильтрация массива?

Vuedraggable (`v-model="column.cards"`) управляет массивом карточек для drag-and-drop. Если мы подменим массив отфильтрованными данными:
1. Drag-and-drop сломается (перетаскивание обновит отфильтрованный массив, не оригинальный)
2. Потеряются скрытые карточки из реактивного состояния

Вместо этого: храним полный массив, скрываем карточки через `v-show`. Vuedraggable видит все элементы, пользователь видит только отфильтрованные.

### Счётчик карточек

При активных фильтрах в заголовке колонки отображается количество **видимых** карточек (не общее). Это вычисляется через `filteredCardCounts` Map в BoardPage.

---

## Backend: серверный API поиска

### Эндпоинт

```
GET /api/v1/boards/{id}/cards/search
```

### Query-параметры

| Параметр | Тип | Обязательный | Описание |
|----------|-----|-------------|----------|
| `search` | string | Нет | Подстрока в title (ILIKE, max 200 символов) |
| `assignee_id` | UUID | Нет | Фильтр по исполнителю |
| `priority` | string | Нет | Фильтр по приоритету (low/medium/high/critical) |
| `task_type` | string | Нет | Фильтр по типу (bug/feature/task/improvement) |

### Ответ

```json
{
  "cards": [
    {
      "id": "...",
      "column_id": "...",
      "title": "...",
      "description": "...",
      "position": "am",
      "assignee_id": "...",
      "creator_id": "...",
      "priority": "high",
      "task_type": "bug",
      "due_date": "2026-04-15T00:00:00Z",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

### Авторизация

- Вызывающий пользователь должен быть **участником доски** (member или owner)
- Проверка через `MembershipRepository.IsMember` (Redis-кеш)

### SQL-запрос

Запрос строится динамически с параметрами:

```sql
SELECT id, column_id, title, description, position, assignee_id, creator_id,
       due_date, priority, task_type, created_at, updated_at
FROM cards
WHERE board_id = $1
  [AND title ILIKE '%' || $N || '%' ESCAPE '\']
  [AND assignee_id = $N]
  [AND priority = $N]
  [AND task_type = $N]
ORDER BY position ASC
```

- `board_id` всегда в WHERE → partition pruning (таблица партиционирована по `board_id`)
- Поиск через ILIKE с `ESCAPE '\'` — защита от SQL LIKE injection (`%`, `_`, `\` экранируются)
- Все параметры передаются через placeholders — защита от SQL injection
- Валидация priority/task_type через allowlist в usecase (невалидные значения игнорируются)

### Производительность

- **Индексы**: `idx_cards_assignee_id`, `idx_cards_priority` уже существуют
- **Партиционирование**: 4 hash-партиции по `board_id` — запрос попадает только в одну
- **Членство**: проверяется через Redis-кеш (`board_roles:{boardID}` HASH) — O(1) микросекунды
- **N+1 защита**: один SQL-запрос возвращает все карточки доски, сервис User не вызывается

### Proto (gRPC)

```protobuf
rpc SearchBoardCards(SearchBoardCardsRequest) returns (SearchBoardCardsResponse);

message SearchBoardCardsRequest {
  string board_id = 1;
  string user_id = 2;
  string search = 3;
  string assignee_id = 4;
  string priority = 5;
  string task_type = 6;
}

message SearchBoardCardsResponse {
  repeated Card cards = 1;
}
```

### Архитектурный путь запроса

```
HTTP GET /boards/{id}/cards/search
  → API Gateway: SearchBoardCards (валидация длин, JWT)
    → gRPC: BoardService.SearchBoardCards
      → UseCase: SearchBoardCardsUseCase.Execute
        → Redis: IsMember (кеш)
        → PostgreSQL: CardRepository.SearchByBoardID (один запрос)
      ← []*domain.Card
    ← SearchBoardCardsResponse (proto)
  ← JSON { "cards": [...] }
```

---

## Визуальные изменения

### Ширина колонок

Колонки расширены с 300px до 360px для лучшей читаемости карточек с фильтрами:

```css
.board-column {
  min-width: 360px;
  max-width: 360px;
}
```
