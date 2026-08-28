# go-template

Стартовый шаблон HTTP API на Go с PostgreSQL, SQL-миграциями и Docker Compose. Локальная разработка работает через Air, production собирается из текущего checkout на сервере.

## Возможности

- Go `1.26.6`, Chi, Zap и конфигурация из environment variables;
- PostgreSQL 18.6 и `golang-migrate`;
- Swagger UI из Go-аннотаций и строгая JSON-валидация HTTP DTO;
- автоматические локальные PostgreSQL backup с SHA256 и защищённым restore;
- hot reload в development;
- минимальный production-образ: distroless, non-root, read-only filesystem;
- отдельный Dozzle для просмотра Docker-логов через SSH tunnel.

## Быстрый старт

### 1. Установите зависимости

Для запуска проекта локально нужны Docker Desktop и Task. Go требуется только для локальных Go-команд.

| Инструмент | macOS | Linux-сервер |
| --- | --- | --- |
| Docker + Compose | [Docker Desktop](https://docs.docker.com/desktop/setup/install/mac-install/) | [Docker Engine + Compose plugin](https://docs.docker.com/engine/install/) |
| Task | `brew install go-task` | см. [официальную установку](https://taskfile.dev/docs/installation) |
| Go `1.26.6` | для format, vet и build | не требуется для deployment |
| `flock` | не требуется для development | пакет `util-linux`, обычно уже установлен |

На Linux установите Task для текущего пользователя:

```bash
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.profile
```

Для запуска Docker без `sudo` добавьте доверенного пользователя в группу `docker`, затем переподключитесь по SSH:

```bash
sudo usermod -aG docker "$USER"
```

### 2. Создайте локальную конфигурацию

```bash
cp .env.example .env
```

Значения по умолчанию подходят для локальной разработки. `.env` и `.env.prod` не попадают в Git.

### 3. Запустите API

```bash
task dev:up
curl http://127.0.0.1:8080/ping
```

`GET /ping` возвращает `pongg`. API доступен на `http://localhost:${PORT}`, PostgreSQL — только на `127.0.0.1:${POSTGRES_PORT}`.

Swagger UI доступен на `http://localhost:${PORT}/swagger/index.html` и защищён значениями `SWAGGER_USERNAME` и `SWAGGER_PASSWORD`. В production Basic Auth должен использоваться только через HTTPS.

## Ежедневная разработка

`task dev:up` запускает PostgreSQL, применяет существующие миграции и запускает API через Air. После изменения Go-файлов Air пересобирает и перезапускает приложение. По умолчанию данные development БД лежат в `data/postgres-dev/` и не попадают в Git.

| Задача | Команда |
| --- | --- |
| Запустить окружение | `task dev:up` |
| Статус сервисов | `task dev:ps` |
| Логи API | `task dev:logs` |
| Остановить окружение | `task dev:down` |
| Форматирование | `task go:fmt` |
| Статический анализ | `task go:vet` |
| Локальная сборка в `bin/api` | `task go:build` |
| Форматирование Swagger-аннотаций | `task swagger:fmt` |
| Генерация Swagger-документации | `task swagger:generate` |
| Тесты | `go test ./...` |

После изменения Swagger-аннотаций выполните `task swagger:generate` и закоммитьте обновлённый каталог `docs/`.

### HTTP DTO и валидация

`httpx.DecodeAndValidateJSON` принимает `application/json` с параметрами, ограничивает тело 1 MiB, запрещает неизвестные поля и проверяет `validate`-теги.

```go
type CreateRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func create(w http.ResponseWriter, r *http.Request) {
	var request CreateRequest
	if !httpx.DecodeAndValidateJSON(w, r, &request) {
		return
	}

	// Обработка валидного request.
}
```

Ошибки тела возвращаются с HTTP-кодами `400`, `413` или `415`; ошибки DTO-валидации — `422` с `details` без исходных значений полей.

### Миграции

```bash
task migrate:create NAME=create_users
task migrate:up
task migrate:version
task migrate:down
```

Миграции лежат в `migrations/`. Пустой каталог — допустимое состояние для нового проекта: запуск не завершится ошибкой.

## Production

На сервере должны быть Docker Engine, Compose plugin, Task и исходный код проекта.

```bash
cp .env.prod.example .env.prod
```

В `.env.prod` обязательно задайте:

- `POSTGRES_PASSWORD` — уникальный пароль;
- `POSTGRES_HOST_DIR` — каталог с данными PostgreSQL относительно корня проекта;
- `DATABASE_URL` — URL с теми же реквизитами PostgreSQL.
- `SWAGGER_USERNAME` и `SWAGGER_PASSWORD` — credentials Swagger UI;
- `BACKUP_HOST_DIR` — каталог backup на сервере без shell-подстановок; пользователь deployment должен иметь право его создать и изменять.

`task prod:deploy` соберёт production-образ API из текущего checkout, затем запустит сервисы:

```bash
task prod:deploy
task prod:ps
task prod:logs
```

Команда создаёт backup-каталог с правами `0700`, ожидает healthy PostgreSQL, применяет миграции и запускает API вместе с backup-sidecar. Docker Compose автоматически создаёт `${POSTGRES_HOST_DIR}` при первом запуске. Данные БД лежат в этом каталоге; `task prod:down` и `docker compose down -v` их не удаляют. Для удаления БД нужно явно удалить этот каталог.

> API публикует `${PORT}` на хосте. Ограничьте внешний доступ firewall или reverse proxy в соответствии с инфраструктурой.

## PostgreSQL backup и restore

Backup-sidecar создаёт custom archive `pg_dump` сразу при первом запуске, затем с интервалом `BACKUP_INTERVAL`. Повторный deploy не создаёт новый архив, если последний корректный backup ещё свежий. Хранятся последние `BACKUP_RETENTION_COUNT` пар:

```text
backup_2026-08-28_12-00-00Z.dump
backup_2026-08-28_12-00-00Z.dump.sha256
```

`.dump` содержит данные БД, а `.dump.sha256` — контрольную сумму: для restore нужны оба файла. При совпадении времени добавляется счётчик, например `_01`.

Доступные команды:

```bash
task backup:create
task backup:list
task backup:logs
task backup:restore \
  FILE=backup_2026-08-28_12-00-00Z.dump \
  CONFIRM=prod:app
```

Пример `task backup:list`:

```text
CREATED (UTC)        TYPE         SIZE           SHA256     FILE
2026-08-28 12:00:00Z backup       12 345 bytes   ok         backup_2026-08-28_12-00-00Z.dump
```

Restore под общим lock проверяет SHA256 и TOC архива, останавливает writers, создаёт `pre_restore`, пересоздаёт БД и запускает ранее работавший API только после успешной проверки. `migrate` автоматически не запускается. При ошибке writers остаются остановленными.

`POSTGRES_USER` считается владельцем БД. ACL и ownership из архива намеренно не восстанавливаются. При разделении ролей на owner/runtime user добавьте отдельный bootstrap ролей и grants.

Локальный backup не является disaster recovery: потеря диска или VM уничтожит `${POSTGRES_HOST_DIR}` и `${BACKUP_HOST_DIR}`. Для критичных данных добавьте off-site storage либо base backup и WAL-архивацию с PITR, а также регулярно выполняйте restore drill на отдельной БД.

## Docker-логи через Dozzle

Dozzle — независимый Compose-проект для просмотра stdout/stderr всех контейнеров Docker host. Он подключается к Docker API через read-only socket proxy; actions, shell и MCP выключены.

```bash
task logs:up
task logs:ps
task logs:logs
task logs:down
```

UI слушает только `127.0.0.1:9999`. На production-сервере подключайтесь через SSH tunnel:

```bash
ssh -N \
  -L 127.0.0.1:9999:127.0.0.1:9999 \
  -o ExitOnForwardFailure=yes \
  -o ServerAliveInterval=30 \
  user@server
```

Откройте `http://127.0.0.1:9999` в локальном браузере. Не открывайте TCP/9999 в firewall и не публикуйте UI через reverse proxy без TLS и аутентификации. `task prod:down` не затрагивает Dozzle; `task logs:down` сохраняет данные UI в named volume.

## Структура

```text
cmd/api/                  точка входа приложения
internal/app/             инициализация зависимостей
internal/infra/           конфигурация и логгер
internal/transport/       HTTP server, router, middleware и helpers
docs/                     сгенерированная Swagger 2.0 документация
migrations/               SQL-миграции
deploy/                   Dockerfile и Compose-конфигурации
Taskfile.yml              все поддерживаемые команды проекта
.env.example              локальная конфигурация
.env.prod.example         production-конфигурация
```

Новые HTTP-маршруты добавляйте в `internal/transport/router`. Конфигурация приложения читается из environment variables, которые Compose загружает из `.env` или `.env.prod`.
