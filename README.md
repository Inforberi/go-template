# go-template

Стартовый шаблон HTTP API на Go с PostgreSQL, SQL-миграциями и Docker Compose. Локальная разработка работает через Air, production — из готового образа registry.

## Возможности

- Go `1.26.6`, Chi, Zap и конфигурация из environment variables;
- PostgreSQL 18 и `golang-migrate`;
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

## Ежедневная разработка

`task dev:up` запускает PostgreSQL, применяет существующие миграции и запускает API через Air. После изменения Go-файлов Air пересобирает и перезапускает приложение.

| Задача | Команда |
| --- | --- |
| Запустить окружение | `task dev:up` |
| Статус сервисов | `task dev:ps` |
| Логи API | `task dev:logs` |
| Остановить окружение | `task dev:down` |
| Форматирование | `task go:fmt` |
| Статический анализ | `task go:vet` |
| Локальная сборка в `bin/api` | `task go:build` |
| Тесты | `go test ./...` |

### Миграции

```bash
task migrate:create NAME=create_users
task migrate:up
task migrate:version
task migrate:down
```

Миграции лежат в `migrations/`. Пустой каталог — допустимое состояние для нового проекта: запуск не завершится ошибкой.

## Production

На сервере должны быть Docker Engine, Compose plugin, Task и доступ к registry с образом API.

```bash
cp .env.prod.example .env.prod
```

В `.env.prod` обязательно задайте:

- `APP_IMAGE` — тег опубликованного образа API;
- `POSTGRES_PASSWORD` — уникальный пароль;
- `DATABASE_URL` — URL с теми же реквизитами PostgreSQL.

Для закрытого registry выполните `docker login`, затем запустите deployment:

```bash
task prod:deploy
task prod:ps
task prod:logs
```

Команда ожидает healthy PostgreSQL, применяет миграции при их наличии и запускает API. `task prod:down` останавливает приложение, но сохраняет PostgreSQL volume.

> API публикует `${PORT}` на хосте. Ограничьте внешний доступ firewall или reverse proxy в соответствии с инфраструктурой.

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
migrations/               SQL-миграции
deploy/                   Dockerfile и Compose-конфигурации
Taskfile.yml              все поддерживаемые команды проекта
.env.example              локальная конфигурация
.env.prod.example         production-конфигурация
```

Новые HTTP-маршруты добавляйте в `internal/transport/router`. Конфигурация приложения читается из environment variables, которые Compose загружает из `.env` или `.env.prod`.
