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
CONFIRM_SCHEMA=0
if [[ "$ACTION" == rollback ]]; then
  case "$NEW" in
    '') ;;
    --confirm-schema-compatible) CONFIRM_SCHEMA=1 ;;
    *) fail 'Unknown rollback argument'; exit 1 ;;
  esac
  NEW=''
fi
compose() {
  local image=$1; shift
  local directory="$ROOT/releases/${image#dorm-app:}"
  DORM_IMAGE="$image" timeout 180 docker compose -p dorm --project-directory "$directory" \
    --env-file "$directory/.env" -f "$directory/docker-compose.yml" -f "$directory/docker-compose.prod.yml" "$@"
}
schema() {
  local exists value
  exists=$(mysql_query "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name='schema_migrations';") || return 1
  [[ "$exists" == 0 || "$exists" == 1 ]] || { fail "Unreadable migration table count: $exists"; return 1; }
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
valid_image() { [[ "$1" =~ ^dorm-app:[a-z0-9][a-z0-9_.-]{0,127}$ ]]; }
release_config() {
  local directory="$ROOT/releases/${1#dorm-app:}" file
  valid_image "$1" || { fail "Invalid release: $1"; return 1; }
  for file in .env docker-compose.yml docker-compose.prod.yml; do
    [[ -r "$directory/$file" ]] || { fail "Release configuration missing: $directory/$file"; return 1; }
  done
}
selected_release() {
  local link
  [[ -L "$ROOT/active" ]] || { fail 'active is missing or is not a symbolic link'; return 1; }
  link=$(readlink "$ROOT/active") || return 1
  [[ "$link" =~ ^releases/([a-z0-9][a-z0-9_.-]{0,127})$ && -d "$ROOT/active" ]] || {
    fail "Invalid or broken active link: $link"; return 1;
  }
  local image="dorm-app:${BASH_REMATCH[1]}"
  release_config "$image" || return 1
  printf '%s\n' "$image"
}
image_id() {
  local id
  id=$(timeout 10 docker image inspect --format '{{.Id}}' "$1") || return 1
  [[ "$id" =~ ^sha256:[0-9a-f]{64}$ ]] || { fail "Invalid image ID for $1: $id"; return 1; }
  printf '%s\n' "$id"
}
container_image() {
  local record id project service extra
  record=$(timeout 5 docker inspect --type container --format '{{.Image}} {{index .Config.Labels "com.docker.compose.project"}} {{index .Config.Labels "com.docker.compose.service"}}' dorm-app) || return 1
  read -r id project service extra <<< "$record"
  [[ "$id" =~ ^sha256:[0-9a-f]{64}$ && "$project" == dorm && "$service" == app && -z "$extra" ]] || {
    fail "Unexpected dorm-app identity/labels: $record"; return 1;
  }
  printf '%s\n' "$id"
}
check_identity() {
  local image=$1 expected=$2 actual tagged
  actual=$(container_image) || return 1
  tagged=$(image_id "$image") || return 1
  [[ "$actual" == "$expected" && "$tagged" == "$expected" ]] || {
    fail "Image identity mismatch for $image: expected=$expected actual=$actual tag_now=$tagged"; return 1;
  }
}
select_release() {
  local image=$1 file
  release_config "$image" || return 1
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
report_state() {
  local selected actual live containers
  echo "Last confirmed release: ${CURRENT:-none} (schema=${CURRENT_SCHEMA:-unknown})" >&2
  echo "Previous confirmed release: ${PREVIOUS:-none}" >&2
  if selected=$(selected_release); then echo "Selected release: $selected" >&2
  else echo 'Selected release: unavailable (see error above)' >&2; fi
  if actual=$(container_image); then echo "Container image ID: $actual" >&2
  else
    if containers=$(timeout 5 docker container ls -a --filter 'name=^/dorm-app$' --format '{{.Names}}') && [[ -z "$containers" ]]; then
      echo 'Container dorm-app: absent' >&2
    else echo 'Container identity: unavailable or unexpected labels (see error above)' >&2; fi
  fi
  if live=$(schema); then echo "Database schema: $live (clean)" >&2
  else echo 'Database schema: dirty or unreadable' >&2; fi
  echo "Attempt: ${NEW:-none}; database backup: ${BACKUP:-not created}" >&2
  echo 'No database rollback was performed.' >&2
}
state_value() {
  awk -v key="$1" 'index($0,key "=")==1 { count++; value=substr($0,length(key)+2) } END { if(count!=1) exit 1; print value }' "$ROOT/release-state" || {
    fail "Metadata field missing, duplicated or unreadable: $1"; return 1;
  }
}
read_confirmed_state() {
  local current current_schema previous previous_schema image
  current=$(state_value current) || return 1
  current_schema=$(state_value current_schema) || return 1
  previous=$(state_value previous) || return 1
  previous_schema=$(state_value previous_schema) || return 1
  if [[ ( -n "$current" && ( ! "$current_schema" =~ ^[0-9]+$ ) ) ||
        ( -z "$current" && ( -n "$current_schema" || -n "$previous" ) ) ||
        ( -n "$previous" && ( ! "$previous_schema" =~ ^[0-9]+$ || "$previous" == "$current" ) ) ||
        ( -z "$previous" && -n "$previous_schema" ) ]]; then
    fail "Contradictory metadata: current=$current current_schema=$current_schema previous=$previous previous_schema=$previous_schema"; return 1
  fi
  for image in "$current" "$previous"; do
    [[ -z "$image" ]] || valid_image "$image" || { fail "Invalid image in metadata: $image"; return 1; }
  done
  CURRENT=$current; CURRENT_SCHEMA=$current_schema
  PREVIOUS=$previous; PREVIOUS_SCHEMA=$previous_schema
}
CURRENT=''; CURRENT_SCHEMA=''; PREVIOUS=''; PREVIOUS_SCHEMA=''; BACKUP=''; STAGE=prepare; SWITCH_STARTED=0; WORKING_SET=0
if [[ -e "$ROOT/release-state" || -L "$ROOT/release-state" ]]; then
  [[ -f "$ROOT/release-state" && ! -L "$ROOT/release-state" && -r "$ROOT/release-state" ]] || { fail 'release-state is not a readable regular file'; exit 1; }
  read_confirmed_state || { report_state; exit 1; }
  for field in attempt result backup; do
    value=$(state_value "$field") || { report_state; exit 1; }
    if [[ "$field" == backup ]]; then BACKUP=$value; fi
  done
fi
write_state() {
  local status=$1 temporary
  temporary=$(mktemp "$ROOT/.state.XXXXXX") || return 1
  if ! printf 'current=%s\ncurrent_schema=%s\nprevious=%s\nprevious_schema=%s\nattempt=%s\nresult=%s\nbackup=%s\n' \
    "$CURRENT" "$CURRENT_SCHEMA" "$PREVIOUS" "$PREVIOUS_SCHEMA" "$NEW" "$status" "$BACKUP" > "$temporary"; then
    rm -f "$temporary"; return 1
  fi
  if ! mv "$temporary" "$ROOT/release-state"; then rm -f "$temporary"; return 1; fi
}
check_release() {
  local image=$1 end=$((SECONDS+120)) health html assets asset headers expected actual
  local directory="$ROOT/releases/${image#dorm-app:}" expected_id
  expected_id=$(image_id "$image") || return 1
  while (( SECONDS < end )); do
    check_identity "$image" "$expected_id" || return 1
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
      check_identity "$image" "$expected_id" || return 1
      [[ ! -f "$directory/legacy" ]] || echo 'Legacy rollback: verified static assets and MySQL; old image has no readiness endpoint' >&2
      return 0
    fi
    sleep 2
  done
  fail "Release validation timed out: $image"
}
# Recovery may repair only configuration: do not restart a healthy confirmed container.
restore_current() {
  local expected actual containers
  expected=$(image_id "$CURRENT") || return 1
  release_config "$CURRENT" || return 1
  selected_release >/dev/null || return 1
  if ! actual=$(container_image); then
    containers=$(timeout 5 docker container ls -a --filter 'name=^/dorm-app$' --format '{{.Names}}') || return 1
    [[ -z "$containers" ]] || return 1
    actual=absent
  fi
  # Repair all aliases as well as active after a partially completed selection.
  select_release "$CURRENT" || return 1
  if [[ "$actual" == "$expected" ]] && check_release "$CURRENT"; then
    echo 'Confirmed container already healthy; no recreation required' >&2
    return 0
  fi
  compose "$CURRENT" up -d --no-deps --force-recreate app || return 1
  check_release "$CURRENT"
}
on_failure() {
  local code=${1:-$?}
  if (( BASH_SUBSHELL > 0 )); then exit "$code"; fi
  trap - ERR HUP INT TERM
  set +e
  echo "Release operation failed during $STAGE" >&2
  timeout 15 docker logs --tail 100 dorm-app >&2
  if [[ "$STAGE" == confirm ]]; then
    # mv is the commit point: a signal may arrive immediately after it. Read
    # durable confirmed values instead of overwriting them with an old snapshot.
    if ! read_confirmed_state; then
      echo 'Confirmation outcome unreadable; no recovery or metadata overwrite performed. Only prior known metadata can be reported.' >&2
      if [[ ${WORKING_SET:-0} == 1 ]]; then
        CURRENT=$WORKING_CURRENT; CURRENT_SCHEMA=$WORKING_SCHEMA
        PREVIOUS=$WORKING_PREVIOUS; PREVIOUS_SCHEMA=$WORKING_PREVIOUS_SCHEMA
      fi
      report_state
      exit "$code"
    fi
  elif [[ ${WORKING_SET:-0} == 1 ]]; then
    CURRENT=$WORKING_CURRENT; CURRENT_SCHEMA=$WORKING_SCHEMA
    PREVIOUS=$WORKING_PREVIOUS; PREVIOUS_SCHEMA=$WORKING_PREVIOUS_SCHEMA
  fi
  local status="failed_$STAGE" live
  if [[ "$SWITCH_STARTED" == 1 && -n "$CURRENT" && ( "$STAGE" == switch || "$STAGE" == verify ) ]]; then
    if live=$(schema) && [[ "$live" == "${BEFORE:-unavailable}" && "$live" == "$CURRENT_SCHEMA" ]]; then
      echo "Schema unchanged; recovering $CURRENT" >&2
      if restore_current; then status=rolled_back
      else
        status=rollback_failed
        echo 'Application recovery failed; confirmed metadata retained' >&2
      fi
    else
      status=failed_schema_changed
      echo 'Automatic rollback blocked: schema changed, dirty or unreadable. Manual compatibility verification required.' >&2
    fi
  fi
  if ! write_state "$status"; then echo 'ERROR: cannot persist failure status; previous atomic release-state retained' >&2; fi
  report_state
  exit "$code"
}
trap on_failure ERR
trap 'on_failure 130' INT
trap 'on_failure 143' TERM
trap 'on_failure 129' HUP
if [[ "$ACTION" == rollback ]]; then
  [[ -n "$CURRENT" ]] || { fail 'No confirmed CURRENT; refusing to guess a rollback target'; report_state; exit 1; }
  check_db
  LIVE=$(schema)
  BEFORE=$LIVE
  SELECTED=$(selected_release)
  CONFIRMED_ID=$(image_id "$CURRENT")
  ACTUAL_ID=$(container_image)
  TARGET=$PREVIOUS; TARGET_SCHEMA=$PREVIOUS_SCHEMA; RECOVER_CURRENT=0
  if [[ "$SELECTED" != "$CURRENT" || "$ACTUAL_ID" != "$CONFIRMED_ID" ]]; then
    TARGET=$CURRENT; TARGET_SCHEMA=$CURRENT_SCHEMA; RECOVER_CURRENT=1
    echo "Selected/container state differs from CURRENT; recovering $CURRENT" >&2
  fi
  [[ -n "$TARGET" ]] || fail 'No previous working release recorded'
  NEW=$TARGET
  image_id "$TARGET" >/dev/null
  release_config "$TARGET"
  if [[ "$LIVE" != "$TARGET_SCHEMA" ]]; then
    [[ "$CONFIRM_SCHEMA" == 1 ]] || fail 'Schema differs. Verify compatibility first, then explicitly pass --confirm-schema-compatible. This does not prove compatibility.'
    echo 'WARNING: one-time manual compatibility confirmation; no database rollback will run' >&2
  fi
  WORKING_CURRENT=$CURRENT; WORKING_SCHEMA=$CURRENT_SCHEMA
  WORKING_PREVIOUS=$PREVIOUS; WORKING_PREVIOUS_SCHEMA=$PREVIOUS_SCHEMA; WORKING_SET=1
  STAGE=switch
  SWITCH_STARTED=1
  if (( RECOVER_CURRENT )); then
    restore_current
  else
    select_release "$NEW"
    compose "$NEW" up -d --no-deps --force-recreate app
  fi
  STAGE=verify
  check_release "$NEW"
  STAGE=confirm
  if (( ! RECOVER_CURRENT )); then
    PREVIOUS=$CURRENT; PREVIOUS_SCHEMA=$CURRENT_SCHEMA
    CURRENT=$NEW; CURRENT_SCHEMA=$TARGET_SCHEMA
  fi
  write_state rolled_back_manually
  echo "Application confirmed on $CURRENT; database unchanged"
  exit 0
fi
# An existing deployment must be readable before any backup/migration/switch.
if [[ -n "$CURRENT" ]]; then
  selected_release >/dev/null
  image_id "$CURRENT" >/dev/null
  container_image >/dev/null
fi
check_image_storage() {
  local archive=$1 size available docker_root endpoint
  size=$(wc -c < "$archive") || return 1
  available=$(df -Pk "$ROOT/releases" | awk 'END {printf "%.0f", $4*1024}') || return 1
  (( available >= size + 67108864 )) || { fail 'Insufficient release filesystem space before image load'; return 1; }
  docker_root=$(timeout 10 docker info --format '{{.DockerRootDir}}') || return 1
  if [[ -n ${DOCKER_HOST:-} ]]; then endpoint=$DOCKER_HOST
  elif ! endpoint=$(timeout 10 docker context inspect --format '{{.Endpoints.docker.Host}}'); then
    echo 'WARNING: Docker endpoint could not be determined; storage space was not verified' >&2
    return 0
  fi
  if [[ "$endpoint" != unix://* || "$docker_root" != /* || ! -d "$docker_root" ]]; then
    echo "WARNING: Docker storage $docker_root is not reliably accessible locally (endpoint=$endpoint); its free space is unknown" >&2
    return 0
  fi
  if ! available=$(df -Pk "$docker_root" | awk 'END {printf "%.0f", $4*1024}'); then
    echo "WARNING: cannot measure Docker storage filesystem: $docker_root" >&2
    return 0
  fi
  (( available >= size * 2 + 67108864 )) || { fail "Insufficient Docker storage space: $docker_root"; return 1; }
}
valid_image "$NEW" || fail 'Invalid release image'
DIRECTORY="$ROOT/releases/${NEW#dorm-app:}"
[[ -d "$DIRECTORY" ]] || fail 'Release upload is missing'
[[ ! -e "$DIRECTORY/loaded" ]] || fail 'Release tag already used'
docker image inspect "$NEW" >/dev/null 2>&1 && fail 'Remote image tag already exists'
pin_env "$DIRECTORY" "$NEW"
check_image_storage "$DIRECTORY/image.tar"
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
STAGE=switch
release_config "$NEW"
image_id "$NEW" >/dev/null
SWITCH_STARTED=1
select_release "$NEW"
compose "$NEW" up -d --no-deps --force-recreate app
STAGE=verify
check_release "$NEW"
STAGE=confirm
PREVIOUS=$CURRENT; PREVIOUS_SCHEMA=$CURRENT_SCHEMA
CURRENT=$NEW; CURRENT_SCHEMA=$AFTER
write_state success
echo "Deployment completed successfully: $CURRENT"
echo "Previous release: ${PREVIOUS:-none}"
echo "Database backup: $BACKUP"
echo 'Images and backups retained; no pruning or volume deletion performed.'
