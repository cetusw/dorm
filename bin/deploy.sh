#!/usr/bin/env bash
set -euo pipefail

SERVER_USER="${SERVER_USER:-ubuntu}"
SERVER_IP="${SERVER_IP:-185.65.201.42}"
SERVER_DIR="${SERVER_DIR:-/home/ubuntu/dorm}"

IMAGE_NAME="dorm-app:latest"
IMAGE_TAR="dorm-app.tar"

REMOTE="${SERVER_USER}@${SERVER_IP}"

COMPOSE_BASE="docker-compose.yml"
COMPOSE_PROD="docker-compose.prod.yml"
COMPOSE_FILES="-f ${COMPOSE_BASE} -f ${COMPOSE_PROD}"

required_paths=(
  "Dockerfile"
  "$COMPOSE_BASE"
  "$COMPOSE_PROD"
  ".env.prod"
  "mysql"
)

for path in "${required_paths[@]}"; do
  if [ ! -e "$path" ]; then
    echo "Missing required path: $path"
    exit 1
  fi
done

cleanup() {
  rm -f "$IMAGE_TAR"
}

trap cleanup EXIT

echo "Building image..."
docker build --pull -t "$IMAGE_NAME" .

echo "Saving image..."
docker save "$IMAGE_NAME" -o "$IMAGE_TAR"

echo "Copying deployment files..."
scp "$IMAGE_TAR" "$REMOTE:$SERVER_DIR/"
scp "$COMPOSE_BASE" "$REMOTE:$SERVER_DIR/"
scp "$COMPOSE_PROD" "$REMOTE:$SERVER_DIR/"
scp ".env.prod" "$REMOTE:$SERVER_DIR/.env"

rsync -av --delete "mysql/" "$REMOTE:$SERVER_DIR/mysql/"

ssh "$REMOTE" "
  chmod 600 '$SERVER_DIR/.env'
"

ssh "$REMOTE" <<EOF
set -euo pipefail

cd "$SERVER_DIR"

COMPOSE_FILES="-f docker-compose.yml -f docker-compose.prod.yml"

echo "Loading image..."
docker load -i "$IMAGE_TAR"
rm -f "$IMAGE_TAR"

echo "Starting database..."
docker compose \$COMPOSE_FILES up -d db

echo "Waiting for database..."
for attempt in \$(seq 1 60); do
  status=\$(docker inspect \
    --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' \
    dorm-db 2>/dev/null || true)

  if [ "\$status" = "healthy" ]; then
    break
  fi

  if [ "\$attempt" -eq 60 ]; then
    docker compose \$COMPOSE_FILES logs --tail=200 db
    exit 1
  fi

  sleep 2
done

echo "Applying migrations..."
docker compose \$COMPOSE_FILES run --rm -T --no-deps \
  migrate ./dorm-migrate up </dev/null

echo "Starting application..."
docker compose \$COMPOSE_FILES up -d \
  --no-deps \
  --force-recreate \
  app

echo "Checking application..."
for attempt in \$(seq 1 30); do
  if curl --fail --silent http://127.0.0.1:8080/app/ >/dev/null; then
    break
  fi

  if curl --fail https://dormkit.ru/app/ >/dev/null; then
      break
  fi

  if [ "\$attempt" -eq 30 ]; then
    docker compose \$COMPOSE_FILES logs --tail=200 app
    exit 1
  fi

  sleep 2
done

docker compose \$COMPOSE_FILES ps
docker compose \$COMPOSE_FILES logs --tail=100 app

docker image prune -f

echo "Deployment completed successfully."
EOF
