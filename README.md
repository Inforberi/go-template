# go-template

Стартовый шаблон HTTP API на Go с PostgreSQL, SQL-миграциями и Docker Compose. Локальная разработка работает через Air, production собирается из текущего checkout на сервере.

## Возможности

- Go `1.26.7`, Chi, Zap и конфигурация из environment variables;
- PostgreSQL 18.6, `pgxpool` и `golang-migrate`;
- Swagger UI из Go-аннотаций и строгая JSON-валидация HTTP DTO;
- liveness и readiness с проверкой PostgreSQL;
- автоматические локальные PostgreSQL backup с SHA256 и защищённым restore;
- hot reload в development;
- минимальный production-образ: distroless, non-root, read-only filesystem;
- отдельный Dozzle для просмотра Docker-логов через SSH tunnel;
- базовый GitHub Actions CI и опциональный управляемый SSH-deploy.

## Быстрый старт

### 1. Установите зависимости

Для запуска проекта локально нужны Docker Desktop и Task. Go требуется только для локальных Go-команд.

| Инструмент | macOS | Linux-сервер |
| --- | --- | --- |
| Docker + Compose | [Docker Desktop](https://docs.docker.com/desktop/setup/install/mac-install/) | [Docker Engine + Compose plugin](https://docs.docker.com/engine/install/) |
| Task | `brew install go-task` | см. [официальную установку](https://taskfile.dev/docs/installation) |
| Go `1.26.7` | для format, vet и build | не требуется для deployment |
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

`GET /ping` возвращает `pong`. API доступен на `http://localhost:${PORT}`, PostgreSQL development — только на `127.0.0.1:${POSTGRES_PORT}`.

Системные проверки:

- `GET /health/live` подтверждает, что HTTP-процесс работает;
- `GET /health/ready` проверяет PostgreSQL через pool и возвращает `503`, если БД недоступна.

При `SWAGGER_ENABLED=true` Swagger UI доступен на `http://localhost:${PORT}/swagger/index.html` и защищён значениями `SWAGGER_USERNAME` и `SWAGGER_PASSWORD`. В production Swagger по умолчанию выключен.

### Использование как шаблона

Перед началом нового проекта замените `github.com/Inforberi/go-template` на новый module path во всех Go-файлах, затем обновите `module`:

```bash
go mod edit -module github.com/acme/orders
go mod tidy
```

Также измените `APP_NAME`, `SERVICE_NAME`, Swagger title и заголовок README. Git remote настраивается отдельно.

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
| Все локальные проверки | `task check` |
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
- `DATABASE_URL` — URL с теми же реквизитами PostgreSQL;
- `DATABASE_MAX_CONNS` — максимальный размер PostgreSQL pool;
- при `SWAGGER_ENABLED=true` — `SWAGGER_USERNAME` и `SWAGGER_PASSWORD`;
- `BACKUP_HOST_DIR` — каталог backup на сервере без shell-подстановок; пользователь deployment должен иметь право его создать и изменять.

`task prod:deploy` соберёт production-образ API из текущего checkout, затем запустит сервисы:

```bash
task prod:deploy
task prod:ps
task prod:logs
```

Команда создаёт backup-каталог с правами `0700`, собирает API, ожидает healthy PostgreSQL, ровно один раз применяет миграции и запускает API вместе с backup-sidecar. Docker Compose автоматически создаёт `${POSTGRES_HOST_DIR}` при первом запуске. Данные БД лежат в этом каталоге; `task prod:down` и `docker compose down -v` их не удаляют. Для удаления БД нужно явно удалить этот каталог.

Production API слушает только `127.0.0.1:${PORT}` и должен публиковаться через host reverse proxy с TLS. PostgreSQL не имеет host-порта; для доступа используйте `docker compose --env-file .env.prod -f deploy/compose.prod.yml exec postgres psql`.

## CI и управляемый deploy

Workflow `CI` на каждый push и pull request проверяет форматирование, `go vet`, race-тесты, Compose/shell-конфигурации и production Docker build.

Локальный и серверный deployment всегда можно выполнить напрямую через `task prod:deploy`. Workflow `Deploy` добавляет два опциональных режима:

- ручной запуск через **Actions → Deploy → Run workflow**;
- automatic deployment успешного commit из `main`, только если Repository Variable `AUTO_DEPLOY` равна `true`.

Для GitHub Environment `production` настройте:

| Тип | Имя | Значение |
| --- | --- | --- |
| Secret | `DEPLOY_HOST` | адрес сервера |
| Secret | `DEPLOY_USER` | SSH-пользователь |
| Secret | `DEPLOY_SSH_KEY` | приватный SSH-ключ |
| Secret | `DEPLOY_KNOWN_HOSTS` | проверенная строка `known_hosts` сервера |
| Variable | `DEPLOY_PATH` | абсолютный путь к checkout на сервере |
| Variable | `DEPLOY_PORT` | SSH-порт, по умолчанию `22` |
| Variable | `AUTO_DEPLOY` | `true` для auto-deploy, иначе выключен |

Workflow передаёт серверу SHA, успешно прошедший CI, проверяет чистоту checkout и совпадение `origin/main` с этим SHA, выполняет только fast-forward и запускает `task prod:deploy`. Если в `main` уже появился другой commit, deployment прекращается. Для дополнительного контроля включите required reviewers у Environment `production`.

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

Restore под общим lock проверяет SHA256 и TOC архива, останавливает writers, создаёт `pre_restore`, пересоздаёт БД, применяет миграции текущего checkout и выполняет SQL healthcheck. Ранее работавший API запускается только после успешного завершения всей последовательности. При ошибке writers остаются остановленными.

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
internal/infra/           конфигурация, логгер и PostgreSQL pool
internal/transport/       HTTP server, router, middleware и helpers
docs/                     сгенерированная Swagger 2.0 документация
migrations/               SQL-миграции
deploy/                   Dockerfile и Compose-конфигурации
Taskfile.yml              все поддерживаемые команды проекта
.env.example              локальная конфигурация
.env.prod.example         production-конфигурация
```

Новые HTTP-маршруты добавляйте в `internal/transport/router`. Конфигурация приложения читается из environment variables, которые Compose загружает из `.env` или `.env.prod`.

Шаблон намеренно не включает прикладную аутентификацию/RBAC, CORS, rate limiting, metrics/tracing, очереди и Kubernetes: эти решения добавляются под требования конкретного проекта.
