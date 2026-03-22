# Генерация Proto файлов — Board Service

Руководство по генерации gRPC-кода из protobuf определений Board Service.

---

## Требования

Вам нужен один из следующих инструментов:
1. **protoc** установлен локально (рекомендуется)
2. **Docker** установлен (альтернатива)

---

## Метод 1: Использование локального protoc (рекомендуется)

### Установка

#### macOS (с Homebrew):
```bash
brew install protobuf
```

#### Ubuntu/Debian:
```bash
# Установить protoc
apt-get update
apt-get install -y protobuf-compiler

# Установить Go-плагины для protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

**Проверка версии:**
```bash
protoc --version
# Должно вывести: libprotoc 3.x.x или выше
```

**Проверка Go-плагинов:**
```bash
which protoc-gen-go
which protoc-gen-go-grpc
# Должны быть в $GOPATH/bin (убедитесь что в PATH)
```

#### Windows (с Chocolatey):
```bash
choco install protoc
```

Затем установите Go-плагины:
```powershell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

#### Ручная установка (Linux/macOS)

Если у вас нет package manager:

```bash
# Скачать protoc (проверьте последнюю версию на GitHub)
PROTOC_VERSION=25.1
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip

# Распаковать
unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d $HOME/.local

# Добавить в PATH (в ~/.bashrc или ~/.zshrc)
export PATH="$PATH:$HOME/.local/bin"

# Установить Go-плагины
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Генерация

После установки `protoc`:

```bash
cd services/board
make proto
```

**Что происходит:**
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/v1/board.proto
```

**Что генерируется:**
- `api/proto/v1/board.pb.go` — Protocol Buffer определения сообщений
- `api/proto/v1/board_grpc.pb.go` — gRPC сервисный код

---

## Метод 2: Использование Docker

Если у вас установлен Docker, но нет `protoc`:

```bash
cd services/board
make proto-docker
```

**Что происходит:**
```bash
docker run --rm \
  -v $(PWD):/workspace \
  -w /workspace \
  protocolbuffers/protoc:latest \
  --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/proto/v1/board.proto
```

**Преимущества:**
- Не нужно устанавливать protoc локально
- Одинаковая версия protoc у всех разработчиков (через Docker image)

**Недостатки:**
- Требует Docker
- Чуть медленнее чем локальный protoc (запуск контейнера)

---

## Метод 3: Использование docker-compose (из корня проекта)

Из корня проекта (`/home/roman/PetProject/yammi`):

```bash
docker-compose run --rm protoc-builder \
  protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    services/board/api/proto/v1/board.proto
```

**Когда использовать:**
- У вас нет protoc локально
- Вы работаете из корня проекта (не из `services/board/`)

---

## Проверка сгенерированных файлов

После генерации проверьте что файлы созданы:

```bash
ls -la services/board/api/proto/v1/*.pb.go
```

**Ожидаемый вывод:**
```
-rw-r--r-- board.pb.go
-rw-r--r-- board_grpc.pb.go
```

**Размер файлов:**
- `board.pb.go` — обычно 10-50 KB (зависит от количества message типов)
- `board_grpc.pb.go` — обычно 5-20 KB (зависит от количества RPC методов)

**Проверка синтаксиса:**
```bash
cd services/board
go build ./api/proto/v1
```

Если нет ошибок → файлы сгенерированы корректно.

---

## Troubleshooting

### "protoc: command not found"

**Проблема:** `protoc` не установлен или не в PATH.

**Решение:**
- Установите protoc одним из методов выше
- ИЛИ используйте `make proto-docker` (если есть Docker)

### "protoc-gen-go: command not found"

**Проблема:** Go-плагин для protoc не установлен или не в PATH.

**Решение:**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Проверьте что $GOPATH/bin в PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Проверьте
which protoc-gen-go
which protoc-gen-go-grpc
```

### Docker image pull errors

**Проблема:** Нет интернета или Docker не может скачать образ.

**Решение:**
- Проверьте интернет-соединение
- Попробуйте скачать образ вручную: `docker pull protocolbuffers/protoc:latest`

### Сгенерированные файлы пустые или отсутствуют

**Проблема:** Ошибка в proto файле или неправильные флаги.

**Решение:**
- Проверьте что `api/proto/v1/board.proto` существует и валиден
- Проверьте синтаксис proto файла: `protoc --lint_errors api/proto/v1/board.proto`
- Проверьте вывод команды (есть ли ошибки)

### "undefined: grpc.SupportPackageIsVersion7"

**Проблема:** Несовместимые версии gRPC библиотек.

**Решение:**
```bash
cd services/board
go get -u google.golang.org/grpc
go mod tidy
```

Затем пересгенерируйте proto:
```bash
make clean
make proto
```

---

## Интеграция с CI/CD

### GitHub Actions

```yaml
- name: Generate gRPC code
  run: |
    cd services/board
    make proto
```

Или с Docker:

```yaml
- name: Generate gRPC code (Docker)
  run: |
    cd services/board
    make proto-docker
```

### GitLab CI

```yaml
before_script:
  - apk add --no-cache protobuf protobuf-dev
  - go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  - go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

script:
  - cd services/board
  - make proto
```

---

## Структура Proto файла

**Файл:** `api/proto/v1/board.proto`

**Package:** `board.v1`

**Services:**
- `BoardService` — операции с досками, колонками, карточками и участниками

### Сообщения (messages)

**Domain модели:**
- `Board` — доска (id, title, description, owner_id, version, timestamps)
- `Column` — колонка (id, board_id, title, position)
- `Card` — карточка (id, column_id, board_id, title, description, position lexorank, assignee_id, version, timestamps, creator_id)
- `BoardMember` — участник доски (user_id, role)

**Request/Response:**
- `CreateBoardRequest`, `CreateBoardResponse`
- `GetBoardRequest`, `GetBoardResponse`
- `UpdateBoardRequest`, `UpdateBoardResponse`
- `DeleteBoardRequest` (batch: `repeated string board_ids`), returns `google.protobuf.Empty`
- `ListBoardsRequest`, `ListBoardsResponse`
- `AddColumnRequest`, `AddColumnResponse`
- `CreateCardRequest`, `CreateCardResponse`
- `MoveCardRequest`, `MoveCardResponse`
- `DeleteCardRequest` (batch: `repeated string card_ids`), returns `google.protobuf.Empty`
- `AddMemberRequest`, `AddMemberResponse`
- ... и другие

### RPC методы (примеры)

```protobuf
service BoardService {
  // Board operations
  rpc CreateBoard(CreateBoardRequest) returns (CreateBoardResponse);
  rpc GetBoard(GetBoardRequest) returns (GetBoardResponse);
  rpc UpdateBoard(UpdateBoardRequest) returns (UpdateBoardResponse);
  rpc DeleteBoard(DeleteBoardRequest) returns (DeleteBoardResponse);
  rpc ListBoards(ListBoardsRequest) returns (ListBoardsResponse);

  // Column operations
  rpc AddColumn(AddColumnRequest) returns (AddColumnResponse);
  rpc UpdateColumn(UpdateColumnRequest) returns (UpdateColumnResponse);
  rpc DeleteColumn(DeleteColumnRequest) returns (DeleteColumnResponse);
  rpc ListColumns(ListColumnsRequest) returns (ListColumnsResponse);

  // Card operations
  rpc CreateCard(CreateCardRequest) returns (CreateCardResponse);
  rpc GetCard(GetCardRequest) returns (GetCardResponse);
  rpc UpdateCard(UpdateCardRequest) returns (UpdateCardResponse);
  rpc DeleteCard(DeleteCardRequest) returns (DeleteCardResponse);
  rpc MoveCard(MoveCardRequest) returns (MoveCardResponse);
  rpc ListCards(ListCardsRequest) returns (ListCardsResponse);

  // Member operations
  rpc AddMember(AddMemberRequest) returns (AddMemberResponse);
  rpc RemoveMember(RemoveMemberRequest) returns (RemoveMemberResponse);
  rpc ListMembers(ListMembersRequest) returns (ListMembersResponse);
}
```

См. `services/board/api/proto/v1/board.proto` для полного определения.

---

## Перегенерация после изменений

Когда вы изменяете `api/proto/v1/board.proto`:

### 1. Очистить старые сгенерированные файлы

```bash
make clean
```

**Что удаляется:**
- `api/proto/v1/*.pb.go`
- `api/proto/v1/*_grpc.pb.go`

### 2. Перегенерировать

```bash
make proto
```

### 3. Проверить что сборка проходит

```bash
make build
```

**Если есть ошибки:**
- Проверьте что новые поля proto файла используются в коде
- Обновите gRPC handlers в `internal/delivery/grpc/`
- Обновите domain модели в `internal/domain/` (если схема изменилась)

---

## Best Practices

### 1. Не коммитить сгенерированные файлы

Добавьте в `.gitignore`:
```gitignore
# Generated proto files
**/*.pb.go
**/*_grpc.pb.go
```

**Почему:**
- Генерируются автоматически при сборке (в CI/CD и локально)
- Засоряют diff в pull requests
- Могут конфликтовать при merge

**Альтернатива (если нужно коммитить):**
- Генерируйте на этапе CI/CD
- Используйте одинаковую версию protoc у всех разработчиков (через Docker)

### 2. Версионирование proto файлов

Структура: `api/proto/v1/board.proto` (v1, v2, v3...)

При breaking changes:
1. Создать `api/proto/v2/board.proto`
2. Оставить v1 для обратной совместимости
3. Постепенно мигрировать клиентов на v2

### 3. Backwards compatibility

**Можно добавлять:**
- Новые поля в message (с тегами > последнего использованного)
- Новые RPC методы
- Новые enum значения (с reserved первыми значениями)

**Нельзя (breaking changes):**
- Удалять/переименовывать поля
- Менять тип поля
- Менять номер тега поля
- Удалять RPC методы

**Пример добавления поля:**
```protobuf
message Card {
  string id = 1;
  string title = 2;
  string description = 3;
  // ... existing fields ...
  string label = 10;  // новое поле (tag > 9)
}
```

### 4. Документирование proto

Используйте комментарии в proto файлах:

```protobuf
// Card представляет карточку задачи на доске.
//
// Position использует lexorank алгоритм (строка вместо INT)
// для эффективного reordering без массовых UPDATE.
message Card {
  // Уникальный идентификатор карточки (UUID v4)
  string id = 1;

  // ID колонки к которой принадлежит карточка
  string column_id = 2;

  // Название карточки (обязательное, max 255 символов)
  string title = 3;

  // Lexorank позиция (строка, например "a", "am", "b")
  // Используется для сортировки карточек в колонке
  string position = 4;

  // Опциональный ID пользователя которому назначена карточка
  optional string assignee_id = 5;

  // ID пользователя создавшего карточку
  string creator_id = 11;
}
```

Комментарии генерируются в Go-коде:
```go
// Card представляет карточку задачи на доске.
//
// Position использует lexorank алгоритм...
type Card struct {
    // Уникальный идентификатор карточки (UUID v4)
    Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
    // ...
}
```

---

## Сравнение методов генерации

| Метод | Скорость | Требования | Рекомендация |
|-------|----------|------------|--------------|
| Локальный protoc | Быстро (~100ms) | protoc + Go plugins | Для разработки |
| Docker | Медленно (~2s) | Docker | Для CI/CD |
| docker-compose | Медленно (~2s) | Docker Compose | Для скриптов |

**Рекомендуемый workflow:**
- **Разработка:** Локальный protoc (`make proto`)
- **CI/CD:** Docker (`make proto-docker`) для одинаковых версий
- **Onboarding:** Docker (новые разработчики без настройки protoc)

---

## Batch Delete операции

Board Service поддерживает batch удаление через `repeated` поля в proto:

### DeleteBoardRequest

```protobuf
message DeleteBoardRequest {
  repeated string board_ids = 1;  // массив ID досок для удаления
  string user_id = 2;
}
```

Удаляет несколько досок за один запрос. Требует права owner для каждой доски. Возвращает `google.protobuf.Empty`.

### DeleteCardRequest

```protobuf
message DeleteCardRequest {
  repeated string card_ids = 1;  // массив ID карточек для удаления
  string board_id = 2;
  string user_id = 3;
}
```

Удаляет несколько карточек за один запрос. Требует членство в доске. Возвращает `google.protobuf.Empty`.

---

## User Service — SearchByEmail RPC

Board Service использует User Service для поиска пользователей при добавлении участников на доску.

**Proto:** `services/user/api/proto/v1/user.proto`

```protobuf
service UserService {
  rpc GetProfile(GetProfileRequest) returns (GetProfileResponse);
  rpc UpdateProfile(UpdateProfileRequest) returns (UpdateProfileResponse);
  rpc SearchByEmail(SearchByEmailRequest) returns (SearchByEmailResponse);
}

message SearchByEmailRequest {
  string query = 1;   // частичный email для поиска (ILIKE)
  int32 limit = 2;    // максимальное количество результатов
}

message SearchByEmailResponse {
  repeated UserInfo users = 1;
}

message UserInfo {
  string id = 1;
  string email = 2;
  string name = 3;
  string avatar_url = 4;
}
```

**Использование:** Фронтенд вызывает SearchByEmail через API Gateway для autocomplete при добавлении участников на доску.

---

## Ссылки

- [Protocol Buffers Documentation](https://developers.google.com/protocol-buffers)
- [gRPC-Go Tutorial](https://grpc.io/docs/languages/go/)
- [Protocol Buffers Go Mapping](https://developers.google.com/protocol-buffers/docs/reference/go-generated)
- [protoc Installation Guide](https://grpc.io/docs/protoc-installation/)
- [gRPC Go Quick Start](https://grpc.io/docs/languages/go/quickstart/)

---

## Частые команды (шпаргалка)

```bash
# Генерация через локальный protoc
make proto

# Генерация через Docker
make proto-docker

# Очистка сгенерированных файлов
make clean

# Проверка что protoc установлен
protoc --version

# Проверка Go-плагинов
which protoc-gen-go
which protoc-gen-go-grpc

# Установка Go-плагинов
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Проверка синтаксиса proto файла
protoc --lint_errors api/proto/v1/board.proto

# Проверка сгенерированных файлов
go build ./api/proto/v1

# Полная пересборка
make clean && make proto && make build
```

---

## Дополнительные инструменты

### buf — современный инструмент для proto

[Buf](https://buf.build/) — замена protoc с лучшим UX:

**Установка:**
```bash
# macOS
brew install bufbuild/buf/buf

# Linux
curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf
```

**Использование:**
```bash
cd services/board

# Генерация (аналог protoc)
buf generate

# Линтинг proto файлов
buf lint

# Проверка breaking changes
buf breaking --against '.git#branch=main'
```

**Конфиг (buf.yaml):**
```yaml
version: v1
lint:
  use:
    - DEFAULT
breaking:
  use:
    - FILE
```

### grpcurl — curl для gRPC

Инструмент для тестирования gRPC API из командной строки:

**Установка:**
```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**Использование:**
```bash
# Список сервисов
grpcurl -plaintext localhost:50053 list

# Список методов сервиса
grpcurl -plaintext localhost:50053 list board.v1.BoardService

# Вызов метода
grpcurl -plaintext \
  -d '{"title": "My Board", "description": "Test"}' \
  localhost:50053 \
  board.v1.BoardService/CreateBoard
```

### evans — интерактивный gRPC клиент

Интерактивная REPL для gRPC:

**Установка:**
```bash
go install github.com/ktr0731/evans@latest
```

**Использование:**
```bash
evans -r repl -p 50053

# В REPL:
> service BoardService
> call CreateBoard
title (TYPE_STRING) => My Board
description (TYPE_STRING) => Test board
# Результат в JSON
```

---

## Итого

**Для локальной разработки:**
```bash
# Установить protoc один раз
brew install protobuf  # (или apt-get)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Использовать при каждом изменении proto
make clean
make proto
make build
```

**Для CI/CD:**
```bash
# В Dockerfile или CI-скрипте
make proto-docker
make build
```

**При проблемах:**
1. Проверьте версию protoc: `protoc --version`
2. Проверьте Go-плагины: `which protoc-gen-go protoc-gen-go-grpc`
3. Используйте Docker как fallback: `make proto-docker`
