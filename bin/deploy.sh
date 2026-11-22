#!/usr/bin/env bash
set -euo pipefail
umask 077

remote_deploy() {
  set -Eeuo pipefail
  umask 077
  root=$1 upload=$2
  local image=$3 expected_sha256=$4 expected_platform=$5 server_id platform
  config_dir=$upload
  stage=preflight backup='' dump='' pid='' asset_container=''
  local file path target size docker_root db_name names has_app=0 has_db=0 service
  local end health html assets asset actual headers mount dump_command
  [[ "$upload" == "$root/.deploy.${upload##*.}" && "${upload##*.}" =~ ^[a-zA-Z0-9]{6}$ && -d "$upload" && ! -L "$upload" && -O "$upload" ]] || { echo 'Invalid staging directory' >&2; return 1; }
  trap '
    code=$?
    trap - EXIT HUP INT TERM
    set +e
    if [[ -n "$pid" ]]; then
      kill -KILL -- "-$pid" 2>/dev/null || kill -KILL "$pid" 2>/dev/null
      wait "$pid" 2>/dev/null
    fi
    [[ -z "$asset_container" ]] || timeout 15 docker rm -f "$asset_container" >/dev/null
    if (( code != 0 )); then
      echo "Deployment failed during $stage; backup: ${backup:-not created}" >&2
      echo "No automatic restart or rollback will run." >&2
      if [[ "$stage" == migration ]]; then echo "DDL may still be running. Inspect MySQL and schema_migrations manually before retrying; no down/force/restore was performed." >&2; fi
      timeout 10 docker ps -a --filter name=dorm-app --filter name=dorm-db >&2
      if [[ $(timeout 10 docker inspect --format "{{index .Config.Labels \"com.docker.compose.project\"}} {{index .Config.Labels \"com.docker.compose.service\"}}" dorm-app) == "dorm app" ]]; then
        timeout 10 docker inspect --format "{{if .State.Running}}Backend is running{{else}}Backend is stopped{{end}}" dorm-app >&2
        timeout 10 docker logs --tail 80 dorm-app >&2
      fi
    fi
    if [[ "$stage" == publication && -d "$root/.deployment-config" ]]; then
      if ! restore_configuration; then echo "Configuration recovery failed; preserve $root/.deployment-config and inspect it manually." >&2; fi
    fi
    [[ -z "$dump" ]] || rm -f -- "$dump"
    rm -rf -- "$upload"
    exit "$code"
  ' EXIT
  trap 'exit 129' HUP
  trap 'exit 130' INT
  trap 'exit 143' TERM
  fail() { echo "$*" >&2; return 1; }
  run() {
    local code=0 duration=$1
    shift
    timeout --kill-after=10 "$duration" "$@" &
    pid=$!
    wait "$pid" || code=$?
    pid=''
    return "$code"
  }
  compose() {
    local duration=180
    if [[ "$1" == --timeout ]]; then duration=$2; shift 2; fi
    run "$duration" docker compose -p dorm --project-directory "$config_dir" --env-file "$config_dir/.env" \
      -f "$config_dir/docker-compose.yml" -f "$config_dir/docker-compose.prod.yml" "$@"
  }
  restore_configuration() {
    local saved="$root/.deployment-config" file
    [[ ! -L "$saved" && -d "$saved" && -O "$saved" && -f "$saved/ready" && ! -L "$saved/ready" ]] || return 1
    [[ $(cat "$saved/ready") == configuration-backup ]] || return 1
    for file in docker-compose.yml docker-compose.prod.yml .env; do
      if [[ -e "$saved/old/$file" || -L "$saved/old/$file" ]]; then
        [[ -f "$saved/old/$file" || -L "$saved/old/$file" ]] || return 1
      else
        [[ -f "$saved/missing/$file" && ! -L "$saved/missing/$file" ]] || return 1
      fi
      [[ ! -d "$root/$file" || -L "$root/$file" ]] || return 1
    done
    for file in docker-compose.yml docker-compose.prod.yml .env; do
      if [[ -e "$saved/old/$file" || -L "$saved/old/$file" ]]; then
        cp -a --no-dereference -- "$saved/old/$file" "$saved/restore-$file" || return 1
        mv -Tf -- "$saved/restore-$file" "$root/$file" || return 1
      else
        rm -f -- "$root/$file" || return 1
      fi
    done
    sync -f "$root" || return 1
    rm -rf -- "$saved" || return 1
    sync -f "$root" || return 1
    echo 'Recovered configuration files from interrupted publication; backend was not restarted.' >&2
  }
  publish_configuration() {
    local file saved="$upload/configuration-backup"
    mkdir "$saved" "$saved/old" "$saved/missing"
    for file in docker-compose.yml docker-compose.prod.yml .env; do
      if [[ -e "$root/$file" || -L "$root/$file" ]]; then
        cp -a --no-dereference -- "$root/$file" "$saved/old/$file"
      else
        : > "$saved/missing/$file"
      fi
    done
    printf '%s\n' configuration-backup > "$saved/ready"
    sync -f "$saved"
    mv -T -- "$saved" "$root/.deployment-config"
    sync -f "$root"
    echo "Publishing configuration; if interrupted, rerun deploy to recover $root/.deployment-config before using Compose manually." >&2
    for file in docker-compose.yml docker-compose.prod.yml .env; do
      mv -Tf -- "$upload/$file" "$root/$file"
    done
    config_dir=$root
    compose config --quiet
    sync -f "$root"
    rm -rf -- "$root/.deployment-config"
    sync -f "$root"
  }
  query() {
    timeout 30 docker exec -i dorm-db sh -c "export MYSQL_PWD=\"\$MYSQL_ROOT_PASSWORD\"; exec mysql -uroot --batch --skip-column-names \"\$MYSQL_DATABASE\"" <<< "$1"
  }
  schema() {
    local exists value
    exists=$(query "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name='schema_migrations';")
    [[ "$exists" == 0 || "$exists" == 1 ]] || fail 'Cannot read migration table state'
    if [[ "$exists" == 0 ]]; then return; fi
    value=$(query 'SELECT version, dirty FROM schema_migrations;')
    [[ -z "$value" || "$value" =~ ^[0-9]+[[:space:]]+0$ ]] || fail 'Schema is dirty or unreadable; manual inspection required'
  }
  space() {
    local directory=$1 required=$2 free
    free=$(df -Pk "$directory" | awk 'END {printf "%.0f", $4*1024}')
    if [[ ! "$free" =~ ^[0-9]+$ ]] || (( free < required )); then fail "Insufficient or unknown free space: $directory"; fi
  }
  validate_assets() {
    local invalid
    [[ -d "$1" && ! -L "$1" ]] || fail 'Assets must be a regular directory'
    invalid=$(find "$1" ! -type d ! -type f -print -quit)
    [[ -z "$invalid" ]] || fail 'Assets contain links or special files'
  }
  publish_assets() {
    local source=$1 entry dest
    validate_assets "$source"
    find "$source" -mindepth 1 -print0 > "$upload/asset-paths"
    while IFS= read -r -d '' entry; do
      dest="$root/shared/assets/${entry#"$source/"}"
      if [[ -d "$entry" ]]; then
        [[ ! -e "$dest" || -d "$dest" ]] || fail 'Asset directory conflict'
      elif [[ -e "$dest" ]]; then
        if [[ ! -f "$dest" ]] || ! cmp -s "$entry" "$dest"; then fail 'Immutable asset conflict; use a new filename/version'; fi
      fi
    done < "$upload/asset-paths"
    while IFS= read -r -d '' entry; do
      dest="$root/shared/assets/${entry#"$source/"}"
      if [[ -d "$entry" ]]; then
        mkdir -p "$dest"
        chmod 755 "$dest"
      elif [[ ! -e "$dest" ]]; then
        chmod 644 "$entry"
        ln -T -- "$entry" "$dest"
      fi
    done < "$upload/asset-paths"
  }

  [[ "$expected_sha256" =~ ^[0-9a-f]{64}$ ]] || fail 'Invalid expected Docker archive checksum'
  [[ "$expected_platform" =~ ^[a-z0-9_]+/[a-z0-9_]+$ ]] || fail 'Invalid expected Docker image platform'
  [[ ! -L "$root/.deployment.lock" && ( ! -e "$root/.deployment.lock" || -f "$root/.deployment.lock" ) ]] || fail 'Unsafe deployment lock'
  exec 9>>"$root/.deployment.lock"
  flock -n 9 || fail 'Another deployment is running'
  if [[ -e "$root/.deployment-config" || -L "$root/.deployment-config" ]]; then
    restore_configuration || fail 'Interrupted configuration publication requires manual recovery'
  fi
  timeout 15 docker info >/dev/null
  for path in "$root/backups" "$root/shared" "$root/shared/assets"; do
    [[ ! -L "$path" && ( ! -e "$path" || -d "$path" ) ]] || fail "Unsafe directory: $path"
    mkdir -p "$path"
  done
  chmod 755 "$root/shared" "$root/shared/assets"
  validate_assets "$root/shared/assets"
  [[ $(stat -c %d "$upload") == $(stat -c %d "$root/shared/assets") ]] || fail 'Staging and assets must share a filesystem'
  for file in docker-compose.yml docker-compose.prod.yml .env mysql; do
    if [[ -L "$root/$file" ]]; then
      [[ $(readlink "$root/$file") == "active/$file" && -L "$root/active" && $(readlink "$root/active") =~ ^releases/[a-zA-Z0-9_.-]+$ ]] || fail "Unknown symlink: $file"
      target=$(readlink -f "$root/$file")
      [[ "$target" == "$root"/releases/* && -r "$target" ]] || fail "Invalid old configuration target: $file"
    fi
    if [[ "$file" == mysql ]]; then
      target="$root/mysql"
      if [[ ! -e "$target" && ! -L "$target" && ( -e "$root/active" || -L "$root/active" ) ]]; then
        [[ -L "$root/active" && $(readlink "$root/active") =~ ^releases/[a-zA-Z0-9_.-]+$ ]] || fail 'Unknown active path; inspect missing mysql/ manually'
        target=$(readlink -f "$root/active/mysql") || fail 'Cannot resolve old mysql/ after interrupted transfer; inspect it manually'
        [[ "$target" == "$root"/releases/* && -d "$target" ]] || fail 'Cannot verify old mysql/ after interrupted transfer; inspect it manually'
      fi
      if [[ -e "$target" || -L "$target" ]]; then
        if [[ ! -d "$target" ]] || ! diff -qr -- "$target" "$upload/mysql" >/dev/null; then fail 'Existing mysql/ differs from uploaded build files; inspect it manually'; fi
      fi
    else
      [[ ! -e "$root/$file" || -f "$root/$file" ]] || fail "Unexpected configuration path: $file"
      [[ -s "$upload/$file" && ! -L "$upload/$file" ]] || fail "Missing uploaded configuration: $file"
    fi
  done
  for service in db app; do
    names=$(timeout 10 docker container ls -a --filter "name=^/dorm-$service$" --format '{{.Names}}')
    if [[ -n "$names" ]]; then
      [[ "$names" == "dorm-$service" && $(timeout 10 docker inspect --format '{{index .Config.Labels "com.docker.compose.project"}} {{index .Config.Labels "com.docker.compose.service"}}' "dorm-$service") == "dorm $service" ]] || fail "Foreign dorm-$service container"
      if [[ "$service" == db ]]; then
        [[ $(timeout 10 docker inspect --format '{{range .Mounts}}{{if eq .Destination "/var/lib/mysql"}}{{.Type}} {{.Name}}{{end}}{{end}}' dorm-db) == 'volume dorm_dorm-db-data' ]] || fail 'Unexpected MySQL volume; refusing to select another database'
        has_db=1
      else has_app=1; fi
    fi
  done
  names=$(timeout 10 docker volume ls --filter 'name=^dorm_dorm-db-data$' --format '{{.Name}}')
  if [[ -n "$names" ]]; then
    [[ "$names" == dorm_dorm-db-data && $(timeout 10 docker volume inspect --format '{{index .Labels "com.docker.compose.project"}} {{index .Labels "com.docker.compose.volume"}}' dorm_dorm-db-data) == 'dorm dorm-db-data' ]] || fail 'Foreign database volume'
  fi
  export DORM_IMAGE="$image" DORM_ASSETS_DIR="$root/shared/assets"
  run 30 docker compose -p dorm --project-directory "$upload" --env-file "$upload/.env" \
    -f "$upload/docker-compose.yml" -f "$upload/docker-compose.prod.yml" config --quiet
  stage=image
  [[ -f "$upload/image.tar" && ! -L "$upload/image.tar" && -s "$upload/image.tar" ]] || fail 'Missing or empty Docker archive'
  actual=$(sha256sum < "$upload/image.tar" | awk '{print $1}')
  [[ "$actual" == "$expected_sha256" ]] || fail 'Transferred Docker archive checksum mismatch'
  names=$(timeout 10 docker image ls --filter "reference=$image" --format '{{.Repository}}:{{.Tag}}')
  [[ -z "$names" ]] || fail 'Deployment image tag already exists on server'
  size=$(wc -c < "$upload/image.tar")
  space "$root" "$((size * 2 + 67108864))"
  docker_root=$(timeout 10 docker info --format '{{.DockerRootDir}}')
  space "$docker_root" "$((size * 2 + 67108864))"
  run 180 docker load -i "$upload/image.tar"
  server_id=$(timeout 10 docker image inspect --format '{{.Id}}' "$image")
  [[ "$server_id" =~ ^sha256:[0-9a-f]{64}$ ]] || fail 'Invalid loaded Docker image ID'
  platform=$(timeout 10 docker image inspect --format '{{.Os}}/{{.Architecture}}' "$image")
  [[ "$platform" == "$expected_platform" ]] || fail 'Loaded Docker image platform mismatch'
  rm -- "$upload/image.tar"
  stage=configuration
  sed -E '/^[[:space:]]*(export[[:space:]]+)?DORM_(IMAGE|ASSETS_DIR)[[:space:]]*=/d' "$upload/.env" > "$upload/env-pinned"
  printf '\nDORM_IMAGE=%s\nDORM_ASSETS_DIR=%s\n' "$image" "$DORM_ASSETS_DIR" >> "$upload/env-pinned"
  mv "$upload/env-pinned" "$upload/.env"
  compose config --quiet
  compose run --rm -T --no-deps app ./dorm --check-config
  stage=mysql
  if (( has_db )); then
    if [[ $(timeout 10 docker inspect --format '{{.State.Running}}' dorm-db) != true ]]; then run 60 docker start dorm-db; fi
  else
    compose up -d --no-deps db
    [[ $(timeout 10 docker inspect --format '{{range .Mounts}}{{if eq .Destination "/var/lib/mysql"}}{{.Type}} {{.Name}}{{end}}{{end}}' dorm-db) == 'volume dorm_dorm-db-data' ]] || fail 'Unexpected MySQL volume'
  fi
  end=$((SECONDS+150))
  until [[ $(timeout 10 docker inspect --format '{{.State.Health.Status}}' dorm-db) == healthy ]]; do
    (( SECONDS < end )) || fail 'MySQL readiness timed out'
    sleep 2
  done
  [[ $(query 'SELECT 1;') == 1 ]] || fail 'MySQL query failed'
  db_name=$(docker exec dorm-db sh -c 'printf "%s" "$MYSQL_DATABASE"')
  compose run --rm -T --no-deps migrate sh -c "test \"\$DB_HOST\" = db && test \"\$DB_PORT\" = 3306 && test \"\$DB_NAME\" = \"\$1\"" sh "$db_name"
  stage=backup
  size=$(query 'SELECT COALESCE(SUM(data_length + index_length),0) FROM information_schema.tables WHERE table_schema=DATABASE();')
  [[ "$size" =~ ^[0-9]+$ ]] || fail 'Cannot estimate database size'
  space "$root/backups" "$((size * 2 + 67108864))"
  dump=$(mktemp "$root/backups/.dump.XXXXXX")
  target="$root/backups/dorm-backup-$(date -u +%Y%m%d-%H%M%S)-${dump##*.}.sql.gz"
  dump_command=$(cat <<'DUMP'
docker exec dorm-db sh -c 'export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"; exec mysqldump -uroot --single-transaction --quick --no-tablespaces --set-gtid-purged=OFF --default-character-set=utf8mb4 --hex-blob --routines --events --triggers --databases "$MYSQL_DATABASE"' | gzip > "$1"
DUMP
)
  run 1800 bash -o pipefail -c "$dump_command" bash "$dump"
  [[ -s "$dump" ]] && gzip -t "$dump"
  gzip -dc "$dump" | LC_ALL=C grep -a -E '^CREATE (DATABASE|TABLE)' >/dev/null
  sync -f "$dump"
  ln -T -- "$dump" "$target"
  backup=$target
  sync -f "$root/backups"
  rm -- "$dump"
  dump=''
  echo "Verified backup: $backup"
  stage=schema
  schema
  if (( has_app )); then
    stage=stop
    health=$(timeout 10 docker inspect --format '{{.State.Status}} {{.State.Running}} {{.State.Restarting}} {{.State.Paused}} {{.State.Pid}}' dorm-app)
    if [[ "$health" == 'running true false false '* ]]; then
      run 45 docker stop --time 30 dorm-app
      [[ $(timeout 10 docker inspect --format '{{.State.Running}} {{.State.ExitCode}} {{.State.OOMKilled}}' dorm-app) == 'false 0 false' ]] || fail 'Backend did not stop cleanly; inspect writers before applying migrations'
    elif [[ "$health" != 'exited false false false 0' && "$health" != 'created false false false 0' ]]; then
      fail 'Backend state is unsafe for migrations; inspect active/restarting/paused processes manually'
    fi
    [[ $(timeout 10 docker inspect --format '{{.State.Status}} {{.State.Running}} {{.State.Restarting}} {{.State.Paused}} {{.State.Pid}}' dorm-app) =~ ^(exited|created)\ false\ false\ false\ 0$ ]] || fail 'Backend is not fully stopped'
  fi
  stage=migration
  compose --timeout 900 run --rm -T --no-deps migrate ./dorm-migrate up
  schema
  query 'SELECT id,user_id,created_at,expires_at,revoked_at FROM auth_session LIMIT 0;' >/dev/null
  [[ $(query "SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='user' AND index_name='uq_active_user_login' AND non_unique=0;") == 1 ]] || fail 'Active login uniqueness is missing'
  stage=assets
  if (( has_app )); then
    mount=$(timeout 10 docker inspect --format '{{range .Mounts}}{{if eq .Destination "/app/web/app/assets"}}{{.Type}} {{.Source}} {{.RW}}{{end}}{{end}}' dorm-app)
    if [[ "$mount" != "bind $root/shared/assets false" ]]; then
      run 120 docker cp dorm-app:/app/web/app/assets "$upload/old-assets"
      publish_assets "$upload/old-assets"
    fi
  fi
  size=$(timeout 10 docker image inspect --format '{{.Size}}' "$image")
  [[ "$size" =~ ^[0-9]+$ ]] || fail 'Cannot estimate asset storage'
  space "$root/shared/assets" "$((size * 2 + 67108864))"
  asset_container="dorm-assets-${upload##*.}"
  run 30 docker create --name "$asset_container" "$image" >/dev/null
  run 120 docker cp "$asset_container:/app/web/app/assets" "$upload/assets"
  run 15 docker rm "$asset_container" >/dev/null
  asset_container=''
  [[ -n $(find "$upload/assets" -type f -print -quit) ]] || fail 'Image has no frontend assets'
  publish_assets "$upload/assets"
  stage=publication
  if [[ -L "$root/mysql" ]]; then unlink "$root/mysql"; fi
  if [[ ! -e "$root/mysql" ]]; then mv -T -- "$upload/mysql" "$root/mysql"; fi
  publish_configuration
  stage=start
  compose up -d --no-deps --force-recreate app
  stage=health
  end=$((SECONDS+120))
  while :; do
    [[ $(timeout 10 docker inspect --format '{{.Image}}' dorm-app) == "$server_id" ]] || fail 'Application image identity mismatch'
    [[ $(timeout 10 docker inspect --format '{{index .Config.Labels "com.docker.compose.project"}} {{index .Config.Labels "com.docker.compose.service"}}' dorm-app) == 'dorm app' ]] || fail 'Application Compose identity mismatch'
    health=$(timeout 10 docker inspect --format '{{.State.Status}} {{if .State.Health}}{{.State.Health.Status}}{{end}}' dorm-app)
    [[ "$health" != exited* && "$health" != dead* && "$health" != *unhealthy ]] || fail 'Backend exited or became unhealthy'
    if [[ "$health" == 'running healthy' ]] && \
      [[ $(curl --noproxy '*' --silent --show-error --max-time 3 --output /dev/null --write-out '%{http_code}' http://127.0.0.1:8080/health/live) == 204 ]] && \
      [[ $(curl --noproxy '*' --silent --show-error --max-time 3 --output /dev/null --write-out '%{http_code}' http://127.0.0.1:8080/health/ready) == 204 ]]; then break; fi
    (( SECONDS < end )) || fail 'Backend readiness timed out'
    sleep 2
  done
  [[ $(timeout 10 docker inspect --format '{{range .Mounts}}{{if eq .Destination "/app/web/app/assets"}}{{.Type}} {{.Source}} {{.RW}}{{end}}{{end}}' dorm-app) == "bind $root/shared/assets false" ]] || fail 'Shared assets mount mismatch'
  html=$(curl --noproxy '*' --fail --silent --show-error --max-time 3 http://127.0.0.1:8080/app/)
  assets=$(printf '%s' "$html" | grep -Eo '/app/assets/[a-zA-Z0-9_./-]+\.(css|js)' | sort -u)
  [[ "$assets" == *'.js'* && "$assets" == *'.css'* ]] || fail 'HTML has no CSS/JS assets'
  while IFS= read -r asset; do
    headers=$(curl --noproxy '*' --fail --silent --show-error --head --max-time 3 "http://127.0.0.1:8080$asset")
    if printf '%s' "$headers" | grep -qi '^Content-Type: text/html'; then fail 'Asset resolved to HTML'; fi
    actual=$(curl --noproxy '*' --fail --silent --show-error --max-time 3 "http://127.0.0.1:8080$asset" | sha256sum | awk '{print $1}')
    [[ "$actual" == "$(sha256sum < "$root/shared/assets/${asset#/app/assets/}" | awk '{print $1}')" ]] || fail 'Served asset content mismatch'
  done <<< "$assets"
  echo "Deployment successful: $image; backup: $backup"
  exit 0
}

cd "$(dirname "$0")/.."
[[ $# == 1 || ( $# == 2 && "$2" == --pull ) ]] || { echo "Usage: $0 user@host [--pull]" >&2; exit 1; }
REMOTE=$1
[[ "$REMOTE" =~ ^[a-zA-Z_][a-zA-Z0-9_.-]*@[a-zA-Z0-9._-]+$ ]] || { echo 'Expected SSH destination user@host' >&2; exit 1; }
for command in docker ssh scp rsync mktemp date timeout sha256sum awk; do command -v "$command" >/dev/null; done
for path in Dockerfile docker-compose.yml docker-compose.prod.yml .env.prod mysql; do [[ -e "$path" ]] || { echo "Missing required path: $path" >&2; exit 1; }; done
timeout 15 docker info >/dev/null
SSH=(ssh -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=15 -o ServerAliveCountMax=3)
ROOT="/home/${REMOTE%@*}/dorm"
LOCAL_DIR=$(mktemp -d)
UPLOAD=''
STARTED=0
LOCAL_STAGE=preflight
cleanup_local() {
  local code=$?
  trap - EXIT HUP INT TERM
  if (( code != 0 )); then echo "Deployment failed during local $LOCAL_STAGE; inspect remote output if the server phase started." >&2; fi
  if [[ -n "$UPLOAD" && "$STARTED" == 0 ]]; then
    "${SSH[@]}" "$REMOTE" bash -s -- "$ROOT" "$UPLOAD" <<'CLEANUP' || true
set -euo pipefail
[[ "$2" == "$1/.deploy.${2##*.}" && "${2##*.}" =~ ^[a-zA-Z0-9]{6}$ && -d "$2" && ! -L "$2" && -O "$2" ]] && rm -rf -- "$2"
CLEANUP
  fi
  rm -rf -- "$LOCAL_DIR"
  exit "$code"
}
trap cleanup_local EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
"${SSH[@]}" "$REMOTE" sh -s <<'REQUIRE'
set -eu
for command in bash docker rsync flock gzip mktemp timeout curl find sha256sum cmp stat ln sync diff awk cat chmod cp date df grep mkdir mv readlink rm sed sleep sort tr unlink wc; do
  command -v "$command" >/dev/null || { echo "Missing required command: $command" >&2; exit 1; }
done
timeout 15 docker info >/dev/null
version=$(docker compose version --short)
printf '%s\n' "$version" | grep -Eq '^v?[2-9][0-9]*\.'
REQUIRE
IMAGE="dorm-app:deploy-$(date -u +%Y%m%d-%H%M%S)-${LOCAL_DIR##*/}"
BUILD_ARGS=()
if [[ "${2:-}" == --pull ]]; then BUILD_ARGS+=(--pull); fi
LOCAL_STAGE=build
docker build "${BUILD_ARGS[@]}" -t "$IMAGE" .
LOCAL_STAGE=configuration
cp docker-compose.yml docker-compose.prod.yml "$LOCAL_DIR/"
cp .env.prod "$LOCAL_DIR/.env"
export DORM_IMAGE="$IMAGE" DORM_ASSETS_DIR="$LOCAL_DIR/assets"
docker compose -p dorm --project-directory "$LOCAL_DIR" --env-file "$LOCAL_DIR/.env" -f "$LOCAL_DIR/docker-compose.yml" -f "$LOCAL_DIR/docker-compose.prod.yml" config --quiet
cat > "$LOCAL_DIR/check.yml" <<'CHECK'
services:
  app:
    image: ${DORM_IMAGE}
    env_file: .env
    environment:
      APP_ENV: production
    network_mode: none
CHECK
docker compose -p "dorm-check-$(printf '%s' "${LOCAL_DIR##*.}" | tr '[:upper:]' '[:lower:]')" --project-directory "$LOCAL_DIR" --env-file "$LOCAL_DIR/.env" \
  -f "$LOCAL_DIR/check.yml" run --rm -T --no-deps app sh -c 'for value in "$DB_USER" "$DB_PASSWORD" "$DB_ROOT_PASSWORD" "$DB_NAME"; do [ -n "$value" ] && [ "$value" != TODO ] || { echo "Incomplete production database configuration" >&2; exit 1; }; done; test "$DB_HOST" = db && test "$DB_PORT" = 3306 && test "$DB_DRIVER" = mysql || exit 1; exec ./dorm --check-config'
EXPECTED_PLATFORM=$(timeout 10 docker image inspect --format '{{.Os}}/{{.Architecture}}' "$IMAGE")
[[ "$EXPECTED_PLATFORM" =~ ^[a-z0-9_]+/[a-z0-9_]+$ ]] || { echo 'Invalid local Docker image platform' >&2; exit 1; }
LOCAL_STAGE=archive
docker save "$IMAGE" -o "$LOCAL_DIR/image.tar"
EXPECTED_SHA256=$(sha256sum < "$LOCAL_DIR/image.tar" | awk '{print $1}')
[[ "$EXPECTED_SHA256" =~ ^[0-9a-f]{64}$ ]] || { echo 'Invalid local Docker archive checksum' >&2; exit 1; }
BYTES=$(wc -c < "$LOCAL_DIR/image.tar")
LOCAL_STAGE=upload
UPLOAD=$("${SSH[@]}" "$REMOTE" bash -s -- "$ROOT" "$BYTES" <<'PREPARE'
set -euo pipefail
umask 077
[[ ! -L "$1" && ( ! -e "$1" || -d "$1" ) ]] || { echo 'Unsafe deployment root' >&2; exit 1; }
mkdir -p "$1"
free=$(df -Pk "$1" | awk 'END {printf "%.0f", $4*1024}')
(( free >= $2 * 2 + 67108864 )) || { echo 'Insufficient upload space' >&2; exit 1; }
mktemp -d "$1/.deploy.XXXXXX"
PREPARE
)
[[ "$UPLOAD" == "$ROOT/.deploy.${UPLOAD##*.}" && "${UPLOAD##*.}" =~ ^[a-zA-Z0-9]{6}$ ]] || { echo 'Invalid upload directory' >&2; exit 1; }
LOCAL_STAGE=transfer
scp -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=15 -o ServerAliveCountMax=3 "$LOCAL_DIR/image.tar" "$LOCAL_DIR/docker-compose.yml" "$LOCAL_DIR/docker-compose.prod.yml" "$LOCAL_DIR/.env" "$REMOTE:$UPLOAD/"
rsync -a -e 'ssh -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=15 -o ServerAliveCountMax=3' -- mysql/ "$REMOTE:$UPLOAD/mysql/"
STARTED=1
LOCAL_STAGE=server
"${SSH[@]}" "$REMOTE" bash -s -- "$ROOT" "$UPLOAD" "$IMAGE" "$EXPECTED_SHA256" "$EXPECTED_PLATFORM" < <(declare -f remote_deploy; printf 'remote_deploy "$@"\n')
