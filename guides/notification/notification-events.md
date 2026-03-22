# Notification Service — Карта событий NATS

> Детальное описание всех событий, которые потребляет и публикует Notification Service.

---

## Потребляемые события (13 типов)

### Потоки (Streams)

| Stream | Subjects | Retention |
|--------|----------|-----------|
| `USERS` | `user.>` | 7 дней |
| `BOARDS` | `board.>`, `column.>`, `card.>`, `member.>` | 7 дней |

---

### User Events

#### `user.created`

**Когда:** Новый пользователь зарегистрировался в Auth Service.

**Payload:**
```json
{
  "event_id": "uuid",
  "user_id": "uuid",
  "name": "Иван Петров",
  "occurred_at": "2026-03-21T12:00:00Z"
}
```

**Действие:** Создаёт welcome-уведомление: _"Добро пожаловать в Yammi!"_

**Кто получает:** Сам зарегистрировавшийся пользователь.

---

### Board Events

#### `board.created`

**Payload:** `{ board_id, title, owner_id, occurred_at }`

**Действие:**
1. Кешируется имя доски (`board_names`)
2. Владелец добавляется в `board_members`
3. Уведомление владельцу: _"Доска 'X' создана"_

#### `board.updated`

**Payload:** `{ board_id, title, actor_id, occurred_at }`

**Действие:**
1. Обновляется кеш имени доски
2. Уведомление всем участникам (кроме `actor_id`): _"Доска 'X' обновлена"_

#### `board.deleted`

**Payload:** `{ board_id, actor_id, occurred_at }`

**Действие:**
1. Уведомление всем участникам (кроме `actor_id`): _"Доска 'X' удалена"_
2. Очищается `board_members` для этой доски
3. Очищается `board_names`

---

### Column Events

#### `column.created` / `column.updated` / `column.deleted`

**Payload:** `{ column_id, board_id, title, actor_id, occurred_at }`

**Действие:**
- Кешируется/обновляется/удаляется имя колонки
- Уведомление участникам доски: _"Колонка 'Y' создана/обновлена/удалена в доске 'X'"_

---

### Card Events

#### `card.created` / `card.updated`

**Payload:** `{ card_id, column_id, board_id, title, actor_id, occurred_at }`

**Действие:**
- Кешируется/обновляется имя карточки
- Уведомление участникам доски: _"Карточка 'Z' создана/обновлена в доске 'X'"_

#### `card.moved`

**Payload:** `{ card_id, board_id, source_column_id, target_column_id, actor_id, occurred_at }`

**Действие:** Уведомление участникам: _"Карточка 'Z' перемещена в доске 'X'"_

#### `card.deleted`

**Payload:** `{ card_id, board_id, actor_id, occurred_at }`

**Действие:**
- Уведомление участникам: _"Карточка 'Z' удалена в доске 'X'"_
- Очищается `card_names`

---

### Member Events

#### `member.added`

**Payload:** `{ board_id, user_id, board_title, role, occurred_at }`

**Действие:**
1. Пользователь добавляется в `board_members`
2. Уведомление **добавленному**: _"Вы добавлены в доску 'X'"_ с ролью

#### `member.removed`

**Payload:** `{ board_id, user_id, board_title, occurred_at }`

**Действие:**
1. Уведомление **удалённому**: _"Вы удалены из доски 'X'"_
2. Пользователь удаляется из `board_members`

---

## Публикуемые события

### `notification.created`

**Stream:** `NOTIFICATIONS` (subjects: `notification.>`, retention 7 дней)

**Когда:** После успешного создания уведомления в БД.

**Payload:**
```json
{
  "event_id": "uuid",
  "occurred_at": "2026-03-21T12:00:00Z",
  "id": "notification-uuid",
  "user_id": "target-user-uuid",
  "type": "card_moved",
  "title": "Карточка 'Задача' перемещена",
  "message": "",
  "metadata": {
    "board_id": "uuid",
    "card_id": "uuid",
    "actor_name": "Иван"
  }
}
```

**Потребитель:** WebSocket Gateway (`services/gateway`) — пушит real-time toast и обновляет счётчик непрочитанных у подключённого клиента.

---

## Обработка ошибок

### Retry с exponential backoff

```
Попытка 1 → ошибка → NAK + delay 2s (±20% jitter)
Попытка 2 → ошибка → NAK + delay 4s
Попытка 3 → ошибка → NAK + delay 8s
...
Попытка 7 → ошибка → DLQ
```

### Dead Letter Queue (DLQ)

Сообщения отправляются в DLQ (`dlq.<original_subject>`) в двух случаях:
1. **Poison message** — невалидный JSON, невозможно распарсить
2. **Max retries** — 7 попыток исчерпаны

DLQ envelope:
```json
{
  "original_subject": "card.moved",
  "consumer_name": "notification-service-card-moved-v2",
  "error": "connection refused",
  "num_delivered": 7,
  "payload": "{...original message...}",
  "failed_at": "2026-03-21T12:00:00Z"
}
```
