# Releases (Board Service)

> Jira-like система релизов/спринтов. Релизы привязаны к доскам — карточки назначаются в релизы, канбан показывает только карточки активного релиза. Карточки без релиза живут в бэклоге. Функционал включается через настройку доски.

---

## Обзор

Релиз — board-scoped сущность с жизненным циклом: **draft → active → completed**.

**Ключевые правила:**
- Функционал включается в настройках доски (`releases_enabled`)
- Когда выключен — табы Релизы/Бэклог скрыты, все карточки видны на доске
- Когда включён — на доске видны только карточки активного релиза; если нет активного — доска пустая
- Карточки назначаются в релиз или остаются в бэклоге (`release_id IS NULL`)
- Только **один активный** релиз на доску (partial unique index в PostgreSQL)
- **Завершённые релизы** — read-only (нельзя редактировать, удалять/добавлять карточки)
- При завершении: карточки не в done-колонке остаются без релиза (бэклог), карточки в done-колонке остаются в релизе
- Done-колонка настраивается в `board_settings.done_column_id`
- Длительность спринта настраивается в `board_settings.sprint_duration_days` (по умолчанию 14, минимум 7)
- При старте релиза `start_date` и `end_date` вычисляются автоматически

---

## Жизненный цикл

```
┌─────────┐   Start()   ┌──────────┐   Complete()   ┌─────────────┐
│  DRAFT  │ ──────────→  │  ACTIVE  │  ──────────→   │  COMPLETED  │
└─────────┘              └──────────┘                └─────────────┘
     │                        │                           │
     │ Можно редактировать    │ Канбан показывает         │ Read-only
     │ Можно удалить          │   карточки релиза         │ Нельзя менять
     │ Можно назначать        │ start_date = now          │ Историческая
     │   карточки             │ end_date = now +          │   запись
     │                        │   sprint_duration_days    │
```

### Start Release
- Только owner может запустить
- Релиз должен быть в статусе `draft`
- На доске не должно быть другого активного релиза
- `start_date` = текущее время, `end_date` = start_date + sprint_duration_days из board_settings

### Complete Release
- Только owner может завершить
- Релиз должен быть в статусе `active`
- Если `done_column_id` настроен:
  - Карточки **в done-колонке** остаются в релизе
  - Карточки **не в done-колонке** → release_id = NULL (бэклог)
- Если `done_column_id` не настроен: все карточки → бэклог
- Карточки **не перемещаются между колонками** — меняется только `release_id`
- После завершения фронт перезагружает доску

---

## Настройки доски (board_settings)

| Поле | Тип | Описание |
|------|-----|----------|
| `releases_enabled` | bool | Включает/выключает функционал релизов (дефолт: false) |
| `done_column_id` | UUID? | Колонка "done" для проверки завершения релиза |
| `sprint_duration_days` | int | Длительность спринта в днях (дефолт: 14, мин: 7) |

---

## Права доступа

| Операция | Owner | Member |
|----------|-------|--------|
| Включить релизы в настройках | + | - |
| Создать релиз | + | + |
| Редактировать релиз | + | - |
| Удалить релиз | + | - |
| Запустить релиз | + | - |
| Завершить релиз | + | - |
| Назначить карточку в релиз | + | + |
| Убрать карточку из релиза | + | + |
| Просмотр релизов/бэклога | + | + |

---

## Database Schema

### releases table
```sql
CREATE TABLE releases (
    id UUID PRIMARY KEY,
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_by UUID NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_releases_board_active ON releases(board_id) WHERE status = 'active';
```

### cards (изменения)
```sql
ALTER TABLE cards ADD COLUMN release_id UUID;
```

### board_settings (изменения)
```sql
ALTER TABLE board_settings ADD COLUMN done_column_id UUID;
ALTER TABLE board_settings ADD COLUMN sprint_duration_days INT NOT NULL DEFAULT 14;
ALTER TABLE board_settings ADD COLUMN releases_enabled BOOLEAN NOT NULL DEFAULT false;
```

Миграции: `000016_releases`, `000017_release_dates`, `000018_sprint_duration`, `000019_releases_enabled`.

---

## API Endpoints (12 routes)

### CRUD
- `POST   /api/v1/boards/:boardId/releases` — создать релиз
- `GET    /api/v1/boards/:boardId/releases` — список релизов доски
- `GET    /api/v1/boards/:boardId/releases/:releaseId` — получить релиз
- `PUT    /api/v1/boards/:boardId/releases/:releaseId` — обновить релиз
- `DELETE /api/v1/boards/:boardId/releases/:releaseId` — удалить релиз

### Lifecycle
- `POST   /api/v1/boards/:boardId/releases/:releaseId/start` — запустить
- `POST   /api/v1/boards/:boardId/releases/:releaseId/complete` — завершить
- `GET    /api/v1/boards/:boardId/releases/active` — активный релиз

### Card Assignment
- `POST   /api/v1/boards/:boardId/releases/:releaseId/cards` — назначить карточку
- `DELETE /api/v1/boards/:boardId/releases/:releaseId/cards/:cardId` — убрать карточку
- `GET    /api/v1/boards/:boardId/releases/:releaseId/cards` — карточки релиза

### Backlog
- `GET    /api/v1/boards/:boardId/backlog` — карточки без релиза

---

## NATS Events

| Subject | Описание |
|---------|----------|
| `release.created` | Релиз создан |
| `release.updated` | Релиз обновлён |
| `release.started` | Релиз запущен (draft → active) |
| `release.completed` | Релиз завершён (active → completed) |
| `release.deleted` | Релиз удалён |
| `card.release_assigned` | Карточка назначена в релиз |
| `card.release_removed` | Карточка убрана из релиза |

---

## Frontend

### Навигация (табы внутри BoardPage)
- **Доска** (`?tab=board` или без query) — канбан
- **Релизы** (`?tab=releases`) — список релизов; `?tab=releases&release={id}` — задачи релиза
- **Бэклог** (`?tab=backlog`) — карточки без релиза

Табы видны только при `releases_enabled = true`. Браузерная кнопка "Назад" работает через URL query params.

### Активный релиз на доске
- Зелёный бэдж с именем релиза прямо в табе "Доска"
- Если нет активного релиза — доска пустая (колонки видны, карточек нет)

### Создание карточки
- Селектор релиза (`BaseSelect`) при создании и редактировании карточки
- Если нет активного релиза — toast-уведомление: "Карточка добавлена в бэклог"

### Общий компонент TaskTable
- Универсальная таблица задач используется в Релизах и Бэклоге
- Колонки: Тип | Задача | Статус | Приоритет | Исполнитель
- Цветные иконки типа, pill-бэдж статуса, аватарки исполнителей

### Компоненты
- `BaseSelect.vue` — универсальный кастомный dropdown (используется везде вместо `<select>`)
- `BoardSubNav.vue` — таб-бар Доска/Релизы/Бэклог
- `TaskTable.vue` — универсальная таблица задач
- `ReleasesPanel.vue` — список релизов + детали с задачами
- `BacklogPanel.vue` — бэклог с кнопкой назначения в релиз
- `ReleaseStatusBadge.vue` — бэдж статуса
- `CreateReleaseModal.vue` — модалка создания

---

## Grafana Dashboard

`deployments/monitoring/grafana/dashboards/release-metrics.json`

Панели: releases created/started/completed, events/sec, RPC rates, latency p95.
