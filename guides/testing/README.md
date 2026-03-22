# Testing Guide — Board Service

## Обзор

**223 теста**, параллельный запуск (8 потоков), общее время ~2s.

| Категория | Тестов | Coverage | Время | Файлы |
|-----------|--------|----------|-------|-------|
| Unit: domain | 111 | **95.9%** | ~0.2s | `internal/domain/*_test.go` |
| Unit: usecase | 47 | **46.9%** | ~0.01s | `internal/usecase/*_test.go` |
| Feature: integration | 65 | **63.8%** (internal/*) | ~1.8s | `tests/integration/*_test.go` |

---

## Запуск

```bash
cd services/board

# Unit тесты (без зависимостей)
go test ./internal/... -parallel 8

# Feature тесты (нужен PostgreSQL)
TEST_DATABASE_URL="postgres://yammi:yammi@localhost:5432/board_test?sslmode=disable" \
  go test ./tests/integration/ -parallel 8 -timeout 180s

# Всё вместе с coverage
go test ./internal/... -cover
TEST_DATABASE_URL="..." go test ./tests/integration/ -cover -coverpkg=./internal/...
```

---

## Unit тесты: Domain (111 тестов, 95.9%)

Файлы в `services/board/internal/domain/`.

### card_test.go
| Тест | Что проверяет |
|------|---------------|
| TestNewCard/valid_card_with_assignee | Создание карточки с исполнителем, все поля |
| TestNewCard/valid_card_without_assignee | Карточка без assignee (nil) |
| TestNewCard/valid_card_with_different_position | Разные lexorank позиции |
| TestNewCard/empty_column_id | Ошибка ErrColumnNotFound |
| TestNewCard/empty_title | Ошибка ErrEmptyCardTitle |
| TestNewCard/empty_position | Ошибка ErrInvalidLexorank |
| TestNewCard/invalid_lexorank_characters | Спецсимволы в позиции |
| TestNewCard/uppercase_letters_in_lexorank | Верхний регистр запрещён |
| TestCard_Update/* | Update title/description/assignee, пустой title → ошибка |
| TestCard_Move/* | Move между колонками, пустой columnID, невалидная позиция |
| TestCard_Reorder/* | Reorder позиции, edge cases |
| TestCard_LexorankValidation/* | Валидные/невалидные lexorank строки (8 кейсов) |

### board_test.go
| Тест | Что проверяет |
|------|---------------|
| TestNewBoard/* | Создание, пустой title, пустой ownerID |
| TestBoard_Update/* | Обновление, пустой title |
| TestBoard_IsOwner/* | Проверка владельца |
| TestBoard_IncrementVersion/* | Version +1, updated_at обновляется |

### column_test.go
| Тест | Что проверяет |
|------|---------------|
| TestNewColumn/* | Создание, пустой title, пустой boardID, отрицательная позиция |
| TestColumn_Update/* | Обновление title |
| TestColumn_UpdatePosition/* | Обновление позиции, отрицательная → ошибка |

### member_test.go
| Тест | Что проверяет |
|------|---------------|
| TestNewMember/* | Создание owner/member, невалидная роль |
| TestMember_CanModifyBoard/* | Owner — true, Member — false |
| TestMember_CanModifyCards/* | Owner и Member — true |

### lexorank_test.go
| Тест | Что проверяет |
|------|---------------|
| TestLexorankBetween/* | Генерация позиции между двумя (9 кейсов), инварианты prev < result < next |
| TestLexorankOrdering | Лексикографическая сортировка |
| TestLexorankSequence | Последовательные вставки |
| TestValidateLexorank/* | Валидные/невалидные строки |
| TestLexorankEdgeCases/* | Длинные строки, trailing zeros |

---

## Unit тесты: Usecase (47 тестов, 46.9%)

Файлы в `services/board/internal/usecase/`. Используют моки (testify/mock).

### delete_card_test.go (9 тестов)
| Тест | Что проверяет |
|------|---------------|
| Owner удаляет чужую карточку | Owner bypasses creator check → success |
| Creator (member) удаляет свою | CreatorID == userID → success |
| Не участник доски | IsMember false → ErrAccessDenied |
| Member удаляет чужую карточку | CreatorID != userID → **ErrAccessDenied** |
| Batch delete (owner) | 3 карточки одним вызовом → success |
| Карточка не найдена | GetByID → ErrCardNotFound |
| Пустой список IDs | Пустой slice → no-op |
| Ошибка IsMember | DB error propagated |
| Ошибка BatchDelete | DB error propagated |

### delete_board_test.go (7 тестов)
| Тест | Что проверяет |
|------|---------------|
| Single delete (owner) | Owner удаляет → success |
| Batch delete (owner) | 3 доски одним вызовом → success |
| Not-owner denied | Member → ErrAccessDenied |
| Not-member denied | Чужой → ErrAccessDenied |
| Partial ownership batch | Owner доски A, member доски B → **ErrAccessDenied на всю операцию** |
| IsMember error | DB error propagated |
| BatchDelete error | DB error propagated |

### create_board_test.go (4 теста)
Создание доски, пустой title, пустой ownerID, ошибка repo.

### create_card_test.go (7 тестов)
С позицией, без позиции (пустая колонка / конец), non-member denied, пустой title, невалидный lexorank, ошибка GetLastInColumn.

### add_column_test.go (6 тестов)
Owner/member add, non-member denied, пустой title, отрицательная позиция, ошибка save.

### add_member_test.go (7 тестов)
Owner добавляет member/owner, not-owner denied, невалидная роль, board not found, duplicate, ошибка add.

### get_board_test.go (5 тестов)
Owner/member get, not found, non-member denied, IsMember error.

### list_boards_test.go (5 тестов)
Дефолтный/кастомный лимит, лимит >100, отрицательный, ошибка repo.

### move_card_test.go (4 теста)
Перемещение в начало/конец/между/пустую колонку.

---

## Feature тесты (65 тестов, реальный PostgreSQL)

Файлы в `services/board/tests/integration/`. Каждый тест работает с реальной PostgreSQL, все данные изолированы через UUID.

### feature_test.go — Бизнес-сценарии (19 тестов)

**Доски:**
| Тест | Сценарий |
|------|----------|
| CreateBoard_OwnerAutoMember | Создание доски → owner автоматически member с ролью RoleOwner |
| ListBoards_OnlyMemberBoards | User A и B создают доски → каждый видит только свои |
| ListBoards_MemberSeesSharedBoards | A создаёт доску, добавляет B → B видит её в списке |
| ListBoards_OwnerOnlyFilter | ownerOnly=true → показывает только свои, не shared |
| ListBoards_SearchByTitle | "Alpha", "Beta", "Alphabet" → поиск "Alph" → 2 результата |
| UpdateBoard_OnlyOwner | Owner обновляет OK, non-member → ErrAccessDenied |
| GetBoard_NonMemberDenied | Owner получает OK, non-member → ErrAccessDenied |

**Колонки:**
| Тест | Сценарий |
|------|----------|
| AddColumn_MemberCanAdd | Member добавляет колонку → success, проверка в БД |
| AddColumn_NonMemberDenied | Non-member → ErrAccessDenied |
| DeleteColumn_OnlyOwnerCanDelete | Non-member denied, member/owner могут |

**Карточки:**
| Тест | Сценарий |
|------|----------|
| CreateCard_SetsCreatorID | Member создаёт карточку → **creator_id == memberID** в БД |
| CreateCard_NonMemberDenied | Non-member → ErrAccessDenied |
| MoveCard_MemberCanMove | Member перемещает → success, column_id обновлён |
| MoveCard_NonMemberDenied | Non-member → ErrAccessDenied, карточка на месте |
| UpdateCard_MemberCanUpdate | Member обновляет title/desc/assignee → success |

**Участники:**
| Тест | Сценарий |
|------|----------|
| AddMember_OnlyOwnerCanAdd | Owner OK, member → ErrNotOwner, non-member → ErrNotOwner |
| RemoveMember_OnlyOwnerCanRemove | Member пытается → ошибка, owner → success |
| RemoveMember_CannotRemoveOwner | Owner self-remove → **ErrCannotRemoveOwner** |
| AfterRemoval_NoAccess | Добавить → удалить → GetBoard → **ErrAccessDenied** |

### delete_test.go — Удаление (9 тестов)

**Доски:**
| Тест | Сценарий |
|------|----------|
| OwnerCanDelete | Owner удаляет → board ErrBoardNotFound |
| MemberCannotDelete | Member → ErrAccessDenied, board на месте |
| BatchDelete | 3 доски одним вызовом → все удалены |
| CascadeDeletesCards | Доска с колонками и карточками → всё удалено каскадно |

**Карточки:**
| Тест | Сценарий |
|------|----------|
| CreatorCanDelete | Member удаляет свою → success |
| OwnerCanDeleteAnyCard | Owner удаляет чужую → success |
| MemberCannotDeleteOthersCard | Member B удаляет карточку A → **ErrAccessDenied**, карточка на месте |
| BatchDelete | 3 карточки одним вызовом → все удалены |
| NonMemberCannotDelete | Non-member → ErrAccessDenied, карточка на месте |

### *_repository_test.go — Репозитории (31 тест)

- **BoardRepository** (7): Create, GetByID_NotFound, Update, OptimisticLocking, Delete, CursorPagination, ListByUserID_EmptyResult
- **CardRepository** (9): Create, CreateWithoutAssignee, GetByID_NotFound, ListByColumnID, LexorankPositioning, Update, Move, Delete, Partitioning
- **ColumnRepository** (6): Create, GetByID_NotFound, ListByBoardID, Update, Delete, CascadeDelete
- **MembershipRepository** (9): AddMember, AddMember_Duplicate, AddMember_InvalidRole, RemoveMember, RemoveMember_NotFound, RemoveMember_CannotRemoveOwner, IsMember, ListMembers, ListMembers_Pagination

### usecase_test.go — Usecase интеграции (6 тестов)

CreateBoard, GetBoard (owner/member/non-member), AddColumn (owner/non-member), CreateCard (member/non-member), MoveCard, AddMember (owner/non-owner).

---

## Изоляция

- Каждый тест использует **уникальные UUID** для user ID → данные не пересекаются
- Тесты запускаются **параллельно на 8 потоках** без race conditions
- Поддержка `TEST_DATABASE_URL` env (CI) или testcontainers (Docker-in-Docker)
- Миграции идемпотентны (IF NOT EXISTS)

## Что проверяется

- CRUD все сущности (board, column, card, member)
- **Права доступа**: owner vs member vs non-member на каждую операцию
- **creator_id**: только создатель или owner может удалять карточки
- **Batch delete**: доски и карточки (одна транзакция)
- **Каскадное удаление**: board → columns + cards + members
- **Optimistic locking**: concurrent updates → ErrInvalidVersion
- **Pagination**: cursor-based и offset
- **Lexorank**: ordering, positioning, generation
- **Partitioning**: распределение по 4 партициям
- **ILIKE search**: поиск по названию доски (pg_trgm)
