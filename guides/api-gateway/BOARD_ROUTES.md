# Board Service API Routes

Добавлены следующие HTTP маршруты в API Gateway для взаимодействия с Board Service.

## Конфигурация

Переменная окружения:
- `BOARD_GRPC_ADDR` - адрес Board Service (по умолчанию: `localhost:50053`)

## Аутентификация

Все маршруты требуют JWT токен в заголовке `Authorization: Bearer <token>`.

## Board Routes

### Создать доску
```
POST /api/v1/boards
Content-Type: application/json

{
  "title": "My Board",
  "description": "Board description"
}

Response 201:
{
  "board": {
    "id": "uuid",
    "title": "My Board",
    "description": "Board description",
    "owner_id": "user-uuid",
    "version": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Получить доску
```
GET /api/v1/boards/{id}

Response 200:
{
  "board": { ... },
  "columns": [ ... ],
  "members": [ ... ]
}
```

### Список досок
```
GET /api/v1/boards?limit=20&cursor=xxx&owner_only=true&search=keyword&sort_by=updated_at

Query params:
  - limit      (int)    — кол-во результатов (по умолчанию 20)
  - cursor     (string) — курсор пагинации
  - owner_only (bool)   — если true, вернуть только доски, где пользователь является Owner
  - search     (string) — поиск по title (ILIKE, регистронезависимый)
  - sort_by    (string) — поле сортировки: updated_at | created_at | title (по умолчанию updated_at)

Response 200:
{
  "boards": [ ... ],
  "next_cursor": "xxx"
}
```

### Обновить доску
```
PUT /api/v1/boards/{id}
Content-Type: application/json

{
  "title": "Updated Title",
  "description": "Updated description",
  "version": 1
}

Response 200:
{
  "board": { ... }
}
```

### Удалить доски (batch)
```
POST /api/v1/boards/delete
Content-Type: application/json

{
  "board_ids": ["board-uuid-1", "board-uuid-2"]
}

Response 200:
{
  "status": "deleted"
}
```

## Column Routes

### Добавить колонку
```
POST /api/v1/boards/{id}/columns
Content-Type: application/json

{
  "title": "To Do",
  "position": 0
}

Response 201:
{
  "column": {
    "id": "uuid",
    "board_id": "board-uuid",
    "title": "To Do",
    "position": 0,
    "version": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Получить колонки
```
GET /api/v1/boards/{id}/columns

Response 200:
{
  "columns": [ ... ]
}
```

### Обновить колонку
```
PUT /api/v1/columns/{id}
Content-Type: application/json

{
  "board_id": "board-uuid",
  "title": "Updated Title",
  "version": 1
}

Response 200:
{
  "column": { ... }
}
```

### Переупорядочить колонки
```
PUT /api/v1/boards/{id}/columns/reorder
Content-Type: application/json

{
  "positions": [
    {"column_id": "uuid1", "position": 0},
    {"column_id": "uuid2", "position": 1}
  ],
  "version": 1
}

Response 200:
{
  "columns": [ ... ]
}
```

### Удалить колонку
```
DELETE /api/v1/columns/{id}
Content-Type: application/json

{
  "board_id": "board-uuid"
}

Response 200:
{
  "status": "deleted"
}
```

## Card Routes

### Создать карточку
```
POST /api/v1/columns/{id}/cards
Content-Type: application/json

{
  "board_id": "board-uuid",
  "title": "Task title",
  "description": "Task description",
  "position": "a"  // lexorank string
}

Response 201:
{
  "card": {
    "id": "uuid",
    "column_id": "column-uuid",
    "board_id": "board-uuid",
    "title": "Task title",
    "description": "Task description",
    "position": "a",
    "assignee_id": "",
    "creator_id": "user-uuid",
    "version": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Получить карточку
```
GET /api/v1/cards/{id}?board_id=board-uuid

Response 200:
{
  "card": { ... }
}
```

### Получить карточки колонки
```
GET /api/v1/columns/{id}/cards?board_id=board-uuid

Response 200:
{
  "cards": [ ... ]
}
```

### Обновить карточку
```
PUT /api/v1/cards/{id}
Content-Type: application/json

{
  "board_id": "board-uuid",
  "title": "Updated title",
  "description": "Updated description",
  "assignee_id": "user-uuid",
  "version": 1
}

Response 200:
{
  "card": { ... }
}
```

### Переместить карточку
```
PUT /api/v1/cards/{id}/move
Content-Type: application/json

{
  "board_id": "board-uuid",
  "from_column_id": "column-uuid-1",
  "to_column_id": "column-uuid-2",
  "position": "b",  // lexorank string
  "version": 1
}

Response 200:
{
  "card": { ... },
  "cards_in_column": [ ... ]
}
```

### Удалить карточки (batch)
```
POST /api/v1/cards/delete
Content-Type: application/json

{
  "card_ids": ["card-uuid-1", "card-uuid-2"],
  "board_id": "board-uuid"
}

Response 200:
{
  "status": "deleted"
}
```

## Member Routes

### Добавить участника
```
POST /api/v1/boards/{id}/members
Content-Type: application/json

{
  "user_id": "user-uuid",
  "role": "member"  // "owner" or "member"
}

Response 201:
{
  "member": {
    "user_id": "user-uuid",
    "role": "member",
    "version": 1,
    "joined_at": "2024-01-01T00:00:00Z"
  }
}
```

### Удалить участника
```
DELETE /api/v1/boards/{boardId}/members/{userId}

Response 200:
{
  "status": "removed"
}
```

### Список участников
```
GET /api/v1/boards/{id}/members

Response 200:
{
  "members": [ ... ]
}
```

## User Routes

### Поиск пользователей
```
GET /api/v1/users/search?q=email-fragment

Query params:
  - q (string) — поисковый запрос по email (частичное совпадение, ILIKE + pg_trgm)

Response 200:
{
  "users": [
    {
      "id": "user-uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "avatar_url": "https://..."
    }
  ]
}
```

## Права доступа (Permissions)

| Действие | Owner | Member | Non-member |
|---|---|---|---|
| Удаление досок | Да | Нет | Нет |
| Удаление любой карточки | Да | Нет | Нет |
| Удаление своей карточки (по `creator_id`) | Да | Да | Нет |
| Удаление колонок | Да | Нет | Нет |
| Управление участниками (добавление/удаление) | Да | Нет | Нет |
| Создание/обновление/перемещение карточек | Да | Да | Нет |
| Просмотр доски | Да | Да | Нет |

- **Owner** — полный доступ: удаление досок, удаление любых карточек, удаление колонок, управление участниками.
- **Member** — создание, обновление и перемещение карточек. Удаление только СВОИХ карточек (определяется по `creator_id`). Не может удалять колонки.
- **Non-member** — доступ запрещён.

## Примечания

- Все временные метки в формате ISO8601 (RFC3339)
- JSON использует snake_case для полей
- Все ответы содержат `Content-Type: application/json`
- Rate limiting: 50 req/min (настраивается через `RATE_LIMIT_DEFAULT`)
- Position для карточек использует lexorank (строка: "a", "am", "b", и т.д.)
- Version используется для optimistic locking
- Все маршруты защищены JWT аутентификацией и rate limiting
- `board.updated_at` обновляется автоматически при любом изменении карточек или колонок (используется для `sort_by=updated_at`)
- Все ответы, содержащие объект Card, включают поле `creator_id` — UUID пользователя, создавшего карточку
