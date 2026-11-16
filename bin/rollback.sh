#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
[[ $# -ge 1 ]] || { echo "Usage: $0 user@host [--confirm-schema-compatible]" >&2; exit 1; }
REMOTE=$1; shift
[[ "$REMOTE" =~ ^[a-zA-Z_][a-zA-Z0-9_.-]*@[a-zA-Z0-9._-]+$ ]] || { echo 'Expected SSH destination user@host' >&2; exit 1; }
CONFIRM=''
if [[ $# == 1 && "$1" == --confirm-schema-compatible ]]; then CONFIRM=$1; shift; fi
[[ $# == 0 ]] || { echo 'Unknown rollback arguments' >&2; exit 1; }
command -v ssh >/dev/null
ssh -o BatchMode=yes -o ConnectTimeout=10 -o ServerAliveInterval=15 -o ServerAliveCountMax=3 "$REMOTE" bash -s -- "/home/${REMOTE%@*}/dorm" rollback "$CONFIRM" < "$SCRIPT_DIR/release.sh"
