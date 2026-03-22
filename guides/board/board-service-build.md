# Board Service — Руководство по сборке

Полное руководство по сборке, тестированию и управлению Board Service.

---

## Быстрый старт

### Требования

- **Go 1.23** или выше
- **PostgreSQL 16** (для локального запуска)
- **Redis 7** (для кеширования)
- **NATS 2** (для событий)
- **protoc** (для генерации proto) ИЛИ **Docker** (альтернатива protoc)

### Сборка сервиса

```bash
cd services/board

# Генерация gRPC-кода (выберите один метод)
make proto              # Если protoc установлен локально
make proto-docker       # Если доступен Docker

# Скачать зависимости
make deps

# Собрать бинарный файл
make build

# Или собрать в debug режиме (без CGO/GOOS/GOARCH флагов)
make build-debug
```

---

## Makefile таргеты

### `make proto`

Генерирует gRPC-код из `api/proto/v1/board.proto`.

**Требования:** `protoc` установлен локально

**Что генерируется:**
- `api/proto/v1/board.pb.go` — Protocol Buffer сообщения
- `api/proto/v1/board_grpc.pb.go` — gRPC сервисные определения

**Обработка ошибок:** Если `protoc` не найден, команда завершается с понятными инструкциями.

### `make proto-docker`

Генерирует gRPC-код используя официальный Docker-образ Protocol Buffers.

**Требования:** Docker установлен и доступен

**Команда:**
```bash
docker run --rm \
  -v $(PWD):/workspace \
  -w /workspace \
  protocolbuffers/protoc:latest \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/proto/v1/board.proto
```

**Преимущество:** Не нужно устанавливать protoc локально

### `make deps`

Скачивает и проверяет зависимости Go модуля.

**Команда:**
```bash
go mod download
go mod verify
```

### `make test`

Запускает unit-тесты из `internal/domain/`.

```bash
go test ./internal/domain/ -v
```

### `make test-integration`

Запускает интеграционные тесты из `tests/integration/`.

```bash
go test ./tests/integration/ -v
```

### `make build`

Собирает production бинарник для Linux (CGO отключен).

**Вывод:** `./bin/board`

**Флаги:**
- `CGO_ENABLED=0` — Статическая сборка, без C-зависимостей
- `GOOS=linux` — Целевая ОС Linux
- `GOARCH=amd64` — Целевая архитектура x86_64 (дефолт)

### `make build-debug`

Собирает debug бинарник для текущей платформы (с CGO).

**Вывод:** `./bin/board`

**Применение:** Локальная разработка и отладка

### `make clean`

Удаляет все сгенерированные файлы:
- Директория `./bin/`
- `api/proto/v1/*.pb.go`
- `api/proto/v1/*_grpc.pb.go`

### `make help`

Отображает все доступные таргеты и их описания.

---

## Структура проекта

```
services/board/
├── Makefile                      # Автоматизация сборки
├── BUILD.md                      # Этот файл (старая версия, англ.)
├── PROTO_GENERATION.md           # Детальное руководство по proto (старая версия, англ.)
├── Dockerfile                    # Определение Docker-образа
├── go.mod                        # Определение Go модуля
├── go.sum                        # Контрольные суммы зависимостей
├── cmd/
│   └── server/
│       └── main.go              # Точка входа приложения
├── api/
│   └── proto/v1/
│       ├── board.proto          # Определение gRPC сервиса
│       ├── board.pb.go          # Сгенерировано: proto сообщения
│       └── board_grpc.pb.go     # Сгенерировано: gRPC сервисы
├── internal/
│   ├── domain/                  # Бизнес-логика (entities, errors)
│   ├── usecase/                 # Логика приложения (оркестрация)
│   ├── delivery/grpc/           # gRPC обработчики
│   ├── repository/postgres/     # Слой доступа к БД
│   └── infrastructure/          # Внешние сервисы (БД, кеш, очередь)
├── migrations/                  # SQL миграции
├── configs/                     # Конфигурационные файлы
├── tests/
│   └── integration/             # Интеграционные тесты
└── bin/
    └── board                    # Скомпилированный бинарник (после сборки)
```

---

## Рабочий процесс разработки

### 1. Изменение Proto-определений

Отредактируйте `api/proto/v1/board.proto` чтобы добавить или изменить gRPC методы/сообщения.

### 2. Перегенерация кода

```bash
make clean
make proto  # или make proto-docker
```

### 3. Запуск тестов

```bash
make test
make test-integration
```

### 4. Сборка бинарника

```bash
make build
```

---

## Docker интеграция

### Сборка Docker-образа

Из корня проекта:

```bash
docker build -t yammi-board:latest services/board
```

Dockerfile использует multi-stage сборку:
1. **Builder stage**: Go 1.24-alpine — компилирует бинарник
2. **Runtime stage**: Alpine 3.19 — запускает бинарник

### Запуск в Docker Compose

Из корня проекта:

```bash
docker-compose up board
```

Сервис будет:
- Подключаться к PostgreSQL на `postgres:5432`
- Подключаться к Redis на `redis:6379`
- Подключаться к NATS на `nats:4222`
- Слушать порт `50053` для gRPC

---

## Переменные окружения

Board Service читает из `.env` или переменных окружения:

| Переменная | Дефолт | Описание |
|----------|---------|----------|
| `DATABASE_URL` | Обязательно | Строка подключения PostgreSQL |
| `REDIS_URL` | Опционально | Строка подключения Redis (host:port) |
| `NATS_URL` | Опционально | URL подключения NATS |
| `BOARD_GRPC_PORT` | `50053` | Порт gRPC сервера |
| `LOG_LEVEL` | `info` | Уровень логирования (debug, info, warn, error) |

Пример `.env`:
```env
DATABASE_URL=postgres://user:pass@localhost:5432/board_db
REDIS_URL=localhost:6380
NATS_URL=nats://localhost:4222
BOARD_GRPC_PORT=50053
LOG_LEVEL=info
```

---

## Тестирование

### Unit-тесты (Domain логика)

```bash
make test
```

Тестирует бизнес-логику в `internal/domain/`.

**Покрываемые области:**
- Валидация Board (title не пустой, owner_id не пустой)
- Валидация Column (title не пустой, position >= 0)
- Валидация Card (title не пустой, lexorank валиден)
- Lexorank алгоритм (LexorankBetween, edge cases)
- Member roles (CanModifyBoard, CanModifyCards)

### Интеграционные тесты

```bash
make test-integration
```

Тестирует весь сервис включая БД, кеш и gRPC взаимодействия. Требует:
- PostgreSQL запущен
- Redis запущен
- NATS запущен

Или использует testcontainers для автоматического поднятия зависимостей.

### Запуск конкретного теста

```bash
cd services/board
go test ./internal/domain -run TestBoardAggregate -v
```

**Примеры:**
```bash
# Тест создания доски
go test ./internal/domain -run TestNewBoard -v

# Тест lexorank
go test ./internal/domain -run TestLexorankBetween -v

# Тест валидации карточки
go test ./internal/domain -run TestCard_Move -v
```

---

## Troubleshooting

### "protoc: command not found"

**Решение:** Используйте `make proto-docker` или установите protoc:
- **macOS**: `brew install protobuf`
- **Linux**: `apt-get install protobuf-compiler`
- **Windows**: `choco install protoc`

См. `/home/roman/PetProject/yammi/guides/board-service-proto.md` для детальных инструкций по установке.

### "docker: command not found"

**Решение:** Установите Docker с https://www.docker.com/

### Сборка падает с "module not found"

**Решение:** Запустите `make deps` для скачивания зависимостей:
```bash
make deps
make build
```

### Тесты падают с "connection refused"

**Решение:** Запустите инфраструктурные сервисы:
```bash
docker-compose up postgres redis nats
```

Затем запустите тесты:
```bash
make test-integration
```

### "optimistic lock failed" в логах

**Что это:** Две параллельные операции пытались обновить одну доску.

**Нормально:** Один из запросов получит 409 Conflict и клиент повторит с новой версией.

**Проблема:** Если происходит часто (> 5% запросов) → возможно слишком много concurrent updates одной доски.

**Решение:** Уменьшить частоту автосохранения на клиенте или добавить debouncing.

---

## Производительность

### Советы по сборке

1. **Пересобрать только proto**: `make clean && make proto && make build`
2. **Пропустить генерацию proto**: `make build` (если уже сгенерировано)
3. **Параллельная сборка**: `go build -p 4` (для 4 параллельных единиц компиляции)

### Профилирование

**CPU профилирование:**
```bash
go test ./internal/domain -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

**Memory профилирование:**
```bash
go test ./internal/domain -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### Бенчмарки

**Lexorank алгоритм:**
```bash
cd services/board
go test ./internal/domain -bench=BenchmarkLexorank -benchmem
```

**Repository операции:**
```bash
cd services/board
go test ./internal/repository/postgres -bench=. -benchmem
```

---

## CI/CD интеграция

### GitHub Actions пример

```yaml
name: Board Service Build

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Install protoc
        run: apt-get update && apt-get install -y protobuf-compiler
      - name: Build Board Service
        run: |
          cd services/board
          make deps
          make proto
          make test
          make build
```

### GitLab CI пример

```yaml
board-service:
  image: golang:1.23-alpine
  before_script:
    - apk add --no-cache protobuf protobuf-dev
  script:
    - cd services/board
    - make deps
    - make proto
    - make test
    - make build
  artifacts:
    paths:
      - services/board/bin/board
```

---

## Миграции базы данных

### Создание новой миграции

```bash
# Формат: <version>_<description>.up.sql
touch services/board/migrations/000002_add_card_labels.up.sql
```

**Пример содержимого:**
```sql
-- 000002_add_card_labels.up.sql
CREATE TABLE card_labels (
    card_id UUID REFERENCES cards(id) ON DELETE CASCADE,
    label VARCHAR(50) NOT NULL,
    color VARCHAR(20) NOT NULL,
    PRIMARY KEY (card_id, label)
);

CREATE INDEX idx_card_labels_card_id ON card_labels(card_id);
```

### Накат миграций

Миграции накатываются автоматически при старте сервиса (`internal/infrastructure/migrator.go`).

**Ручной накат (для тестирования):**
```bash
# Используя migrate CLI (если установлен)
migrate -path services/board/migrations \
        -database "postgres://user:pass@localhost:5432/yammi_board?sslmode=disable" \
        up
```

### Откат миграций (down)

```bash
# Создайте .down.sql файл
touch services/board/migrations/000002_add_card_labels.down.sql
```

**Содержимое:**
```sql
-- 000002_add_card_labels.down.sql
DROP TABLE card_labels;
```

**Откат последней миграции:**
```bash
migrate -path services/board/migrations \
        -database "postgres://..." \
        down 1
```

---

## Мониторинг

### Health Check

```bash
# gRPC health check (требует grpcurl)
grpcurl -plaintext localhost:50053 grpc.health.v1.Health/Check
```

**Ответ:**
```json
{
  "status": "SERVING"
}
```

### Метрики (будущее)

**Планируется:**
- Prometheus metrics endpoint на `:9090/metrics`
- Метрики: requests_total, request_duration, cache_hit_rate, db_query_duration

**Пример использования:**
```bash
curl http://localhost:9090/metrics | grep board_
```

---

## Локальная разработка

### Запуск без Docker

```bash
# 1. Поднять зависимости в Docker
docker-compose up postgres redis nats

# 2. Накатить миграции (автоматически при старте, или вручную)
# (сервис сам накатит при запуске)

# 3. Запустить сервис
cd services/board
go run cmd/server/main.go
```

**Переменные окружения (.env в services/board/):**
```env
DATABASE_URL=postgres://yammi_user:yammi_pass@localhost:5432/yammi_board?sslmode=disable
REDIS_URL=localhost:6380
NATS_URL=nats://localhost:4222
BOARD_GRPC_PORT=50053
LOG_LEVEL=debug
```

### Hot reload (air)

Установите [air](https://github.com/cosmtrek/air):
```bash
go install github.com/cosmtrek/air@latest
```

**Конфиг (.air.toml в services/board/):**
```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main cmd/server/main.go"
  bin = "tmp/main"
  include_ext = ["go", "proto"]
  exclude_dir = ["tmp", "bin"]
  delay = 1000
```

**Запуск:**
```bash
cd services/board
air
```

При изменении `.go` или `.proto` файлов → автоматический rebuild и restart.

---

## Ссылки

- [Схема БД](/home/roman/PetProject/yammi/guides/database-schema.md)
- [Lexorank алгоритм](/home/roman/PetProject/yammi/guides/lexorank-explained.md)
- [Архитектура проекта](/home/roman/PetProject/yammi/guides/architecture.md)
- [Генерация Proto](/home/roman/PetProject/yammi/guides/board-service-proto.md)
- [Protocol Buffers Documentation](https://developers.google.com/protocol-buffers)
- [gRPC Go Guide](https://grpc.io/docs/languages/go/)
- [Docker Documentation](https://docs.docker.com/)

---

## Частые команды (шпаргалка)

```bash
# Полная пересборка с нуля
make clean && make proto && make deps && make build

# Только unit-тесты
make test

# Только один тест
go test ./internal/domain -run TestNewBoard -v

# Запуск сервиса локально
go run cmd/server/main.go

# Генерация proto через Docker (без локального protoc)
make proto-docker

# Сборка Docker-образа
docker build -t yammi-board:latest .

# Запуск в Docker Compose
docker-compose up board

# Проверка health
grpcurl -plaintext localhost:50053 grpc.health.v1.Health/Check

# Профилирование CPU
go test ./internal/domain -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Форматирование кода
go fmt ./...

# Линтинг
go vet ./...
```
