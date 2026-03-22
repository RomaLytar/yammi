# Integration тесты Board Service

Комплексные integration тесты с использованием testcontainers для проверки работы Board Service с реальной PostgreSQL базой данных.

## Структура тестов

```
tests/integration/
├── setup_test.go                      # Setup testcontainers + миграции
├── board_repository_test.go           # Тесты BoardRepository
├── column_repository_test.go          # Тесты ColumnRepository
├── card_repository_test.go            # Тесты CardRepository (включая partitioning)
├── membership_repository_test.go      # Тесты MembershipRepository
├── usecase_test.go                    # Тесты Use Cases
└── README.md                          # Эта документация
```

## Зависимости

Добавьте в `go.mod`:

```go
require (
    github.com/testcontainers/testcontainers-go v0.26.0
)
```

Выполните:

```bash
cd services/board
go mod tidy
```

## Запуск тестов

### Все integration тесты:

```bash
cd services/board
go test ./tests/integration/... -v
```

### Отдельный тест:

```bash
cd services/board
go test ./tests/integration/ -run TestBoardRepository_Create -v
```

### С таймаутом:

```bash
cd services/board
go test ./tests/integration/... -v -timeout 10m
```

### С подробным выводом:

```bash
cd services/board
go test ./tests/integration/... -v -count=1
```

## Покрытие тестами

### BoardRepository
- ✅ Create (создание board + автоматическое добавление owner в members)
- ✅ GetByID (включая NotFound)
- ✅ Update (с проверкой version increment)
- ✅ Delete (включая каскадное удаление)
- ✅ OptimisticLocking (конкурентные обновления, конфликт версий)
- ✅ CursorPagination (ListByUserID с cursor-based pagination, 3 страницы)
- ✅ EmptyResult (пустой список для пользователя без досок)

### ColumnRepository
- ✅ Create (создание колонки)
- ✅ GetByID (включая NotFound)
- ✅ ListByBoardID (сортировка по position)
- ✅ Update (обновление title и position)
- ✅ Delete (удаление колонки)
- ✅ CascadeDelete (каскадное удаление при удалении board)

### CardRepository
- ✅ Create (создание card с assignee)
- ✅ CreateWithoutAssignee (создание card без assignee)
- ✅ GetByID (включая NotFound)
- ✅ ListByColumnID (сортировка по lexorank position)
- ✅ LexorankPositioning (тест лексикографической сортировки: a, am, b, c, m, z)
- ✅ Update (обновление title, description, assignee)
- ✅ Move (перемещение card между columns с новой position)
- ✅ Delete (удаление card)
- ✅ **Partitioning** (распределение 100 cards по 4 партициям, проверка tableoid)

### MembershipRepository
- ✅ AddMember (добавление member с ролью)
- ✅ AddMember_Duplicate (ошибка при дубликате)
- ✅ AddMember_InvalidRole (ошибка при невалидной роли)
- ✅ RemoveMember (удаление member)
- ✅ RemoveMember_NotFound (ошибка при удалении несуществующего member)
- ✅ RemoveMember_CannotRemoveOwner (запрет удаления owner)
- ✅ IsMember (проверка членства и роли)
- ✅ ListMembers (список members с pagination)
- ✅ ListMembers_Pagination (offset-based pagination)

### Use Cases
- ✅ CreateBoardUseCase (создание board + owner membership)
- ✅ GetBoardUseCase (получение board с проверкой access control)
- ✅ AddColumnUseCase (добавление column, проверка прав owner)
- ✅ CreateCardUseCase (создание card, проверка прав member)
- ✅ MoveCardUseCase (перемещение card между columns)
- ✅ AddMemberUseCase (добавление member, только owner)

## Особенности тестов

### Testcontainers
- Каждый тест запускает свой PostgreSQL 16 контейнер
- Автоматическое применение миграций из `migrations/000001_init.up.sql`
- Cleanup контейнера после теста

### Optimistic Locking
Тест `TestBoardRepository_OptimisticLocking` проверяет:
1. Загрузка board дважды (version 1)
2. Обновление первого экземпляра (version 1 → 2) ✅
3. Попытка обновления второго (version 1) → ErrInvalidVersion ❌

### Partitioning
Тест `TestCardRepository_Partitioning` проверяет:
1. Создание 100 cards для 10 boards
2. Распределение по 4 партициям (`cards_p0`, `cards_p1`, `cards_p2`, `cards_p3`)
3. Query: `SELECT tableoid::regclass, COUNT(*) FROM cards GROUP BY tableoid`
4. Верификация использования всех партиций

### Cursor Pagination
Тест `TestBoardRepository_CursorPagination` проверяет:
1. Создание 25 boards
2. Страница 1 (limit 10, cursor="") → 10 boards + cursor1
3. Страница 2 (limit 10, cursor=cursor1) → 10 boards + cursor2
4. Страница 3 (limit 10, cursor=cursor2) → 5 boards + cursor=""
5. Отсутствие дубликатов между страницами

### Lexorank
Тест `TestCardRepository_LexorankPositioning` проверяет:
1. Создание cards с позициями: "a", "am", "b", "c", "m", "z"
2. Загрузка с `ORDER BY position ASC`
3. Верификация лексикографического порядка

## CI/CD Integration

Для запуска в CI/CD убедитесь, что Docker доступен:

```yaml
# .github/workflows/test.yml
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:dind
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Run integration tests
        run: |
          cd services/board
          go test ./tests/integration/... -v -timeout 10m
```

## Troubleshooting

### Ошибка: "Cannot connect to Docker daemon"
Убедитесь, что Docker запущен:
```bash
docker ps
```

### Ошибка: "Failed to start container"
Увеличьте timeout в `setupPostgresContainer`:
```go
WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(120 * time.Second)
```

### Медленные тесты
Testcontainers запускает контейнер для каждого теста. Это нормально для integration тестов.
Для ускорения можно переиспользовать контейнер:
```go
// В setup_test.go создайте shared container с TestMain
func TestMain(m *testing.M) {
    // Setup shared container
    // Run tests
    // Cleanup
}
```

## Performance

Типичное время выполнения на локальной машине:
- BoardRepository: ~10s (7 тестов)
- ColumnRepository: ~8s (6 тестов)
- CardRepository: ~12s (9 тестов)
- MembershipRepository: ~10s (8 тестов)
- UseCases: ~8s (6 тестов)

**Итого:** ~48s для всех 65 integration тестов.

## Новые тесты

С момента первоначального создания добавлены:
- **Delete тесты** — batch удаление досок и карточек (DeleteBoard с repeated board_ids, DeleteCard с repeated card_ids)
- **Feature тесты** — дополнительные сценарии для новых возможностей (creator_id в карточках, search, sort и т.д.)

Общее количество integration тестов: **65** (было 36).

Полная актуальная информация: [../testing/README.md](../testing/README.md)

## Следующие шаги

Рекомендуемые дополнительные тесты:
- [ ] Тесты производительности (benchmark) для pagination
- [ ] Тесты concurrent updates с race detector
- [ ] Тесты для edge cases (очень длинные titles, special characters)
- [ ] Тесты для транзакций (rollback scenarios)
