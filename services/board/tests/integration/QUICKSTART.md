# Quick Start - Integration Tests

## Предварительные требования

1. **Docker** должен быть запущен:
   ```bash
   docker ps
   ```

2. **Go 1.23+** установлен:
   ```bash
   go version
   ```

## Установка зависимостей

```bash
cd /home/roman/PetProject/yammi/services/board
go mod download
go mod tidy
```

## Запуск тестов

### Вариант 1: Через Makefile (рекомендуется)

```bash
cd /home/roman/PetProject/yammi/services/board/tests/integration

# Все тесты
make test

# С подробным выводом
make test-verbose

# Отдельный тест
make test-single TEST=TestBoardRepository_Create
```

### Вариант 2: Через скрипт

```bash
cd /home/roman/PetProject/yammi/services/board
./scripts/run-integration-tests.sh
```

### Вариант 3: Напрямую через go test

```bash
cd /home/roman/PetProject/yammi/services/board

# Все integration тесты
go test ./tests/integration/... -v -timeout 10m

# Без кэша
go test ./tests/integration/... -v -count=1 -timeout 10m

# Только BoardRepository тесты
go test ./tests/integration/ -run TestBoard -v
```

## Примеры команд

### Запустить все тесты для BoardRepository
```bash
go test ./tests/integration/ -run TestBoardRepository -v
```

### Запустить тест optimistic locking
```bash
go test ./tests/integration/ -run TestBoardRepository_OptimisticLocking -v
```

### Запустить тест partitioning
```bash
go test ./tests/integration/ -run TestCardRepository_Partitioning -v
```

### Запустить все Use Case тесты
```bash
go test ./tests/integration/ -run UseCase -v
```

### С coverage report
```bash
cd tests/integration
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
# Откройте coverage.html в браузере
```

## Ожидаемый результат

```
🧪 Running Board Service Integration Tests
==========================================
📦 Installing dependencies...
🏃 Running integration tests...

=== RUN   TestBoardRepository_Create
--- PASS: TestBoardRepository_Create (2.31s)
=== RUN   TestBoardRepository_OptimisticLocking
--- PASS: TestBoardRepository_OptimisticLocking (2.15s)
=== RUN   TestCardRepository_Partitioning
--- PASS: TestCardRepository_Partitioning (3.47s)
...
PASS
ok      github.com/RomaLytar/yammi/services/board/tests/integration    48.234s

✅ All integration tests passed!
```

## Troubleshooting

### Ошибка: "Cannot connect to Docker daemon"
**Решение:** Запустите Docker
```bash
# Linux/WSL
sudo systemctl start docker

# macOS
open -a Docker

# Windows
# Запустите Docker Desktop
```

### Ошибка: "Failed to start container"
**Решение:** Увеличьте timeout или очистите Docker
```bash
# Очистка неиспользуемых контейнеров
docker system prune -f

# Проверка доступных ресурсов
docker info
```

### Тесты зависают
**Решение:** Проверьте логи контейнера
```bash
# Найдите зависший контейнер
docker ps

# Проверьте логи
docker logs <container_id>
```

### Медленные тесты
**Причина:** Testcontainers создает новый контейнер для каждого теста

**Ускорение (опционально):**
- Используйте SSD
- Увеличьте Docker memory/CPU
- Или запустите только нужные тесты:
  ```bash
  go test ./tests/integration/ -run TestBoardRepository_Create -v
  ```

## Следующие шаги

1. ✅ Запустите все тесты и убедитесь, что они проходят
2. ✅ Изучите покрытие: `make test-coverage`
3. ✅ Прочитайте подробную документацию: `README.md`
4. ✅ Изучите структуру тестов: `TESTS_SUMMARY.md`
5. ✅ Добавьте тесты в CI/CD pipeline

## Полезные ссылки

- [Testcontainers Go Documentation](https://golang.testcontainers.org/)
- [PostgreSQL Docker Hub](https://hub.docker.com/_/postgres)
- [Board Service CLAUDE.md](/home/roman/PetProject/yammi/CLAUDE.md)

## Контакты

Вопросы или проблемы? Создайте issue в репозитории.
