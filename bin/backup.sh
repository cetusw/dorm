#!/usr/bin/env bash
set -euo pipefail
umask 077
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
[[ $# -ge 1 ]] || { echo "Usage: $0 user@host [--output-dir ./backups] [--remote-only]" >&2; exit 1; }
REMOTE=$1; shift
[[ "$REMOTE" =~ ^[a-zA-Z_][a-zA-Z0-9_.-]*@[a-zA-Z0-9._-]+$ ]] || { echo 'Expected SSH destination user@host' >&2; exit 1; }
OUTPUT=backups
REMOTE_ONLY=0
while (( $# )); do
  case $1 in
    --output-dir) [[ $# -ge 2 && -n "$2" ]] || { echo 'Missing output directory' >&2; exit 1; }; OUTPUT=$2; shift 2 ;;
    --remote-only) REMOTE_ONLY=1; shift ;;
    *) echo "Unknown argument: $1" >&2; exit 1 ;;
  esac
done
for command in ssh scp gzip mktemp; do command -v "$command" >/dev/null; done
SERVER_DIR="/home/${REMOTE%@*}/dorm"
BACKUP=$(ssh -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=15 -o ServerAliveCountMax=3 "$REMOTE" bash -s -- "$SERVER_DIR" backup < "$SCRIPT_DIR/release.sh")
[[ "$BACKUP" =~ ^$SERVER_DIR/backups/dorm-backup-[0-9]{8}-[0-9]{6}(-(preliminary|final)-[a-zA-Z0-9]{6})?\.sql\.gz$ ]] || { echo 'Invalid backup path returned by server' >&2; exit 1; }
echo "Server backup: $REMOTE:$BACKUP"
if (( REMOTE_ONLY )); then exit 0; fi
mkdir -p -- "$OUTPUT"
OUTPUT=$(cd "$OUTPUT" && pwd)
FINAL="$OUTPUT/${BACKUP##*/}"
[[ ! -e "$FINAL" ]] || { echo "Refusing to overwrite: $FINAL (server backup retained)" >&2; exit 1; }
TEMP=$(mktemp "$OUTPUT/.download.XXXXXX")
trap '[[ -z "$TEMP" ]] || rm -f -- "$TEMP"' EXIT
scp -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=15 -o ServerAliveCountMax=3 "$REMOTE:$BACKUP" "$TEMP"
[[ -s "$TEMP" ]] && gzip -t "$TEMP"
mv -n -- "$TEMP" "$FINAL"
[[ ! -e "$TEMP" ]] || { echo "Refusing to overwrite: $FINAL" >&2; exit 1; }
TEMP=''
echo "Local backup: $FINAL"
