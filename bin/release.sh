#!/usr/bin/env bash
# Internal remote implementation, streamed over SSH by deploy/backup/rollback.
set -Eeuo pipefail
umask 077
ROOT=${1:?Missing deployment directory}
ACTION=${2:?Missing action}
NEW=${3:-}
[[ "$ROOT" =~ ^/[a-zA-Z0-9_./-]+$ && "$ROOT" != / ]] || { echo 'Invalid deployment directory' >&2; exit 1; }
for command in docker flock gzip mktemp timeout curl; do command -v "$command" >/dev/null; done
mkdir -p "$ROOT/backups" "$ROOT/releases"
exec 9>"$ROOT/.deployment.lock"
flock -n 9 || { echo 'Another deployment, backup or rollback is running' >&2; exit 1; }
TEMP=''
trap '[[ -z "$TEMP" ]] || rm -f -- "$TEMP"' EXIT
fail() { echo "$*" >&2; return 1; }
mysql_query() {
  timeout 30 docker exec -i dorm-db sh -c 'export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"; exec mysql -uroot --batch --skip-column-names "$MYSQL_DATABASE"' <<< "$1"
}
check_db() {
  [[ $(docker inspect -f '{{.State.Running}}' dorm-db) == true ]] || fail 'MySQL container is not running'
  [[ $(docker inspect -f '{{index .Config.Labels "com.docker.compose.project"}}' dorm-db) == dorm ]] || fail 'Unexpected MySQL Compose project; refusing to select another volume'
  [[ $(mysql_query 'SELECT 1;') == 1 ]] || fail 'MySQL is unavailable'
}
create_backup() {
  check_db
  local size available usage name target
  size=$(mysql_query 'SELECT COALESCE(SUM(data_length + index_length),0) FROM information_schema.tables WHERE table_schema=DATABASE();')
  [[ "$size" =~ ^[0-9]+$ ]] || fail 'Cannot estimate database size'
  read -r available usage < <(df -Pk "$ROOT/backups" | awk 'END {gsub(/%/, "", $5); printf "%.0f %s\n", $4*1024, $5}')
  (( available >= size * 2 + 67108864 )) || fail 'Insufficient backup disk space (requires twice the estimated DB size plus 64 MiB)'
  (( usage < 85 )) || echo "WARNING: backup filesystem is ${usage}% full; retention is manual" >&2
  name="dorm-backup-$(date -u +%Y%m%d-%H%M%S).sql.gz"
  target="$ROOT/backups/$name"
  [[ ! -e "$target" ]] || fail "Backup already exists: $target"
  TEMP=$(mktemp "$ROOT/backups/.dump.XXXXXX")
  timeout 1800 docker exec dorm-db sh -c '
    export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"
    exec mysqldump -uroot --single-transaction --quick --no-tablespaces \
      --set-gtid-purged=OFF --default-character-set=utf8mb4 --hex-blob \
      --routines --events --triggers --databases "$MYSQL_DATABASE"
  ' | gzip > "$TEMP"
  [[ -s "$TEMP" ]] && gzip -t "$TEMP"
  gzip -dc "$TEMP" | LC_ALL=C grep -a -E '^CREATE (DATABASE|TABLE)' >/dev/null
  mv -n -- "$TEMP" "$target"
  [[ ! -e "$TEMP" ]] || fail "Refusing to overwrite backup: $target"
  TEMP=''
  BACKUP=$target
  echo "Backup saved: $BACKUP" >&2
}
if [[ "$ACTION" == backup ]]; then
  create_backup
  printf '%s\n' "$BACKUP"
  exit 0
fi
[[ "$ACTION" == deploy || "$ACTION" == rollback ]] || fail 'Unknown release action'
compose() {
  local image=$1; shift
  local directory="$ROOT/releases/${image#dorm-app:}"
  DORM_IMAGE="$image" timeout 180 docker compose -p dorm --project-directory "$directory" \
    --env-file "$directory/.env" -f "$directory/docker-compose.yml" -f "$directory/docker-compose.prod.yml" "$@"
}
schema() {
  local exists value
  exists=$(mysql_query "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name='schema_migrations';") || return 1
  if [[ "$exists" == 0 ]]; then printf '0\n'; return; fi
  value=$(mysql_query 'SELECT version, dirty FROM schema_migrations;') || return 1
  if [[ -z "$value" ]]; then printf '0\n'; return; fi
  if [[ ! "$value" =~ ^([0-9]+)[[:space:]]+0$ ]]; then
    echo "Migration version/dirty: $value" >&2
    fail 'Database migration state is dirty or unreadable; inspect schema manually (no force/down/restore was run)'
    return 1
  fi
  printf '%s\n' "${BASH_REMATCH[1]}"
}
pin_env() {
  local directory=$1 image=$2
  sed -E '/^[[:space:]]*(export[[:space:]]+)?DORM_IMAGE[[:space:]]*=/d' "$directory/.env" > "$directory/.env.tmp"
  printf '\nDORM_IMAGE=%s\n' "$image" >> "$directory/.env.tmp"
  chmod 600 "$directory/.env.tmp"
  mv "$directory/.env.tmp" "$directory/.env"
}
select_release() {
  local image=$1 file
  rm -f "$ROOT/.active.tmp" || return 1
  ln -s "releases/${image#dorm-app:}" "$ROOT/.active.tmp" || return 1
  mv -Tf "$ROOT/.active.tmp" "$ROOT/active" || return 1
  for file in .env docker-compose.yml docker-compose.prod.yml; do
    rm -f "$ROOT/.link.tmp" || return 1
    ln -s "active/$file" "$ROOT/.link.tmp" || return 1
    mv -Tf "$ROOT/.link.tmp" "$ROOT/$file" || return 1
  done
  if [[ ! -e "$ROOT/mysql" && ! -L "$ROOT/mysql" ]]; then ln -s active/mysql "$ROOT/mysql" || return 1; fi
}
state_value() { sed -n "s/^$1=//p" "$ROOT/release-state"; }
CURRENT=''; CURRENT_SCHEMA=''; PREVIOUS=''; PREVIOUS_SCHEMA=''; BACKUP=''; LAST_RESULT=''; STAGE=prepare
if [[ -f "$ROOT/release-state" ]]; then
  CURRENT=$(state_value current); CURRENT_SCHEMA=$(state_value current_schema)
  PREVIOUS=$(state_value previous); PREVIOUS_SCHEMA=$(state_value previous_schema)
  BACKUP=$(state_value backup)
  LAST_RESULT=$(state_value result)
fi
write_state() {
  local status=$1 temporary
  temporary=$(mktemp "$ROOT/.state.XXXXXX")
  printf 'current=%s\ncurrent_schema=%s\nprevious=%s\nprevious_schema=%s\nattempt=%s\nresult=%s\nbackup=%s\n' \
    "$CURRENT" "$CURRENT_SCHEMA" "$PREVIOUS" "$PREVIOUS_SCHEMA" "$NEW" "$status" "$BACKUP" > "$temporary"
  mv "$temporary" "$ROOT/release-state"
}
valid_image() { [[ "$1" =~ ^dorm-app:[a-z0-9][a-z0-9_.-]{0,127}$ ]]; }
for image in "$CURRENT" "$PREVIOUS"; do
  [[ -z "$image" ]] || valid_image "$image" || fail 'Invalid image in release metadata'
done
check_release() {
  local image=$1 end=$((SECONDS+120)) health html assets asset headers expected actual
  local directory="$ROOT/releases/${image#dorm-app:}"
  while (( SECONDS < end )); do
    health=$(timeout 5 docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Running}}{{end}}' dorm-app 2>/dev/null || true)
    case $(timeout 5 docker inspect -f '{{.State.Status}}' dorm-app 2>/dev/null || true) in
      exited|dead) fail "Application container exited: $image"; return 1 ;;
    esac
    if [[ "$health" == unhealthy && ! -f "$directory/legacy" ]]; then
      fail "Application healthcheck failed: $image"
      return 1
    fi
    if [[ -f "$directory/legacy" ]]; then
      [[ $(timeout 5 docker inspect -f '{{.State.Running}}' dorm-app) == true ]] || { sleep 2; continue; }
    else
      [[ "$health" == healthy ]] || { sleep 2; continue; }
      [[ $(curl --silent --output /dev/null --write-out '%{http_code}' --max-time 3 http://127.0.0.1:8080/health/ready || true) == 204 ]] || { sleep 2; continue; }
    fi
    html=$(curl --fail --silent --max-time 3 http://127.0.0.1:8080/app/) || { sleep 2; continue; }
    # A successful fallback page must not pass for a missing asset.
    assets=$(printf '%s' "$html" | grep -Eo '/app/assets/[a-zA-Z0-9_./-]+\.(css|js)' | sort -u) || { sleep 2; continue; }
    [[ "$assets" == *'.js'* && "$assets" == *'.css'* ]] || { sleep 2; continue; }
    local good=1
    while IFS= read -r asset; do
      (( SECONDS < end )) || return 1
      headers=$(curl --fail --silent --head --max-time 3 "http://127.0.0.1:8080$asset") || { good=0; break; }
      if printf '%s' "$headers" | grep -qi '^Content-Type: text/html'; then good=0; break; fi
      # Compare the served content to the selected image, not only its HTTP status.
      expected=$(timeout 10 docker run --rm --entrypoint sha256sum "$image" "/app/web$asset" | awk '{print $1}') || { good=0; break; }
      actual=$(curl --fail --silent --max-time 3 "http://127.0.0.1:8080$asset" | sha256sum | awk '{print $1}') || { good=0; break; }
      [[ "$expected" == "$actual" ]] || { good=0; break; }
    done <<< "$assets"
    if (( good )); then
      [[ ! -f "$directory/legacy" ]] || echo 'Legacy rollback: verified static assets and MySQL; old image has no readiness endpoint' >&2
      return 0
    fi
    sleep 2
  done
  fail "Release validation timed out: $image"
}
on_failure() {
  local code=${1:-$?}
  if (( BASH_SUBSHELL > 0 )); then exit "$code"; fi
  trap - ERR HUP INT TERM
  set +e
  echo "Deployment failed during $STAGE; backup: ${BACKUP:-not created}" >&2
  timeout 15 docker logs --tail 100 dorm-app >&2
  if [[ ${WORKING_SET:-0} == 1 ]]; then
    CURRENT=$WORKING_CURRENT; CURRENT_SCHEMA=$WORKING_SCHEMA
    PREVIOUS=$WORKING_PREVIOUS; PREVIOUS_SCHEMA=$WORKING_PREVIOUS_SCHEMA
  fi
  local status="failed_$STAGE" live
  if [[ "$STAGE" == app && -n "$CURRENT" ]]; then
    live=$(schema)
    if [[ -n "$live" && "$live" == "$BEFORE" && "$live" == "$CURRENT_SCHEMA" ]]; then
      echo "Schema unchanged; restoring $CURRENT" >&2
      if docker image inspect "$CURRENT" >/dev/null && select_release "$CURRENT" && compose "$CURRENT" up -d --no-deps --force-recreate app && check_release "$CURRENT"; then
        status=rolled_back
      else status=rollback_failed; fi
    else
      status=failed_schema_changed
      echo 'Schema changed or is dirty/unreadable. No automatic application/DB rollback. Inspect schema and backup manually.' >&2
    fi
  fi
  write_state "$status"
  exit "$code"
}
trap on_failure ERR
trap 'on_failure 130' INT
trap 'on_failure 143' TERM HUP
if [[ "$ACTION" == rollback ]]; then
  check_db
  LIVE=$(schema)
  BEFORE=$LIVE
  TARGET=$PREVIOUS; TARGET_SCHEMA=$PREVIOUS_SCHEMA
  SELECTED=$(readlink "$ROOT/active" || true)
  if [[ -n "$CURRENT" && ( "$LAST_RESULT" == failed_* || "$LAST_RESULT" == rollback_failed || "$SELECTED" != "releases/${CURRENT#dorm-app:}" ) ]]; then
    TARGET=$CURRENT; TARGET_SCHEMA=$CURRENT_SCHEMA
    echo "Restoring the last confirmed release after an unsuccessful deployment: $TARGET" >&2
  fi
  [[ -n "$TARGET" ]] || fail 'No previous working release recorded'
  docker image inspect "$TARGET" >/dev/null
  if [[ "$LIVE" != "$TARGET_SCHEMA" ]]; then
    [[ "$NEW" == --confirm-schema-compatible ]] || fail 'Schema differs. Verify compatibility first, then explicitly pass --confirm-schema-compatible. This does not prove compatibility.'
    echo 'WARNING: manually confirmed application compatibility; no database rollback will run' >&2
  fi
  NEW=$TARGET
  WORKING_CURRENT=$CURRENT; WORKING_SCHEMA=$CURRENT_SCHEMA
  WORKING_PREVIOUS=$PREVIOUS; WORKING_PREVIOUS_SCHEMA=$PREVIOUS_SCHEMA; WORKING_SET=1
  STAGE=app
  select_release "$NEW"
  compose "$NEW" up -d --no-deps --force-recreate app
  check_release "$NEW"
  if [[ "$NEW" != "$CURRENT" ]]; then
    PREVIOUS=$CURRENT; PREVIOUS_SCHEMA=$CURRENT_SCHEMA
  fi
  CURRENT=$NEW; CURRENT_SCHEMA=$LIVE
  write_state rolled_back_manually
  echo "Application switched to $CURRENT; database unchanged"
  exit 0
fi
valid_image "$NEW" || fail 'Invalid release image'
DIRECTORY="$ROOT/releases/${NEW#dorm-app:}"
[[ -d "$DIRECTORY" ]] || fail 'Release upload is missing'
[[ ! -e "$DIRECTORY/loaded" ]] || fail 'Release tag already used'
docker image inspect "$NEW" >/dev/null 2>&1 && fail 'Remote image tag already exists'
pin_env "$DIRECTORY" "$NEW"
docker load -i "$DIRECTORY/image.tar"
touch "$DIRECTORY/loaded"
rm "$DIRECTORY/image.tar"
# Do not recreate an existing database or change its volume/configuration.
if ! docker inspect dorm-db >/dev/null 2>&1; then compose "$NEW" up -d db; fi
for attempt in {1..60}; do
  if [[ $(docker inspect -f '{{.State.Health.Status}}' dorm-db) == healthy ]]; then break; fi
  (( attempt < 60 )) || fail 'MySQL did not become healthy'
  sleep 2
done
check_db
DB_NAME=$(timeout 10 docker exec dorm-db sh -c 'printf "%s" "$MYSQL_DATABASE"')
compose "$NEW" run --rm -T --no-deps migrate sh -c 'test "$DB_HOST" = db && test "$DB_PORT" = 3306 && test "$DB_NAME" = "$1"' sh "$DB_NAME" </dev/null
STAGE=backup
BACKUP=''
write_state creating_backup
create_backup
write_state backed_up
STAGE=schema
BEFORE=$(schema)
# Bootstrap the old latest deployment using its actual image ID, never a mutable tag.
if [[ -z "$CURRENT" ]] && docker inspect dorm-app >/dev/null 2>&1; then
  LEGACY="dorm-app:legacy-$(date -u +%Y%m%d-%H%M%S)"
  docker image inspect "$LEGACY" >/dev/null 2>&1 && fail 'Legacy preservation tag already exists'
  OLD_DIRECTORY="$ROOT/releases/${LEGACY#dorm-app:}"
  mkdir "$OLD_DIRECTORY"
  cp "$ROOT/docker-compose.yml" "$ROOT/docker-compose.prod.yml" "$ROOT/.env" "$OLD_DIRECTORY/"
  sed -i 's@image: dorm-app:latest@image: ${DORM_IMAGE:?Set DORM_IMAGE to a versioned release}@g' "$OLD_DIRECTORY/docker-compose.yml"
  pin_env "$OLD_DIRECTORY" "$LEGACY"
  docker tag "$(docker inspect -f '{{.Image}}' dorm-app)" "$LEGACY"
  touch "$OLD_DIRECTORY/legacy"
  check_release "$LEGACY"
  CURRENT=$LEGACY; CURRENT_SCHEMA=$BEFORE
fi
write_state backed_up
WORKING_CURRENT=$CURRENT; WORKING_SCHEMA=$CURRENT_SCHEMA
WORKING_PREVIOUS=$PREVIOUS; WORKING_PREVIOUS_SCHEMA=$PREVIOUS_SCHEMA; WORKING_SET=1
STAGE=migration
# A failed migration may leave dirty DDL; never run force, down or restore.
if ! compose "$NEW" run --rm -T --no-deps migrate ./dorm-migrate up </dev/null; then
  echo 'Migration failed. Current migration state:' >&2
  schema >&2 || true
  echo 'Inspect the schema manually; no down/force/restore was performed.' >&2
  false
fi
AFTER=$(schema)
echo "Schema before=$BEFORE after=$AFTER"
STAGE=app
select_release "$NEW"
compose "$NEW" up -d --no-deps --force-recreate app
check_release "$NEW"
PREVIOUS=$CURRENT; PREVIOUS_SCHEMA=$CURRENT_SCHEMA
CURRENT=$NEW; CURRENT_SCHEMA=$AFTER
write_state success
echo "Deployment completed successfully: $CURRENT"
echo "Previous release: ${PREVIOUS:-none}"
echo "Database backup: $BACKUP"
echo 'Images and backups retained; no pruning or volume deletion performed.'
