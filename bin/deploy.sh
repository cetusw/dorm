#!/bin/bash
set -euo pipefail

SERVER_USER="ubuntu"
SERVER_IP="185.65.201.42"
SERVER_DIR="/home/ubuntu/dorm"

IMAGE_NAME="dorm-app"
IMAGE_TAR="dorm-app.tar"

REMOTE="$SERVER_USER@$SERVER_IP"

COMPOSE_BASE="docker-compose.yml"
COMPOSE_PROD="docker-compose.prod.yml"

echo "Checking required local files..."

required_paths=(
  "Dockerfile"
  "$COMPOSE_BASE"
  "$COMPOSE_PROD"
  ".env.prod"
  "bin"
  "mysql"
)

for path in "${required_paths[@]}"; do
  if [ ! -e "$path" ]; then
    echo "Missing required path: $path"
    exit 1
  fi
done

echo "Preparing remote directory..."
ssh "$REMOTE" "mkdir -p '$SERVER_DIR'"

echo "Building Docker image..."
docker build -t "$IMAGE_NAME" .

echo "Saving Docker image to tar..."
docker save "$IMAGE_NAME" -o "$IMAGE_TAR"

echo "Copying deployment files..."

scp "$IMAGE_TAR" "$REMOTE:$SERVER_DIR/"
scp "$COMPOSE_BASE" "$REMOTE:$SERVER_DIR/"
scp "$COMPOSE_PROD" "$REMOTE:$SERVER_DIR/"
scp ".env.prod" "$REMOTE:$SERVER_DIR/.env"

rsync -av --delete "bin/" "$REMOTE:$SERVER_DIR/bin/"
rsync -av --delete "mysql/" "$REMOTE:$SERVER_DIR/mysql/"

echo "Removing local image tar..."
rm "$IMAGE_TAR"

echo "Deploying on remote server..."
ssh "$REMOTE" << EOF
set -euo pipefail

cd "$SERVER_DIR"

COMPOSE_FILES="-f docker-compose.yml -f docker-compose.prod.yml"

echo "Setting execution permissions for scripts..."
chmod +x ./bin/* || true

echo "Loading Docker image..."
docker load -i "$IMAGE_TAR"
rm "$IMAGE_TAR"

echo "Stopping old containers..."
docker compose \$COMPOSE_FILES down

echo "Starting database..."
docker compose \$COMPOSE_FILES up -d db

echo "Waiting for database health..."
until [ "\$(docker inspect -f '{{.State.Health.Status}}' dorm-db)" = "healthy" ]; do
  sleep 2
done

echo "Applying migrations..."

MIGRATE_CONTAINER_ID=\$(docker compose \$COMPOSE_FILES run -d --no-deps migrate ./dorm-migrate up)

docker wait "\$MIGRATE_CONTAINER_ID" >/dev/null

echo "Migration logs:"
docker logs "\$MIGRATE_CONTAINER_ID"

MIGRATE_EXIT_CODE=\$(docker inspect "\$MIGRATE_CONTAINER_ID" --format='{{.State.ExitCode}}')
docker rm "\$MIGRATE_CONTAINER_ID" >/dev/null

if [ "\$MIGRATE_EXIT_CODE" -ne 0 ]; then
  echo "Migration failed with exit code \$MIGRATE_EXIT_CODE"
  exit "\$MIGRATE_EXIT_CODE"
fi

echo "Starting application..."
docker compose \$COMPOSE_FILES up -d --force-recreate app

echo "Deployment result:"
docker compose \$COMPOSE_FILES ps -a

echo "App logs:"
docker compose \$COMPOSE_FILES logs --tail=100 app || true

echo "Cleaning unused Docker images..."
docker image prune -f || true

echo "Deployment done."
EOF