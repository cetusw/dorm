#!/usr/bin/env bash
set -euo pipefail
umask 077
cd "$(dirname "$0")/.."
[[ $# -ge 1 ]] || { echo "Usage: $0 user@host [--pull]" >&2; exit 1; }
REMOTE=$1; shift
[[ "$REMOTE" =~ ^[a-zA-Z_][a-zA-Z0-9_.-]*@[a-zA-Z0-9._-]+$ ]] || { echo 'Expected SSH destination user@host' >&2; exit 1; }
BUILD_ARGS=()
if [[ $# == 1 && "$1" == --pull ]]; then BUILD_ARGS+=(--pull); shift; fi
[[ $# == 0 ]] || { echo 'Unknown arguments (only --pull is supported)' >&2; exit 1; }
for command in docker git ssh scp rsync mktemp; do command -v "$command" >/dev/null; done
for path in Dockerfile docker-compose.yml docker-compose.prod.yml .env.prod mysql bin/release.sh; do
  [[ -e "$path" ]] || { echo "Missing required path: $path" >&2; exit 1; }
done
BRANCH=$(git symbolic-ref --quiet --short HEAD) || { echo 'Detached HEAD: deploy from a named Git branch' >&2; exit 1; }
BRANCH_TAG=$(printf '%s' "$BRANCH" | LC_ALL=C tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9_.-]+/-/g; s/^[^a-z0-9]+//; s/[-.]+$//')
BRANCH_TAG=${BRANCH_TAG:0:110}
[[ -n "$BRANCH_TAG" ]] || { echo 'Branch name cannot form a Docker tag' >&2; exit 1; }
RELEASE="$BRANCH_TAG-$(date -u +%Y%m%d-%H%M%S)"
IMAGE="dorm-app:$RELEASE"
SERVER_DIR="/home/${REMOTE%@*}/dorm"
SSH=(ssh -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=15 -o ServerAliveCountMax=3)
"${SSH[@]}" "$REMOTE" 'command -v docker >/dev/null && command -v bash >/dev/null' </dev/null
if docker image inspect "$IMAGE" >/dev/null 2>&1; then echo "Release image already exists: $IMAGE" >&2; exit 1; fi
if [[ -n $(git status --porcelain) ]]; then echo 'Warning: building with uncommitted local changes' >&2; fi
RESERVATION="${TMPDIR:-/tmp}/dorm-image-$RELEASE.lock"
mkdir "$RESERVATION" || { echo "Release tag is reserved by another build: $RELEASE" >&2; exit 1; }
TEMP_DIR=''
trap '[[ -z "$TEMP_DIR" ]] || rm -rf -- "$TEMP_DIR"; rmdir "$RESERVATION"' EXIT
TEMP_DIR=$(mktemp -d)
echo "Building $IMAGE (branch=$BRANCH, UTC timestamp)"
docker build "${BUILD_ARGS[@]}" -t "$IMAGE" .
docker save "$IMAGE" -o "$TEMP_DIR/image.tar"
# Reserve room for both the transferred tar and its unpacked layers.
IMAGE_BYTES=$(wc -c < "$TEMP_DIR/image.tar")
"${SSH[@]}" "$REMOTE" bash -s -- "$SERVER_DIR" "$RELEASE" "$IMAGE_BYTES" <<'REMOTE_PREPARE'
set -euo pipefail
umask 077
root=$1; release=$2; image_bytes=$3
mkdir -p "$root/releases"
available=$(df -Pk "$root/releases" | awk 'END {printf "%.0f", $4*1024}')
if (( available < image_bytes * 2 + 67108864 )); then
  echo 'Insufficient space to transfer and load the release image' >&2
  exit 1
fi
# mkdir reserves the name; retries never overwrite an earlier upload.
mkdir "$root/releases/$release"
REMOTE_PREPARE
scp -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=15 -o ServerAliveCountMax=3 "$TEMP_DIR/image.tar" docker-compose.yml docker-compose.prod.yml "$REMOTE:$SERVER_DIR/releases/$RELEASE/"
scp -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=15 -o ServerAliveCountMax=3 .env.prod "$REMOTE:$SERVER_DIR/releases/$RELEASE/.env"
rsync -a -e 'ssh -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=15 -o ServerAliveCountMax=3' -- mysql/ "$REMOTE:$SERVER_DIR/releases/$RELEASE/mysql/"
"${SSH[@]}" "$REMOTE" bash -s -- "$SERVER_DIR" deploy "$IMAGE" < bin/release.sh
