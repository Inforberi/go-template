set -Eeuo pipefail

readonly PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
readonly ENV_FILE="${PROJECT_ROOT}/.env.prod"
readonly COMPOSE_FILE="${PROJECT_ROOT}/deploy/compose.prod.yml"

readonly LOCK_FILENAME=".backup-restore.lock"

readonly -a STOP_SERVICES=(api migrate)
readonly -a RESTARTABLE_SERVICES=(api)

compose() {
    docker compose \
        --project-directory "$PROJECT_ROOT" \
        --env-file "$ENV_FILE" \
        -f "$COMPOSE_FILE" \
        "$@"
}

fail() {
    printf 'ERROR: %s\n' "$*" >&2
    exit 1
}

read_env_value() {
    local key="$1"
    local value

    value="$(sed -n "s/^${key}=//p" "$ENV_FILE" | tail -n 1)"
    value="${value%$'\r'}"

    if [[ "$value" == \"*\" && "$value" == *\" ]]; then
        value="${value:1:${#value}-2}"
    elif [[ "$value" == \'*\' && "$value" == *\' ]]; then
        value="${value:1:${#value}-2}"
    fi

    printf '%s\n' "$value"
}

resolve_backup_host_dir() {
    local configured

    configured="${BACKUP_HOST_DIR:-$(read_env_value BACKUP_HOST_DIR)}"
    configured="${configured:-./backups}"

    if [[ "$configured" == /* ]]; then
        printf '%s\n' "$configured"
    else
        printf '%s/%s\n' "$PROJECT_ROOT" "${configured#./}"
    fi
}

prepare_backup_dir() {
    local backup_host_dir

    backup_host_dir="$(resolve_backup_host_dir)"

    mkdir -p "$backup_host_dir"
    chmod 0700 "$backup_host_dir"

    touch "${backup_host_dir}/${LOCK_FILENAME}"
    chmod 0600 "${backup_host_dir}/${LOCK_FILENAME}"

    [[ -d "$backup_host_dir" && -w "$backup_host_dir" ]] \
        || fail "backup directory is not writable: $backup_host_dir"
}

run_backup_unlocked() {
    compose run \
        --rm \
        --no-deps \
        -e BACKUP_LOCK_HELD=1 \
        backup \
        "$@"
}

run_migrations() {
    compose run \
        --rm \
        --no-deps \
        migrate
}

is_service_running() {
    local service="$1"
    local running_service

    while IFS= read -r running_service; do
        [[ "$running_service" == "$service" ]] && return 0
    done < <(compose ps --status running --services)

    return 1
}

restore_database() {
    local archive="${RESTORE_FILE:-}"
    local confirmation="${RESTORE_CONFIRM:-}"

    local postgres_db
    local backup_host_dir
    local service

    local -a restart_services=()

    [[ -f "$ENV_FILE" ]] \
        || fail "missing $ENV_FILE"

    [[ "$archive" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*\.dump$ ]] \
        || fail "invalid backup filename: $archive"

    postgres_db="${POSTGRES_DB:-$(read_env_value POSTGRES_DB)}"

    [[ -n "$postgres_db" ]] \
        || fail "POSTGRES_DB is not configured"

    [[ "$confirmation" == "prod:${postgres_db}" ]] \
        || fail "confirmation must be prod:${postgres_db}"

    prepare_backup_dir

    backup_host_dir="$(resolve_backup_host_dir)"

    command -v flock >/dev/null 2>&1 \
        || fail "flock is required; install util-linux"

    # Один restore/backup одновременно.
    exec 9>"${backup_host_dir}/${LOCK_FILENAME}"

    flock -n 9 \
        || fail "backup/restore operation is already running"

    # Сначала проверяем backup, пока ничего не остановили.
    run_backup_unlocked verify-unlocked "$archive"

    # Запоминаем, какие сервисы работали.
    for service in "${RESTARTABLE_SERVICES[@]}"; do
        if is_service_running "$service"; then
            restart_services+=("$service")
        fi
    done

    # Останавливаем приложение и возможную миграцию.
    compose stop "${STOP_SERVICES[@]}"

    # Закрываем подключения к восстанавливаемой БД.
    run_backup_unlocked terminate-sessions-unlocked

    # Проверяем, можно ли безопасно DROP DATABASE.
    run_backup_unlocked check-drop-blockers-unlocked

    # Страховочный backup текущего состояния перед уничтожением БД.
    run_backup_unlocked create-unlocked pre_restore false

    # Пересоздаём БД.
    run_backup_unlocked recreate-database-unlocked

    # Восстанавливаем выбранный dump.
    run_backup_unlocked restore-unlocked "$archive"

    # Доводим старую схему backup до версии текущего Go-кода.
    run_migrations

    # Проверяем восстановленную БД.
    run_backup_unlocked healthcheck-unlocked

    # Применяем retention.
    run_backup_unlocked retention-unlocked

    # Возвращаем API только если он работал до restore.
    if ((${#restart_services[@]} > 0)); then
        compose start "${restart_services[@]}"
    fi

    printf 'Restore completed: %s\n' "$archive"
}

main() {
    local command="${1:-}"

    case "$command" in
        prepare)
            [[ -f "$ENV_FILE" ]] \
                || fail "missing $ENV_FILE"

            prepare_backup_dir
            ;;

        restore)
            restore_database
            ;;

        *)
            fail "usage: $0 {prepare|restore}"
            ;;
    esac
}

main "$@"