# 🛡️ GophKeeper

**GophKeeper** — CLI-менеджер для безопасного хранения логинов, заметок, карт и бинарных данных, построенный на Go с использованием gRPC, шифрования и поддержки мнемоник.

---

## 🚀 Возможности

* Хранение логинов, заметок, карт, файлов.
* Шифрование данных на клиенте (AES-128 GCM + seed от мнемоники).
* CLI-оболочка с интерактивным `shell`-режимом.
* Поддержка множественных контекстов (профилей).
* Регистрация и логин с HMAC-хешированием паролей.
* Сервер на gRPC с JWT-аутентификацией.
* Логирование на основе `zap`.
* Конфигурация через `viper`.

---

## 📦 Установка

```bash
git clone https://github.com/yourname/go-goph-keeper.git
cd go-goph-keeper
```

Сборка клиента:

```bash
GOOS=windows GOARCH=amd64   go build -o bin/gk.exe     ./cmd/client
GOOS=linux   GOARCH=amd64   go build -o bin/gk-linux   ./cmd/client
GOOS=darwin  GOARCH=amd64   go build -o bin/gk-darwin  ./cmd/client
```

Сборка сервера:

```bash
GOOS=windows GOARCH=amd64   go build -o bin/gk-server.exe     ./cmd/server
GOOS=linux   GOARCH=amd64   go build -o bin/gk-server-linux   ./cmd/server
GOOS=darwin  GOARCH=amd64   go build -o bin/gk-server-darwin  ./cmd/server
```

---

## ⚙️ Конфигурация

Пример `config.client.yaml`:

```yaml
databaseKV:
  dirPath: ./data

master: "your-master-key"
```

Пример `config.server.yaml`:

```yaml
server:
  port: "50051"

database:
  dsn: "host=localhost port=5432 user=postgres dbname=gophkeeper password=pass sslmode=disable"

master: "your-master-key"
```

---

## 🧠 Как пользоваться

Всё взаимодействие с приложением осуществляется через интерактивную оболочку:

```bash
gk shell
```

Внутри интерактивной сессии доступны следующие команды:

```bash
login              войти в аккаунт или создать новый
register           зарегистрировать новый аккаунт
contexts           список всех контекстов
use <name>         сменить контекст
list               показать все записи
get <id>           показать запись по ID
delete <id>        удалить запись по ID
create             создать новую запись
me                 вывести текущую информацию о контексте
exit / quit / q    выйти из программы
help / ?           список команд
```

---

## 🔐 Шифрование

* Генерация seed'а из мнемоники + пароль.
* Шифрование с `AES-GCM (128 бит)` на клиенте.
* Расшифровка также на клиенте, сервер не видит содержимого.

---

## 🛠️ Используемые технологии

* Go 1.21+
* Cobra (CLI)
* gRPC
* zap (logger)
* Viper (config)
* samber/do (DI)
* GORM (ORM for server)
* RoseDB (KV storage for client)
* Protocol Buffers

---

## 📁 Структура проекта

````
cmd/
  client/        — CLI-интерфейс
    internal/kv     — хранилище данных (RoseDB)
    internal/crypto — шифрование и генерация seed
  server/        — gRPC-сервер
internal/
  api/           — gRPC-протоколы
  config/        — загрузка конфигов
  service/       — бизнес-логика
  storage/       — слои хранения
  logger/        — zap-логгер
  server/        — gRPC-интерсепторы и сервер
  kv/            — внутреннее хранилище клиента на RoseDB (internal/client/kv)
  crypto/        — шифрование и работа с seed/mnemonic```

---

## 🧰 Taskfile.yml

Пример задач для сборки и разработки:

```yaml
version: '3'

tasks:

  build-win:
    desc: Сборка под Windows
    cmds:
      - GOOS=windows GOARCH=amd64 go build -o gk.exe ./cmd/client
      - GOOS=windows GOARCH=amd64 go build -o gk-server.exe ./cmd/server

  build-linux:
    desc: Сборка под Linux
    cmds:
      - GOOS=linux GOARCH=amd64 go build -o bin/gk-linux ./cmd/client
      - GOOS=linux GOARCH=amd64 go build -o bin/gk-server-linux ./cmd/server

  build-mac:
    desc: Сборка под macOS
    cmds:
      - GOOS=darwin GOARCH=amd64 go build -o bin/gk-darwin ./cmd/client
      - GOOS=darwin GOARCH=amd64 go build -o bin/gk-server-darwin ./cmd/server


  generate:
    desc: генерация билд инфы
    cmds:
      - go generate ./cmd/client/

  server:
    desc: Запуск сервачка
    cmds:
      - go run ./cmd/server/. -c=./config/config.server.yaml
  client:
    desc: Запуск клиентика
    cmds:
      - go run ./cmd/client/.


  mock:
    desc: генерация mocks
    cmds:
      - mockgen -source=internal/storage/interface.go -destination=internal/mocks/mock_storage.go -package=mocks
      - mockgen -source=internal/service/interface.go -destination=internal/mocks/mock_service.go -package=mocks
      - mockgen -source=internal/api/server_grpc.pb.go -destination=cmd/client/internal/mocks/mock_gophkeeper.go -package=mocks GophKeeperClient
      - mockgen -source=cmd/client/internal/kv/interface.go -destination=cmd/client/internal/mocks/mock_storage.go -package=mocks


  test:
    desc: тестирование с coverage
    cmds:
      - go test -v -coverpkg=./... -coverprofile=coverage.tmp.out -covermode=count ./...
      - go run ./cmd/cover .
      - go tool cover -func coverage.out

  docker:
    desc: билд базы в докере
    cmds:
      - docker run --name my-postgres -e POSTGRES_USER=username -e POSTGRES_PASSWORD=password -e POSTGRES_DB=praktikum -p 5432:5432 -d postgres

  protoc:
    cmds:
      - go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
      - go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest


  protoc-create:
    desc: генерация прото
    cmds:
      - protoc -I=proto/ --go_out=./internal/api/ --go_opt=paths=source_relative --go-grpc_out=./internal/api/ --go-grpc_opt=paths=source_relative proto/*.proto
````

---

## 📜 Генерация gRPC

```bash
protoc \
  --go_out=. \
  --go-grpc_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  internal/api/api.proto
```

---

## ✅ ToDo

* [ ] Обновление записей
* [ ] Поддержка OTP
* [ ] UI через bubbletea (TUI)
* [ ] Интеграционные тесты
* [ ] Экспорт/импорт данных

---

## 🧑‍💻 Автор

[@wicked](https://github.com/wickedv43)

---