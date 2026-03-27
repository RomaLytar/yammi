 # Board Settings, Global Labels, Automation Engine, Templates

> Новые фичи Board Service: глобальные метки, настройки доски, движок автоматизации, шаблоны

---

## Глобальные метки (User Labels)

Метки, которые видны на **всех досках** пользователя.

### Архитектура
- Хранятся в таблице `user_labels` (Board Service DB, НЕ User Service — чтобы избежать cross-service вызовов)
- Scope: `user_id` (создатель)
- Max 50 глобальных меток на пользователя
- При загрузке доски: `ListAvailableLabels` мержит board labels + owner's user_labels
- **При привязке к карточке**: если метка глобальная — автоматически копируется в `labels` доски с тем же UUID, затем прикрепляется. Это позволяет toggle attach/detach работать по одному ID.

### Доступ из UI
- Иконка тега в **AppHeader** (рядом с переключателем темы) — `GlobalLabelsModal.vue`
- Доступно с любой страницы, без привязки к доске

### API (7 новых routes)
```
POST   /api/v1/user-labels              — создать глобальную метку
GET    /api/v1/user-labels              — список глобальных меток
PUT    /api/v1/user-labels/:id          — обновить
DELETE /api/v1/user-labels/:id          — удалить
GET    /api/v1/boards/:id/settings      — получить настройки доски
PUT    /api/v1/boards/:id/settings      — обновить настройки (owner only)
GET    /api/v1/boards/:id/available-labels — мерж board + global labels
```

### Ответ available-labels API
```json
{
  "board_labels": [...],
  "user_labels": [...],
  "use_board_labels_only": false
}
```
Ключ — `user_labels` (НЕ `global_labels`). Frontend маппит `data.user_labels` → `globalLabels`.

### Таблицы (Migration 000014)
```sql
board_settings (board_id PK, use_board_labels_only BOOLEAN, created_at, updated_at)
user_labels (id PK, user_id, name, color, created_at, UNIQUE(user_id, name))
```

### Frontend
- `GlobalLabelsModal.vue` — модалка CRUD глобальных меток (вызывается из AppHeader)
- `BoardSettingsPage.vue` — страница настроек доски (4 вкладки: Метки доски, Участники, Автоматизация, Настройки)
- Метки доски: CRUD с цветовым пикером (16 цветов)
- Toggle: "Использовать только метки этой доски" (скрывает глобальные метки в пикере карточек)
- `allAvailableLabels` computed в store — мержит board + global, респектит настройку
- Глобальные метки НЕ привязаны к доске — управляются отдельно через AppHeader

---

## Board Settings

Настройки конкретной доски, доступные через `/boards/:boardId/settings` (иконка шестерёнки в шапке доски, только owner).

### Текущие настройки
| Настройка | Тип | Default | Описание |
|-----------|-----|---------|----------|
| `use_board_labels_only` | boolean | false | Если true, глобальные метки не показываются |

### Access Control
- GET: member
- PUT: owner only

### Lazy creation
Запись в `board_settings` создаётся при первом UPDATE (Upsert). GET возвращает дефолты если записи нет.

---

## Automation Execution Engine

Автоматическое выполнение правил при перемещении и создании карточек.

### Как работает
1. Пользователь перемещает карточку в колонку **или** создаёт карточку
2. `MoveCardUseCase` / `CreateCardUseCase` завершает операцию
3. **Async goroutine** (10s timeout) вызывает `ExecuteAutomationsUseCase`
4. Загружает matching правила: `ListEnabledByBoardAndTrigger(boardID, triggerType)`
5. Для каждого правила проверяет `triggerConfig["column_id"]` — если указан, матчит только для этой колонки; если не указан — матчит любую
6. Выполняет действие (assign_member, add_label, set_priority)
7. Записывает результат в `automation_executions` (success/failed)

### Поддерживаемые триггеры (UI)
| Trigger | Что проверяет | Config |
|---------|-------------|--------|
| `card_moved_to_column` | Карточка перемещена в колонку | `{"column_id": "..."}` (опционально) |
| `card_created` | Карточка создана в колонке | `{"column_id": "..."}` (опционально) |

Триггеры `label_added`, `due_date_passed`, `checklist_completed` определены в домене, но убраны из UI — не реализованы и не полезны.

### Поддерживаемые действия
| Action | Что делает | Config |
|--------|-----------|--------|
| `assign_member` | Устанавливает assignee на карточку | `{"user_id": "..."}` |
| `add_label` | Добавляет метку на карточку | `{"label_id": "..."}` |
| `set_priority` | Устанавливает приоритет | `{"priority": "high"}` |
| `move_card` | Пропускается (защита от рекурсии) | — |

### Partial Update
`UpdateAutomationRule` поддерживает partial update: если `name` пустой — сохраняется текущее значение. Это позволяет toggle enabled/disabled без передачи всех полей.

### Non-blocking
- Ошибки логируются через `slog.Error`, НЕ пропагируются в MoveCard/CreateCard
- `automationExecutor` nilable — если nil, автоматизация пропускается

### UI — вкладка "Автоматизация" в настройках доски
- Список правил с toggle вкл/выкл
- Создание: выбор триггера → колонки → действия → параметра (участник/метка/приоритет/колонка)
- Редактирование и удаление (owner only, кнопки видны при hover)
- В действии "Добавить метку" показываются `allAvailableLabels` (board + global), глобальные помечены "(глобальная)"

### Тесты
- 7 unit тестов (matching, non-matching, disabled, actions, failures)
- 3 интеграционных теста (end-to-end с реальной PostgreSQL)

---

## Templates

Три типа шаблонов: карточки, колонки, доски.

### Таблицы (Migration 000015)
```sql
card_templates    (id, board_id?, user_id, name, title, description, priority, task_type, checklist_data JSONB, label_ids UUID[], created_at, updated_at)
column_templates  (id, board_id?, user_id, name, columns_data JSONB, created_at, updated_at)
board_templates   (id, user_id, name, description, columns_data JSONB, labels_data JSONB, created_at, updated_at)
```

### API (12 новых routes)
```
POST   /api/v1/boards/:id/card-templates           — создать шаблон карточки
GET    /api/v1/boards/:id/card-templates           — список шаблонов
DELETE /api/v1/card-templates/:id                   — удалить
POST   /api/v1/boards/:boardId/cards/from-template — создать карточку из шаблона

POST   /api/v1/boards/:id/column-templates         — создать шаблон колонок
GET    /api/v1/boards/:id/column-templates         — список
DELETE /api/v1/column-templates/:id                 — удалить
POST   /api/v1/boards/:boardId/columns/from-template — создать колонки из шаблона

POST   /api/v1/board-templates                      — создать шаблон доски
GET    /api/v1/board-templates                      — список
DELETE /api/v1/board-templates/:id                   — удалить
POST   /api/v1/boards/from-template                 — создать доску из шаблона
```

### JSONB структуры
```json
// checklist_data (card_templates)
[{"title": "QA Checklist", "items": ["Unit tests", "Manual test", "Code review"]}]

// columns_data (column_templates, board_templates)
[{"title": "Backlog", "position": 0}, {"title": "In Progress", "position": 1}]

// labels_data (board_templates)
[{"name": "Bug", "color": "#ef4444"}, {"name": "Feature", "color": "#10b981"}]
```

### Create from template flow
**Card**: загружает шаблон → создаёт карточку → создаёт чек-листы + items → привязывает метки
**Columns**: загружает шаблон → создаёт колонки с позициями
**Board**: загружает шаблон → создаёт доску → добавляет owner как member → создаёт колонки → создаёт метки

### Frontend
- `TemplateManager.vue` — управление шаблонами (карточки, колонки, доски)
- CreateCardModal: dropdown "Шаблон" заполняет форму из шаблона
- EditCardModal: кнопка "Сохранить как шаблон"
- CreateBoardModal: выбор шаблона доски, эмитит `createFromTemplate` → `BoardListPage` обрабатывает через `boardsApi.createBoardFromTemplate()`

### Тесты
- 15 domain тестов (конструкторы, валидация)
- 26 usecase тестов (CRUD, access control, create-from-template)

---

## Кеширование

| Данные | Кеш | Комментарий |
|--------|-----|-------------|
| Board settings | Pinia store | Загружается при `fetchBoard()` |
| User labels | Pinia store | Через `fetchAvailableLabels()` |
| Card templates | Pinia store | Через `fetchCardTemplates()` |
| Automation rules | Нет кеша | Запрашиваются только при move card |
| Board/Column templates | On-demand | Загружаются в модалках |

Новый Redis кеш не нужен — все данные low-frequency.

---

## Миграции
- `000014_board_settings_user_labels.up.sql` — board_settings + user_labels
- `000015_templates.up.sql` — card_templates + column_templates + board_templates
