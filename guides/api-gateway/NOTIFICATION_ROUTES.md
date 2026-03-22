# Notification Service API Routes

Добавлены следующие HTTP маршруты в API Gateway для взаимодействия с Notification Service.

## Конфигурация

Переменная окружения:
- `NOTIFICATION_GRPC_ADDR` - адрес Notification Service (по умолчанию: `localhost:50055`)

## Аутентификация

Все маршруты требуют JWT токен в заголовке `Authorization: Bearer <token>`.
`user_id` извлекается из JWT автоматически.

---

## Notification Routes

### Список уведомлений
```
GET /api/v1/notifications?limit=20&cursor=&type=card&search=задача

Response 200:
{
  "notifications": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "type": "card_moved",
      "title": "Карточка \"Задача\" перемещена в доске \"Спринт 1\"",
      "message": "",
      "metadata": {
        "board_id": "uuid",
        "card_id": "uuid",
        "actor_name": "Иван"
      },
      "is_read": false,
      "created_at": "2026-03-21T12:00:00Z"
    }
  ],
  "next_cursor": "2026-03-21T11:59:00.123456789Z",
  "total_unread": 5
}
```

**Query Parameters:**

| Параметр | Тип | Default | Описание |
|----------|-----|---------|----------|
| `limit` | int | 20 | Количество (max 100) |
| `cursor` | string | — | Курсор пагинации (из `next_cursor` предыдущего ответа) |
| `type` | string | — | Фильтр по типу (префикс: `card` → `card_created`, `card_moved`, ...) |
| `search` | string | — | Поиск по заголовку (ILIKE) |

Если `next_cursor` пустой — записей больше нет.

### Пометить прочитанными
```
POST /api/v1/notifications/read
Content-Type: application/json

{
  "notification_ids": ["uuid-1", "uuid-2"]
}

Response 200: {}
```

### Пометить все прочитанными
```
POST /api/v1/notifications/read-all

Response 200: {}
```

### Количество непрочитанных
```
GET /api/v1/notifications/unread-count

Response 200:
{
  "count": 5
}
```

### Получить настройки
```
GET /api/v1/notifications/settings

Response 200:
{
  "settings": {
    "user_id": "uuid",
    "enabled": true,
    "realtime_enabled": true
  }
}
```

### Обновить настройки
```
PUT /api/v1/notifications/settings
Content-Type: application/json

{
  "enabled": true,
  "realtime_enabled": false
}

Response 200:
{
  "settings": {
    "user_id": "uuid",
    "enabled": true,
    "realtime_enabled": false
  }
}
```

---

## Сводная таблица

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/v1/notifications` | Список уведомлений (cursor pagination) |
| POST | `/api/v1/notifications/read` | Пометить выбранные прочитанными |
| POST | `/api/v1/notifications/read-all` | Пометить все прочитанными |
| GET | `/api/v1/notifications/unread-count` | Счётчик непрочитанных |
| GET | `/api/v1/notifications/settings` | Получить настройки |
| PUT | `/api/v1/notifications/settings` | Обновить настройки |
