#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/.env}"
DB_CONTAINER="${DB_CONTAINER:-dorm-db}"

die() {
    echo "Ошибка: $*" >&2
    exit 1
}

require_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        die "не найдена команда '$1'"
    fi
}

read_required() {
    local prompt="$1"
    local variable_name="$2"
    local value

    while true; do
        read -r -p "$prompt: " value

        if [[ -n "${value//[[:space:]]/}" ]]; then
            printf -v "$variable_name" '%s' "$value"
            return
        fi

        echo "Значение обязательно."
    done
}

read_password() {
    local first
    local second

    while true; do
        read -r -s -p "Пароль первого пользователя: " first
        echo

        if [[ ${#first} -lt 8 ]]; then
            echo "Пароль должен содержать не менее 8 символов."
            continue
        fi

        read -r -s -p "Повторите пароль: " second
        echo

        if [[ "$first" != "$second" ]]; then
            echo "Пароли не совпадают."
            continue
        fi

        PASSWORD="$first"
        return
    done
}

reject_newlines() {
    local name="$1"
    local value="$2"

    if [[ "$value" == *$'\n'* || "$value" == *$'\r'* ]]; then
        die "поле '$name' не должно содержать переносы строк"
    fi
}

sql_escape() {
    local value="$1"

    # SQL ниже включает NO_BACKSLASH_ESCAPES, поэтому достаточно
    # удвоить одинарные кавычки.
    value=${value//\'/\'\'}

    printf '%s' "$value"
}

mysql_exec() {
    docker exec \
        -e MYSQL_PWD="$DB_PASSWORD" \
        -i "$DB_CONTAINER" \
        mysql \
        --default-character-set=utf8mb4 \
        --batch \
        --skip-column-names \
        -u"$DB_USER" \
        "$DB_NAME" \
        "$@"
}

require_command docker
require_command sha256sum

[[ -f "$ENV_FILE" ]] || die "не найден файл $ENV_FILE"

# .env является доверенным конфигурационным файлом проекта.
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

: "${DB_USER:?В .env отсутствует DB_USER}"
: "${DB_PASSWORD:?В .env отсутствует DB_PASSWORD}"
: "${DB_NAME:?В .env отсутствует DB_NAME}"

if ! docker inspect "$DB_CONTAINER" >/dev/null 2>&1; then
    die "контейнер '$DB_CONTAINER' не найден"
fi

DB_RUNNING="$(
    docker inspect \
        --format='{{.State.Running}}' \
        "$DB_CONTAINER"
)"

[[ "$DB_RUNNING" == "true" ]] ||
    die "контейнер '$DB_CONTAINER' не запущен"

echo "Проверка подключения к базе данных..."

mysql_exec -e "SELECT 1;" >/dev/null

USER_COUNT="$(
    mysql_exec -e \
        "SELECT COUNT(*) FROM \`user\` WHERE deleted_at IS NULL;"
)"

DORMITORY_COUNT="$(
    mysql_exec -e \
        "SELECT COUNT(*) FROM dormitory;"
)"

if [[ "$USER_COUNT" != "0" ]]; then
    die "в системе уже существуют пользователи: $USER_COUNT"
fi

if [[ "$DORMITORY_COUNT" != "0" ]]; then
    die "в системе уже существуют общежития: $DORMITORY_COUNT"
fi

echo
echo "Данные первого пользователя"

read_required "Имя" FIRST_NAME
read_required "Фамилия" LAST_NAME
read_required "Логин" LOGIN
read_password

if [[ ! "$LOGIN" =~ ^[a-zA-Z0-9._-]+$ ]]; then
    die "логин может содержать только латинские буквы, цифры, '.', '_' и '-'"
fi

echo
echo "Данные первого общежития"

read_required "Название общежития" DORMITORY_NAME
read_required "Город" DORMITORY_CITY
read_required "Тип улицы, например улица или переулок" STREET_TYPE
read_required "Название улицы" STREET_NAME
read_required "Номер дома" HOUSE_NUMBER

reject_newlines "Имя" "$FIRST_NAME"
reject_newlines "Фамилия" "$LAST_NAME"
reject_newlines "Логин" "$LOGIN"
reject_newlines "Название общежития" "$DORMITORY_NAME"
reject_newlines "Город" "$DORMITORY_CITY"
reject_newlines "Тип улицы" "$STREET_TYPE"
reject_newlines "Название улицы" "$STREET_NAME"
reject_newlines "Номер дома" "$HOUSE_NUMBER"

EXISTING_LOGIN_COUNT="$(
    mysql_exec -e "
        SELECT COUNT(*)
        FROM \`user\`
        WHERE login = '$(sql_escape "$LOGIN")';
    "
)"

if [[ "$EXISTING_LOGIN_COUNT" != "0" ]]; then
    die "пользователь с логином '$LOGIN' уже существует"
fi

USER_ID="$(cat /proc/sys/kernel/random/uuid)"
PASSWORD_HASH="$(
    printf '%s' "$PASSWORD" |
        sha256sum |
        awk '{print $1}'
)"

FIRST_NAME_SQL="$(sql_escape "$FIRST_NAME")"
LAST_NAME_SQL="$(sql_escape "$LAST_NAME")"
LOGIN_SQL="$(sql_escape "$LOGIN")"
DORMITORY_NAME_SQL="$(sql_escape "$DORMITORY_NAME")"
DORMITORY_CITY_SQL="$(sql_escape "$DORMITORY_CITY")"
STREET_TYPE_SQL="$(sql_escape "$STREET_TYPE")"
STREET_NAME_SQL="$(sql_escape "$STREET_NAME")"
HOUSE_NUMBER_SQL="$(sql_escape "$HOUSE_NUMBER")"

echo
echo "Создание пользователя и общежития..."

mysql_exec <<SQL
SET SESSION sql_mode = CONCAT_WS(
    ',',
    @@SESSION.sql_mode,
    'NO_BACKSLASH_ESCAPES'
);

SET NAMES utf8mb4;

START TRANSACTION;

INSERT INTO \`user\` (
    id,
    login,
    password_hash,
    first_name,
    middle_name,
    last_name,
    team_id,
    room_number,
    floor_number,
    dormitory_id,
    created_at,
    deleted_at
)
VALUES (
    UUID_TO_BIN('$USER_ID'),
    '$LOGIN_SQL',
    '$PASSWORD_HASH',
    '$FIRST_NAME_SQL',
    NULL,
    '$LAST_NAME_SQL',
    NULL,
    NULL,
    NULL,
    NULL,
    NOW(),
    NULL
);

INSERT INTO dormitory (
    leader_id,
    name,
    city,
    street_type,
    street_name,
    house_number
)
VALUES (
    UUID_TO_BIN('$USER_ID'),
    '$DORMITORY_NAME_SQL',
    '$DORMITORY_CITY_SQL',
    '$STREET_TYPE_SQL',
    '$STREET_NAME_SQL',
    '$HOUSE_NUMBER_SQL'
);

SET @new_dormitory_id = LAST_INSERT_ID();

UPDATE \`user\`
SET dormitory_id = @new_dormitory_id
WHERE id = UUID_TO_BIN('$USER_ID');

COMMIT;

SELECT
    BIN_TO_UUID(u.id) AS user_id,
    u.login,
    CONCAT_WS(' ', u.first_name, u.last_name) AS user_name,
    d.id AS dormitory_id,
    d.name AS dormitory_name,
    BIN_TO_UUID(d.leader_id) AS dormitory_leader_id
FROM \`user\` u
JOIN dormitory d ON d.id = u.dormitory_id
WHERE u.id = UUID_TO_BIN('$USER_ID');
SQL

unset PASSWORD
unset PASSWORD_HASH

echo
echo "Инициализация завершена."
echo "Логин: $LOGIN"
echo "Пользователь: $FIRST_NAME $LAST_NAME"
echo "Общежитие: $DORMITORY_NAME"
echo "ID пользователя: $USER_ID"
